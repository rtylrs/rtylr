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
    expect(state.title).toBe('What should we work on?');
    expect(state.primaryAction.label).toBe('Start with a brief');
    expect(state.composerPlaceholder).toBe('Ask a question about this page…');
    expect(state.contextSignals).toContainEqual({
      label: 'Screen',
      value: 'Dashboard home',
    });
  });
});
