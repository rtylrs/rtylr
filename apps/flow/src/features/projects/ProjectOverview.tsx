import React from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { motion } from 'motion/react';
import {
  Activity,
  CheckCircle2,
  Clock,
  AlertCircle,
  Users,
  Milestone as MilestoneIcon,
  MessageSquare,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useProjectOverview } from './useProjectOverview';
import { useFlowStore } from '../../shared/stores/flowStore';
import { AvatarStack } from '../../shared/components/AssigneeAvatar';
import { formatDate } from '../../shared/lib/utils';
import { displayDescription } from '../../shared/lib/stringUtils';
import ProjectViewTabs from './ProjectViewTabs';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { withAlpha } from '../../shared/lib/colorUtils';

// labelKey, not label: module scope again, so it cannot call `t` itself.
const HEALTH: Record<string, { color: string; labelKey: string }> = {
  on_track: {
    color: 'var(--flow-status-approved)',
    labelKey: 'flow.projectOverview.health.onTrack',
  },
  at_risk: { color: '#f59e0b', labelKey: 'flow.milestones.status.atRisk' },
  off_track: { color: 'var(--flow-danger-fg)', labelKey: 'flow.projectOverview.health.offTrack' },
};

const ProjectOverview: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId = '' } = useParams({ strict: false });
  useProjectRealtime(projectId);

  const { data, isLoading } = useProjectOverview(projectId);
  const { projects, members } = useFlowStore();
  const project = projects.find((p) => p.id === projectId);

  if (isLoading)
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.project.loading')}
      </div>
    );

  if (!data || !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.notFound')}
      </div>
    );
  }
  const { stats, milestones } = data;
  const team = members.filter((m) => project.memberAvatars.includes(m.initials));
  const recentComments: {
    id: string;
    author: string;
    authorInitials: string;
    body: string;
    createdAt: string;
  }[] = [];
  const health = HEALTH[project.health];

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={t('flow.projectOverview.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={stats.total}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        rightSlot={
          <span
            className="text-meta px-2 py-0.5 rounded-full font-semibold"
            style={{
              background: withAlpha(health.color, 10),
              color: health.color,
              border: `1px solid ${withAlpha(health.color, 20)}`,
            }}
          >
            {t(health.labelKey)}
          </span>
        }
      />

      {displayDescription(project.description) && (
        <div
          className="px-5 py-3 shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {displayDescription(project.description)}
          </p>
        </div>
      )}

      <div className="p-5 space-y-5 max-w-5xl w-full mx-auto">
        {/* Progress */}
        <motion.section
          initial={{ opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
          className="rounded-xl p-5"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <div className="flex items-center justify-between mb-3">
            <h3
              className="text-meta font-semibold uppercase tracking-wider"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.milestones.progress')}
            </h3>
            <span className="text-screen font-bold" style={{ color: 'var(--flow-text)' }}>
              {stats.progress}%
            </span>
          </div>
          <div
            className="h-2 rounded-full overflow-hidden mb-4"
            style={{ background: 'var(--flow-hover)' }}
          >
            <div
              className="h-full rounded-full transition-all"
              style={{ width: `${stats.progress}%`, background: project.color }}
            />
          </div>
          <div className="grid grid-cols-4 gap-3">
            <StatCard
              icon={CheckCircle2}
              label={t('flow.taskDetail.done')}
              value={stats.done}
              color="var(--flow-status-approved)"
            />
            <StatCard
              icon={Activity}
              label={t('flow.taskDetail.inProgress')}
              value={stats.inProgress}
              color="#3b82f6"
            />
            <StatCard
              icon={Clock}
              label={t('flow.projectOverview.total')}
              value={stats.total}
              color="var(--accent)"
            />
            <StatCard
              icon={AlertCircle}
              label={t('common.dashboard.stats.overdue')}
              value={stats.overdue}
              color="var(--flow-danger-fg)"
            />
          </div>
        </motion.section>

        {/* Milestones */}
        <section
          className="rounded-xl p-5"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <div className="flex items-center justify-between mb-3">
            <h3
              className="text-meta font-semibold uppercase tracking-wider"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.shell.milestones')}
            </h3>
            <MilestoneIcon size={12} style={{ color: 'var(--flow-text-dim)' }} />
          </div>
          {milestones.length === 0 ? (
            <EmptyState
              icon={MilestoneIcon}
              title={t('flow.projectOverview.noMilestones')}
              action={{
                label: t('flow.projectOverview.goToMilestonesCta'),
                onClick: () => navigate({ to: '/milestones' }),
              }}
              size="compact"
            />
          ) : (
            <div className="space-y-3">
              {milestones.map((m) => (
                <div key={m.id}>
                  <div className="flex items-center justify-between mb-1">
                    <span className="text-meta font-medium" style={{ color: 'var(--flow-text)' }}>
                      {m.name}
                    </span>
                    <span className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                      {formatDate(m.targetDate)}
                    </span>
                  </div>
                  <div
                    className="h-1.5 rounded-full overflow-hidden"
                    style={{ background: 'var(--flow-hover)' }}
                  >
                    <div
                      className="h-full"
                      style={{ width: `${m.progress}%`, background: 'var(--accent)' }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Team */}
        <section
          className="rounded-xl p-5"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <div className="flex items-center justify-between mb-3">
            <h3
              className="text-meta font-semibold uppercase tracking-wider"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.shell.team')}
            </h3>
            <Users size={12} style={{ color: 'var(--flow-text-dim)' }} />
          </div>
          <div className="flex items-center gap-3">
            <AvatarStack
              members={team.map((t) => ({ name: t.name, initials: t.initials }))}
              size={30}
              max={6}
            />
            <div className="flex-1 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {team.length} {team.length === 1 ? 'member' : 'members'}
            </div>
          </div>
        </section>

        {/* Recent activity */}
        <section
          className="rounded-xl p-5"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <h3
            className="text-meta font-semibold uppercase tracking-wider mb-3"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.projectOverview.recentComments')}
          </h3>
          {recentComments.length === 0 ? (
            <EmptyState
              icon={MessageSquare}
              title={t('flow.projectOverview.noComments')}
              action={{
                label: t('flow.projectOverview.goToBoardCta'),
                onClick: () =>
                  navigate({
                    to: '/p/$projectId/board',
                    params: { projectId },
                  }),
              }}
              size="compact"
            />
          ) : (
            <div className="space-y-2">
              {recentComments.map((c) => (
                <div key={c.id} className="flex items-start gap-2 text-meta">
                  <span
                    className="h-5 w-5 rounded-full flex items-center justify-center text-meta font-bold shrink-0 mt-0.5"
                    style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
                  >
                    {c.authorInitials}
                  </span>
                  <div className="flex-1 min-w-0">
                    <span className="font-semibold" style={{ color: 'var(--flow-text)' }}>
                      {c.author}
                    </span>{' '}
                    <span style={{ color: 'var(--flow-text-muted)' }}>{c.body}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
};

const StatCard: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  label: string;
  value: number;
  color: string;
}> = ({ icon: Icon, label, value, color }) => (
  <div
    className="rounded-lg p-3"
    style={{ background: 'var(--flow-hover)', border: '1px solid var(--flow-border)' }}
  >
    <div className="flex items-center gap-1.5 mb-1">
      <Icon size={11} />
      <span
        className="text-meta uppercase tracking-wider font-semibold"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        {label}
      </span>
    </div>
    <p className="text-title font-bold" style={{ color }}>
      {value}
    </p>
  </div>
);

export default ProjectOverview;
