import { DEFAULT_MODE, type CaretMode } from '../types/mode';

const AGENT_PHRASES = [
  'add ',
  'approve ',
  'book ',
  'cancel ',
  'change ',
  'create ',
  'draft ',
  'make ',
  'mark ',
  'move ',
  'notify ',
  'record ',
  'refund ',
  'remember ',
  'remove ',
  'save ',
  'schedule ',
  'send ',
  'set ',
  'transfer ',
  'update ',
];

const ANALYST_PHRASES = [
  'best seller',
  'breakdown',
  'compare',
  'dashboard',
  'forecast',
  'how many',
  'how much',
  'kpi',
  'margin',
  'metric',
  'owner brief',
  'report',
  'revenue',
  'sales',
  'show my top',
  'slow mover',
  'summarize',
  'total',
  'trend',
  'what changed',
];

const SUPPORT_PHRASES = [
  'can i find',
  'explain this page',
  'help me find',
  'how can i',
  'how do i',
  'how does',
  'show me where',
  'what does this mean',
  'what is this',
  'where can i',
  'where do i',
  'why can',
];

function normalizePrompt(value: string): string {
  return ` ${value
    .toLowerCase()
    .replace(/[^a-z0-9$%]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()} `;
}

function containsAny(text: string, phrases: string[]): boolean {
  return phrases.some((phrase) => text.includes(` ${phrase.trim()} `));
}

function looksLikeAnalytics(text: string): boolean {
  if (containsAny(text, ANALYST_PHRASES)) return true;
  return (
    /\b(top|bottom)\s+\d+\b/.test(text) ||
    /\b(this|last|next)\s+(day|week|month|quarter|year)\b/.test(text)
  );
}

function looksLikeAction(text: string): boolean {
  if (containsAny(text, AGENT_PHRASES)) return true;
  return /\b(can you|please)\s+(add|approve|change|create|record|schedule|send|set|update)\b/.test(
    text
  );
}

function looksLikeSupport(text: string): boolean {
  return containsAny(text, SUPPORT_PHRASES);
}

export function inferCaretModeForPrompt(
  prompt: string,
  currentMode: CaretMode = DEFAULT_MODE
): CaretMode {
  const text = normalizePrompt(prompt);
  if (!text.trim()) return currentMode;

  if (looksLikeAction(text)) return 'agent';
  if (looksLikeAnalytics(text)) return 'analyst';
  if (looksLikeSupport(text)) return 'support';

  return currentMode;
}
