import React, { useState } from "react";
import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
import { caretRadius, caretTokens } from "../ui/tokens";
import { useCaretMemories } from "../hooks/useCaretMemories";

const MAX_MEMORY_LENGTH = 500;

export interface MemorySettingsProps {
  userId: number;
  apiBaseUrl: string;
}

const iconButton: React.CSSProperties = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: 28,
  height: 28,
  flexShrink: 0,
  border: "none",
  borderRadius: caretRadius.sm,
  background: "transparent",
  color: caretTokens.textMuted,
  cursor: "pointer",
};

const field: React.CSSProperties = {
  flex: 1,
  minWidth: 0,
  padding: "7px 9px",
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  background: caretTokens.bg,
  color: caretTokens.text,
  font: "inherit",
  fontSize: 13,
  resize: "vertical",
};

/** Lets the person see, correct, add, and remove what Caret remembers. */
export const MemorySettings: React.FC<MemorySettingsProps> = ({
  userId,
  apiBaseUrl,
}) => {
  const { memories, isLoading, error, save, remove, isSaving } =
    useCaretMemories({ userId, apiBaseUrl });
  const [draft, setDraft] = useState("");
  const [editing, setEditing] = useState<{ id: string; value: string } | null>(
    null,
  );
  const [failed, setFailed] = useState(false);

  const run = async (action: () => Promise<unknown>) => {
    setFailed(false);
    try {
      await action();
      return true;
    } catch {
      setFailed(true);
      return false;
    }
  };

  const add = async (event: React.FormEvent) => {
    event.preventDefault();
    const value = draft.trim();
    if (!value) return;
    if (await run(() => save(value))) setDraft("");
  };

  const commitEdit = async () => {
    if (!editing) return;
    const value = editing.value.trim();
    if (!value) return;
    if (await run(() => save(value, editing.id))) setEditing(null);
  };

  return (
    <div
      style={{
        overflow: "hidden",
        background: caretTokens.surface,
        border: `1px solid ${caretTokens.border}`,
        borderRadius: caretRadius.lg,
      }}
    >
      <form
        onSubmit={add}
        style={{ display: "flex", gap: 6, padding: "10px 12px" }}
      >
        <input
          aria-label="Add something for Caret to remember"
          placeholder="Add something for Caret to remember"
          value={draft}
          maxLength={MAX_MEMORY_LENGTH}
          onChange={(event) => setDraft(event.target.value)}
          style={field}
        />
        <button
          type="submit"
          aria-label="Add memory"
          disabled={!draft.trim() || isSaving}
          style={{ ...iconButton, color: caretTokens.text }}
        >
          <Plus size={15} />
        </button>
      </form>

      {failed && (
        <p
          role="alert"
          style={{
            margin: 0,
            padding: "0 12px 10px",
            color: caretTokens.textMuted,
            fontSize: 12,
          }}
        >
          That didn't save. Try again.
        </p>
      )}

      {error ? (
        <p role="alert" style={emptyStyle}>
          Couldn't load memories.
        </p>
      ) : isLoading ? (
        <p role="status" style={emptyStyle}>
          Loading…
        </p>
      ) : memories.length === 0 ? (
        <p style={emptyStyle}>
          Nothing yet. Tell Caret "remember that…" in a chat, or add it here.
        </p>
      ) : (
        <ul style={{ listStyle: "none", margin: 0, padding: 0 }}>
          {memories.map((memory) => {
            const isEditing = editing?.id === memory.id;
            return (
              <li
                key={memory.id}
                className="caret-memory-row"
                style={{
                  display: "flex",
                  alignItems: isEditing ? "flex-start" : "center",
                  gap: 6,
                  padding: "8px 12px",
                  borderTop: `1px solid ${caretTokens.border}`,
                }}
              >
                {isEditing ? (
                  <>
                    <textarea
                      aria-label="Edit memory"
                      autoFocus
                      rows={2}
                      maxLength={MAX_MEMORY_LENGTH}
                      value={editing.value}
                      onChange={(event) =>
                        setEditing({ id: memory.id, value: event.target.value })
                      }
                      onKeyDown={(event) => {
                        if (event.key === "Enter" && !event.shiftKey) {
                          event.preventDefault();
                          void commitEdit();
                        }
                        if (event.key === "Escape") setEditing(null);
                      }}
                      style={field}
                    />
                    <button
                      type="button"
                      aria-label="Save memory"
                      onClick={() => void commitEdit()}
                      disabled={!editing.value.trim() || isSaving}
                      style={{ ...iconButton, color: caretTokens.text }}
                    >
                      <Check size={15} />
                    </button>
                    <button
                      type="button"
                      aria-label="Cancel editing"
                      onClick={() => setEditing(null)}
                      style={iconButton}
                    >
                      <X size={15} />
                    </button>
                  </>
                ) : (
                  <>
                    <span
                      style={{
                        flex: 1,
                        minWidth: 0,
                        color: caretTokens.text,
                        fontSize: 13,
                        lineHeight: 1.45,
                        overflowWrap: "anywhere",
                      }}
                    >
                      {memory.value}
                      {memory.origin && (
                        <span
                          style={{
                            display: "block",
                            marginTop: 2,
                            fontSize: 11,
                            color: caretTokens.textDim,
                          }}
                        >
                          {memory.origin === "feedback"
                            ? "Learned from your feedback"
                            : "Picked up from a chat"}
                        </span>
                      )}
                    </span>
                    <span
                      className="caret-memory-actions"
                      style={{ display: "flex" }}
                    >
                      <button
                        type="button"
                        aria-label={`Edit memory: ${memory.value}`}
                        onClick={() =>
                          setEditing({ id: memory.id, value: memory.value })
                        }
                        style={iconButton}
                      >
                        <Pencil size={13} />
                      </button>
                      <button
                        type="button"
                        aria-label={`Forget: ${memory.value}`}
                        onClick={() => void run(() => remove(memory.id))}
                        style={iconButton}
                      >
                        <Trash2 size={13} />
                      </button>
                    </span>
                  </>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
};

const emptyStyle: React.CSSProperties = {
  margin: 0,
  padding: "10px 12px 12px",
  borderTop: `1px solid ${caretTokens.border}`,
  color: caretTokens.textMuted,
  fontSize: 12,
  lineHeight: 1.45,
};

export default MemorySettings;
