import React from 'react';
import { useTranslation } from 'react-i18next';
import { LogIn, LogOut, QrCode } from 'lucide-react';
import { useClockIn, useClockOut } from '../../shared/hooks/queries/useAttendance';
import { useHRConfigStore } from '../../shared/stores/hrConfigStore';
import type { AttendanceRecordPayload } from '@voxire/contracts/hr/v1/attendance';
import { Card, Button, Pill, ACCENT } from '../../shared/components/ui';

interface ClockInWidgetProps {
  activeRecord?: AttendanceRecordPayload;
}

function formatTime(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  });
}

const ClockInWidget = ({ activeRecord }: ClockInWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();
  const config = useHRConfigStore((s) => s.config);
  const clockIn = useClockIn();
  const clockOut = useClockOut();

  const isClockedIn = activeRecord && !activeRecord.clockOutAt;

  const handleClockIn = (): void => {
    clockIn.mutate({ method: 'app', locationId: 0 });
  };

  const handleClockOut = (): void => {
    if (!activeRecord) return;
    clockOut.mutate({ attendanceRecordId: activeRecord.id });
  };

  return (
    <Card
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 12,
        padding: 16,
        border: isClockedIn ? `1px solid ${ACCENT}40` : undefined,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--hr-text)' }}>
          {t('attendance.clockWidget.title')}
        </span>
        <Pill color={isClockedIn ? 'var(--hr-status-active)' : 'var(--hr-status-inactive)'}>
          {isClockedIn
            ? t('attendance.clockWidget.clockedIn')
            : t('attendance.clockWidget.clockedOut')}
        </Pill>
      </div>

      {isClockedIn && activeRecord ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ fontSize: 12, color: 'var(--hr-text-muted)' }}>
            {t('attendance.clockWidget.since', { time: formatTime(activeRecord.clockInAt) })}
            {activeRecord.isLate && (
              <Pill color="var(--hr-error)" style={{ marginLeft: 8, fontSize: 9 }}>
                {t('attendance.late', { minutes: activeRecord.lateMinutes })}
              </Pill>
            )}
          </div>
          <Button
            variant="danger"
            onClick={handleClockOut}
            disabled={clockOut.isPending}
            icon={<LogOut size={14} />}
          >
            {t('attendance.clockWidget.clockOutBtn')}
          </Button>
        </div>
      ) : (
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
          <Button
            variant="primary"
            onClick={handleClockIn}
            disabled={clockIn.isPending}
            icon={<LogIn size={14} />}
          >
            {t('attendance.clockWidget.clockInApp')}
          </Button>
          <Button
            variant="secondary"
            onClick={() => alert(t('attendance.qrComingSoon'))}
            icon={<QrCode size={14} />}
          >
            {t('attendance.clockWidget.clockInQR')}
          </Button>
        </div>
      )}
    </Card>
  );
};

export default ClockInWidget;
