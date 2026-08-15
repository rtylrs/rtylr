import React, { useMemo, useState } from 'react';
import { Plus } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@voxire/ui';
import { formatDurationClock, parseDurationInput } from '../../shared/lib/durationUtils';
import DurationField from '../tasks/components/DurationField';
import { useProjects } from '../projects/useProjects';
import { newClientKey, useCreateWorkDayLine, type WorkType } from './useWorkDay';

const labelClass = 'mb-1 block text-[10px] font-semibold uppercase tracking-wider';
const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border)',
  color: 'var(--flow-text)',
};

interface LogTimeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  date: string;
  workDayId: string;
  workspaceId: string | undefined;
  workTypes: WorkType[];
}

/**
 * Logging a line gets its own screen rather than an inline row: a business's
 * work-type tree can run two levels deep (a type, then its named sub-tasks),
 * and that only reads well with room - a picker this size does not belong
 * squeezed under the day's clock figures.
 *
 * The first description is never typed: it is the breadcrumb of whatever
 * project and work type were picked, computed here and shown read-only. The
 * second is the person's own account of what they did. Two different kinds
 * of "what is this" - the classification and the account of it - collapsing
 * them into one field is what made the old inline row feel thin.
 */
const LogTimeDialog: React.FC<LogTimeDialogProps> = ({
  open,
  onOpenChange,
  date,
  workDayId,
  workspaceId,
  workTypes,
}) => {
  const { t } = useTranslation();
  const { data: projects } = useProjects(workspaceId);
  const createLine = useCreateWorkDayLine();

  const [duration, setDuration] = useState('');
  const [description, setDescription] = useState('');
  const [projectId, setProjectId] = useState('');
  const [topLevelId, setTopLevelId] = useState('');
  const [childId, setChildId] = useState('');

  const topLevelTypes = useMemo(() => workTypes.filter((wt) => !wt.parentId), [workTypes]);
  const childTypes = useMemo(
    () => workTypes.filter((wt) => wt.parentId === topLevelId),
    [workTypes, topLevelId]
  );

  const selectedTopLevel = topLevelTypes.find((wt) => wt.id === topLevelId);
  const selectedChild = childTypes.find((wt) => wt.id === childId);
  const workTypeId = selectedChild?.id || selectedTopLevel?.id || '';

  const taskDescription = [selectedTopLevel?.name, selectedChild?.name].filter(Boolean).join(' / ');

  const canAddLine = parseDurationInput(duration) !== null && projectId !== '';

  const reset = (): void => {
    setDuration('');
    setDescription('');
    setProjectId('');
    setTopLevelId('');
    setChildId('');
  };

  const pickTopLevel = (id: string): void => {
    setTopLevelId(id);
    setChildId('');
  };

  const addLine = (): void => {
    const minutes = parseDurationInput(duration);
    if (minutes === null || !projectId) return;
    createLine.mutate(
      {
        date,
        workDayId,
        // Generated here, before the request leaves, so a retry after a
        // dropped connection resolves to the same entry rather than a
        // second one.
        clientKey: newClientKey(),
        durationMinutes: minutes,
        description: description.trim(),
        projectId,
        workTypeId: workTypeId || undefined,
        isBillable: true,
      },
      {
        onSuccess: () => {
          reset();
          onOpenChange(false);
        },
      }
    );
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next) reset();
        onOpenChange(next);
      }}
    >
      <DialogContent
        className="max-w-lg"
        style={{ background: 'var(--flow-card)', color: 'var(--flow-text)' }}
      >
        <DialogHeader>
          <DialogTitle style={{ color: 'var(--flow-text)' }}>{t('flow.time.logTime')}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <DurationField
            id="log-time-duration"
            label={t('flow.time.duration')}
            value={duration}
            onChange={setDuration}
            placeholder={t('flow.time.durationPlaceholder')}
            invalidMessage={t('flow.time.durationInvalid')}
          />

          <div>
            <label
              htmlFor="log-time-project"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.project')}
            </label>
            <select
              id="log-time-project"
              aria-label={t('flow.time.project')}
              value={projectId}
              onChange={(e) => setProjectId(e.target.value)}
              required
              className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
              style={inputStyle}
            >
              <option value="">{t('flow.time.selectProject')}</option>
              {(projects ?? []).map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name}
                </option>
              ))}
            </select>
          </div>

          {topLevelTypes.length > 0 && (
            <div>
              <span className={labelClass} style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.time.workType')}
              </span>
              <div className="flex flex-wrap gap-1.5">
                {topLevelTypes.map((wt) => {
                  const active = wt.id === topLevelId;
                  return (
                    <button
                      key={wt.id}
                      type="button"
                      onClick={() => pickTopLevel(wt.id ?? '')}
                      className="flow-focus min-h-[44px] rounded-md px-3 text-[13px] font-medium md:min-h-0 md:py-1.5"
                      style={{
                        background: active ? 'var(--accent)' : 'var(--flow-hover)',
                        color: active ? 'var(--flow-on-accent)' : 'var(--flow-text)',
                      }}
                    >
                      {wt.name}
                    </button>
                  );
                })}
              </div>

              {childTypes.length > 0 && (
                <div className="mt-2 flex flex-wrap gap-1.5">
                  {childTypes.map((wt) => {
                    const active = wt.id === childId;
                    return (
                      <button
                        key={wt.id}
                        type="button"
                        onClick={() => setChildId(wt.id ?? '')}
                        className="flow-focus min-h-[44px] rounded-md px-3 text-[13px] font-medium md:min-h-0 md:py-1.5"
                        style={{
                          background: active ? 'var(--accent)' : 'var(--flow-hover)',
                          color: active ? 'var(--flow-on-accent)' : 'var(--flow-text)',
                        }}
                      >
                        {wt.name}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          <div>
            <label
              htmlFor="log-time-task-description"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.taskDescription')}
            </label>
            <input
              id="log-time-task-description"
              aria-label={t('flow.time.taskDescription')}
              value={taskDescription}
              readOnly
              placeholder={t('flow.time.taskDescriptionPlaceholder')}
              className="min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
              style={{ ...inputStyle, color: 'var(--flow-text-muted)' }}
            />
          </div>

          <div>
            <label
              htmlFor="log-time-description"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.description')}
            </label>
            <input
              id="log-time-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && addLine()}
              className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
              style={inputStyle}
            />
          </div>

          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="flow-focus min-h-[44px] rounded px-3 text-[13px] font-medium md:min-h-0 md:py-2"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.cancel')}
            </button>
            <button
              type="button"
              onClick={addLine}
              disabled={!canAddLine || createLine.isPending}
              className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded px-4 text-[13px] font-semibold disabled:opacity-40 md:min-h-0 md:py-2.5"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              <Plus size={14} aria-hidden="true" />
              {parseDurationInput(duration) !== null
                ? t('flow.time.logAmount', {
                    amount: formatDurationClock(parseDurationInput(duration) ?? 0),
                  })
                : t('flow.time.addLine')}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default LogTimeDialog;
