import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Timer, ArrowRight, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { DatePicker, EmptyState } from '@voxire/ui';
import { useCycles, useCreateCycle } from './useCycles';
import { type ApiCycle } from '../../shared/services/flowApi';
import { formatDate } from '../../shared/lib/utils';
import CycleBadge from '../../shared/components/CycleBadge';
import { useProjects } from '../projects/useProjects';
import GateNotice from '../../shared/components/GateNotice';

const NO_PROJECT_GATE_ID = 'cycles-project-gate';

const Cycles: React.FC = () => {
  const { t, i18n } = useTranslation();
  // Cycles belong to a project on v2, and this screen is business-level, so it
  // asks which project to show. Defaults to the first rather than rendering an
  // empty board with no explanation.
  const { data: projects = [] } = useProjects();
  const [projectId, setProjectId] = useState('');
  const selectedProjectId = projectId || projects[0]?.id || '';
  const { data: cycles = [], isLoading } = useCycles(selectedProjectId || undefined);
  const createCycle = useCreateCycle();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newStart, setNewStart] = useState('');
  const [newEnd, setNewEnd] = useState('');

  if (isLoading)
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.cycles.loading')}
      </div>
    );

  const active = cycles.find((c) => c.status === 'active');
  const upcoming = cycles.filter((c) => c.status === 'upcoming');
  const completed = cycles.filter((c) => c.status === 'completed');

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={`${t('flow.shell.cycles')} · Flow`} />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.shell.cycles')}
          </h1>
          <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.cycles.subtitle')}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <label className="flex items-center gap-1.5">
            {/* Visible, not sr-only. S-15 asks for a Project label, and a
                selector holding a 60-character truncated project name is
                exactly the case where the reader needs to know what the
                truncated thing is. */}
            <span
              className="hidden text-meta sm:inline"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.cycles.project')}
            </span>
            <select
              value={selectedProjectId}
              onChange={(e) => setProjectId(e.target.value)}
              aria-label={t('flow.cycles.project')}
              className="min-h-[44px] rounded-md px-2 text-title sm:min-h-0 sm:py-1.5 sm:text-meta"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            >
              {projects.length === 0 ? (
                <option value="">{t('flow.cycles.noProjects')}</option>
              ) : (
                projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))
              )}
            </select>
          </label>
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            disabled={!selectedProjectId}
            aria-describedby={selectedProjectId ? undefined : NO_PROJECT_GATE_ID}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-40"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            <Plus size={11} /> {t('flow.cycles.newCycle')}
          </button>
        </div>
      </div>

      {/* A cycle belongs to a project, so with no project there is nothing to
          create one against. The button was simply dim, which is the case
          root CLAUDE.md rule 12 exists for: state the reason or drop the
          control. */}
      {!selectedProjectId && (
        <GateNotice
          id={NO_PROJECT_GATE_ID}
          className="mx-6 mt-3"
          reason={t('flow.cycles.newCycleGate')}
        />
      )}

      {showCreate && (
        <div
          className="mx-6 mt-4 rounded-xl p-5 space-y-3"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder={t('flow.cycles.namePlaceholder')}
            className="w-full rounded-md px-3 py-2 text-meta"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
            autoFocus
          />
          <div className="flex gap-2">
            <DatePicker
              value={newStart}
              onChange={setNewStart}
              locale={i18n.language}
              aria-label={t('flow.cycles.startDate')}
              labels={{
                placeholder: t('flow.cycles.startDate'),
                previousMonth: t('common.datePicker.previousMonth'),
                nextMonth: t('common.datePicker.nextMonth'),
              }}
              className="text-meta"
            />
            <DatePicker
              value={newEnd}
              onChange={setNewEnd}
              minDate={newStart || undefined}
              locale={i18n.language}
              aria-label={t('flow.cycles.endDate')}
              labels={{
                placeholder: t('flow.cycles.endDate'),
                previousMonth: t('common.datePicker.previousMonth'),
                nextMonth: t('common.datePicker.nextMonth'),
              }}
              className="text-meta"
            />
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => {
                if (!newName.trim() || !newStart || !newEnd || !selectedProjectId) {
                  toast.error(t('errors.generic'));
                  return;
                }
                createCycle.mutate(
                  {
                    projectId: selectedProjectId,
                    name: newName,
                    startDate: newStart,
                    endDate: newEnd,
                  },
                  {
                    onSuccess: () => {
                      toast.success(t('toasts.cycle.created'));
                      setShowCreate(false);
                      setNewName('');
                      setNewStart('');
                      setNewEnd('');
                    },
                  }
                );
              }}
              disabled={createCycle.isPending || !newName.trim() || !newStart || !newEnd}
              className="px-3 py-1.5 rounded-sm text-meta font-semibold disabled:opacity-50"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {createCycle.isPending ? t('flow.cycles.creating') : t('flow.cycles.create')}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowCreate(false);
                setNewName('');
                setNewStart('');
                setNewEnd('');
              }}
              className="px-3 py-1.5 rounded-sm text-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.cycles.cancel')}
            </button>
          </div>
        </div>
      )}
      {cycles.length === 0 && !showCreate ? (
        <div className="flex-1 flex flex-col items-center justify-center p-6">
          <EmptyState
            icon={Timer}
            title={t('flow.cycles.emptyTitle')}
            description={t('flow.cycles.emptyDescription')}
            action={{
              label: t('flow.cycles.createFirst'),
              onClick: () => setShowCreate(true),
            }}
          />
        </div>
      ) : (
        <div className="p-6 max-w-4xl space-y-8">
          {active && (
            <motion.section
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              className="rounded-xl p-6"
              style={{
                background:
                  'linear-gradient(135deg, var(--accent-muted) 0%, rgba(100,197,185,0.02) 100%)',
                border: '1px solid var(--accent-border)',
              }}
            >
              <div className="flex items-center gap-2 mb-3">
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ background: 'var(--flow-status-approved)' }}
                />
                <span
                  className="text-meta uppercase tracking-wider font-semibold"
                  style={{ color: 'var(--flow-status-approved)' }}
                >
                  {t('flow.cycles.activeCycle')}
                </span>
              </div>
              <h2 className="text-screen font-bold mb-2" style={{ color: 'var(--flow-text)' }}>
                {active.name}
              </h2>
              <div
                className="flex items-center gap-2 text-meta"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <span>{formatDate(active.startDate)}</span>
                <ArrowRight size={10} />
                <span>{formatDate(active.endDate)}</span>
              </div>
              <div className="mt-4">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.cycles.taskCount', { done: active.completed, total: active.scope })}
                  </span>
                  <span className="text-meta font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {active.scope > 0 ? Math.round((active.completed / active.scope) * 100) : 0}%
                  </span>
                </div>
                <div
                  className="h-2 rounded-full overflow-hidden"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <div
                    className="h-full rounded-full"
                    style={{
                      width: `${active.scope > 0 ? (active.completed / active.scope) * 100 : 0}%`,
                      background: 'var(--accent)',
                    }}
                  />
                </div>
              </div>
            </motion.section>
          )}

          {upcoming.length > 0 && (
            <CycleSection title={t('flow.cycles.sectionUpcoming')} cycles={upcoming} />
          )}
          {completed.length > 0 && (
            <CycleSection title={t('flow.cycles.sectionCompleted')} cycles={completed} />
          )}
        </div>
      )}
    </div>
  );
};

const CycleSection: React.FC<{ title: string; cycles: ApiCycle[] }> = ({ title, cycles }) => {
  const { t } = useTranslation();
  return (
    <section>
      <h3
        className="text-meta font-semibold uppercase tracking-wider mb-3"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {title}
      </h3>
      <div className="space-y-2">
        {cycles.map((c) => (
          <div
            key={c.id}
            className="rounded-lg px-4 py-3 flex items-center gap-4"
            style={{
              background: 'var(--flow-card)',
              border: '1px solid var(--flow-border)',
            }}
          >
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <Timer size={12} style={{ color: 'var(--flow-text-muted)' }} />
                <span className="text-body font-semibold" style={{ color: 'var(--flow-text)' }}>
                  {c.name}
                </span>
              </div>
              <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                {t('flow.cycles.dateRange', {
                  start: formatDate(c.startDate),
                  end: formatDate(c.endDate),
                })}
              </p>
            </div>
            <CycleBadge cycle={c} compact />
          </div>
        ))}
      </div>
    </section>
  );
};

export default Cycles;
