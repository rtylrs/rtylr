import React from 'react';
import { Activity } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { useActivity } from '../../../../shared/hooks/queries/useActivity';
import { ActivityEntryRow } from '../../../../shared/components/dash';
import WidgetHeader from './WidgetHeader';

const ActivityWidget: React.FC = () => {
  const navigate = useNavigate();
  const { data: raw = [] } = useActivity(5);
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
          <div className="flex flex-col items-center justify-center h-full text-center py-6">
            <p className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
              No activity yet
            </p>
            <p className="text-[11px] mt-1" style={{ color: 'var(--dash-text-dim)' }}>
              Actions will appear here
            </p>
          </div>
        ) : (
          activity.map((e, i) => <ActivityEntryRow key={e.id} entry={e} index={i} />)
        )}
      </div>
    </div>
  );
};

export default ActivityWidget;
