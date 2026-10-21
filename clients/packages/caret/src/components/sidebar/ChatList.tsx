import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { X } from "lucide-react";
import { caretTokens, caretRadius } from "../../ui/tokens";
import ChatDialog from "./ChatDialog";
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
  chatActions: string;
  renameChat: string;
  renameTitle: string;
  renameHint: string;
  renameFailed: string;
  save: string;
  close: string;
  pinned: string;
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
  /** Rejects when the rename fails; the dialog then stays open with an error. */
  onRename?: (id: string, title: string) => void | Promise<void>;
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
  onRename,
  onTogglePin,
  showNewChat = true,
}) => {
  const [deleting, setDeleting] = useState<CaretChatSummary | null>(null);
  const [renaming, setRenaming] = useState<CaretChatSummary | null>(null);
  const [draft, setDraft] = useState("");
  const [renameError, setRenameError] = useState("");
  const [busy, setBusy] = useState(false);
  const renameInputRef = useRef<HTMLInputElement>(null);

  const closeDialogs = useCallback((): void => {
    setDeleting(null);
    setRenaming(null);
    setRenameError("");
  }, []);
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
  const startRename = (chat: CaretChatSummary): void => {
    setDraft(chat.title || "");
    setRenameError("");
    setRenaming(chat);
  };
  useEffect(() => {
    if (renaming) {
      renameInputRef.current?.focus();
      renameInputRef.current?.select();
    }
  }, [renaming]);
  const trimmedDraft = draft.trim();
  const submitRename = async (): Promise<void> => {
    if (!renaming || !onRename || !trimmedDraft) return;
    if (trimmedDraft === renaming.title) {
      closeDialogs();
      return;
    }
    setBusy(true);
    setRenameError("");
    try {
      await onRename(renaming.id, trimmedDraft);
      closeDialogs();
    } catch {
      setRenameError(labels?.renameFailed ?? "Unable to rename this chat.");
    } finally {
      setBusy(false);
    }
  };

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
                onRename={onRename ? () => startRename(c) : undefined}
                onTogglePin={onTogglePin}
              />
            ))}
          </>
        )}
      </div>
      {deleting && (
        <ChatDialog
          role="alertdialog"
          labelledBy="caret-delete-dialog-title"
          describedBy="caret-delete-dialog-message"
          busy={busy}
          onClose={closeDialogs}
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
          <div style={dialogActionsStyle}>
            <button
              type="button"
              autoFocus
              disabled={busy}
              onClick={closeDialogs}
              style={secondaryButtonStyle}
            >
              {labels?.cancel ?? "Cancel"}
            </button>
            <button
              type="button"
              disabled={busy}
              onClick={() => {
                void confirmDelete();
              }}
              style={{ ...primaryButtonStyle, background: "#b42318", color: "#fff" }}
            >
              {labels?.confirmDelete ?? "Delete"}
            </button>
          </div>
        </ChatDialog>
      )}
      {renaming && (
        <ChatDialog
          labelledBy="caret-rename-dialog-title"
          busy={busy}
          onClose={closeDialogs}
        >
          <form
            onSubmit={(event) => {
              event.preventDefault();
              void submitRename();
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "flex-start",
                justifyContent: "space-between",
                gap: 12,
              }}
            >
              <h2
                id="caret-rename-dialog-title"
                style={{ margin: "0 0 4px", fontSize: 18 }}
              >
                {labels?.renameTitle ?? "Rename chat"}
              </h2>
              <button
                type="button"
                aria-label={labels?.close ?? "Close"}
                disabled={busy}
                onClick={closeDialogs}
                style={{
                  display: "grid",
                  placeItems: "center",
                  width: 26,
                  height: 26,
                  margin: "-2px -4px 0 0",
                  border: "none",
                  borderRadius: caretRadius.sm,
                  background: "transparent",
                  color: caretTokens.textMuted,
                  cursor: "pointer",
                }}
              >
                <X size={16} />
              </button>
            </div>
            <p
              style={{
                margin: "0 0 14px",
                color: caretTokens.textMuted,
                fontSize: 13,
              }}
            >
              {labels?.renameHint ?? "Keep it short and recognizable"}
            </p>
            <input
              ref={renameInputRef}
              value={draft}
              maxLength={80}
              disabled={busy}
              aria-label={labels?.renameTitle ?? "Rename chat"}
              aria-invalid={renameError ? true : undefined}
              onChange={(event) => setDraft(event.target.value)}
              style={{
                width: "100%",
                boxSizing: "border-box",
                padding: "10px 12px",
                border: `1px solid ${caretTokens.border}`,
                borderRadius: caretRadius.md,
                background: caretTokens.surface,
                color: caretTokens.text,
                fontSize: 14,
                fontFamily: "inherit",
                outline: "none",
              }}
            />
            {renameError && (
              <p
                role="alert"
                style={{
                  margin: "8px 0 0",
                  color: caretTokens.dangerText,
                  fontSize: 12,
                }}
              >
                {renameError}
              </p>
            )}
            <div style={dialogActionsStyle}>
              <button
                type="button"
                disabled={busy}
                onClick={closeDialogs}
                style={secondaryButtonStyle}
              >
                {labels?.cancel ?? "Cancel"}
              </button>
              <button
                type="submit"
                disabled={busy || !trimmedDraft}
                style={{
                  ...primaryButtonStyle,
                  opacity: busy || !trimmedDraft ? 0.5 : 1,
                }}
              >
                {labels?.save ?? "Save"}
              </button>
            </div>
          </form>
        </ChatDialog>
      )}
    </div>
  );
};

export default ChatList;

const dialogActionsStyle: React.CSSProperties = {
  display: "flex",
  justifyContent: "flex-end",
  gap: 8,
  marginTop: 18,
};
const secondaryButtonStyle: React.CSSProperties = {
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  padding: "8px 12px",
  background: "transparent",
  color: caretTokens.text,
  cursor: "pointer",
  fontSize: 12.5,
};
const primaryButtonStyle: React.CSSProperties = {
  border: 0,
  borderRadius: caretRadius.sm,
  padding: "8px 14px",
  background: caretTokens.accent,
  color: caretTokens.accentInk,
  cursor: "pointer",
  fontSize: 12.5,
  fontWeight: 600,
};
