/**
 * Re-export of caret types under a dash-namespaced path so app modules
 * don't need to import directly from @voxire/caret in every file.
 */

export type {
  CaretMessageType,
  ContentBlock,
  TextBlock,
  ThinkingBlock,
  ToolCallBlock,
  CaretChat,
  CaretChatSummary,
  CaretMode,
  CaretModel,
} from '@voxire/caret';
