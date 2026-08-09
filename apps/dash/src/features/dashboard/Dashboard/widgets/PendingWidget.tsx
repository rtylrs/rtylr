import React from 'react';
import { Clock } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../../../shared/lib/api';

interface PendingRequest {
  staffName?: string;
  staff_name?: string;
  type?: string;
  status?: string;
}

const PendingWidget: React.FC = () => {
  const { data: raw } = useQuery({
    queryKey: ['pending-widget'],
    queryFn: async (): Promise<PendingRequest[]> => {
      try {
        const res = await unwrap(
          safe.get<{ requests?: PendingRequest[] } | PendingRequest[]>('/hr/time-off/v1/requests')
        );
        return ((Array.isArray(res) ? res : res.requests) ?? []).filter(
          (request) => request.status?.toLowerCase() === 'pending'
        );
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });
  const requests: PendingRequest[] = Array.isArray(raw)
    ? (raw as PendingRequest[]).slice(0, 5)
    : [];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center gap-2 px-4 py-3"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <Clock size={14} style={{ color: '#f59e0b' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Pending approvals
        </p>
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {requests.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              Nothing pending
            </p>
          </div>
        ) : (
          requests.map((r, i) => (
            <div
              key={i}
              className="flex items-center gap-3 px-2 py-1.5 rounded-lg"
              style={{ background: 'var(--dash-bg)' }}
            >
              <div className="h-2 w-2 rounded-full shrink-0" style={{ background: '#f59e0b' }} />
              <div className="flex-1 min-w-0">
                <p className="text-xs font-medium truncate" style={{ color: 'var(--dash-text)' }}>
                  {r.staffName ?? r.staff_name ?? 'Team member'}
                </p>
                <p className="text-[10px]" style={{ color: 'var(--dash-text-dim)' }}>
                  {r.type || 'Time off request'}
                </p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default PendingWidget;
