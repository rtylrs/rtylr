import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Archive, RotateCcw, Trash2 } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import type { Project } from '@voxire/voxie/flow/project/v2/types';
import ConfirmDialog from '../../../shared/components/ConfirmDialog';
import { useArchivedProjects } from '../../projects-v2/useAllProjects';
import { useRestoreProject, usePurgeProject } from '../../projects-v2/useBin';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';

const PAGE_SIZE = 50;

/**
 * The business-wide recycle bin.
 *
 * Only reachable for owner/admin/operator, which is also the condition under
 * which listArchivedProjects returns the whole business rather than just the
 * caller's own archived projects (`includeAll` in ListArchivedProjectsV2).
 *
 * There is no per-user counterpart. One was written and never reachable, behind
 * the literal `false` removed in rtylr#1212, so this is the only bin the
 * product has.
 *
 * Selection is deliberately explicit rather than a "purge everything" button:
 * the action cascades through tasks, comments and attachments with no undo, so
 * the operator has to name what goes.
 */
const BinSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [confirming, setConfirming] = useState(false);
  const restore = useRestoreProject();
  const purge = usePurgeProject();

  const archivedQuery = useArchivedProjects(1, PAGE_SIZE);
  const { data, isLoading, refetch } = archivedQuery;
  const projects: Project[] = data?.projects ?? [];
  const total = data?.total ?? 0;

  const toggle = (id: string): void => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const allSelected = projects.length > 0 && selected.size === projects.length;
  const toggleAll = (): void => {
    setSelected(
      allSelected
        ? new Set()
        : new Set(projects.map((p) => p.id).filter((id): id is string => !!id))
    );
  };

  // Sequential rather than concurrent: each purge is a cascading delete, and a
  // partial failure part-way through a batch should stop rather than race the
  // rest. Failures surface through the mutation's own error toast.
  const confirmPurge = async (): Promise<void> => {
    const remaining = new Set(selected);
    let failed = false;
    for (const id of selected) {
      try {
        await purge.mutateAsync(id);
      } catch {
        failed = true;
        break;
      }
      remaining.delete(id);
    }
    if (failed) {
      // Keep the failed project and the ones the stop left unprocessed
      // selected, and the dialog open, so a retry is one click.
      setSelected(remaining);
      return;
    }
    setSelected(new Set());
    setConfirming(false);
  };

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flow.settings.bin.section')}
        </h2>
        <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.bin.description')}
        </p>
      </div>

      {isLoading ? (
        <div className="space-y-2" aria-live="polite" aria-busy="true">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-12 rounded-lg animate-pulse"
              style={{ background: 'var(--flow-hover)' }}
            />
          ))}
        </div>
      ) : hasQueryFailed(archivedQuery) ? (
        // A failed request emptied the bin exactly as an empty bin does, so a
        // 500 told the user nothing was archived when something might be.
        <ListLoadFailed onRetry={() => void refetch()} />
      ) : projects.length === 0 ? (
        <EmptyState
          icon={Archive}
          title={t('flow.settings.bin.emptyTitle')}
          description={t('flow.settings.bin.emptyDescription')}
          action={{
            label: t('flow.settings.bin.refreshCta'),
            onClick: () => void refetch(),
          }}
          size="compact"
        />
      ) : (
        <>
          <div className="flex items-center justify-between gap-2 flex-wrap">
            <label
              className="flex items-center gap-2 min-h-[44px] text-meta cursor-pointer"
              style={{ color: 'var(--flow-text)' }}
            >
              <input
                type="checkbox"
                checked={allSelected}
                onChange={toggleAll}
                className="h-4 w-4"
              />
              {t('flow.settings.bin.selectAll', { count: projects.length })}
            </label>

            <button
              type="button"
              onClick={() => setConfirming(true)}
              disabled={selected.size === 0 || purge.isPending}
              className="min-h-[44px] px-3 rounded-lg text-meta font-semibold disabled:opacity-40"
              style={{ background: 'var(--flow-danger-fg)', color: 'var(--flow-surface)' }}
            >
              {t('flow.settings.bin.purgeSelected', { count: selected.size })}
            </button>
          </div>

          <ul className="space-y-1">
            {projects.map((p) => (
              <li
                key={p.id}
                className="flex items-center gap-2 rounded-lg border px-3 py-2"
                style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-surface)' }}
              >
                <input
                  type="checkbox"
                  checked={p.id !== undefined && selected.has(p.id)}
                  onChange={() => p.id && toggle(p.id)}
                  className="h-4 w-4"
                  aria-label={t('flow.settings.bin.selectProject', { name: p.name })}
                />
                <span
                  className="flex-1 min-w-0 truncate text-body"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {p.name}
                </span>
                {p.purgeAt ? (
                  <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.settings.bin.purgesOn', { date: p.purgeAt.slice(0, 10) })}
                  </span>
                ) : null}
                <button
                  type="button"
                  onClick={() => p.id && restore.mutate(p.id)}
                  disabled={restore.isPending}
                  className="h-11 w-11 rounded-lg flex items-center justify-center hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-muted)' }}
                  aria-label={t('flow.bin.restore', { name: p.name })}
                >
                  <RotateCcw size={14} aria-hidden />
                </button>
              </li>
            ))}
          </ul>

          {total > projects.length ? (
            <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.settings.bin.truncated', { shown: projects.length, total })}
            </p>
          ) : null}
        </>
      )}

      <ConfirmDialog
        open={confirming}
        onOpenChange={setConfirming}
        title={t('flow.settings.bin.purgeTitle')}
        description={t('flow.settings.bin.purgeConfirm', { count: selected.size })}
        requireText={t('flow.settings.bin.purgeKeyword')}
        confirmLabel={t('flow.bin.purgeCta')}
        cancelLabel={t('common.cancel')}
        pending={purge.isPending}
        onConfirm={() => void confirmPurge()}
      />

      <p
        className="flex items-center gap-1.5 text-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <Trash2 size={11} aria-hidden />
        {t('flow.settings.bin.retentionNote')}
      </p>
    </div>
  );
};

export default BinSection;
