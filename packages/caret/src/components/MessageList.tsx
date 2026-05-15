import React from 'react';
import MessageBubble from './MessageBubble';
import { useScrollToBottom } from '../hooks/useScrollToBottom';
import type { CaretMessage } from '../types/message';

export interface MessageListProps {
  messages: CaretMessage[];
  isStreaming: boolean;
  showThinking?: boolean;
  onApproveTool?: (id: string) => void;
  onDenyTool?: (id: string) => void;
}

export const MessageList: React.FC<MessageListProps> = ({
  messages,
  isStreaming,
  showThinking,
  onApproveTool,
  onDenyTool,
}) => {
  const scrollRef = useScrollToBottom<HTMLDivElement>([messages.length, isStreaming]);

  return (
    <div
      ref={scrollRef}
      style={{
        flex: 1,
        overflowY: 'auto',
        padding: '16px',
      }}
      aria-live="polite"
      aria-atomic="false"
    >
      {messages.map((msg, i) => {
        const isLastAssistant =
          i === messages.length - 1 && msg.role === 'assistant' && isStreaming;
        return (
          <MessageBubble
            key={msg.id}
            message={msg}
            isStreaming={isLastAssistant}
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
