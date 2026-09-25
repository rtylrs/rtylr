import React, { useEffect, useState } from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useNavigate, useParams } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import {
  ArrowLeft,
  Copy,
  Tag,
  User as UserIcon,
  Calendar as CalendarIcon,
  Clock,
  Milestone as MilestoneIcon,
  Timer,
  Flag,
  CircleDot,
  Trash2,
  Archive,
  GitBranch,
  LayoutList,
  Link2,
  ListChecks,
  Gauge,
  Hourglass,
  Pin,
  Lightbulb,
  Activity,
  Type,
  PartyPopper,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTask } from './useTask';
import {
  useUpdateTask,
  useDeleteTask,
  useAddComment,
  useAddSubtask,
  useToggleSubtask,
  useTaskAttachments,
  useAddAttachment,
  useRemoveAttachment,
} from './useTasks';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { useLabels, useSetTaskLabels } from './useLabels';
import { useProjectStatuses, useSetTaskStatus } from '../projects/useStatuses';
import { useTaskActivity } from './useTaskActivity';
import { useTaskRealtime } from '../../shared/lib/useTaskRealtime';
import { useCurrentProject } from '../projects/useCurrentProject';
import { useApprovedLeaveInRange } from '../../shared/hooks/queries/useApprovedLeaveInRange';
import TimePanel from './components/TimePanel';
import { useFlowStore, Attachment, Task } from '../../shared/stores/flowStore';
import type { ApiTask, ApiTaskDetail, ApiTaskActivity } from '../../shared/services/flowApi';
import { useUserStore } from '../../shared/stores/userStore';
import { useCanEditTask, useCanDeleteResource } from '../../shared/lib/currentUser';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import CommentList from '../../shared/components/CommentList';
import CommentInput from '../../shared/components/CommentInput';
import SubtaskList from '../../shared/components/SubtaskList';
import AttachmentUpload from '../../shared/components/AttachmentUpload';
import AssigneeAvatar from '../../shared/components/AssigneeAvatar';
import LabelChip from '../../shared/components/LabelChip';
import DueDatePicker from '../../shared/components/DueDatePicker';
import { getInitials } from '../../shared/lib/utils';
import { displayDescription } from '../../shared/lib/stringUtils';
import { EFFORT_PRESETS, formatEffortMinutes } from '../../shared/lib/effortPresets';
import { TASK_STATUSES, STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import { withAlpha } from '../../shared/lib/colorUtils';
import { toLocalIso, todayIso } from '../../shared/lib/dateUtils';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatRelative(dateStr: string, t: TFunction): string {
  const d = new Date(dateStr);
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return t('flow.taskDetail.justNow');
  if (mins < 60) return t('flow.taskDetail.minutesAgo', { count: mins });
  const hours = Math.floor(mins / 60);
  if (hours < 24) return t('flow.taskDetail.hoursAgo', { count: hours });
  const days = Math.floor(hours / 24);
  if (days < 30) return t('flow.taskDetail.daysAgo', { count: days });
  return d.toLocaleDateString(t('flow.taskDetail.dateLocale'), {
    month: 'short',
    day: 'numeric',
  });
}

function humanizeValue(action: string, value: string): string {
  if (!value) return '-';
  if (action.includes('status') || action === 'status_changed' || action === 'status_change') {
    return value
      .split('_')
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ');
  }
  if (
    action.includes('priority') ||
    action === 'priority_changed' ||
    action === 'priority_change'
  ) {
    return value.charAt(0).toUpperCase() + value.slice(1);
  }
  return value;
}

function inferAttachmentMime(name: string): string {
  if (/\.(apng|avif|gif|jpe?g|png|svg|webp)$/i.test(name)) {
    return 'image/*';
  }
  return 'application/octet-stream';
}

function describeActivity(a: ApiTaskActivity, t: TFunction): React.ReactNode {
  const action = a.action;
  const oldVal = humanizeValue(action, a.oldValue);
  const newVal = humanizeValue(action, a.newValue);

  if (action === 'status_changed' || action === 'status_change') {
    return t('flow.taskDetail.statusChanged', { from: oldVal, to: newVal });
  }
  if (action === 'priority_changed' || action === 'priority_change') {
    return t('flow.taskDetail.priorityChanged', { from: oldVal, to: newVal });
  }
  if (action === 'assignee_changed' || action === 'assignee_change') {
    return t('flow.taskDetail.assigneeChanged', { to: newVal });
  }
  if (action === 'title_changed' || action === 'title_change') {
    return t('flow.taskDetail.titleChanged');
  }
  if (action === 'due_date_changed' || action === 'due_date_change') {
    return t('flow.taskDetail.dueDateChanged', { to: newVal });
  }
  // generic
  const label = action
    .replace(/_/g, ' ')
    .replace(/\bchanged\b/, 'changed')
    .trim();
  return a.newValue
    ? t('flow.taskDetail.genericActivity', { label, value: newVal })
    : t('flow.taskDetail.genericActivityNoValue', { label });
}

function activityIcon(action: string): React.ComponentType<{ size?: number }> {
  if (action.includes('status')) return CircleDot;
  if (action.includes('priority')) return Flag;
  if (action.includes('assignee')) return UserIcon;
  if (action.includes('title')) return Type;
  if (action.includes('due_date')) return CalendarIcon;
  if (action.includes('label')) return Tag;
  return Activity;
}

// ── Component ────────────────────────────────────────────────────────────────

/**
 * `taskId` overrides the route param, so the same component serves the
 * standalone `/task/$id` page and the drawer that opens over a list
 * (rtylr#1079). The drawer knows which task it is showing; the page reads it
 * from the URL.
 */
const TaskDetail: React.FC<{ taskId?: string }> = ({ taskId }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const openTask = useOpenTask();
  const { id: routeId } = useParams({ strict: false });
  const id = taskId ?? routeId;
  useTaskRealtime(id);
  const taskQuery = useTask(id);
  const { data: taskData, isLoading } = taskQuery;
  const user = useUserStore((s) => s.user);
  // Read up here rather than below the isLoading early return: these gates are
  // hooks, so they have to run on every render, and they need the task's project.
  const detail = taskData as ApiTaskDetail | undefined;
  const canEdit = useCanEditTask(detail?.task?.projectId);
  // F-A: empty means the project has not configured any custom statuses, the
  // normal case - the picker below stays hidden rather than showing an
  // empty dropdown.
  const { data: projectStatuses = [] } = useProjectStatuses(detail?.task?.projectId ?? '');
  const setTaskStatus = useSetTaskStatus();
  // HR->FLOW-02 (#838): warn before assigning work into approved leave.
  // Window is today through the task's due date (or two weeks out when
  // undated). Hidden entirely when HR is absent or unreachable; assignment
  // itself is never blocked. Ranges are shown without names because leave
  // records key on HR employee ids — the hub-user mapping lands with #848.
  const warningFrom = todayIso();
  const warningTo = detail?.task?.dueDate || toLocalIso(new Date(Date.now() + 14 * 86_400_000));
  const { data: overlappingLeave } = useApprovedLeaveInRange(warningFrom, warningTo, canEdit);
  const canRemove = useCanDeleteResource(detail?.task?.projectId);
  const updateTaskMut = useUpdateTask();
  const deleteTaskMut = useDeleteTask();
  const addCommentMut = useAddComment();
  const addSubtaskMut = useAddSubtask();
  const toggleSubtaskMut = useToggleSubtask();
  const setTaskLabelsMut = useSetTaskLabels();

  const { data: taskActivityData = [] } = useTaskActivity(id ?? '');
  const { data: taskAttachmentData } = useTaskAttachments(id);
  const addAttachmentMut = useAddAttachment();
  const removeAttachmentMut = useRemoveAttachment();

  const taskWorkspaceId = detail?.task?.workspaceId;
  const taskProjectId = detail?.task?.projectId;
  const { project: taskProject } = useCurrentProject(taskWorkspaceId, taskProjectId);
  const { data: allLabels = [] } = useLabels();
  const {
    members,
    cycles,
    milestones,
    archiveTask,
    attachments,
    projects,
    tasks: allTasks,
    sections,
    customFieldDefinitions,
    setActiveWorkspace,
  } = useFlowStore();

  const [labelPickerOpen, setLabelPickerOpen] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  useEffect(() => {
    if (taskWorkspaceId && taskWorkspaceId !== 'personal') {
      setActiveWorkspace(taskWorkspaceId);
    }
  }, [setActiveWorkspace, taskWorkspaceId]);

  const updateTask = (
    taskId: string,
    patch: Partial<Task>,
    options?: { onSuccess?: () => void }
  ) => {
    // Only pass options when there are some: `mutate(payload, undefined)` and
    // `mutate(payload)` behave identically, but they are not the same call, and
    // every existing caller here passes none.
    if (options) {
      updateTaskMut.mutate({ id: taskId, data: patch as Partial<ApiTask> }, options);
      return;
    }
    updateTaskMut.mutate({ id: taskId, data: patch as Partial<ApiTask> });
  };
  const addComment = (
    taskId: string,
    body: string,
    _author?: { id: string; name: string; initials: string }
  ) => {
    addCommentMut.mutate({ taskId, content: body });
  };
  const addSubtask = (taskId: string, title: string) => {
    addSubtaskMut.mutate({ taskId, title });
  };

  const [tab, setTab] = useState<'activity' | 'comments'>('comments');
  const [depId, setDepId] = useState('');

  if (isLoading) {
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.taskDetail.loading')}
      </div>
    );
  }

  const task = detail?.task;
  const taskComments = (detail?.comments ?? []).map((c) => {
    const author = members.find((member) => String(member.id) === String(c.userId));
    const authorName = author?.name || t('flow.time.unknownAuthor');
    return {
      id: c.id,
      taskId: c.taskId,
      authorId: c.userId,
      author: authorName,
      authorInitials: author?.initials || getInitials(authorName),
      body: c.content,
      createdAt: c.createdAt,
    };
  });

  const taskSubtasks = (detail?.subtasks ?? []).map((s) => ({
    id: s.id,
    taskId: task?.id ?? '',
    title: s.title,
    completed: s.status === 'done',
  }));

  const toggleSubtask = (subtaskId: string) => {
    const raw = detail?.subtasks?.find((s) => s.id === subtaskId);
    toggleSubtaskMut.mutate({ id: subtaskId, completed: raw?.status === 'done' });
  };

  // A failed request leaves `task` undefined exactly as a deleted one does,
  // so a 500 used to report the task as gone.
  if (hasQueryFailed(taskQuery)) {
    return <ListLoadFailed onRetry={() => void taskQuery.refetch()} />;
  }

  if (!task) {
    return (
      <div className="p-10 text-center">
        <p className="text-body" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.taskDetail.notFound')}
        </p>
      </div>
    );
  }

  const project = taskProject ?? projects.find((p) => p.id === task.projectId);
  const projectSections = sections
    .filter((s) => s.projectId === task.projectId)
    .sort((a, b) => a.order - b.order);
  const fieldDefs = customFieldDefinitions
    .filter((f) => f.projectId === task.projectId)
    .sort((a, b) => a.order - b.order);
  const dependencyTasks = allTasks.filter((ot) => task.dependencies?.includes(ot.id));
  const dependencyCandidates = allTasks.filter(
    (ot) =>
      ot.workspaceId === task.workspaceId &&
      ot.id !== task.id &&
      !task.dependencies?.includes(ot.id)
  );
  const apiAttachments: Attachment[] = (taskAttachmentData?.attachments ?? []).map(
    (attachment) => ({
      id: String(attachment.id),
      taskId: String(attachment.taskId),
      name: attachment.name,
      size: attachment.size,
      url: attachment.url,
      uploadedBy: String(attachment.uploadedBy),
      uploadedAt: attachment.createdAt,
      mime: inferAttachmentMime(attachment.name),
    })
  );
  const taskAttachments =
    taskAttachmentData !== undefined
      ? apiAttachments
      : attachments.filter((a) => a.taskId === task.id);

  // Labels may arrive as label_ids (snake_case from API before normalization)
  // or labels (post-mapping). Accept either shape.
  const taskWithLabelIds = task as Task & { label_ids?: number[] };
  const rawLabels = taskWithLabelIds.label_ids ?? task.labels ?? [];
  // Compare as strings. `ApiLabel` declares `id: number`, but `useLabels` casts
  // the response rather than converting it, so at runtime the id is the int64
  // wire string. Coercing the task's ids to numbers and testing them against
  // that string with Array.includes never matched, so an attached label saved
  // correctly and then never displayed.
  const taskLabelIds: string[] = rawLabels.map((l) => String(l));
  const taskLabels = allLabels.filter((l) => taskLabelIds.includes(String(l.id)));

  const cycle = cycles.find((c) => c.id === task.cycleId);
  const milestone = milestones.find((m) => m.id === task.milestoneId);

  // Totals and the entry list now live in TimePanel, which reads the time-log
  // query itself. The old store fallback went with them: it only ever applied
  // while the query was still in flight, and TimePanel has a loading state.
  const estimatedHours = task.estimatedHours ?? task.estimate ?? null;

  const priorities: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];
  const statuses = TASK_STATUSES;

  // Handle set labels via new API
  const handleToggleLabel = (labelId: number) => {
    const current: number[] = taskLabelIds.map((id) => Number(id));
    const next = current.includes(labelId)
      ? current.filter((id) => id !== labelId)
      : [...current, labelId];
    setTaskLabelsMut.mutate({ task_id: Number(task.id), label_ids: next });
    setLabelPickerOpen(false);
  };

  return (
    <div className="flex h-full min-h-0 flex-col overflow-y-auto lg:flex-row lg:overflow-hidden">
      <HelmetContainer title={`${t('flow.taskDetail.project')} · rtylr`} />
      {/* Main area. shrink-0 with an auto basis below lg: as flex-1 its basis is
          0%, so in this fixed-height scrolling column the taller rail claimed
          every pixel and the title and description collapsed out of view on a
          phone. It only shares the row with the rail from lg up. */}
      <div className="flex min-h-0 w-full min-w-0 shrink-0 flex-col lg:min-h-0 lg:flex-1 lg:overflow-hidden">
        <div
          className="flex shrink-0 items-center justify-between gap-2 px-4 py-3 sm:px-6"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div className="flex min-w-0 items-center gap-2 sm:gap-3">
            <button
              type="button"
              onClick={() => window.history.back()}
              aria-label={t('common.back')}
              className="h-6 w-6 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <ArrowLeft size={13} className="flow-directional-icon" />
            </button>
            <Breadcrumbs
              items={[
                {
                  label: project?.name || task.projectName || t('flow.taskDetail.project'),
                  to: `/p/${task.projectId}`,
                },
                { label: task.identifier ?? task.id },
              ]}
            />
          </div>
          <div className="flex shrink-0 items-center gap-0.5 sm:gap-1">
            <button
              type="button"
              onClick={() => {
                navigator.clipboard?.writeText(task.identifier ?? task.id);
                toast.success(t('toasts.task.copied', { id: task.identifier ?? task.id }));
              }}
              className="flex h-8 items-center gap-1 rounded-sm px-2 text-meta hover:bg-[var(--flow-hover)] sm:h-7"
              style={{ color: 'var(--flow-text-muted)' }}
              title={t('flow.taskDetail.copyId')}
            >
              <Copy size={11} />
              <span className="hidden sm:inline">{t('flow.taskDetail.copyId')}</span>
            </button>
            <button
              type="button"
              onClick={() => {
                updateTaskMut.mutate(
                  { id: task.id, data: { archived: true } },
                  {
                    onSuccess: () => {
                      archiveTask(task.id);
                      toast.success(t('toasts.task.archived'));
                      window.history.back();
                    },
                  }
                );
              }}
              className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
              style={{ color: 'var(--flow-text-muted)' }}
              title={!canEdit ? t('common.readOnlyAccess') : t('flow.taskDetail.archive')}
              aria-label={!canEdit ? t('common.readOnlyAccess') : t('flow.taskDetail.archive')}
              disabled={updateTaskMut.isPending || !canEdit}
            >
              <Archive size={12} />
            </button>
            <button
              type="button"
              onClick={() => setConfirmingDelete(true)}
              className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
              style={{ color: 'var(--flow-danger-fg)' }}
              title={!canRemove ? t('common.deleteRestricted') : t('flow.taskDetail.delete')}
              aria-label={!canRemove ? t('common.deleteRestricted') : t('flow.taskDetail.delete')}
              disabled={deleteTaskMut.isPending || !canRemove}
            >
              <Trash2 size={12} />
            </button>
          </div>
        </div>

        {!canEdit && (
          <div
            className="px-4 py-2 text-meta sm:px-6"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
          >
            {t('common.readOnlyAccess')}
          </div>
        )}

        <motion.div
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.18 }}
          className="min-h-0 flex-1 overflow-y-auto lg:min-h-0"
        >
          <div className="max-w-3xl mx-auto px-4 py-6 sm:px-8 sm:py-8">
            {/* Title */}
            <input
              aria-label={t('flow.taskDetail.taskTitle')}
              value={task.title}
              onChange={(e) => updateTask(task.id, { title: e.target.value })}
              disabled={!canEdit}
              // 20px at every width. S-12 asks for 20px; `sm:text-2xl` took it
              // to 24 on desktop, and the review measured 30px before that.
              className="w-full bg-transparent font-bold mb-4 text-xl disabled:opacity-70"
              style={{
                color: 'var(--flow-text)',
              }}
            />
            {/* Description */}
            <textarea
              value={displayDescription(task.description)}
              onChange={(e) => updateTask(task.id, { description: e.target.value })}
              placeholder={t('flow.taskDetail.descriptionPlaceholder')}
              rows={5}
              disabled={!canEdit}
              className="w-full bg-transparent text-body leading-relaxed mb-6 resize-none disabled:opacity-70"
              style={{ color: 'var(--flow-text-muted)' }}
            />

            {/* Subtasks */}
            <section className="mb-8">
              <div className="flex items-center gap-2 mb-3">
                <GitBranch size={12} style={{ color: 'var(--flow-text-muted)' }} />
                <h3
                  className="text-meta font-semibold uppercase tracking-wider"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t('flow.taskDetail.subtasks')}
                </h3>
              </div>
              <SubtaskList
                subtasks={taskSubtasks}
                onToggle={toggleSubtask}
                onAdd={(title) => addSubtask(task.id, title)}
                disabled={!canEdit}
              />
            </section>

            <section className="mb-8">
              <h3
                className="text-meta font-semibold uppercase tracking-wider mb-3"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.taskDetail.attachments')}
              </h3>
              <AttachmentUpload
                attachments={taskAttachments}
                onAdd={(name, size, mime, url) =>
                  addAttachmentMut.mutate({ taskId: task.id, name, size, mime, url })
                }
                onRemove={(attachmentId) =>
                  removeAttachmentMut.mutate({ taskId: task.id, attachmentId })
                }
                disabled={!canEdit}
              />
            </section>

            {/* Tabs */}
            <div
              className="flex items-center gap-1 mb-4"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              {(['comments', 'activity'] as const).map((tabName) => (
                <button
                  type="button"
                  key={tabName}
                  onClick={() => setTab(tabName)}
                  className="px-3 py-2 text-meta font-medium capitalize"
                  style={{
                    color: tab === tabName ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                    borderBottom:
                      tab === tabName ? '2px solid var(--accent)' : '2px solid transparent',
                  }}
                >
                  {tabName === 'comments'
                    ? t('flow.taskDetail.comments')
                    : t('flow.taskDetail.activity')}
                  {tabName === 'comments' && (
                    <span
                      className="ms-1.5 text-meta px-1 rounded-sm"
                      style={{
                        background: 'var(--flow-hover)',
                        color: 'var(--flow-text-dim)',
                      }}
                    >
                      {taskComments.length}
                    </span>
                  )}
                </button>
              ))}
            </div>

            {tab === 'comments' && (
              <div>
                <div className="mb-4">
                  <CommentList comments={taskComments} />
                </div>
                <CommentInput
                  onSubmit={(body) =>
                    addComment(
                      task.id,
                      body,
                      user
                        ? { id: user.id, name: user.name, initials: getInitials(user.name) }
                        : {
                            id: 'local',
                            name: t('flow.taskDetail.you'),
                            initials: getInitials(t('flow.taskDetail.you')),
                          }
                    )
                  }
                  disabled={!canEdit}
                />
              </div>
            )}

            {tab === 'activity' && (
              <ActivitySection
                entries={taskActivityData}
                t={t}
                onGoToComments={() => setTab('comments')}
              />
            )}
          </div>
        </motion.div>
      </div>

      {/* Side panel, below main on small screens */}
      <aside
        className="w-full shrink-0 overflow-y-auto border-t border-[var(--flow-border)] px-4 py-6 sm:px-5 lg:w-[256px] lg:border-s lg:border-t-0"
        style={{
          background: 'var(--flow-surface)',
        }}
      >
        <DetailField icon={CircleDot} label={t('flow.taskDetail.status')}>
          <select
            aria-label={t('flow.taskDetail.status')}
            value={task.status}
            onChange={(e) => {
              const next = e.target.value as Task['status'];
              // The celebration belongs on the mutation's success path. Firing
              // it here showed "Nice work!" next to "Failed to update task"
              // while the status reverted and nothing was saved.
              const justCompleted = next === 'done' && task.status !== 'done';
              updateTask(
                task.id,
                { status: next },
                justCompleted
                  ? {
                      onSuccess: () =>
                        toast.success(t('toasts.task.niceWork'), {
                          description: t('flow.taskDetail.taskMarkedComplete'),
                          icon: <PartyPopper size={16} />,
                        }),
                    }
                  : undefined
              );
            }}
            disabled={!canEdit}
            className="bg-transparent text-meta w-full disabled:opacity-70"
            style={{ color: 'var(--flow-text)' }}
          >
            {statuses.map((s) => (
              <option key={s} value={s} style={{ background: 'var(--flow-card)' }}>
                {t(STATUS_LABEL_KEYS[s])}
              </option>
            ))}
          </select>

          {/* One STATUS row, with the project's own statuses inside it.
              These were two adjacent rows reading `STATUS: In Progress` and
              `STATUSES: No custom status`, which is one concept presented as
              two and a label that differed from its neighbour by a letter
              (rtylr#1079, S-12).

              Still two controls, because they are two fields with two
              mutations: the built-in status drives the board column, the
              project's own is a label on top of it. Each keeps its own
              accessible name, so a screen reader announces which is which
              rather than hearing one group called "status" twice. */}
          {projectStatuses.length > 0 && (
            <select
              aria-label={t('flow.status.settingsTitle')}
              value={task.statusId ?? ''}
              onChange={(e) => {
                setTaskStatus.mutate({ taskId: task.id, statusId: e.target.value || '0' });
              }}
              disabled={!canEdit}
              className="mt-1.5 w-full bg-transparent text-meta disabled:opacity-70"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <option value="" style={{ background: 'var(--flow-card)' }}>
                {t('flow.status.none')}
              </option>
              {projectStatuses.map((s) => (
                <option key={s.id} value={s.id} style={{ background: 'var(--flow-card)' }}>
                  {s.name}
                </option>
              ))}
            </select>
          )}
        </DetailField>

        <DetailField icon={Flag} label={t('flow.taskDetail.priority')}>
          <select
            aria-label={t('flow.taskDetail.priority')}
            value={task.priority}
            onChange={(e) => updateTask(task.id, { priority: e.target.value as Task['priority'] })}
            disabled={!canEdit}
            className="bg-transparent text-meta w-full disabled:opacity-70"
            style={{ color: 'var(--flow-text)' }}
          >
            {priorities.map((p) => (
              <option key={p} value={p} style={{ background: 'var(--flow-card)' }}>
                {p === 'no_priority' ? t('flow.taskDetail.noPriority') : t(`priority.${p}`)}
              </option>
            ))}
          </select>
        </DetailField>

        <DetailField icon={UserIcon} label={t('flow.taskDetail.assignee')}>
          <div className="flex items-center gap-2">
            <AssigneeAvatar name={task.assignee} initials={task.assigneeInitials} size={18} />
            <select
              aria-label={t('flow.taskDetail.assignee')}
              value={task.assigneeId}
              onChange={(e) => {
                const m = members.find((m) => m.id === e.target.value);
                if (m)
                  updateTask(task.id, {
                    assigneeId: m.id,
                    assignee: m.name,
                    assigneeInitials: m.initials,
                  });
              }}
              disabled={!canEdit}
              className="bg-transparent text-meta flex-1 disabled:opacity-70"
              style={{ color: 'var(--flow-text)' }}
            >
              {members.map((m) => (
                <option key={m.id} value={m.id} style={{ background: 'var(--flow-card)' }}>
                  {m.name}
                </option>
              ))}
            </select>
          </div>
        </DetailField>

        {overlappingLeave && overlappingLeave.length > 0 && (
          <div
            role="status"
            style={{
              margin: '-6px 0 8px',
              padding: '6px 10px',
              borderRadius: 'var(--radius-md)',
              background: 'var(--flow-status-pending-bg)',
              color: 'var(--flow-text)',
              fontSize: 'var(--text-meta)',
            }}
          >
            {t('flow.taskDetail.leaveOverlap', { count: overlappingLeave.length })}{' '}
            <span style={{ color: 'var(--hub-text-dim)' }}>
              {overlappingLeave
                .slice(0, 3)
                .map((l) => `${l.startDate ?? '?'} → ${l.endDate ?? '?'}`)
                .join(', ')}
            </span>
          </div>
        )}

        <DetailField icon={CalendarIcon} label={t('flow.taskDetail.dueDate')}>
          <DueDatePicker
            value={task.dueDate}
            onChange={(d) => updateTask(task.id, { dueDate: d })}
            disabled={!canEdit}
          />
        </DetailField>

        <DetailField icon={Clock} label={t('flow.taskDetail.estimate')}>
          <input
            type="number"
            // The resolved value, the same one TimePanel is given. This read
            // `task.estimate` alone, so whenever the API sent `estimatedHours`
            // the field showed only its "hours" placeholder while the panel
            // beside it said "of 40:00" (rtylr#1079).
            value={estimatedHours ?? ''}
            onChange={(e) => updateTask(task.id, { estimate: Number(e.target.value) })}
            placeholder={t('flow.taskDetail.hours')}
            disabled={!canEdit}
            className="bg-transparent text-meta w-full disabled:opacity-70"
            style={{ color: 'var(--flow-text)' }}
          />
        </DetailField>

        <TimePanel
          taskId={task.id}
          estimateHours={estimatedHours}
          canLog={canEdit}
          canLogForOthers={false}
          canDelete={canRemove}
          currentUserId={user?.id ?? null}
        />

        <DetailField icon={Gauge} label={t('flow.taskDetail.effort')}>
          <div className="flex flex-wrap gap-1">
            <button
              type="button"
              onClick={() => updateTask(task.id, { effortMinutes: undefined })}
              disabled={!canEdit}
              className="text-meta px-2 py-1 rounded-md disabled:opacity-40"
              style={{
                background: !task.effortMinutes ? 'var(--flow-hover)' : 'transparent',
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
                onClick={() => updateTask(task.id, { effortMinutes: m })}
                disabled={!canEdit}
                className="text-meta px-2 py-1 rounded-md transition-colors disabled:opacity-40"
                style={{
                  background: task.effortMinutes === m ? 'var(--accent-muted)' : 'transparent',
                  color: 'var(--flow-text)',
                  border:
                    task.effortMinutes === m
                      ? `1px solid ${withAlpha('var(--accent)', 40)}`
                      : '1px solid var(--flow-border)',
                }}
                title={
                  !canEdit ? t('common.readOnlyAccess') : `${m} ${t('flow.taskDetail.minutes')}`
                }
              >
                {formatEffortMinutes(m)}
              </button>
            ))}
          </div>
        </DetailField>

        <DetailField icon={Hourglass} label={t('flow.taskDetail.waitingOn')}>
          <input
            value={task.waitingOn || ''}
            onChange={(e) => updateTask(task.id, { waitingOn: e.target.value || undefined })}
            placeholder={t('flow.taskDetail.waitingPlaceholder')}
            disabled={!canEdit}
            className="bg-transparent text-meta w-full placeholder:opacity-50 disabled:opacity-70"
            style={{ color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Lightbulb} label={t('flow.taskDetail.decision')}>
          <textarea
            value={task.decisionNote || ''}
            onChange={(e) => updateTask(task.id, { decisionNote: e.target.value || undefined })}
            placeholder={t('flow.taskDetail.decisionPlaceholder')}
            rows={2}
            disabled={!canEdit}
            className="w-full bg-transparent text-meta resize-none placeholder:opacity-50 rounded-md px-2 py-1.5 disabled:opacity-70"
            style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Pin} label={t('flow.taskDetail.today')}>
          <button
            type="button"
            onClick={() => updateTask(task.id, { focusPin: !task.focusPin })}
            disabled={!canEdit}
            title={!canEdit ? t('common.readOnlyAccess') : undefined}
            className="text-meta px-2.5 py-1.5 rounded-md w-full text-start transition-colors disabled:opacity-60"
            style={{
              background: task.focusPin ? 'var(--accent-muted)' : 'var(--flow-hover)',
              color: 'var(--flow-text)',
              border: task.focusPin
                ? `1px solid ${withAlpha('var(--accent)', 33)}`
                : '1px solid transparent',
            }}
          >
            {task.focusPin ? t('flow.taskDetail.pinnedToToday') : t('flow.taskDetail.pinToToday')}
          </button>
        </DetailField>

        {cycle && (
          <DetailField icon={Timer} label={t('flow.taskDetail.cycle')}>
            <span className="text-meta" style={{ color: 'var(--flow-text)' }}>
              {cycle.name}
            </span>
          </DetailField>
        )}

        {milestone && (
          <DetailField icon={MilestoneIcon} label={t('flow.taskDetail.milestone')}>
            <span className="text-meta" style={{ color: 'var(--flow-text)' }}>
              {milestone.name}
            </span>
          </DetailField>
        )}

        {projectSections.length > 0 && (
          <DetailField icon={LayoutList} label={t('flow.taskDetail.section')}>
            <select
              aria-label={t('flow.taskDetail.section')}
              value={task.sectionId ?? ''}
              onChange={(e) => updateTask(task.id, { sectionId: e.target.value || undefined })}
              disabled={!canEdit}
              className="bg-transparent text-meta w-full disabled:opacity-70"
              style={{ color: 'var(--flow-text)' }}
            >
              <option value="">{t('flow.taskDetail.noSection')}</option>
              {projectSections.map((s) => (
                <option key={s.id} value={s.id} style={{ background: 'var(--flow-card)' }}>
                  {s.name}
                </option>
              ))}
            </select>
          </DetailField>
        )}

        {fieldDefs.map((def) => (
          <DetailField key={def.id} icon={ListChecks} label={def.name}>
            {def.type === 'dropdown' ? (
              <select
                aria-label={def.name}
                value={String(task.customFieldValues?.[def.id] ?? '')}
                onChange={(e) =>
                  updateTask(task.id, {
                    customFieldValues: {
                      ...(task.customFieldValues ?? {}),
                      [def.id]: e.target.value,
                    },
                  })
                }
                disabled={!canEdit}
                className="bg-transparent text-meta w-full disabled:opacity-70"
                style={{ color: 'var(--flow-text)' }}
              >
                <option value="">-</option>
                {(def.options ?? []).map((opt) => (
                  <option key={opt} value={opt} style={{ background: 'var(--flow-card)' }}>
                    {opt}
                  </option>
                ))}
              </select>
            ) : def.type === 'number' ? (
              <input
                aria-label={def.name}
                type="number"
                value={task.customFieldValues?.[def.id] ?? ''}
                onChange={(e) =>
                  updateTask(task.id, {
                    customFieldValues: {
                      ...(task.customFieldValues ?? {}),
                      [def.id]: e.target.value === '' ? '' : Number(e.target.value),
                    },
                  })
                }
                disabled={!canEdit}
                className="bg-transparent text-meta w-full disabled:opacity-70"
                style={{ color: 'var(--flow-text)' }}
              />
            ) : (
              <input
                aria-label={def.name}
                value={String(task.customFieldValues?.[def.id] ?? '')}
                onChange={(e) =>
                  updateTask(task.id, {
                    customFieldValues: {
                      ...(task.customFieldValues ?? {}),
                      [def.id]: e.target.value,
                    },
                  })
                }
                disabled={!canEdit}
                className="bg-transparent text-meta w-full disabled:opacity-70"
                style={{ color: 'var(--flow-text)' }}
              />
            )}
          </DetailField>
        ))}

        <DetailField icon={Link2} label={t('flow.taskDetail.dependencies')}>
          <div className="space-y-1.5">
            {dependencyTasks.map((dt) => (
              <div key={dt.id} className="flex items-center justify-between gap-1">
                <button
                  type="button"
                  onClick={() => {
                    // In the drawer, following a dependency swaps which task the
                    // drawer shows. On the standalone page there is no drawer to
                    // swap, so it navigates (rtylr#1079).
                    if (taskId !== undefined) {
                      openTask(String(dt.id));
                      return;
                    }
                    void navigate({ to: '/task/$id', params: { id: String(dt.id) } });
                  }}
                  className="text-meta text-start truncate flex-1 hover:underline"
                  style={{ color: 'var(--accent)' }}
                >
                  {dt.identifier} · {dt.title}
                </button>
                {canEdit && (
                  <button
                    type="button"
                    className="text-meta shrink-0"
                    style={{ color: 'var(--flow-text-dim)' }}
                    onClick={() =>
                      updateTask(task.id, {
                        dependencies: (task.dependencies ?? []).filter((x) => x !== dt.id),
                      })
                    }
                  >
                    {t('flow.taskDetail.remove')}
                  </button>
                )}
              </div>
            ))}
            {dependencyCandidates.length > 0 && canEdit && (
              <div className="flex gap-1 pt-1">
                <select
                  aria-label={t('flow.taskDetail.addDependency')}
                  value={depId}
                  onChange={(e) => setDepId(e.target.value)}
                  className="bg-transparent text-meta flex-1 min-w-0 rounded-sm px-1"
                  style={{ color: 'var(--flow-text)', border: '1px solid var(--flow-border)' }}
                >
                  <option value="">{t('flow.taskDetail.addDependency')}</option>
                  {dependencyCandidates.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.identifier} {c.title}
                    </option>
                  ))}
                </select>
                <button
                  type="button"
                  disabled={!depId}
                  onClick={() => {
                    if (!depId) return;
                    updateTask(task.id, {
                      dependencies: [...(task.dependencies ?? []), depId],
                    });
                    setDepId('');
                  }}
                  className="text-meta px-1.5 py-0.5 rounded-sm font-semibold disabled:opacity-40"
                  style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
                >
                  {t('flow.taskDetail.add')}
                </button>
              </div>
            )}
          </div>
        </DetailField>

        <DetailField icon={Tag} label={t('flow.taskDetail.labels')}>
          <div className="flex flex-wrap gap-1 relative">
            {taskLabels.length === 0 && canEdit && (
              <button
                type="button"
                onClick={() => setLabelPickerOpen((v) => !v)}
                className="text-meta px-1.5 py-0.5 rounded-sm hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                {t('flow.taskDetail.addLabel')}
              </button>
            )}
            {taskLabels.map((l) => (
              <LabelChip
                key={l.id}
                label={{ id: String(l.id), name: l.name, color: l.color }}
                small
                onRemove={canEdit ? () => handleToggleLabel(l.id) : undefined}
              />
            ))}
            {taskLabels.length > 0 && canEdit && (
              <button
                type="button"
                onClick={() => setLabelPickerOpen((v) => !v)}
                className="text-meta px-1.5 py-0.5 rounded-sm hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                {t('flow.taskDetail.addShort')}
              </button>
            )}
            {labelPickerOpen && (
              <div
                className="absolute top-6 start-0 z-50 rounded-lg shadow-lg p-2 min-w-[160px] space-y-0.5"
                style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
              >
                {allLabels.length === 0 && (
                  <p className="text-meta px-2 py-1" style={{ color: 'var(--flow-text-dim)' }}>
                    {t('flow.taskDetail.noLabels')}
                  </p>
                )}
                {allLabels.map((l) => {
                  const active = taskLabelIds.includes(String(l.id));
                  return (
                    <button
                      type="button"
                      key={l.id}
                      onClick={() => handleToggleLabel(l.id)}
                      className="w-full flex items-center gap-2 px-2 py-1 rounded-sm text-start hover:bg-[var(--flow-hover)]"
                    >
                      <span
                        className="h-2 w-2 rounded-full shrink-0"
                        style={{ background: l.color }}
                      />
                      <span
                        className="text-meta flex-1"
                        style={{
                          color: active ? l.color : 'var(--flow-text)',
                          fontWeight: active ? 600 : 400,
                        }}
                      >
                        {l.name}
                      </span>
                      {active && (
                        <span className="text-meta" style={{ color: l.color }}>
                          ✓
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </DetailField>
      </aside>
      <ConfirmDialog
        open={confirmingDelete}
        onOpenChange={setConfirmingDelete}
        title={t('flow.taskMenu.deleteTitle')}
        description={
          <>
            {t('flow.taskMenu.deleteConfirmPrefix')} <strong>{task.title}</strong>
            {t('flow.taskMenu.deleteConfirmSuffix')}
          </>
        }
        confirmLabel={t('common.delete')}
        cancelLabel={t('common.cancel')}
        pending={deleteTaskMut.isPending}
        onConfirm={() => {
          // Wait for the mutation to actually succeed before toasting and
          // navigating away: on failure the user would otherwise see a success
          // toast, leave the page, and the task would still be alive in cache.
          deleteTaskMut.mutate(task.id, {
            onSuccess: () => {
              toast.success(t('toasts.task.deleted'));
              setConfirmingDelete(false);
              window.history.back();
            },
            // useDeleteTask already toasts on error.
          });
        }}
      />
    </div>
  );
};

// ── Activity Section ──────────────────────────────────────────────────────────

const ActivitySection: React.FC<{
  entries: ApiTaskActivity[];
  t: TFunction;
  onGoToComments: () => void;
}> = ({ entries, t, onGoToComments }) => {
  if (entries.length === 0) {
    return (
      <EmptyState
        icon={Activity}
        title={t('flow.taskDetail.noActivity')}
        action={{
          label: t('flow.taskDetail.goToComments'),
          onClick: onGoToComments,
        }}
        size="compact"
      />
    );
  }
  return (
    <div className="space-y-3">
      {entries.map((a) => {
        const Icon = activityIcon(a.action);
        return (
          <div key={a.id} className="flex items-start gap-2.5">
            <span
              className="h-5 w-5 rounded-full flex items-center justify-center shrink-0 mt-0.5"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
            >
              <Icon size={10} />
            </span>
            <div className="flex-1 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {describeActivity(a, t)}
              <span className="ms-2" style={{ color: 'var(--flow-text-dim)' }}>
                {formatRelative(a.createdAt, t)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
};

// ── DetailField ───────────────────────────────────────────────────────────────

const DetailField: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  label: string;
  children: React.ReactNode;
}> = ({ icon: Icon, label, children }) => (
  <div className="mb-5">
    <div className="flex items-center gap-1.5 mb-1.5">
      <Icon size={11} />
      {/* muted, not dim: --flow-text-dim measures 2.24:1 and never carries
       * text. TimePanel's own heading uses the same treatment, so leaving this
       * dim would put two versions of one label style side by side. */}
      <span
        className="text-meta uppercase tracking-wider font-semibold"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {label}
      </span>
    </div>
    <div>{children}</div>
  </div>
);

export default TaskDetail;
