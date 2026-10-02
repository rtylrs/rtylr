import React from 'react';
import { StaffMember } from '../services/hrApi';
import { initials, paletteFor } from './EmployeeCard';

export interface OrgTreeNode {
  staff: StaffMember;
  children: OrgTreeNode[];
}

interface NodeProps {
  node: OrgTreeNode;
  depth?: number;
}

const OrgNode: React.FC<NodeProps> = ({ node, depth = 0 }) => {
  const { bg, fg } = paletteFor(node.staff.id);
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 16 }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '10px 14px',
          borderRadius: 12,
          border: '1px solid var(--hr-border)',
          background: 'var(--hr-card)',
          minWidth: 180,
          boxShadow: depth === 0 ? `0 0 0 2px ${bg}` : undefined,
        }}
      >
        <span
          style={{
            width: 32,
            height: 32,
            borderRadius: 8,
            background: bg,
            color: fg,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontWeight: 700,
            fontSize: 11,
          }}
        >
          {initials(node.staff.name)}
        </span>
        <div style={{ textAlign: 'left' }}>
          <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--hr-text)' }}>
            {node.staff.name}
          </div>
          <div style={{ fontSize: 10, color: 'var(--hr-text-muted)' }}>{node.staff.role}</div>
        </div>
      </div>
      {node.children.length > 0 && (
        <div style={{ position: 'relative', display: 'flex', gap: 24, paddingTop: 12 }}>
          <div
            style={{
              position: 'absolute',
              top: 0,
              left: '50%',
              width: 1,
              height: 12,
              background: 'var(--hr-border)',
            }}
          />
          {node.children.map((child) => (
            <OrgNode key={child.staff.id} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
};

export default OrgNode;
