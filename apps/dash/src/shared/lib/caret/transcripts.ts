import type { CaretMessageType } from '@voxire/caret';

/** Format a transcript as plain text for clipboard / download. */
export function transcriptToText(messages: CaretMessageType[]): string {
  return messages
    .map((m) => {
      const who = m.role === 'user' ? 'You' : 'Caret';
      const text = m.content ?? m.blocks?.map((b) => (b.type === 'text' ? b.content : '')).join('');
      return `${who}: ${text}`;
    })
    .join('\n\n');
}

/** Format a transcript as Markdown for rich pasting. */
export function transcriptToMarkdown(messages: CaretMessageType[]): string {
  return messages
    .map((m) => {
      const who = m.role === 'user' ? '**You**' : '**Caret**';
      const text = m.content ?? m.blocks?.map((b) => (b.type === 'text' ? b.content : '')).join('');
      return `${who}\n\n${text}`;
    })
    .join('\n\n---\n\n');
}
