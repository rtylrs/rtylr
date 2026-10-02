import { useRecruitPermission } from '../../shared/hooks/useRecruitPermission';
import AdminQueryState from '../../shared/components/AdminQueryState';
import React, { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Plus, Trash2, ExternalLink, Edit3, Briefcase } from 'lucide-react';
import { RECRUIT_PUBLIC_URL } from '@voxire/config';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';

import {
  useAdminCareers,
  useAdminSubmissions,
  useCreateCareer,
  useDeleteCareer,
  useUpdateCareer,
  useAdminCustomization,
} from '../../shared/hooks/queries/useAdmin';
import type { RecruitCareer } from '../../shared/lib/types';
import CareerEditor from './CareerEditor';

const CareersAdmin: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation('recruitCommon');
  const { data: careersData, isLoading, isError, refetch } = useAdminCareers();
  const { data: customData } = useAdminCustomization();
  const canReadCandidates = useRecruitPermission('recruit:candidates:view');
  const { data: submissionsData } = useAdminSubmissions(undefined, 20, 0, canReadCandidates);
  const canManage = useRecruitPermission('recruit:jobs:manage');
  const create = useCreateCareer();
  const update = useUpdateCareer();
  const del = useDeleteCareer();

  const [editing, setEditing] = useState<RecruitCareer | null>(null);
  const [creating, setCreating] = useState(false);

  const careers = careersData?.careers ?? [];
  const submissions = submissionsData?.submissions ?? [];
  const slug = customData?.customization?.slug;

  if (isError) return <AdminQueryState retry={() => void refetch()} />;

  return (
    <>
      <HelmetContainer title={t('recruitAdmin.careers.metaTitle')} />
      <div className="px-4 py-6 sm:px-10 sm:py-8">
        <div className="flex flex-col sm:flex-row sm:items-baseline justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">
              {t('recruitAdmin.careers.heading')}
            </h1>
            <p className="mt-1 text-sm text-[var(--dash-text-muted)]">
              {t('recruitAdmin.careers.roleCount', { count: careers.length })} ·{' '}
              {t('recruitAdmin.careers.publishedCount', {
                count: careers.filter((c) => c.isPublished).length,
              })}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            {slug && (
              <a
                href={`${RECRUIT_PUBLIC_URL}/${slug}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex min-h-[44px] items-center gap-2 rounded-md border border-[var(--dash-border)] px-3.5 py-2.5 text-sm hover:bg-[var(--dash-hover)]"
              >
                <ExternalLink className="h-4 w-4" />
                {t('recruitAdmin.careers.viewPublic')}
              </a>
            )}
            {canManage && (
              <button
                type="button"
                onClick={() => setCreating(true)}
                className="inline-flex min-h-[44px] items-center gap-2 rounded-md bg-[var(--dash-hover)] px-4 py-2.5 text-sm font-semibold text-[var(--dash-text)] hover:bg-[var(--dash-active)]"
              >
                <Plus className="h-4 w-4" />
                {t('recruitAdmin.careers.newRole')}
              </button>
            )}
          </div>
        </div>

        <div
          className="mt-8 overflow-hidden rounded-lg border"
          style={{
            borderColor: 'var(--recruit-admin-border)',
            background: 'var(--recruit-admin-surface)',
          }}
        >
          {isLoading ? (
            <div className="p-10 text-center text-sm text-[var(--dash-text-muted)]">
              {t('recruitAdmin.careers.loading')}
            </div>
          ) : careers.length === 0 ? (
            <EmptyState
              icon={Briefcase}
              title={t('recruitCommon.emptyStates.rolesNone')}
              action={
                canManage
                  ? {
                      label: t('recruitCommon.emptyStates.createRole'),
                      onClick: () => setCreating(true),
                    }
                  : undefined
              }
            />
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b" style={{ borderColor: 'var(--recruit-admin-border)' }}>
                  <Th>{t('recruitAdmin.careers.colTitle')}</Th>
                  <Th>{t('recruitAdmin.careers.colDepartment')}</Th>
                  <Th>{t('recruitAdmin.careers.colLocation')}</Th>
                  <Th>{t('recruitAdmin.careers.colStatus')}</Th>
                  <Th align="right">{t('recruitAdmin.careers.colActions')}</Th>
                </tr>
              </thead>
              <tbody>
                {careers.map((c) => (
                  <tr
                    key={c.id}
                    className="border-b transition hover:bg-[var(--dash-hover)]"
                    style={{ borderColor: 'var(--recruit-admin-border)' }}
                  >
                    <td className="px-4 py-3 font-medium">{c.title}</td>
                    <td className="px-4 py-3 text-sm text-[var(--dash-text-muted)]">
                      {c.department || '-'}
                    </td>
                    <td className="px-4 py-3 text-sm text-[var(--dash-text-muted)]">
                      {c.location || '-'}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className="inline-block rounded-full px-2.5 py-0.5 text-xs font-medium"
                        style={{
                          background: c.isPublished
                            ? 'color-mix(in srgb, var(--recruit-admin-success) 15%, transparent)'
                            : 'var(--dash-hover)',
                          color: c.isPublished
                            ? 'var(--recruit-admin-success)'
                            : 'var(--dash-text-muted)',
                        }}
                      >
                        {c.isPublished
                          ? t('recruitAdmin.careers.published')
                          : t('recruitAdmin.careers.draft')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      {canManage && (
                        <div className="inline-flex items-center gap-1">
                          <IconButton
                            onClick={() => setEditing(c)}
                            title={t('recruitAdmin.careers.edit')}
                          >
                            <Edit3 className="h-4 w-4" />
                          </IconButton>
                          <IconButton
                            onClick={() => {
                              const openCount =
                                submissions.filter(
                                  (s) =>
                                    s.careerId === c.id &&
                                    s.status !== 'hired' &&
                                    s.status !== 'rejected'
                                ).length ?? 0;
                              const message =
                                openCount > 0
                                  ? `Archive "${c.title}"? It has ${openCount} open application(s).`
                                  : `Archive "${c.title}"?`;
                              toast.warning(message, {
                                action: {
                                  label: t('recruitAdmin.careers.archive'),
                                  onClick: () =>
                                    del.mutate(c.id, {
                                      onSuccess: () =>
                                        toast.success(t('recruitAdmin.careers.archived')),
                                      onError: () =>
                                        toast.error(t('recruitAdmin.careers.archiveFailed')),
                                    }),
                                },
                              });
                            }}
                            title={t('recruitAdmin.careers.archive')}
                          >
                            <Trash2 className="h-4 w-4" />
                          </IconButton>
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {(creating || editing) && (
          <CareerEditor
            career={editing ?? undefined}
            pending={create.isPending || update.isPending}
            onClose={() => {
              setEditing(null);
              setCreating(false);
            }}
            onSave={(body) => {
              if (editing) {
                update.mutate(
                  { ...body, id: editing.id },
                  {
                    onSuccess: () => {
                      toast.success(t('recruitAdmin.careers.updated'));
                      setEditing(null);
                    },
                    onError: () => toast.error(t('recruitAdmin.careers.updateFailed')),
                  }
                );
              } else {
                create.mutate(body, {
                  onSuccess: () => {
                    toast.success(t('recruitAdmin.careers.created'));
                    setCreating(false);
                  },
                  onError: () => toast.error(t('recruitAdmin.careers.createFailed')),
                });
              }
            }}
          />
        )}
      </div>
    </>
  );
};

const Th: React.FC<{ children: React.ReactNode; align?: 'left' | 'right' }> = ({
  children,
  align = 'left',
}) => (
  <th
    className="px-4 py-3 text-xs font-medium uppercase tracking-wider text-[var(--dash-text-dim)]"
    style={{ textAlign: align }}
  >
    {children}
  </th>
);

const IconButton: React.FC<{
  children: React.ReactNode;
  onClick: () => void;
  title: string;
}> = ({ children, onClick, title }) => (
  <button
    type="button"
    onClick={onClick}
    title={title}
    className="rounded-md p-1.5 text-[var(--dash-text-muted)] transition hover:bg-[var(--dash-hover)] hover:text-[var(--dash-text)]"
  >
    {children}
  </button>
);

export default CareersAdmin;
