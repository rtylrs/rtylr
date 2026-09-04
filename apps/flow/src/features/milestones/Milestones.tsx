import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Milestone as MilestoneIcon, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { DatePicker, EmptyState } from '@voxire/ui';
import { useMilestones, useCreateMilestone } from './useMilestones';
import { formatDate } from '../../shared/lib/utils';
import { CardSkeleton } from '../../shared/components/Skeleton';

const STATUS_STYLES: Record<string, { bg: string; color: string }> = {
  planning: { bg: 'rgba(107,114,128,0.15)', color: '#9ca3af' },
  active: { bg: 'rgba(34,197,94,0.15)', color: '#22c55e' },
  at_risk: { bg: 'rgba(245,158,11,0.15)', color: '#f59e0b' },
  completed: { bg: 'rgba(59,130,246,0.15)', color: '#3b82f6' },
};

const STATUS_LABEL_KEYS: Record<string, string> = {
  planning: 'planning',
  active: 'active',
  at_risk: 'atRisk',
  completed: 'completed',
};

import { useProjects } from '../projects/useProjects';

const Milestones: React.FC = () => {
  const { t, i18n } = useTranslation();
  // Business-wide list; v2 takes project_id only as an optional narrowing
  // filter, which this screen does not use.
  const { data: milestones = [], isLoading } = useMilestones();
  const { data: projects = [] } = useProjects();
  const createMilestone = useCreateMilestone();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDate, setNewDate] = useState('');
  // Every milestone belongs to a project on v2 (project_id is @required), so
  // the form has to ask. Defaults to the first project rather than leaving an
  // empty select that fails on submit.
  const [newProjectId, setNewProjectId] = useState('');
  const selectedProjectId = newProjectId || projects[0]?.id || '';

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
        <HelmetContainer title={`${t('flow.milestones.heading')} · Flow`} />
        <div
          className="flex items-center justify-between px-6 py-4 shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <div>
            <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
              {t('flow.milestones.heading')}
            </h1>
            <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.milestones.trackedCount', { count: 0 })}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-meta font-semibold"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            <Plus size={11} /> {t('flow.milestones.newMilestone')}
          </button>
        </div>
        <div className="flex-1 flex flex-col items-center justify-center p-6">
          <EmptyState
            icon={MilestoneIcon}
            title={t('flow.milestones.emptyTitle')}
            description={t('flow.milestones.emptyDescription')}
            action={{
              label: t('flow.milestones.createFirst'),
              onClick: () => setShowCreate(true),
            }}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.milestones.heading')} · Flow`} />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.milestones.heading')}
          </h1>
          <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.milestones.trackedCount', { count: milestones.length })}
          </p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-meta font-semibold"
          style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
        >
          <Plus size={11} /> {t('flow.milestones.newMilestone')}
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
            placeholder={t('flow.milestones.namePlaceholder')}
            className="w-full rounded-md px-3 py-2 text-meta"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
            autoFocus
          />
          <label className="block">
            <span className="mb-1 block text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.milestones.project')}
            </span>
            <select
              value={selectedProjectId}
              onChange={(e) => setNewProjectId(e.target.value)}
              className="w-full rounded-md px-3 py-2 text-title sm:text-meta"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            >
              {projects.length === 0 ? (
                <option value="">{t('flow.milestones.noProjects')}</option>
              ) : (
                projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))
              )}
            </select>
          </label>
          <DatePicker
            value={newDate}
            onChange={setNewDate}
            locale={i18n.language}
            aria-label={t('flow.milestones.targetDate')}
            labels={{
              placeholder: t('flow.milestones.targetDate'),
              previousMonth: t('common.datePicker.previousMonth'),
              nextMonth: t('common.datePicker.nextMonth'),
            }}
            className="text-meta"
          />
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => {
                if (!newName.trim() || !newDate || !selectedProjectId) {
                  toast.error(t('errors.generic'));
                  return;
                }
                createMilestone.mutate(
                  {
                    projectId: selectedProjectId,
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
              disabled={
                createMilestone.isPending || !newName.trim() || !newDate || !selectedProjectId
              }
              className="px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-50"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {createMilestone.isPending
                ? t('flow.milestones.creating')
                : t('flow.milestones.create')}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowCreate(false);
                setNewName('');
                setNewDate('');
              }}
              className="px-3 py-1.5 rounded-sm text-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.milestones.cancel')}
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
                    <p className="text-body font-bold" style={{ color: 'var(--flow-text)' }}>
                      {m.name}
                    </p>
                    <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                      {m.project?.name}
                    </p>
                  </div>
                </div>
                <span
                  className="text-meta px-2 py-0.5 rounded-full font-semibold capitalize"
                  style={{ background: style.bg, color: style.color }}
                >
                  {t(`flow.milestones.status.${STATUS_LABEL_KEYS[m.status]}`)}
                </span>
              </div>
              {m.description && (
                <p className="text-meta mb-3" style={{ color: 'var(--flow-text-muted)' }}>
                  {m.description}
                </p>
              )}
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {t('flow.milestones.progress')}
                </span>
                <span className="text-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
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
                className="flex items-center justify-between text-meta"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span>
                  {t('flow.milestones.taskCount', { done: m.doneCount, total: m.taskCount })}
                </span>
                <span>
                  {t('flow.milestones.target')} {formatDate(m.targetDate)}
                </span>
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
};

export default Milestones;
