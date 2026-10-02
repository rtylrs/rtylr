import React from "react";
import { Trash2, Pin, PinOff } from "lucide-react";
import { caretTokens, caretRadius } from "../../ui/tokens";
import IconButton from "../../ui/IconButton";
import type { ChatListLabels } from "./ChatList";
import type { CaretChatSummary } from "../../types/chat";

export interface ChatListItemProps {
  chat: CaretChatSummary;
  labels?: Partial<ChatListLabels>;
  active?: boolean;
  onSelect: (id: string) => void;
  onTogglePin?: (id: string) => void;
  onDelete?: (id: string) => void | Promise<void>;
}

export const ChatListItem: React.FC<ChatListItemProps> = ({
  chat,
  labels,
  active = false,
  onSelect,
  onTogglePin,
  onDelete,
}) => {
  return (
    <div
      className={active ? "caret-chat-row is-active" : "caret-chat-row"}
      style={{
        display: "flex",
        alignItems: "center",
        minHeight: 36,
        padding: "0 4px 0 0",
        borderRadius: caretRadius.sm,
        background: active ? caretTokens.surfaceHigh : "transparent",
        color: active ? caretTokens.text : caretTokens.textMuted,
        position: "relative",
      }}
    >
      <button
        type="button"
        onClick={() => onSelect(chat.id)}
        style={{
          display: "flex",
          alignItems: "center",
          gap: 0,
          flex: 1,
          minWidth: 0,
          minHeight: 36,
          padding: "7px 6px 7px 10px",
          background: "transparent",
          border: "none",
          borderRadius: caretRadius.sm,
          color: "inherit",
          cursor: "pointer",
          textAlign: "start",
        }}
      >
        <span
          style={{
            flex: 1,
            minWidth: 0,
            fontSize: 13,
            whiteSpace: "nowrap",
            overflow: "hidden",
            textOverflow: "ellipsis",
          }}
        >
          {chat.title || labels?.newChat || "New chat"}
        </span>
      </button>
      <div className="caret-chat-actions" style={{ display: "flex", gap: 2 }}>
        {onTogglePin && (
          <IconButton
            size="sm"
            label={
              chat.pinned
                ? (labels?.unpinChat ?? "Unpin chat")
                : (labels?.pinChat ?? "Pin chat")
            }
            onClick={() => onTogglePin(chat.id)}
          >
            {chat.pinned ? <PinOff size={11} /> : <Pin size={11} />}
          </IconButton>
        )}
        {onDelete && (
          <IconButton
            size="sm"
            label={labels?.deleteChat ?? "Delete chat"}
            onClick={() => {
              void onDelete(chat.id);
            }}
          >
            <Trash2 size={11} />
          </IconButton>
        )}
      </div>
    </div>
  );
};

export default ChatListItem;
