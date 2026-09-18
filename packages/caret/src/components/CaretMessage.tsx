import React from 'react';
import MessageBlocks from './MessageBlocks';
import StreamingDot from './blocks/StreamingDot';
import { caretTokens } from '../ui/tokens';
import type { CaretMessage as CaretMessageType } from '../stores/caretStore';
import type { ContentBlock } from '../types/message';

interface CaretMessageProps {
  message: CaretMessageType;
  isStreaming?: boolean;
}

function messageBlocks(message: CaretMessageType): ContentBlock[] {
  if (message.blocks.length > 0) return message.blocks;
  if (!message.content) return [];
  return [{ id: `${message.id}-text`, type: 'text', content: message.content }];
}

function blocksToText(blocks: ContentBlock[]): string {
  return blocks
    .map((block) => {
      if (block.type === 'text') return block.content;
      if (block.type === 'thinking') return '';
      return '[business check]';
    })
    .filter(Boolean)
    .join('\n');
}

const CaretMessage = ({ message, isStreaming = false }: CaretMessageProps): React.JSX.Element => {
  const isUser = message.role === 'user';
  const blocks = messageBlocks(message);
  const text = message.content ?? blocksToText(blocks);

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: isUser ? 'flex-end' : 'flex-start',
        marginBottom: '12px',
      }}
    >
      <div
        style={{
          maxWidth: '80%',
          padding: '10px 14px',
          borderRadius: 4,
          background: isUser ? caretTokens.surfaceHigh : 'transparent',
          color: caretTokens.text,
          border: isUser ? `1px solid ${caretTokens.border}` : 'none',
          fontSize: '14px',
          lineHeight: '1.6',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
        }}
      >
        {isUser ? text : <MessageBlocks blocks={blocks} />}
        {isStreaming && <StreamingDot />}
      </div>
    </div>
  );
};

export default CaretMessage;
