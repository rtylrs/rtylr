import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { CalendarCheck } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useGetMyAttendance } from '../../shared/hooks/queries/useAttendance';
import type { AttendanceRecordPayload } from '@voxire/voxie/hr/attendance/v1/types';
import {
  SectionHeader,
  Card,
  Pill,
  QueryError,
  Skeleton,
  ACCENT,
} from '../../shared/components/ui';
import ClockInWidget from './ClockInWidget';

function formatTime(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

const STATUS_COLORS: Record<string, string> = {
  present: 'var(--hr-status-approved)',
  absent: 'var(--hr-status-rejected)',
  partial: 'var(--hr-status-submitted)',
};

const Attendance = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, isError, isFetching, refetch } = useGetMyAttendance({});

  const records = useMemo(() => data?.records ?? [], [data]);

  const activeRecord = useMemo((): AttendanceRecordPayload | undefined => {
    return records.find((r) => !r.clockOutAt);
  }, [records]);

  return (
    <>
      <HelmetContainer title="Attendance · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <SectionHeader title={t('attendance.title')} subtitle={t('attendance.subtitle')} />

        <ClockInWidget activeRecord={activeRecord} />

        <h3
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: 'var(--hr-text)',
            margin: '8px 0 0',
          }}
        >
          {t('attendance.history')}
        </h3>

        {isError ? (
          <QueryError onRetry={() => refetch()} isRetrying={isFetching} />
        ) : isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Skeleton height={56} />
            <Skeleton height={56} />
            <Skeleton height={56} />
            <Skeleton height={56} />
          </div>
        ) : records.length === 0 ? (
          <EmptyState
            icon={CalendarCheck}
            title={t('attendance.emptyTitle')}
            description={t('attendance.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <Card style={{ padding: 0, overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
              <thead>
                <tr style={{ background: 'var(--hr-hover)' }}>
                  {[
                    t('attendance.table.date'),
                    t('attendance.table.clockIn'),
                    t('attendance.table.clockOut'),
                    t('attendance.table.duration'),
                    t('attendance.table.method'),
                    t('attendance.table.status'),
                  ].map((header) => (
                    <th
                      key={header}
                      style={{
                        textAlign: 'left',
                        padding: '10px 12px',
                        color: 'var(--hr-text-muted)',
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        fontSize: 10,
                        letterSpacing: '0.05em',
                      }}
                    >
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {records.map((record, idx) => (
                  <motion.tr
                    key={record.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: idx * 0.02 }}
                    style={{ borderTop: idx > 0 ? '1px solid var(--hr-border)' : 'none' }}
                  >
                    <td style={{ padding: '12px', color: 'var(--hr-text)', fontWeight: 500 }}>
                      {new Date(record.date + 'T00:00:00').toLocaleDateString(undefined, {
                        weekday: 'short',
                        month: 'short',
                        day: 'numeric',
                      })}
                    </td>
                    <td style={{ padding: '12px', color: 'var(--hr-text)' }}>
                      {formatTime(Number(record.clockInAt ?? 0))}
                      {record.isLate && (
                        <Pill color="var(--hr-error)" style={{ marginLeft: 6, fontSize: 9 }}>
                          {t('attendance.late', { minutes: record.lateMinutes })}
                        </Pill>
                      )}
                    </td>
                    <td style={{ padding: '12px', color: 'var(--hr-text)' }}>
                      {record.clockOutAt ? formatTime(Number(record.clockOutAt)) : '-'}
                    </td>
                    <td
                      style={{
                        padding: '12px',
                        color: ACCENT,
                        fontWeight: 600,
                      }}
                    >
                      {record.durationMinutes ? formatDuration(record.durationMinutes) : '-'}
                    </td>
                    <td style={{ padding: '12px', color: 'var(--hr-text-muted)' }}>
                      {t(`attendance.methods.${record.clockInMethod}`)}
                    </td>
                    <td style={{ padding: '12px' }}>
                      <Pill color={STATUS_COLORS[record.status ?? ''] ?? 'var(--hr-status-draft)'}>
                        {t(`attendance.statuses.${record.status}`)}
                      </Pill>
                    </td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}
      </div>
    </>
  );
};

export default Attendance;
