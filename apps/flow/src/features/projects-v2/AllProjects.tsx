import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { Search, Plus, Lock, Globe, Archive } from 'lucide-react';
import { useAllProjects } from './useAllProjects';
import type { Project } from '@voxire/voxie/flow/project/v2/types';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import ProjectIcon from '../../shared/components/ProjectIcon';

const STATUS_FILTERS = ['', 'active', 'completed'] as const;
const VISIBILITY_FILTERS = ['', 'private', 'public'] as const;

function ProjectCard({ project }: { project: Project }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isArchived = !!project.archivedAt;

  return (
    <button
      type="button"
      onClick={() => navigate({ to: `/p/${project.id}` })}
      className="w-full text-left rounded-lg border p-4 transition-colors hover:bg-[var(--flow-hover)] focus-visible:focus-visible:ring-2 focus-visible:ring-ring"
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
              background: project.color ? `${project.color}20` : 'var(--accent-muted)',
              color: project.color ?? 'var(--accent)',
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
        <span>{project.taskCount ?? 0} tasks</span>
        <span>{project.completedCount ?? 0} done</span>
        <span>{project.memberCount ?? 0} members</span>
      </div>
    </button>
  );
}

export const AllProjects: React.FC = () => {
  const { t } = useTranslation();
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
          <h1 className="text-screen font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.allProjects.title')}
          </h1>

          {/* Search + Filters */}
          <div className="mt-3 flex flex-wrap items-center gap-2">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <Search
                size={14}
                className="absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none"
                style={{ color: 'var(--flow-text-muted)' }}
                aria-hidden
              />
              <input
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder={t('flow.allProjects.search')}
                className="w-full h-9 pl-8 pr-3 rounded-lg border text-sm focus:ring-1 focus:ring-ring"
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
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <p className="font-semibold text-sm mb-1" style={{ color: 'var(--flow-text)' }}>
              {hasFilters ? t('flow.allProjects.noResults') : t('flow.allProjects.emptyTitle')}
            </p>
            <p className="text-xs opacity-60 mb-4" style={{ color: 'var(--flow-text-muted)' }}>
              {hasFilters
                ? t('flow.allProjects.tryClearing')
                : t('flow.allProjects.emptyDescription')}
            </p>
            {!hasFilters && (
              <button
                type="button"
                onClick={() => setQ('')}
                className="inline-flex items-center gap-1.5 px-3 h-11 md:h-8 rounded-lg font-semibold text-xs bg-accent text-accent-foreground hover:opacity-90 transition-opacity"
              >
                <Plus size={12} /> {t('flow.allProjects.createFirst')}
              </button>
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
                  className="mb-2 text-meta font-semibold uppercase tracking-wider"
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
                className="mb-2 text-meta font-semibold uppercase tracking-wider"
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
