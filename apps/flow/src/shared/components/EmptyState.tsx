import React from 'react';
import { LucideIcon } from 'lucide-react';

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon: Icon,
  title,
  description,
  action,
}) => {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div
        className="h-12 w-12 rounded-full flex items-center justify-center mb-3"
        style={{ background: 'var(--flow-hover)' }}
      >
        <Icon size={22} style={{ color: 'var(--flow-text-dim)' }} />
      </div>
      <h3 className="text-[14px] font-semibold mb-1" style={{ color: 'var(--flow-text)' }}>
        {title}
      </h3>
      {description && (
        <p className="text-[12px] max-w-sm mb-3" style={{ color: 'var(--flow-text-muted)' }}>
          {description}
        </p>
      )}
      {action}
    </div>
  );
};

export default EmptyState;
