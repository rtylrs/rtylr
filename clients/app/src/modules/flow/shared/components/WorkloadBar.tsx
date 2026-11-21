import React from 'react';

interface WorkloadBarProps {
  name: string;
  initials: string;
  load: number; // 0-1 where 1 = at capacity
  /** Contracted hours per week. Absent when the member has no work schedule (#868). */
  capacity?: number; // hours
  assigned: number; // hours
  /**
   * Overrides the colour thresholds. The team view turns amber at 70%; cycle
   * capacity turns amber at 90% and red only once the time is exceeded.
   */
  tone?: 'ok' | 'warn' | 'over';
}

const TONE_COLORS = {
  ok: 'var(--flow-status-approved)',
  warn: 'var(--flow-warn-fg)',
  over: 'var(--flow-danger-fg)',
} as const;

export const WorkloadBar: React.FC<WorkloadBarProps> = ({
  name,
  initials,
  load,
  capacity,
  assigned,
  tone,
}) => {
  const clamped = Math.min(1.5, Math.max(0, load));
  const color = TONE_COLORS[tone ?? (clamped < 0.7 ? 'ok' : clamped < 1 ? 'warn' : 'over')];
  return (
    <div className="flex items-center gap-3 py-2" data-tone={tone}>
      <span
        className="h-7 w-7 rounded-full flex items-center justify-center text-flow-meta font-bold shrink-0"
        style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
      >
        {initials}
      </span>
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-1">
          <span className="text-flow-meta font-medium" style={{ color: 'var(--flow-text)' }}>
            {name}
          </span>
          <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {capacity === undefined ? `${assigned}h` : `${assigned}h / ${capacity}h`}
          </span>
        </div>
        <div
          className="h-1.5 rounded-full overflow-hidden"
          style={{ background: 'var(--flow-hover)' }}
        >
          <div
            className="h-full transition-all rounded-full"
            style={{
              width: `${Math.min(100, clamped * 100)}%`,
              background: color,
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default WorkloadBar;
