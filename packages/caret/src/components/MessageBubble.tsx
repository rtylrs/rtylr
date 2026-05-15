import React from 'react';
import { caretTokens, caretRadius } from '../ui/tokens';
import Avatar from '../ui/Avatar';
import MessageBlocks from './MessageBlocks';
import MessageActions from './MessageActions';
import StreamingDot from './blocks/StreamingDot';
import type { CaretMessage, ContentBlock } from '../types/message';

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
      if (b.type === 'text') return b.content;
      if (b.type === 'thinking') return '';
      return `[tool:${b.name}]`;
    })
    .filter(Boolean)
    .join('\n');
}

export const MessageBubble: React.FC<MessageBubbleProps> = ({
  message,
  isStreaming = false,
  showThinking = true,
  onApproveTool,
  onDenyTool,
  onRetry,
}) => {
  const isUser = message.role === 'user';
  const blocks: ContentBlock[] =
    message.blocks.length > 0
      ? message.blocks
      : message.content
        ? [{ id: `${message.id}-text`, type: 'text', content: message.content }]
        : [];

  const text = isUser ? (message.content ?? blocksToText(blocks)) : blocksToText(blocks);

  if (isUser) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'flex-end',
          marginBottom: 10,
        }}
      >
        <div
          style={{
            maxWidth: '85%',
            padding: '8px 12px',
            borderRadius: caretRadius.md,
            background: caretTokens.accent,
            color: '#ffffff',
            fontSize: 13.5,
            lineHeight: 1.55,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          {text}
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        display: 'flex',
        gap: 8,
        marginBottom: 12,
      }}
    >
      <Avatar role="assistant" initials="C" />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div
          style={{
            color: caretTokens.text,
            fontSize: 13.5,
            lineHeight: 1.55,
          }}
        >
          <MessageBlocks
            blocks={blocks}
            onApprove={onApproveTool}
            onDeny={onDenyTool}
            showThinking={showThinking}
          />
          {isStreaming && <StreamingDot />}
        </div>
        {!isStreaming && blocks.some((b) => b.type === 'text') && (
          <MessageActions text={text} onRetry={onRetry} />
        )}
      </div>
    </div>
  );
};

export default MessageBubble;
