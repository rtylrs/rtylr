import React from 'react';
import { ListChecks, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Switch } from '@voxire/ui';

import GateNotice from '../../../shared/components/GateNotice';
import type { BoardGroupKey } from '../../../shared/lib/boardUtils';
import { useDialogBehavior } from '../../../shared/lib/useDialogBehavior';
import { displayStatusName } from '../statusUtils';

const GROUP_KEYS: BoardGroupKey[] = ['none', 'assignee', 'priority', 'due', 'epic'];

/**
 * Card fields the design puts on a board card. None of them can be stored:
 * neither the column nor the project carries a card-field set, so these render
 * disabled with the reason rather than writing a team setting to localStorage,
 * which would make a shared setting silently personal.
 */
const CARD_FIELDS = ['loggedTime', 'estimate', 'overEstimate', 'assignee'] as const;

interface BoardSettingsPanelProps {
  open: boolean;
  onClose: () => void;
  /** The board's columns, which are the project's statuses, in order. */
  columns: readonly { id: string; name: string }[];
  /** Project creator or admin. A viewer sees the panel read-only with the reason. */
  canManage: boolean;
  groupBy: BoardGroupKey;
  onGroupByChange: (next: BoardGroupKey) => void;
  /** Columns are statuses, edited in the statuses panel. */
  onManageStatuses: () => void;
}

/**
 * Board settings: what the board shows and how it groups. Columns are the
 * project's statuses, so they are managed in the statuses panel this links to.
 *
 * The panel is never hidden from a viewer. Hiding it makes the board's behaviour
 * unexplainable to the person living with it.
 */
const BoardSettingsPanel = ({
  open,
  onClose,
  columns,
  canManage,
  groupBy,
  onGroupByChange,
  onManageStatuses,
}: BoardSettingsPanelProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const { panelRef } = useDialogBehavior<HTMLDivElement>(open, onClose);

  if (!open) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t('flow.board.settingsTitle')}
      className="fixed inset-0 z-50 flex flex-col overflow-y-auto p-4 md:items-center md:justify-center"
      style={{ background: 'rgba(0,0,0,.5)' }}
    >
      <div
        ref={panelRef}
        tabIndex={-1}
        // A panel taller than the viewport was centred by the overlay, which
        // pushed its top off the screen: the measured case was 1242px in a
        // 900px viewport with its own title at -171px. Bounding the height and
        // scrolling the panel keeps the title reachable at any viewport.
        className="max-h-[calc(100dvh-64px)] w-full overflow-y-auto rounded-lg p-4 md:max-w-[520px]"
        style={{
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        <div className="mb-4 flex items-start justify-between gap-3">
          <h2 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.board.settingsTitle')}
          </h2>
          <button
            type="button"
            onClick={onClose}
            aria-label={t('flow.board.closeSettings')}
            className="flow-focus flex h-[44px] w-[44px] items-center justify-center rounded-sm md:h-8 md:w-8"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <X size={16} aria-hidden="true" />
          </button>
        </div>

        {!canManage && <GateNotice className="mb-4" reason={t('flow.board.readOnlyReason')} />}

        <section className="mb-6">
          <h3
            className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.columns')}
          </h3>
          <p
            className="mb-2 text-flow-meta leading-snug"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.columnsAreStatuses')}
          </p>
          <button
            type="button"
            onClick={onManageStatuses}
            className="flow-focus inline-flex min-h-[44px] items-center gap-1.5 rounded-sm px-3 text-flow-body font-semibold md:min-h-0 md:py-1.5"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
          >
            <ListChecks size={14} aria-hidden="true" />
            {t('flow.status.openSettings')}
          </button>
        </section>

        <section className="mb-6">
          <h3
            className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.promptsForHours')}
          </h3>
          <GateNotice className="mb-2" reason={t('flow.board.promptsForHoursBlocked')} />
          <ul className="space-y-1.5">
            {columns.map((col) => (
              <li key={col.id} className="flex items-center justify-between gap-3">
                <span
                  className="truncate text-flow-body"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {displayStatusName(col.name, t)}
                </span>
                <Switch
                  checked={false}
                  disabled
                  aria-label={t('flow.board.promptsForHoursFor', { name: col.name })}
                  className="shrink-0"
                />
              </li>
            ))}
          </ul>
        </section>

        <section className="mb-6">
          <h3
            className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.onTheCard')}
          </h3>
          <GateNotice className="mb-2" reason={t('flow.board.cardFieldsBlocked')} />
          <ul className="space-y-1.5">
            {CARD_FIELDS.map((field) => (
              <li key={field} className="flex items-center justify-between gap-3">
                <span className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
                  {t(`flow.board.cardField.${field}`)}
                </span>
                <Switch
                  checked={false}
                  disabled
                  aria-label={t(`flow.board.cardField.${field}`)}
                  className="shrink-0"
                />
              </li>
            ))}
          </ul>
        </section>

        <section>
          <h3
            className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.groupByHeading')}
          </h3>
          {/* Group-by is the one setting here that is a personal preference. The
           * copy says which is which, because one changes only your view and the
           * other changes what your team sees. */}
          <p
            className="mb-2 text-flow-meta leading-snug"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.scopeExplained')}
          </p>
          <div role="radiogroup" aria-label={t('flow.board.groupByHeading')} className="space-y-1">
            {GROUP_KEYS.map((key) => (
              <button
                key={key}
                type="button"
                role="radio"
                aria-checked={groupBy === key}
                onClick={() => onGroupByChange(key)}
                className="flow-focus flex min-h-[44px] w-full items-center gap-2 rounded-sm px-2 text-start text-flow-body md:min-h-0 md:py-1.5"
                style={{
                  background: groupBy === key ? 'var(--flow-active)' : 'transparent',
                  color: groupBy === key ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                }}
              >
                <span
                  className="h-1.5 w-1.5 shrink-0 rounded-full"
                  style={{ background: groupBy === key ? 'var(--accent)' : 'transparent' }}
                />
                {t(`flow.board.groupBy.${key}`)}
              </button>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
};

export default BoardSettingsPanel;
