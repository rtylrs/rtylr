import React, { useMemo, useState, useCallback } from 'react';
import { Link, useParams } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  CheckCircle2,
  Clock,
  AlertTriangle,
  TrendingUp,
  ArrowRight,
  Calendar,
  Gauge,
  Flame,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../shared/stores/userStore';
import { useFlowStore } from '../../shared/stores/flowStore';
import { cn, formatDate } from '../../shared/lib/utils';
import { type DashboardStats, type ApiTask } from '../../shared/services/flowApi';
import { useFlowDashboard } from './useDashboard';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { getCurrentUserId, isTaskMine } from '../../shared/lib/currentUser';
import { isDueThisCalendarWeek } from '../../shared/lib/lifeRhythm';
import { getWeeklyReviewStreak, recordWeeklyReviewDone } from '../../shared/lib/weeklyReviewStreak';

const PRIORITY_COLORS: Record<string, string> = {
  urgent: '#ef4444',
  high: '#f97316',
  medium: '#3b82f6',
  low: '#6b7280',
  no_priority: '#4b5563',
};

const STATUS_LABELS: Record<string, string> = {
  backlog: 'Backlog',
  todo: 'Todo',
  in_progress: 'In Progress',
  in_review: 'In Review',
  done: 'Done',
  cancelled: 'Cancelled',
};

const STATUS_COLORS: Record<string, { bg: string; text: string }> = {
  backlog: { bg: 'rgba(107,114,128,0.15)', text: '#6b7280' },
  todo: { bg: 'rgba(156,163,175,0.15)', text: '#9ca3af' },
  in_progress: { bg: 'rgba(59,130,246,0.15)', text: '#3b82f6' },
  in_review: { bg: 'rgba(245,158,11,0.15)', text: '#f59e0b' },
  done: { bg: 'rgba(34,197,94,0.15)', text: '#22c55e' },
  cancelled: { bg: 'rgba(239,68,68,0.15)', text: '#ef4444' },
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
      <span className="text-[12px] font-medium" style={{ color: 'var(--flow-text-muted)' }}>
        {label}
      </span>
      <span
        className="h-7 w-7 rounded-lg flex items-center justify-center"
        style={{ background: `${color}18` }}
      >
        {icon}
      </span>
    </div>
    <span className="text-2xl font-bold" style={{ color: 'var(--flow-text)' }}>
      {value}
    </span>
  </motion.div>
);

const WorkspaceDashboard: React.FC = () => {
  const { t } = useTranslation();
  const { workspaceId = PERSONAL_WORKSPACE_ID } = useParams({ strict: false });
  const { user } = useUserStore();
  const { tasks, projects, workspaces, members } = useFlowStore();

  const { data: apiStats } = useFlowDashboard(workspaceId);

  const workspace = workspaces.find((w) => w.id === workspaceId);
  const workspaceProjects = projects.filter((p) => p.workspaceId === workspaceId);
  const workspaceTasks = tasks.filter((t) => t.workspaceId === workspaceId);

  const firstName = user?.name?.split(' ')[0] || 'there';
  const hour = new Date().getHours();
  const greeting = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening';

  const myTasks: ApiTask[] =
    apiStats?.myTasks ??
    (workspaceTasks.filter(
      (t) => isTaskMine(t.assigneeId) && t.status !== 'done' && t.status !== 'cancelled'
    ) as ApiTask[]);

  const openTasks = apiStats?.openTasks ?? workspaceTasks.filter((t) => t.status === 'todo').length;
  const inProgress =
    apiStats?.inProgress ?? workspaceTasks.filter((t) => t.status === 'in_progress').length;
  const completedThisWeek =
    apiStats?.completedThisWeek ?? workspaceTasks.filter((t) => t.status === 'done').length;
  const overdue =
    apiStats?.overdue ??
    workspaceTasks.filter((t) => {
      const d = new Date(t.dueDate);
      return d < new Date() && t.status !== 'done';
    }).length;

  const uid = getCurrentUserId();
  const myCapacityHours = members.find((m) => m.id === uid)?.capacity ?? 40;

  const weekLoad = useMemo(() => {
    const mine = workspaceTasks.filter(
      (t) =>
        isTaskMine(t.assigneeId) &&
        isDueThisCalendarWeek(t.dueDate) &&
        t.status !== 'done' &&
        t.status !== 'cancelled'
    );
    const minutes = mine.reduce((sum, t) => {
      if (t.effortMinutes && t.effortMinutes > 0) return sum + t.effortMinutes;
      if (t.estimate != null && t.estimate > 0) return sum + t.estimate * 60;
      return sum;
    }, 0);
    return { minutes, count: mine.length };
  }, [workspaceTasks]);

  const weekLoadPct = Math.min(
    100,
    myCapacityHours > 0 ? (weekLoad.minutes / (myCapacityHours * 60)) * 100 : 0
  );

  const [reviewStreak, setReviewStreak] = useState(() => getWeeklyReviewStreak().streak);

  const onWeeklyReviewDone = useCallback(() => {
    const { streak, alreadyLogged } = recordWeeklyReviewDone();
    setReviewStreak(streak);
    if (alreadyLogged) {
      toast.message(t('toasts.weeklyReview.alreadyLogged'), {
        description: t('toasts.weeklyReview.alreadyLoggedStreak', { count: streak }),
      });
    } else {
      toast.success(t('toasts.weeklyReview.logged'), {
        description: t('toasts.weeklyReview.streak', { count: streak }),
      });
    }
  }, [t]);

  return (
    <>
      <HelmetContainer title={t('flow.workspaceDashboard.pageTitle')} />
      <div className="max-w-5xl mx-auto p-4 pb-6 sm:p-6">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2 }}
          className="mb-8"
        >
          <h1 className="text-2xl font-bold mb-1" style={{ color: 'var(--flow-text)' }}>
            {greeting}, {firstName} 👋
          </h1>
          <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
            Here's what's happening in{' '}
            <span style={{ color: workspace?.color || 'var(--accent)' }}>{workspace?.name}</span>{' '}
            today.
          </p>
        </motion.div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 gap-3 mb-8 sm:gap-4 xl:grid-cols-4">
          {statCard(
            <TrendingUp size={15} style={{ color: '#3b82f6' }} />,
            'Open Tasks',
            openTasks,
            '#3b82f6',
            0
          )}
          {statCard(
            <Clock size={15} style={{ color: 'var(--accent)' }} />,
            'In Progress',
            inProgress,
            'var(--accent)',
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

        <div className="grid grid-cols-1 gap-4 mb-8 md:grid-cols-2">
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2, delay: 0.08 }}
            className="rounded-xl p-4 flex flex-col gap-3"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
            }}
          >
            <div className="flex items-center justify-between">
              <span className="text-[12px] font-medium" style={{ color: 'var(--flow-text-muted)' }}>
                Week load (your tasks due this week)
              </span>
              <Gauge size={15} style={{ color: 'var(--accent)' }} />
            </div>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold" style={{ color: 'var(--flow-text)' }}>
                {Math.round(weekLoad.minutes / 60)}
              </span>
              <span className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
                h scheduled
              </span>
              <span className="text-[12px] ml-auto" style={{ color: 'var(--flow-text-muted)' }}>
                of ~{myCapacityHours}h capacity
              </span>
            </div>
            <div className="w-full rounded-full h-2" style={{ background: 'var(--flow-hover)' }}>
              <div
                className="h-2 rounded-full transition-all duration-500"
                style={{
                  width: `${weekLoadPct}%`,
                  background: weekLoadPct > 95 ? '#f97316' : 'var(--accent)',
                }}
              />
            </div>
            <p className="text-[11px] leading-snug" style={{ color: 'var(--flow-text-dim)' }}>
              Uses effort chips or estimate hours on tasks. Planned backlog stays visible in
              projects; this is only what you already scheduled for the week.
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.2, delay: 0.1 }}
            className="rounded-xl p-4 flex flex-col gap-3"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
            }}
          >
            <div className="flex items-center justify-between">
              <span className="text-[12px] font-medium" style={{ color: 'var(--flow-text-muted)' }}>
                Weekly review streak
              </span>
              <Flame size={15} style={{ color: '#f97316' }} />
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--flow-text)' }}>
              {reviewStreak}
              <span
                className="text-[14px] font-medium ml-1"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                weeks
              </span>
            </p>
            <button
              type="button"
              onClick={onWeeklyReviewDone}
              className="text-[12px] font-semibold py-2 px-3 rounded-lg w-fit transition-opacity hover:opacity-90"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              I did my review this week
            </button>
            <p className="text-[11px] leading-snug" style={{ color: 'var(--flow-text-dim)' }}>
              Honest ritual, not app opens. Tap once after you finish your weekly close-out.
            </p>
          </motion.div>
        </div>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-5 lg:gap-6">
          {/* My Tasks */}
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.22, delay: 0.12 }}
            className="rounded-xl overflow-hidden lg:col-span-3"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
            }}
          >
            <div
              className="flex items-center justify-between px-4 py-4 sm:px-5"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <h2
                className="text-[13px] sm:text-[14px] font-semibold"
                style={{ color: 'var(--flow-text)' }}
              >
                My Tasks
              </h2>
              <span
                className="text-[12px] px-2 py-0.5 rounded-full"
                style={{
                  background: 'var(--accent-muted)',
                  color: 'var(--accent)',
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
                  className="flex flex-col gap-2 px-4 py-3.5 hover:bg-[var(--flow-hover)] transition-colors cursor-pointer sm:flex-row sm:items-center sm:gap-3 sm:px-5"
                >
                  <div className="flex items-start gap-3 min-w-0 sm:items-center sm:flex-1">
                    <span
                      className="h-2 w-2 rounded-full shrink-0 mt-1.5 sm:mt-0"
                      style={{ background: PRIORITY_COLORS[task.priority] }}
                    />
                    <div className="flex-1 min-w-0">
                      <p
                        className="text-[13px] font-medium line-clamp-2 sm:truncate"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {task.title}
                      </p>
                      <p className="text-[11px] mt-0.5" style={{ color: 'var(--flow-text-muted)' }}>
                        {task.projectName}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2.5 shrink-0 pl-5 sm:pl-0">
                    <span
                      className="text-[11px] flex items-center gap-1"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      <Calendar size={11} />
                      {formatDate(task.dueDate)}
                    </span>
                    <span
                      className="text-[11px] px-2 py-0.5 rounded-full font-medium whitespace-nowrap"
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
            className="rounded-xl overflow-hidden lg:col-span-2"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
            }}
          >
            <div
              className="flex items-center justify-between px-5 py-4"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <h2 className="text-[14px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                Projects
              </h2>
              <Link
                to={`/w/${workspaceId}` as string}
                className="flex items-center gap-1 text-[11px] transition-colors hover:opacity-70"
                style={{ color: 'var(--accent)' }}
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
                      to={`/w/${workspaceId}/p/${project.id}` as string}
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
                                outlineColor: 'var(--flow-card)',
                              }}
                            >
                              {initials}
                            </span>
                          ))}
                        </div>
                        <span className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
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
    </>
  );
};

export default WorkspaceDashboard;
