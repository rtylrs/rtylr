import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const REVENUE = [9800, 11200, 10400, 12800, 18600, 21400, 14200];
const COVERS = [142, 163, 148, 184, 268, 309, 204];

const W = 580;
const H = 140;
const PAD = { t: 10, r: 10, b: 24, l: 40 };

const max = Math.max(...REVENUE);

const pts = REVENUE.map((v, i) => {
  const x = PAD.l + (i / (REVENUE.length - 1)) * (W - PAD.l - PAD.r);
  const y = PAD.t + (1 - v / max) * (H - PAD.t - PAD.b);
  return [x, y] as [number, number];
});

const linePath = pts
  .map((p, i) => `${i === 0 ? 'M' : 'L'}${p[0].toFixed(1)},${p[1].toFixed(1)}`)
  .join(' ');

const areaPath =
  linePath +
  ` L${pts[pts.length - 1][0].toFixed(1)},${(H - PAD.b).toFixed(1)} L${pts[0][0].toFixed(1)},${(H - PAD.b).toFixed(1)} Z`;

const total = REVENUE.reduce((s, v) => s + v, 0);
const avgOrder = Math.round(total / COVERS.reduce((s, v) => s + v, 0));
const peakDay = DAYS[REVENUE.indexOf(Math.max(...REVENUE))];
const totalCovers = COVERS.reduce((s, v) => s + v, 0);

const KPIs = [
  { label: 'Revenue', value: `$${(total / 1000).toFixed(1)}k`, delta: '+12%', good: true },
  { label: 'Covers', value: totalCovers.toLocaleString(), delta: '+8%', good: true },
  { label: 'Avg order', value: `$${avgOrder}`, delta: '+4%', good: true },
  { label: 'Peak day', value: peakDay, delta: 'vs Thu', good: true },
];

export const InsightsDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const [drawn, setDrawn] = useState(false);
  const [hoveredDay, setHoveredDay] = useState<number | null>(null);

  useEffect(() => {
    if (isActive) {
      const t = setTimeout(() => setDrawn(true), 150);
      return () => clearTimeout(t);
    } else {
      setDrawn(false);
    }
  }, [isActive]);

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
        padding: '20px',
      }}
    >
      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 20,
        }}
      >
        <div>
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>Revenue</span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12, marginLeft: 10 }}>
            Last 7 days
          </span>
        </div>
        <div style={{ display: 'flex', gap: 6 }}>
          {['7d', '30d', '90d'].map((r, i) => (
            <div
              key={r}
              style={{
                padding: '4px 10px',
                borderRadius: 6,
                fontSize: 11,
                background: i === 0 ? 'var(--bg-surface-b)' : 'transparent',
                color: i === 0 ? 'var(--ink)' : 'rgba(19,19,19,0.45)',
                border: i === 0 ? '1px solid var(--hairline-strong)' : '1px solid transparent',
                cursor: 'pointer',
              }}
            >
              {r}
            </div>
          ))}
        </div>
      </div>

      {/* KPI tiles */}
      <div
        style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 8, marginBottom: 20 }}
      >
        {KPIs.map((k) => (
          <div
            key={k.label}
            style={{
              background: 'var(--bg-surface-b)',
              border: '1px solid var(--hairline)',
              borderRadius: 8,
              padding: '10px 12px',
            }}
          >
            <div style={{ fontSize: 10, color: 'rgba(19,19,19,0.55)', marginBottom: 5 }}>
              {k.label}
            </div>
            <div style={{ fontSize: 16, fontWeight: 700, color: 'var(--ink)', marginBottom: 3 }}>
              {k.value}
            </div>
            <div style={{ fontSize: 10, color: k.good ? '#16a34a' : '#dc2626' }}>{k.delta}</div>
          </div>
        ))}
      </div>

      {/* Chart */}
      <div style={{ flex: 1, minHeight: 0, position: 'relative' }}>
        <svg
          width="100%"
          height="100%"
          viewBox={`0 0 ${W} ${H}`}
          preserveAspectRatio="none"
          style={{ overflow: 'visible' }}
        >
          {/* Y-axis labels */}
          {[0, 0.5, 1].map((t) => {
            const y = PAD.t + (1 - t) * (H - PAD.t - PAD.b);
            return (
              <text
                key={t}
                x={PAD.l - 6}
                y={y + 4}
                textAnchor="end"
                fontSize={8}
                fill="rgba(19,19,19,0.4)"
              >
                ${((max * t) / 1000).toFixed(0)}k
              </text>
            );
          })}

          {/* Grid lines */}
          {[0, 0.5, 1].map((t) => {
            const y = PAD.t + (1 - t) * (H - PAD.t - PAD.b);
            return (
              <line
                key={t}
                x1={PAD.l}
                y1={y}
                x2={W - PAD.r}
                y2={y}
                stroke="rgba(19,19,19,0.08)"
                strokeWidth={0.5}
              />
            );
          })}

          {/* Area fill */}
          <motion.path
            d={areaPath}
            fill="url(#grad)"
            initial={{ opacity: 0 }}
            animate={{ opacity: drawn ? 1 : 0 }}
            transition={{ duration: 0.6 }}
          />
          <defs>
            <linearGradient id="grad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#131313" stopOpacity="0.12" />
              <stop offset="100%" stopColor="#131313" stopOpacity="0" />
            </linearGradient>
          </defs>

          {/* Line */}
          <motion.path
            d={linePath}
            fill="none"
            stroke="#131313"
            strokeWidth={1.5}
            strokeLinecap="round"
            strokeLinejoin="round"
            initial={{ pathLength: 0 }}
            animate={{ pathLength: drawn ? 1 : 0 }}
            transition={{ duration: 1, ease: 'easeOut' }}
          />

          {/* Day labels + dots */}
          {pts.map(([x, y], i) => (
            <g
              key={i}
              onMouseEnter={() => setHoveredDay(i)}
              onMouseLeave={() => setHoveredDay(null)}
            >
              <circle
                cx={x}
                cy={y}
                r={hoveredDay === i ? 5 : 3}
                fill="#131313"
                style={{ transition: 'r 0.15s ease' }}
                opacity={drawn ? 1 : 0}
              />
              <text
                x={x}
                y={H - PAD.b + 14}
                textAnchor="middle"
                fontSize={9}
                fill={hoveredDay === i ? '#131313' : 'rgba(19,19,19,0.45)'}
                style={{ transition: 'fill 0.15s ease' }}
              >
                {DAYS[i]}
              </text>
              {hoveredDay === i && (
                <g>
                  <rect
                    x={x - 28}
                    y={y - 28}
                    width={56}
                    height={22}
                    rx={4}
                    fill="#131313"
                    stroke="rgba(255,255,255,0.12)"
                    strokeWidth={0.5}
                  />
                  <text
                    x={x}
                    y={y - 13}
                    textAnchor="middle"
                    fontSize={9}
                    fontWeight="700"
                    fill="#ffffff"
                  >
                    ${(REVENUE[i] / 1000).toFixed(1)}k
                  </text>
                </g>
              )}
            </g>
          ))}
        </svg>
      </div>
    </div>
  );
};
