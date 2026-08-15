import React from 'react';
import { Inbox } from 'lucide-react';
import { ACCENT } from '../lib/tokens';

interface EmptyStateProps {
  icon?: React.ComponentType<{ size?: number; style?: React.CSSProperties; className?: string }>;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

const EmptyState: React.FC<EmptyStateProps> = ({
  icon: Icon = Inbox,
  title,
  description,
  action,
}) => (
  <div
    style={{
      padding: '48px 24px',
      textAlign: 'center',
      background: 'var(--hub-card)',
      border: '1px dashed var(--hub-border)',
      borderRadius: 12,
    }}
  >
    <div
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: 48,
        height: 48,
        borderRadius: 12,
        background: 'var(--finance-accent-soft)',
        marginBottom: 14,
      }}
    >
      <Icon size={22} style={{ color: ACCENT }} />
    </div>
    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hub-text)', marginBottom: 6 }}>
      {title}
    </div>
    {description && (
      <div
        style={{
          fontSize: 12,
          color: 'var(--hub-text-muted)',
          marginBottom: 18,
          maxWidth: 360,
          margin: '0 auto 18px',
        }}
      >
        {description}
      </div>
    )}
    {action && <div style={{ marginTop: 8 }}>{action}</div>}
  </div>
);

export default EmptyState;
