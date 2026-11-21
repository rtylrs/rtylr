import React, { useState } from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { useNavigate, useParams } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
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
  Plus,
  Link2,
  Gauge,
  Hourglass,
  Pin,
  Lightbulb,
  PartyPopper,
  UserPen,
  UserPlus,
  Printer,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import { useTask } from './useTask';
import {
  useTasks,
  useAddTaskDependency,
  useRemoveTaskDependency,
  useUpdateTask,
  useDeleteTask,
  useAddComment,
  useUpdateComment,
  useAddSubtask,
  useToggleSubtask,
  useTaskAttachments,
  useAddAttachment,
  useRemoveAttachment,
} from './useTasks';
import { useCycles } from '../cycles/useCycles';
import { useMilestones } from '../milestones/useMilestones';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { useLabels, useSetTaskLabels } from './useLabels';
import { useCreateStatus, useProjectStatuses, useSetTaskStatus } from '../projects/useStatuses';
import { displayStatusName, insertionPlan, sortStatuses } from '../projects/statusUtils';
import NewStatusDialog from '../projects/components/NewStatusDialog';
import { useTaskActivity } from './useTaskActivity';
import { useTaskRealtime } from '../../shared/lib/useTaskRealtime';
import { useCurrentProject } from '../projects/useCurrentProject';
import { useApprovedLeaveInRange } from '../../shared/hooks/queries/useApprovedLeaveInRange';
import DescriptionEditor from './components/DescriptionEditor';
import TimePanel from './components/TimePanel';
import { useFlowStore, Attachment, Task } from '../../shared/stores/flowStore';
import type { ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';
import { useUserStore } from '../../shared/stores/userStore';
import { useCanEditTask, useCanDeleteResource } from '../../shared/lib/currentUser';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import CommentList from '../../shared/components/CommentList';
import CommentInput, { type MentionCandidate } from '../../shared/components/CommentInput';
import { useProjectMembers } from '../projects/useProjectMembers';
import SubtaskList from '../../shared/components/SubtaskList';
import AttachmentUpload from '../../shared/components/AttachmentUpload';
import AssigneeAvatar from '../../shared/components/AssigneeAvatar';
import FlowSelect from '../../shared/components/FlowSelect';
import PriorityIcon from '../../shared/components/PriorityIcon';
import LabelChip from '../../shared/components/LabelChip';
import DueDatePicker from '../../shared/components/DueDatePicker';
import TaskTextField from '../../shared/components/TaskTextField';
import { getInitials } from '../../shared/lib/utils';
import { displayUserName } from '../../shared/lib/userDisplay';
import { priorityLabelKey } from '../../shared/lib/priorityLabels';
import { assigneeOptions } from '../../shared/lib/memberUtils';
import { displayDescription } from '../../shared/lib/stringUtils';
import { EFFORT_PRESETS, formatEffortMinutes } from '../../shared/lib/effortPresets';
import { TASK_STATUSES, STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import { withAlpha } from '../../shared/lib/colorUtils';
import { toLocalIso, todayIso } from '../../shared/lib/dateUtils';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import DevelopmentSection from './components/DevelopmentSection';
import { useIsDeveloperProject } from '../projects-v2/useDeveloperMode';
import LinksSection from './components/LinksSection';
import TaskCustomFields from './components/TaskCustomFields';
import TaskActivityTabs from './components/TaskActivityTabs';
import TaskPrintView from './components/TaskPrintView';

// ── Helpers ──────────────────────────────────────────────────────────────────

function inferAttachmentMime(name: string): string {
  if (/\.(apng|avif|gif|jpe?g|png|svg|webp)$/i.test(name)) {
    return 'image/*';
  }
  return 'application/octet-stream';
}

// ── Component ────────────────────────────────────────────────────────────────

/**
 * `taskId` overrides the route param, so the same component serves the
 * standalone `/flow/task/$id` page and the drawer that opens over a list
 * (rtylr#1079). The drawer knows which task it is showing; the page reads it
 * from the URL.
 */
const TaskDetail: React.FC<{ taskId?: string }> = ({ taskId }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const openTask = useOpenTask();
  const { id: routeId } = useParams({ strict: false });
  const id = taskId ?? routeId;
  const closeTask = () => {
    if (taskId) {
      void navigate({
        to: '.',
        search: (previous: Record<string, unknown>) => {
          const { openTask: _openTask, ...rest } = previous;
          return rest;
        },
      });
    } else {
      void navigate({ to: '/flow/my-issues' });
    }
  };
  useTaskRealtime(id);
  const taskQuery = useTask(id);
  const { data: taskData } = taskQuery;
  const taskLoadFailed = hasQueryFailed(taskQuery);
  const isLoading = taskQuery.isPending && !taskLoadFailed;
  const user = useUserStore((s) => s.user);
  // Read up here rather than below the isLoading early return: these gates are
  // hooks, so they have to run on every render, and they need the task's project.
  const detail = taskData as ApiTaskDetail | undefined;
  const canEdit = useCanEditTask(detail?.task?.projectId);
  // The project's statuses are its board columns; while they load (or for a
  // task whose project has none) the built-in statuses stand in.
  const { data: rawProjectStatuses = [] } = useProjectStatuses(detail?.task?.projectId ?? '');
  const projectStatuses = sortStatuses(rawProjectStatuses);
  const setTaskStatus = useSetTaskStatus();
  const createStatus = useCreateStatus(detail?.task?.projectId ?? '');
  const [newStatusOpen, setNewStatusOpen] = useState(false);
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
  const updateCommentMut = useUpdateComment();
  const addSubtaskMut = useAddSubtask();
  const toggleSubtaskMut = useToggleSubtask();
  const setTaskLabelsMut = useSetTaskLabels();
  const addDependency = useAddTaskDependency();
  const removeDependency = useRemoveTaskDependency();

  const { data: taskActivityData = [] } = useTaskActivity(id ?? '');
  const { data: taskAttachmentData } = useTaskAttachments(id);
  const addAttachmentMut = useAddAttachment();
  const removeAttachmentMut = useRemoveAttachment();

  const taskProjectId = detail?.task?.projectId;
  const isDeveloperProject = useIsDeveloperProject(taskProjectId);
  const { project: taskProject } = useCurrentProject(taskProjectId);
  const { data: projectMembers = [] } = useProjectMembers(taskProjectId);
  const { data: allLabels = [] } = useLabels();
  const tasksQuery = useTasks({});
  const allTasks = (tasksQuery.data ?? []).map(apiTaskToTask).filter((task) => !task.archived);
  const { data: cycles = [] } = useCycles(taskProjectId);
  const milestonesQuery = useMilestones(taskProjectId ?? '');
  const { data: milestones = [] } = milestonesQuery;
  const members = useFlowStore((state) => state.members);
  const projects = useFlowStore((state) => state.projects);

  const [labelPickerOpen, setLabelPickerOpen] = useState(false);
  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [printing, setPrinting] = useState(false);

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

  const [depId, setDepId] = useState('');

  if (isLoading) {
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.taskDetail.loading')}
      </div>
    );
  }

  const task = detail?.task;
  // Names by user id. The business team list needs the team-view permission,
  // so the project's own member list fills in for everyone else; without it a
  // comment read "Unknown" and an attachment named its uploader by raw id.
  const peopleNames: Record<string, string> = {};
  for (const member of projectMembers) {
    if (member.name) peopleNames[String(member.userId)] = member.name;
  }
  for (const member of members) {
    if (member.name) peopleNames[String(member.id)] = member.name;
  }
  const currentUserId = user?.id === undefined ? undefined : String(user.id);
  // The signed-in person is not always on either list (no team-view access,
  // not a project member yet), yet they are often the reporter.
  const currentUserName = displayUserName(user);
  if (currentUserId !== undefined && currentUserName && !peopleNames[currentUserId]) {
    peopleNames[currentUserId] = currentUserName;
  }
  // The people a comment here can mention: the project's members, who are the
  // ones the server notifies when their name follows an "@".
  const mentionCandidates: MentionCandidate[] = projectMembers
    .filter((member) => member.name && String(member.userId) !== currentUserId)
    .map((member) => ({
      id: String(member.userId),
      name: member.name,
      initials: getInitials(member.name),
    }))
    .sort((a, b) => a.name.localeCompare(b.name));
  const taskComments = (detail?.comments ?? []).map((c) => {
    const authorName =
      c.actorName?.trim() || peopleNames[String(c.userId)] || t('flow.time.unknownAuthor');
    return {
      id: c.id,
      taskId: c.taskId,
      authorId: c.userId,
      author: authorName,
      authorInitials: c.actorInitials || getInitials(authorName),
      body: c.content,
      createdAt: c.createdAt,
      edited: c.isEdited,
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
  if (taskLoadFailed) {
    return <ListLoadFailed onRetry={() => void taskQuery.refetch()} />;
  }

  if (!task) {
    return (
      <div className="p-10 text-center">
        <p className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.taskDetail.notFound')}
        </p>
      </div>
    );
  }

  const { assignee, assigneeInitials } = apiTaskToTask(task);
  // Reporter comes as an id on the task. It resolves through the same people
  // map comments use, so it costs no request of its own.
  const reporterId = task.reporterId && task.reporterId !== '0' ? String(task.reporterId) : '';
  const reporterName = reporterId ? (peopleNames[reporterId] ?? t('flow.time.unknownAuthor')) : '';
  // Hidden rather than disabled: it is a shortcut, and to a viewer, or to the
  // person already assigned, it has nothing to offer.
  const canAssignToMe =
    canEdit && currentUserId !== undefined && String(task.assigneeId ?? '') !== currentUserId;
  const assignToMe = (): void => {
    if (currentUserId === undefined) return;
    updateTask(task.id, {
      assigneeId: currentUserId,
      assignee: currentUserName,
      assigneeInitials: getInitials(currentUserName),
    });
  };
  const project = taskProject ?? projects.find((p) => p.id === task.projectId);
  const dependencyTasks = allTasks.filter((ot) => task.dependencies?.includes(ot.id));
  const dependencyCandidates = allTasks.filter(
    (ot) => ot.id !== task.id && !task.dependencies?.includes(ot.id)
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
  const taskAttachments = apiAttachments;

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
  const currentStatusId = String(task.statusId || task.sectionId || '');
  const celebrate = (): void => {
    toast.success(t('flowToasts.task.niceWork'), {
      description: t('flow.taskDetail.taskMarkedComplete'),
      icon: <PartyPopper size={16} />,
    });
  };
  // A status made from the task panel is an in-flight state, appended at the end.
  const createAndApplyStatus = (name: string): void => {
    createStatus.mutate(
      { name, bucket: 'active', ...insertionPlan(projectStatuses, projectStatuses.length) },
      {
        onSuccess: (created) => {
          setNewStatusOpen(false);
          if (created?.id) setTaskStatus.mutate({ taskId: task.id, statusId: String(created.id) });
        },
      }
    );
  };

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
    <div className="flow-task-detail flex h-full min-h-0 flex-col overflow-y-auto lg:flex-row lg:overflow-hidden">
      <HelmetContainer title={task.title} />
      {/* Main area. shrink-0 with an auto basis below lg: as flex-1 its basis is
          0%, so in this fixed-height scrolling column the taller rail claimed
          every pixel and the title and description collapsed out of view on a
          phone. It only shares the row with the rail from lg up. */}
      <div className="flex min-h-0 w-full min-w-0 shrink-0 flex-col lg:min-h-0 lg:flex-1 lg:overflow-hidden">
        <div
          className="flow-task-detail-toolbar flex shrink-0 items-center justify-between gap-2 px-4 py-3 sm:px-6"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div className="flex min-w-0 items-center gap-2 sm:gap-3">
            <button
              type="button"
              onClick={() => closeTask()}
              aria-label={t('flowCommon.back')}
              className="h-6 w-6 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <ArrowLeft size={13} className="flow-directional-icon" />
            </button>
            <Breadcrumbs
              items={[
                {
                  label: project?.name || task.projectName || t('flow.taskDetail.project'),
                  to: `/flow/p/${task.projectId}`,
                },
                { label: task.identifier ?? task.id },
              ]}
            />
          </div>
          <div className="flex shrink-0 items-center gap-0.5 sm:gap-1">
            {canAssignToMe && (
              <button
                type="button"
                onClick={assignToMe}
                disabled={updateTaskMut.isPending}
                className="flex h-8 items-center gap-1 rounded-sm px-2 text-flow-meta hover:bg-[var(--flow-hover)] disabled:opacity-40 sm:h-7"
                style={{ color: 'var(--flow-text-muted)' }}
                title={t('flow.taskDetail.assignToMe')}
                aria-label={t('flow.taskDetail.assignToMe')}
              >
                <UserPlus size={11} aria-hidden="true" />
                <span className="hidden sm:inline">{t('flow.taskDetail.assignToMe')}</span>
              </button>
            )}
            <button
              type="button"
              onClick={async () => {
                try {
                  await navigator.clipboard.writeText(task.identifier ?? task.id);
                  toast.success(t('flowToasts.task.copied', { id: task.identifier ?? task.id }));
                } catch {
                  toast.error(t('common.copyFailed'));
                }
              }}
              className="flex h-8 items-center gap-1 rounded-sm px-2 text-flow-meta hover:bg-[var(--flow-hover)] sm:h-7"
              style={{ color: 'var(--flow-text-muted)' }}
              title={t('flow.taskDetail.copyId')}
            >
              <Copy size={11} />
              <span className="hidden sm:inline">{t('flow.taskDetail.copyId')}</span>
            </button>
            <button
              type="button"
              onClick={() => setPrinting(true)}
              disabled={printing}
              // Not on phones: the header is full there, the task key was
              // squeezed to one letter, and printing from a phone is rare.
              className="hidden h-7 w-7 rounded-sm items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40 sm:flex"
              style={{ color: 'var(--flow-text-muted)' }}
              title={t('flow.taskDetail.print.action')}
              aria-label={t('flow.taskDetail.print.action')}
            >
              <Printer size={12} aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={() => {
                updateTaskMut.mutate(
                  { id: task.id, data: { archived: true } },
                  {
                    onSuccess: () => {
                      toast.success(t('flowToasts.task.archived'));
                      closeTask();
                    },
                  }
                );
              }}
              className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
              style={{ color: 'var(--flow-text-muted)' }}
              title={!canEdit ? t('flowCommon.readOnlyAccess') : t('flow.taskDetail.archive')}
              aria-label={!canEdit ? t('flowCommon.readOnlyAccess') : t('flow.taskDetail.archive')}
              disabled={updateTaskMut.isPending || !canEdit}
            >
              <Archive size={12} />
            </button>
            <button
              type="button"
              onClick={() => setConfirmingDelete(true)}
              className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
              style={{ color: 'var(--flow-danger-fg)' }}
              title={!canRemove ? t('flowCommon.deleteRestricted') : t('flow.taskDetail.delete')}
              aria-label={
                !canRemove ? t('flowCommon.deleteRestricted') : t('flow.taskDetail.delete')
              }
              disabled={deleteTaskMut.isPending || !canRemove}
            >
              <Trash2 size={12} />
            </button>
          </div>
        </div>

        {!canEdit && (
          <div
            className="px-4 py-2 text-flow-meta sm:px-6"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
          >
            {t('flowCommon.readOnlyAccess')}
          </div>
        )}

        <motion.div
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.18 }}
          className="min-h-0 flex-1 overflow-y-auto lg:min-h-0"
        >
          <div className="flow-task-detail-body max-w-3xl mx-auto px-4 py-6 sm:px-8 sm:py-8">
            {/* Title */}
            <TaskTextField
              aria-label={t('flow.taskDetail.taskTitle')}
              key={`${task.id}-title`}
              wrap
              value={task.title}
              required
              onSave={(title) => updateTask(task.id, { title })}
              disabled={!canEdit}
              // 20px at every width. S-12 asks for 20px; `sm:text-2xl` took it
              // to 24 on desktop, and the review measured 30px before that.
              className="w-full bg-transparent font-bold mb-4 text-xl resize-none disabled:opacity-70"
              style={{
                color: 'var(--flow-text)',
              }}
            />
            {/* Description */}
            <DescriptionEditor
              key={`${task.id}-description`}
              value={displayDescription(task.description)}
              onSave={(description) => updateTask(task.id, { description })}
              placeholder={t('flow.taskDetail.descriptionPlaceholder')}
              disabled={!canEdit}
            />

            {/* Subtasks */}
            <section className="mb-8">
              <div className="flex items-center gap-2 mb-3">
                <GitBranch size={12} style={{ color: 'var(--flow-text-muted)' }} />
                <h3
                  className="text-flow-meta font-semibold uppercase tracking-wider"
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

            {isDeveloperProject && <DevelopmentSection taskId={task.id} />}

            <section className="mb-8">
              <h3
                className="text-flow-meta font-semibold uppercase tracking-wider mb-3"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.taskDetail.attachments')}
              </h3>
              <AttachmentUpload
                attachments={taskAttachments}
                peopleNames={peopleNames}
                // Matches the server: your own uploads, or anyone's as a
                // project admin.
                canRemove={(attachment) =>
                  canRemove || String(attachment.uploadedBy) === currentUserId
                }
                onAdd={(name, size, mime, url) =>
                  addAttachmentMut.mutate({ taskId: task.id, name, size, mime, url })
                }
                onRemove={(attachmentId) =>
                  removeAttachmentMut.mutate({ taskId: task.id, attachmentId })
                }
                disabled={!canEdit}
              />
            </section>

            <LinksSection taskId={task.id} canEdit={canEdit} peopleNames={peopleNames} />

            <TaskActivityTabs
              taskId={task.id}
              projectId={task.projectId}
              currentUserId={currentUserId}
              comments={taskComments}
              activity={taskActivityData}
              peopleNames={peopleNames}
              renderComments={(comments) => (
                <CommentList
                  comments={comments}
                  currentUserId={currentUserId}
                  mentionCandidates={mentionCandidates}
                  mentionNames={Object.values(peopleNames)}
                  onEdit={
                    canEdit
                      ? (commentId, body) =>
                          updateCommentMut.mutate({ id: commentId, content: body })
                      : undefined
                  }
                />
              )}
              composer={
                <CommentInput
                  mentionCandidates={mentionCandidates}
                  onSubmit={(body) =>
                    addComment(
                      task.id,
                      body,
                      user
                        ? {
                            id: String(user.id ?? ''),
                            name: user.fullName ?? '',
                            initials: getInitials(user.fullName ?? ''),
                          }
                        : {
                            id: 'local',
                            name: t('flow.taskDetail.you'),
                            initials: getInitials(t('flow.taskDetail.you')),
                          }
                    )
                  }
                  disabled={!canEdit}
                />
              }
            />
          </div>
        </motion.div>
      </div>

      {/* Side panel, below main on small screens */}
      <aside
        className="flow-task-detail-properties w-full shrink-0 overflow-y-auto border-t border-[var(--flow-border)] px-4 py-6 sm:px-5 lg:w-[256px] lg:border-s lg:border-t-0"
        style={{
          background: 'var(--flow-surface)',
        }}
      >
        <DetailField icon={CircleDot} label={t('flow.taskDetail.status')}>
          {projectStatuses.length > 0 ? (
            <FlowSelect
              label={t('flow.taskDetail.status')}
              value={currentStatusId}
              options={projectStatuses.map((s) => ({
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
              onChange={(value) => {
                const next = projectStatuses.find((s) => s.id === value);
                const current = projectStatuses.find((s) => s.id === currentStatusId);
                // Celebrate only once the server has the change; see below.
                const justCompleted = next?.bucket === 'done' && current?.bucket !== 'done';
                setTaskStatus.mutate(
                  { taskId: task.id, statusId: value },
                  justCompleted ? { onSuccess: celebrate } : undefined
                );
              }}
              disabled={!canEdit}
              action={
                canEdit && canRemove
                  ? {
                      label: t('flow.status.newStatusAction'),
                      icon: <Plus size={13} aria-hidden="true" className="shrink-0" />,
                      onSelect: () => setNewStatusOpen(true),
                    }
                  : undefined
              }
            />
          ) : (
            <FlowSelect
              label={t('flow.taskDetail.status')}
              value={task.status}
              options={statuses.map((s) => ({ value: s, label: t(STATUS_LABEL_KEYS[s]) }))}
              onChange={(value) => {
                const next = value as Task['status'];
                // The celebration belongs on the mutation's success path. Firing
                // it here showed "Nice work!" next to "Failed to update task"
                // while the status reverted and nothing was saved.
                const justCompleted = next === 'done' && task.status !== 'done';
                updateTask(
                  task.id,
                  { status: next },
                  justCompleted ? { onSuccess: celebrate } : undefined
                );
              }}
              disabled={!canEdit}
            />
          )}
        </DetailField>

        <DetailField icon={Flag} label={t('flow.taskDetail.priority')}>
          <FlowSelect
            label={t('flow.taskDetail.priority')}
            value={task.priority}
            options={priorities.map((p) => ({
              value: p,
              label: p === 'no_priority' ? t('flow.taskDetail.noPriority') : t(`flowPriority.${p}`),
              icon: (
                <span aria-hidden="true">
                  <PriorityIcon priority={p} />
                </span>
              ),
            }))}
            onChange={(value) => updateTask(task.id, { priority: value as Task['priority'] })}
            disabled={!canEdit}
          />
        </DetailField>

        <DetailField icon={UserIcon} label={t('flow.taskDetail.assignee')}>
          <FlowSelect
            label={t('flow.taskDetail.assignee')}
            value={task.assigneeId ? String(task.assigneeId) : ''}
            leading={
              <AssigneeAvatar
                name={assignee}
                initials={assigneeInitials}
                seed={task.assigneeId ? String(task.assigneeId) : undefined}
                size={18}
              />
            }
            options={assigneeOptions(
              members,
              { id: task.assigneeId ? String(task.assigneeId) : '', name: assignee },
              t('flow.quickCreate.unassigned')
            ).map((option) => ({ value: option.id, label: option.name }))}
            onChange={(value) => {
              const m = members.find((m) => String(m.id) === value);
              updateTask(task.id, {
                assigneeId: m?.id ?? '',
                assignee: m?.name ?? '',
                assigneeInitials: m?.initials ?? '',
              });
            }}
            disabled={!canEdit}
          />
        </DetailField>

        {reporterId && (
          <DetailField icon={UserPen} label={t('flow.taskDetail.reporter')}>
            <div className="flex min-w-0 items-center gap-2">
              <AssigneeAvatar
                name={reporterName}
                initials={getInitials(reporterName)}
                seed={reporterId}
                size={18}
              />
              <span className="truncate text-flow-meta" style={{ color: 'var(--flow-text)' }}>
                {reporterName}
              </span>
            </div>
          </DetailField>
        )}

        {overlappingLeave && overlappingLeave.length > 0 && (
          <div
            role="status"
            style={{
              margin: '-6px 0 8px',
              padding: '6px 10px',
              borderRadius: 'var(--radius-md)',
              background: 'var(--flow-status-pending-bg)',
              color: 'var(--flow-text)',
              fontSize: 'var(--text-flow-meta)',
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
          <TaskTextField
            type="number"
            min={0}
            step="any"
            value={estimatedHours == null ? '' : String(estimatedHours)}
            onSave={(value) => {
              const estimate = Number(value);
              if (Number.isFinite(estimate) && estimate >= 0) updateTask(task.id, { estimate });
            }}
            placeholder={t('flow.taskDetail.hours')}
            disabled={!canEdit}
            className="bg-transparent text-flow-meta w-full disabled:opacity-70"
            style={{ color: 'var(--flow-text)' }}
          />
        </DetailField>

        <TimePanel
          taskId={task.id}
          estimateHours={estimatedHours}
          canLog={canEdit}
          canLogForOthers={false}
          canDelete={canRemove}
          currentUserId={user?.id === undefined ? null : String(user.id)}
        />

        <DetailField icon={Gauge} label={t('flow.taskDetail.effort')}>
          <div className="flex flex-wrap gap-1">
            <button
              type="button"
              onClick={() => updateTask(task.id, { effortMinutes: 0, estimate: 0 })}
              disabled={!canEdit}
              className="text-flow-meta px-2 py-1 rounded-md disabled:opacity-40"
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
                className="text-flow-meta px-2 py-1 rounded-md transition-colors disabled:opacity-40"
                style={{
                  background: task.effortMinutes === m ? 'var(--accent-muted)' : 'transparent',
                  color: 'var(--flow-text)',
                  border:
                    task.effortMinutes === m
                      ? `1px solid ${withAlpha('var(--accent)', 40)}`
                      : '1px solid var(--flow-border)',
                }}
                title={
                  !canEdit ? t('flowCommon.readOnlyAccess') : `${m} ${t('flow.taskDetail.minutes')}`
                }
              >
                {formatEffortMinutes(m)}
              </button>
            ))}
          </div>
        </DetailField>

        <DetailField icon={Hourglass} label={t('flow.taskDetail.waitingOn')}>
          <TaskTextField
            key={`${task.id}-waiting`}
            value={task.waitingOn || ''}
            onSave={(waitingOn) => updateTask(task.id, { waitingOn })}
            placeholder={t('flow.taskDetail.waitingPlaceholder')}
            disabled={!canEdit}
            className="bg-transparent text-flow-meta w-full placeholder:opacity-50 disabled:opacity-70"
            style={{ color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Lightbulb} label={t('flow.taskDetail.decision')}>
          <TaskTextField
            key={`${task.id}-decision`}
            multiline
            value={task.decisionNote || ''}
            onSave={(decisionNote) => updateTask(task.id, { decisionNote })}
            placeholder={t('flow.taskDetail.decisionPlaceholder')}
            rows={2}
            disabled={!canEdit}
            className="w-full bg-transparent text-flow-meta resize-none placeholder:opacity-50 rounded-md px-2 py-1.5 disabled:opacity-70"
            style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Pin} label={t('flow.taskDetail.today')}>
          <button
            type="button"
            onClick={() => updateTask(task.id, { focusPin: !task.focusPin })}
            disabled={!canEdit}
            title={!canEdit ? t('flowCommon.readOnlyAccess') : undefined}
            className="text-flow-meta px-2.5 py-1.5 rounded-md w-full text-start transition-colors disabled:opacity-60"
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
            <span className="text-flow-meta" style={{ color: 'var(--flow-text)' }}>
              {cycle.name}
            </span>
          </DetailField>
        )}

        {hasQueryFailed(milestonesQuery) ? (
          <div role="alert" className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
            <span>{t('flow.loadFailed.description')} </span>
            <button
              type="button"
              onClick={() => void milestonesQuery.refetch()}
              className="flow-focus underline"
              style={{ color: 'var(--flow-text)' }}
            >
              {t('flow.loadFailed.retry')}
            </button>
          </div>
        ) : (
          <DetailField icon={MilestoneIcon} label={t('flow.taskDetail.milestone')}>
            <FlowSelect
              label={t('flow.taskDetail.milestone')}
              value={milestone?.id ?? ''}
              options={
                milestones.length === 0
                  ? []
                  : [
                      { value: '', label: t('flow.taskDetail.noMilestone') },
                      ...milestones.map((m) => ({ value: m.id, label: m.name })),
                    ]
              }
              onChange={(milestoneId) => updateTask(task.id, { milestoneId })}
              disabled={!canEdit || milestones.length === 0}
              placeholder={t('flow.taskDetail.noMilestonesInProject')}
              muted={!milestone}
            />
          </DetailField>
        )}

        {task.projectId && (
          <TaskCustomFields
            taskId={String(task.id)}
            projectId={String(task.projectId)}
            values={task.fieldValues}
            canEdit={canEdit}
            people={projectMembers.map((m) => ({ id: String(m.userId), name: m.name }))}
            Row={DetailField}
          />
        )}

        <DetailField icon={Link2} label={t('flow.taskDetail.dependencies')}>
          <div className="space-y-1.5">
            {hasQueryFailed(tasksQuery) && (
              <ListLoadFailed onRetry={() => void tasksQuery.refetch()} />
            )}
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
                    void navigate({ to: '/flow/task/$id', params: { id: String(dt.id) } });
                  }}
                  className="text-flow-meta text-start truncate flex-1 hover:underline"
                  style={{ color: 'var(--accent)' }}
                >
                  {dt.identifier} · {dt.title}
                </button>
                {canEdit && (
                  <button
                    type="button"
                    className="text-flow-meta shrink-0"
                    style={{ color: 'var(--flow-text-dim)' }}
                    disabled={removeDependency.isPending}
                    onClick={() =>
                      removeDependency.mutate({ taskId: task.id, dependsOnTaskId: dt.id })
                    }
                  >
                    {t('flow.taskDetail.remove')}
                  </button>
                )}
              </div>
            ))}
            {dependencyCandidates.length > 0 && canEdit && (
              <div className="flex gap-1 pt-1">
                <div className="min-w-0 flex-1">
                  <FlowSelect
                    label={t('flow.taskDetail.addDependency')}
                    value={depId}
                    placeholder={t('flow.taskDetail.addDependency')}
                    options={[
                      { value: '', label: t('flow.taskDetail.addDependency') },
                      ...dependencyCandidates.map((c) => ({
                        value: String(c.id),
                        label: `${c.identifier} ${c.title}`,
                      })),
                    ]}
                    onChange={setDepId}
                    muted
                  />
                </div>
                <button
                  type="button"
                  disabled={!depId || addDependency.isPending}
                  onClick={() => {
                    if (!depId) return;
                    addDependency.mutate(
                      { taskId: task.id, dependsOnTaskId: depId },
                      {
                        onSuccess: () => setDepId(''),
                      }
                    );
                  }}
                  className="text-flow-meta px-1.5 py-0.5 rounded-sm font-semibold disabled:opacity-40"
                  style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
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
                className="text-flow-meta px-1.5 py-0.5 rounded-sm hover:bg-[var(--flow-hover)]"
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
                className="text-flow-meta px-1.5 py-0.5 rounded-sm hover:bg-[var(--flow-hover)]"
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
                  <p className="text-flow-meta px-2 py-1" style={{ color: 'var(--flow-text-dim)' }}>
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
                        className="text-flow-meta flex-1"
                        style={{
                          color: active ? l.color : 'var(--flow-text)',
                          fontWeight: active ? 600 : 400,
                        }}
                      >
                        {l.name}
                      </span>
                      {active && (
                        <span className="text-flow-meta" style={{ color: l.color }}>
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
      {printing && (
        <TaskPrintView
          identifier={task.identifier ?? task.id}
          projectName={project?.name || task.projectName || ''}
          title={task.title}
          description={displayDescription(task.description)}
          fields={[
            {
              label: t('flow.taskDetail.status'),
              value: (() => {
                const current = projectStatuses.find((s) => s.id === currentStatusId);
                return current
                  ? displayStatusName(current.name, t)
                  : t(STATUS_LABEL_KEYS[task.status]);
              })(),
            },
            {
              label: t('flow.taskDetail.priority'),
              value: t(priorityLabelKey(task.priority)),
            },
            {
              label: t('flow.taskDetail.assignee'),
              value: task.assigneeId ? assignee : t('flow.quickCreate.unassigned'),
            },
            { label: t('flow.taskDetail.reporter'), value: reporterName },
            { label: t('flow.taskDetail.dueDate'), value: task.dueDate },
            {
              label: t('flow.taskDetail.estimate'),
              value:
                estimatedHours == null ? '' : `${estimatedHours} ${t('flow.taskDetail.hours')}`,
            },
            { label: t('flow.taskDetail.labels'), value: taskLabels.map((l) => l.name).join(', ') },
          ]}
          subtasks={taskSubtasks}
          comments={taskComments}
          mentionNames={Object.values(peopleNames)}
          onDone={() => setPrinting(false)}
        />
      )}
      <NewStatusDialog
        open={newStatusOpen}
        onOpenChange={setNewStatusOpen}
        onCreate={createAndApplyStatus}
        pending={createStatus.isPending}
      />
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
        confirmLabel={t('flowCommon.delete')}
        cancelLabel={t('flowCommon.cancel')}
        pending={deleteTaskMut.isPending}
        onConfirm={() => {
          // Wait for the mutation to actually succeed before toasting and
          // navigating away: on failure the user would otherwise see a success
          // toast, leave the page, and the task would still be alive in cache.
          deleteTaskMut.mutate(task.id, {
            onSuccess: () => {
              toast.success(t('flowToasts.task.deleted'));
              setConfirmingDelete(false);
              closeTask();
            },
            // useDeleteTask already toasts on error.
          });
        }}
      />
    </div>
  );
};

// ── DetailField ───────────────────────────────────────────────────────────────

const DetailField: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  label: string;
  children: React.ReactNode;
}> = ({ icon: Icon, label, children }) => (
  <div className="flow-task-detail-field mb-5">
    <div className="flex items-center gap-1.5 mb-1.5">
      <Icon size={11} />
      <span
        className="flow-task-detail-label text-flow-meta font-medium"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {label}
      </span>
    </div>
    <div>{children}</div>
  </div>
);

export default TaskDetail;
