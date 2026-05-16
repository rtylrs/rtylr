import React from 'react';
import type { ToolDef } from '../../../shared/api/caret';
import ToolBadge from './ToolBadge';

export interface ToolsListProps {
  tools: ToolDef[];
}

/** Compact list of available tools, grouped by module. */
export const ToolsList: React.FC<ToolsListProps> = ({ tools }) => {
  const grouped: Record<string, ToolDef[]> = {};
  for (const t of tools) {
    (grouped[t.module] ??= []).push(t);
  }
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {Object.entries(grouped).map(([module, list]) => (
        <section key={module}>
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
            {module}
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
