import { describe, it, expect } from 'vitest';
import {
  DRAG_THRESHOLD_PX,
  columnIdAtPoint,
  exceedsDragThreshold,
  resolveDropTarget,
  type ColumnRect,
} from './boardDrag';

describe('exceedsDragThreshold', () => {
  it('treats a still pointer as a click, not a drag', () => {
    expect(exceedsDragThreshold(0, 0)).toBe(false);
  });

  it('ignores movement below the threshold on a single axis', () => {
    expect(exceedsDragThreshold(DRAG_THRESHOLD_PX - 1, 0)).toBe(false);
    expect(exceedsDragThreshold(0, -(DRAG_THRESHOLD_PX - 1))).toBe(false);
  });

  it('counts movement at the threshold', () => {
    expect(exceedsDragThreshold(DRAG_THRESHOLD_PX, 0)).toBe(true);
  });

  // The reason for radial distance rather than two axis checks: 5 and 5 is 7.07
  // px of travel, which is past a 6px threshold even though neither axis is.
  it('counts a diagonal nudge whose axes are each below the threshold', () => {
    expect(exceedsDragThreshold(5, 5)).toBe(true);
  });

  it('is direction agnostic', () => {
    expect(exceedsDragThreshold(-20, 0)).toBe(true);
    expect(exceedsDragThreshold(0, 20)).toBe(true);
  });
});

describe('columnIdAtPoint', () => {
  const rects: ColumnRect[] = [
    { id: 'a', left: 0, right: 100, top: 0, bottom: 200 },
    { id: 'b', left: 120, right: 220, top: 0, bottom: 200 },
  ];

  it('finds the column under the point', () => {
    expect(columnIdAtPoint(rects, 50, 100)).toBe('a');
    expect(columnIdAtPoint(rects, 150, 100)).toBe('b');
  });

  it('returns null in the gutter between columns', () => {
    expect(columnIdAtPoint(rects, 110, 100)).toBeNull();
  });

  it('returns null outside the board vertically', () => {
    expect(columnIdAtPoint(rects, 50, 500)).toBeNull();
  });

  it('includes both edges so the seam is never dead', () => {
    expect(columnIdAtPoint(rects, 0, 0)).toBe('a');
    expect(columnIdAtPoint(rects, 100, 200)).toBe('a');
  });

  it('returns null when there are no columns', () => {
    expect(columnIdAtPoint([], 10, 10)).toBeNull();
  });
});

describe('resolveDropTarget', () => {
  it('moves the task when the target differs', () => {
    expect(resolveDropTarget('b', 'a')).toBe('b');
  });

  it('is a no-op when dropped on its own column', () => {
    expect(resolveDropTarget('a', 'a')).toBeNull();
  });

  it('is a no-op when dropped outside every column', () => {
    expect(resolveDropTarget(null, 'a')).toBeNull();
  });

  it('still moves a task whose source column is unknown', () => {
    expect(resolveDropTarget('b', null)).toBe('b');
  });
});
