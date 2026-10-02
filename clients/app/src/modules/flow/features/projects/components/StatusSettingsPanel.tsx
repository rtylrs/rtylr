import React, { useMemo, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { ChevronDown, ChevronUp, Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import GateNotice from '../../../shared/components/GateNotice';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import {
  useProjectStatuses,
  useCreateStatus,
  useUpdateStatus,
  useDeleteStatus,
} from '../useStatuses';
import type { ApiStatus } from '../../../shared/services/flowApi';
import { useDialogBehavior } from '../../../shared/lib/useDialogBehavior';

const STATUS_COLORS = [
  '#ef4444',
  '#f97316',
  '#eab308',
  '#22c55e',
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
  '#64C5B9',
];

const BUCKETS = ['backlog', 'active', 'done', 'cancelled'] as const;
type Bucket = (typeof BUCKETS)[number];

interface StatusSettingsPanelProps {
  open: boolean;
  onClose: () => void;
  projectId: string;
  /** Project creator or admin. Matches BoardSettingsPanel's gate: a viewer sees
   * the set read-only rather than the panel being hidden, per F-A's "config is
   * admin, task mutation is member" split. */
  canManage: boolean;
}

/**
 * Per-project custom status configuration (F-A). Reorder is up/down buttons
 * plus Alt+Arrow, matching BoardSettingsPanel's columns list rather than
 * drag - same reason: the board's drag handling has no touch support, and
 * buttons serve pointer, touch and keyboard from one implementation.
 *
 * An empty set is the normal, unconfigured case, not an error state: the
 * project falls back to the six built-in string statuses. The empty state
 * says this explicitly rather than reading as broken.
 */
const StatusSettingsPanel = ({
  open,
  onClose,
  projectId,
  canManage,
}: StatusSettingsPanelProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const { panelRef } = useDialogBehavior<HTMLDivElement>(open, onClose);
  const statusesQuery = useProjectStatuses(projectId);
  const {
    data: statuses = [],
    isPending: statusesPending,
    refetch: refetchStatuses,
  } = statusesQuery;
  const statusesLoadFailed = hasQueryFailed(statusesQuery);
  const statusesReady = !statusesPending && !statusesLoadFailed && statusesQuery.data !== undefined;
  const queryClient = useQueryClient();
  const createStatus = useCreateStatus(projectId);
  const updateStatus = useUpdateStatus(projectId);
  const deleteStatus = useDeleteStatus(projectId);

  const invalidateStatusViews = (): void => {
    queryClient.invalidateQueries({ queryKey: ['tasks'] });
    queryClient.invalidateQueries({ queryKey: ['my-tasks'] });
    queryClient.invalidateQueries({ queryKey: ['task'] });
  };

  const [newName, setNewName] = useState('');
  const [newBucket, setNewBucket] = useState<Bucket>('backlog');
  const [newColor, setNewColor] = useState(STATUS_COLORS[0]);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);
  const [announcement, setAnnouncement] = useState('');

  const ordered = useMemo(
    () => [...statuses].sort((a, b) => a.sortOrder - b.sortOrder),
    [statuses]
  );

  if (!open) return null;

  const rename = (status: ApiStatus, raw: string, restore: () => void): void => {
    const name = raw.trim();
    if (!name) {
      restore();
      return;
    }
    if (name === status.name) return;
    updateStatus.mutate(
      {
        id: status.id,
        name,
        bucket: status.bucket,
        color: status.color,
        sortOrder: status.sortOrder,
        isDefault: status.isDefault,
      },
      { onSuccess: invalidateStatusViews }
    );
  };

  const recolor = (status: ApiStatus, color: string, nameOverride?: string): void => {
    const name = nameOverride === undefined ? status.name : nameOverride.trim();
    if (!name) return;
    updateStatus.mutate(
      {
        id: status.id,
        name,
        bucket: status.bucket,
        color,
        sortOrder: status.sortOrder,
        isDefault: status.isDefault,
      },
      { onSuccess: invalidateStatusViews }
    );
  };

  const rebucket = (status: ApiStatus, bucket: Bucket, nameOverride?: string): void => {
    const name = nameOverride === undefined ? status.name : nameOverride.trim();
    if (!name) return;
    updateStatus.mutate(
      {
        id: status.id,
        name,
        bucket,
        color: status.color,
        sortOrder: status.sortOrder,
        isDefault: status.isDefault,
      },
      { onSuccess: invalidateStatusViews }
    );
  };

  const makeDefault = (status: ApiStatus, nameOverride?: string): void => {
    const name = nameOverride === undefined ? status.name : nameOverride.trim();
    if (!name) return;
    updateStatus.mutate(
      {
        id: status.id,
        name,
        bucket: status.bucket,
        color: status.color,
        sortOrder: status.sortOrder,
        isDefault: true,
      },
      { onSuccess: invalidateStatusViews }
    );
  };

  /** Two independent PUTs, no bulk reorder on the contract - same pattern as
   * BoardSettingsPanel.move: if the second fails, the invalidated query pulls
   * the server's real order back rather than leaving the local list lying. The
   * current row draft is carried into the first replacement update. */
  const move = (index: number, direction: -1 | 1, nameOverride?: string): void => {
    const target = index + direction;
    if (target < 0 || target >= ordered.length) return;
    const a = ordered[index];
    const b = ordered[target];
    const aName = nameOverride === undefined ? a.name : nameOverride.trim();
    if (!aName) return;
    updateStatus.mutate(
      {
        id: a.id,
        name: aName,
        bucket: a.bucket,
        color: a.color,
        sortOrder: b.sortOrder,
        isDefault: a.isDefault,
      },
      { onSuccess: invalidateStatusViews }
    );
    updateStatus.mutate(
      {
        id: b.id,
        name: b.name,
        bucket: b.bucket,
        color: b.color,
        sortOrder: a.sortOrder,
        isDefault: b.isDefault,
      },
      { onSuccess: invalidateStatusViews }
    );
    setAnnouncement(
      t('flow.status.movedAnnouncement', {
        name: aName,
        position: target + 1,
        total: ordered.length,
      })
    );
  };

  const addStatus = (): void => {
    const name = newName.trim();
    if (!name) {
      toast.error(t('flowErrors.generic'));
      return;
    }
    const nextSortOrder = ordered.reduce((max, status) => Math.max(max, status.sortOrder), -1) + 1;
    createStatus.mutate(
      { name, bucket: newBucket, color: newColor, sortOrder: nextSortOrder },
      {
        onSuccess: () => {
          setNewName('');
          setNewBucket('backlog');
          setNewColor(STATUS_COLORS[0]);
          toast.success(t('flowToasts.status.created'));
        },
      }
    );
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t('flow.status.settingsTitle')}
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
        className="max-h-[calc(100dvh-64px)] w-full overflow-y-auto rounded-lg p-4 md:max-w-[560px]"
        style={{
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <h2 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
              {t('flow.status.settingsTitle')}
            </h2>
            <p
              className="mt-1 text-flow-meta leading-snug"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.status.settingsSubtitle')}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label={t('flow.status.closeSettings')}
            className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm md:h-8 md:w-8"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <X size={16} aria-hidden="true" />
          </button>
        </div>

        {!canManage && <GateNotice className="mb-4" reason={t('flow.status.readOnlyReason')} />}

        <p role="status" aria-live="polite" className="sr-only">
          {announcement}
        </p>

        {statusesLoadFailed ? (
          <div className="mb-4">
            <ListLoadFailed onRetry={() => void refetchStatuses()} />
          </div>
        ) : statusesPending || statusesQuery.data === undefined ? (
          <p className="mb-4 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.shell.loading')}
          </p>
        ) : (
          <>
            {ordered.length === 0 && (
              <p
                className="mb-4 text-flow-meta leading-snug"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.status.emptyExplainer')}
              </p>
            )}

            <ul className="space-y-1.5">
              {ordered.map((status, index) => (
                <StatusRow
                  key={status.id}
                  status={status}
                  index={index}
                  total={ordered.length}
                  canManage={canManage}
                  pendingDelete={pendingDelete === status.id}
                  onRename={rename}
                  onRecolor={recolor}
                  onRebucket={rebucket}
                  onMakeDefault={makeDefault}
                  onMove={move}
                  onRequestDelete={() => setPendingDelete(status.id)}
                  onCancelDelete={() => setPendingDelete(null)}
                  onConfirmDelete={() => {
                    deleteStatus.mutate(status.id, { onSuccess: invalidateStatusViews });
                    setPendingDelete(null);
                  }}
                />
              ))}
            </ul>
          </>
        )}

        {canManage && statusesReady && (
          <div
            className="mt-4 space-y-2 rounded-md p-3"
            style={{ background: 'var(--flow-hover)' }}
          >
            <div className="flex flex-col gap-2 sm:flex-row">
              <input
                aria-label={t('flow.status.newStatusName')}
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    addStatus();
                  }
                }}
                placeholder={t('flow.status.newStatusPlaceholder')}
                className="flow-focus min-h-[44px] min-w-0 flex-1 rounded-sm px-2 text-flow-title md:min-h-0 md:py-1.5 md:text-flow-body"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              />
              <select
                aria-label={t('flow.status.bucketLabel')}
                value={newBucket}
                onChange={(e) => setNewBucket(e.target.value as Bucket)}
                className="flow-focus min-h-[44px] rounded-sm px-2 text-flow-title md:min-h-0 md:py-1.5 md:text-flow-body"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              >
                {BUCKETS.map((b) => (
                  <option key={b} value={b}>
                    {t(`flow.status.bucket.${b}`)}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {STATUS_COLORS.map((c) => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setNewColor(c)}
                  aria-label={t('flow.status.chooseColorLabel', { color: c })}
                  aria-pressed={newColor === c}
                  className="flow-focus h-6 w-6 shrink-0 rounded-full transition-all"
                  style={{
                    background: c,
                    outline: newColor === c ? `2px solid ${c}` : 'none',
                    outlineOffset: 2,
                    boxShadow: newColor === c ? '0 0 0 2px var(--flow-bg)' : 'none',
                  }}
                  title={c}
                />
              ))}
              <button
                type="button"
                disabled={!newName.trim() || createStatus.isPending}
                onClick={addStatus}
                className="flow-focus ms-auto inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded-sm px-3 text-flow-body font-semibold disabled:opacity-40 md:min-h-0 md:py-2"
                style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
              >
                <Plus size={14} aria-hidden="true" />
                {t('flow.status.addStatus')}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

interface StatusRowProps {
  status: ApiStatus;
  index: number;
  total: number;
  canManage: boolean;
  pendingDelete: boolean;
  onRename: (status: ApiStatus, raw: string, restore: () => void) => void;
  onRecolor: (status: ApiStatus, color: string, name?: string) => void;
  onRebucket: (status: ApiStatus, bucket: Bucket, name?: string) => void;
  onMakeDefault: (status: ApiStatus, name?: string) => void;
  onMove: (index: number, direction: -1 | 1, name?: string) => void;
  onRequestDelete: () => void;
  onCancelDelete: () => void;
  onConfirmDelete: () => void;
}

const StatusRow = ({
  status,
  index,
  total,
  canManage,
  pendingDelete,
  onRename,
  onRecolor,
  onRebucket,
  onMakeDefault,
  onMove,
  onRequestDelete,
  onCancelDelete,
  onConfirmDelete,
}: StatusRowProps): React.JSX.Element => {
  const { t } = useTranslation();
  const [draft, setDraft] = useState(status.name);

  const currentDraftName = (): string | null => {
    const name = draft.trim();
    if (!name) {
      setDraft(status.name);
      return null;
    }
    return name;
  };

  const moveWithDraft = (direction: -1 | 1): void => {
    const name = currentDraftName();
    if (name) onMove(index, direction, name);
  };

  return (
    <li className="rounded-md p-2" style={{ background: 'var(--flow-hover)' }}>
      <div className="flex items-center gap-2">
        <span
          aria-hidden="true"
          className="h-2.5 w-2.5 shrink-0 rounded-full"
          style={{ background: status.color }}
        />
        <input
          aria-label={t('flow.status.statusName')}
          value={draft}
          disabled={!canManage}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={() => onRename(status, draft, () => setDraft(status.name))}
          onKeyDown={(e) => {
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
        <select
          aria-label={t('flow.status.bucketFor', { name: status.name })}
          value={status.bucket}
          disabled={!canManage}
          onChange={(e) => {
            const name = currentDraftName();
            if (name) onRebucket(status, e.target.value as Bucket, name);
          }}
          className="flow-focus min-h-[44px] rounded-sm px-1.5 text-flow-body disabled:opacity-70 md:min-h-0 md:py-1"
          style={{
            background: 'var(--flow-input-bg)',
            border: '1px solid var(--flow-border-control)',
            color: 'var(--flow-text)',
          }}
        >
          {BUCKETS.map((b) => (
            <option key={b} value={b}>
              {t(`flow.status.bucket.${b}`)}
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={() => moveWithDraft(-1)}
          disabled={!canManage || index === 0}
          aria-label={t('flow.status.moveUpLabel', { name: status.name })}
          className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <ChevronUp size={14} aria-hidden="true" />
        </button>
        <button
          type="button"
          onClick={() => moveWithDraft(1)}
          disabled={!canManage || index === total - 1}
          aria-label={t('flow.status.moveDownLabel', { name: status.name })}
          className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <ChevronDown size={14} aria-hidden="true" />
        </button>
        <button
          type="button"
          onClick={onRequestDelete}
          disabled={!canManage}
          aria-label={t('flow.status.removeStatusLabel', { name: status.name })}
          className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <X size={14} aria-hidden="true" />
        </button>
      </div>

      <div className="mt-2 flex items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-1.5">
          {[
            '#ef4444',
            '#f97316',
            '#eab308',
            '#22c55e',
            '#3b82f6',
            '#8b5cf6',
            '#ec4899',
            '#64C5B9',
          ].map((c) => (
            <button
              key={c}
              type="button"
              disabled={!canManage}
              onClick={() => {
                const name = currentDraftName();
                if (name) onRecolor(status, c, name);
              }}
              aria-label={t('flow.status.chooseColorLabel', { color: c })}
              aria-pressed={status.color === c}
              className="flow-focus h-5 w-5 shrink-0 rounded-full disabled:opacity-40"
              style={{
                background: c,
                outline: status.color === c ? `2px solid ${c}` : 'none',
                outlineOffset: 2,
                boxShadow: status.color === c ? '0 0 0 2px var(--flow-bg)' : 'none',
              }}
              title={c}
            />
          ))}
        </div>
        <label
          className="flex items-center gap-1.5 text-flow-meta"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <input
            type="radio"
            name="flow-status-default"
            checked={status.isDefault}
            disabled={!canManage}
            onChange={() => {
              const name = currentDraftName();
              if (name) onMakeDefault(status, name);
            }}
            aria-label={t('flow.status.isDefaultFor', { name: status.name })}
          />
          {t('flow.status.isDefault')}
        </label>
      </div>

      {pendingDelete && (
        <div className="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center">
          <p className="text-flow-meta leading-snug" style={{ color: 'var(--flow-text)' }}>
            {t('flow.status.removeStatusConfirm', { name: status.name })}
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
              {t('flow.status.confirmRemoveStatus')}
            </button>
            <button
              type="button"
              onClick={onCancelDelete}
              className="flow-focus min-h-[44px] rounded-sm px-3 text-flow-meta md:min-h-0 md:py-1"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
            >
              {t('flow.status.cancel')}
            </button>
          </div>
        </div>
      )}
    </li>
  );
};

export default StatusSettingsPanel;
