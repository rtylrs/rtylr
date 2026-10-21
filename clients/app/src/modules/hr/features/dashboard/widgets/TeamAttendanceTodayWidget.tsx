import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users } from 'lucide-react';
import WidgetShell from './WidgetShell';
import { safeRatio, toFiniteNumber } from '@voxire/api';

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
  const { t } = useTranslation('hrCommon');
  const total = toFiniteNumber(data?.total);

  return (
    <WidgetShell
      title={t('hrDashboard.widgets.teamAttendance')}
      icon={<Users size={14} />}
      isLoading={isLoading}
      isEmpty={!data}
      emptyMessage={t('hrDashboard.widgets.noData')}
    >
      {data && (
        <>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
            {[
              {
                label: t('hrDashboard.widgets.present'),
                value: toFiniteNumber(data.present),
                color: 'var(--hr-status-approved)',
              },
              {
                label: t('hrDashboard.widgets.absent'),
                value: toFiniteNumber(data.absent),
                color: 'var(--hr-error)',
              },
              {
                label: t('hrDashboard.widgets.late'),
                value: toFiniteNumber(data.late),
                color: 'var(--hr-status-pending)',
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
          {total > 0 && (
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
                  width: `${safeRatio(data.present, total) * 100}%`,
                  background: 'var(--hr-status-approved)',
                }}
              />
              <div
                style={{
                  width: `${safeRatio(data.late, total) * 100}%`,
                  background: 'var(--hr-status-pending)',
                }}
              />
              <div
                style={{
                  width: `${safeRatio(data.absent, total) * 100}%`,
                  background: 'var(--hr-error)',
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
