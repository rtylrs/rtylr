import React from 'react';
import { motion } from 'motion/react';
import { Link } from '@tanstack/react-router';
import { Map as MapIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useRoadmap } from './useRoadmap';
import { formatDate } from '../../shared/lib/utils';
import { AvatarStack } from '../../shared/components/AssigneeAvatar';
import Breadcrumbs from '../../shared/components/Breadcrumbs';

const Roadmap: React.FC = () => {
  const { t } = useTranslation();
  const { data: projects = [], isLoading } = useRoadmap();

  if (isLoading)
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        Loading roadmap...
      </div>
    );

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.roadmap.pageTitle')} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <Breadcrumbs items={[{ label: 'Roadmap' }]} />
        <h1 className="text-[15px] font-semibold mt-1" style={{ color: 'var(--flow-text)' }}>
          Roadmap
        </h1>
        <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
          High-level view of projects in flight
        </p>
      </div>
      <div className="flex-1 overflow-auto p-6 space-y-3">
        {projects.map((p, i) => (
          <motion.div
            key={p.id}
            initial={{ opacity: 0, x: -8 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.18, delay: i * 0.04 }}
          >
            <Link
              to={`/w/${p.workspaceId}/p/${p.id}/overview` as string}
              className="block rounded-xl p-5"
              style={{
                background: 'var(--flow-card)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                  <span
                    className="h-8 w-8 rounded-md flex items-center justify-center font-bold text-[11px]"
                    style={{ background: `${p.color}22`, color: p.color }}
                  >
                    {p.name.charAt(0)}
                  </span>
                  <div>
                    <p className="text-[14px] font-bold" style={{ color: 'var(--flow-text)' }}>
                      {p.name}
                    </p>
                    <p className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                      {p.description}
                    </p>
                  </div>
                </div>
                <AvatarStack
                  members={(p.memberAvatars ?? []).map((a) => ({ name: a, initials: a }))}
                  size={22}
                  max={4}
                />
              </div>
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                  {p.doneCount} of {p.tasks.length || p.taskCount} tasks done
                </span>
                <span className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
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
                className="flex items-center justify-between text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span>Start: {formatDate(p.startDate)}</span>
                <span>Target: {formatDate(p.targetDate)}</span>
              </div>
            </Link>
          </motion.div>
        ))}
      </div>
    </div>
  );
};

export default Roadmap;
