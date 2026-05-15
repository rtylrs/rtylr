import { describe, expect, it } from 'vitest';
import { visualForTool } from './toolIcons';

describe('visualForTool', () => {
  it('returns the CRM visual for crm.* tools', () => {
    const v = visualForTool('crm.list_customers');
    expect(v.color).toBe('#a855f7');
  });

  it('returns the fallback for unknown modules', () => {
    const v = visualForTool('unknown.thing');
    expect(v.color).toBe('#a1a1aa');
  });

  it('handles a name without a dot', () => {
    expect(visualForTool('plain').color).toBe('#a1a1aa');
  });
});
