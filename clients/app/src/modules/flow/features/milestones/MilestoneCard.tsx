import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { Milestone as MilestoneIcon, Pencil, Trash2 } from 'lucide-react';
import { DatePicker } from '@voxire/ui';
import type { ApiMilestone } from '../../shared/services/flowApi';
import { useUpdateMilestone, useDeleteMilestone } from './useMilestones';
import { useCanDeleteResource } from '../../shared/lib/currentUser';
import { onTint, withAlpha } from '../../shared/lib/colorUtils';
import { formatDueDate } from '../../shared/lib/utils';
import { EMPTY_TASK_FILTERS, serializeTaskFilters } from '../../shared/filters/taskFilters';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import FlowSelect from '../../shared/components/FlowSelect';

type MilestoneStatus = ApiMilestone['status'];

const STATUS_STYLES: Record<MilestoneStatus, { bg: string; color: string }> = {
  planning: { bg: withAlpha('var(--flow-neutral-fg)', 12), color: 'var(--flow-neutral-fg)' },
  active: { bg: 'var(--flow-status-approved-bg)', color: 'var(--flow-status-approved)' },
  at_risk: { bg: withAlpha('var(--flow-warn-fg)', 12), color: 'var(--flow-warn-fg)' },
  completed: { bg: withAlpha('var(--flow-info-fg)', 12), color: 'var(--flow-info-fg)' },
};

const STATUS_LABEL_KEYS: Record<MilestoneStatus, string> = {
  planning: 'planning',
  active: 'active',
  at_risk: 'atRisk',
  completed: 'completed',
};

const STATUSES = Object.keys(STATUS_LABEL_KEYS) as MilestoneStatus[];

// Older milestones can carry a status this screen does not offer (the column
// also allows "upcoming" and "missed"); show those as planning rather than
// failing to render the card.
const knownStatus = (status: string): MilestoneStatus =>
  (STATUSES as string[]).includes(status) ? (status as MilestoneStatus) : 'planning';

export function milestonePercent(done: number, total: number): number {
  if (total <= 0) return 0;
  return Math.min(100, Math.max(0, Math.round((done / total) * 100)));
}

interface MilestoneCardProps {
  milestone: ApiMilestone;
  index: number;
}

const MilestoneCard: React.FC<MilestoneCardProps> = ({ milestone: m, index }) => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const canDelete = useCanDeleteResource(m.projectId);
  const updateMilestone = useUpdateMilestone();
  const deleteMilestone = useDeleteMilestone();

  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(m.name);
  const [targetDate, setTargetDate] = useState(m.targetDate);
  const [status, setStatus] = useState<MilestoneStatus>(knownStatus(m.status));
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const shownStatus = knownStatus(m.status);
  const style = STATUS_STYLES[shownStatus];
  const taskCount = m.taskCount ?? 0;
  const doneCount = m.doneCount ?? 0;
  const overdueCount = m.overdueCount ?? 0;
  const percent = milestonePercent(doneCount, taskCount);
  const accent = m.project?.color || 'var(--accent)';

  const openTasks = (): void => {
    void navigate({
      to: '/flow/p/$projectId/list',
      params: { projectId: m.projectId },
      search: { filter: serializeTaskFilters({ ...EMPTY_TASK_FILTERS, milestones: [m.id] }) },
    });
  };

  const startEditing = (): void => {
    setName(m.name);
    setTargetDate(m.targetDate);
    setStatus(shownStatus);
    setEditing(true);
  };

  const save = (): void => {
    const trimmed = name.trim();
    if (!trimmed || !targetDate) return;
    updateMilestone.mutate(
      { id: m.id, data: { name: trimmed, targetDate, status } },
      {
        onSuccess: () => {
          toast.success(t('flowToasts.milestoneUpdated'));
          setEditing(false);
        },
      }
    );
  };

  const inputStyle: React.CSSProperties = {
    background: 'var(--flow-input-bg)',
    border: '1px solid var(--flow-border-control)',
    color: 'var(--flow-text)',
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.16, delay: index * 0.03 }}
      className="rounded-xl p-5"
      style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
    >
      {editing ? (
        <form
          className="space-y-3"
          onSubmit={(e) => {
            e.preventDefault();
            save();
          }}
        >
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            aria-label={t('flow.milestones.namePlaceholder')}
            placeholder={t('flow.milestones.namePlaceholder')}
            className="w-full rounded-md px-3 py-2 text-flow-meta"
            style={inputStyle}
            autoFocus
          />
          <DatePicker
            value={targetDate}
            onChange={setTargetDate}
            locale={i18n.language}
            aria-label={t('flow.milestones.targetDate')}
            labels={{
              placeholder: t('flow.milestones.targetDate'),
              previousMonth: t('flowCommon.datePicker.previousMonth'),
              nextMonth: t('flowCommon.datePicker.nextMonth'),
            }}
            className="text-flow-meta"
          />
          <FlowSelect
            label={t('flowCommon.status')}
            value={status}
            options={STATUSES.map((s) => ({
              value: s,
              label: t(`flow.milestones.status.${STATUS_LABEL_KEYS[s]}`),
            }))}
            onChange={(value) => setStatus(knownStatus(value))}
          />
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={updateMilestone.isPending || !name.trim() || !targetDate}
              className="px-3 py-1.5 rounded-sm text-flow-meta font-semibold disabled:opacity-50"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            >
              {updateMilestone.isPending ? t('flowCommon.saving') : t('flowCommon.save')}
            </button>
            <button
              type="button"
              onClick={() => setEditing(false)}
              className="px-3 py-1.5 rounded-sm text-flow-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flowCommon.cancel')}
            </button>
          </div>
        </form>
      ) : (
        <>
          <div className="flex items-start justify-between gap-2 mb-3">
            <button
              type="button"
              onClick={openTasks}
              aria-label={t('flow.milestones.openTasks', { name: m.name })}
              className="flow-focus flex min-w-0 flex-1 items-center gap-2 rounded-md text-start"
            >
              <span
                className="h-8 w-8 shrink-0 rounded-md flex items-center justify-center"
                style={{ background: withAlpha(accent, 13) }}
              >
                <MilestoneIcon size={14} style={{ color: onTint(accent) }} aria-hidden="true" />
              </span>
              <span className="min-w-0">
                <span
                  className="block truncate text-flow-body font-bold"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {m.name}
                </span>
                <span
                  className="block truncate text-flow-meta"
                  style={{ color: 'var(--flow-text-dim)' }}
                >
                  {m.project?.name}
                </span>
              </span>
            </button>
            <div className="flex shrink-0 items-center gap-1">
              <span
                className="text-flow-meta px-2 py-0.5 rounded-full font-semibold"
                style={{ background: style.bg, color: style.color }}
              >
                {t(`flow.milestones.status.${STATUS_LABEL_KEYS[shownStatus]}`)}
              </span>
              <button
                type="button"
                onClick={startEditing}
                aria-label={t('flow.milestones.editMilestone', { name: m.name })}
                title={t('flowCommon.edit')}
                className="flow-focus rounded-md p-1.5 hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <Pencil size={13} aria-hidden="true" />
              </button>
              {canDelete && (
                <button
                  type="button"
                  onClick={() => setConfirmingDelete(true)}
                  aria-label={t('flow.milestones.deleteMilestone', { name: m.name })}
                  title={t('flowCommon.delete')}
                  className="flow-focus rounded-md p-1.5 hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  <Trash2 size={13} aria-hidden="true" />
                </button>
              )}
            </div>
          </div>
          {m.description && (
            <p className="text-flow-meta mb-3" style={{ color: 'var(--flow-text-muted)' }}>
              {m.description}
            </p>
          )}
          {taskCount > 0 ? (
            <div className="mb-3">
              <div
                className="mb-1 flex items-center justify-between text-flow-meta"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span>{t('flow.milestones.progress')}</span>
                <span style={{ color: 'var(--flow-text)' }}>{percent}%</span>
              </div>
              <div
                role="progressbar"
                aria-label={t('flow.milestones.progress')}
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={percent}
                className="h-1.5 rounded-full overflow-hidden"
                style={{ background: 'var(--flow-hover)' }}
              >
                <div
                  className="h-full rounded-full"
                  style={{ width: `${percent}%`, background: accent }}
                />
              </div>
            </div>
          ) : null}
          <div
            className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 text-flow-meta"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            <span className="flex items-center gap-3">
              {taskCount > 0 ? (
                <span>
                  {t('flow.milestones.taskCount', { done: doneCount, total: taskCount })}
                </span>
              ) : (
                <span>{t('flow.milestones.noTasksHint')}</span>
              )}
              {overdueCount > 0 && (
                <span className="font-semibold" style={{ color: 'var(--flow-warn-fg)' }}>
                  {t('flow.milestones.overdueCount', { count: overdueCount })}
                </span>
              )}
            </span>
            <span>
              {t('flow.milestones.target')} {formatDueDate(m.targetDate, t)}
            </span>
          </div>
        </>
      )}
      <ConfirmDialog
        open={confirmingDelete}
        onOpenChange={setConfirmingDelete}
        title={t('flow.milestones.deleteTitle')}
        description={t('flow.milestones.deleteDescription', { name: m.name })}
        confirmLabel={t('flowCommon.delete')}
        cancelLabel={t('flowCommon.cancel')}
        pending={deleteMilestone.isPending}
        onConfirm={() =>
          deleteMilestone.mutate(m.id, {
            onSuccess: () => {
              toast.success(t('flowToasts.milestoneDeleted'));
              setConfirmingDelete(false);
            },
          })
        }
      />
    </motion.div>
  );
};

export default MilestoneCard;
