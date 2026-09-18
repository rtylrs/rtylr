import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';

import { findLocale } from '../../../../lib/locales';

interface Staff {
  initials: string;
  name: string;
  roleKey: 'floorManager' | 'barista' | 'chef' | 'server';
  start: number;
  end: number;
  color: string;
  clockedIn?: boolean;
}

const STAFF: Staff[] = [
  {
    initials: 'MC',
    name: 'Maya Chen',
    roleKey: 'floorManager',
    start: 8,
    end: 16,
    color: '#b3c5e8',
    clockedIn: true,
  },
  {
    initials: 'SP',
    name: 'Sam Park',
    roleKey: 'barista',
    start: 9,
    end: 17,
    color: '#a2d9af',
    clockedIn: true,
  },
  {
    initials: 'JK',
    name: 'James K.',
    roleKey: 'chef',
    start: 10,
    end: 22,
    color: '#f3c08a',
    clockedIn: true,
  },
  {
    initials: 'AR',
    name: 'Aisha R.',
    roleKey: 'server',
    start: 11,
    end: 19,
    color: '#c9a48a',
    clockedIn: false,
  },
  {
    initials: 'LB',
    name: 'Luca B.',
    roleKey: 'barista',
    start: 14,
    end: 22,
    color: '#d6b88e',
    clockedIn: false,
  },
  {
    initials: 'NA',
    name: 'Nadia A.',
    roleKey: 'server',
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
  const { t, i18n } = useTranslation();
  const [now, setNow] = useState(10.5);
  // transform-origin has no widely-supported logical keyword yet, so the
  // scaleX growth anchor for each shift bar is set explicitly from the
  // active locale's direction rather than a CSS property alone.
  const isRtl = findLocale(i18n.language)?.dir === 'rtl';

  useEffect(() => {
    if (!isActive) return;
    const timer = setInterval(() => {
      setNow((h) => {
        const next = h + 0.15;
        return next > 22 ? 8.5 : next;
      });
    }, 120);
    return () => clearInterval(timer);
  }, [isActive]);

  const nowPct = ((now - DAY_START) / TOTAL_HRS) * 100;
  const clocked = STAFF.filter((s) => now >= s.start && now < s.end).length;

  // 24-hour, not 12-hour with an am/pm suffix: am/pm has no equivalent
  // worth inventing in French or Arabic, and 24-hour is the locale-neutral
  // convention anyway, so this sidesteps translating it at all.
  const fmtHour = (h: number): string => {
    const hh = Math.floor(h);
    const mm = Math.round((h - hh) * 60);
    return mm === 0 ? `${hh}:00` : `${hh}:${String(mm).padStart(2, '0')}`;
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
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>
            {t('home.features.demos.hr.scheduleTitle')}
          </span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12, marginInlineStart: 10 }}>
            {t('home.features.demos.hr.today')}
          </span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <div style={{ width: 7, height: 7, borderRadius: '50%', background: '#22c55e' }} />
          <span style={{ color: 'rgba(19,19,19,0.65)', fontSize: 12 }}>
            {t('home.features.demos.hr.onShift', { count: clocked })}
          </span>
        </div>
      </div>

      {/* Timeline header */}
      <div
        style={{
          paddingBlock: '8px 6px',
          paddingInlineStart: 148,
          paddingInlineEnd: 20,
          display: 'flex',
          justifyContent: 'space-between',
          flexShrink: 0,
        }}
      >
        {[8, 10, 12, 14, 16, 18, 20, 22].map((h) => (
          <span
            key={h}
            dir="ltr"
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
            insetInlineStart: `calc(148px + ${nowPct}% * ((100% - 148px) / 100))`,
            width: 1,
            background: 'var(--ink)',
            zIndex: 10,
            pointerEvents: 'none',
          }}
        >
          <div
            dir="ltr"
            style={{
              position: 'absolute',
              top: 0,
              insetInlineStart: -18,
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
                paddingBlock: 7,
                paddingInlineEnd: 20,
                borderBottom: '1px solid var(--hairline)',
              }}
            >
              {/* Staff info */}
              <div
                style={{
                  width: 148,
                  paddingInlineStart: 20,
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
                  {/* dir="ltr": a Latin name ending in a period (an
                      abbreviated surname) puts the period at the visual
                      start under dir="rtl" otherwise, e.g. ".James K". */}
                  <div
                    dir="ltr"
                    style={{
                      fontSize: 11,
                      fontWeight: 500,
                      color: isOn ? 'var(--ink)' : 'rgba(19,19,19,0.5)',
                    }}
                  >
                    {s.name}
                  </div>
                  <div style={{ fontSize: 10, color: 'rgba(19,19,19,0.45)' }}>
                    {t(`home.features.demos.hr.roles.${s.roleKey}`)}
                  </div>
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
                    insetInlineStart: `${leftPct}%`,
                    width: `${widthPct}%`,
                    height: '100%',
                    background: isOn ? `${s.color}33` : 'var(--hairline)',
                    borderRadius: 4,
                    border: `1px solid ${isOn ? `${s.color}66` : 'var(--hairline-strong)'}`,
                    transformOrigin: isRtl ? 'right' : 'left',
                    display: 'flex',
                    alignItems: 'center',
                    paddingInlineStart: 6,
                    overflow: 'hidden',
                  }}
                >
                  <span
                    dir="ltr"
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
