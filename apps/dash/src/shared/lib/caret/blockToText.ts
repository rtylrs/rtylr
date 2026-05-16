import type { ContentBlock } from '@voxire/caret';

/** Flatten a block to plain-text representation. */
export function blockToText(b: ContentBlock): string {
  if (b.type === 'text') return b.content;
  if (b.type === 'thinking') return `[thinking]\n${b.content}`;
  return `[tool ${b.name}: ${b.status}]`;
}
