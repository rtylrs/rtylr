import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';

interface Staff {
  initials: string;
  name: string;
  role: string;
  start: number;
  end: number;
  color: string;
  clockedIn?: boolean;
}

const STAFF: Staff[] = [
  {
    initials: 'MC',
    name: 'Maya Chen',
    role: 'Floor Manager',
    start: 8,
    end: 16,
    color: '#b3c5e8',
    clockedIn: true,
  },
  {
    initials: 'SP',
    name: 'Sam Park',
    role: 'Barista',
    start: 9,
    end: 17,
    color: '#a2d9af',
    clockedIn: true,
  },
  {
    initials: 'JK',
    name: 'James K.',
    role: 'Chef',
    start: 10,
    end: 22,
    color: '#f3c08a',
    clockedIn: true,
  },
  {
    initials: 'AR',
    name: 'Aisha R.',
    role: 'Server',
    start: 11,
    end: 19,
    color: '#c9a48a',
    clockedIn: false,
  },
  {
    initials: 'LB',
    name: 'Luca B.',
    role: 'Barista',
    start: 14,
    end: 22,
    color: '#d6b88e',
    clockedIn: false,
  },
  {
    initials: 'NA',
    name: 'Nadia A.',
    role: 'Server',
    start: 16,
    end: 23,
    color: '#b3c5e8',
    clockedIn: false,
  },
];

const DAY_START = 7;
const DAY_END = 24;
const TOTAL_HRS = DAY_END - DAY_START;

export const HRDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const [now, setNow] = useState(10.5);

  useEffect(() => {
    if (!isActive) return;
    const t = setInterval(() => {
      setNow((h) => {
        const next = h + 0.15;
        return next > 22 ? 8.5 : next;
      });
    }, 120);
    return () => clearInterval(t);
  }, [isActive]);

  const nowPct = ((now - DAY_START) / TOTAL_HRS) * 100;
  const clocked = STAFF.filter((s) => now >= s.start && now < s.end).length;

  const fmtHour = (h: number): string => {
    const hh = Math.floor(h);
    const suffix = hh >= 12 ? 'pm' : 'am';
    return `${hh > 12 ? hh - 12 : hh}${suffix}`;
  };

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: 'var(--bg-surface-a)',
        fontFamily: 'inherit',
        fontSize: 13,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '16px 20px 14px',
          borderBottom: '1px solid var(--hairline)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}
      >
        <div>
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>Schedule</span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12, marginLeft: 10 }}>
            Today · Mon 5 May
          </span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <div style={{ width: 7, height: 7, borderRadius: '50%', background: '#22c55e' }} />
          <span style={{ color: 'rgba(19,19,19,0.65)', fontSize: 12 }}>{clocked} on shift</span>
        </div>
      </div>

      {/* Timeline header */}
      <div
        style={{
          padding: '8px 20px 6px 148px',
          display: 'flex',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}
      >
        {[8, 10, 12, 14, 16, 18, 20, 22].map((h) => (
          <span
            key={h}
            style={{ fontSize: 9, color: 'rgba(19,19,19,0.4)', width: 24, textAlign: 'center' }}
          >
            {fmtHour(h)}
          </span>
        ))}
      </div>

      {/* Staff rows */}
      <div style={{ flex: 1, overflow: 'hidden', position: 'relative' }}>
        {/* Now line */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            bottom: 0,
            left: `calc(148px + ${nowPct}% * ((100% - 148px) / 100))`,
            width: 1,
            background: 'var(--ink)',
            zIndex: 10,
            pointerEvents: 'none',
          }}
        >
          <div
            style={{
              position: 'absolute',
              top: 0,
              left: -18,
              background: 'var(--ink)',
              color: '#ffffff',
              borderRadius: 4,
              padding: '1px 5px',
              fontSize: 9,
              fontWeight: 700,
              whiteSpace: 'nowrap',
            }}
          >
            {fmtHour(now)}
          </div>
        </div>

        {STAFF.map((s) => {
          const isOn = now >= s.start && now < s.end;
          const leftPct = ((s.start - DAY_START) / TOTAL_HRS) * 100;
          const widthPct = ((s.end - s.start) / TOTAL_HRS) * 100;
          return (
            <div
              key={s.initials}
              style={{
                display: 'flex',
                alignItems: 'center',
                padding: '7px 20px 7px 0',
                borderBottom: '1px solid var(--hairline)',
              }}
            >
              {/* Staff info */}
              <div
                style={{
                  width: 148,
                  paddingLeft: 20,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  flexShrink: 0,
                }}
              >
                <div
                  style={{
                    width: 28,
                    height: 28,
                    borderRadius: '50%',
                    background: isOn ? s.color : 'var(--hairline)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 10,
                    fontWeight: 700,
                    color: isOn ? '#0a0a0a' : 'rgba(19,19,19,0.4)',
                    transition: 'all 0.3s ease',
                    flexShrink: 0,
                  }}
                >
                  {s.initials}
                </div>
                <div>
                  <div
                    style={{
                      fontSize: 11,
                      fontWeight: 500,
                      color: isOn ? 'var(--ink)' : 'rgba(19,19,19,0.5)',
                    }}
                  >
                    {s.name}
                  </div>
                  <div style={{ fontSize: 10, color: 'rgba(19,19,19,0.45)' }}>{s.role}</div>
                </div>
              </div>

              {/* Shift bar track */}
              <div style={{ flex: 1, position: 'relative', height: 20 }}>
                <motion.div
                  initial={{ scaleX: 0 }}
                  animate={{ scaleX: 1 }}
                  transition={{ duration: 0.6, delay: 0.1, ease: 'easeOut' }}
                  style={{
                    position: 'absolute',
                    left: `${leftPct}%`,
                    width: `${widthPct}%`,
                    height: '100%',
                    background: isOn ? `${s.color}33` : 'var(--hairline)',
                    borderRadius: 4,
                    border: `1px solid ${isOn ? `${s.color}66` : 'var(--hairline-strong)'}`,
                    transformOrigin: 'left',
                    display: 'flex',
                    alignItems: 'center',
                    paddingLeft: 6,
                    overflow: 'hidden',
                  }}
                >
                  <span
                    style={{
                      fontSize: 9,
                      color: isOn ? s.color : 'rgba(19,19,19,0.4)',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {fmtHour(s.start)}–{fmtHour(s.end)}
                  </span>
                </motion.div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
