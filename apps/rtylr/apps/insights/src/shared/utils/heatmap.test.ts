import { describe, it, expect } from 'vitest';
import {
  heatmapMin,
  heatmapMax,
  normalizeCell,
  heatmapRows,
  heatmapCols,
  cellsForRow,
  hotspots,
} from './heatmap';
import type { HeatmapCell } from './heatmap';

function makeCell(row: string, col: string, value: number): HeatmapCell {
  return { row, col, value };
}

const cells = [
  makeCell('Mon', '9am', 10),
  makeCell('Mon', '10am', 50),
  makeCell('Tue', '9am', 30),
  makeCell('Tue', '10am', 80),
];

describe('heatmapMin', () => {
  it('returns minimum value', () => {
    expect(heatmapMin(cells)).toBe(10);
  });

  it('returns 0 for empty cells', () => {
    expect(heatmapMin([])).toBe(0);
  });
});

describe('heatmapMax', () => {
  it('returns maximum value', () => {
    expect(heatmapMax(cells)).toBe(80);
  });

  it('returns 0 for empty cells', () => {
    expect(heatmapMax([])).toBe(0);
  });
});

describe('normalizeCell', () => {
  it('normalizes to 0 for min value', () => {
    expect(normalizeCell(cells[0], 10, 80)).toBe(0);
  });

  it('normalizes to 1 for max value', () => {
    expect(normalizeCell(cells[3], 10, 80)).toBe(1);
  });

  it('returns 0 when min equals max', () => {
    expect(normalizeCell(cells[0], 10, 10)).toBe(0);
  });
});

describe('heatmapRows', () => {
  it('returns unique row labels', () => {
    expect(heatmapRows(cells)).toEqual(['Mon', 'Tue']);
  });
});

describe('heatmapCols', () => {
  it('returns unique column labels', () => {
    expect(heatmapCols(cells)).toEqual(['9am', '10am']);
  });
});

describe('cellsForRow', () => {
  it('returns cells in specified row', () => {
    expect(cellsForRow(cells, 'Mon')).toHaveLength(2);
  });
});

describe('hotspots', () => {
  it('returns cells above 90th percentile threshold', () => {
    const hot = hotspots(cells, 0.9);
    expect(hot.every((c) => c.value >= 10 + (80 - 10) * 0.9)).toBe(true);
  });

  it('returns empty for empty cells', () => {
    expect(hotspots([])).toHaveLength(0);
  });
});
