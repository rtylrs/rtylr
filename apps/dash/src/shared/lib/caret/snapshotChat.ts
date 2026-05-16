import type { CaretMessageType } from '@voxire/caret';
import { messagePlainText } from './messageMeta';

/** Build a serialisable, compact representation of a chat for diagnostics. */
export interface ChatSnapshot {
  sessionId: string | null;
  messageCount: number;
  preview: string;
  capturedAt: string;
}

export function snapshotChat(
  sessionId: string | null,
  messages: CaretMessageType[],
  maxPreview = 320
): ChatSnapshot {
  const last = messages[messages.length - 1];
  const text = last ? messagePlainText(last) : '';
  const preview = text.length > maxPreview ? text.slice(0, maxPreview - 1) + '…' : text;
  return {
    sessionId,
    messageCount: messages.length,
    preview,
    capturedAt: new Date().toISOString(),
  };
}
