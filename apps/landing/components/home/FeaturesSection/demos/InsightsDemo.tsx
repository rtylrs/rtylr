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
        background: '#0c0c0f',
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
          <span style={{ color: '#e8e6e1', fontWeight: 600, fontSize: 14 }}>Revenue</span>
          <span style={{ color: 'rgba(255,255,255,0.35)', fontSize: 12, marginLeft: 10 }}>
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
                background: i === 0 ? 'rgba(255,255,255,0.08)' : 'transparent',
                color: i === 0 ? '#e8e6e1' : 'rgba(255,255,255,0.35)',
                border: i === 0 ? '1px solid rgba(255,255,255,0.12)' : '1px solid transparent',
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
              background: 'rgba(255,255,255,0.03)',
              border: '1px solid rgba(255,255,255,0.07)',
              borderRadius: 8,
              padding: '10px 12px',
            }}
          >
            <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.38)', marginBottom: 5 }}>
              {k.label}
            </div>
            <div style={{ fontSize: 16, fontWeight: 700, color: '#e8e6e1', marginBottom: 3 }}>
              {k.value}
            </div>
            <div style={{ fontSize: 10, color: k.good ? '#22c55e' : '#f87171' }}>{k.delta}</div>
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
                fill="rgba(255,255,255,0.25)"
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
                stroke="rgba(255,255,255,0.06)"
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
              <stop offset="0%" stopColor="#e8e6e1" stopOpacity="0.12" />
              <stop offset="100%" stopColor="#e8e6e1" stopOpacity="0" />
            </linearGradient>
          </defs>

          {/* Line */}
          <motion.path
            d={linePath}
            fill="none"
            stroke="#e8e6e1"
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
                fill={hoveredDay === i ? '#fff' : '#e8e6e1'}
                style={{ transition: 'r 0.15s ease' }}
                opacity={drawn ? 1 : 0}
              />
              <text
                x={x}
                y={H - PAD.b + 14}
                textAnchor="middle"
                fontSize={9}
                fill={hoveredDay === i ? '#e8e6e1' : 'rgba(255,255,255,0.3)'}
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
                    fill="#1e1e1e"
                    stroke="rgba(255,255,255,0.12)"
                    strokeWidth={0.5}
                  />
                  <text
                    x={x}
                    y={y - 13}
                    textAnchor="middle"
                    fontSize={9}
                    fontWeight="700"
                    fill="#e8e6e1"
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
