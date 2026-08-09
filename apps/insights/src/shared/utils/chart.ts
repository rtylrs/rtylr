// Chart helper utilities for the insights app.
// These functions transform raw API series data into shapes that
// recharts (or any other chart lib) can consume directly. They are
// pure functions with no side effects.

export interface LabeledPoint {
  label: string;
  value: number;
}

/**
 * reindexByLabel converts a parallel pair of label and value arrays
 * into an array of {label, value} objects. Useful when the backend
 * sends dates and amounts as two separate arrays.
 */
export function reindexByLabel(labels: string[], values: number[]): LabeledPoint[] {
  const len = Math.min(labels.length, values.length);
  const out: LabeledPoint[] = [];
  for (let i = 0; i < len; i++) {
    out.push({ label: labels[i], value: values[i] });
  }
  return out;
}

/**
 * topN returns the n highest-value items from a LabeledPoint array,
 * sorted descending. Ties are broken by the original order.
 */
export function topN(points: LabeledPoint[], n: number): LabeledPoint[] {
  return [...points].sort((a, b) => b.value - a.value).slice(0, n);
}

/**
 * aggregateByLabel sums values that share the same label. Useful for
 * collapsing duplicate date entries that sometimes appear in daily
 * rollup queries.
 */
export function aggregateByLabel(points: LabeledPoint[]): LabeledPoint[] {
  const map = new Map<string, number>();
  for (const p of points) {
    map.set(p.label, (map.get(p.label) ?? 0) + p.value);
  }
  return Array.from(map.entries()).map(([label, value]) => ({ label, value }));
}

/**
 * fillMissingDates inserts zero-value points for any date in the range
 * [from, to] that does not appear in the input series. Input must be
 * sorted ascending by label (ISO date string YYYY-MM-DD).
 */
export function fillMissingDates(points: LabeledPoint[], from: string, to: string): LabeledPoint[] {
  const existing = new Map(points.map((p) => [p.label, p.value]));
  const result: LabeledPoint[] = [];
  const cur = new Date(from + 'T00:00:00Z');
  const end = new Date(to + 'T00:00:00Z');
  while (cur <= end) {
    const label = cur.toISOString().slice(0, 10);
    result.push({ label, value: existing.get(label) ?? 0 });
    cur.setUTCDate(cur.getUTCDate() + 1);
  }
  return result;
}

/**
 * pctShare converts an array of numeric values into percentage-of-total
 * shares. Returns an equal-length array. If the sum is zero every entry
 * gets 0 so division by zero is avoided.
 */
export function pctShare(values: number[]): number[] {
  const total = values.reduce((a, b) => a + b, 0);
  if (total === 0) return values.map(() => 0);
  return values.map((v) => (v / total) * 100);
}
