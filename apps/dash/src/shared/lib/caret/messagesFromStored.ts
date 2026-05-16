import type { CaretMessageType } from '@voxire/caret';
import type { StoredMessage } from '../../api/caret';

/** Convert backend StoredMessage shape into the in-memory Caret message. */
export function fromStored(m: StoredMessage): CaretMessageType {
  return {
    id: m.id,
    role: m.role,
    blocks: [],
    content: m.content,
    createdAt: m.createdAt,
  };
}

export function fromStoredMany(messages: StoredMessage[]): CaretMessageType[] {
  return messages.map(fromStored);
}
