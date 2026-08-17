import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Milestone as MilestoneIcon, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useMilestones, useCreateMilestone } from './useMilestones';
import { formatDate } from '../../shared/lib/utils';
import EmptyState from '../../shared/components/EmptyState';
import { CardSkeleton } from '../../shared/components/Skeleton';

const STATUS_STYLES: Record<string, { bg: string; color: string }> = {
  planning: { bg: 'rgba(107,114,128,0.15)', color: '#9ca3af' },
  active: { bg: 'rgba(34,197,94,0.15)', color: '#22c55e' },
  at_risk: { bg: 'rgba(245,158,11,0.15)', color: '#f59e0b' },
  completed: { bg: 'rgba(59,130,246,0.15)', color: '#3b82f6' },
};

import { useFlowStore } from '../../shared/stores/flowStore';

const Milestones: React.FC = () => {
  const { t } = useTranslation();
  const { activeWorkspaceId } = useFlowStore();
  const { data: milestones = [], isLoading } = useMilestones({
    workspaceId: activeWorkspaceId || undefined,
  });
  const createMilestone = useCreateMilestone();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDate, setNewDate] = useState('');

  if (isLoading) {
    return (
      <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  }

  if (milestones.length === 0 && !showCreate) {
    return (
      <div className="flex flex-col h-full">
        <HelmetContainer title="Milestones · rtylr" />
        <div
          className="flex items-center justify-between px-6 py-4 shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div>
            <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
              Milestones
            </h1>
            <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
              0 tracked
            </p>
          </div>
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
            style={{ background: 'var(--accent)', color: '#0a0a0a' }}
          >
            <Plus size={11} /> New milestone
          </button>
        </div>
        <div className="flex-1 flex flex-col items-center justify-center p-6">
          <EmptyState
            icon={MilestoneIcon}
            title="No milestones yet"
            description="Track major delivery targets across your projects."
          />
          <button
            onClick={() => setShowCreate(true)}
            className="mt-4 flex items-center gap-1.5 px-4 py-2 rounded-lg text-[12px] font-semibold"
            style={{ background: 'var(--accent)', color: '#0a0a0a' }}
          >
            <Plus size={13} /> Create first milestone
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title="Milestones · rtylr" />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            Milestones
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {milestones.length} tracked
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
        >
          <Plus size={11} /> New milestone
        </button>
      </div>
      {showCreate && (
        <div
          className="mx-6 mt-4 rounded-xl p-5 space-y-3"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="Milestone name"
            className="w-full rounded-md px-3 py-2 text-[12px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
            autoFocus
          />
          <input
            type="date"
            value={newDate}
            onChange={(e) => setNewDate(e.target.value)}
            className="rounded-md px-3 py-2 text-[12px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
          />
          <div className="flex gap-2">
            <button
              onClick={() => {
                if (!newName.trim() || !newDate) {
                  toast.error(t('errors.generic'));
                  return;
                }
                createMilestone.mutate(
                  {
                    workspaceId: activeWorkspaceId || undefined,
                    name: newName,
                    targetDate: newDate,
                  },
                  {
                    onSuccess: () => {
                      toast.success(t('toasts.milestoneCreated'));
                      setShowCreate(false);
                      setNewName('');
                      setNewDate('');
                    },
                    onError: () => toast.error(t('errors.generic')),
                  }
                );
              }}
              disabled={createMilestone.isPending || !newName.trim() || !newDate}
              className="px-3 py-1.5 rounded text-[11px] font-semibold disabled:opacity-50"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              {createMilestone.isPending ? 'Creating...' : 'Create'}
            </button>
            <button
              onClick={() => {
                setShowCreate(false);
                setNewName('');
                setNewDate('');
              }}
              className="px-3 py-1.5 rounded text-[11px]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
      <div className="flex-1 overflow-auto p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
        {milestones.map((m, i) => {
          const style = STATUS_STYLES[m.status];
          return (
            <motion.div
              key={m.id}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.16, delay: i * 0.03 }}
              className="rounded-xl p-5"
              style={{
                background: 'var(--flow-card)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span
                    className="h-8 w-8 rounded-md flex items-center justify-center"
                    style={{ background: `${m.project?.color || 'var(--accent)'}22` }}
                  >
                    <MilestoneIcon
                      size={14}
                      style={{ color: m.project?.color || 'var(--accent)' }}
                    />
                  </span>
                  <div>
                    <p className="text-[13px] font-bold" style={{ color: 'var(--flow-text)' }}>
                      {m.name}
                    </p>
                    <p className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                      {m.project?.name}
                    </p>
                  </div>
                </div>
                <span
                  className="text-[10px] px-2 py-0.5 rounded-full font-semibold capitalize"
                  style={{ background: style.bg, color: style.color }}
                >
                  {m.status.replace('_', ' ')}
                </span>
              </div>
              {m.description && (
                <p className="text-[12px] mb-3" style={{ color: 'var(--flow-text-muted)' }}>
                  {m.description}
                </p>
              )}
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
                  Progress
                </span>
                <span className="text-[11px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                  {m.computedProgress}%
                </span>
              </div>
              <div
                className="h-1.5 rounded-full overflow-hidden mb-3"
                style={{ background: 'var(--flow-hover)' }}
              >
                <div
                  className="h-full transition-all"
                  style={{ width: `${m.computedProgress}%`, background: style.color }}
                />
              </div>
              <div
                className="flex items-center justify-between text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span>
                  {m.doneCount} of {m.taskCount} tasks
                </span>
                <span>Target: {formatDate(m.targetDate)}</span>
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
};

export default Milestones;
