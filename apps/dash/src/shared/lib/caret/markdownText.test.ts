import { describe, expect, it } from 'vitest';
import { markdownToPreview } from './markdownText';

describe('markdownToPreview', () => {
  it('strips heading markers', () => {
    expect(markdownToPreview('## hello')).toBe('hello');
  });
  it('keeps link text only', () => {
    expect(markdownToPreview('click [here](https://x.com)')).toBe('click here');
  });
  it('drops fenced code blocks', () => {
    expect(markdownToPreview('intro\n```\ncode\n```\nend')).toBe('intro end');
  });
  it('truncates with ellipsis', () => {
    const long = 'a'.repeat(120);
    expect(markdownToPreview(long, 10).endsWith('…')).toBe(true);
  });
});
