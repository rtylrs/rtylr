import type { ContentBlock } from '@voxire/caret';

/** Quick predicate for whether a content block holds any visible text. */
export function blockHasText(block: ContentBlock): boolean {
  if (block.type === 'text') return block.content.trim().length > 0;
  if (block.type === 'thinking') return block.content.trim().length > 0;
  return false;
}
