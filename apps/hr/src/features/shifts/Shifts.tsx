import React, { useState, useMemo } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { motion } from 'motion/react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Clock, CheckCircle2, Play, Calendar, ChevronLeft, ChevronRight } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import type { Shift } from '../../shared/schemas/shifts';
import { useShifts } from './useShifts';
import { SectionHeader, Skeleton, Button, Pill, useIsMobile } from '../../shared/components/ui';

const ACCENT = 'var(--hr-accent)';

const DAY_KEYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const;
const todayIdx = new Date().getDay() === 0 ? 6 : new Date().getDay() - 1;

const statusConfig: Record<string, { color: string; bg: string; icon: typeof Clock }> = {
  Scheduled: { color: ACCENT, bg: 'var(--accent-muted)', icon: Calendar },
  'In Progress': {
    color: 'var(--hr-shift-green)',
    bg: 'color-mix(in srgb, var(--hr-shift-green) 12%, transparent)',
    icon: Play,
  },
  Completed: {
    color: 'var(--hr-status-draft)',
    bg: 'color-mix(in srgb, var(--hr-status-draft) 12%, transparent)',
    icon: CheckCircle2,
  },
};

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const container = { hidden: {}, show: { transition: { staggerChildren: 0.06 } } };
const rowAnim = {
  hidden: { opacity: 0, x: -8 },
  show: { opacity: 1, x: 0, transition: { duration: 0.2 } },
};

function getWeekRange(offset: number): { start: Date; end: Date } {
  const now = new Date();
  const dayOfWeek = now.getDay() === 0 ? 6 : now.getDay() - 1;
  const start = new Date(now);
  start.setDate(now.getDate() - dayOfWeek + offset * 7);
  const end = new Date(start);
  end.setDate(start.getDate() + 6);
  return { start, end };
}

function formatDate(date: Date): string {
  return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

const Shifts = (): React.JSX.Element => {
  const { t } = useTranslation();
  const isMobile = useIsMobile(900);
  const [weekOffset, setWeekOffset] = useState(0);
  const [selectedDay, setSelectedDay] = useState(todayIdx);
  const week = useMemo(() => getWeekRange(weekOffset), [weekOffset]);
  const { data: fetchedShifts, isLoading, refetch } = useShifts();
  const shifts: Shift[] = fetchedShifts ?? [];
  const days = DAY_KEYS.map((k) => t(`timesheets.days.${k}`));

  const dayShifts = shifts.filter((s) => s.day === days[selectedDay]);
  const todayShifts = shifts.filter((s) => s.day === days[todayIdx]);
  const scheduled = todayShifts.filter((s) => s.status === 'Scheduled').length;
  const inProgress = todayShifts.filter((s) => s.status === 'In Progress').length;
  const completed = todayShifts.filter((s) => s.status === 'Completed').length;

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <Skeleton height={32} width={200} />
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} height={80} />
          ))}
        </div>
        <Skeleton height={300} />
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.shifts'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('shifts.title')}
          subtitle={t('shifts.weekOf', {
            start: formatDate(week.start),
            end: formatDate(week.end),
          })}
          right={
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Button
                variant="ghost"
                onClick={() => setWeekOffset((o) => o - 1)}
                aria-label={t('shifts.prevWeek')}
              >
                <ChevronLeft size={14} />
              </Button>
              <Button
                variant="ghost"
                onClick={() => setWeekOffset((o) => o + 1)}
                aria-label={t('shifts.nextWeek')}
              >
                <ChevronRight size={14} />
              </Button>
            </div>
          }
        />

        {/* Stats row */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            gap: 12,
          }}
        >
          {[
            {
              label: t('shifts.statuses.scheduled'),
              value: scheduled,
              color: ACCENT,
              bg: 'rgba(139,92,246,0.12)',
            },
            {
              label: t('shifts.statuses.inProgress'),
              value: inProgress,
              color: 'var(--hr-shift-green)',
              bg: 'color-mix(in srgb, var(--hr-shift-green) 12%, transparent)',
            },
            {
              label: t('shifts.statuses.completed'),
              value: completed,
              color: 'var(--hr-status-draft)',
              bg: 'color-mix(in srgb, var(--hr-status-draft) 12%, transparent)',
            },
          ].map((s) => (
            <div
              key={s.label}
              style={{
                background: s.bg,
                borderRadius: 10,
                padding: '14px 18px',
                border: `1px solid ${s.color}30`,
              }}
            >
              <p style={{ margin: 0, fontSize: 24, fontWeight: 700, color: s.color }}>{s.value}</p>
              <p style={{ margin: '4px 0 0', fontSize: 12, color: s.color, fontWeight: 500 }}>
                {s.label} {t('shifts.today')}
              </p>
            </div>
          ))}
        </div>

        {/* Week selector */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(7, minmax(72px, 1fr))',
            gap: 6,
            overflowX: 'auto',
          }}
        >
          {days.map((day, i) => {
            const count = shifts.filter((s) => s.day === day).length;
            const isToday = i === todayIdx && weekOffset === 0;
            const isSelected = i === selectedDay;
            const dateNum = new Date(week.start.getTime() + i * 86400000).getDate();
            return (
              <button
                type="button"
                key={day}
                onClick={() => setSelectedDay(i)}
                style={{
                  padding: '10px 6px',
                  borderRadius: 10,
                  border: isSelected ? `1px solid ${ACCENT}` : '1px solid var(--hr-border)',
                  background: isSelected ? 'var(--accent-muted)' : 'var(--hr-card)',
                  cursor: 'pointer',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 4,
                }}
              >
                <span
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    color: isToday ? ACCENT : 'var(--hr-text-muted)',
                    textTransform: 'uppercase',
                  }}
                >
                  {day}
                </span>
                <span
                  style={{
                    fontSize: 14,
                    fontWeight: 700,
                    color: isSelected ? ACCENT : 'var(--hr-text)',
                  }}
                >
                  {dateNum}
                </span>
                {count > 0 && (
                  <span
                    style={{
                      width: 6,
                      height: 6,
                      borderRadius: '50%',
                      background: isSelected ? ACCENT : 'var(--hr-text-dim)',
                    }}
                  />
                )}
              </button>
            );
          })}
        </div>

        {/* Shift list */}
        <motion.div
          variants={container}
          initial="hidden"
          animate="show"
          key={`${weekOffset}-${selectedDay}`}
          style={{
            background: 'var(--hr-card)',
            border: '1px solid var(--hr-border)',
            borderRadius: 12,
            overflow: 'hidden',
          }}
        >
          <div style={{ padding: '14px 20px', borderBottom: '1px solid var(--hr-border)' }}>
            <h3 style={{ margin: 0, fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
              {days[selectedDay]}: {dayShifts.length}{' '}
              {t('shifts.shiftsCount', { count: dayShifts.length })}
            </h3>
          </div>

          {dayShifts.length === 0 ? (
            <EmptyState
              icon={Calendar}
              title={t('shifts.emptyDay')}
              description={t('shifts.emptyDayDescription')}
              action={{ label: t('common.refresh'), onClick: () => refetch() }}
            />
          ) : (
            dayShifts.map((s, i) => {
              const cfg = statusConfig[s.status] ?? statusConfig.Scheduled;
              const StatusIcon = cfg.icon;
              return (
                <motion.div
                  key={s.id}
                  variants={rowAnim}
                  style={{
                    display: 'flex',
                    alignItems: isMobile ? 'flex-start' : 'center',
                    flexDirection: isMobile ? 'column' : 'row',
                    gap: 14,
                    padding: '14px 20px',
                    borderBottom: i < dayShifts.length - 1 ? '1px solid var(--hr-border)' : 'none',
                  }}
                >
                  <span
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      height: 36,
                      width: 36,
                      borderRadius: '50%',
                      background: 'var(--accent-muted)',
                      color: ACCENT,
                      fontSize: 12,
                      fontWeight: 700,
                      flexShrink: 0,
                    }}
                  >
                    {getInitials(s.name)}
                  </span>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p
                      style={{ margin: 0, fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}
                    >
                      {s.name}
                    </p>
                    <p style={{ margin: '2px 0 0', fontSize: 11, color: 'var(--hr-text-muted)' }}>
                      {s.role}
                    </p>
                  </div>
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 6,
                      color: 'var(--hr-text-muted)',
                      fontSize: 13,
                    }}
                  >
                    <Clock size={13} style={{ color: 'var(--hr-text-dim)' }} />
                    {s.start} – {s.end}
                  </div>
                  <Pill color={cfg.color} style={{ background: cfg.bg }}>
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                      <StatusIcon size={11} />
                      {t(
                        `shifts.statuses.${s.status === 'In Progress' ? 'inProgress' : s.status.toLowerCase()}`
                      )}
                    </span>
                  </Pill>
                </motion.div>
              );
            })
          )}
        </motion.div>
      </div>
    </>
  );
};

export default Shifts;
