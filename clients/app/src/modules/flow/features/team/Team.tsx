import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { motion } from 'motion/react';
import { Users } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useTeamWorkload } from './useTeamWorkload';
import WorkloadBar from '../../shared/components/WorkloadBar';
import PriorityIcon from '../../shared/components/PriorityIcon';
import StatusIcon from '../../shared/components/StatusIcon';
import { useNavigate } from '@tanstack/react-router';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const Team: React.FC = () => {
  const { t } = useTranslation();
  const openTask = useOpenTask();
  const teamQuery = useTeamWorkload();
  const { data: members = [] } = teamQuery;
  const loadFailed = hasQueryFailed(teamQuery);
  const isLoading = teamQuery.isPending && !loadFailed;
  const navigate = useNavigate();

  if (isLoading)
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.team.loading')}
      </div>
    );

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={`${t('flow.shell.team')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <div className="flex items-center gap-2">
          <Users size={14} style={{ color: 'var(--flow-text-muted)' }} />
          <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.shell.team')}
          </h1>
        </div>
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.team.memberCount', { count: members.length })}
        </p>
      </div>
      <div className="p-6 max-w-4xl space-y-4">
        {/* A failed request and an empty team both left this region blank. */}
        {loadFailed ? (
          <ListLoadFailed onRetry={() => void teamQuery.refetch()} />
        ) : members.length === 0 ? (
          <EmptyState
            icon={Users}
            title={t('flow.team.emptyTitle')}
            description={t('flow.team.emptyDescription')}
            action={{
              label: t('flow.team.emptyCta'),
              onClick: () => void navigate({ to: '/flow/projects' }),
            }}
          />
        ) : (
          members.map((m, i) => (
            <motion.div
              key={m.id}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.18, delay: i * 0.04 }}
              className="rounded-xl p-5"
              style={{
                background: 'var(--flow-card)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <WorkloadBar
                name={m.name}
                initials={m.initials}
                load={m.load}
                capacity={m.capacity}
                assigned={m.assignedHours}
              />
              {m.tasks.length > 0 && (
                <div
                  className="mt-3 pt-3 space-y-1.5"
                  style={{ borderTop: '1px solid var(--flow-border)' }}
                >
                  {m.tasks.slice(0, 5).map((t) => (
                    <button
                      type="button"
                      key={t.id}
                      onClick={() => openTask(String(t.id))}
                      className="w-full flex items-center gap-2 py-1 text-left hover:bg-[var(--flow-hover)] rounded-sm px-1.5"
                    >
                      <PriorityIcon priority={t.priority} size={11} />
                      <StatusIcon status={t.status} size={11} />
                      <span
                        className="text-flow-meta flex-1 truncate"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {t.title}
                      </span>
                      <span
                        className="text-flow-meta shrink-0"
                        style={{ color: 'var(--flow-text-dim)' }}
                      >
                        {t.estimate || 4}h
                      </span>
                    </button>
                  ))}
                  {m.tasks.length > 5 && (
                    <p className="text-flow-meta px-1.5" style={{ color: 'var(--flow-text-dim)' }}>
                      {t('flow.team.moreCount', { count: m.tasks.length - 5 })}
                    </p>
                  )}
                </div>
              )}
            </motion.div>
          ))
        )}
      </div>
    </div>
  );
};

export default Team;
