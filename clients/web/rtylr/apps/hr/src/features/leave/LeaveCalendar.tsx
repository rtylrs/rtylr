import React, { useState, useMemo, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react';
import { useGetLeaveCalendar } from '../../shared/hooks/queries/useLeave';
import { Card, Button, Skeleton, Select, Label } from '../../shared/components/ui';
import { ErrorBoundary, EmptyState } from '@voxire/ui';
import { toLocalIso } from '../../shared/lib/dateUtils';

const DAYS_OF_WEEK_KEYS = [
  'timesheets.days.mon',
  'timesheets.days.tue',
  'timesheets.days.wed',
  'timesheets.days.thu',
  'timesheets.days.fri',
  'timesheets.days.sat',
  'timesheets.days.sun',
] as const;

const LeaveCalendarInner = (): React.JSX.Element => {
  const { t } = useTranslation();
  const now = new Date();
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [year, setYear] = useState(now.getFullYear());
  const [departmentId, setDepartmentId] = useState<number | undefined>(undefined);
  const [locationId, setLocationId] = useState<number | undefined>(undefined);

  const {
    data: entries,
    isLoading,
    refetch,
  } = useGetLeaveCalendar({
    month,
    year,
    departmentId,
    locationId,
  });

  const prevMonth = useCallback((): void => {
    if (month === 1) {
      setMonth(12);
      setYear((y) => y - 1);
    } else {
      setMonth((m) => m - 1);
    }
  }, [month]);

  const nextMonth = useCallback((): void => {
    if (month === 12) {
      setMonth(1);
      setYear((y) => y + 1);
    } else {
      setMonth((m) => m + 1);
    }
  }, [month]);

  const calendarDays = useMemo((): Date[] => {
    const firstDay = new Date(year, month - 1, 1);
    const lastDay = new Date(year, month, 0);
    const startOffset = (firstDay.getDay() + 6) % 7;
    const days: Date[] = [];

    for (let i = -startOffset; i <= lastDay.getDate() - 1; i++) {
      days.push(new Date(year, month - 1, i + 1));
    }
    const remaining = 7 - (days.length % 7);
    if (remaining < 7) {
      for (let i = 0; i < remaining; i++) {
        days.push(new Date(year, month, i + 1));
      }
    }
    return days;
  }, [month, year]);

  const getEntriesForDate = useCallback(
    (date: Date) => {
      if (!entries) return [];
      const dateStr = toLocalIso(date);
      return entries.filter((e) => dateStr >= (e.startDate ?? '') && dateStr <= (e.endDate ?? ''));
    },
    [entries]
  );

  const monthLabel = new Date(year, month - 1).toLocaleDateString('en-US', {
    month: 'long',
    year: 'numeric',
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: 12,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <Button variant="ghost" onClick={prevMonth} aria-label={t('leave.calendar.prevMonth')}>
            <ChevronLeft size={14} />
          </Button>
          <span
            style={{
              fontSize: 14,
              fontWeight: 700,
              color: 'var(--hr-text)',
              minWidth: 160,
              textAlign: 'center',
            }}
          >
            {monthLabel}
          </span>
          <Button variant="ghost" onClick={nextMonth} aria-label={t('leave.calendar.nextMonth')}>
            <ChevronRight size={14} />
          </Button>
        </div>

        <div style={{ display: 'flex', gap: 8 }}>
          <div>
            <Label style={{ marginBottom: 2 }}>{t('leave.calendar.department')}</Label>
            <Select
              value={departmentId ?? ''}
              onChange={(e) => setDepartmentId(e.target.value ? Number(e.target.value) : undefined)}
              style={{ width: 140, fontSize: 11, padding: '4px 8px' }}
            >
              <option value="">{t('common.filter')}</option>
            </Select>
          </div>
          <div>
            <Label style={{ marginBottom: 2 }}>{t('leave.calendar.location')}</Label>
            <Select
              value={locationId ?? ''}
              onChange={(e) => setLocationId(e.target.value ? Number(e.target.value) : undefined)}
              style={{ width: 140, fontSize: 11, padding: '4px 8px' }}
            >
              <option value="">{t('common.filter')}</option>
            </Select>
          </div>
        </div>
      </div>

      {isLoading ? (
        <Skeleton height={400} />
      ) : (
        <Card style={{ padding: 0, overflow: 'hidden' }}>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)' }}>
            {DAYS_OF_WEEK_KEYS.map((key) => (
              <div
                key={key}
                style={{
                  padding: '8px 4px',
                  textAlign: 'center',
                  fontSize: 10,
                  fontWeight: 600,
                  color: 'var(--hr-text-muted)',
                  textTransform: 'uppercase',
                  borderBottom: '1px solid var(--hr-border)',
                }}
              >
                {t(key)}
              </div>
            ))}

            {calendarDays.map((date, idx) => {
              const isCurrentMonth = date.getMonth() === month - 1;
              const isToday = date.toDateString() === now.toDateString();
              const dayEntries = getEntriesForDate(date);

              return (
                <div
                  key={idx}
                  style={{
                    minHeight: 72,
                    padding: 4,
                    borderBottom: '1px solid var(--hr-border)',
                    borderRight: (idx + 1) % 7 !== 0 ? '1px solid var(--hr-border)' : undefined,
                    opacity: isCurrentMonth ? 1 : 0.35,
                    background: isToday
                      ? 'color-mix(in srgb, var(--hr-accent) 3%, transparent)'
                      : undefined,
                  }}
                >
                  <div
                    style={{
                      fontSize: 11,
                      fontWeight: isToday ? 700 : 500,
                      color: isToday ? 'var(--hr-accent)' : 'var(--hr-text)',
                      marginBottom: 2,
                      textAlign: 'right',
                      padding: '2px 4px',
                    }}
                  >
                    {date.getDate()}
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                    {dayEntries.slice(0, 3).map((entry) => (
                      <div
                        key={entry.id}
                        style={{
                          fontSize: 9,
                          padding: '1px 4px',
                          borderRadius: 4,
                          background: 'color-mix(in srgb, var(--hr-accent) 13%, transparent)',
                          color: 'var(--hr-accent)',
                          fontWeight: 600,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        #{entry.employeeId}
                      </div>
                    ))}
                    {dayEntries.length > 3 && (
                      <div
                        style={{ fontSize: 9, color: 'var(--hr-text-dim)', textAlign: 'center' }}
                      >
                        +{dayEntries.length - 3}
                      </div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </Card>
      )}

      {!isLoading && (!entries || entries.length === 0) && (
        <EmptyState
          icon={Calendar}
          title={t('leave.calendar.emptyTitle')}
          description={t('leave.calendar.emptyDescription')}
          action={
            departmentId !== undefined || locationId !== undefined
              ? {
                  label: t('common.clearFilters'),
                  onClick: () => {
                    setDepartmentId(undefined);
                    setLocationId(undefined);
                  },
                }
              : { label: t('common.refresh'), onClick: () => refetch() }
          }
        />
      )}
    </div>
  );
};

const LeaveCalendar = (): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <ErrorBoundary
      fallback={
        <div style={{ padding: 24, textAlign: 'center', color: 'var(--hr-text-muted)' }}>
          {t('errors.calendarFailed')}
        </div>
      }
    >
      <LeaveCalendarInner />
    </ErrorBoundary>
  );
};

export default LeaveCalendar;
