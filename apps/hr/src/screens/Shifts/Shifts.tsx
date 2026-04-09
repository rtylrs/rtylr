import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { CalendarPlus, Clock, CheckCircle2, Play, Calendar, ChevronLeft, ChevronRight } from 'lucide-react';
import { hrApi, Shift } from '../../services/hrApi';

const ACCENT = '#8b5cf6';

const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const todayIdx = new Date().getDay() === 0 ? 6 : new Date().getDay() - 1;

const statusConfig: Record<Shift['status'], { color: string; bg: string; icon: typeof Clock }> = {
  Scheduled:   { color: ACCENT,    bg: 'rgba(139,92,246,0.12)', icon: Calendar   },
  'In Progress':{ color: '#10b981', bg: 'rgba(16,185,129,0.12)',  icon: Play       },
  Completed:   { color: '#6b7280', bg: 'rgba(107,114,128,0.12)', icon: CheckCircle2 },
};

function getInitials(name: string) {
  return name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2);
}

const container = { hidden: {}, show: { transition: { staggerChildren: 0.06 } } };
const rowAnim   = { hidden: { opacity: 0, x: -8 }, show: { opacity: 1, x: 0, transition: { duration: 0.2 } } };

const Shifts: React.FC = () => {
  const [selectedDay, setSelectedDay] = useState(todayIdx);
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    hrApi.listShifts()
      .then((res) => { if (res.data?.length) setShifts(res.data); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const dayShifts = shifts.filter((s) => s.day === days[selectedDay]);
  const todayShifts = shifts.filter((s) => s.day === days[todayIdx]);
  const scheduled = todayShifts.filter((s) => s.status === 'Scheduled').length;
  const inProgress = todayShifts.filter((s) => s.status === 'In Progress').length;
  const completed = todayShifts.filter((s) => s.status === 'Completed').length;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: '18px', fontWeight: 700, color: 'var(--hr-text)' }}>Shifts</h2>
          <p style={{ margin: '4px 0 0', fontSize: '13px', color: 'var(--hr-text-muted)' }}>Week of April 7 – 13, 2026</p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <button style={{ padding: '8px', borderRadius: '8px', background: 'var(--hr-surface)', border: '1px solid var(--hr-border)', color: 'var(--hr-text-muted)', cursor: 'pointer' }}>
            <ChevronLeft size={14} />
          </button>
          <button style={{ padding: '8px', borderRadius: '8px', background: 'var(--hr-surface)', border: '1px solid var(--hr-border)', color: 'var(--hr-text-muted)', cursor: 'pointer' }}>
            <ChevronRight size={14} />
          </button>
          <button
            style={{
              display: 'flex', alignItems: 'center', gap: '8px',
              padding: '9px 16px', borderRadius: '8px',
              background: ACCENT, border: 'none', color: '#fff',
              fontSize: '13px', fontWeight: 600, cursor: 'pointer',
            }}
          >
            <CalendarPlus size={15} />
            Add Shift
          </button>
        </div>
      </div>

      {/* Stats row */}
      {(() => {
        const statItems = [
          { label: 'Scheduled',    value: scheduled,  color: ACCENT,     bg: 'rgba(139,92,246,0.12)' },
          { label: 'In Progress',  value: inProgress, color: '#10b981',  bg: 'rgba(16,185,129,0.12)' },
          { label: 'Completed',    value: completed,  color: '#6b7280',  bg: 'rgba(107,114,128,0.12)' },
        ];
        return (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
            {statItems.map((s) => (
              <div key={s.label} style={{ background: s.bg, borderRadius: '10px', padding: '14px 18px', border: `1px solid ${s.color}30` }}>
                <p style={{ margin: 0, fontSize: '24px', fontWeight: 700, color: s.color }}>{s.value}</p>
                <p style={{ margin: '4px 0 0', fontSize: '12px', color: s.color, fontWeight: 500 }}>{s.label} Today</p>
              </div>
            ))}
          </div>
        );
      })()}

      {/* Week selector */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: '6px' }}>
        {days.map((day, i) => {
          const count = shifts.filter((s) => s.day === day).length;
          const isToday = i === todayIdx;
          const isSelected = i === selectedDay;
          return (
            <button
              key={day}
              onClick={() => setSelectedDay(i)}
              style={{
                padding: '10px 6px',
                borderRadius: '10px',
                border: isSelected ? `1px solid ${ACCENT}` : '1px solid var(--hr-border)',
                background: isSelected ? 'rgba(139,92,246,0.12)' : 'var(--hr-card)',
                cursor: 'pointer',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: '4px',
              }}
            >
              <span style={{ fontSize: '11px', fontWeight: 600, color: isToday ? ACCENT : 'var(--hr-text-muted)', textTransform: 'uppercase' }}>
                {day}
              </span>
              <span style={{ fontSize: '14px', fontWeight: 700, color: isSelected ? ACCENT : 'var(--hr-text)' }}>
                {7 + i}
              </span>
              {count > 0 && (
                <span style={{ width: 6, height: 6, borderRadius: '50%', background: isSelected ? ACCENT : 'var(--hr-text-dim)' }} />
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
        key={selectedDay}
        style={{ background: 'var(--hr-card)', border: '1px solid var(--hr-border)', borderRadius: '12px', overflow: 'hidden' }}
      >
        <div style={{ padding: '14px 20px', borderBottom: '1px solid var(--hr-border)' }}>
          <h3 style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)' }}>
            {days[selectedDay]} — {dayShifts.length} shift{dayShifts.length !== 1 ? 's' : ''}
          </h3>
        </div>

        {dayShifts.length === 0 ? (
          <div style={{ padding: '40px', textAlign: 'center', color: 'var(--hr-text-muted)', fontSize: '13px' }}>
            No shifts scheduled for this day.
          </div>
        ) : (
          dayShifts.map((s, i) => {
            const cfg = statusConfig[s.status];
            const StatusIcon = cfg.icon;
            return (
              <motion.div
                key={s.id}
                variants={rowAnim}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '14px',
                  padding: '14px 20px',
                  borderBottom: i < dayShifts.length - 1 ? '1px solid var(--hr-border)' : 'none',
                }}
              >
                {/* Avatar */}
                <span
                  style={{
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    height: 36, width: 36, borderRadius: '50%',
                    background: 'rgba(139,92,246,0.12)', color: ACCENT,
                    fontSize: '12px', fontWeight: 700, flexShrink: 0,
                  }}
                >
                  {getInitials(s.name)}
                </span>

                <div style={{ flex: 1, minWidth: 0 }}>
                  <p style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--hr-text)' }}>{s.name}</p>
                  <p style={{ margin: '2px 0 0', fontSize: '11px', color: 'var(--hr-text-muted)' }}>{s.role}</p>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '6px', color: 'var(--hr-text-muted)', fontSize: '13px' }}>
                  <Clock size={13} style={{ color: 'var(--hr-text-dim)' }} />
                  {s.start} – {s.end}
                </div>

                <span
                  style={{
                    display: 'flex', alignItems: 'center', gap: '5px',
                    padding: '4px 10px', borderRadius: '20px',
                    fontSize: '11px', fontWeight: 600,
                    background: cfg.bg, color: cfg.color,
                    flexShrink: 0,
                  }}
                >
                  <StatusIcon size={11} />
                  {s.status}
                </span>
              </motion.div>
            );
          })
        )}
      </motion.div>
    </div>
  );
};

export default Shifts;
