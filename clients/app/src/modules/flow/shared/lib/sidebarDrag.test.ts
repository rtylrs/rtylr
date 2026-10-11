import { describe, expect, it } from 'vitest';
import { isWithinSameList, placementOver, reorderGroupIds, samePlacement } from './sidebarDrag';

describe('placementOver', () => {
  const list = ['a', 'b', 'c', 'd'];

  it('goes before the hovered row when the pointer is in its upper half', () => {
    expect(placementOver(list, 'a', 'c', false)).toEqual({
      groupId: undefined,
      beforeItemId: 'c',
    });
  });

  it('goes before the next row when the pointer is in its lower half', () => {
    expect(placementOver(list, 'a', 'b', true)).toEqual({
      groupId: undefined,
      beforeItemId: 'c',
    });
  });

  it('goes to the end when the hovered row is the last one and the pointer is below it', () => {
    expect(placementOver(list, 'a', 'd', true)).toEqual({
      groupId: undefined,
      beforeItemId: undefined,
    });
  });

  it('skips the dragged row when finding the neighbour', () => {
    // Dragging "b" over the lower half of "a": the next row is "c", not "b".
    expect(placementOver(list, 'b', 'a', true).beforeItemId).toBe('c');
    // Dragging "c" over the lower half of "b": the next row is "d".
    expect(placementOver(list, 'c', 'b', true).beforeItemId).toBe('d');
  });

  it('carries the group id through', () => {
    expect(placementOver(list, 'a', 'c', false, 'g1').groupId).toBe('g1');
  });

  it('falls back to the list end for a row that is not in the list', () => {
    expect(placementOver(list, 'a', 'zzz', false, 'g1')).toEqual({ groupId: 'g1' });
  });
});

describe('samePlacement', () => {
  it('compares group and neighbour', () => {
    expect(
      samePlacement({ groupId: 'g', beforeItemId: 'x' }, { groupId: 'g', beforeItemId: 'x' })
    ).toBe(true);
    expect(
      samePlacement({ groupId: 'g', beforeItemId: 'x' }, { groupId: 'g', beforeItemId: 'y' })
    ).toBe(false);
    expect(samePlacement({ groupId: 'g' }, { groupId: 'h' })).toBe(false);
  });

  it('treats two missing placements as equal and one missing as different', () => {
    expect(samePlacement(undefined, undefined)).toBe(true);
    expect(samePlacement(undefined, { groupId: 'g' })).toBe(false);
  });
});

describe('isWithinSameList', () => {
  it("is true only when the placement targets the row's own list", () => {
    expect(isWithinSameList(undefined, { groupId: undefined, beforeItemId: 'a' })).toBe(true);
    expect(isWithinSameList('g1', { groupId: 'g1' })).toBe(true);
    expect(isWithinSameList('g1', { groupId: 'g2' })).toBe(false);
    expect(isWithinSameList(undefined, { groupId: 'g2' })).toBe(false);
  });
});

describe('reorderGroupIds', () => {
  const order = ['projects', 'plan', 'time'];

  it('moves the dragged group to the target position', () => {
    expect(reorderGroupIds(order, 'projects', 'time')).toEqual(['plan', 'time', 'projects']);
    expect(reorderGroupIds(order, 'time', 'projects')).toEqual(['time', 'projects', 'plan']);
  });

  it('is a no-op for the same or an unknown group', () => {
    expect(reorderGroupIds(order, 'plan', 'plan')).toEqual(order);
    expect(reorderGroupIds(order, 'plan', 'nope')).toEqual(order);
  });

  it('does not mutate the input', () => {
    const input = ['a', 'b'];
    reorderGroupIds(input, 'a', 'b');
    expect(input).toEqual(['a', 'b']);
  });
});
