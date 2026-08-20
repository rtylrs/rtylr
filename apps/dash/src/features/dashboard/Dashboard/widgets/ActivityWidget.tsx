import React from 'react';
import { useTranslation } from 'react-i18next';
import { Activity } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { EmptyState } from '@voxire/ui';
import { useActivity } from '../../../../shared/hooks/queries/useActivity';
import { ActivityEntryRow } from '../../../../shared/components/dash';
import WidgetHeader from './WidgetHeader';

const ActivityWidget: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: raw = [], refetch } = useActivity(5);
  const activity = Array.isArray(raw) ? raw : [];

  return (
    <div className="flex flex-col h-full">
      <WidgetHeader
        icon={Activity}
        title="Recent activity"
        action={{ label: 'View all', onClick: () => navigate({ to: '/audit' }) }}
      />
      <div className="flex-1 overflow-y-auto p-2">
        {activity.length === 0 ? (
          <EmptyState
            icon={Activity}
            title={t('dashboard.activityWidget.emptyTitle')}
            description={t('dashboard.activityWidget.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => void refetch() }}
            size="compact"
          />
        ) : (
          activity.map((e, i) => <ActivityEntryRow key={e.id} entry={e} index={i} />)
        )}
      </div>
    </div>
  );
};

export default ActivityWidget;
