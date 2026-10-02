import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { Search, Plus, Lock, Globe, Archive, FolderOpen } from 'lucide-react';
import { useAllProjects } from './useAllProjects';
import type { Project } from '@voxire/voxie/flow/project/v2/types';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import ProjectIcon from '../../shared/components/ProjectIcon';
import { Button } from '../../../../shared/components/ui/button';
import { useUserStore } from '../../shared/stores/userStore';
import { useFlowStore } from '../../shared/stores/flowStore';
import { canFlowPermission } from '../../shared/lib/roleUtils';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';

const STATUS_FILTERS = ['', 'active', 'completed'] as const;
const VISIBILITY_FILTERS = ['', 'private', 'public'] as const;

function ProjectCard({ project }: { project: Project }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isArchived = !!project.archivedAt;

  return (
    <button
      type="button"
      onClick={() => navigate({ to: `/flow/p/${project.id}` })}
      className="w-full text-start rounded-lg border p-4 transition-colors hover:bg-[var(--flow-hover)] focus-visible:ring-2 focus-visible:ring-ring"
      style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-surface)' }}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2 min-w-0">
          {/* G-I1: a lucide glyph, not the stored value painted as text. A
              project whose icon is "building 2" rendered those words in
              coloured 14px type, wrapping onto two lines and colliding with
              the title. */}
          <span
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
            style={{
              background: project.color ? withAlpha(project.color, 13) : 'var(--accent-muted)',
              color: project.color ? onTint(project.color) : 'var(--accent-on-muted)',
            }}
            aria-hidden
          >
            <ProjectIcon icon={project.icon} size={16} />
          </span>
          <div className="min-w-0">
            <p className="font-medium text-sm truncate" style={{ color: 'var(--flow-text)' }}>
              {project.name}
            </p>
            {project.description ? (
              <p className="text-xs truncate" style={{ color: 'var(--flow-text-muted)' }}>
                {project.description}
              </p>
            ) : null}
          </div>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {project.visibility === 'public' ? (
            <Globe
              size={13}
              style={{ color: 'var(--flow-text-muted)' }}
              aria-label={t('flow.visibility.public')}
            />
          ) : (
            <Lock
              size={13}
              style={{ color: 'var(--flow-text-muted)' }}
              aria-label={t('flow.visibility.private')}
            />
          )}
          {isArchived && (
            <Archive
              size={13}
              style={{ color: 'var(--flow-text-muted)' }}
              aria-label={t('flow.visibility.archived')}
            />
          )}
        </div>
      </div>
      <div
        className="mt-3 flex items-center gap-3 text-xs"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <span>{t('flow.allProjects.taskCount', { count: project.taskCount ?? 0 })}</span>
        <span>{t('flow.allProjects.completedCount', { count: project.completedCount ?? 0 })}</span>
        <span>{t('flow.allProjects.memberCount', { count: project.memberCount ?? 0 })}</span>
      </div>
    </button>
  );
}

export const AllProjects: React.FC = () => {
  const { t } = useTranslation();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const canCreate = canFlowPermission('flow:projects:create', access);
  const setProjectCreateOpen = useFlowStore((state) => state.setProjectCreateOpen);
  const [q, setQ] = useState('');
  const [status, setStatus] = useState<string>('');
  const [visibility, setVisibility] = useState<string>('');

  const allProjectsQuery = useAllProjects({
    q: q || undefined,
    status: status || undefined,
    visibility: visibility || undefined,
    limit: 100,
  });
  const { data, isLoading } = allProjectsQuery;
  const loadFailed = hasQueryFailed(allProjectsQuery);

  const projects = data?.projects ?? [];
  const total = data?.total ?? 0;

  const hasFilters = q !== '' || status !== '' || visibility !== '';

  // visible_via_report is a per-caller projection from the server: true only
  // when the caller sees the project because they manage someone on it. With no
  // reports every project falls in the first bucket and the grouped layout is
  // never rendered, so nothing changes for the common case.
  const ownProjects = projects.filter((p) => !p.visibleViaReport);
  const reportProjects = projects.filter((p) => p.visibleViaReport);

  return (
    <div className="h-full overflow-y-auto">
      <div className="mx-auto max-w-4xl px-4 py-6 sm:px-6">
        <header className="mb-6">
          <div className="flex items-center justify-between gap-3">
            <h1>{t('flow.allProjects.title')}</h1>
            {canCreate && (
              <Button onClick={() => setProjectCreateOpen(true)}>
                <Plus size={16} />
                {t('flow.shell.newProject')}
              </Button>
            )}
          </div>

          {/* Search + Filters */}
          <div className="mt-3 flex flex-wrap items-center gap-2">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <Search
                size={14}
                className="absolute start-2.5 top-1/2 -translate-y-1/2 pointer-events-none"
                style={{ color: 'var(--flow-text-muted)' }}
                aria-hidden
              />
              <input
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder={t('flow.allProjects.search')}
                className="w-full h-9 ps-8 pe-3 rounded-lg border text-sm focus:ring-1 focus:ring-ring"
                style={{
                  borderColor: 'var(--flow-border-control)',
                  background: 'var(--flow-surface)',
                  color: 'var(--flow-text)',
                }}
                aria-label={t('flow.allProjects.search')}
              />
            </div>

            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              className="h-9 px-2 rounded-lg border text-sm"
              style={{
                borderColor: 'var(--flow-border-control)',
                background: 'var(--flow-surface)',
                color: 'var(--flow-text)',
              }}
              aria-label={t('flow.allProjects.filterStatus')}
            >
              {STATUS_FILTERS.map((s) => (
                <option key={s} value={s}>
                  {s === '' ? t('flow.allProjects.allStatuses') : t(`flow.projectStatus.${s}`)}
                </option>
              ))}
            </select>

            <select
              value={visibility}
              onChange={(e) => setVisibility(e.target.value)}
              className="h-9 px-2 rounded-lg border text-sm"
              style={{
                borderColor: 'var(--flow-border-control)',
                background: 'var(--flow-surface)',
                color: 'var(--flow-text)',
              }}
              aria-label={t('flow.allProjects.filterVisibility')}
            >
              {VISIBILITY_FILTERS.map((v) => (
                <option key={v} value={v}>
                  {v === '' ? t('flow.allProjects.allVisibilities') : t(`flow.visibility.${v}`)}
                </option>
              ))}
            </select>
          </div>
        </header>

        {/* Loading skeleton */}
        {isLoading ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {[...Array(6)].map((_, i) => (
              <div
                key={i}
                className="h-[88px] rounded-lg animate-pulse"
                style={{ background: 'var(--flow-hover)' }}
              />
            ))}
          </div>
        ) : loadFailed ? (
          <ListLoadFailed onRetry={() => void allProjectsQuery.refetch()} />
        ) : projects.length === 0 ? (
          <div className="portal-flow-empty py-16 text-center">
            <div className="portal-flow-empty-icon">
              <FolderOpen size={24} />
            </div>
            <p className="font-semibold text-sm mb-1" style={{ color: 'var(--flow-text)' }}>
              {hasFilters ? t('flow.allProjects.noResults') : t('flow.allProjects.emptyTitle')}
            </p>
            <p className="text-xs opacity-60 mb-4" style={{ color: 'var(--flow-text-muted)' }}>
              {hasFilters
                ? t('flow.allProjects.tryClearing')
                : t('flow.allProjects.emptyDescription')}
            </p>
            {!hasFilters && canCreate && (
              <Button type="button" onClick={() => setProjectCreateOpen(true)} className="mt-2">
                <Plus size={16} /> {t('flow.allProjects.createFirst')}
              </Button>
            )}
          </div>
        ) : reportProjects.length > 0 ? (
          // P9: projects visible only because the caller manages someone on them
          // are grouped separately. Mixed into one list they would read as the
          // caller's own work, which is the thing visible_via_report exists to
          // prevent.
          <>
            {ownProjects.length > 0 && (
              <section className="mb-6">
                <h2
                  className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t('flow.allProjects.yourProjects')}
                </h2>
                <div className="grid gap-3 sm:grid-cols-2">
                  {ownProjects.map((p) => (
                    <ProjectCard key={p.id} project={p} />
                  ))}
                </div>
              </section>
            )}
            <section>
              <h2
                className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.allProjects.yourReports')}
              </h2>
              <div className="grid gap-3 sm:grid-cols-2">
                {reportProjects.map((p) => (
                  <ProjectCard key={p.id} project={p} />
                ))}
              </div>
            </section>
          </>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {projects.map((p) => (
              <ProjectCard key={p.id} project={p} />
            ))}
          </div>
        )}

        {!isLoading && projects.length > 0 && (
          <p className="mt-4 text-xs text-center" style={{ color: 'var(--flow-text-muted)' }}>
            {total} {total === 1 ? 'project' : 'projects'}
          </p>
        )}
      </div>
    </div>
  );
};

export default AllProjects;
