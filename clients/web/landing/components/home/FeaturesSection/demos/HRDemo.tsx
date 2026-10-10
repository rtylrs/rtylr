import React from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../../../lib/cn';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { useCountUp } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './HRDemo.module.scss';

type Role = 'ops' | 'sales' | 'support' | 'finance' | 'design';

interface Shift {
  name: string;
  initials: string;
  role: Role;
  start: number;
  end: number;
  color: string;
}

const TEAM: Shift[] = [
  { name: 'Sara Haddad', initials: 'SH', role: 'ops', start: 8, end: 16, color: '#10b981' },
  { name: 'Marc Khoury', initials: 'MK', role: 'sales', start: 9, end: 17, color: '#3b82f6' },
  { name: 'Lina Rahme', initials: 'LR', role: 'support', start: 10.5, end: 18.5, color: '#8b5cf6' },
  { name: 'Jad Daou', initials: 'JD', role: 'finance', start: 12, end: 20, color: '#f59e0b' },
  { name: 'Nour Saad', initials: 'NS', role: 'design', start: 9, end: 13, color: '#ef4444' },
];

const DAY_START = 8;
const DAY_END = 20;
const SPAN = DAY_END - DAY_START;
const TICKS = [8, 10, 12, 14, 16, 18, 20];
// "Now" walks from 9:45 to 11:15 while the tab is open, so Lina's shift
// visibly starts mid-demo.
const NOW_FROM = 9.75;
const NOW_TRAVEL = 1.5;

const pct = (h: number): string => `${((h - DAY_START) / SPAN) * 100}%`;
const fmt = (h: number): string => {
  const hh = Math.floor(h);
  const mm = Math.floor((h - hh) * 60);
  return `${hh}:${String(mm).padStart(2, '0')}`;
};

export const HRDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const now = NOW_FROM + useCountUp(NOW_TRAVEL, isActive, 5200);
  const onShift = TEAM.filter((s) => now >= s.start && now < s.end).length;

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="hr"
      crumb={t('home.features.demos.hr.crumb', { returnObjects: true }) as string[]}
    >
      <div className={styles.spread}>
        <span className={styles.title}>{t('home.features.demos.hr.today')}</span>
        <span className={styles.pill} data-tone="green">
          <span className={local.liveDot} />
          {t('home.features.demos.hr.onShift', { count: onShift })}
        </span>
      </div>

      <div className={local.grid} dir="ltr">
        <div />
        <div className={local.ticks} dir="ltr">
          {TICKS.map((h) => (
            <span key={h} style={{ left: pct(h) }}>
              {h}:00
            </span>
          ))}
        </div>

        {TEAM.map((s) => {
          const on = now >= s.start && now < s.end;
          return (
            <React.Fragment key={s.name}>
              <div className={local.person}>
                <span className={styles.avatar} style={{ background: s.color }}>
                  {s.initials}
                </span>
                <div>
                  <div className={local.name}>{s.name}</div>
                  <div className={styles.muted}>{t(`home.features.demos.hr.roles.${s.role}`)}</div>
                </div>
              </div>
              <div className={local.track} dir="ltr">
                <span
                  className={cn(local.shift, on && local.shiftOn)}
                  style={
                    {
                      left: pct(s.start),
                      width: `${((s.end - s.start) / SPAN) * 100}%`,
                      '--shift': s.color,
                    } as React.CSSProperties
                  }
                >
                  {fmt(s.start)}–{fmt(s.end)}
                </span>
              </div>
            </React.Fragment>
          );
        })}

        {/* Now line spans the timeline column only. */}
        <div className={local.nowLayer} dir="ltr" aria-hidden="true">
          <span className={local.now} style={{ left: pct(now) }}>
            <span className={local.nowLabel}>{fmt(now)}</span>
          </span>
        </div>
      </div>
    </PortalFrame>
  );
};
