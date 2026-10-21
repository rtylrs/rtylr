import React, { useRef, useState } from "react";
import {
  MoreHorizontal,
  Pencil,
  Pin,
  PinOff,
  Trash2,
} from "lucide-react";
import { caretTokens, caretRadius } from "../../ui/tokens";
import IconButton from "../../ui/IconButton";
import ChatActionsMenu, { type ChatMenuEntry } from "./ChatActionsMenu";
import type { ChatListLabels } from "./ChatList";
import type { CaretChatSummary } from "../../types/chat";

export interface ChatListItemProps {
  chat: CaretChatSummary;
  labels?: Partial<ChatListLabels>;
  active?: boolean;
  onSelect: (id: string) => void;
  onTogglePin?: (id: string) => void;
  onRename?: (id: string) => void;
  onDelete?: (id: string) => void | Promise<void>;
}

export const ChatListItem: React.FC<ChatListItemProps> = ({
  chat,
  labels,
  active = false,
  onSelect,
  onTogglePin,
  onRename,
  onDelete,
}) => {
  const rowRef = useRef<HTMLDivElement>(null);
  const [menu, setMenu] = useState<{
    x: number;
    y: number;
    alignEnd: boolean;
  } | null>(null);

  const entries: ChatMenuEntry[] = [];
  if (onRename)
    entries.push({
      id: "rename",
      label: labels?.renameChat ?? "Rename",
      icon: <Pencil size={15} />,
      onSelect: () => onRename(chat.id),
    });
  if (onTogglePin)
    entries.push({
      id: "pin",
      label: chat.pinned
        ? (labels?.unpinChat ?? "Unpin chat")
        : (labels?.pinChat ?? "Pin chat"),
      icon: chat.pinned ? <PinOff size={15} /> : <Pin size={15} />,
      onSelect: () => onTogglePin(chat.id),
    });
  if (onDelete) {
    if (entries.length > 0) entries.push({ id: "divider", separator: true });
    entries.push({
      id: "delete",
      label: labels?.deleteChat ?? "Delete chat",
      icon: <Trash2 size={15} />,
      destructive: true,
      onSelect: () => {
        void onDelete(chat.id);
      },
    });
  }
  const shell =
    rowRef.current?.closest<HTMLElement>(".caret-panel-shell") ?? document.body;

  return (
    <div
      ref={rowRef}
      className={
        "caret-chat-row" +
        (active ? " is-active" : "") +
        (menu ? " is-menu-open" : "")
      }
      onContextMenu={(event) => {
        if (entries.length === 0) return;
        event.preventDefault();
        setMenu({ x: event.clientX, y: event.clientY, alignEnd: false });
      }}
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
        onDoubleClick={onRename ? () => onRename(chat.id) : undefined}
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
      {chat.pinned && (
        <Pin
          size={11}
          aria-label={labels?.pinned ?? "Pinned"}
          style={{ flex: "0 0 auto", margin: "0 4px", color: caretTokens.textDim }}
        />
      )}
      {entries.length > 0 && (
        <div className="caret-chat-actions" style={{ display: "flex", gap: 2 }}>
          <IconButton
            size="sm"
            label={labels?.chatActions ?? "Chat actions"}
            aria-haspopup="menu"
            aria-expanded={menu !== null}
            onClick={(event) => {
              const rect = event.currentTarget.getBoundingClientRect();
              setMenu({ x: rect.right, y: rect.bottom + 4, alignEnd: true });
            }}
          >
            <MoreHorizontal size={14} />
          </IconButton>
        </div>
      )}
      {menu && (
        <ChatActionsMenu
          x={menu.x}
          y={menu.y}
          alignEnd={menu.alignEnd}
          label={labels?.chatActions ?? "Chat actions"}
          container={shell}
          entries={entries}
          onClose={() => setMenu(null)}
        />
      )}
    </div>
  );
};

export default ChatListItem;
