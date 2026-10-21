import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { SalesActivityDay } from '@voxire/voxie/crm/sales/v1/types';

/** A round top for an axis reaching `max`: 1, 2, 4 or 5 times a power of ten. */
export function niceCeiling(max: number): number {
  if (max <= 4) return 4;
  const power = 10 ** Math.floor(Math.log10(max));
  return ([1, 2, 4, 5, 10].find((step) => step * power >= max) ?? 10) * power;
}

/** Which days get an x-axis label: every day of a week, else four spread out. */
export function labelledDays(count: number): number[] {
  if (count <= 7) return Array.from({ length: count }, (_, i) => i);
  return [0, Math.round((count - 1) / 3), Math.round(((count - 1) * 2) / 3), count - 1];
}

const dayOf = (iso: string) => new Date(`${iso}T00:00:00Z`);

/**
 * Calls per day as columns: the calls that reached someone at the base in the
 * accent, the rest stacked above in gray. Hovering a column, or moving
 * through the days with the arrow keys once the chart has focus, shows that
 * day's numbers; a visually hidden table carries every value.
 */
export const ActivityChart: React.FC<{ days: SalesActivityDay[] }> = ({ days }) => {
  const { t, i18n } = useTranslation();
  const [active, setActive] = useState<number | null>(null);
  const rtl = i18n.dir() === 'rtl';
  const number = new Intl.NumberFormat(i18n.language);
  const short = new Intl.DateTimeFormat(i18n.language, {
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  });
  const long = new Intl.DateTimeFormat(i18n.language, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  });

  const count = days.length;
  const top = niceCeiling(Math.max(0, ...days.map((day) => day.calls ?? 0)));
  const ticks = Number.isInteger(top / 2) ? [0, top / 2, top] : [0, top];
  const gap = count > 60 ? 2 : count > 20 ? 4 : 12;
  const labels = labelledDays(count);

  const move = (event: React.KeyboardEvent) => {
    const later = rtl ? 'ArrowLeft' : 'ArrowRight';
    const earlier = rtl ? 'ArrowRight' : 'ArrowLeft';
    const current = active ?? count - 1;
    const next =
      event.key === later
        ? Math.min(count - 1, current + 1)
        : event.key === earlier
          ? Math.max(0, current - 1)
          : event.key === 'Home'
            ? 0
            : event.key === 'End'
              ? count - 1
              : null;
    if (next === null) return;
    event.preventDefault();
    setActive(next);
  };

  const shown = active === null ? undefined : days[active];
  const centre = active === null ? 0 : (active + 0.5) / count;
  // Near either edge the tooltip hangs inward instead of centring on its column.
  const tooltipPlace: React.CSSProperties =
    centre < 0.2
      ? { insetInlineStart: `${(active ?? 0) * (100 / count)}%` }
      : centre > 0.8
        ? { insetInlineEnd: `${(count - (active ?? 0) - 1) * (100 / count)}%` }
        : {
            insetInlineStart: `${centre * 100}%`,
            transform: `translateX(${rtl ? '50%' : '-50%'})`,
          };

  return (
    <>
      <ul className="crm-dash-legend">
        <li>
          <span className="crm-dash-swatch" style={{ background: 'var(--crm-accent)' }} />
          {t('crm.dashboard.calls.reached')}
        </li>
        <li>
          <span className="crm-dash-swatch" style={{ background: 'var(--crm-chart-muted)' }} />
          {t('crm.dashboard.calls.notReached')}
        </li>
      </ul>
      <div
        className="crm-dash-chart"
        role="group"
        tabIndex={0}
        aria-label={t('crm.dashboard.calls.chartLabel')}
        onKeyDown={move}
        onFocus={() => setActive((current) => current ?? count - 1)}
        onBlur={() => setActive(null)}
        onPointerLeave={() => setActive(null)}
      >
        <div className="crm-dash-yaxis" aria-hidden="true">
          {ticks.map((tick) => (
            <span key={tick} style={{ bottom: `${(tick / top) * 100}%` }}>
              {number.format(tick)}
            </span>
          ))}
        </div>
        <div className="crm-dash-plot">
          {ticks.map((tick) => (
            <div
              key={tick}
              className="crm-dash-gridline"
              style={{ bottom: `${(tick / top) * 100}%` }}
              aria-hidden="true"
            />
          ))}
          <div
            className="crm-dash-columns"
            style={{ gridTemplateColumns: `repeat(${count}, minmax(0, 1fr))`, columnGap: gap }}
            aria-hidden="true"
          >
            {days.map((day, index) => {
              const calls = day.calls ?? 0;
              const reached = day.reached ?? 0;
              const missed = Math.max(0, calls - reached);
              return (
                <div
                  key={day.date}
                  className={`crm-dash-column${active === index ? ' is-active' : ''}`}
                  onPointerEnter={() => setActive(index)}
                >
                  {calls > 0 && (
                    <div
                      className="crm-dash-column-bar"
                      style={{ height: `${(calls / top) * 100}%` }}
                    >
                      {reached > 0 && (
                        <div
                          className={`crm-dash-segment is-reached${missed === 0 ? ' is-top' : ''}`}
                          style={{ flex: `${reached} 1 0`, minHeight: 2 }}
                        />
                      )}
                      {missed > 0 && (
                        <div
                          className="crm-dash-segment is-missed is-top"
                          style={{ flex: `${missed} 1 0`, minHeight: 2 }}
                        />
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
          {shown && (
            <div className="crm-dash-tooltip" style={tooltipPlace} role="status">
              <div className="crm-dash-tooltip-date">{long.format(dayOf(shown.date ?? ''))}</div>
              <div className="crm-dash-tooltip-row">
                <strong>{number.format(shown.calls ?? 0)}</strong>
                {t('crm.dashboard.calls.callsLabel', { count: shown.calls ?? 0 })}
              </div>
              <div className="crm-dash-tooltip-row">
                <span className="crm-dash-key" style={{ background: 'var(--crm-accent)' }} />
                <strong>{number.format(shown.reached ?? 0)}</strong>
                {t('crm.dashboard.calls.reached')}
              </div>
              <div className="crm-dash-tooltip-row">
                <span className="crm-dash-key" style={{ background: 'var(--crm-chart-muted)' }} />
                <strong>
                  {number.format(Math.max(0, (shown.calls ?? 0) - (shown.reached ?? 0)))}
                </strong>
                {t('crm.dashboard.calls.notReached')}
              </div>
            </div>
          )}
        </div>
        <div className="crm-dash-xaxis" aria-hidden="true">
          {labels.map((index, position) => (
            <span
              key={index}
              className={
                position === 0 && count > 7
                  ? 'is-first'
                  : position === labels.length - 1 && count > 7
                    ? 'is-last'
                    : undefined
              }
              style={{
                insetInlineStart:
                  count > 7 && position === 0
                    ? 0
                    : count > 7 && position === labels.length - 1
                      ? '100%'
                      : `${((index + 0.5) / count) * 100}%`,
              }}
            >
              {count <= 7
                ? new Intl.DateTimeFormat(i18n.language, {
                    weekday: 'short',
                    timeZone: 'UTC',
                  }).format(dayOf(days[index].date ?? ''))
                : short.format(dayOf(days[index].date ?? ''))}
            </span>
          ))}
        </div>
      </div>
      <table className="sales-visually-hidden">
        <caption>{t('crm.dashboard.calls.tableCaption')}</caption>
        <thead>
          <tr>
            <th scope="col">{t('crm.dashboard.calls.day')}</th>
            <th scope="col">{t('crm.dashboard.calls.title')}</th>
            <th scope="col">{t('crm.dashboard.calls.reached')}</th>
          </tr>
        </thead>
        <tbody>
          {days.map((day) => (
            <tr key={day.date}>
              <th scope="row">{long.format(dayOf(day.date ?? ''))}</th>
              <td>{number.format(day.calls ?? 0)}</td>
              <td>{number.format(day.reached ?? 0)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
};
