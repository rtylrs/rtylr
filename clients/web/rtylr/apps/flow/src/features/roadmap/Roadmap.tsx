import React from 'react';
import { motion } from 'motion/react';
import { Link, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useRoadmap } from './useRoadmap';
import { formatDueDate, formatShortDate } from '../../shared/lib/utils';
import { displayDescription } from '../../shared/lib/stringUtils';
import { AvatarStack } from '../../shared/components/AssigneeAvatar';
import Breadcrumbs from '../../shared/components/Breadcrumbs';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { Map } from 'lucide-react';
import { EmptyState } from '@voxire/ui';

const Roadmap: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const roadmapQuery = useRoadmap();
  const { data: projects = [], isLoading } = roadmapQuery;

  if (isLoading)
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.roadmap.loading')}
      </div>
    );

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.shell.roadmap')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <Breadcrumbs items={[{ label: t('flow.shell.roadmap') }]} />
        <h1 className="text-title font-semibold mt-1" style={{ color: 'var(--flow-text)' }}>
          {t('flow.shell.roadmap')}
        </h1>
        <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.roadmap.subtitle')}
        </p>
      </div>
      <div className="flex-1 overflow-auto p-6 space-y-3">
        {/* A failed request and an empty roadmap both left this region blank,
            so a 500 read as "nothing planned" with nothing to act on. */}
        {hasQueryFailed(roadmapQuery) ? (
          <ListLoadFailed onRetry={() => void roadmapQuery.refetch()} />
        ) : projects.length === 0 ? (
          <EmptyState
            icon={Map}
            title={t('flow.roadmap.emptyTitle')}
            description={t('flow.roadmap.emptyDescription')}
            action={{
              label: t('flow.roadmap.emptyCta'),
              onClick: () => void navigate({ to: '/projects' }),
            }}
          />
        ) : (
          projects.map((p, i) => (
            <motion.div
              key={p.id}
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.18, delay: i * 0.04 }}
            >
              <Link
                to="/p/$projectId/overview"
                params={{ projectId: p.id }}
                className="block rounded-xl p-5"
                style={{
                  background: 'var(--flow-card)',
                  border: '1px solid var(--flow-border)',
                }}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <span
                      className="h-8 w-8 rounded-md flex items-center justify-center font-bold text-meta"
                      style={{ background: withAlpha(p.color, 13), color: onTint(p.color) }}
                    >
                      {p.name.charAt(0)}
                    </span>
                    <div>
                      <p className="text-section font-bold" style={{ color: 'var(--flow-text)' }}>
                        {p.name}
                      </p>
                      {displayDescription(p.description) && (
                        <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                          {displayDescription(p.description)}
                        </p>
                      )}
                    </div>
                  </div>
                  <AvatarStack
                    members={(p.memberAvatars ?? []).map((a) => ({ name: a, initials: a }))}
                    size={22}
                    max={4}
                  />
                </div>
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.roadmap.taskCount', {
                      done: p.doneCount,
                      total: p.tasks.length || p.taskCount,
                    })}
                  </span>
                  <span className="text-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {p.progress}%
                  </span>
                </div>
                <div
                  className="h-1.5 rounded-full overflow-hidden mb-3"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <div
                    className="h-full transition-all"
                    style={{ width: `${p.progress}%`, background: p.color }}
                  />
                </div>
                <div
                  className="flex items-center justify-between text-meta"
                  style={{ color: 'var(--flow-text-dim)' }}
                >
                  <span>
                    {t('flow.roadmap.start')} {formatShortDate(p.startDate, t('common.dateLocale'))}
                  </span>
                  <span>
                    {t('flow.roadmap.target')} {formatDueDate(p.targetDate, t)}
                  </span>
                </div>
              </Link>
            </motion.div>
          ))
        )}
      </div>
    </div>
  );
};

export default Roadmap;
