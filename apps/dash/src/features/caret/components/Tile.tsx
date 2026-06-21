import React from 'react';

export interface TileProps {
  title: string;
  value: React.ReactNode;
  hint?: string;
  accent?: string;
}

/** Small metric tile used in the Caret overview/dashboard. */
export const Tile: React.FC<TileProps> = ({ title, value, hint, accent = '#64c5b9' }) => (
  <div
    style={{
      background: 'var(--dash-surface)',
      border: '1px solid var(--dash-border)',
      borderRadius: 10,
      padding: 14,
      borderLeft: `3px solid ${accent}`,
    }}
  >
    <div style={{ fontSize: 11, color: 'var(--dash-text-dim)', textTransform: 'uppercase' }}>
      {title}
    </div>
    <div style={{ fontSize: 22, fontWeight: 600, marginTop: 2 }}>{value}</div>
    {hint && (
      <div style={{ fontSize: 12, color: 'var(--dash-text-muted)', marginTop: 2 }}>{hint}</div>
    )}
  </div>
);

export default Tile;
