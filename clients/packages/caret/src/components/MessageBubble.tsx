import React from 'react';
import { caretTokens } from '../ui/tokens';
import MessageBlocks from './MessageBlocks';
import MessageActions from './MessageActions';
import FollowUpChips from './FollowUpChips';
import { StreamingOrb } from './CaretOrb';
import type { CaretStarter } from '../lib/suggestions';
import type { CaretMessage, ContentBlock } from '../types/message';

export interface MessageBubbleProps {
  message: CaretMessage;
  isStreaming?: boolean;
  showThinking?: boolean;
  onApproveTool?: (id: string) => void;
  onDenyTool?: (id: string) => void;
  onRetry?: () => void;
  followUps?: CaretStarter[];
  onFollowUp?: (item: CaretStarter) => void;
}

/** Flatten content blocks to plain text for copy. */
function blocksToText(blocks: ContentBlock[]): string {
  return blocks
    .map((b) => {
      if (b.type === 'text') return b.content;
      if (b.type === 'thinking') return '';
      return '[business check]';
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
  followUps = [],
  onFollowUp,
}) => {
  const isUser = message.role === 'user';
  const blocks: ContentBlock[] =
    message.blocks.length > 0
      ? message.blocks
      : message.content
        ? [{ id: `${message.id}-text`, type: 'text', content: message.content }]
        : [];

  const text = isUser ? (message.content ?? blocksToText(blocks)) : blocksToText(blocks);

  // User message - quiet, right-aligned bubble.
  if (isUser) {
    return (
      <div
        style={{
          display: 'flex',
          justifyContent: 'flex-end',
          margin: '12px 0',
        }}
      >
        <div
          style={{
            maxWidth: '82%',
            padding: '10px 14px',
            borderRadius: '16px 16px 4px 16px',
            background: caretTokens.accentSoft,
            color: caretTokens.text,
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

  // Assistant message - full-width plain text, no avatar, no bubble (ChatGPT-style).
  return (
    <div className="caret-assistant-message" style={{ margin: '12px 0' }}>
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
      {!isStreaming && blocks.some((b) => b.type === 'text') && (
        <MessageActions text={text} onRetry={onRetry} />
      )}
      {!isStreaming && <FollowUpChips items={followUps} onSelect={onFollowUp} />}
    </div>
  );
};

export default MessageBubble;
