import React from 'react';
import { Button } from '../ui';
import { cn } from '../../lib/utils';

export interface EmptyStateProps {
  icon?: React.ComponentType<{ size?: number; className?: string; style?: React.CSSProperties }>;
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
  className?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon: Icon,
  title,
  description,
  actionLabel,
  onAction,
  className,
}) => {
  return (
    <div
      className={cn(
        'rounded-xl px-8 py-14 flex flex-col items-center justify-center text-center',
        className
      )}
      style={{ background: 'var(--dash-card)', border: '1px dashed var(--dash-border)' }}
    >
      {Icon && (
        <div
          className="h-12 w-12 rounded-xl flex items-center justify-center mb-4"
          style={{ background: 'var(--dash-hover)' }}
        >
          <Icon size={20} style={{ color: 'var(--dash-text-dim)' }} />
        </div>
      )}
      <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
        {title}
      </p>
      {description && (
        <p className="text-xs mt-1 max-w-sm" style={{ color: 'var(--dash-text-muted)' }}>
          {description}
        </p>
      )}
      {actionLabel && onAction && (
        <Button className="mt-4" size="sm" onClick={onAction}>
          {actionLabel}
        </Button>
      )}
    </div>
  );
};

export default EmptyState;
