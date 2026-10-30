import { describe, expect, it } from 'vitest';
import { moodOf } from './useBuddyMood';

const reply = (blocks: unknown[]) =>
  [
    { id: 'u', role: 'user', blocks: [], createdAt: '' },
    { id: 'a', role: 'assistant', blocks, createdAt: '' },
  ] as never;

describe('moodOf', () => {
  it('thinks before the first words, then talks', () => {
    expect(moodOf({ streaming: true, messages: reply([]) })).toBe('thinking');
    expect(
      moodOf({ streaming: true, messages: reply([{ id: 't', type: 'text', content: 'Hi' }]) })
    ).toBe('talking');
  });

  it('works while a tool runs', () => {
    expect(
      moodOf({
        streaming: true,
        messages: reply([
          { id: 'c', type: 'tool_call', name: 'crm.list', input: {}, status: 'running' },
        ]),
      })
    ).toBe('working');
  });

  it('waits on an approval even after the stream ends', () => {
    expect(
      moodOf({
        streaming: false,
        messages: reply([
          {
            id: 'c',
            type: 'tool_call',
            name: 'buddy.create_routine',
            input: {},
            status: 'pending',
            requiresApproval: true,
          },
        ]),
      })
    ).toBe('waiting');
  });

  it('rests when nothing is happening', () => {
    expect(moodOf({ streaming: false, messages: [] as never })).toBe('idle');
  });
});
