import React from "react";
import Markdown from "./blocks/Markdown";
import ThinkingBlock from "./blocks/ThinkingBlock";
import ToolCallCard from "./blocks/ToolCallCard";
import { useCaretStore } from "../stores/caretStore";
import type { ContentBlock } from "../types/message";

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
  // The store marks the open thought until thinking_end; match its id so older
  // thoughts in the same assistant turn stay rendered as completed.
  const liveThinkingBlockId = useCaretStore((state) => {
    const tab = state.tabs.find(
      (candidate) => candidate.id === state.activeTabId,
    );
    if (!tab?.isStreaming) return undefined;

    const message = tab.messages[tab.messages.length - 1];
    if (message?.role !== "assistant") return undefined;

    const block = message.blocks[message.blocks.length - 1];
    return block?.type === "thinking" && block.durationMs === undefined
      ? block.id
      : undefined;
  });

  return (
    <>
      {blocks.map((b) => {
        if (b.type === "text")
          return <Markdown key={b.id} source={b.content} />;
        if (b.type === "thinking") {
          if (!showThinking) return null;
          return (
            <ThinkingBlock
              key={b.id}
              content={b.content}
              durationMs={b.durationMs}
              isLive={liveThinkingBlockId === b.id}
            />
          );
        }
        return (
          <ToolCallCard
            key={b.id}
            call={b}
            onApprove={onApprove}
            onDeny={onDeny}
          />
        );
      })}
    </>
  );
};

export default MessageBlocks;
