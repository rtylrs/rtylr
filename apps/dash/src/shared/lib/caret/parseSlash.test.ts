import { describe, expect, it } from 'vitest';
import { parseSlash } from './parseSlash';

describe('parseSlash', () => {
  it('parses a bare command', () => {
    expect(parseSlash('/help')).toEqual({ command: 'help', args: '' });
  });
  it('captures arguments', () => {
    expect(parseSlash('/report sales month')).toEqual({
      command: 'report',
      args: 'sales month',
    });
  });
  it('returns null for non-slash input', () => {
    expect(parseSlash('hello')).toBeNull();
  });
});
