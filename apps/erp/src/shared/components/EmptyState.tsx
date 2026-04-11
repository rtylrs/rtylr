import React from 'react';
import { PackageOpen } from 'lucide-react';

interface EmptyStateProps {
  icon?: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

const EmptyState: React.FC<EmptyStateProps> = ({
  icon: Icon = PackageOpen,
  title,
  description,
  action,
}) => {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div
        className="flex items-center justify-center h-16 w-16 rounded-full mb-4"
        style={{ background: 'rgba(245,158,11,0.08)' }}
      >
        <Icon size={28} style={{ color: '#f59e0b' }} />
      </div>
      <h3 className="text-sm font-semibold mb-1" style={{ color: 'var(--hub-text)' }}>
        {title}
      </h3>
      {description && (
        <p className="text-xs max-w-sm mb-4" style={{ color: 'var(--hub-text-muted)' }}>
          {description}
        </p>
      )}
      {action}
    </div>
  );
};

export default EmptyState;
