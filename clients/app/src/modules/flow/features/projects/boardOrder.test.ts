import { describe, expect, it } from 'vitest';
import {
  applyBoardMove,
  insertionIndex,
  isSamePlace,
  orderBoardTasks,
  orderedIdsAfterDrop,
} from './boardOrder';

describe('orderBoardTasks', () => {
  it('keeps the server order for cards nobody has placed', () => {
    const tasks = [{ id: 'c' }, { id: 'a' }, { id: 'b' }];
    expect(orderBoardTasks(tasks).map((t) => t.id)).toEqual(['c', 'a', 'b']);
  });

  it('sorts placed cards by their stored position', () => {
    const tasks = [
      { id: 'a', sortOrder: 3 },
      { id: 'b', sortOrder: 1 },
      { id: 'c', sortOrder: 2 },
    ];
    expect(orderBoardTasks(tasks).map((t) => t.id)).toEqual(['b', 'c', 'a']);
  });

  it('puts unplaced cards above placed ones and keeps their relative order', () => {
    const tasks = [
      { id: 'placed-2', sortOrder: 2 },
      { id: 'new-1', sortOrder: 0 },
      { id: 'placed-1', sortOrder: 1 },
      { id: 'new-2' },
    ];
    expect(orderBoardTasks(tasks).map((t) => t.id)).toEqual([
      'new-1',
      'new-2',
      'placed-1',
      'placed-2',
    ]);
  });

  it('does not mutate its input', () => {
    const tasks = [
      { id: 'a', sortOrder: 2 },
      { id: 'b', sortOrder: 1 },
    ];
    orderBoardTasks(tasks);
    expect(tasks.map((t) => t.id)).toEqual(['a', 'b']);
  });
});

describe('insertionIndex', () => {
  const slots = [
    { id: 'a', mid: 40 },
    { id: 'b', mid: 120 },
    { id: 'c', mid: 200 },
  ];

  it('is the number of other cards above the pointer', () => {
    expect(insertionIndex(slots, 'x', 10)).toBe(0);
    expect(insertionIndex(slots, 'x', 100)).toBe(1);
    expect(insertionIndex(slots, 'x', 160)).toBe(2);
    expect(insertionIndex(slots, 'x', 500)).toBe(3);
  });

  it('ignores the dragged card itself', () => {
    // Dragging "b" with the pointer below its own centre: only "a" is above.
    expect(insertionIndex(slots, 'b', 130)).toBe(1);
    expect(insertionIndex(slots, 'b', 500)).toBe(2);
  });

  it('is zero for an empty column', () => {
    expect(insertionIndex([], 'x', 300)).toBe(0);
  });
});

describe('orderedIdsAfterDrop', () => {
  it('inserts a card from another column at the index', () => {
    expect(orderedIdsAfterDrop(['a', 'b', 'c'], 'x', 1)).toEqual(['a', 'x', 'b', 'c']);
  });

  it('moves a card inside its own column', () => {
    expect(orderedIdsAfterDrop(['a', 'b', 'c'], 'a', 2)).toEqual(['b', 'c', 'a']);
    expect(orderedIdsAfterDrop(['a', 'b', 'c'], 'c', 0)).toEqual(['c', 'a', 'b']);
  });

  it('clamps an out-of-range index', () => {
    expect(orderedIdsAfterDrop(['a', 'b'], 'x', 99)).toEqual(['a', 'b', 'x']);
    expect(orderedIdsAfterDrop(['a', 'b'], 'x', -3)).toEqual(['x', 'a', 'b']);
  });

  it('puts the card alone into an empty column', () => {
    expect(orderedIdsAfterDrop([], 'x', 0)).toEqual(['x']);
  });
});

describe('isSamePlace', () => {
  const source = { columnId: '1', index: 2 };

  it('treats no target as unchanged', () => {
    expect(isSamePlace(source, null)).toBe(true);
  });

  it('is true only for the same column and index', () => {
    expect(isSamePlace(source, { columnId: '1', index: 2 })).toBe(true);
    expect(isSamePlace(source, { columnId: '1', index: 1 })).toBe(false);
    expect(isSamePlace(source, { columnId: '2', index: 2 })).toBe(false);
  });
});

describe('applyBoardMove', () => {
  const tasks = [
    { id: '1', sectionId: '10', sortOrder: 0 },
    { id: '2', sectionId: '10', sortOrder: 0 },
    { id: '3', sectionId: '20', sortOrder: 0 },
    { id: '4', sectionId: '30', sortOrder: 0 },
  ];

  it('places the moved card and stores the listed order', () => {
    const next = applyBoardMove(tasks, {
      taskId: '1',
      columnId: '20',
      orderedTaskIds: ['3', '1'],
    });
    expect(next.find((t) => t.id === '1')).toMatchObject({ sectionId: '20', sortOrder: 2 });
    expect(next.find((t) => t.id === '3')).toMatchObject({ sectionId: '20', sortOrder: 1 });
  });

  it('leaves every other card as the same object', () => {
    const next = applyBoardMove(tasks, {
      taskId: '1',
      columnId: '20',
      orderedTaskIds: ['3', '1'],
    });
    expect(next.find((t) => t.id === '2')).toBe(tasks[1]);
    expect(next.find((t) => t.id === '4')).toBe(tasks[3]);
  });

  it('moves the column only when no order is given', () => {
    const next = applyBoardMove(tasks, { taskId: '2', columnId: '30' });
    expect(next.find((t) => t.id === '2')).toMatchObject({ sectionId: '30', sortOrder: 0 });
  });

  it('returns the same array when nothing matches', () => {
    expect(applyBoardMove(tasks, { taskId: '99', columnId: '10' })).toBe(tasks);
  });
});
