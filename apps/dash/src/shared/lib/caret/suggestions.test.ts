import { describe, expect, it } from 'vitest';
import { suggestionsFor } from './suggestions';

describe('suggestionsFor', () => {
  it('returns the dashboard list at the root path', () => {
    expect(suggestionsFor('/')[0]).toBe('How much revenue today?');
  });

  it('returns team-specific suggestions on /team', () => {
    const s = suggestionsFor('/team');
    expect(s.some((p) => p.toLowerCase().includes('team'))).toBe(true);
  });

  it('matches subpaths against the most specific known prefix', () => {
    expect(suggestionsFor('/menus/123')).toEqual(suggestionsFor('/menus'));
  });

  it('falls back to the dashboard list for unknown routes', () => {
    expect(suggestionsFor('/no/such/page')).toEqual(suggestionsFor('/'));
  });
});
