import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Ban } from 'lucide-react';
import { Tooltip } from '../../../../shared/components/ui/tooltip';
import { useTaskRelations } from '../../features/tasks/useTaskRelations';
import { withAlpha } from '../lib/colorUtils';

interface BlockedBadgeProps {
  taskId: string;
  /** Open blockers, from the task listing. */
  count: number;
}

/**
 * Marks a task that another open task blocks, on board cards and list rows.
 * The tooltip names the blockers. They are fetched the first time someone
 * points at the badge, not for every card on the board.
 */
const BlockedBadge: React.FC<BlockedBadgeProps> = ({ taskId, count }) => {
  const { t } = useTranslation();
  const [wanted, setWanted] = useState(false);
  const relations = useTaskRelations(taskId, wanted);
  const blockers =
    relations.data
      ?.find((group) => group.relation === 'blocked_by')
      ?.tasks.filter((task) => task.statusBucket !== 'done' && task.statusBucket !== 'cancelled') ??
    [];

  const content = (
    <span className="block">
      <span className="block font-semibold">{t('flow.blockedBadge.tooltipTitle', { count })}</span>
      {relations.isPending ? (
        <span className="block opacity-80">{t('flow.blockedBadge.loading')}</span>
      ) : (
        blockers.map((blocker) => (
          <span key={blocker.relationId} className="block truncate">
            {blocker.identifier} · {blocker.title}
          </span>
        ))
      )}
    </span>
  );

  return (
    <span
      className="inline-flex shrink-0"
      onMouseEnter={() => setWanted(true)}
      onFocus={() => setWanted(true)}
      data-testid="blocked-badge"
    >
      <Tooltip content={content}>
        <span
          className="inline-flex items-center gap-0.5 rounded-sm px-1.5 py-0.5 text-flow-meta font-semibold"
          style={{
            background: withAlpha('var(--flow-danger-fg)', 13),
            color: 'var(--flow-danger-fg)',
          }}
          aria-label={t('flow.blockedBadge.label', { count })}
        >
          <Ban size={10} aria-hidden="true" />
          {t('flow.blockedBadge.text')}
        </span>
      </Tooltip>
    </span>
  );
};

export default BlockedBadge;
