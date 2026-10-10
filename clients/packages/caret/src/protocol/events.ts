/**
 * Structured event stream emitted by the Caret backend over SSE. The legacy
 * stream emitted only `{text, done}` chunks; the new protocol is a tagged
 * union so we can support thinking blocks, tool calls, and per-message
 * metadata without breaking older clients.
 */

import type { CaretMode } from '../types/mode';

export type CaretStreamEvent =
  | { type: 'session'; sessionId: string }
  | { type: 'message_start'; messageId: string; role: 'assistant' }
  | { type: 'text_delta'; messageId: string; delta: string }
  | { type: 'thinking_start'; messageId: string; blockId: string }
  | { type: 'thinking_delta'; messageId: string; blockId: string; delta: string }
  | { type: 'thinking_end'; messageId: string; blockId: string; durationMs?: number }
  | {
      type: 'tool_call';
      messageId: string;
      blockId: string;
      name: string;
      mode?: CaretMode;
      input: Record<string, unknown>;
      preview?: string;
      requiresApproval?: boolean;
    }
  | {
      type: 'tool_result';
      messageId: string;
      blockId: string;
      status: 'success' | 'error';
      output?: unknown;
      summary?: string;
      error?: string;
    }
  | { type: 'message_end'; messageId: string; stored_id?: string }
  | { type: 'follow_ups'; messageId: string; items: string[] }
  | { type: 'error'; error: string; recoverable?: boolean }
  | { type: 'done' };

/** Legacy compatibility shape that the original stream produced. */
export interface LegacyStreamToken {
  text: string;
  done: boolean;
}

/** Try to parse a JSON line into a CaretStreamEvent.
 *  Falls back to a `text_delta` envelope so we stay compatible with the old
 *  `{text, done}` server output during the rollout window. */
export function parseStreamLine(line: string, fallbackMessageId: string): CaretStreamEvent | null {
  const trimmed = line.trim();
  if (!trimmed) return null;
  try {
    const obj = JSON.parse(trimmed) as Record<string, unknown>;
    if (typeof obj.type === 'string') {
      return obj as CaretStreamEvent;
    }
    // Legacy: { text, done }
    if (obj.done === true) return { type: 'done' };
    if (typeof obj.text === 'string') {
      return { type: 'text_delta', messageId: fallbackMessageId, delta: obj.text };
    }
    return null;
  } catch {
    return null;
  }
}
