import React, { useEffect, useState } from 'react';
import { ChevronDown, ChevronUp, Plus, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import GateNotice from '../../../shared/components/GateNotice';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import { canFlowPermission } from '../../../shared/lib/roleUtils';
import type { ApiStatusTemplateItem } from '../../../shared/services/flowApi';
import { useUserStore } from '../../../../../shared/stores/userStore';
import { useSaveStatusTemplate, useStatusTemplate } from '../../projects/useStatuses';
import { STATUS_BUCKETS, STATUS_COLORS, type StatusBucket } from '../../projects/statusUtils';

/** The server caps a template at this many statuses. */
const MAX_STATUSES = 30;

interface DraftItem extends ApiStatusTemplateItem {
  /** Local identity, so rows keep focus and state while being reordered. */
  key: number;
}

let nextKey = 0;

const withKeys = (items: readonly ApiStatusTemplateItem[]): DraftItem[] =>
  items.map((item) => ({ ...item, key: nextKey++ }));

const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};

/** Why the draft cannot be saved, as a translation key, or null when it can. */
export function templateProblem(items: readonly ApiStatusTemplateItem[]): string | null {
  if (items.length === 0) return 'flow.settings.statuses.errorEmpty';
  if (items.some((item) => !item.name.trim())) return 'flow.settings.statuses.errorName';
  const names = items.map((item) => item.name.trim().toLowerCase());
  if (new Set(names).size !== names.length) return 'flow.settings.statuses.errorDuplicate';
  if (!items.some((item) => item.bucket === 'done')) return 'flow.settings.statuses.errorNoDone';
  return null;
}

/**
 * The workspace's status template: the set every new project starts with.
 * Existing projects keep their own statuses. Editing needs team management;
 * everyone else sees the template read-only.
 */
const StatusTemplateSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const access = useUserStore((s) => s.user?.effectiveAccess);
  const canManage = canFlowPermission('flow:team:manage', access);
  const templateQuery = useStatusTemplate();
  const saveTemplate = useSaveStatusTemplate();
  const [draft, setDraft] = useState<DraftItem[]>([]);
  const [dirty, setDirty] = useState(false);
  const [newName, setNewName] = useState('');

  // A fresh server copy replaces the draft only while there are no local edits.
  useEffect(() => {
    if (templateQuery.data && !dirty) setDraft(withKeys(templateQuery.data));
  }, [templateQuery.data, dirty]);

  const change = (next: DraftItem[]): void => {
    setDraft(next);
    setDirty(true);
  };

  const patch = (key: number, update: Partial<ApiStatusTemplateItem>): void =>
    change(draft.map((item) => (item.key === key ? { ...item, ...update } : item)));

  const move = (index: number, direction: -1 | 1): void => {
    const target = index + direction;
    if (target < 0 || target >= draft.length) return;
    const next = [...draft];
    [next[index], next[target]] = [next[target], next[index]];
    change(next);
  };

  const add = (): void => {
    const name = newName.trim();
    if (!name || draft.length >= MAX_STATUSES) return;
    change([...draft, { key: nextKey++, name, bucket: 'active', color: STATUS_COLORS[4] }]);
    setNewName('');
  };

  const problem = templateProblem(draft);

  const save = (): void => {
    if (problem) return;
    saveTemplate.mutate(
      draft.map(({ name, bucket, legacyStatus, color }, i) => ({
        name: name.trim(),
        bucket,
        legacyStatus,
        color,
        sortOrder: i * 10,
      })),
      { onSuccess: () => setDirty(false) }
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-flow-title font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
          {t('flow.settings.statuses.section')}
        </h2>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.settings.statuses.subtitle')}
        </p>
      </div>

      {!canManage && <GateNotice reason={t('flow.settings.statuses.readOnlyReason')} />}

      {hasQueryFailed(templateQuery) ? (
        <ListLoadFailed onRetry={() => void templateQuery.refetch()} />
      ) : templateQuery.isPending ? (
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flowCommon.loading')}
        </p>
      ) : (
        <>
          <ul className="space-y-1.5">
            {draft.map((item, index) => (
              <li
                key={item.key}
                className="rounded-md p-2"
                style={{ background: 'var(--flow-hover)' }}
              >
                <div className="flex items-center gap-2">
                  <span
                    aria-hidden="true"
                    className="h-2.5 w-2.5 shrink-0 rounded-full"
                    style={{ background: item.color }}
                  />
                  <input
                    aria-label={t('flow.status.statusName')}
                    value={item.name}
                    disabled={!canManage}
                    maxLength={100}
                    onChange={(e) => patch(item.key, { name: e.target.value })}
                    className="flow-focus min-h-[44px] min-w-0 flex-1 rounded-sm bg-transparent px-2 text-flow-body disabled:opacity-70 md:min-h-0 md:py-1"
                    style={{ color: 'var(--flow-text)' }}
                  />
                  <select
                    aria-label={t('flow.status.bucketFor', { name: item.name })}
                    value={item.bucket}
                    disabled={!canManage}
                    // The built-in value follows the bucket, so a stale one is dropped.
                    onChange={(e) =>
                      patch(item.key, {
                        bucket: e.target.value as StatusBucket,
                        legacyStatus: undefined,
                      })
                    }
                    className="flow-focus min-h-[44px] rounded-sm px-1.5 text-flow-body disabled:opacity-70 md:min-h-0 md:py-1"
                    style={inputStyle}
                  >
                    {STATUS_BUCKETS.map((b) => (
                      <option key={b} value={b}>
                        {t(`flow.status.bucket.${b}`)}
                      </option>
                    ))}
                  </select>
                  <button
                    type="button"
                    onClick={() => move(index, -1)}
                    disabled={!canManage || index === 0}
                    aria-label={t('flow.status.moveUpLabel', { name: item.name })}
                    className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <ChevronUp size={14} aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    onClick={() => move(index, 1)}
                    disabled={!canManage || index === draft.length - 1}
                    aria-label={t('flow.status.moveDownLabel', { name: item.name })}
                    className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <ChevronDown size={14} aria-hidden="true" />
                  </button>
                  <button
                    type="button"
                    onClick={() => change(draft.filter((other) => other.key !== item.key))}
                    disabled={!canManage}
                    aria-label={t('flow.status.removeStatusLabel', { name: item.name })}
                    className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm disabled:opacity-30 md:h-8 md:w-8"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <X size={14} aria-hidden="true" />
                  </button>
                </div>
                <div className="mt-2 flex flex-wrap items-center gap-1.5">
                  {STATUS_COLORS.map((c) => (
                    <button
                      key={c}
                      type="button"
                      disabled={!canManage}
                      onClick={() => patch(item.key, { color: c })}
                      aria-label={t('flow.status.chooseColorLabel', { color: c })}
                      aria-pressed={item.color === c}
                      className="flow-focus h-5 w-5 shrink-0 rounded-full disabled:opacity-40"
                      style={{
                        background: c,
                        outline: item.color === c ? `2px solid ${c}` : 'none',
                        outlineOffset: 2,
                      }}
                    />
                  ))}
                </div>
              </li>
            ))}
          </ul>

          {canManage && (
            <div className="flex flex-col gap-2 sm:flex-row">
              <input
                aria-label={t('flow.status.newStatusName')}
                value={newName}
                maxLength={100}
                onChange={(e) => setNewName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    add();
                  }
                }}
                placeholder={t('flow.status.newStatusPlaceholder')}
                className="flow-focus min-h-[44px] min-w-0 flex-1 rounded-sm px-2 text-flow-body md:min-h-0 md:py-1.5"
                style={inputStyle}
              />
              <button
                type="button"
                onClick={add}
                disabled={!newName.trim() || draft.length >= MAX_STATUSES}
                className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded-sm px-3 text-flow-body font-semibold disabled:opacity-40 md:min-h-0 md:py-1.5"
                style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
              >
                <Plus size={14} aria-hidden="true" />
                {t('flow.status.addStatus')}
              </button>
            </div>
          )}

          {canManage && problem && dirty && (
            <p role="alert" className="text-flow-meta" style={{ color: 'var(--flow-danger-fg)' }}>
              {t(problem)}
            </p>
          )}

          <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.settings.statuses.newProjectsOnly')}
          </p>

          {canManage && (
            <button
              type="button"
              onClick={save}
              disabled={!dirty || problem !== null || saveTemplate.isPending}
              className="flow-focus inline-flex min-h-[44px] items-center rounded-sm px-4 text-flow-body font-semibold disabled:opacity-40 md:min-h-0 md:py-2"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            >
              {saveTemplate.isPending
                ? t('flow.settings.statuses.saving')
                : t('flow.settings.statuses.save')}
            </button>
          )}
        </>
      )}
    </div>
  );
};

export default StatusTemplateSection;
