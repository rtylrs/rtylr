import { describe, expect, it } from 'vitest';
import { parseMentions } from './parseMention';

describe('parseMentions', () => {
  it('finds a leading mention', () => {
    expect(parseMentions('@crm find john')).toEqual([{ app: 'crm', start: 0, end: 4 }]);
  });
  it('finds a mid-string mention', () => {
    const out = parseMentions('hello @pos thing');
    expect(out.length).toBe(1);
    expect(out[0].app).toBe('pos');
  });
  it('returns empty for no mentions', () => {
    expect(parseMentions('plain text')).toEqual([]);
  });
});
