import { describe, expect, it } from 'vitest';
import { buildCaretEmptyState } from './suggestions';

describe('buildCaretEmptyState', () => {
  it('does not expose the removed signals CTA', () => {
    const state = buildCaretEmptyState({
      appId: 'insights',
      pageContext: { app: 'insights', module: 'analytics', page: 'Dashboard home' },
    });
    const labels = [state.primaryAction, ...state.suggestions].map((item) => item.label);

    expect(labels).not.toContain('Explain the signals');
  });
});
