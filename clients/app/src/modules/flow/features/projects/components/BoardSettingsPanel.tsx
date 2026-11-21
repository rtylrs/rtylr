import React from 'react';
import { ListChecks, RotateCcw, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Switch } from '@voxire/ui';

import GateNotice from '../../../shared/components/GateNotice';
import { useDialogBehavior } from '../../../shared/lib/useDialogBehavior';
import { displayStatusName } from '../statusUtils';
import {
  BOARD_GROUP_KEYS,
  CARD_FIELDS,
  COLUMN_WIDTHS,
  HIDE_DONE_CHOICES,
  type BoardViewSettings,
  type HideDoneAfter,
} from '../boardViewSettings';

interface BoardSettingsPanelProps {
  open: boolean;
  onClose: () => void;
  /** The board's columns, which are the project's statuses, in order. */
  columns: readonly { id: string; name: string }[];
  /** Project creator or admin. A viewer sees the board setup read-only with the reason. */
  canManage: boolean;
  /** The viewer's own settings for this board. */
  settings: BoardViewSettings;
  onSettingsChange: (patch: Partial<BoardViewSettings>) => void;
  onResetSettings: () => void;
  /** Columns are statuses, edited in the statuses panel. */
  onManageStatuses: () => void;
}

const hideDoneKey = (days: HideDoneAfter): string =>
  days === null ? 'flow.board.hideDone.never' : `flow.board.hideDone.days${days}`;

const sectionHeading = 'mb-2 text-flow-meta font-semibold uppercase tracking-wider';

/** A labelled set of radio buttons in the panel's one style. */
function Choices<T extends string | number | null>({
  label,
  options,
  value,
  optionLabel,
  onChange,
}: {
  label: string;
  options: readonly T[];
  value: T;
  optionLabel: (option: T) => string;
  onChange: (next: T) => void;
}): React.JSX.Element {
  return (
    <div role="radiogroup" aria-label={label} className="space-y-1">
      {options.map((option) => {
        const checked = option === value;
        return (
          <button
            key={String(option)}
            type="button"
            role="radio"
            aria-checked={checked}
            onClick={() => onChange(option)}
            className="flow-focus flex min-h-[44px] w-full items-center gap-2 rounded-sm px-2 text-start text-flow-body md:min-h-0 md:py-1.5"
            style={{
              background: checked ? 'var(--flow-active)' : 'transparent',
              color: checked ? 'var(--flow-text)' : 'var(--flow-text-muted)',
            }}
          >
            <span
              className="h-1.5 w-1.5 shrink-0 rounded-full"
              style={{ background: checked ? 'var(--accent)' : 'transparent' }}
            />
            {optionLabel(option)}
          </button>
        );
      })}
    </div>
  );
}

/**
 * Board settings, in a side panel. Two kinds live here and the panel keeps them
 * apart: "Your view" (grouping, card fields, how long done cards stay, column
 * width) is the viewer's own and changes nobody else's board; "Board setup"
 * (columns, which are statuses) is the team's and needs manage project.
 *
 * The panel is never hidden from a viewer. Hiding it makes the board's behaviour
 * unexplainable to the person living with it, and their own view is theirs to set.
 */
const BoardSettingsPanel = ({
  open,
  onClose,
  columns,
  canManage,
  settings,
  onSettingsChange,
  onResetSettings,
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
      // A side panel on the inline end, so the board stays in view beside it
      // while cards change. flex-row follows the document direction, so in
      // Arabic it opens from the left.
      className="fixed inset-0 z-50 flex justify-end"
      style={{ background: 'rgba(0,0,0,.35)' }}
    >
      <div
        ref={panelRef}
        tabIndex={-1}
        // Bounded to the viewport and scrolling inside, so the title stays
        // reachable however long the panel grows.
        className="h-full max-h-dvh w-full overflow-y-auto p-4 md:max-w-[400px]"
        style={{
          background: 'var(--flow-card)',
          borderInlineStart: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        <div className="mb-3 flex items-start justify-between gap-3">
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

        <h3 className="text-flow-body font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.board.yourView')}
        </h3>
        <p className="mb-4 text-flow-meta leading-snug" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.board.scopeExplained')}
        </p>

        <section className="mb-6">
          <h4 className={sectionHeading} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.board.groupByHeading')}
          </h4>
          <Choices
            label={t('flow.board.groupByHeading')}
            options={BOARD_GROUP_KEYS}
            value={settings.groupBy}
            optionLabel={(key) => t(`flow.board.groupBy.${key}`)}
            onChange={(groupBy) => onSettingsChange({ groupBy })}
          />
        </section>

        <section className="mb-6">
          <h4 className={sectionHeading} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.board.onTheCard')}
          </h4>
          <ul className="space-y-1.5">
            {CARD_FIELDS.map((field) => (
              <li key={field} className="flex items-center justify-between gap-3">
                <span className="text-flow-body" style={{ color: 'var(--flow-text)' }}>
                  {t(`flow.board.cardField.${field}`)}
                </span>
                <Switch
                  checked={settings.cardFields[field]}
                  onCheckedChange={(on: boolean) =>
                    onSettingsChange({ cardFields: { ...settings.cardFields, [field]: on } })
                  }
                  aria-label={t(`flow.board.cardField.${field}`)}
                  className="shrink-0"
                />
              </li>
            ))}
          </ul>
        </section>

        <section className="mb-6">
          <h4 className={sectionHeading} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.board.hideDoneHeading')}
          </h4>
          <p
            className="mb-2 text-flow-meta leading-snug"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.hideDoneExplained')}
          </p>
          <Choices
            label={t('flow.board.hideDoneHeading')}
            options={HIDE_DONE_CHOICES}
            value={settings.hideDoneAfterDays}
            optionLabel={(days) => t(hideDoneKey(days))}
            onChange={(hideDoneAfterDays) => onSettingsChange({ hideDoneAfterDays })}
          />
        </section>

        <section className="mb-4">
          <h4 className={sectionHeading} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.board.columnWidthHeading')}
          </h4>
          <Choices
            label={t('flow.board.columnWidthHeading')}
            options={COLUMN_WIDTHS}
            value={settings.columnWidth}
            optionLabel={(width) => t(`flow.board.columnWidth.${width}`)}
            onChange={(columnWidth) => onSettingsChange({ columnWidth })}
          />
        </section>

        <button
          type="button"
          onClick={onResetSettings}
          className="flow-focus mb-8 inline-flex min-h-[44px] items-center gap-1.5 rounded-sm px-2 text-flow-meta font-medium hover:bg-[var(--flow-hover)] md:min-h-0 md:py-1.5"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <RotateCcw size={12} aria-hidden="true" />
          {t('flow.board.resetView')}
        </button>

        <h3
          className="mb-3 border-t pt-4 text-flow-body font-semibold"
          style={{ color: 'var(--flow-text)', borderColor: 'var(--flow-border)' }}
        >
          {t('flow.board.boardSetup')}
        </h3>

        {!canManage && <GateNotice className="mb-4" reason={t('flow.board.readOnlyReason')} />}

        <section className="mb-6">
          <h4 className={sectionHeading} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.board.columns')}
          </h4>
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

        <section>
          <h4 className={sectionHeading} style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.board.promptsForHours')}
          </h4>
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
      </div>
    </div>
  );
};

export default BoardSettingsPanel;
