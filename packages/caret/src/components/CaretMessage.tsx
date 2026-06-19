import React from 'react';
import MessageBlocks from './MessageBlocks';
import StreamingDot from './blocks/StreamingDot';
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
      return `[tool:${block.name}]`;
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
          borderRadius: isUser ? '18px 18px 4px 18px' : '18px 18px 18px 4px',
          background: isUser ? 'var(--caret-accent, #8b5cf6)' : 'var(--caret-card, #1e1e1e)',
          color: '#ffffff',
          border: isUser ? 'none' : '1px solid var(--caret-border, #2a2a2a)',
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
