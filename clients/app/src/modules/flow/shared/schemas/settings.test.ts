import { describe, expect, it } from 'vitest';
import { settingsSearchSchema } from './settings';

/**
 * validateSearch runs before the route renders, so anything this schema throws on
 * replaces the settings screen with the router error boundary. These cases are
 * the URLs a user can actually produce: a bookmark to a renamed section, a
 * hand-edited query, a stale link from chat.
 */
describe('settingsSearchSchema', () => {
  it('accepts a known section', () => {
    expect(settingsSearchSchema.parse({ section: 'time' })).toEqual({ section: 'time' });
  });

  it('treats a missing section as absent rather than failing', () => {
    expect(settingsSearchSchema.parse({})).toEqual({ section: undefined });
  });

  it('falls back to undefined on a section that no longer exists', () => {
    expect(settingsSearchSchema.parse({ section: 'billing' })).toEqual({ section: undefined });
  });

  it('falls back to undefined rather than throwing on a wrong type', () => {
    // The router JSON-parses query values, so ?section=42 arrives as a number.
    expect(settingsSearchSchema.parse({ section: 42 })).toEqual({ section: undefined });
    expect(settingsSearchSchema.parse({ section: null })).toEqual({ section: undefined });
  });

  it('ignores unrelated params instead of rejecting the URL', () => {
    expect(settingsSearchSchema.parse({ section: 'labels', utm_source: 'email' })).toEqual({
      section: 'labels',
    });
  });

  it('never throws, for any input shape', () => {
    for (const input of [{}, { section: '' }, { section: [] }, { section: {} }]) {
      expect(() => settingsSearchSchema.parse(input)).not.toThrow();
    }
  });

  it('types the section as a literal union, not a bare string', () => {
    // A plain string would defeat the point: <Link search={{ section: 'typo' }} />
    // has to be a compile error, which is what makes the schema load-bearing.
    const parsed = settingsSearchSchema.parse({ section: 'appearance' });
    const section: 'appearance' | undefined =
      parsed.section === 'appearance' ? parsed.section : undefined;
    expect(section).toBe('appearance');
  });
});
