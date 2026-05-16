import type { ToolDef } from '../../api/caret';

/** Partition a tool registry into read-only and write-scope arrays. */
export function partitionTools(tools: ToolDef[]): { read: ToolDef[]; write: ToolDef[] } {
  const read: ToolDef[] = [];
  const write: ToolDef[] = [];
  for (const t of tools) {
    (t.scope === 'read' ? read : write).push(t);
  }
  return { read, write };
}
