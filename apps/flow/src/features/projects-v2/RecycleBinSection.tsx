import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Archive, RotateCcw, Trash2, ChevronDown, ChevronRight } from 'lucide-react';
import { useUserStore } from '@voxire/auth';
import type { Project } from '@voxire/voxie/flow/project/v2/types';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { canPurgeProject } from './canPurgeProject';
import { useArchivedProjects } from './useAllProjects';
import { useRestoreProject, usePurgeProject } from './useBin';

export const RecycleBinSection: React.FC = () => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [pendingPurge, setPendingPurge] = useState<Project | null>(null);
  const restore = useRestoreProject();
  const purge = usePurgeProject();
  const userId = useUserStore((s) => s.userId);
  const effectiveAccess = useUserStore((s) => s.effectiveAccess);

  const { data } = useArchivedProjects(1, 10);
  const projects = data?.projects ?? [];

  if (projects.length === 0) return null;

  const confirmPurge = (): void => {
    if (!pendingPurge?.id) return;
    purge.mutate(pendingPurge.id, { onSettled: () => setPendingPurge(null) });
  };

  return (
    <div className="mt-1" role="region" aria-label={t('flow.bin.title')}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center gap-2 px-3 py-1.5 text-meta font-semibold uppercase tracking-wide rounded transition-colors hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-muted)' }}
        aria-expanded={open}
      >
        {open ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        <Archive size={13} aria-hidden />
        {t('flow.bin.title')}
        <span
          className="ml-auto rounded-full px-1.5 py-0.5 text-meta"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
          aria-label={`${projects.length} ${t('flow.bin.archived')}`}
        >
          {projects.length}
        </span>
      </button>

      {open && (
        <ul className="px-2 mt-0.5 space-y-0.5">
          {projects.map((p) => (
            <li key={p.id} className="flex items-center gap-1 group">
              <span
                className="flex items-center gap-1.5 flex-1 min-w-0 px-2 py-1 rounded text-xs truncate"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <Archive size={11} aria-hidden />
                <span className="truncate">{p.name}</span>
              </span>
              {/* The row is 28px tall by design, so each control keeps its
                  compact box and grows only its hit area: after:-inset-2 adds
                  8px on every side, giving the 44px touch target the row has no
                  room to show. */}
              <button
                type="button"
                onClick={() => restore.mutate(p.id!)}
                disabled={restore.isPending}
                className="relative h-7 w-7 rounded flex items-center justify-center opacity-0 group-hover:opacity-100 focus-visible:opacity-100 hover:bg-[var(--flow-hover)] transition-opacity after:absolute after:-inset-2 after:content-['']"
                style={{ color: 'var(--flow-text-muted)' }}
                aria-label={t('flow.bin.restore', { name: p.name })}
                title={t('flow.bin.restore', { name: p.name })}
              >
                <RotateCcw size={12} aria-hidden />
              </button>
              {canPurgeProject(p, userId, effectiveAccess) && (
                <button
                  type="button"
                  onClick={() => setPendingPurge(p)}
                  disabled={purge.isPending}
                  className="relative h-7 w-7 rounded flex items-center justify-center opacity-0 group-hover:opacity-100 focus-visible:opacity-100 hover:bg-[var(--flow-hover)] transition-opacity after:absolute after:-inset-2 after:content-['']"
                  style={{ color: 'var(--flow-danger-fg)' }}
                  aria-label={t('flow.bin.purge', { name: p.name })}
                  title={t('flow.bin.purge', { name: p.name })}
                >
                  <Trash2 size={12} aria-hidden />
                </button>
              )}
            </li>
          ))}
        </ul>
      )}

      <ConfirmDialog
        open={pendingPurge !== null}
        onOpenChange={(next) => {
          if (!next) setPendingPurge(null);
        }}
        title={t('flow.bin.purgeTitle')}
        description={t('flow.bin.purgeConfirm', { name: pendingPurge?.name ?? '' })}
        requireText={pendingPurge?.name ?? undefined}
        confirmLabel={t('flow.bin.purgeCta')}
        cancelLabel={t('common.cancel')}
        pending={purge.isPending}
        onConfirm={confirmPurge}
      />
    </div>
  );
};

export default RecycleBinSection;
