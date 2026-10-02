import React from 'react';
import { useTranslation } from 'react-i18next';
import { CalendarDays } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import WidgetShell from './WidgetShell';

interface ShiftEntry {
  id: number;
  name: string;
  start: string;
  end: string;
  role: string;
}

interface UpcomingShiftsWidgetProps {
  data: ShiftEntry[] | null;
  isLoading: boolean;
  onRefresh: () => void;
}

const UpcomingShiftsWidget = ({
  data,
  isLoading,
  onRefresh,
}: UpcomingShiftsWidgetProps): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const isEmpty = !data || data.length === 0;

  return (
    <WidgetShell
      title={t('hrDashboard.widgets.upcomingShifts')}
      icon={<CalendarDays size={14} />}
      isLoading={isLoading}
    >
      {isEmpty ? (
        <EmptyState
          icon={CalendarDays}
          title={t('hrDashboard.widgets.noShifts')}
          action={{ label: t('hrCommon.refresh'), onClick: onRefresh }}
          size="compact"
        />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {(data ?? []).map((s) => (
            <div
              key={s.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '8px 10px',
                borderRadius: 8,
                background: 'var(--hr-hover)',
                gap: 8,
              }}
            >
              <div style={{ minWidth: 0 }}>
                <p
                  style={{
                    margin: 0,
                    fontSize: 12,
                    fontWeight: 600,
                    color: 'var(--hr-text)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {s.name}
                </p>
                <p style={{ margin: 0, fontSize: 10, color: 'var(--hr-text-muted)' }}>{s.role}</p>
              </div>
              <span
                style={{
                  fontSize: 11,
                  color: 'var(--hr-text-muted)',
                  flexShrink: 0,
                  whiteSpace: 'nowrap',
                }}
              >
                {s.start} – {s.end}
              </span>
            </div>
          ))}
        </div>
      )}
    </WidgetShell>
  );
};

export default UpcomingShiftsWidget;
