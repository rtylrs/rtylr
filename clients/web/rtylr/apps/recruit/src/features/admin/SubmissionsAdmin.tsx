import React, { useState } from 'react';
import { cn, EmptyState } from '@voxire/ui';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Mail, Phone, FileText, Inbox } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import {
  useAdminSubmissions,
  useUpdateSubmissionStatus,
  useAdminFormFields,
  fetchResumeURL,
} from '../../shared/hooks/queries/useAdmin';
import type { Submission } from '../../shared/lib/types';

/**
 * SubmissionsAdmin - HR-facing inbox for all candidate applications.
 * Kept deliberately simple (no kanban, no stages swimlane) for MVP;
 * the status dropdown on each row is enough to move a candidate
 * through the funnel.
 */
// Exported so the locale suite can assert a label exists for every stage.
// The labels are looked up with a runtime built key, which the key resolution
// guard cannot see, so nothing else would catch a missing one.
export const STATUSES = ['new', 'reviewed', 'interviewing', 'offered', 'hired', 'rejected'];

const PAGE_SIZE = 20;

const SubmissionsAdmin: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  // A stage value is what the API stores, so it stays as it is and only its
  // label is translated. One helper keeps the filter chips, the dropdown and
  // the confirmation prompt from drifting apart.
  const statusLabel = (status: string): string => t(`admin.submissions.status.${status}`);
  const [filter, setFilter] = useState<string>('');
  const [page, setPage] = useState<number>(0);
  const { data, isLoading, refetch } = useAdminSubmissions(filter, PAGE_SIZE, page * PAGE_SIZE);
  const update = useUpdateSubmissionStatus();
  const submissions = data?.submissions ?? [];

  const handleFilterChange = (newFilter: string) => {
    setFilter(newFilter);
    setPage(0);
  };

  return (
    <>
      <HelmetContainer title={t('admin.submissions.metaTitle')} />
      <div className="px-4 py-6 sm:px-10 sm:py-8">
        <h1 className="text-2xl font-semibold tracking-tight">{t('admin.submissions.heading')}</h1>
        <p className="mt-1 text-sm text-white/60">
          {t('admin.submissions.applicationCount', { count: submissions.length })}
          {filter && ` ${t('admin.submissions.inFilter', { stage: statusLabel(filter) })}`}
        </p>

        <div className="mt-6 flex flex-wrap items-center gap-2 text-sm">
          <FilterChip active={filter === ''} onClick={() => handleFilterChange('')}>
            {t('admin.submissions.all')}
          </FilterChip>
          {STATUSES.map((s) => (
            <FilterChip key={s} active={filter === s} onClick={() => handleFilterChange(s)}>
              {statusLabel(s)}
            </FilterChip>
          ))}
        </div>

        <div
          className="mt-6 overflow-hidden rounded-lg border"
          style={{
            borderColor: 'var(--recruit-admin-border)',
            background: 'var(--recruit-admin-surface)',
          }}
        >
          {isLoading ? (
            <div className="p-10 text-center text-sm text-white/60">
              {t('admin.submissions.loading')}
            </div>
          ) : submissions.length === 0 ? (
            <EmptyState
              icon={Inbox}
              title={
                filter
                  ? t('common.emptyStates.submissionsNoneFiltered')
                  : t('common.emptyStates.submissionsNone')
              }
              action={
                filter
                  ? {
                      label: t('common.emptyStates.clearFilter'),
                      onClick: () => handleFilterChange(''),
                    }
                  : { label: t('common.emptyStates.refresh'), onClick: () => void refetch() }
              }
            />
          ) : (
            <>
              <ul className="divide-y" style={{ borderColor: 'var(--recruit-admin-border)' }}>
                {submissions.map((s) => (
                  <li
                    key={s.id}
                    className="flex items-start gap-6 px-6 py-5 transition hover:bg-white/5"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="flex items-baseline gap-3">
                        <div className="font-semibold">{s.candidateName}</div>
                        <div className="text-xs text-white/50">
                          {new Date(s.createdAt).toLocaleDateString()}
                        </div>
                      </div>
                      <div className="mt-0.5 text-sm text-white/70">
                        Applied for <span className="text-white">{s.careerTitle || '-'}</span>
                      </div>
                      <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-white/60">
                        <a
                          href={`mailto:${s.email}`}
                          className="inline-flex items-center gap-1 hover:text-white"
                        >
                          <Mail className="h-3 w-3" />
                          {s.email}
                        </a>
                        {s.phone && (
                          <span className="inline-flex items-center gap-1">
                            <Phone className="h-3 w-3" />
                            {s.phone}
                          </span>
                        )}
                        {s.resumeUrl && (
                          <button
                            type="button"
                            onClick={async () => {
                              try {
                                const url = await fetchResumeURL(s.id);
                                const parsed = new URL(url);
                                if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
                                  toast.error(t('admin.submissions.invalidResumeScheme'));
                                  return;
                                }
                                window.open(url, '_blank', 'noopener,noreferrer');
                              } catch {
                                toast.error(t('admin.submissions.openFailed'));
                              }
                            }}
                            className="inline-flex items-center gap-1 hover:text-white"
                          >
                            <FileText className="h-3 w-3" />
                            {t('admin.submissions.resume')}
                          </button>
                        )}
                      </div>
                      {s.coverLetter && (
                        <p className="mt-3 line-clamp-3 text-sm text-white/70">{s.coverLetter}</p>
                      )}

                      {/* Custom question answers - shown after the cover
                      letter so the reviewer sees them in reading order.
                      We render labels only once we've fetched the field
                      definitions for this submission's career. */}
                      <CustomAnswers submission={s} />
                    </div>

                    <select
                      value={s.status}
                      onChange={(e) => {
                        const newStatus = e.target.value;
                        if (s.status === 'hired' && newStatus !== 'hired') {
                          toast.warning(
                            t('admin.submissions.confirmUnhire', {
                              name: s.candidateName,
                              from: statusLabel('hired'),
                              to: statusLabel(newStatus),
                            }),
                            {
                              action: {
                                label: t('admin.submissions.confirm'),
                                onClick: () =>
                                  update.mutate(
                                    { id: s.id, status: newStatus },
                                    {
                                      onSuccess: () =>
                                        toast.success(t('admin.submissions.statusUpdated')),
                                      onError: () =>
                                        toast.error(t('admin.submissions.statusFailed')),
                                    }
                                  ),
                              },
                            }
                          );
                          return;
                        }
                        update.mutate(
                          { id: s.id, status: newStatus },
                          {
                            onSuccess: () => toast.success(t('admin.submissions.statusUpdated')),
                            onError: () => toast.error(t('admin.submissions.statusFailed')),
                          }
                        );
                      }}
                      className="rounded-md border bg-black/40 px-3 py-1.5 text-sm outline-none focus:border-white/40"
                      style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
                    >
                      {STATUSES.map((opt) => (
                        <option key={opt} value={opt}>
                          {statusLabel(opt)}
                        </option>
                      ))}
                    </select>
                  </li>
                ))}
              </ul>
              {submissions.length > 0 && (
                <div
                  className="flex items-center justify-between border-t px-6 py-3 text-xs text-white/60"
                  style={{ borderColor: 'var(--recruit-admin-border)' }}
                >
                  <span>{t('admin.submissions.page', { page: page + 1 })}</span>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      disabled={page === 0}
                      onClick={() => setPage((p) => Math.max(0, p - 1))}
                      className="rounded border px-3 py-1 text-xs text-white hover:bg-white/10 disabled:opacity-40"
                      style={{ borderColor: 'var(--recruit-admin-border)' }}
                    >
                      {t('admin.submissions.previous')}
                    </button>
                    <button
                      type="button"
                      disabled={submissions.length < PAGE_SIZE}
                      onClick={() => setPage((p) => p + 1)}
                      className="rounded border px-3 py-1 text-xs text-white hover:bg-white/10 disabled:opacity-40"
                      style={{ borderColor: 'var(--recruit-admin-border)' }}
                    >
                      {t('admin.submissions.next')}
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </>
  );
};

/**
 * CustomAnswers renders a submission's per-role custom field answers.
 * Fetches the field definitions for this career to label each answer;
 * TanStack Query dedupes the same career_id across all rows on the
 * page so a batch of submissions for the same role fires one fetch.
 */
const CustomAnswers: React.FC<{ submission: Submission }> = ({ submission }) => {
  const { data } = useAdminFormFields(submission.careerId);
  if (!submission.answersJson) return null;

  let parsed: Record<string, string | string[]> = {};
  try {
    const raw = JSON.parse(submission.answersJson);
    if (raw && typeof raw === 'object') parsed = raw as Record<string, string | string[]>;
  } catch {
    return null;
  }
  const ids = Object.keys(parsed);
  if (ids.length === 0) return null;

  const fields = data?.fields ?? [];
  const byId = new Map(fields.map((f) => [String(f.id), f]));

  return (
    <div
      className="mt-3 space-y-1.5 rounded-md border p-3"
      style={{ borderColor: 'var(--recruit-admin-border)', background: 'rgba(255,255,255,0.02)' }}
    >
      {ids.map((id) => {
        const field = byId.get(id);
        const value = parsed[id];
        const display = Array.isArray(value)
          ? value.join(', ')
          : value === '1'
            ? 'Yes'
            : String(value);
        return (
          <div key={id} className="flex gap-2 text-xs">
            <span className="min-w-[140px] text-white/50">{field?.label ?? `Question ${id}`}</span>
            <span className="text-white/80">{display || '-'}</span>
          </div>
        );
      })}
    </div>
  );
};

const FilterChip: React.FC<{
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}> = ({ active, onClick, children }) => (
  <button
    type="button"
    onClick={onClick}
    className={cn(
      'inline-flex min-h-[44px] items-center rounded-full px-4 py-2 text-xs font-medium transition',
      active ? 'bg-white text-black' : 'border border-white/10 text-white/70 hover:bg-white/5'
    )}
  >
    {children}
  </button>
);

export default SubmissionsAdmin;
