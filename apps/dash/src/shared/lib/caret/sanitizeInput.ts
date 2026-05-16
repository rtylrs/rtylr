/**
 * Strip the unicode line/paragraph separators and zero-width spaces that
 * can sneak in via paste from rich-text editors. Keeps newlines.
 */
export function sanitizeInput(s: string): string {
  return s
    .replace(/[​-‍﻿]/g, '')
    .replace(/[  ]/g, '\n')
    .replace(/\r\n?/g, '\n')
    .trimEnd();
}
