import React from 'react';
import { useTranslation } from 'react-i18next';
import { Clock } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { EmptyState } from '@voxire/ui';
import { safe } from '../../../../shared/lib/api';
import WidgetHeader from './WidgetHeader';

interface PendingRequest {
  staffName?: string;
  staff_name?: string;
  type?: string;
  status?: string;
}

const PendingWidget: React.FC = () => {
  const { t } = useTranslation();
  const {
    data: raw,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['pending-widget'],
    queryFn: async (): Promise<PendingRequest[]> => {
      const res = await unwrap(
        safe.get<{ requests?: PendingRequest[] } | PendingRequest[]>('/hr/time-off/v1/requests')
      );
      return ((Array.isArray(res) ? res : res.requests) ?? []).filter(
        (request) => request.status?.toLowerCase() === 'pending'
      );
    },
    staleTime: 60_000,
  });
  const requests: PendingRequest[] = Array.isArray(raw)
    ? (raw as PendingRequest[]).slice(0, 5)
    : [];

  return (
    <div className="flex flex-col h-full">
      <WidgetHeader icon={Clock} iconColor="var(--status-trial)" title="Pending HR approvals" />
      <div className="flex-1 overflow-y-auto p-3 space-y-1.5">
        {isError ? (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-xs text-red-500">Failed to load pending approvals</p>
          </div>
        ) : requests.length === 0 ? (
          <EmptyState
            icon={Clock}
            title={t('dashboard.pendingWidget.emptyTitle')}
            description={t('dashboard.pendingWidget.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => void refetch() }}
            size="compact"
          />
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
