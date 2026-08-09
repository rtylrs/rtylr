import React from 'react';
import { ACCENT } from './ui';

interface Props {
  progress: number; // 0-100
  label?: string;
  modules?: number;
}

const TrainingProgress: React.FC<Props> = ({ progress, label, modules }) => {
  const clamped = Math.max(0, Math.min(100, progress));
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      {label && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--hr-text)' }}>{label}</span>
          <span style={{ fontSize: 11, color: 'var(--hr-text-muted)' }}>
            {clamped}%{modules !== undefined && ` • ${modules} modules`}
          </span>
        </div>
      )}
      <div
        style={{
          height: 6,
          borderRadius: 999,
          background: 'var(--hr-hover)',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: `${clamped}%`,
            height: '100%',
            background: clamped >= 100 ? '#10b981' : ACCENT,
            transition: 'width 0.3s ease',
          }}
        />
      </div>
    </div>
  );
};

export default TrainingProgress;
