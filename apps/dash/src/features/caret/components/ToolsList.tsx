import React from 'react';
import type { ToolDef } from '../../../shared/api/caret';
import ToolBadge from './ToolBadge';

export interface ToolsListProps {
  tools: ToolDef[];
}

/** Compact list of available tools, grouped by category. */
export const ToolsList: React.FC<ToolsListProps> = ({ tools }) => {
  const grouped: Record<string, ToolDef[]> = {};
  for (const t of tools) {
    (grouped[t.category] ??= []).push(t);
  }
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {Object.entries(grouped).map(([category, list]) => (
        <section key={category}>
          <h3
            style={{
              margin: '0 0 6px',
              fontSize: 11,
              fontWeight: 600,
              textTransform: 'uppercase',
              color: 'var(--dash-text-dim)',
              letterSpacing: 0.4,
            }}
          >
            {category}
          </h3>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
            {list.map((t) => (
              <ToolBadge key={t.name} toolName={t.name} />
            ))}
          </div>
        </section>
      ))}
    </div>
  );
};

export default ToolsList;
