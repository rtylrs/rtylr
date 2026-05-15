import { describe, expect, it } from 'vitest';
import { transcriptToText, transcriptToMarkdown } from './transcripts';
import type { CaretMessageType } from '@voxire/caret';

const sampleMessages: CaretMessageType[] = [
  {
    id: 'u1',
    role: 'user',
    blocks: [],
    content: 'Hello?',
    createdAt: '2026-05-16T12:00:00Z',
  },
  {
    id: 'a1',
    role: 'assistant',
    blocks: [{ id: 'b1', type: 'text', content: 'Hi there.' }],
    content: 'Hi there.',
    createdAt: '2026-05-16T12:00:01Z',
  },
];

describe('transcriptToText', () => {
  it('labels each turn', () => {
    expect(transcriptToText(sampleMessages)).toBe('You: Hello?\n\nCaret: Hi there.');
  });
});

describe('transcriptToMarkdown', () => {
  it('uses bold role labels separated by horizontal rules', () => {
    const md = transcriptToMarkdown(sampleMessages);
    expect(md).toContain('**You**');
    expect(md).toContain('**Caret**');
    expect(md).toContain('---');
  });
});
