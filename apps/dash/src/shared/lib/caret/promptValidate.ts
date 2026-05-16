import { CARET_LIMITS } from './limits';

/** Validate a composer prompt before sending. Returns null when valid. */
export function validatePrompt(prompt: string): string | null {
  const trimmed = prompt.trim();
  if (!trimmed) return 'Type something first.';
  if (trimmed.length > CARET_LIMITS.maxMessageChars) {
    return `Message is too long (${trimmed.length}/${CARET_LIMITS.maxMessageChars} chars).`;
  }
  return null;
}
