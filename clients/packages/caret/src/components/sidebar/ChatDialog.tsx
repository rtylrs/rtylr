import React, { useEffect, useRef } from "react";
import { caretTokens, caretRadius } from "../../ui/tokens";
import { useFocusTrap } from "../../hooks/useFocusTrap";

export interface ChatDialogProps {
  labelledBy: string;
  describedBy?: string;
  role?: "dialog" | "alertdialog";
  /** While busy the dialog can't be dismissed. */
  busy?: boolean;
  width?: number;
  onClose: () => void;
  children: React.ReactNode;
}

/** The centred modal shared by the chat list's rename and delete prompts. */
export const ChatDialog: React.FC<ChatDialogProps> = ({
  labelledBy,
  describedBy,
  role = "dialog",
  busy = false,
  width = 420,
  onClose,
  children,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  useFocusTrap(ref, true);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape" && !busy) onClose();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [busy, onClose]);

  return (
    <div
      className="caret-chat-dialog-backdrop"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !busy) onClose();
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
        ref={ref}
        role={role}
        aria-modal="true"
        aria-labelledby={labelledBy}
        aria-describedby={describedBy}
        style={{
          width: `min(${width}px, 100%)`,
          background: caretTokens.bg,
          border: `1px solid ${caretTokens.border}`,
          borderRadius: caretRadius.lg,
          padding: 20,
          boxSizing: "border-box",
          color: caretTokens.text,
          boxShadow: caretTokens.shadow,
        }}
      >
        {children}
      </div>
    </div>
  );
};

export default ChatDialog;
