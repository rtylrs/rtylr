import React, { useMemo } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Send, AlertCircle, FileText } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useGetTimesheet, useSubmitTimesheet } from '../../shared/hooks/queries/useTimesheets';
import type { TimeEntryPayload } from '@voxire/voxie/hr/time_entry/v1/types';
import {
  SectionHeader,
  Card,
  Button,
  Pill,
  Skeleton,
  Breadcrumbs,
  ACCENT,
} from '../../shared/components/ui';

const WEEKDAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;

function formatDuration(minutes: number): string {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  if (h === 0) return `${m}m`;
  if (m === 0) return `${h}h`;
  return `${h}h ${m}m`;
}

function getDayOfWeek(dateStr: string): string {
  const d = new Date(dateStr + 'T00:00:00');
  return WEEKDAYS[d.getDay() === 0 ? 6 : d.getDay() - 1];
}

function groupByDay(entries: TimeEntryPayload[]): Record<string, number> {
  const result: Record<string, number> = {};
  for (const day of WEEKDAYS) {
    result[day] = 0;
  }
  for (const entry of entries) {
    const day = getDayOfWeek(entry.date ?? '');
    result[day] = (result[day] ?? 0) + (entry.durationMinutes ?? 0);
  }
  return result;
}

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-status-draft)',
  submitted: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  rejected: 'var(--hr-status-rejected)',
  locked: 'var(--hr-status-locked)',
};

const TimesheetDetail = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const { id } = useParams({ strict: false });
  const navigate = useNavigate();
  const timesheetId = id != null ? Number(id) : undefined;
  const { data: timesheet, isLoading } = useGetTimesheet(timesheetId);
  const submitTimesheet = useSubmitTimesheet();

  const dayTotals = useMemo(
    () => (timesheet ? groupByDay(timesheet.entries ?? []) : null),
    [timesheet]
  );

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={80} />
        <Skeleton height={200} />
      </div>
    );
  }

  if (!timesheet) {
    return (
      <EmptyState
        icon={FileText}
        title={t('hrErrors.notFound')}
        action={{ label: t('hrCommon.back'), onClick: () => navigate({ to: '/hr/timesheets' }) }}
      />
    );
  }

  const entriesList = timesheet.entries ?? [];

  const periodLabel = `${new Date(timesheet.periodStart + 'T00:00:00').toLocaleDateString(
    undefined,
    {
      month: 'short',
      day: 'numeric',
    }
  )} – ${new Date(timesheet.periodEnd + 'T00:00:00').toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })}`;

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrCommon.pageTitles.timesheetDetail'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Breadcrumbs
          items={[
            { label: t('hrTimesheets.title'), href: '/hr/timesheets' },
            { label: periodLabel },
          ]}
        />

        <SectionHeader
          title={periodLabel}
          right={
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Pill color={STATUS_COLORS[timesheet.status ?? ''] ?? 'var(--hr-status-draft)'}>
                {t(`hrTimesheets.statuses.${timesheet.status}`)}
              </Pill>
              {timesheet.status === 'draft' && (
                <Button
                  variant="primary"
                  onClick={() => submitTimesheet.mutate({ timesheetId: Number(timesheet.id ?? 0) })}
                  disabled={submitTimesheet.isPending}
                  icon={<Send size={13} />}
                >
                  {t('hrTimesheets.submit')}
                </Button>
              )}
            </div>
          }
        />

        {/* Rejection reason. The reason the approver typed is stored
            server-side but is not part of TimesheetPayload yet, so the card
            must not claim that none was given. */}
        {timesheet.status === 'rejected' && (
          <Card
            style={{
              display: 'flex',
              alignItems: 'flex-start',
              gap: 10,
              padding: '12px 14px',
              border: '1px solid color-mix(in srgb, var(--hr-error) 30%, transparent)',
              background: 'color-mix(in srgb, var(--hr-error) 6%, transparent)',
            }}
          >
            <AlertCircle
              size={16}
              style={{ color: 'var(--hr-error)', flexShrink: 0, marginTop: 1 }}
            />
            <div style={{ fontSize: 12, color: 'var(--hr-text)' }}>
              <span style={{ fontWeight: 600 }}>{t('hrTimesheets.rejectionReason')}:</span>{' '}
              {t('hrTimesheets.rejectionReasonUnavailable')}
            </div>
          </Card>
        )}

        {/* Weekly grid */}
        <Card style={{ padding: 0, overflow: 'hidden' }}>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(7, 1fr)',
              borderBottom: '1px solid var(--hr-border)',
            }}
          >
            {WEEKDAYS.map((day) => (
              <div
                key={day}
                style={{
                  padding: '10px 8px',
                  textAlign: 'center',
                  fontSize: 10,
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  color: 'var(--hr-text-muted)',
                  letterSpacing: '0.05em',
                }}
              >
                {t(`hrTimesheets.days.${day}`)}
              </div>
            ))}
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)' }}>
            {WEEKDAYS.map((day) => {
              const mins = dayTotals?.[day] ?? 0;
              return (
                <div
                  key={day}
                  style={{
                    padding: '16px 8px',
                    textAlign: 'center',
                    fontSize: 14,
                    fontWeight: 700,
                    color: mins > 0 ? 'var(--hr-text)' : 'var(--hr-text-dim)',
                  }}
                >
                  {mins > 0 ? formatDuration(mins) : '-'}
                </div>
              );
            })}
          </div>
          <div
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              padding: '10px 14px',
              borderTop: '1px solid var(--hr-border)',
              fontSize: 13,
              fontWeight: 700,
              color: ACCENT,
            }}
          >
            {t('hrTimesheets.total')}: {formatDuration(timesheet.totalMinutes ?? 0)}
          </div>
        </Card>

        {/* Time entries list */}
        <h3
          style={{
            fontSize: 14,
            fontWeight: 600,
            color: 'var(--hr-text)',
            margin: '8px 0 0',
          }}
        >
          {t('hrTimesheets.entriesTitle')} ({entriesList.length})
        </h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {entriesList.length === 0 ? (
            <p style={{ fontSize: 12, color: 'var(--hr-text-muted)', padding: '8px 0' }}>
              {t('hrTimesheets.noEntries')}
            </p>
          ) : (
            entriesList.map((entry, idx) => (
              <motion.div
                key={entry.id}
                initial={{ opacity: 0, y: 4 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.02 }}
              >
                <Card
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    padding: '10px 14px',
                  }}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 12, fontWeight: 500, color: 'var(--hr-text)' }}>
                      {entry.description || t('hrTimeLogging.noDescription')}
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 2 }}>
                      {new Date(entry.date + 'T00:00:00').toLocaleDateString(undefined, {
                        weekday: 'short',
                        month: 'short',
                        day: 'numeric',
                      })}
                      {entry.startTime &&
                        entry.endTime &&
                        ` · ${entry.startTime} – ${entry.endTime}`}
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
                </Card>
              </motion.div>
            ))
          )}
        </div>
      </div>
    </>
  );
};

export default TimesheetDetail;
