import React, { useEffect, useRef, useState } from 'react';
import { useOpenTask } from '../lib/useOpenTask';
import { toast } from 'sonner';
import { Calendar, User, CircleUser, FolderInput, Flag, Gauge, Tag } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { DatePicker, Dialog, DialogContent, DialogTitle, DialogDescription } from '@voxire/ui';
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
import { withAlpha } from '../lib/colorUtils';

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
  const { quickCreateOpen, quickCreateColumnId, setQuickCreateOpen } = useFlowStore();
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
    }
  }, [quickCreateOpen, currentProjectId]);

  const submit = async () => {
    if (!title.trim()) {
      toast.error(t('flowToasts.task.titleRequired'));
      return;
    }
    if (!canCreate || createTaskMutation.isPending) return;

    const uid = getCurrentUserId();

    try {
      const res = await createTaskMutation.mutateAsync({
        title: title.trim(),
        projectId: projectId || undefined,
        sectionId: projectId === currentProjectId ? quickCreateColumnId : undefined,
        priority,
        description: description.trim() || undefined,
        dueDate: dueDate || undefined,
        assigneeId: assignToMe ? (uid ?? undefined) : undefined,
        effortMinutes,
        labels: labelId ? [labelId] : undefined,
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
        className="portal-flow-dialog flex max-h-[calc(100dvh-32px)] flex-col gap-0 overflow-hidden p-0 sm:max-w-lg"
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
        {/* Header */}
        <div
          className="flex items-center justify-between gap-3 px-4 py-3 shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div className="min-w-0">
            <DialogTitle
              className="text-flow-body font-semibold"
              style={{ color: 'var(--flow-text)' }}
            >
              {t('flow.quickCreate.title')}
            </DialogTitle>
            <DialogDescription
              className="text-flow-meta mt-0.5"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              {t('flow.quickCreate.subtitle')}
            </DialogDescription>
          </div>
        </div>

        {/* Body */}
        <div className="flex-1 min-h-0 overflow-y-auto px-4 py-3">
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
                <ProjectSelect value={projectId} onChange={setProjectId} allowPersonal />
              </div>
            </FieldRow>

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

        {/* Footer */}
        <div
          className="flex items-center justify-end gap-2 px-4 py-3 shrink-0"
          style={{ borderTop: '1px solid var(--flow-border)' }}
        >
          <Button variant="ghost" onClick={() => setQuickCreateOpen(false)}>
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
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default QuickCreate;
