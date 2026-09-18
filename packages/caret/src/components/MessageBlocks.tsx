import React from 'react';
import Markdown from './blocks/Markdown';
import ThinkingBlock from './blocks/ThinkingBlock';
import ToolCallCard from './blocks/ToolCallCard';
import type { ContentBlock } from '../types/message';

export interface MessageBlocksProps {
  blocks: ContentBlock[];
  onApprove?: (toolCallId: string) => void;
  onDeny?: (toolCallId: string) => void;
  showThinking?: boolean;
}

/** Render a heterogeneous array of content blocks (text + thinking + tools). */
export const MessageBlocks: React.FC<MessageBlocksProps> = ({
  blocks,
  onApprove,
  onDeny,
  showThinking = true,
}) => {
  return (
    <>
      {blocks.map((b) => {
        if (b.type === 'text') return <Markdown key={b.id} source={b.content} />;
        if (b.type === 'thinking') {
          if (!showThinking) return null;
          return <ThinkingBlock key={b.id} content={b.content} durationMs={b.durationMs} />;
        }
        return <ToolCallCard key={b.id} call={b} onApprove={onApprove} onDeny={onDeny} />;
      })}
    </>
  );
};

export default MessageBlocks;
