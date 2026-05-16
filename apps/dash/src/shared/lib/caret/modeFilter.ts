import type { CaretMode } from '@voxire/caret';
import type { ToolDef } from '../../api/caret';

/** Drop write-scope tools when the mode is read-only. */
export function filterToolsByMode(tools: ToolDef[], mode: CaretMode): ToolDef[] {
  if (mode === 'ask' || mode === 'analyst') {
    return tools.filter((t) => t.scope === 'read');
  }
  if (mode === 'support') return [];
  return tools;
}
