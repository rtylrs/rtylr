import React from 'react';
import { Inbox } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface EmptyStateProps {
  title?: string;
  description?: string;
  icon?: React.ComponentType<{ size?: number; className?: string }>;
}

const EmptyState: React.FC<EmptyStateProps> = ({ title, description, icon: Icon = Inbox }) => {
  const { t } = useTranslation();
  const defaultTitle = t('insights.empty.title', 'No data available');
  const defaultDescription = t(
    'insights.empty.description',
    'Data will appear here once it is available.'
  );

  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3">
      <Icon size={28} style={{ color: 'var(--hub-text-dim)' }} />
      <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>
        {title !== undefined ? title : defaultTitle}
      </p>
      <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
        {description !== undefined ? description : defaultDescription}
      </p>
    </div>
  );
};

export default EmptyState;
