import { describe, expect, it } from 'vitest';
import { messagePlainText, messageWordCount, chatWordCount } from './messageMeta';
import type { CaretMessageType } from '@voxire/caret';

const msg = (content: string, blocks?: CaretMessageType['blocks']): CaretMessageType => ({
  id: 'm',
  role: 'assistant',
  blocks: blocks ?? [],
  content,
  createdAt: '2026-05-16T12:00:00Z',
});

describe('messagePlainText', () => {
  it('returns content when present', () => {
    expect(messagePlainText(msg('hello'))).toBe('hello');
  });

  it('falls back to text blocks', () => {
    expect(messagePlainText(msg('', [{ id: 'a', type: 'text', content: 'from block' }]))).toBe(
      'from block'
    );
  });
});

describe('messageWordCount', () => {
  it('counts whitespace-separated words', () => {
    expect(messageWordCount(msg('one two three'))).toBe(3);
  });

  it('returns 0 for empty strings', () => {
    expect(messageWordCount(msg(''))).toBe(0);
  });
});

describe('chatWordCount', () => {
  it('sums words across messages', () => {
    expect(chatWordCount([msg('one two'), msg('three')])).toBe(3);
  });
});
