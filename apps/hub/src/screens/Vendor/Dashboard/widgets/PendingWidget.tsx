import React from 'react';
import { Clock } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import api from '../../../../lib/api';

const PendingWidget: React.FC = () => {
  const { data: raw } = useQuery({
    queryKey: ['pending-widget'],
    queryFn: async () => {
      const res = await api.get('/hr/time-off?status=pending');
      return res.data?.requests ?? res.data ?? [];
    },
    staleTime: 60_000,
  });
  const requests = Array.isArray(raw) ? raw.slice(0, 5) : [];

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-3" style={{ borderBottom: '1px solid var(--hub-border)' }}>
        <Clock size={14} style={{ color: '#f59e0b' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Pending approvals</p>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {requests.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>Nothing pending</p>
          </div>
        ) : (
          requests.map((r: any, i: number) => (
            <div key={i} className="flex items-center gap-3 px-2 py-1.5 rounded-lg" style={{ background: 'var(--hub-bg)' }}>
              <div className="h-2 w-2 rounded-full shrink-0" style={{ background: '#f59e0b' }} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>{r.staffName || 'Team member'}</p>
                <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>{r.type || 'Time off request'}</p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default PendingWidget;
