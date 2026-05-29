/**
 * Tree and hierarchical data helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export interface TreeNode<T = unknown> {
  id: string;
  parentId: string | null;
  data: T;
}

export function buildChildMap<T>(nodes: TreeNode<T>[]): Map<string | null, TreeNode<T>[]> {
  const map = new Map<string | null, TreeNode<T>[]>();
  for (const node of nodes) {
    const list = map.get(node.parentId) ?? [];
    list.push(node);
    map.set(node.parentId, list);
  }
  return map;
}

export function rootNodes<T>(nodes: TreeNode<T>[]): TreeNode<T>[] {
  return nodes.filter((n) => n.parentId === null);
}

export function childrenOf<T>(nodes: TreeNode<T>[], parentId: string): TreeNode<T>[] {
  return nodes.filter((n) => n.parentId === parentId);
}

export function descendants<T>(nodes: TreeNode<T>[], rootId: string): TreeNode<T>[] {
  const result: TreeNode<T>[] = [];
  const queue = [rootId];
  while (queue.length > 0) {
    const current = queue.shift()!;
    const kids = childrenOf(nodes, current);
    result.push(...kids);
    queue.push(...kids.map((k) => k.id));
  }
  return result;
}

export function nodeDepth<T>(nodes: TreeNode<T>[], id: string): number {
  let depth = 0;
  let current: TreeNode<T> | undefined = nodes.find((n) => n.id === id);
  while (current && current.parentId !== null) {
    depth++;
    current = nodes.find((n) => n.id === current!.parentId);
  }
  return depth;
}

export function ancestorIds<T>(nodes: TreeNode<T>[], id: string): string[] {
  const ids: string[] = [];
  let current: TreeNode<T> | undefined = nodes.find((n) => n.id === id);
  while (current && current.parentId !== null) {
    ids.push(current.parentId);
    current = nodes.find((n) => n.id === current!.parentId);
  }
  return ids;
}

export function isAncestor<T>(
  nodes: TreeNode<T>[],
  ancestorId: string,
  descendantId: string
): boolean {
  return ancestorIds(nodes, descendantId).includes(ancestorId);
}
