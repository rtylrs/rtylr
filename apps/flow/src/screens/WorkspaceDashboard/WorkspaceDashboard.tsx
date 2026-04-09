import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  CheckCircle2,
  Clock,
  AlertTriangle,
  TrendingUp,
  ArrowRight,
  Calendar,
} from 'lucide-react';
import { useUserStore } from '../../stores/userStore';
import { useFlowStore } from '../../stores/flowStore';
import { cn, formatDate } from '../../lib/utils';
import { flowApi, type DashboardStats } from '../../services/flowApi';

const ACCENT = '#64C5B9';

const PRIORITY_COLORS: Record<string, string> = {
  urgent: '#ef4444',
  high: '#f97316',
  medium: '#3b82f6',
  low: '#6b7280',
};

const STATUS_LABELS: Record<string, string> = {
  todo: 'Todo',
  in_progress: 'In Progress',
  in_review: 'In Review',
  done: 'Done',
};

const STATUS_COLORS: Record<string, { bg: string; text: string }> = {
  todo: { bg: 'rgba(107,114,128,0.15)', text: '#6b7280' },
  in_progress: { bg: 'rgba(59,130,246,0.15)', text: '#3b82f6' },
  in_review: { bg: 'rgba(245,158,11,0.15)', text: '#f59e0b' },
  done: { bg: 'rgba(34,197,94,0.15)', text: '#22c55e' },
};

const statCard = (
  icon: React.ReactNode,
  label: string,
  value: string | number,
  color: string,
  index: number
) => (
  <motion.div
    key={label}
    initial={{ opacity: 0, y: 12 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.2, delay: index * 0.06 }}
    className="rounded-xl p-4 flex flex-col gap-3"
    style={{
      background: 'var(--flow-card)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <div className="flex items-center justify-between">
      <span
        className="text-[12px] font-medium"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {label}
      </span>
      <span
        className="h-7 w-7 rounded-lg flex items-center justify-center"
        style={{ background: `${color}18` }}
      >
        {icon}
      </span>
    </div>
    <span
      className="text-2xl font-bold"
      style={{ color: 'var(--flow-text)' }}
    >
      {value}
    </span>
  </motion.div>
);

const WorkspaceDashboard: React.FC = () => {
  const { workspaceId = 'default' } = useParams<{ workspaceId: string }>();
  const { user } = useUserStore();
  const { tasks, projects, workspaces } = useFlowStore();

  const [apiStats, setApiStats] = useState<DashboardStats | null>(null);

  useEffect(() => {
    flowApi.getDashboard(workspaceId).then(setApiStats).catch(() => {});
  }, [workspaceId]);

  const workspace = workspaces.find((w) => w.id === workspaceId);
  const workspaceProjects = projects.filter((p) => p.workspaceId === workspaceId);
  const workspaceTasks = tasks.filter((t) => t.workspaceId === workspaceId);

  const firstName = user?.name?.split(' ')[0] || 'there';
  const hour = new Date().getHours();
  const greeting =
    hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening';

  const myTasks = apiStats?.myTasks ?? workspaceTasks.filter(
    (t) => t.assigneeInitials === 'MH' && t.status !== 'done'
  );

  const openTasks = apiStats?.openTasks ?? workspaceTasks.filter((t) => t.status === 'todo').length;
  const inProgress = apiStats?.inProgress ?? workspaceTasks.filter((t) => t.status === 'in_progress').length;
  const completedThisWeek = apiStats?.completedThisWeek ?? workspaceTasks.filter((t) => t.status === 'done').length;
  const overdue = apiStats?.overdue ?? workspaceTasks.filter((t) => {
    const d = new Date(t.dueDate);
    return d < new Date() && t.status !== 'done';
  }).length;

  return (
    <div className="p-6 max-w-5xl mx-auto">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
        className="mb-8"
      >
        <h1
          className="text-2xl font-bold mb-1"
          style={{ color: 'var(--flow-text)' }}
        >
          {greeting}, {firstName} 👋
        </h1>
        <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
          Here's what's happening in{' '}
          <span style={{ color: workspace?.color || ACCENT }}>
            {workspace?.name}
          </span>{' '}
          today.
        </p>
      </motion.div>

      {/* Stats Grid */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {statCard(
          <TrendingUp size={15} style={{ color: '#3b82f6' }} />,
          'Open Tasks',
          openTasks,
          '#3b82f6',
          0
        )}
        {statCard(
          <Clock size={15} style={{ color: ACCENT }} />,
          'In Progress',
          inProgress,
          ACCENT,
          1
        )}
        {statCard(
          <CheckCircle2 size={15} style={{ color: '#22c55e' }} />,
          'Completed This Week',
          completedThisWeek,
          '#22c55e',
          2
        )}
        {statCard(
          <AlertTriangle size={15} style={{ color: '#ef4444' }} />,
          'Overdue',
          overdue,
          '#ef4444',
          3
        )}
      </div>

      <div className="grid grid-cols-5 gap-6">
        {/* My Tasks */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.22, delay: 0.12 }}
          className="col-span-3 rounded-xl overflow-hidden"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
          }}
        >
          <div
            className="flex items-center justify-between px-5 py-4"
            style={{ borderBottom: '1px solid var(--flow-border)' }}
          >
            <h2
              className="text-[14px] font-semibold"
              style={{ color: 'var(--flow-text)' }}
            >
              My Tasks
            </h2>
            <span
              className="text-[12px] px-2 py-0.5 rounded-full"
              style={{
                background: `${ACCENT}18`,
                color: ACCENT,
              }}
            >
              {myTasks.length} active
            </span>
          </div>

          <div className="divide-y" style={{ borderColor: 'var(--flow-border)' }}>
            {myTasks.slice(0, 4).map((task, i) => (
              <motion.div
                key={task.id}
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.18, delay: 0.15 + i * 0.05 }}
                className="flex items-center gap-3 px-5 py-3.5 hover:bg-[var(--flow-hover)] transition-colors cursor-pointer"
              >
                {/* Priority dot */}
                <span
                  className="h-2 w-2 rounded-full shrink-0"
                  style={{ background: PRIORITY_COLORS[task.priority] }}
                />

                <div className="flex-1 min-w-0">
                  <p
                    className="text-[13px] font-medium truncate"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {task.title}
                  </p>
                  <p
                    className="text-[11px] mt-0.5"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {task.projectName}
                  </p>
                </div>

                <div className="flex items-center gap-2.5 shrink-0">
                  <span
                    className="text-[11px] flex items-center gap-1"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    <Calendar size={11} />
                    {formatDate(task.dueDate)}
                  </span>
                  <span
                    className="text-[11px] px-2 py-0.5 rounded-full font-medium"
                    style={{
                      background: STATUS_COLORS[task.status].bg,
                      color: STATUS_COLORS[task.status].text,
                    }}
                  >
                    {STATUS_LABELS[task.status]}
                  </span>
                </div>
              </motion.div>
            ))}
          </div>
        </motion.div>

        {/* Recent Projects */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.22, delay: 0.16 }}
          className="col-span-2 rounded-xl overflow-hidden"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
          }}
        >
          <div
            className="flex items-center justify-between px-5 py-4"
            style={{ borderBottom: '1px solid var(--flow-border)' }}
          >
            <h2
              className="text-[14px] font-semibold"
              style={{ color: 'var(--flow-text)' }}
            >
              Projects
            </h2>
            <Link
              to={`/w/${workspaceId}`}
              className="flex items-center gap-1 text-[11px] transition-colors hover:opacity-70"
              style={{ color: ACCENT }}
            >
              View all
              <ArrowRight size={11} />
            </Link>
          </div>

          <div className="p-3 space-y-2">
            {(apiStats?.recentProjects ?? workspaceProjects).slice(0, 3).map((project, i) => {
              const progress =
                project.taskCount > 0
                  ? Math.round((project.completedCount / project.taskCount) * 100)
                  : 0;
              return (
                <motion.div
                  key={project.id}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.18, delay: 0.2 + i * 0.05 }}
                >
                  <Link
                    to={`/w/${workspaceId}/p/${project.id}`}
                    className="block p-3 rounded-lg transition-colors hover:bg-[var(--flow-hover)]"
                  >
                    <div className="flex items-center gap-2 mb-2.5">
                      <span
                        className="h-2.5 w-2.5 rounded-full shrink-0"
                        style={{ background: project.color }}
                      />
                      <span
                        className="text-[13px] font-medium flex-1 truncate"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {project.name}
                      </span>
                    </div>

                    {/* Progress bar */}
                    <div
                      className="w-full rounded-full h-1.5 mb-2"
                      style={{ background: 'var(--flow-hover)' }}
                    >
                      <div
                        className="h-1.5 rounded-full transition-all duration-500"
                        style={{
                          width: `${progress}%`,
                          background: project.color,
                        }}
                      />
                    </div>

                    <div className="flex items-center justify-between">
                      {/* Member avatars */}
                      <div className="flex -space-x-1.5">
                        {project.memberAvatars.slice(0, 3).map((initials) => (
                          <span
                            key={initials}
                            className="h-5 w-5 rounded-full flex items-center justify-center text-[9px] font-bold ring-1"
                            style={{
                              background: `${project.color}22`,
                              color: project.color,
                              ringColor: 'var(--flow-card)',
                            }}
                          >
                            {initials}
                          </span>
                        ))}
                      </div>
                      <span
                        className="text-[11px]"
                        style={{ color: 'var(--flow-text-dim)' }}
                      >
                        {progress}% done
                      </span>
                    </div>
                  </Link>
                </motion.div>
              );
            })}
          </div>
        </motion.div>
      </div>
    </div>
  );
};

export default WorkspaceDashboard;
