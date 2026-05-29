// Heatmap data helpers for the insights analytics feature.

export interface HeatmapCell {
  row: string;
  col: string;
  value: number;
}

/**
 * heatmapMin returns the minimum value across all cells.
 */
export function heatmapMin(cells: HeatmapCell[]): number {
  if (cells.length === 0) return 0;
  return Math.min(...cells.map((c) => c.value));
}

/**
 * heatmapMax returns the maximum value across all cells.
 */
export function heatmapMax(cells: HeatmapCell[]): number {
  if (cells.length === 0) return 0;
  return Math.max(...cells.map((c) => c.value));
}

/**
 * normalizeCell returns a 0-1 normalized value for a cell relative to min/max.
 */
export function normalizeCell(cell: HeatmapCell, min: number, max: number): number {
  if (max === min) return 0;
  return Math.round(((cell.value - min) / (max - min)) * 1000) / 1000;
}

/**
 * heatmapRows returns unique row labels from the cells.
 */
export function heatmapRows(cells: HeatmapCell[]): string[] {
  return [...new Set(cells.map((c) => c.row))];
}

/**
 * heatmapCols returns unique column labels from the cells.
 */
export function heatmapCols(cells: HeatmapCell[]): string[] {
  return [...new Set(cells.map((c) => c.col))];
}

/**
 * cellsForRow returns all cells belonging to a specific row.
 */
export function cellsForRow(cells: HeatmapCell[], row: string): HeatmapCell[] {
  return cells.filter((c) => c.row === row);
}

/**
 * hotspots returns cells with values above the given percentile threshold.
 */
export function hotspots(cells: HeatmapCell[], pct = 0.9): HeatmapCell[] {
  if (cells.length === 0) return [];
  const max = heatmapMax(cells);
  const min = heatmapMin(cells);
  const threshold = min + (max - min) * pct;
  return cells.filter((c) => c.value >= threshold);
}
