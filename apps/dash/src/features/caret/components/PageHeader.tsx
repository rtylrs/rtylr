import React from 'react';
import { CaretLogo } from '@voxire/caret';

/**
 * Hero header used at the top of dedicated Caret pages (history, settings).
 * Kept separate from the panel HeaderBar so we can swap layouts without
 * coupling them.
 */
export const PageHeader: React.FC<{ title: string; subtitle?: string }> = ({ title, subtitle }) => (
  <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
    <span
      style={{
        width: 32,
        height: 32,
        borderRadius: 10,
        background: 'rgba(100,197,185,0.16)',
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexShrink: 0,
      }}
    >
      <CaretLogo variant="teal" size={26} />
    </span>
    <div>
      <h1 style={{ margin: 0, fontSize: 16, fontWeight: 600 }}>{title}</h1>
      {subtitle && (
        <p style={{ margin: '2px 0 0', fontSize: 12.5, color: 'var(--dash-text-muted)' }}>
          {subtitle}
        </p>
      )}
    </div>
  </div>
);

export default PageHeader;
