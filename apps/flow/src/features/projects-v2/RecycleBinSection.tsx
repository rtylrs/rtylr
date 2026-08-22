import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Archive, RotateCcw, Trash2, ChevronDown, ChevronRight } from 'lucide-react';
import { useArchivedProjects } from './useAllProjects';
import { useRestoreProject } from './useBin';

export const RecycleBinSection: React.FC = () => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const restore = useRestoreProject();

  const { data } = useArchivedProjects(1, 10);
  const projects = data?.projects ?? [];

  if (projects.length === 0) return null;

  return (
    <div className="mt-1" role="region" aria-label={t('flow.bin.title')}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center gap-2 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wide rounded transition-colors hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-muted)' }}
        aria-expanded={open}
      >
        {open ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        <Archive size={13} aria-hidden />
        {t('flow.bin.title')}
        <span
          className="ml-auto rounded-full px-1.5 py-0.5 text-[10px]"
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
              <button
                type="button"
                onClick={() => restore.mutate(p.id!)}
                disabled={restore.isPending}
                className="h-7 w-7 rounded flex items-center justify-center opacity-0 group-hover:opacity-100 focus-visible:opacity-100 hover:bg-[var(--flow-hover)] transition-opacity"
                style={{ color: 'var(--flow-text-muted)' }}
                aria-label={t('flow.bin.restore', { name: p.name })}
                title={t('flow.bin.restore', { name: p.name })}
              >
                <RotateCcw size={12} aria-hidden />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default RecycleBinSection;
