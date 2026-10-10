import React from "react";
import { caretTokens } from "../../ui/tokens";

export interface ChartSpec {
  type: "bar" | "line";
  title?: string;
  unit?: string;
  labels: string[];
  series: Array<{ name: string; values: number[] }>;
}

const SERIES_COLORS = [caretTokens.accent, "#8fa8ff", "#f2b56b"];
const MAX_POINTS = 60;
const MAX_SERIES = 3;

/** Parses a ```chart block; null when it isn't a usable chart. */
export function parseChartSpec(raw: string): ChartSpec | null {
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch {
    return null;
  }
  if (!value || typeof value !== "object") return null;
  const spec = value as Record<string, unknown>;
  const labels = Array.isArray(spec.labels) ? spec.labels.map(String) : [];
  const series = (Array.isArray(spec.series) ? spec.series : [])
    .slice(0, MAX_SERIES)
    .map((item) => {
      const entry = (item ?? {}) as Record<string, unknown>;
      const values = Array.isArray(entry.values)
        ? entry.values.map((v) => Number(v))
        : [];
      return { name: String(entry.name ?? ""), values };
    })
    .filter(
      (item) =>
        item.values.length === labels.length &&
        item.values.every((v) => Number.isFinite(v)),
    );
  if (labels.length < 2 || labels.length > MAX_POINTS || series.length === 0)
    return null;
  return {
    type: spec.type === "line" ? "line" : "bar",
    title: typeof spec.title === "string" ? spec.title : undefined,
    unit: typeof spec.unit === "string" ? spec.unit : undefined,
    labels,
    series,
  };
}

function formatNumber(value: number): string {
  const abs = Math.abs(value);
  if (abs >= 1e9) return `${(value / 1e9).toFixed(1)}B`;
  if (abs >= 1e6) return `${(value / 1e6).toFixed(1)}M`;
  if (abs >= 1e4) return `${(value / 1e3).toFixed(0)}k`;
  if (abs >= 1e3) return `${(value / 1e3).toFixed(1)}k`;
  return Number.isInteger(value) ? String(value) : value.toFixed(2);
}

/** A small SVG bar or line chart drawn from a ```chart block. */
export const ChartBlock: React.FC<{ spec: ChartSpec }> = ({ spec }) => {
  const width = 640;
  const height = 240;
  const pad = { top: 12, right: 12, bottom: 34, left: 48 };
  const plotW = width - pad.left - pad.right;
  const plotH = height - pad.top - pad.bottom;
  const all = spec.series.flatMap((s) => s.values);
  const max = Math.max(0, ...all);
  const min = Math.min(0, ...all);
  const range = max - min || 1;
  const y = (v: number) => pad.top + plotH - ((v - min) / range) * plotH;
  const step = plotW / spec.labels.length;
  const labelEvery = Math.ceil(spec.labels.length / 10);
  const ticks = [min, min + range / 2, max];

  return (
    <figure className="caret-chart" style={{ margin: "10px 0" }}>
      {spec.title && (
        <figcaption
          style={{ fontSize: 12.5, color: caretTokens.textMuted, marginBottom: 4 }}
        >
          {spec.title}
          {spec.unit ? ` (${spec.unit})` : ""}
        </figcaption>
      )}
      <svg
        role="img"
        aria-label={spec.title ?? "Chart"}
        viewBox={`0 0 ${width} ${height}`}
        style={{ width: "100%", height: "auto", display: "block" }}
      >
        {ticks.map((tick) => (
          <g key={tick}>
            <line
              x1={pad.left}
              x2={width - pad.right}
              y1={y(tick)}
              y2={y(tick)}
              stroke={caretTokens.border}
              strokeDasharray="3 4"
            />
            <text
              x={pad.left - 6}
              y={y(tick) + 4}
              textAnchor="end"
              fontSize="11"
              fill={caretTokens.textDim}
            >
              {formatNumber(tick)}
            </text>
          </g>
        ))}
        {spec.labels.map((label, i) =>
          i % labelEvery === 0 ? (
            <text
              key={`${label}-${i}`}
              x={pad.left + step * i + step / 2}
              y={height - pad.bottom + 16}
              textAnchor="middle"
              fontSize="11"
              fill={caretTokens.textDim}
            >
              {label.length > 10 ? `${label.slice(0, 9)}…` : label}
            </text>
          ) : null,
        )}
        {spec.series.map((series, s) => {
          const color = SERIES_COLORS[s % SERIES_COLORS.length];
          if (spec.type === "line") {
            const points = series.values
              .map((v, i) => `${pad.left + step * i + step / 2},${y(v)}`)
              .join(" ");
            return (
              <g key={series.name || s}>
                <polyline points={points} fill="none" stroke={color} strokeWidth="2" />
                {series.values.map((v, i) => (
                  <circle key={i} cx={pad.left + step * i + step / 2} cy={y(v)} r="3" fill={color}>
                    <title>{`${spec.labels[i]}: ${formatNumber(v)}${spec.unit ? ` ${spec.unit}` : ""}`}</title>
                  </circle>
                ))}
              </g>
            );
          }
          const barW = (step * 0.7) / spec.series.length;
          return (
            <g key={series.name || s}>
              {series.values.map((v, i) => (
                <rect
                  key={i}
                  x={pad.left + step * i + step * 0.15 + barW * s}
                  y={Math.min(y(v), y(0))}
                  width={Math.max(1, barW - 1)}
                  height={Math.max(1, Math.abs(y(v) - y(0)))}
                  rx="2"
                  fill={color}
                >
                  <title>{`${spec.labels[i]}: ${formatNumber(v)}${spec.unit ? ` ${spec.unit}` : ""}`}</title>
                </rect>
              ))}
            </g>
          );
        })}
      </svg>
      {spec.series.length > 1 && (
        <div style={{ display: "flex", gap: 12, fontSize: 11.5, color: caretTokens.textMuted }}>
          {spec.series.map((series, s) => (
            <span key={series.name || s} style={{ display: "inline-flex", alignItems: "center", gap: 5 }}>
              <span style={{ width: 9, height: 9, borderRadius: 2, background: SERIES_COLORS[s % SERIES_COLORS.length] }} />
              {series.name}
            </span>
          ))}
        </div>
      )}
    </figure>
  );
};

export default ChartBlock;
