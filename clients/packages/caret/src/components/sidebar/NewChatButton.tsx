import React from "react";
import { SquarePen } from "lucide-react";
import { caretTokens, caretRadius } from "../../ui/tokens";

export interface NewChatButtonProps {
  onClick: () => void;
  label?: string;
}

export const NewChatButton: React.FC<NewChatButtonProps> = ({
  onClick,
  label = "New chat",
}) => (
  <button
    className="caret-new-chat-button"
    type="button"
    onClick={onClick}
    style={{
      display: "flex",
      alignItems: "center",
      gap: 9,
      width: "100%",
      minHeight: 34,
      padding: "6px 9px",
      background: "transparent",
      color: caretTokens.text,
      border: "1px solid transparent",
      borderRadius: caretRadius.md,
      fontSize: 12.5,
      fontWeight: 500,
      cursor: "pointer",
      transition: "background 0.12s ease, color 0.12s ease",
    }}
  >
    <SquarePen size={15} />
    {label}
  </button>
);

export default NewChatButton;
