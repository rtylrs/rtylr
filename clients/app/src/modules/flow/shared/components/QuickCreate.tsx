import React, { useEffect, useRef, useState } from 'react';
import { useOpenTask } from '../lib/useOpenTask';
import { toast } from 'sonner';
import {
  Calendar,
  User,
  CircleUser,
  FolderInput,
  Flag,
  Gauge,
  Tag,
  CircleDot,
  Milestone as MilestoneIcon,
  SlidersHorizontal,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { DatePicker } from '@voxire/ui';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../../shared/components/ui/dialog';
import { useParams, useNavigate, useRouterState } from '@tanstack/react-router';
import { useProjects } from '../../features/projects/useProjects';
import { apiProjectToProject } from '../lib/projectMapper';
import { canFlowPermission } from '../lib/roleUtils';
import { useFlowStore, Task } from '../stores/flowStore';
import { priorityLabelKey } from '../lib/priorityLabels';
import { useUserStore } from '../stores/userStore';
import { Button } from '../../../../shared/components/ui/button';
import ProjectSelect from './ProjectSelect';
import PriorityIcon from './PriorityIcon';
import { cn, getInitials } from '../lib/utils';
import { getCurrentUserId } from '../lib/currentUser';
import { displayUserName } from '../lib/userDisplay';
import { EFFORT_PRESETS, formatEffortMinutes } from '../lib/effortPresets';
import { useCreateTask } from '../../features/tasks/useTasks';
import { useLabels } from '../../features/tasks/useLabels';
import { useProjectStatuses } from '../../features/projects/useStatuses';
import { useMilestones } from '../../features/milestones/useMilestones';
import {
  defaultStatus,
  displayStatusName,
  sortStatuses,
} from '../../features/projects/statusUtils';
import FlowSelect from './FlowSelect';
import CustomFieldInput from './CustomFieldInput';
import { withAlpha } from '../lib/colorUtils';
import { useProjectFields } from '../../features/projects/useFields';
import { useProjectMembers } from '../../features/projects/useProjectMembers';
import type { ApiField, ApiTaskFieldValue } from '../services/flowApi';
import { initialFieldValues, missingRequiredFields, toFieldValueInputs } from '../lib/customFields';

const NO_FIELDS: ApiField[] = [];

function FieldRow({
  label,
  icon: Icon,
  children,
}: {
  label: string;
  icon?: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  children: React.ReactNode;
}) {
  return (
    <div
      className="flex items-start gap-3 py-2.5"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <div
        className="flex items-center gap-2 w-[100px] shrink-0 pt-0.5"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {Icon && <Icon size={14} className="shrink-0 opacity-70" />}
        <span className="text-flow-meta font-medium">{label}</span>
      </div>
      <div className="flex-1 min-w-0 text-flow-body">{children}</div>
    </div>
  );
}

export const QuickCreate: React.FC = () => {
  const { t, i18n } = useTranslation();
  const openTask = useOpenTask();
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const quickCreateOpen = useFlowStore((state) => state.quickCreateOpen);
  const quickCreateColumnId = useFlowStore((state) => state.quickCreateColumnId);
  const quickCreateCycleId = useFlowStore((state) => state.quickCreateCycleId);
  const setQuickCreateOpen = useFlowStore((state) => state.setQuickCreateOpen);
  const user = useUserStore((s) => s.user);
  const { projectId: currentProjectId } = useParams({ strict: false });
  const canCreate = canFlowPermission('flow:tasks:create', user?.effectiveAccess);
  const { data: availableProjects } = useProjects({
    enabled: quickCreateOpen && canFlowPermission('flow:projects:view', user?.effectiveAccess),
  });
  useEffect(() => {
    if (availableProjects)
      useFlowStore.getState().setProjects(availableProjects.map(apiProjectToProject));
  }, [availableProjects]);
  const createTaskMutation = useCreateTask();
  const { data: labels = [] } = useLabels();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [projectId, setProjectId] = useState<string>('');
  const [priority, setPriority] = useState<Task['priority']>('medium');
  const [dueDate, setDueDate] = useState('');
  const [assignToMe, setAssignToMe] = useState(true);
  const [effortMinutes, setEffortMinutes] = useState<number | undefined>(30);
  const [labelId, setLabelId] = useState('');
  const [statusId, setStatusId] = useState('');
  const [milestoneId, setMilestoneId] = useState('');
  const { data: projectMilestones = [] } = useMilestones(quickCreateOpen ? projectId : '');
  const { data: rawStatuses } = useProjectStatuses(quickCreateOpen ? projectId : '');
  const statuses = rawStatuses ? sortStatuses(rawStatuses) : [];
  // A stable empty list: a fresh `[]` each render would re-run the reset below forever.
  const projectFields = useProjectFields(quickCreateOpen ? projectId : '').data ?? NO_FIELDS;
  const { data: projectMembers = [] } = useProjectMembers(
    quickCreateOpen && projectFields.some((f) => f.type === 'person') ? projectId : undefined
  );
  const [fieldValues, setFieldValues] = useState<Record<string, ApiTaskFieldValue>>({});
  const [missingFields, setMissingFields] = useState<string[]>([]);

  const titleRef = useRef<HTMLTextAreaElement>(null);
  const returnFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (quickCreateOpen) {
      setTitle('');
      setDescription('');
      setProjectId(currentProjectId ?? '');
      setPriority('medium');
      setDueDate('');
      setAssignToMe(true);
      setEffortMinutes(30);
      setLabelId('');
      setMilestoneId('');
    }
  }, [quickCreateOpen, currentProjectId]);

  // Custom fields belong to the project, so a new project starts from its
  // own defaults rather than carrying values the next one cannot hold.
  useEffect(() => {
    if (!quickCreateOpen) return;
    setFieldValues(initialFieldValues(projectFields));
    setMissingFields([]);
  }, [quickCreateOpen, projectFields]);

  // The status follows the project: a board column's own status when the
  // dialog was opened from that column, otherwise the project's default.
  useEffect(() => {
    if (!quickCreateOpen || !rawStatuses) return;
    setStatusId((current) => {
      if (rawStatuses.some((s) => s.id === current)) return current;
      const fromColumn =
        projectId === currentProjectId && quickCreateColumnId
          ? rawStatuses.find((s) => s.id === quickCreateColumnId)
          : undefined;
      return (fromColumn ?? defaultStatus(rawStatuses))?.id ?? '';
    });
  }, [quickCreateOpen, rawStatuses, projectId, currentProjectId, quickCreateColumnId]);

  const submit = async () => {
    if (!title.trim()) {
      toast.error(t('flowToasts.task.titleRequired'));
      return;
    }
    if (!canCreate || createTaskMutation.isPending) return;
    const missing = missingRequiredFields(projectFields, fieldValues);
    if (missing.length > 0) {
      const names = missing.map((f) => f.name).join(', ');
      setMissingFields(missing.map((f) => f.id));
      toast.error(t('flow.fields.requiredMissing', { names }));
      return;
    }

    const uid = getCurrentUserId();

    try {
      const res = await createTaskMutation.mutateAsync({
        title: title.trim(),
        projectId: projectId || undefined,
        // A task's column is its status, so section_id carries the pick.
        sectionId: statuses.some((s) => s.id === statusId) ? statusId : undefined,
        priority,
        description: description.trim() || undefined,
        dueDate: dueDate || undefined,
        assigneeId: assignToMe ? (uid ?? undefined) : undefined,
        effortMinutes,
        labels: labelId ? [labelId] : undefined,
        milestoneId: projectMilestones.some((m) => m.id === milestoneId) ? milestoneId : undefined,
        fieldValues: toFieldValueInputs(projectFields, fieldValues),
        // Opened from a Backlog section: the task joins that cycle, unless the
        // project was changed in the dialog.
        cycleId:
          quickCreateCycleId && projectId === currentProjectId ? quickCreateCycleId : undefined,
      });
      const task = res.task;
      const taskId = String(task.id);
      toast.success(t('flowToasts.task.createdWithId', { id: task.identifier ?? taskId }), {
        description: task.title,
      });
      setQuickCreateOpen(false);
      if (canFlowPermission('flow:tasks:view', user?.effectiveAccess)) {
        if (pathname.startsWith('/flow/')) openTask(taskId);
        else void navigate({ to: '/flow/my-issues', search: { openTask: taskId } });
      }
    } catch {
      // error toast handled by mutation onError
    }
  };

  const priorities: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];
  const assigneeName = displayUserName(user);

  return (
    <Dialog open={quickCreateOpen} onOpenChange={(open) => setQuickCreateOpen(open)}>
      <DialogContent
        className="portal-flow-dialog flex flex-col sm:max-w-lg"
        dir={i18n.dir()}
        closeLabel={t('flowCommon.close')}
        onOpenAutoFocus={(event) => {
          event.preventDefault();
          returnFocusRef.current =
            document.activeElement instanceof HTMLElement &&
            document.activeElement !== document.body
              ? document.activeElement
              : null;
          titleRef.current?.focus();
        }}
        onCloseAutoFocus={(event) => {
          event.preventDefault();
          const target = returnFocusRef.current?.isConnected
            ? returnFocusRef.current
            : document.querySelector<HTMLButtonElement>('[data-portal-command-trigger]');
          target?.focus();
        }}
      >
        <DialogHeader className="shrink-0">
          <DialogTitle>{t('flow.quickCreate.title')}</DialogTitle>
          <DialogDescription>{t('flow.quickCreate.subtitle')}</DialogDescription>
        </DialogHeader>

        {/* Body */}
        <div className="min-h-0 flex-1 overflow-y-auto px-6 pb-6">
          <textarea
            ref={titleRef}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                submit();
              }
            }}
            aria-label={t('flow.quickCreate.titlePlaceholder')}
            placeholder={t('flow.quickCreate.titlePlaceholder')}
            rows={2}
            className="flow-task-name"
          />

          <div className="mt-1">
            <FieldRow label={t('flowCommon.assignee')} icon={User}>
              <div className="flex items-center gap-2 flex-wrap">
                <button
                  type="button"
                  onClick={() => setAssignToMe(true)}
                  className={cn(
                    'inline-flex items-center gap-2 rounded-md px-2.5 py-1.5 text-flow-meta font-medium transition-colors',
                    assignToMe ? '' : 'hover:bg-[var(--flow-hover)]'
                  )}
                  style={{
                    background: assignToMe ? 'var(--flow-hover)' : 'transparent',
                    color: 'var(--flow-text)',
                    border: assignToMe
                      ? `1px solid ${withAlpha('var(--accent)', 27)}`
                      : '1px solid transparent',
                  }}
                >
                  <span
                    className="h-6 w-6 rounded-full flex items-center justify-center text-flow-meta font-bold"
                    style={{
                      background: 'var(--accent-muted)',
                      color: 'var(--accent-on-muted)',
                    }}
                  >
                    {getInitials(assigneeName)}
                  </span>
                  {assigneeName}
                </button>
                <button
                  type="button"
                  onClick={() => setAssignToMe(false)}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-flow-meta transition-colors',
                    !assignToMe ? '' : 'hover:bg-[var(--flow-hover)]'
                  )}
                  style={{
                    background: !assignToMe ? 'var(--flow-hover)' : 'transparent',
                    color: 'var(--flow-text-muted)',
                    border: !assignToMe ? '1px solid var(--flow-border)' : '1px solid transparent',
                  }}
                >
                  <CircleUser size={14} />
                  {t('flow.quickCreate.unassigned')}
                </button>
              </div>
            </FieldRow>

            <FieldRow label={t('flow.taskDetail.effort')} icon={Gauge}>
              <div className="flex flex-wrap gap-1">
                <button
                  type="button"
                  onClick={() => setEffortMinutes(undefined)}
                  className="text-flow-meta px-2 py-1 rounded-md"
                  style={{
                    background: effortMinutes == null ? 'var(--flow-hover)' : 'transparent',
                    color: 'var(--flow-text-muted)',
                    border: '1px solid var(--flow-border)',
                  }}
                >
                  {t('flow.quickCreate.noEffort')}
                </button>
                {EFFORT_PRESETS.map((m) => (
                  <button
                    key={m}
                    type="button"
                    onClick={() => setEffortMinutes(m)}
                    className="text-flow-meta px-2 py-1 rounded-md"
                    style={{
                      background: effortMinutes === m ? 'var(--accent-muted)' : 'transparent',
                      color: 'var(--flow-text)',
                      border:
                        effortMinutes === m
                          ? `1px solid ${withAlpha('var(--accent)', 33)}`
                          : '1px solid var(--flow-border)',
                    }}
                  >
                    {formatEffortMinutes(m)}
                  </button>
                ))}
              </div>
            </FieldRow>

            <FieldRow label={t('flowCommon.dueDate')} icon={Calendar}>
              <DatePicker
                value={dueDate}
                onChange={setDueDate}
                locale={i18n.language}
                aria-label={t('flowCommon.dueDate')}
                wrapperClassName="max-w-[200px]"
                className="h-9 min-h-0 py-1.5 text-flow-body"
                labels={{
                  placeholder: t('flowCommon.dueDate'),
                  previousMonth: t('flowCommon.datePicker.previousMonth'),
                  nextMonth: t('flowCommon.datePicker.nextMonth'),
                }}
              />
            </FieldRow>

            <FieldRow label={t('flow.time.project')} icon={FolderInput}>
              <div className="-ml-1">
                <ProjectSelect
                  value={projectId}
                  onChange={(next) => {
                    setProjectId(next);
                    // A milestone belongs to one project.
                    setMilestoneId('');
                  }}
                  allowPersonal
                />
              </div>
            </FieldRow>

            {statuses.length > 0 && (
              <FieldRow label={t('flow.taskDetail.status')} icon={CircleDot}>
                <div className="max-w-[240px]">
                  <FlowSelect
                    label={t('flow.taskDetail.status')}
                    value={statusId}
                    options={statuses.map((s) => ({
                      value: s.id,
                      label: displayStatusName(s.name, t),
                      icon: (
                        <span
                          aria-hidden="true"
                          className="h-2 w-2 shrink-0 rounded-full"
                          style={{ background: s.color }}
                        />
                      ),
                    }))}
                    onChange={setStatusId}
                  />
                </div>
              </FieldRow>
            )}

            {projectMilestones.length > 0 && (
              <FieldRow label={t('flow.taskDetail.milestone')} icon={MilestoneIcon}>
                <div className="max-w-[240px]">
                  <FlowSelect
                    label={t('flow.taskDetail.milestone')}
                    value={milestoneId}
                    options={[
                      { value: '', label: t('flow.taskDetail.noMilestone') },
                      ...projectMilestones.map((m) => ({ value: m.id, label: m.name })),
                    ]}
                    onChange={setMilestoneId}
                    muted={!milestoneId}
                  />
                </div>
              </FieldRow>
            )}

            {projectFields.map((field) => {
              const missing = missingFields.includes(field.id);
              return (
                <FieldRow
                  key={field.id}
                  label={field.required ? `${field.name} *` : field.name}
                  icon={SlidersHorizontal}
                >
                  <div className="max-w-[240px]" data-quick-create-field={field.id}>
                    <CustomFieldInput
                      field={field}
                      value={fieldValues[field.id]}
                      people={projectMembers.map((m) => ({ id: String(m.userId), name: m.name }))}
                      onChange={(next) => {
                        setFieldValues((current) => ({ ...current, [field.id]: next }));
                        setMissingFields((current) => current.filter((id) => id !== field.id));
                      }}
                    />
                    {missing && (
                      <p
                        role="alert"
                        className="mt-1 text-flow-meta"
                        style={{ color: 'var(--flow-danger-fg)' }}
                      >
                        {t('flow.fields.required')}
                      </p>
                    )}
                  </div>
                </FieldRow>
              );
            })}

            <FieldRow label={t('flowCommon.priority')} icon={Flag}>
              <div className="flex items-center gap-1 flex-wrap">
                {priorities.map((p) => (
                  <button
                    key={p}
                    type="button"
                    onClick={() => setPriority(p)}
                    className="h-8 w-8 rounded-md flex items-center justify-center transition-colors"
                    style={{
                      background: priority === p ? 'var(--flow-hover)' : 'transparent',
                      border:
                        priority === p ? '1px solid var(--flow-border)' : '1px solid transparent',
                    }}
                    title={t(p === 'no_priority' ? 'flowPriority.noPriority' : `flowPriority.${p}`)}
                    aria-label={t(priorityLabelKey(p))}
                  >
                    <PriorityIcon priority={p} />
                  </button>
                ))}
              </div>
            </FieldRow>

            <FieldRow label={t('flowCommon.settings.labels')} icon={Tag}>
              <select
                aria-label={t('flowCommon.settings.labels')}
                value={labelId}
                onChange={(e) => setLabelId(e.target.value)}
                disabled={labels.length === 0}
                className="w-full max-w-[240px] rounded-md px-2 py-1.5 text-flow-body disabled:opacity-50"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              >
                {labels.length === 0 ? (
                  <option value="">{t('flowCommon.noData')}</option>
                ) : (
                  <>
                    <option value="">{t('flow.quickCreate.noLabel')}</option>
                    {labels.map((label) => (
                      <option key={label.id} value={label.id}>
                        {label.name}
                      </option>
                    ))}
                  </>
                )}
              </select>
            </FieldRow>
          </div>

          <div className="mt-4">
            <p
              className="text-flow-meta font-semibold uppercase tracking-wide mb-2"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flowCommon.description')}
            </p>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('flow.quickCreate.descriptionPlaceholder')}
              rows={6}
              className="w-full rounded-lg px-3 py-2.5 text-flow-body leading-relaxed resize-y min-h-[120px] placeholder:opacity-45"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            />
          </div>
        </div>

        <DialogFooter className="shrink-0">
          <Button variant="outline" onClick={() => setQuickCreateOpen(false)}>
            {t('flowCommon.cancel')}
          </Button>
          <Button
            onClick={submit}
            disabled={!title.trim() || !canCreate || createTaskMutation.isPending}
          >
            {createTaskMutation.isPending
              ? t('flow.quickCreate.creating')
              : t('flow.quickCreate.createTask')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default QuickCreate;
