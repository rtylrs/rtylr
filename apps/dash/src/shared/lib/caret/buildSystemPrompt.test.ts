import { describe, expect, it } from 'vitest';
import { buildSystemPromptPreview } from './buildSystemPrompt';

describe('buildSystemPromptPreview', () => {
  it('includes the mode addendum', () => {
    const out = buildSystemPromptPreview({ mode: 'ask', systemContext: 'on dashboard' });
    expect(out).toContain('Ask mode');
  });

  it('includes the system context', () => {
    const out = buildSystemPromptPreview({ mode: 'ask', systemContext: 'on dashboard' });
    expect(out).toContain('on dashboard');
  });
});
