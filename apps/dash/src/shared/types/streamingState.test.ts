import { describe, expect, it } from 'vitest';
import { streamingLabel } from './streamingState';

describe('streamingLabel', () => {
  it('returns empty for idle', () => {
    expect(streamingLabel('idle')).toBe('');
  });
  it('returns Thinking… for thinking', () => {
    expect(streamingLabel('thinking')).toBe('Thinking…');
  });
});
