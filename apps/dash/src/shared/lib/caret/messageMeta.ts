import type { CaretMessageType } from '@voxire/caret';

/** Best-effort plain-text extract from a message's blocks or content. */
export function messagePlainText(m: CaretMessageType): string {
  if (m.content) return m.content;
  return (m.blocks ?? [])
    .map((b) => {
      if (b.type === 'text') return b.content;
      if (b.type === 'tool_call') return `[tool: ${b.name}]`;
      return '';
    })
    .join('');
}

export function messageWordCount(m: CaretMessageType): number {
  const text = messagePlainText(m).trim();
  if (!text) return 0;
  return text.split(/\s+/).length;
}

export function chatWordCount(messages: CaretMessageType[]): number {
  return messages.reduce((sum, m) => sum + messageWordCount(m), 0);
}
