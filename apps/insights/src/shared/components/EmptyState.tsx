import React from 'react';
import { Inbox } from 'lucide-react';

interface EmptyStateProps {
  title?: string;
  description?: string;
  icon?: React.ComponentType<{ size?: number; className?: string }>;
}

const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'No data available',
  description = 'Data will appear here once it is available.',
  icon: Icon = Inbox,
}) => (
  <div className="flex flex-col items-center justify-center py-16 gap-3">
    <Icon size={28} style={{ color: 'var(--hub-text-dim)' }} />
    <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>
      {title}
    </p>
    <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
      {description}
    </p>
  </div>
);

export default EmptyState;
