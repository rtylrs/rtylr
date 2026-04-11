import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
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
} from 'lucide-react';
import { useTask } from '../../hooks/useTask';
import { useFlowStore, PRIORITY_LABELS, STATUS_LABELS, Task } from '../../stores/flowStore';
import { useUserStore } from '../../stores/userStore';
import Breadcrumbs from '../../components/Breadcrumbs';
import CommentList from '../../components/CommentList';
import CommentInput from '../../components/CommentInput';
import SubtaskList from '../../components/SubtaskList';
import ActivityFeed from '../../components/ActivityFeed';
import AttachmentUpload from '../../components/AttachmentUpload';
import AssigneeAvatar from '../../components/AssigneeAvatar';
import LabelChip from '../../components/LabelChip';
import DueDatePicker from '../../components/DueDatePicker';
import { getInitials } from '../../lib/utils';

const ACCENT = '#64C5B9';

const TaskDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: taskData, isLoading } = useTask(id);
  const user = useUserStore((s) => s.user);
  const {
    updateTask,
    addComment,
    toggleSubtask,
    addSubtask,
    labels,
    members,
    cycles,
    milestones,
    deleteTask,
    archiveTask,
    logTime,
    comments,
    activity,
    subtasks,
    attachments,
    timelogs,
    projects,
    tasks: allTasks,
    sections,
    customFieldDefinitions,
  } = useFlowStore();

  const [tab, setTab] = useState<'activity' | 'comments'>('comments');
  const [depId, setDepId] = useState('');

  if (isLoading) {
    return <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>Loading task...</div>;
  }

  if (!taskData) {
    return (
      <div className="p-10 text-center">
        <p
          className="text-[13px]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          Task not found.
        </p>
      </div>
    );
  }

  const task = taskData as Task;
  const project = projects.find((p) => p.id === task.projectId);
  const projectSections = sections
    .filter((s) => s.projectId === task.projectId)
    .sort((a, b) => a.order - b.order);
  const fieldDefs = customFieldDefinitions
    .filter((f) => f.projectId === task.projectId)
    .sort((a, b) => a.order - b.order);
  const dependencyTasks = (task.dependencies ?? [])
    .map((tid) => allTasks.find((ot) => ot.id === tid))
    .filter((ot): ot is Task => !!ot);
  const dependencyCandidates = allTasks.filter(
    (ot) =>
      ot.workspaceId === task.workspaceId &&
      ot.id !== task.id &&
      !(task.dependencies ?? []).includes(ot.id)
  );
  const taskComments = comments.filter((c) => c.taskId === task.id);
  const taskActivity = activity.filter((a) => a.taskId === task.id);
  const taskSubtasks = subtasks.filter((s) => s.taskId === task.id);
  const taskAttachments = attachments.filter((a) => a.taskId === task.id);
  const taskTimelogs = timelogs.filter((tl) => tl.taskId === task.id);
  const taskLabels = labels.filter((l) => task.labels?.includes(l.id));
  const cycle = cycles.find((c) => c.id === task.cycleId);
  const milestone = milestones.find((m) => m.id === task.milestoneId);
  const totalMinutes = taskTimelogs.reduce((s, t) => s + t.minutes, 0);

  const priorities: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];
  const statuses: Task['status'][] = ['backlog', 'todo', 'in_progress', 'in_review', 'done', 'cancelled'];

  return (
    <div className="flex h-full">
      {/* Main area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        <div
          className="px-6 py-3 flex items-center justify-between shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div className="flex items-center gap-3">
            <button
              onClick={() => navigate(-1)}
              className="h-6 w-6 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <ArrowLeft size={13} />
            </button>
            <Breadcrumbs
              items={[
                { label: project?.name || 'Project', to: `/w/${task.workspaceId}/p/${task.projectId}` },
                { label: task.identifier },
              ]}
            />
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => {
                navigator.clipboard?.writeText(task.identifier);
                toast.success(`Copied ${task.identifier}`);
              }}
              className="flex items-center gap-1 h-7 px-2 rounded text-[11px] hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Copy size={11} /> Copy ID
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
          className="flex-1 overflow-y-auto"
        >
          <div className="max-w-3xl mx-auto px-8 py-8">
            {/* Title */}
            <input
              value={task.title}
              onChange={(e) => updateTask(task.id, { title: e.target.value })}
              className="w-full bg-transparent outline-none font-bold mb-4"
              style={{
                fontSize: 24,
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

      {/* Side panel */}
      <aside
        className="w-64 shrink-0 overflow-y-auto px-5 py-6"
        style={{
          background: 'var(--flow-surface)',
          borderLeft: '1px solid var(--flow-border)',
        }}
      >
        <DetailField icon={CircleDot} label="Status">
          <select
            value={task.status}
            onChange={(e) => updateTask(task.id, { status: e.target.value as Task['status'] })}
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
                if (m) updateTask(task.id, { assigneeId: m.id, assignee: m.name, assigneeInitials: m.initials });
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
          <DueDatePicker value={task.dueDate} onChange={(d) => updateTask(task.id, { dueDate: d })} />
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
          <div className="flex flex-wrap gap-1">
            {taskLabels.map((l) => (
              <LabelChip key={l.id} label={l} small />
            ))}
            <button
              className="text-[10px] px-1.5 py-0.5 rounded hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              +
            </button>
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
