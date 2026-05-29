import { describe, it, expect } from 'vitest';
import {
  buildChildMap,
  rootNodes,
  childrenOf,
  descendants,
  nodeDepth,
  ancestorIds,
  isAncestor,
  type TreeNode,
} from './treeUtils';

const nodes: TreeNode<string>[] = [
  { id: 'root', parentId: null, data: 'Root' },
  { id: 'a', parentId: 'root', data: 'A' },
  { id: 'b', parentId: 'root', data: 'B' },
  { id: 'a1', parentId: 'a', data: 'A1' },
  { id: 'a2', parentId: 'a', data: 'A2' },
  { id: 'a1x', parentId: 'a1', data: 'A1X' },
];

describe('buildChildMap', () => {
  it('groups by parent', () => {
    const map = buildChildMap(nodes);
    expect(map.get('root')).toHaveLength(2);
    expect(map.get('a')).toHaveLength(2);
  });
});

describe('rootNodes', () => {
  it('returns nodes with no parent', () => {
    const roots = rootNodes(nodes);
    expect(roots).toHaveLength(1);
    expect(roots[0].id).toBe('root');
  });
});

describe('childrenOf', () => {
  it('returns direct children', () =>
    expect(
      childrenOf(nodes, 'a')
        .map((n) => n.id)
        .sort()
    ).toEqual(['a1', 'a2']));
  it('empty for leaf', () => expect(childrenOf(nodes, 'a1x')).toHaveLength(0));
});

describe('descendants', () => {
  it('returns all descendants', () =>
    expect(
      descendants(nodes, 'a')
        .map((n) => n.id)
        .sort()
    ).toEqual(['a1', 'a1x', 'a2']));
  it('empty for leaf', () => expect(descendants(nodes, 'b')).toHaveLength(0));
});

describe('nodeDepth', () => {
  it('root=0', () => expect(nodeDepth(nodes, 'root')).toBe(0));
  it('first level=1', () => expect(nodeDepth(nodes, 'a')).toBe(1));
  it('deep=3', () => expect(nodeDepth(nodes, 'a1x')).toBe(3));
});

describe('ancestorIds', () => {
  it('from parent to root', () => expect(ancestorIds(nodes, 'a1x')).toEqual(['a1', 'a', 'root']));
  it('empty for root', () => expect(ancestorIds(nodes, 'root')).toEqual([]));
});

describe('isAncestor', () => {
  it('true for direct parent', () => expect(isAncestor(nodes, 'a', 'a1')).toBe(true));
  it('true for indirect ancestor', () => expect(isAncestor(nodes, 'root', 'a1x')).toBe(true));
  it('false for non-ancestor', () => expect(isAncestor(nodes, 'b', 'a1')).toBe(false));
});
