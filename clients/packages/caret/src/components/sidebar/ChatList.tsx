import React, { useEffect, useMemo, useRef, useState } from "react";
import { caretTokens, caretRadius } from "../../ui/tokens";
import { useFocusTrap } from "../../hooks/useFocusTrap";
import ChatListItem from "./ChatListItem";
import NewChatButton from "./NewChatButton";
import type { CaretChatSummary } from "../../types/chat";

export interface ChatListLabels {
  heading: string;
  description: string;
  newChat: string;
  loading: string;
  empty: string;
  loadFailed: string;
  retry: string;
  deleteChat: string;
  deleteTitle: string;
  deleteMessage: string;
  confirmDelete: string;
  cancel: string;
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
  /** Hide the list's own button when the host places New chat elsewhere. */
  showNewChat?: boolean;
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
  showNewChat = true,
}) => {
  const [deleting, setDeleting] = useState<CaretChatSummary | null>(null);
  const [busy, setBusy] = useState(false);
  const dialogRef = useRef<HTMLDivElement>(null);
  useFocusTrap(dialogRef, deleting !== null);

  const closeDelete = (): void => {
    if (!busy) setDeleting(null);
  };
  const confirmDelete = async (): Promise<void> => {
    if (!deleting || !onDelete) return;
    setBusy(true);
    try {
      await onDelete(deleting.id);
    } finally {
      setBusy(false);
      setDeleting(null);
    }
  };
  useEffect(() => {
    if (deleting === null) return;
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape") closeDelete();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  });

  // Searching chats lives in the command palette, so this list only orders them.
  const filtered = useMemo(
    () =>
      [...chats].sort((a, b) => {
        if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
        return b.updatedAt.localeCompare(a.updatedAt);
      }),
    [chats],
  );

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
      {showNewChat && (
        <NewChatButton onClick={onNew} label={labels?.newChat} />
      )}
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
                {labels?.empty ?? "No saved chats yet."}
              </p>
            )}
            {filtered.map((c) => (
              <ChatListItem
                key={c.id}
                chat={c}
                labels={labels}
                active={c.id === activeId}
                onSelect={onSelect}
                onDelete={onDelete ? () => setDeleting(c) : undefined}
                onTogglePin={onTogglePin}
              />
            ))}
          </>
        )}
      </div>
      {deleting && (
        <div
          className="caret-delete-backdrop"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) closeDelete();
          }}
          style={{
            position: "fixed",
            inset: 0,
            zIndex: 10001,
            display: "grid",
            placeItems: "center",
            padding: 18,
            background: "rgba(0,0,0,.38)",
          }}
        >
          <div
            ref={dialogRef}
            role="alertdialog"
            aria-modal="true"
            aria-labelledby="caret-delete-dialog-title"
            aria-describedby="caret-delete-dialog-message"
            style={{
              width: "min(420px, 100%)",
              background: caretTokens.bg,
              border: `1px solid ${caretTokens.border}`,
              borderRadius: caretRadius.lg,
              padding: 20,
              boxSizing: "border-box",
              color: caretTokens.text,
              boxShadow: caretTokens.shadow,
            }}
          >
            <h2
              id="caret-delete-dialog-title"
              style={{ margin: "0 0 9px", fontSize: 18 }}
            >
              {labels?.deleteTitle ?? "Delete chat?"}
            </h2>
            <p
              id="caret-delete-dialog-message"
              style={{
                margin: "0 0 8px",
                color: caretTokens.textMuted,
                fontSize: 13,
                lineHeight: 1.5,
              }}
            >
              {labels?.deleteMessage ??
                "This chat and its messages will be permanently deleted. This can't be undone."}
            </p>
            <p
              style={{
                margin: 0,
                fontSize: 13,
                fontWeight: 600,
                overflowWrap: "anywhere",
              }}
            >
              {deleting.title || labels?.newChat || "New chat"}
            </p>
            <div
              style={{
                display: "flex",
                justifyContent: "flex-end",
                gap: 8,
                marginTop: 18,
              }}
            >
              <button
                type="button"
                autoFocus
                disabled={busy}
                onClick={closeDelete}
                style={{
                  border: `1px solid ${caretTokens.border}`,
                  borderRadius: caretRadius.sm,
                  padding: "8px 12px",
                  background: "transparent",
                  color: caretTokens.text,
                  cursor: "pointer",
                  fontSize: 12.5,
                }}
              >
                {labels?.cancel ?? "Cancel"}
              </button>
              <button
                type="button"
                disabled={busy}
                onClick={() => {
                  void confirmDelete();
                }}
                style={{
                  border: 0,
                  borderRadius: caretRadius.sm,
                  padding: "8px 12px",
                  background: "#b42318",
                  color: "#fff",
                  cursor: "pointer",
                  fontSize: 12.5,
                  fontWeight: 600,
                }}
              >
                {labels?.confirmDelete ?? "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ChatList;
