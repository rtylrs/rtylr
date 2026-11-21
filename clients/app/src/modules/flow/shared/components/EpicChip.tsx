import React, { createContext, useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { Layers } from 'lucide-react';
import type { Task } from '../stores/flowStore';
import { withAlpha, onTint } from '../lib/colorUtils';
import { cn } from '../lib/utils';
import { isEpic } from '../lib/taskHierarchy';

/**
 * Narrows the current view to one epic. A project view that owns a filter
 * (board, list) provides it; anywhere else the chip is a plain label, since
 * there is no filter for it to drive.
 */
export const EpicFilterContext = createContext<((epicId: string) => void) | null>(null);

interface EpicChipProps {
  task: Pick<Task, 'id' | 'title' | 'taskType' | 'epicId' | 'epicTitle' | 'epicColor'>;
  className?: string;
}

/**
 * The epic a task sits under, on its card and its list row. An epic itself
 * shows an "Epic" badge in its own colour instead, which filters to the epic
 * the same way. Renders nothing for a task with no epic.
 */
export const EpicChip: React.FC<EpicChipProps> = ({ task, className }) => {
  const { t } = useTranslation();
  const filterByEpic = useContext(EpicFilterContext);
  const ownEpic = isEpic(task);
  if (!ownEpic && (!task.epicId || !task.epicTitle)) return null;
  const epicId = ownEpic ? task.id : (task.epicId ?? '');
  const name = ownEpic ? t('flow.taskType.epic') : (task.epicTitle ?? '');
  const epicName = ownEpic ? task.title : name;
  const tone = task.epicColor;
  const style: React.CSSProperties = {
    ...(tone ? { ['--tone' as string]: tone } : {}),
    background: tone ? withAlpha('var(--tone)', 13) : 'var(--accent-muted)',
    color: tone ? onTint('var(--tone)') : 'var(--accent-on-muted)',
  };
  const classes = cn(
    'inline-flex min-w-0 max-w-[10rem] shrink items-center gap-1 rounded-sm px-1.5 py-0.5 text-flow-meta font-semibold',
    className
  );
  const content = (
    <>
      <Layers size={10} aria-hidden="true" className="shrink-0" />
      <span className="truncate">{name}</span>
    </>
  );

  if (!filterByEpic) {
    return (
      <span className={classes} style={style} title={name}>
        {content}
      </span>
    );
  }
  return (
    <button
      type="button"
      className={cn('flow-focus hover:opacity-80', classes)}
      style={style}
      title={t('flow.epics.filterBy', { name: epicName })}
      aria-label={t('flow.epics.filterBy', { name: epicName })}
      onClick={(e) => {
        e.stopPropagation();
        filterByEpic(epicId);
      }}
      onKeyDown={(e) => e.stopPropagation()}
    >
      {content}
    </button>
  );
};

export default EpicChip;
