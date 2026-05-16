import { describe, expect, it } from 'vitest';
import { filterToolsByMode } from './modeFilter';
import type { ToolDef } from '../../api/caret';

const tools: ToolDef[] = [
  { name: 'a.read', description: '', module: 'a', scope: 'read', requiresApproval: false },
  { name: 'a.write', description: '', module: 'a', scope: 'write', requiresApproval: true },
];

describe('filterToolsByMode', () => {
  it('keeps writes in agent mode', () => {
    expect(filterToolsByMode(tools, 'agent')).toHaveLength(2);
  });
  it('drops writes in ask mode', () => {
    expect(filterToolsByMode(tools, 'ask')).toHaveLength(1);
  });
  it('returns no tools in support mode', () => {
    expect(filterToolsByMode(tools, 'support')).toHaveLength(0);
  });
});
