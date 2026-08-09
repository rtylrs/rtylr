import React from 'react';
import { Activity, ArrowUpRight } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { useActivity } from '../../../../shared/hooks/queries/useActivity';
import { ActivityEntryRow } from '../../../../shared/components/dash';

const ActivityWidget: React.FC = () => {
  const navigate = useNavigate();
  const { data: raw = [] } = useActivity(5);
  const activity = Array.isArray(raw) ? raw : [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center justify-between px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <div className="flex items-center gap-2">
          <Activity size={14} style={{ color: 'var(--dash-text-muted)' }} />
          <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            Recent activity
          </p>
        </div>
        <button
          onClick={() => navigate({ to: '/audit' })}
          className="flex items-center gap-1 text-xs font-medium"
          style={{ color: '#4583C4' }}
        >
          View all <ArrowUpRight size={11} />
        </button>
      </div>
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
