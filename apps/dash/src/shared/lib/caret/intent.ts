/**
 * Tiny keyword-based intent hint. Lets the composer warn the user (or pre-
 * select a different mode) when their message clearly maps to a tool that
 * requires Agent mode. Not a substitute for server-side routing.
 */

export type IntentHint = 'read' | 'write' | 'unknown';

const WRITE_KEYWORDS = [
  'create',
  'add',
  'update',
  'edit',
  'delete',
  'remove',
  'refund',
  'cancel',
  'send',
  'invite',
  'promote',
  'demote',
  'set ',
  'change ',
  'mark',
];

const READ_KEYWORDS = [
  'show ',
  'list ',
  'how many',
  'what is',
  'what are',
  "what's",
  'tell me',
  'find',
  'search',
  'compare',
  'analyze',
  'top',
  'breakdown',
];

export function inferIntent(prompt: string): IntentHint {
  const p = prompt.toLowerCase();
  if (WRITE_KEYWORDS.some((k) => p.includes(k))) return 'write';
  if (READ_KEYWORDS.some((k) => p.startsWith(k) || p.includes(' ' + k))) return 'read';
  return 'unknown';
}
