import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { toast } from 'sonner';
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
} from 'lucide-react';
import { useTask } from './useTask';
import { useUpdateTask, useDeleteTask, useAddComment, useAddSubtask, useLogTime } from './useTasks';
import { useLabels } from './useLabels';
import { useFlowStore, PRIORITY_LABELS, STATUS_LABELS, Task } from '../../shared/stores/flowStore';
import type { ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';
import { useUserStore } from '../../shared/stores/userStore';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import CommentList from '../../shared/components/CommentList';
import CommentInput from '../../shared/components/CommentInput';
import SubtaskList from '../../shared/components/SubtaskList';
import ActivityFeed from '../../shared/components/ActivityFeed';
import AttachmentUpload from '../../shared/components/AttachmentUpload';
import AssigneeAvatar from '../../shared/components/AssigneeAvatar';
import LabelChip from '../../shared/components/LabelChip';
import DueDatePicker from '../../shared/components/DueDatePicker';
import { getInitials } from '../../shared/lib/utils';
import { EFFORT_PRESETS, formatEffortMinutes } from '../../shared/lib/effortPresets';

const ACCENT = '#64C5B9';

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: taskData, isLoading } = useTask(id);
  const user = useUserStore((s) => s.user);
  const updateTaskMut = useUpdateTask();
  const deleteTaskMut = useDeleteTask();
  const addCommentMut = useAddComment();
  const addSubtaskMut = useAddSubtask();
  const logTimeMut = useLogTime();
  const { data: allLabels = [] } = useLabels();
  const {
    toggleSubtask,
    members,
    cycles,
    milestones,
    archiveTask,
    attachments,
    timelogs,
    projects,
    tasks: allTasks,
    sections,
    customFieldDefinitions,
  } = useFlowStore();
  const [labelPickerOpen, setLabelPickerOpen] = useState(false);

  const updateTask = (id: string, patch: Partial<Task>) => {
    updateTaskMut.mutate({ id, data: patch as Partial<ApiTask> });
  };
  const deleteTask = (id: string) => {
    deleteTaskMut.mutate(id);
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
  const logTime = (taskId: string, minutes: number, note?: string) => {
    logTimeMut.mutate({ taskId, minutes, note });
  };

  const [tab, setTab] = useState<'activity' | 'comments'>('comments');
  const [depId, setDepId] = useState('');

  if (isLoading) {
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        Loading task...
      </div>
    );
  }

  const detail = taskData as ApiTaskDetail | undefined;
  const task = detail?.task;
  const taskComments = (detail?.comments ?? []).map((c) => ({
    id: c.id,
    taskId: c.taskId,
    authorId: c.userId,
    author: c.userId,
    authorInitials: c.userId.slice(0, 2).toUpperCase(),
    body: c.content,
    createdAt: c.createdAt,
  }));
  const taskActivity = (detail?.activity ?? []).map((a) => ({
    id: a.id,
    taskId: a.taskId,
    type: 'status_change' as const,
    actor: a.userId,
    actorInitials: a.userId.slice(0, 2).toUpperCase(),
    from: a.oldValue,
    to: a.newValue,
    createdAt: a.createdAt,
    detail: a.action,
  }));
  const taskSubtasks = (detail?.subtasks ?? []).map((s) => ({
    id: s.id,
    taskId: task?.id ?? '',
    title: s.title,
    completed: s.status === 'done',
  }));

  if (!task) {
    return (
      <div className="p-10 text-center">
        <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
          Task not found.
        </p>
      </div>
    );
  }

  const project = projects.find((p) => p.id === task.projectId);
  const projectSections = sections
    .filter((s) => s.projectId === task.projectId)
    .sort((a, b) => a.order - b.order);
  const fieldDefs = customFieldDefinitions
    .filter((f) => f.projectId === task.projectId)
    .sort((a, b) => a.order - b.order);
  const dependencyTasks = allTasks.filter((ot) => (task as any).dependencies?.includes(ot.id));
  const dependencyCandidates = allTasks.filter(
    (ot) =>
      ot.workspaceId === task.workspaceId &&
      ot.id !== task.id &&
      !(task as any).dependencies?.includes(ot.id)
  );
  const taskAttachments = attachments.filter((a) => a.taskId === task.id);
  const taskTimelogs = timelogs.filter((tl) => tl.taskId === task.id);
  const taskLabels = allLabels.filter((l) => (task as any).labels?.includes(l.id));
  const cycle = cycles.find((c) => c.id === (task as any).cycleId);
  const milestone = milestones.find((m) => m.id === (task as any).milestoneId);
  const totalMinutes = taskTimelogs.reduce((s, t) => s + t.minutes, 0);

  const priorities: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];
  const statuses: Task['status'][] = [
    'backlog',
    'todo',
    'in_progress',
    'in_review',
    'done',
    'cancelled',
  ];

  return (
    <div className="flex h-full min-h-0 flex-col overflow-y-auto lg:flex-row lg:overflow-hidden">
      {/* Main area */}
      <div className="flex min-h-0 flex-1 flex-col lg:overflow-hidden">
        <div
          className="flex shrink-0 items-center justify-between gap-2 px-4 py-3 sm:px-6"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div className="flex min-w-0 items-center gap-2 sm:gap-3">
            <button
              onClick={() => navigate(-1)}
              className="h-6 w-6 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <ArrowLeft size={13} />
            </button>
            <Breadcrumbs
              items={[
                {
                  label: project?.name || 'Project',
                  to: `/w/${task.workspaceId}/p/${task.projectId}`,
                },
                { label: task.identifier ?? task.id },
              ]}
            />
          </div>
          <div className="flex shrink-0 items-center gap-0.5 sm:gap-1">
            <button
              onClick={() => {
                navigator.clipboard?.writeText(task.identifier ?? task.id);
                toast.success(`Copied ${task.identifier ?? task.id}`);
              }}
              className="flex h-8 items-center gap-1 rounded px-2 text-[11px] hover:bg-[var(--flow-hover)] sm:h-7"
              style={{ color: 'var(--flow-text-muted)' }}
              title="Copy ID"
            >
              <Copy size={11} /> <span className="hidden sm:inline">Copy ID</span>
            </button>
            <button
              onClick={() => {
                archiveTask(task.id);
                toast.success('Task archived');
                navigate(-1);
              }}
              className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
              title="Archive"
            >
              <Archive size={12} />
            </button>
            <button
              onClick={() => {
                deleteTask(task.id);
                toast.success('Task deleted');
                navigate(-1);
              }}
              className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
              style={{ color: '#ef4444' }}
              title="Delete"
            >
              <Trash2 size={12} />
            </button>
          </div>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.18 }}
          className="min-h-0 flex-1 overflow-y-auto lg:min-h-0"
        >
          <div className="max-w-3xl mx-auto px-4 py-6 sm:px-8 sm:py-8">
            {/* Title */}
            <input
              value={task.title}
              onChange={(e) => updateTask(task.id, { title: e.target.value })}
              className="w-full bg-transparent outline-none font-bold mb-4 text-xl sm:text-2xl"
              style={{
                color: 'var(--flow-text)',
              }}
            />
            {/* Description */}
            <textarea
              value={task.description || ''}
              onChange={(e) => updateTask(task.id, { description: e.target.value })}
              placeholder="Add a description…"
              rows={5}
              className="w-full bg-transparent outline-none text-[13px] leading-relaxed mb-6 resize-none"
              style={{ color: 'var(--flow-text-muted)' }}
            />

            {/* Attachments */}
            <section className="mb-8">
              <div className="flex items-center gap-2 mb-3">
                <GitBranch size={12} style={{ color: 'var(--flow-text-muted)' }} />
                <h3
                  className="text-[11px] font-semibold uppercase tracking-wider"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  Subtasks
                </h3>
              </div>
              <SubtaskList
                subtasks={taskSubtasks}
                onToggle={toggleSubtask}
                onAdd={(title) => addSubtask(task.id, title)}
              />
            </section>

            <section className="mb-8">
              <h3
                className="text-[11px] font-semibold uppercase tracking-wider mb-3"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                Attachments
              </h3>
              <AttachmentUpload attachments={taskAttachments} />
            </section>

            {/* Tabs */}
            <div
              className="flex items-center gap-1 mb-4"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              {(['comments', 'activity'] as const).map((t) => (
                <button
                  key={t}
                  onClick={() => setTab(t)}
                  className="px-3 py-2 text-[12px] font-medium capitalize"
                  style={{
                    color: tab === t ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                    borderBottom: tab === t ? '2px solid #64C5B9' : '2px solid transparent',
                  }}
                >
                  {t}
                  {t === 'comments' && (
                    <span
                      className="ml-1.5 text-[10px] px-1 rounded"
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
                        : { id: 'local', name: 'You', initials: 'YO' }
                    )
                  }
                />
              </div>
            )}
            {tab === 'activity' && <ActivityFeed entries={taskActivity} />}
          </div>
        </motion.div>
      </div>

      {/* Side panel — below main on small screens */}
      <aside
        className="w-full shrink-0 overflow-y-auto border-t border-[var(--flow-border)] px-4 py-6 sm:px-5 lg:w-64 lg:border-l lg:border-t-0"
        style={{
          background: 'var(--flow-surface)',
        }}
      >
        <DetailField icon={CircleDot} label="Status">
          <select
            value={task.status}
            onChange={(e) => {
              const next = e.target.value as Task['status'];
              if (next === 'done' && task.status !== 'done') {
                toast.success('Nice work!', {
                  description: 'Task marked complete.',
                  icon: '🎉',
                });
              }
              updateTask(task.id, { status: next });
            }}
            className="bg-transparent text-[12px] outline-none w-full"
            style={{ color: 'var(--flow-text)' }}
          >
            {statuses.map((s) => (
              <option key={s} value={s} style={{ background: 'var(--flow-card)' }}>
                {STATUS_LABELS[s]}
              </option>
            ))}
          </select>
        </DetailField>

        <DetailField icon={Flag} label="Priority">
          <select
            value={task.priority}
            onChange={(e) => updateTask(task.id, { priority: e.target.value as Task['priority'] })}
            className="bg-transparent text-[12px] outline-none w-full"
            style={{ color: 'var(--flow-text)' }}
          >
            {priorities.map((p) => (
              <option key={p} value={p} style={{ background: 'var(--flow-card)' }}>
                {PRIORITY_LABELS[p]}
              </option>
            ))}
          </select>
        </DetailField>

        <DetailField icon={UserIcon} label="Assignee">
          <div className="flex items-center gap-2">
            <AssigneeAvatar name={task.assignee} initials={task.assigneeInitials} size={18} />
            <select
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
              className="bg-transparent text-[12px] outline-none flex-1"
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

        <DetailField icon={CalendarIcon} label="Due date">
          <DueDatePicker
            value={task.dueDate}
            onChange={(d) => updateTask(task.id, { dueDate: d })}
          />
        </DetailField>

        <DetailField icon={Clock} label="Estimate">
          <input
            type="number"
            value={task.estimate || ''}
            onChange={(e) => updateTask(task.id, { estimate: Number(e.target.value) })}
            placeholder="hours"
            className="bg-transparent text-[12px] outline-none w-full"
            style={{ color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Gauge} label="Effort">
          <div className="flex flex-wrap gap-1">
            <button
              type="button"
              onClick={() => updateTask(task.id, { effortMinutes: undefined })}
              className="text-[10px] px-2 py-1 rounded-md"
              style={{
                background: !task.effortMinutes ? 'var(--flow-hover)' : 'transparent',
                color: 'var(--flow-text-muted)',
                border: '1px solid var(--flow-border)',
              }}
            >
              —
            </button>
            {EFFORT_PRESETS.map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => updateTask(task.id, { effortMinutes: m })}
                className="text-[10px] px-2 py-1 rounded-md transition-colors"
                style={{
                  background: task.effortMinutes === m ? `${ACCENT}28` : 'transparent',
                  color: 'var(--flow-text)',
                  border:
                    task.effortMinutes === m
                      ? `1px solid ${ACCENT}66`
                      : '1px solid var(--flow-border)',
                }}
                title={`${m} minutes`}
              >
                {formatEffortMinutes(m)}
              </button>
            ))}
          </div>
        </DetailField>

        <DetailField icon={Hourglass} label="Waiting on">
          <input
            value={task.waitingOn || ''}
            onChange={(e) => updateTask(task.id, { waitingOn: e.target.value || undefined })}
            placeholder="Person, team, approval…"
            className="bg-transparent text-[12px] outline-none w-full placeholder:opacity-50"
            style={{ color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Lightbulb} label="Decision">
          <textarea
            value={task.decisionNote || ''}
            onChange={(e) => updateTask(task.id, { decisionNote: e.target.value || undefined })}
            placeholder="Why this approach (one line)"
            rows={2}
            className="w-full bg-transparent text-[12px] outline-none resize-none placeholder:opacity-50 rounded-md px-2 py-1.5"
            style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text)' }}
          />
        </DetailField>

        <DetailField icon={Pin} label="Today">
          <button
            type="button"
            onClick={() => updateTask(task.id, { focusPin: !task.focusPin })}
            className="text-[11px] px-2.5 py-1.5 rounded-md w-full text-left transition-colors"
            style={{
              background: task.focusPin ? `${ACCENT}22` : 'var(--flow-hover)',
              color: 'var(--flow-text)',
              border: task.focusPin ? `1px solid ${ACCENT}55` : '1px solid transparent',
            }}
          >
            {task.focusPin ? 'Pinned to Today view' : 'Pin to Today view'}
          </button>
        </DetailField>

        <DetailField icon={Timer} label="Time tracked">
          <div className="flex items-center gap-2">
            <span className="text-[12px]" style={{ color: 'var(--flow-text)' }}>
              {Math.round((totalMinutes / 60) * 10) / 10}h
            </span>
            <button
              onClick={() => {
                logTime(task.id, 30, 'Quick log');
                toast.success('Logged 30m');
              }}
              className="text-[10px] px-1.5 py-0.5 rounded hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              +30m
            </button>
          </div>
        </DetailField>

        {cycle && (
          <DetailField icon={Timer} label="Cycle">
            <span className="text-[12px]" style={{ color: 'var(--flow-text)' }}>
              {cycle.name}
            </span>
          </DetailField>
        )}

        {milestone && (
          <DetailField icon={MilestoneIcon} label="Milestone">
            <span className="text-[12px]" style={{ color: 'var(--flow-text)' }}>
              {milestone.name}
            </span>
          </DetailField>
        )}

        {projectSections.length > 0 && (
          <DetailField icon={LayoutList} label="Section">
            <select
              value={task.sectionId ?? ''}
              onChange={(e) => updateTask(task.id, { sectionId: e.target.value || undefined })}
              className="bg-transparent text-[12px] outline-none w-full"
              style={{ color: 'var(--flow-text)' }}
            >
              <option value="">No section</option>
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
                value={String(task.customFieldValues?.[def.id] ?? '')}
                onChange={(e) =>
                  updateTask(task.id, {
                    customFieldValues: {
                      ...(task.customFieldValues ?? {}),
                      [def.id]: e.target.value,
                    },
                  })
                }
                className="bg-transparent text-[12px] outline-none w-full"
                style={{ color: 'var(--flow-text)' }}
              >
                <option value="">—</option>
                {(def.options ?? []).map((opt) => (
                  <option key={opt} value={opt} style={{ background: 'var(--flow-card)' }}>
                    {opt}
                  </option>
                ))}
              </select>
            ) : def.type === 'number' ? (
              <input
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
                className="bg-transparent text-[12px] outline-none w-full"
                style={{ color: 'var(--flow-text)' }}
              />
            ) : (
              <input
                value={String(task.customFieldValues?.[def.id] ?? '')}
                onChange={(e) =>
                  updateTask(task.id, {
                    customFieldValues: {
                      ...(task.customFieldValues ?? {}),
                      [def.id]: e.target.value,
                    },
                  })
                }
                className="bg-transparent text-[12px] outline-none w-full"
                style={{ color: 'var(--flow-text)' }}
              />
            )}
          </DetailField>
        ))}

        <DetailField icon={Link2} label="Dependencies">
          <div className="space-y-1.5">
            {dependencyTasks.map((dt) => (
              <div key={dt.id} className="flex items-center justify-between gap-1">
                <button
                  type="button"
                  onClick={() => navigate(`/task/${dt.id}`)}
                  className="text-[11px] text-left truncate flex-1 hover:underline"
                  style={{ color: ACCENT }}
                >
                  {dt.identifier} · {dt.title}
                </button>
                <button
                  type="button"
                  className="text-[10px] shrink-0"
                  style={{ color: 'var(--flow-text-dim)' }}
                  onClick={() =>
                    updateTask(task.id, {
                      dependencies: (task.dependencies ?? []).filter((x) => x !== dt.id),
                    })
                  }
                >
                  Remove
                </button>
              </div>
            ))}
            {dependencyCandidates.length > 0 && (
              <div className="flex gap-1 pt-1">
                <select
                  value={depId}
                  onChange={(e) => setDepId(e.target.value)}
                  className="bg-transparent text-[11px] outline-none flex-1 min-w-0 rounded px-1"
                  style={{ color: 'var(--flow-text)', border: '1px solid var(--flow-border)' }}
                >
                  <option value="">Add dependency…</option>
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
                  className="text-[10px] px-1.5 py-0.5 rounded font-semibold disabled:opacity-40"
                  style={{ background: ACCENT, color: '#0a0a0a' }}
                >
                  Add
                </button>
              </div>
            )}
          </div>
        </DetailField>

        <DetailField icon={Tag} label="Labels">
          <div className="flex flex-wrap gap-1 relative">
            {taskLabels.map((l) => (
              <LabelChip
                key={l.id}
                label={l}
                small
                onRemove={() =>
                  updateTask(task.id, {
                    labels: ((task as any).labels ?? []).filter((lid: string) => lid !== l.id),
                  })
                }
              />
            ))}
            <button
              onClick={() => setLabelPickerOpen((v) => !v)}
              className="text-[10px] px-1.5 py-0.5 rounded hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              + Add
            </button>
            {labelPickerOpen && (
              <div
                className="absolute top-6 left-0 z-50 rounded-lg shadow-lg p-2 min-w-[160px] space-y-0.5"
                style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
              >
                {allLabels.length === 0 && (
                  <p className="text-[11px] px-2 py-1" style={{ color: 'var(--flow-text-dim)' }}>
                    No labels yet
                  </p>
                )}
                {allLabels.map((l) => {
                  const active = ((task as any).labels ?? []).includes(l.id);
                  return (
                    <button
                      key={l.id}
                      onClick={() => {
                        const current: string[] = (task as any).labels ?? [];
                        updateTask(task.id, {
                          labels: active
                            ? current.filter((lid) => lid !== l.id)
                            : [...current, l.id],
                        });
                        setLabelPickerOpen(false);
                      }}
                      className="w-full flex items-center gap-2 px-2 py-1 rounded text-left hover:bg-[var(--flow-hover)]"
                    >
                      <span
                        className="h-2 w-2 rounded-full shrink-0"
                        style={{ background: l.color }}
                      />
                      <span
                        className="text-[11px] flex-1"
                        style={{
                          color: active ? l.color : 'var(--flow-text)',
                          fontWeight: active ? 600 : 400,
                        }}
                      >
                        {l.name}
                      </span>
                      {active && (
                        <span className="text-[10px]" style={{ color: l.color }}>
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
    </div>
  );
};

const DetailField: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  label: string;
  children: React.ReactNode;
}> = ({ icon: Icon, label, children }) => (
  <div className="mb-5">
    <div className="flex items-center gap-1.5 mb-1.5">
      <Icon size={11} />
      <span
        className="text-[10px] uppercase tracking-wider font-semibold"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        {label}
      </span>
    </div>
    <div>{children}</div>
  </div>
);

export default TaskDetail;
