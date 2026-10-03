import React from "react";
import MessageBubble from "./MessageBubble";
import { useScrollToBottom } from "../hooks/useScrollToBottom";
import type { CaretMessage } from "../types/message";

export interface MessageListProps {
  messages: CaretMessage[];
  isStreaming: boolean;
  scrollKey?: string;
  showThinking?: boolean;
  onApproveTool?: (id: string) => void;
  onDenyTool?: (id: string) => void;
}

export const MessageList: React.FC<MessageListProps> = ({
  messages,
  isStreaming,
  scrollKey,
  showThinking,
  onApproveTool,
  onDenyTool,
}) => {
  const scrollRef = useScrollToBottom<HTMLDivElement>(
    [messages.length, isStreaming],
    scrollKey,
  );

  return (
    <div
      className="caret-message-list"
      ref={scrollRef}
      style={{
        flex: 1,
        overflowY: "auto",
        padding: "16px",
      }}
      aria-live="polite"
      aria-atomic="false"
    >
      {messages.map((msg, i) => {
        const isLastAssistant =
          i === messages.length - 1 && msg.role === "assistant";
        return (
          <MessageBubble
            key={msg.id}
            message={msg}
            isStreaming={isLastAssistant && isStreaming}
            showThinking={showThinking}
            onApproveTool={onApproveTool}
            onDenyTool={onDenyTool}
          />
        );
      })}
    </div>
  );
};

export default MessageList;
