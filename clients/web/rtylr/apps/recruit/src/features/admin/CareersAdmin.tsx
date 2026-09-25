import React, { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Plus, Trash2, ExternalLink, Edit3, Briefcase } from 'lucide-react';
import { APP_URLS } from '@voxire/config';
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

/**
 * CareersAdmin - the HR-facing list of this business's openings. Lets
 * an admin create, edit, publish/unpublish, and archive roles. The
 * editor opens as a side drawer rather than a separate route so the
 * list stays visible (and rank-ordered by created_at) while editing.
 */
const CareersAdmin: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: careersData, isLoading } = useAdminCareers();
  const { data: customData } = useAdminCustomization();
  const { data: submissionsData } = useAdminSubmissions();
  const create = useCreateCareer();
  const update = useUpdateCareer();
  const del = useDeleteCareer();

  const [editing, setEditing] = useState<RecruitCareer | null>(null);
  const [creating, setCreating] = useState(false);

  const careers = careersData?.careers ?? [];
  const submissions = submissionsData?.submissions ?? [];
  const slug = customData?.customization?.slug;

  return (
    <>
      <HelmetContainer title={t('admin.careers.metaTitle')} />
      <div className="px-4 py-6 sm:px-10 sm:py-8">
        <div className="flex flex-col sm:flex-row sm:items-baseline justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{t('admin.careers.heading')}</h1>
            <p className="mt-1 text-sm text-white/60">
              {t('admin.careers.roleCount', { count: careers.length })} ·{' '}
              {t('admin.careers.publishedCount', {
                count: careers.filter((c) => c.isPublished).length,
              })}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            {slug && (
              <a
                href={`${APP_URLS.recruit}/${slug}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex min-h-[44px] items-center gap-2 rounded-md border border-white/10 px-3.5 py-2.5 text-sm hover:bg-white/5"
              >
                <ExternalLink className="h-4 w-4" />
                {t('admin.careers.viewPublic')}
              </a>
            )}
            <button
              type="button"
              onClick={() => setCreating(true)}
              className="inline-flex min-h-[44px] items-center gap-2 rounded-md bg-white px-4 py-2.5 text-sm font-semibold text-black hover:bg-white/90"
            >
              <Plus className="h-4 w-4" />
              {t('admin.careers.newRole')}
            </button>
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
            <div className="p-10 text-center text-sm text-white/60">
              {t('admin.careers.loading')}
            </div>
          ) : careers.length === 0 ? (
            <EmptyState
              icon={Briefcase}
              title={t('common.emptyStates.rolesNone')}
              action={{
                label: t('common.emptyStates.createRole'),
                onClick: () => setCreating(true),
              }}
            />
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b" style={{ borderColor: 'var(--recruit-admin-border)' }}>
                  <Th>{t('admin.careers.colTitle')}</Th>
                  <Th>{t('admin.careers.colDepartment')}</Th>
                  <Th>{t('admin.careers.colLocation')}</Th>
                  <Th>{t('admin.careers.colStatus')}</Th>
                  <Th align="right">{t('admin.careers.colActions')}</Th>
                </tr>
              </thead>
              <tbody>
                {careers.map((c) => (
                  <tr
                    key={c.id}
                    className="border-b transition hover:bg-white/5"
                    style={{ borderColor: 'var(--recruit-admin-border)' }}
                  >
                    <td className="px-4 py-3 font-medium">{c.title}</td>
                    <td className="px-4 py-3 text-sm text-white/70">{c.department || '-'}</td>
                    <td className="px-4 py-3 text-sm text-white/70">{c.location || '-'}</td>
                    <td className="px-4 py-3">
                      <span
                        className="inline-block rounded-full px-2.5 py-0.5 text-xs font-medium"
                        style={{
                          background: c.isPublished
                            ? 'color-mix(in srgb, var(--recruit-admin-success) 15%, transparent)'
                            : 'rgba(255,255,255,0.08)',
                          color: c.isPublished
                            ? 'var(--recruit-admin-success)'
                            : 'rgba(255,255,255,0.6)',
                        }}
                      >
                        {c.isPublished ? t('admin.careers.published') : t('admin.careers.draft')}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="inline-flex items-center gap-1">
                        <IconButton onClick={() => setEditing(c)} title={t('admin.careers.edit')}>
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
                                label: t('admin.careers.archive'),
                                onClick: () =>
                                  del.mutate(c.id, {
                                    onSuccess: () => toast.success(t('admin.careers.archived')),
                                    onError: () => toast.error(t('admin.careers.archiveFailed')),
                                  }),
                              },
                            });
                          }}
                          title={t('admin.careers.archive')}
                        >
                          <Trash2 className="h-4 w-4" />
                        </IconButton>
                      </div>
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
                      toast.success(t('admin.careers.updated'));
                      setEditing(null);
                    },
                    onError: () => toast.error(t('admin.careers.updateFailed')),
                  }
                );
              } else {
                create.mutate(body, {
                  onSuccess: () => {
                    toast.success(t('admin.careers.created'));
                    setCreating(false);
                  },
                  onError: () => toast.error(t('admin.careers.createFailed')),
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
    className="px-4 py-3 text-xs font-medium uppercase tracking-wider text-white/50"
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
    className="rounded-md p-1.5 text-white/70 transition hover:bg-white/10 hover:text-white"
  >
    {children}
  </button>
);

export default CareersAdmin;
