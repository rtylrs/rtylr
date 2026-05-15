/**
 * Derive a short chat title from the first user message. Used as a
 * placeholder before the backend generates a real summary-based title.
 */
export function deriveChatTitle(firstMessage: string, maxLen = 48): string {
  const cleaned = firstMessage.trim().replace(/\s+/g, ' ');
  if (cleaned.length <= maxLen) return cleaned;
  return cleaned.slice(0, maxLen - 1).trimEnd() + '…';
}
