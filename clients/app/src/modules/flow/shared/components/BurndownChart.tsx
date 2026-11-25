import React from 'react';
import { useTranslation } from 'react-i18next';
import { formatShortDate } from '../lib/utils';

const WIDTH = 640;
const HEIGHT = 200;
const PAD = { top: 12, right: 12, bottom: 24, left: 32 };

export interface BurndownChartPoint {
  /** yyyy-mm-dd; the first and last label the x axis when present. */
  date?: string;
  /** Work left at the end of the day, or null for days still to come. */
  remaining: number | null;
  ideal: number;
}

interface BurndownChartProps {
  points: BurndownChartPoint[];
  /** Label for the y axis unit, e.g. "tasks" or "points". */
  unitLabel: string;
}

/**
 * Remaining work per day against the ideal straight line, for a cycle page and
 * Insights. Plain SVG, so Flow does not load a chart library for one line.
 */
const BurndownChart: React.FC<BurndownChartProps> = ({ points, unitLabel }) => {
  const { t, i18n } = useTranslation();
  if (points.length < 2) {
    return (
      <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.cycles.detail.burndownEmpty')}
      </p>
    );
  }
  const max = Math.max(1, ...points.map((p) => Math.max(p.ideal, p.remaining ?? 0)));
  const innerW = WIDTH - PAD.left - PAD.right;
  const innerH = HEIGHT - PAD.top - PAD.bottom;
  const x = (i: number) => PAD.left + (i / (points.length - 1)) * innerW;
  const y = (v: number) => PAD.top + innerH - (v / max) * innerH;
  const ideal = points.map((p, i) => `${x(i)},${y(p.ideal)}`).join(' ');
  const actual = points
    .map((p, i) => (p.remaining === null ? null : `${x(i)},${y(p.remaining)}`))
    .filter((v): v is string => v !== null)
    .join(' ');
  const last = points.length - 1;
  const summary = t('flow.cycles.detail.burndownSummary', {
    remaining: [...points].reverse().find((p) => p.remaining !== null)?.remaining ?? max,
    unit: unitLabel,
  });

  return (
    <figure className="w-full">
      <svg
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        className="w-full h-auto"
        role="img"
        aria-label={summary}
        // Time runs left to right in both directions of the page.
        style={{ direction: 'ltr' }}
      >
        <line
          x1={PAD.left}
          x2={WIDTH - PAD.right}
          y1={y(0)}
          y2={y(0)}
          stroke="var(--flow-border)"
        />
        <line x1={PAD.left} x2={PAD.left} y1={PAD.top} y2={y(0)} stroke="var(--flow-border)" />
        <text
          x={PAD.left - 6}
          y={y(max) + 4}
          textAnchor="end"
          fontSize="10"
          fill="var(--flow-text-dim)"
        >
          {max}
        </text>
        <text
          x={PAD.left - 6}
          y={y(0) + 4}
          textAnchor="end"
          fontSize="10"
          fill="var(--flow-text-dim)"
        >
          0
        </text>
        {points[0].date && (
          <text x={x(0)} y={HEIGHT - 6} fontSize="10" fill="var(--flow-text-dim)">
            {formatShortDate(points[0].date, i18n.language)}
          </text>
        )}
        {points[last].date && (
          <text
            x={x(last)}
            y={HEIGHT - 6}
            textAnchor="end"
            fontSize="10"
            fill="var(--flow-text-dim)"
          >
            {formatShortDate(points[last].date, i18n.language)}
          </text>
        )}
        <polyline
          points={ideal}
          fill="none"
          stroke="var(--flow-text-dim)"
          strokeDasharray="4 4"
          strokeWidth={1.5}
        />
        {actual && (
          <polyline points={actual} fill="none" stroke="var(--accent)" strokeWidth={2.5} />
        )}
      </svg>
      <figcaption
        className="mt-2 flex items-center gap-4 text-flow-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <span className="flex items-center gap-1.5">
          <span className="h-0.5 w-4" style={{ background: 'var(--accent)' }} />
          {t('flow.cycles.detail.remaining')}
        </span>
        <span className="flex items-center gap-1.5">
          <span
            className="h-0 w-4 border-t border-dashed"
            style={{ borderColor: 'var(--flow-text-dim)' }}
          />
          {t('flow.cycles.detail.ideal')}
        </span>
      </figcaption>
    </figure>
  );
};

export default BurndownChart;
