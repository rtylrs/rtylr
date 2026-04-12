import React, { useState, useMemo } from 'react';
import { motion } from 'motion/react';
import { ChevronLeft, ChevronRight, Calendar as CalendarIcon } from 'lucide-react';
import { useCalendarEvents } from './useCalendar';

const INDIGO = '#6366f1';

type ViewMode = 'day' | 'week' | 'month';

const HOURS = Array.from({ length: 14 }, (_, i) => i + 7); // 7am to 8pm

function formatDateISO(d: Date): string {
  return d.toISOString().split('T')[0];
}

function addDays(d: Date, n: number): Date {
  const r = new Date(d);
  r.setDate(r.getDate() + n);
  return r;
}

function startOfWeek(d: Date): Date {
  const r = new Date(d);
  const day = r.getDay();
  r.setDate(r.getDate() - day);
  return r;
}

function getDaysInMonth(y: number, m: number): number {
  return new Date(y, m + 1, 0).getDate();
}

const DAY_NAMES = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

function getStatusColor(status: string): string {
  if (status === 'confirmed') return '#10b981';
  if (status === 'pending') return '#f59e0b';
  if (status === 'cancelled') return '#ef4444';
  return INDIGO;
}

const SkeletonBlock: React.FC = () => (
  <div className="h-64 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
);

const CalendarScreen: React.FC = () => {
  const [view, setView] = useState<ViewMode>('week');
  const [currentDate, setCurrentDate] = useState(new Date());

  const dateRange = useMemo(() => {
    if (view === 'day') {
      return { start: formatDateISO(currentDate), end: formatDateISO(currentDate) };
    }
    if (view === 'week') {
      const ws = startOfWeek(currentDate);
      return { start: formatDateISO(ws), end: formatDateISO(addDays(ws, 6)) };
    }
    const y = currentDate.getFullYear();
    const m = currentDate.getMonth();
    return {
      start: formatDateISO(new Date(y, m, 1)),
      end: formatDateISO(new Date(y, m, getDaysInMonth(y, m))),
    };
  }, [view, currentDate]);

  const { data: events, isLoading, isError } = useCalendarEvents(dateRange);

  const navigate = (dir: number) => {
    const d = new Date(currentDate);
    if (view === 'day') d.setDate(d.getDate() + dir);
    else if (view === 'week') d.setDate(d.getDate() + dir * 7);
    else d.setMonth(d.getMonth() + dir);
    setCurrentDate(d);
  };

  const headerLabel = useMemo(() => {
    if (view === 'day') {
      return currentDate.toLocaleDateString('en-US', {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
        year: 'numeric',
      });
    }
    if (view === 'week') {
      const ws = startOfWeek(currentDate);
      const we = addDays(ws, 6);
      return `${ws.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
      })} - ${we.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}`;
    }
    return currentDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
  }, [view, currentDate]);

  const weekDays = useMemo(() => {
    const ws = startOfWeek(currentDate);
    return Array.from({ length: 7 }, (_, i) => addDays(ws, i));
  }, [currentDate]);

  const monthDays = useMemo(() => {
    const y = currentDate.getFullYear();
    const m = currentDate.getMonth();
    const firstDay = new Date(y, m, 1).getDay();
    const daysInMonth = getDaysInMonth(y, m);
    const days: (number | null)[] = [];
    for (let i = 0; i < firstDay; i++) days.push(null);
    for (let i = 1; i <= daysInMonth; i++) days.push(i);
    return days;
  }, [currentDate]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-4 max-w-7xl"
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate(-1)}
            className="h-7 w-7 rounded-lg flex items-center justify-center"
            style={{ color: 'var(--hub-text-muted)', background: 'var(--hub-hover)' }}
          >
            <ChevronLeft size={14} />
          </button>
          <h2
            className="text-sm font-semibold min-w-[220px] text-center"
            style={{ color: 'var(--hub-text)' }}
          >
            {headerLabel}
          </h2>
          <button
            onClick={() => navigate(1)}
            className="h-7 w-7 rounded-lg flex items-center justify-center"
            style={{ color: 'var(--hub-text-muted)', background: 'var(--hub-hover)' }}
          >
            <ChevronRight size={14} />
          </button>
          <button
            onClick={() => setCurrentDate(new Date())}
            className="px-2 py-1 rounded-lg text-xs font-medium"
            style={{ color: INDIGO, background: 'rgba(99,102,241,0.1)' }}
          >
            Today
          </button>
        </div>
        <div className="flex gap-1">
          {(['day', 'week', 'month'] as ViewMode[]).map((v) => (
            <button
              key={v}
              onClick={() => setView(v)}
              className="px-3 py-1 rounded-lg text-xs font-medium capitalize"
              style={{
                background: view === v ? 'rgba(99,102,241,0.12)' : 'var(--hub-hover)',
                color: view === v ? INDIGO : 'var(--hub-text-muted)',
                border: view === v ? '1px solid rgba(99,102,241,0.3)' : '1px solid transparent',
              }}
            >
              {v}
            </button>
          ))}
        </div>
      </div>

      {isError && (
        <div
          className="rounded-xl p-4 text-sm"
          style={{
            background: 'rgba(239,68,68,0.08)',
            color: '#ef4444',
            border: '1px solid rgba(239,68,68,0.2)',
          }}
        >
          Failed to load calendar events. Please try again later.
        </div>
      )}

      {isLoading && <SkeletonBlock />}

      {!isLoading && !isError && view === 'month' && (
        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="grid grid-cols-7">
            {DAY_NAMES.map((d) => (
              <div
                key={d}
                className="px-2 py-2 text-center text-[10px] font-semibold uppercase"
                style={{
                  color: 'var(--hub-text-dim)',
                  borderBottom: '1px solid var(--hub-border)',
                }}
              >
                {d}
              </div>
            ))}
          </div>
          <div className="grid grid-cols-7">
            {monthDays.map((day, i) => {
              const isToday =
                day != null &&
                formatDateISO(new Date(currentDate.getFullYear(), currentDate.getMonth(), day)) ===
                  formatDateISO(new Date());
              const dateStr =
                day != null
                  ? formatDateISO(new Date(currentDate.getFullYear(), currentDate.getMonth(), day))
                  : '';
              const dayEvents = (events ?? []).filter((e) => e.date === dateStr);
              return (
                <div
                  key={i}
                  className="min-h-[80px] p-1"
                  style={{
                    borderBottom: '1px solid var(--hub-border)',
                    borderRight: (i + 1) % 7 !== 0 ? '1px solid var(--hub-border)' : 'none',
                  }}
                >
                  {day != null && (
                    <>
                      <span
                        className="inline-flex items-center justify-center h-5 w-5 rounded-full text-[10px] font-medium"
                        style={{
                          background: isToday ? INDIGO : 'transparent',
                          color: isToday ? '#fff' : 'var(--hub-text-muted)',
                        }}
                      >
                        {day}
                      </span>
                      {dayEvents.slice(0, 3).map((ev, j: number) => (
                        <div
                          key={j}
                          className="mt-0.5 px-1 py-0.5 rounded text-[9px] truncate"
                          style={{
                            background: `${getStatusColor(ev.status)}18`,
                            color: getStatusColor(ev.status),
                          }}
                        >
                          {ev.time} {ev.customerName}
                        </div>
                      ))}
                      {dayEvents.length > 3 && (
                        <p className="text-[9px] px-1" style={{ color: 'var(--hub-text-dim)' }}>
                          +{dayEvents.length - 3} more
                        </p>
                      )}
                    </>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {!isLoading && !isError && view === 'week' && (
        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="grid grid-cols-[60px_repeat(7,1fr)]">
            <div style={{ borderBottom: '1px solid var(--hub-border)' }} />
            {weekDays.map((d, i) => {
              const isToday = formatDateISO(d) === formatDateISO(new Date());
              return (
                <div
                  key={i}
                  className="px-2 py-2 text-center"
                  style={{
                    borderBottom: '1px solid var(--hub-border)',
                    borderLeft: '1px solid var(--hub-border)',
                  }}
                >
                  <p
                    className="text-[10px] font-semibold uppercase"
                    style={{ color: 'var(--hub-text-dim)' }}
                  >
                    {DAY_NAMES[d.getDay()]}
                  </p>
                  <p
                    className="text-sm font-semibold mt-0.5"
                    style={{ color: isToday ? INDIGO : 'var(--hub-text)' }}
                  >
                    {d.getDate()}
                  </p>
                </div>
              );
            })}
          </div>
          <div className="max-h-[500px] overflow-y-auto">
            {HOURS.map((hour) => (
              <div key={hour} className="grid grid-cols-[60px_repeat(7,1fr)] min-h-[48px]">
                <div
                  className="px-2 py-1 text-[10px] font-medium text-right pr-3"
                  style={{
                    color: 'var(--hub-text-dim)',
                    borderBottom: '1px solid var(--hub-border)',
                  }}
                >
                  {hour > 12 ? `${hour - 12} PM` : hour === 12 ? '12 PM' : `${hour} AM`}
                </div>
                {weekDays.map((d, i) => {
                  const dateStr = formatDateISO(d);
                  const hourStr = String(hour).padStart(2, '0');
                  const hourEvents = (events ?? []).filter(
                    (e) => e.date === dateStr && e.time?.startsWith(hourStr)
                  );
                  return (
                    <div
                      key={i}
                      className="relative p-0.5"
                      style={{
                        borderBottom: '1px solid var(--hub-border)',
                        borderLeft: '1px solid var(--hub-border)',
                      }}
                    >
                      {hourEvents.map((ev, j: number) => (
                        <div
                          key={j}
                          className="px-1 py-0.5 rounded text-[9px] truncate mb-0.5"
                          style={{
                            background: `${getStatusColor(ev.status)}18`,
                            color: getStatusColor(ev.status),
                          }}
                        >
                          {ev.customerName} - {ev.serviceName}
                        </div>
                      ))}
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      )}

      {!isLoading && !isError && view === 'day' && (
        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="max-h-[600px] overflow-y-auto">
            {HOURS.map((hour) => {
              const dateStr = formatDateISO(currentDate);
              const hourStr = String(hour).padStart(2, '0');
              const hourEvents = (events ?? []).filter(
                (e) => e.date === dateStr && e.time?.startsWith(hourStr)
              );
              return (
                <div
                  key={hour}
                  className="grid grid-cols-[80px_1fr] min-h-[56px]"
                  style={{ borderBottom: '1px solid var(--hub-border)' }}
                >
                  <div
                    className="px-3 py-2 text-xs font-medium text-right pr-4"
                    style={{ color: 'var(--hub-text-dim)' }}
                  >
                    {hour > 12 ? `${hour - 12}:00 PM` : hour === 12 ? '12:00 PM' : `${hour}:00 AM`}
                  </div>
                  <div className="p-1" style={{ borderLeft: '1px solid var(--hub-border)' }}>
                    {hourEvents.map((ev, j: number) => (
                      <div
                        key={j}
                        className="px-2 py-1.5 rounded-lg text-xs mb-1"
                        style={{
                          background: `${getStatusColor(ev.status)}18`,
                          borderLeft: `3px solid ${getStatusColor(ev.status)}`,
                        }}
                      >
                        <p className="font-medium" style={{ color: 'var(--hub-text)' }}>
                          {ev.customerName}
                        </p>
                        <p style={{ color: 'var(--hub-text-muted)' }}>
                          {ev.serviceName} - {ev.duration} min
                        </p>
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {!isLoading && !isError && (!events || events.length === 0) && (
        <div className="py-12 text-center">
          <CalendarIcon
            size={36}
            className="mx-auto mb-3"
            style={{ color: 'var(--hub-text-dim)' }}
          />
          <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>
            No events in this period
          </p>
          <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
            Schedule appointments to see them here
          </p>
        </div>
      )}
    </motion.div>
  );
};

export default CalendarScreen;
