import React from 'react';
import { Skeleton } from '../../../shared/components/ui';

interface WidgetShellProps {
  title: string;
  icon?: React.ReactNode;
  isLoading?: boolean;
  isEmpty?: boolean;
  emptyMessage?: string;
  children: React.ReactNode;
}

const WidgetShell = ({
  title,
  icon,
  isLoading,
  isEmpty,
  emptyMessage,
  children,
}: WidgetShellProps): React.JSX.Element => {
  return (
    <div
      style={{
        background: 'var(--hr-card)',
        border: '1px solid var(--hr-border)',
        borderRadius: 12,
        padding: 16,
        display: 'flex',
        flexDirection: 'column',
        gap: 12,
        height: '100%',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        {icon && (
          <span
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: 28,
              width: 28,
              borderRadius: 8,
              background: 'var(--hr-hover)',
              color: 'var(--hr-text-muted)',
              flexShrink: 0,
            }}
          >
            {icon}
          </span>
        )}
        <h3
          style={{
            margin: 0,
            fontSize: 13,
            fontWeight: 700,
            color: 'var(--hr-text)',
          }}
        >
          {title}
        </h3>
      </div>
      {isLoading ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Skeleton height={20} width="60%" />
          <Skeleton height={14} width="80%" />
          <Skeleton height={14} width="40%" />
        </div>
      ) : isEmpty ? (
        <p
          style={{
            margin: 0,
            fontSize: 12,
            color: 'var(--hr-text-muted)',
            textAlign: 'center',
            padding: '12px 0',
          }}
        >
          {emptyMessage}
        </p>
      ) : (
        children
      )}
    </div>
  );
};

export default WidgetShell;
