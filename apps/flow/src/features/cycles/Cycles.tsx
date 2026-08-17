import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Timer, ArrowRight, Plus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useCycles, useCreateCycle } from './useCycles';
import { type ApiCycle } from '../../shared/services/flowApi';
import { formatDate } from '../../shared/lib/utils';
import CycleBadge from '../../shared/components/CycleBadge';

const Cycles: React.FC = () => {
  const { t } = useTranslation();
  const { data: cycles = [], isLoading } = useCycles();
  const createCycle = useCreateCycle();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newStart, setNewStart] = useState('');
  const [newEnd, setNewEnd] = useState('');

  if (isLoading)
    return (
      <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
        Loading cycles...
      </div>
    );

  const active = cycles.find((c) => c.status === 'active');
  const upcoming = cycles.filter((c) => c.status === 'upcoming');
  const completed = cycles.filter((c) => c.status === 'completed');

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title="Cycles · rtylr" />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            Cycles
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            Two-week delivery sprints
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded text-[11px] font-semibold"
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
        >
          <Plus size={11} /> New cycle
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
            placeholder="Cycle name"
            className="w-full rounded-md px-3 py-2 text-[12px] outline-none"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border)',
              color: 'var(--flow-text)',
            }}
            autoFocus
          />
          <div className="flex gap-2">
            <input
              type="date"
              value={newStart}
              onChange={(e) => setNewStart(e.target.value)}
              className="rounded-md px-3 py-2 text-[12px] outline-none"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
            <input
              type="date"
              value={newEnd}
              onChange={(e) => setNewEnd(e.target.value)}
              className="rounded-md px-3 py-2 text-[12px] outline-none"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => {
                if (!newName.trim() || !newStart || !newEnd) {
                  toast.error(t('errors.generic'));
                  return;
                }
                createCycle.mutate(
                  {
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
              className="px-3 py-1.5 rounded text-[11px] font-semibold disabled:opacity-50"
              style={{ background: 'var(--accent)', color: '#0a0a0a' }}
            >
              {createCycle.isPending ? 'Creating...' : 'Create'}
            </button>
            <button
              onClick={() => {
                setShowCreate(false);
                setNewName('');
                setNewStart('');
                setNewEnd('');
              }}
              className="px-3 py-1.5 rounded text-[11px]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              Cancel
            </button>
          </div>
        </div>
      )}
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
              <span className="h-2 w-2 rounded-full" style={{ background: '#22c55e' }} />
              <span
                className="text-[10px] uppercase tracking-wider font-semibold"
                style={{ color: '#22c55e' }}
              >
                Active cycle
              </span>
            </div>
            <h2 className="text-[22px] font-bold mb-2" style={{ color: 'var(--flow-text)' }}>
              {active.name}
            </h2>
            <div
              className="flex items-center gap-2 text-[11px]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <span>{formatDate(active.startDate)}</span>
              <ArrowRight size={10} />
              <span>{formatDate(active.endDate)}</span>
            </div>
            <div className="mt-4">
              <div className="flex items-center justify-between mb-1.5">
                <span className="text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
                  {active.completed} of {active.scope} tasks done
                </span>
                <span className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
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

        {upcoming.length > 0 && <CycleSection title="Upcoming" cycles={upcoming} />}
        {completed.length > 0 && <CycleSection title="Completed" cycles={completed} />}
      </div>
    </div>
  );
};

const CycleSection: React.FC<{ title: string; cycles: ApiCycle[] }> = ({ title, cycles }) => (
  <section>
    <h3
      className="text-[11px] font-semibold uppercase tracking-wider mb-3"
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
              <span className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                {c.name}
              </span>
            </div>
            <p className="text-[10px]" style={{ color: 'var(--flow-text-dim)' }}>
              {formatDate(c.startDate)} to {formatDate(c.endDate)}
            </p>
          </div>
          <CycleBadge cycle={c} compact />
        </div>
      ))}
    </div>
  </section>
);

export default Cycles;
