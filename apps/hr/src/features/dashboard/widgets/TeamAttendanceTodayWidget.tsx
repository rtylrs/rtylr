import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users } from 'lucide-react';
import WidgetShell from './WidgetShell';

interface AttendanceData {
  present: number;
  absent: number;
  late: number;
  total: number;
}

interface TeamAttendanceTodayWidgetProps {
  data: AttendanceData | null;
  isLoading: boolean;
}

const TeamAttendanceTodayWidget = ({
  data,
  isLoading,
}: TeamAttendanceTodayWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <WidgetShell
      title={t('dashboard.widgets.teamAttendance')}
      icon={<Users size={14} />}
      isLoading={isLoading}
      isEmpty={!data}
      emptyMessage={t('dashboard.widgets.noData')}
    >
      {data && (
        <>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            {[
              {
                label: t('dashboard.widgets.present'),
                value: data.present,
                color: '#10b981',
              },
              {
                label: t('dashboard.widgets.absent'),
                value: data.absent,
                color: '#ef4444',
              },
              {
                label: t('dashboard.widgets.late'),
                value: data.late,
                color: '#f59e0b',
              },
            ].map((s) => (
              <div key={s.label} style={{ flex: 1, minWidth: 60, textAlign: 'center' }}>
                <div style={{ fontSize: 22, fontWeight: 700, color: s.color }}>{s.value}</div>
                <div style={{ fontSize: 10, color: 'var(--hr-text-muted)', marginTop: 2 }}>
                  {s.label}
                </div>
              </div>
            ))}
          </div>
          {data.total > 0 && (
            <div
              style={{
                display: 'flex',
                height: 6,
                borderRadius: 3,
                overflow: 'hidden',
                background: 'var(--hr-hover)',
              }}
            >
              <div
                style={{
                  width: `${(data.present / data.total) * 100}%`,
                  background: '#10b981',
                }}
              />
              <div
                style={{
                  width: `${(data.late / data.total) * 100}%`,
                  background: '#f59e0b',
                }}
              />
              <div
                style={{
                  width: `${(data.absent / data.total) * 100}%`,
                  background: '#ef4444',
                }}
              />
            </div>
          )}
        </>
      )}
    </WidgetShell>
  );
};

export default TeamAttendanceTodayWidget;
