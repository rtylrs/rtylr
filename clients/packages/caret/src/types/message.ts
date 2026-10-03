import type { CaretMode } from "./mode";

export type MessageRole = "user" | "assistant" | "system" | "tool";

export interface ThinkingBlock {
  id: string;
  type: "thinking";
  content: string;
  durationMs?: number;
}

export interface TextBlock {
  id: string;
  type: "text";
  content: string;
}

export interface ToolCallBlock {
  id: string;
  type: "tool_call";
  name: string;
  mode?: CaretMode;
  input: Record<string, unknown>;
  status: "pending" | "approved" | "denied" | "running" | "success" | "error";
  output?: unknown;
  preview?: string;
  summary?: string;
  error?: string;
  requiresApproval?: boolean;
}

export type ContentBlock = TextBlock | ThinkingBlock | ToolCallBlock;

export interface CaretImageAttachment {
  name: string;
  dataUrl: string;
}

export interface CaretMessage {
  id: string;
  role: MessageRole;
  blocks: ContentBlock[];
  createdAt: string;
  /** Legacy/flattened content used when blocks aren't populated yet. */
  content?: string;
  attachments?: CaretImageAttachment[];
}
