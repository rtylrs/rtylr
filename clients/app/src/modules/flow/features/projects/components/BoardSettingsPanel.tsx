import React, { useMemo, useState } from 'react';
import { ChevronDown, ChevronUp, Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Switch } from '@voxire/ui';

import { useCreateColumn, useUpdateColumn, useDeleteColumn } from '../useColumns';
import GateNotice from '../../../shared/components/GateNotice';
import type { ApiProjectColumn } from '../../../shared/services/flowApi';
import type { BoardGroupKey } from '../../../shared/lib/boardUtils';
import { useDialogBehavior } from '../../../shared/lib/useDialogBehavior';

const GROUP_KEYS: BoardGroupKey[] = ['none', 'assignee', 'priority', 'due'];

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
  projectId: string;
  columns: ApiProjectColumn[];
  /** Project creator or admin. A viewer sees the panel read-only with the reason. */
  canManage: boolean;
  groupBy: BoardGroupKey;
  onGroupByChange: (next: BoardGroupKey) => void;
}

/**
 * Board settings: the columns a board has, and how it groups.
 *
 * Full-screen sheet below 768px, centred panel above it. Reordering is up/down
 * buttons plus Alt+Arrow rather than drag: the board's existing drag handling is
 * raw HTML5 events with no touch support, and buttons serve pointer, touch and
 * keyboard from one implementation.
 *
 * The panel is never hidden from a viewer. Hiding it makes the board's behaviour
 * unexplainable to the person living with it.
 */
const BoardSettingsPanel = ({
  open,
  onClose,
  projectId,
  columns,
  canManage,
  groupBy,
  onGroupByChange,
}: BoardSettingsPanelProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const { panelRef } = useDialogBehavior<HTMLDivElement>(open, onClose);
  const createColumn = useCreateColumn(projectId);
  const updateColumn = useUpdateColumn(projectId);
  const deleteColumn = useDeleteColumn(projectId);

  const [newColumnName, setNewColumnName] = useState('');
  const [pendingDelete, setPendingDelete] = useState<number | null>(null);
  const [announcement, setAnnouncement] = useState('');

  const ordered = useMemo(() => [...columns].sort((a, b) => a.sortOrder - b.sortOrder), [columns]);
  const isLastColumn = ordered.length <= 1;

  if (!open) return null;

  const renameColumn = (col: ApiProjectColumn, raw: string, restore: () => void): void => {
    const name = raw.trim();
    if (!name) {
      restore();
      return;
    }
    if (name === col.name) return;
    updateColumn.mutate({ id: String(col.id), name, color: col.color, sortOrder: col.sortOrder });
  };

  /**
   * Swaps a column with its neighbour. Two independent PUTs, because the column
   * API has no bulk reorder - if the second fails the invalidated query pulls the
   * server's real order back rather than leaving the local list lying. The name
   * override is the row's current draft so a reorder cannot write the old name
   * back over an edit that has not blurred yet.
   */
  const move = (index: number, direction: -1 | 1, nameOverride?: string): void => {
    const target = index + direction;
    if (target < 0 || target >= ordered.length) return;
    const a = ordered[index];
    const b = ordered[target];
    const aName = nameOverride === undefined ? a.name : nameOverride.trim();
    if (!aName) return;
    updateColumn.mutate({ id: String(a.id), name: aName, color: a.color, sortOrder: b.sortOrder });
    updateColumn.mutate({ id: String(b.id), name: b.name, color: b.color, sortOrder: a.sortOrder });
    setAnnouncement(
      t('flow.board.movedAnnouncement', {
        name: aName,
        position: target + 1,
        total: ordered.length,
      })
    );
  };

  const addColumn = (): void => {
    const name = newColumnName.trim();
    if (!name) return;
    const nextSortOrder = ordered.reduce((max, col) => Math.max(max, col.sortOrder), -1) + 1;
    createColumn.mutate({ name, sortOrder: nextSortOrder });
    setNewColumnName('');
  };

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

        {/* Screen-reader channel for reorder moves, which are otherwise silent. */}
        <p role="status" aria-live="polite" className="sr-only">
          {announcement}
        </p>

        <section className="mb-6">
          <h3
            className="mb-2 text-flow-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.columns')}
          </h3>

          <GateNotice className="mb-2" reason={t('flow.board.wipLimitBlocked')} />

          <ul className="space-y-1.5">
            {ordered.map((col, index) => (
              <ColumnRow
                key={col.id}
                column={col}
                index={index}
                total={ordered.length}
                canManage={canManage}
                isLastColumn={isLastColumn}
                pendingDelete={pendingDelete === col.id}
                onRename={renameColumn}
                onMove={move}
                onRequestDelete={() => setPendingDelete(col.id)}
                onCancelDelete={() => setPendingDelete(null)}
                onConfirmDelete={() => {
                  deleteColumn.mutate(String(col.id));
                  setPendingDelete(null);
                }}
              />
            ))}
          </ul>

          {isLastColumn && (
            <p className="mt-2 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.board.lastColumnProtected')}
            </p>
          )}

          {canManage && (
            <div className="mt-3 flex flex-col gap-2 sm:flex-row">
              <input
                aria-label={t('flow.board.newColumnName')}
                value={newColumnName}
                onChange={(e) => setNewColumnName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    addColumn();
                  }
                }}
                placeholder={t('flow.board.newColumnPlaceholder')}
                className="flow-focus min-h-[44px] rounded-sm px-2 text-flow-title sm:flex-1 md:min-h-0 md:py-1.5 md:text-flow-body"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              />
              <button
                type="button"
                onClick={addColumn}
                className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded-sm px-3 text-flow-body font-semibold md:min-h-0 md:py-2"
                style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
              >
                <Plus size={14} aria-hidden="true" />
                {t('flow.board.addColumn')}
              </button>
            </div>
          )}
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
            {ordered.map((col) => (
              <li key={col.id} className="flex items-center justify-between gap-3">
                <span
                  className="truncate text-flow-body"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {col.name}
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

interface ColumnRowProps {
  column: ApiProjectColumn;
  index: number;
  total: number;
  canManage: boolean;
  isLastColumn: boolean;
  pendingDelete: boolean;
  onRename: (col: ApiProjectColumn, raw: string, restore: () => void) => void;
  onMove: (index: number, direction: -1 | 1, name?: string) => void;
  onRequestDelete: () => void;
  onCancelDelete: () => void;
  onConfirmDelete: () => void;
}

const ColumnRow = ({
  column,
  index,
  total,
  canManage,
  isLastColumn,
  pendingDelete,
  onRename,
  onMove,
  onRequestDelete,
  onCancelDelete,
  onConfirmDelete,
}: ColumnRowProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [draft, setDraft] = useState(column.name);

  const moveWithDraft = (direction: -1 | 1): void => {
    const name = draft.trim();
    if (!name) {
      setDraft(column.name);
      return;
    }
    onMove(index, direction, name);
  };

  return (
    <li className="rounded-md p-2" style={{ background: 'var(--flow-hover)' }}>
      <div className="flex items-center gap-2">
        <input
          aria-label={t('flow.board.columnName')}
          value={draft}
          disabled={!canManage}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={() => onRename(column, draft, () => setDraft(column.name))}
          onKeyDown={(e) => {
            // Alt+Arrow reorders without drag, announced by the panel's live
            // region. Drag-and-drop on a touch board fights the scroll.
            if (e.altKey && (e.key === 'ArrowUp' || e.key === 'ArrowDown')) {
              e.preventDefault();
              if (canManage) moveWithDraft(e.key === 'ArrowUp' ? -1 : 1);
              return;
            }
            if (e.key === 'Enter') (e.target as HTMLInputElement).blur();
          }}
          className="flow-focus min-h-[44px] min-w-0 flex-1 rounded-sm bg-transparent px-2 text-flow-title disabled:opacity-70 md:min-h-0 md:py-1 md:text-flow-body"
          style={{ color: 'var(--flow-text)' }}
        />

        <button
          type="button"
          onClick={() => moveWithDraft(-1)}
          disabled={!canManage || index === 0}
          aria-label={t('flow.board.moveUpLabel', { name: column.name })}
          className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <ChevronUp size={14} aria-hidden="true" />
        </button>
        <button
          type="button"
          onClick={() => moveWithDraft(1)}
          disabled={!canManage || index === total - 1}
          aria-label={t('flow.board.moveDownLabel', { name: column.name })}
          className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <ChevronDown size={14} aria-hidden="true" />
        </button>
        <button
          type="button"
          onClick={onRequestDelete}
          disabled={!canManage || isLastColumn}
          aria-label={t('flow.board.removeColumnLabel', { name: column.name })}
          className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <X size={14} aria-hidden="true" />
        </button>
      </div>

      {/* WIP limit has no column field. Present and disabled with the reason
       * above the list, rather than absent - the board shows a limit it does not
       * have nowhere else either. */}
      <div className="mt-2 flex items-center gap-2">
        <label
          htmlFor={`wip-${column.id}`}
          className="text-flow-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.board.wipLimit')}
        </label>
        <input
          id={`wip-${column.id}`}
          type="number"
          disabled
          value=""
          placeholder={t('flow.board.wipLimitNone')}
          aria-label={t('flow.board.wipLimitFor', { name: column.name })}
          className="w-[80px] rounded-sm px-2 py-1 text-flow-body opacity-50"
          style={{
            background: 'var(--flow-input-bg)',
            border: '1px solid var(--flow-border-control)',
            color: 'var(--flow-text-muted)',
          }}
        />
      </div>

      {pendingDelete && (
        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
          <p className="text-flow-meta leading-snug" style={{ color: 'var(--flow-text)' }}>
            {t('flow.board.removeColumnConfirm', { name: column.name })}
          </p>
          <div className="flex gap-1.5">
            <button
              type="button"
              onClick={onConfirmDelete}
              className="flow-focus min-h-[44px] rounded-sm px-3 text-flow-meta font-semibold md:min-h-0 md:py-1"
              style={{
                background: 'var(--flow-status-rejected-bg)',
                border: '1px solid var(--flow-status-rejected-border)',
                color: 'var(--flow-status-rejected)',
              }}
            >
              {t('flow.board.confirmRemoveColumn')}
            </button>
            <button
              type="button"
              onClick={onCancelDelete}
              className="flow-focus min-h-[44px] rounded-sm px-3 text-flow-meta md:min-h-0 md:py-1"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
            >
              {t('flow.board.cancel')}
            </button>
          </div>
        </div>
      )}
    </li>
  );
};

export default BoardSettingsPanel;
