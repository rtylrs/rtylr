import React from "react";
import { caretTokens } from "../ui/tokens";
import MessageBlocks from "./MessageBlocks";
import MessageActions from "./MessageActions";
import { StreamingOrb } from "./CaretOrb";
import type { CaretMessage, ContentBlock } from "../types/message";

export interface MessageBubbleProps {
  message: CaretMessage;
  isStreaming?: boolean;
  showThinking?: boolean;
  onApproveTool?: (id: string) => void;
  onDenyTool?: (id: string) => void;
  onRetry?: () => void;
}

/** Flatten content blocks to plain text for copy. */
function blocksToText(blocks: ContentBlock[]): string {
  return blocks
    .map((b) => {
      if (b.type === "text") return b.content;
      if (
        b.type === "tool_call" &&
        b.name === "web.search" &&
        b.output &&
        typeof b.output === "object"
      ) {
        const answer = (b.output as Record<string, unknown>).answer;
        return typeof answer === "string" ? answer : "";
      }
      return "";
    })
    .filter(Boolean)
    .join("\n");
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({
  message,
  isStreaming = false,
  showThinking = true,
  onApproveTool,
  onDenyTool,
  onRetry,
}) => {
  const isUser = message.role === "user";
  const blocks: ContentBlock[] =
    message.blocks.length > 0
      ? message.blocks
      : message.content
        ? [{ id: `${message.id}-text`, type: "text", content: message.content }]
        : [];

  const text = isUser
    ? (message.content ?? blocksToText(blocks))
    : blocksToText(blocks);
  const hasAssistantText =
    blocks.some((block) => block.type === "text") ||
    blocks.some(
      (block) =>
        block.type === "tool_call" &&
        block.name === "web.search" &&
        Boolean(blocksToText([block])),
    );

  // User message - quiet, right-aligned bubble.
  if (isUser) {
    return (
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "flex-end",
          gap: 8,
          minWidth: 0,
          margin: "12px 0",
        }}
      >
        {message.attachments?.map((attachment, index) =>
          attachment.dataUrl.startsWith("data:image/") ? (
            <img
              key={`${attachment.name}-${index}`}
              src={attachment.dataUrl}
              alt={attachment.name || `Attached image ${index + 1}`}
              style={{
                display: "block",
                width: "auto",
                height: "auto",
                maxWidth: "min(320px, 82%)",
                maxHeight: 280,
                objectFit: "contain",
                borderRadius: 10,
              }}
            />
          ) : null,
        )}
        {text && (
          <div
            style={{
              maxWidth: "82%",
              padding: "10px 14px",
              borderRadius: "16px 16px 4px 16px",
              background: caretTokens.accentSoft,
              color: caretTokens.text,
              fontSize: 13.5,
              lineHeight: 1.55,
              whiteSpace: "pre-wrap",
              wordBreak: "break-word",
            }}
          >
            {text}
          </div>
        )}
      </div>
    );
  }

  // Assistant message - full-width plain text, no avatar, no bubble (ChatGPT-style).
  return (
    <div className="caret-assistant-message" style={{ margin: "12px 0" }}>
      <div
        style={{
          color: caretTokens.text,
          fontSize: 14,
          lineHeight: 1.6,
        }}
      >
        <MessageBlocks
          blocks={blocks}
          onApprove={onApproveTool}
          onDeny={onDenyTool}
          showThinking={showThinking}
        />
        {isStreaming && <StreamingOrb hasText={text.trim().length > 0} />}
      </div>
      {!isStreaming && hasAssistantText && (
        <MessageActions
          text={text}
          onRetry={onRetry}
          responseMs={message.responseMs}
        />
      )}
    </div>
  );
};

export default MessageBubble;
