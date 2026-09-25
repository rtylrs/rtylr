import React, { useState, useMemo, useCallback } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Plus, Clock, Trash2 } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListTimeEntries, useDeleteTimeEntry } from '../../shared/hooks/queries/useTimeEntries';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { getWeekRange } from '../../shared/utils/dateRange';
import type { TimeEntryPayload } from '@voxire/voxie/hr/time_entry/v1/types';
import { SectionHeader, Card, Pill, Skeleton, ACCENT } from '../../shared/components/ui';
import ActiveTimer from './ActiveTimer';
import TimeEntryForm from './TimeEntryForm';

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

function groupByDate(entries: TimeEntryPayload[]): Map<string, TimeEntryPayload[]> {
  const groups = new Map<string, TimeEntryPayload[]>();
  for (const entry of entries) {
    const existing = groups.get(entry.date ?? '') ?? [];
    existing.push(entry);
    groups.set(entry.date ?? '', existing);
  }
  return groups;
}

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-status-draft)',
  submitted: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  rejected: 'var(--hr-status-rejected)',
};

const TimeLogging = (): React.JSX.Element => {
  const { t } = useTranslation();
  const week = useMemo(getWeekRange, []);
  const {
    data: myEmployeeId,
    isError: employeeLookupFailed,
    refetch: refetchEmployeeId,
  } = useMyEmployeeId();
  const { data, isLoading } = useListTimeEntries(
    { date: week.start, employeeId: myEmployeeId === undefined ? undefined : String(myEmployeeId) },
    { enabled: myEmployeeId !== undefined }
  );
  const deleteEntry = useDeleteTimeEntry();

  const [formOpen, setFormOpen] = useState(false);
  const [prefill, setPrefill] = useState<{
    durationMinutes?: number;
    description?: string;
    source?: 'manual' | 'timer';
  }>({});

  const handleTimerStop = useCallback((durationMinutes: number, description: string): void => {
    setPrefill({ durationMinutes, description, source: 'timer' });
    setFormOpen(true);
  }, []);

  const handleNewEntry = useCallback((): void => {
    setPrefill({});
    setFormOpen(true);
  }, []);

  const entries = useMemo(() => data?.entries ?? [], [data]);
  const grouped = useMemo(() => groupByDate(entries), [entries]);
  const totalMinutes = entries.reduce((sum, e) => sum + (e.durationMinutes ?? 0), 0);

  const sortedDates = useMemo(() => {
    return Array.from(grouped.keys()).sort((a, b) => b.localeCompare(a));
  }, [grouped]);

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.timeLogging'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <SectionHeader
          title={t('timeLogging.title')}
          subtitle={t('timeLogging.subtitle', {
            start: week.start,
            end: week.end,
            total: formatDuration(totalMinutes),
          })}
        />

        <ActiveTimer onStop={handleTimerStop} />

        {employeeLookupFailed ? (
          <EmptyState
            icon={Clock}
            title={t('timeLogging.employeeLookupFailed')}
            description={t('errors.unauthorized')}
            action={{ label: t('common.retry'), onClick: () => refetchEmployeeId() }}
          />
        ) : isLoading || myEmployeeId === undefined ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <Skeleton height={60} />
            <Skeleton height={60} />
            <Skeleton height={60} />
          </div>
        ) : entries.length === 0 ? (
          <EmptyState
            icon={Clock}
            title={t('timeLogging.emptyTitle')}
            description={t('timeLogging.emptyDescription')}
            action={{ label: t('timeLogging.addEntry'), onClick: handleNewEntry }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            {sortedDates.map((date, dateIdx) => {
              const dayEntries = grouped.get(date) ?? [];
              const dayTotal = dayEntries.reduce((sum, e) => sum + (e.durationMinutes ?? 0), 0);
              const dateLabel = new Date(date + 'T00:00:00').toLocaleDateString(undefined, {
                weekday: 'short',
                month: 'short',
                day: 'numeric',
              });

              return (
                <motion.div
                  key={date}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: dateIdx * 0.04 }}
                >
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      marginBottom: 6,
                      padding: '0 4px',
                    }}
                  >
                    <span
                      style={{
                        fontSize: 12,
                        fontWeight: 600,
                        color: 'var(--hr-text-muted)',
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                      }}
                    >
                      {dateLabel}
                    </span>
                    <span style={{ fontSize: 12, fontWeight: 700, color: ACCENT }}>
                      {formatDuration(dayTotal)}
                    </span>
                  </div>
                  <Card style={{ padding: 0 }}>
                    {dayEntries.map((entry, i) => (
                      <div
                        key={entry.id}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: 10,
                          padding: '10px 14px',
                          borderTop: i > 0 ? '1px solid var(--hr-border)' : 'none',
                        }}
                      >
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div
                            style={{
                              fontSize: 13,
                              fontWeight: 500,
                              color: 'var(--hr-text)',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap',
                            }}
                          >
                            {entry.description || t('timeLogging.noDescription')}
                          </div>
                          <div
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 8,
                              marginTop: 3,
                              fontSize: 11,
                              color: 'var(--hr-text-muted)',
                            }}
                          >
                            {entry.startTime && entry.endTime && (
                              <span>
                                {entry.startTime} – {entry.endTime}
                              </span>
                            )}
                            {entry.isBillable && (
                              <Pill color="var(--hr-status-approved)" style={{ fontSize: 9 }}>
                                {t('timeLogging.billable')}
                              </Pill>
                            )}
                            <Pill
                              color={STATUS_COLORS[entry.status ?? ''] ?? 'var(--hr-status-draft)'}
                              style={{ fontSize: 9 }}
                            >
                              {t(`timeLogging.statuses.${entry.status}`)}
                            </Pill>
                          </div>
                        </div>
                        <span
                          style={{
                            fontSize: 13,
                            fontWeight: 700,
                            color: 'var(--hr-text)',
                            flexShrink: 0,
                          }}
                        >
                          {formatDuration(entry.durationMinutes ?? 0)}
                        </span>
                        {entry.status === 'draft' && (
                          <button
                            type="button"
                            onClick={() => deleteEntry.mutate(Number(entry.id ?? 0))}
                            style={{
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                              width: 28,
                              height: 28,
                              borderRadius: 6,
                              border: 'none',
                              background: 'transparent',
                              color: 'var(--hr-text-dim)',
                              cursor: 'pointer',
                              flexShrink: 0,
                            }}
                            aria-label={t('common.delete')}
                          >
                            <Trash2 size={13} />
                          </button>
                        )}
                      </div>
                    ))}
                  </Card>
                </motion.div>
              );
            })}
          </div>
        )}

        {/* Floating add button */}
        <button
          type="button"
          onClick={handleNewEntry}
          style={{
            position: 'fixed',
            bottom: 24,
            right: 24,
            width: 48,
            height: 48,
            borderRadius: '50%',
            background: ACCENT,
            color: 'var(--hr-on-accent, #fff)',
            border: 'none',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 12px var(--accent-border)',
            zIndex: 40,
          }}
          aria-label={t('timeLogging.addEntry')}
        >
          <Plus size={22} />
        </button>

        <TimeEntryForm open={formOpen} onClose={() => setFormOpen(false)} prefill={prefill} />
      </div>
    </>
  );
};

export default TimeLogging;
