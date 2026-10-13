import React, { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { caretTokens, caretRadius } from "../../ui/tokens";

export type ChatMenuEntry =
  | {
      id: string;
      label: string;
      icon?: React.ReactNode;
      destructive?: boolean;
      onSelect: () => void;
    }
  | { id: string; separator: true };

export interface ChatActionsMenuProps {
  /** Viewport point the menu opens at; `alignEnd` anchors its right edge there. */
  x: number;
  y: number;
  alignEnd?: boolean;
  label: string;
  /** Rendered inside the Caret shell so it inherits the theme's colours. */
  container: HTMLElement;
  entries: ChatMenuEntry[];
  onClose: () => void;
}

const MARGIN = 8;

/**
 * A small action menu for one chat row. It is portalled to the body so the
 * scrolling chat list can't clip it.
 */
export const ChatActionsMenu: React.FC<ChatActionsMenuProps> = ({
  x,
  y,
  alignEnd = false,
  label,
  container,
  entries,
  onClose,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  const [position, setPosition] = useState({ left: x, top: y });

  // Place the menu at the point, then keep it on screen: shift it left or
  // flip it above. The shell can be the containing block for fixed elements,
  // so the real offset is measured rather than assumed.
  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.style.left = `${x}px`;
    el.style.top = `${y}px`;
    const rect = el.getBoundingClientRect();
    const offsetX = rect.left - x;
    const offsetY = rect.top - y;
    let left = alignEnd ? x - rect.width : x;
    let top = y;
    if (left + rect.width > window.innerWidth - MARGIN)
      left = window.innerWidth - MARGIN - rect.width;
    if (top + rect.height > window.innerHeight - MARGIN)
      top = Math.max(MARGIN, y - rect.height - 8);
    setPosition({
      left: Math.max(MARGIN, left) - offsetX,
      top: top - offsetY,
    });
  }, [x, y, alignEnd]);

  useEffect(() => {
    ref.current?.querySelector<HTMLButtonElement>("button")?.focus();
    const onMouseDown = (event: MouseEvent): void => {
      if (!ref.current?.contains(event.target as Node)) onClose();
    };
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape") {
        event.preventDefault();
        event.stopPropagation();
        onClose();
        return;
      }
      if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
      const items = Array.from(
        ref.current?.querySelectorAll<HTMLButtonElement>("button") ?? [],
      );
      if (items.length === 0) return;
      event.preventDefault();
      const current = items.indexOf(document.activeElement as HTMLButtonElement);
      const step = event.key === "ArrowDown" ? 1 : -1;
      items[(current + step + items.length) % items.length].focus();
    };
    document.addEventListener("mousedown", onMouseDown);
    document.addEventListener("keydown", onKeyDown, true);
    window.addEventListener("resize", onClose);
    window.addEventListener("blur", onClose);
    // Capture scrolls from any ancestor, including the chat list.
    window.addEventListener("scroll", onClose, true);
    return () => {
      document.removeEventListener("mousedown", onMouseDown);
      document.removeEventListener("keydown", onKeyDown, true);
      window.removeEventListener("resize", onClose);
      window.removeEventListener("blur", onClose);
      window.removeEventListener("scroll", onClose, true);
    };
  }, [onClose]);

  return createPortal(
    <div
      ref={ref}
      role="menu"
      className="caret-chat-menu"
      aria-label={label}
      style={{
        position: "fixed",
        left: position.left,
        top: position.top,
        zIndex: 10002,
        minWidth: 190,
        padding: 6,
        boxSizing: "border-box",
        background: caretTokens.bg,
        border: `1px solid ${caretTokens.border}`,
        borderRadius: caretRadius.lg,
        boxShadow:
          "0 1px 2px rgba(15, 15, 20, 0.06), 0 12px 28px -8px rgba(15, 15, 20, 0.2)",
      }}
    >
      {entries.map((entry) =>
        "separator" in entry ? (
          <div
            key={entry.id}
            role="separator"
            style={{
              height: 1,
              margin: "5px 4px",
              background: caretTokens.border,
            }}
          />
        ) : (
          <button
            key={entry.id}
            type="button"
            role="menuitem"
            onClick={() => {
              onClose();
              entry.onSelect();
            }}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 10,
              width: "100%",
              minHeight: 34,
              padding: "6px 10px",
              border: "none",
              borderRadius: caretRadius.md,
              background: "transparent",
              color: entry.destructive ? caretTokens.danger : caretTokens.text,
              cursor: "pointer",
              fontSize: 13,
              textAlign: "start",
            }}
            onMouseEnter={(event) => event.currentTarget.focus()}
          >
            {entry.icon}
            <span>{entry.label}</span>
          </button>
        ),
      )}
    </div>,
    container,
  );
};

export default ChatActionsMenu;
