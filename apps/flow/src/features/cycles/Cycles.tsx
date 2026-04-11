import React from 'react';
import { motion } from 'motion/react';
import { Timer, ArrowRight } from 'lucide-react';
import { useCycles } from './useCycles';
import { formatDate } from '../../shared/lib/utils';
import CycleBadge from '../../shared/components/CycleBadge';

const Cycles: React.FC = () => {
  const { data: cycles = [], isLoading } = useCycles();

  if (isLoading) return <div className="p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>Loading cycles...</div>;

  const active = cycles.find((c) => c.status === 'active');
  const upcoming = cycles.filter((c) => c.status === 'upcoming');
  const completed = cycles.filter((c) => c.status === 'completed');

  return (
    <div className="flex flex-col h-full overflow-auto">
      <div
        className="px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
          Cycles
        </h1>
        <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
          Two-week delivery sprints
        </p>
      </div>

      <div className="p-6 max-w-4xl space-y-8">
        {active && (
          <motion.section
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            className="rounded-xl p-6"
            style={{
              background:
                'linear-gradient(135deg, rgba(100,197,185,0.08) 0%, rgba(100,197,185,0.02) 100%)',
              border: '1px solid rgba(100,197,185,0.3)',
            }}
          >
            <div className="flex items-center gap-2 mb-3">
              <span
                className="h-2 w-2 rounded-full"
                style={{ background: '#22c55e' }}
              />
              <span className="text-[10px] uppercase tracking-wider font-semibold"
                style={{ color: '#22c55e' }}>
                Active cycle
              </span>
            </div>
            <h2 className="text-[22px] font-bold mb-2" style={{ color: 'var(--flow-text)' }}>
              {active.name}
            </h2>
            <div className="flex items-center gap-2 text-[11px]"
              style={{ color: 'var(--flow-text-muted)' }}>
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
              <div className="h-2 rounded-full overflow-hidden"
                style={{ background: 'var(--flow-hover)' }}>
                <div
                  className="h-full rounded-full"
                  style={{
                    width: `${active.scope > 0 ? (active.completed / active.scope) * 100 : 0}%`,
                    background: '#64C5B9',
                  }}
                />
              </div>
            </div>
          </motion.section>
        )}

        {upcoming.length > 0 && (
          <CycleSection title="Upcoming" cycles={upcoming} />
        )}
        {completed.length > 0 && (
          <CycleSection title="Completed" cycles={completed} />
        )}
      </div>
    </div>
  );
};

const CycleSection: React.FC<{ title: string; cycles: any[] }> = ({ title, cycles }) => (
  <section>
    <h3 className="text-[11px] font-semibold uppercase tracking-wider mb-3"
      style={{ color: 'var(--flow-text-muted)' }}>
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
