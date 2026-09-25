const PLACEHOLDER_TITLES = new Set(['', 'Auto-generated session', 'Untitled chat']);
const FALLBACK_TITLE = 'New chat';
const MAX_TITLE_LENGTH = 80;

function compact(value: string): string {
  return value.trim().replace(/\s+/g, ' ');
}

export function titleFromUserMessage(message: string): string {
  const title = compact(message);
  if (!title) return FALLBACK_TITLE;

  const chars = Array.from(title);
  if (chars.length <= MAX_TITLE_LENGTH) return title;
  return `${chars.slice(0, MAX_TITLE_LENGTH).join('')}...`;
}

export function normalizeChatTitle(title?: string | null): string {
  const compacted = compact(title ?? '');
  return PLACEHOLDER_TITLES.has(compacted) ? FALLBACK_TITLE : titleFromUserMessage(compacted);
}
