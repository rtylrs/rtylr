import type { CaretChatSummary } from '@voxire/caret';

/** Sort chats so pinned items float to the top, then by most-recent activity. */
export function sortByRecency(chats: CaretChatSummary[]): CaretChatSummary[] {
  return [...chats].sort((a, b) => {
    if (Boolean(a.pinned) !== Boolean(b.pinned)) return a.pinned ? -1 : 1;
    return b.updatedAt.localeCompare(a.updatedAt);
  });
}
