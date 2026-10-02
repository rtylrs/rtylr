import React, { useMemo, useState } from "react";
import { caretTokens } from "../../ui/tokens";
import ChatListItem from "./ChatListItem";
import ChatSearch from "./ChatSearch";
import NewChatButton from "./NewChatButton";
import { useDebouncedValue } from "../../hooks/useDebouncedValue";
import type { CaretChatSummary } from "../../types/chat";

export interface ChatListLabels {
  heading: string;
  description: string;
  newChat: string;
  search: string;
  clearSearch: string;
  loading: string;
  empty: string;
  noResults: string;
  loadFailed: string;
  retry: string;
  deleteChat: string;
  pinChat: string;
  unpinChat: string;
}

export interface ChatListProps {
  chats: CaretChatSummary[];
  labels?: Partial<ChatListLabels>;
  isLoading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
  activeId?: string | null;
  onSelect: (id: string) => void;
  onNew: () => void;
  onDelete?: (id: string) => void | Promise<void>;
  onTogglePin?: (id: string) => void;
}

export const ChatList: React.FC<ChatListProps> = ({
  chats,
  labels,
  isLoading = false,
  error = null,
  onRetry,
  activeId,
  onSelect,
  onNew,
  onDelete,
  onTogglePin,
}) => {
  const [query, setQuery] = useState("");
  const debounced = useDebouncedValue(query, 150);

  const filtered = useMemo(() => {
    const q = debounced.toLowerCase().trim();
    const ordered = [...chats].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
      return b.updatedAt.localeCompare(a.updatedAt);
    });
    if (!q) return ordered;
    return ordered.filter(
      (c) =>
        c.title.toLowerCase().includes(q) ||
        (c.summary ?? "").toLowerCase().includes(q),
    );
  }, [chats, debounced]);

  return (
    <div
      className="caret-chat-list"
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 9,
        padding: "18px 14px 14px",
        flex: 1,
        minHeight: 0,
        overflow: "hidden",
        boxSizing: "border-box",
      }}
    >
      <div className="caret-chat-list-heading" style={{ margin: "0 2px 7px" }}>
        <h2
          style={{
            margin: 0,
            color: caretTokens.text,
            fontSize: 15,
            fontWeight: 650,
            lineHeight: 1.4,
          }}
        >
          {labels?.heading ?? "History"}
        </h2>
        {labels?.description !== "" && (
          <p
            style={{
              margin: "3px 0 0",
              color: caretTokens.textMuted,
              fontSize: 11.5,
            }}
          >
            {labels?.description ?? "Open a past chat or start another."}
          </p>
        )}
      </div>
      <NewChatButton onClick={onNew} label={labels?.newChat} />
      <ChatSearch
        value={query}
        onChange={setQuery}
        label={labels?.search}
        clearLabel={labels?.clearSearch}
      />
      <div
        className="caret-chat-list-scroll"
        style={{
          display: "flex",
          flex: 1,
          flexDirection: "column",
          gap: 2,
          minHeight: 0,
          overflowY: "auto",
          paddingRight: 2,
        }}
      >
        {error !== null ? (
          <div
            role="alert"
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "flex-start",
              gap: 8,
              padding: "8px 4px",
            }}
          >
            <p
              style={{
                fontSize: 12,
                color: caretTokens.textDim,
                margin: 0,
              }}
            >
              {labels?.loadFailed ?? "Unable to load saved chats."}
            </p>
            {onRetry && (
              <button
                type="button"
                onClick={onRetry}
                style={{
                  border: `1px solid ${caretTokens.border}`,
                  borderRadius: 6,
                  background: caretTokens.surface,
                  color: caretTokens.text,
                  cursor: "pointer",
                  fontSize: 12,
                  padding: "5px 9px",
                }}
              >
                {labels?.retry ?? "Try again"}
              </button>
            )}
          </div>
        ) : isLoading ? (
          <p
            role="status"
            style={{
              fontSize: 12,
              color: caretTokens.textDim,
              padding: "8px 4px",
              margin: 0,
            }}
          >
            {labels?.loading ?? "Loading saved chats…"}
          </p>
        ) : (
          <>
            {filtered.length === 0 && (
              <p
                style={{
                  fontSize: 12,
                  color: caretTokens.textDim,
                  padding: "8px 4px",
                  margin: 0,
                }}
              >
                {query
                  ? (labels?.noResults ?? "No chats match your search.")
                  : (labels?.empty ?? "No saved chats yet.")}
              </p>
            )}
            {filtered.map((c) => (
              <ChatListItem
                key={c.id}
                chat={c}
                labels={labels}
                active={c.id === activeId}
                onSelect={onSelect}
                onDelete={onDelete}
                onTogglePin={onTogglePin}
              />
            ))}
          </>
        )}
      </div>
    </div>
  );
};

export default ChatList;
