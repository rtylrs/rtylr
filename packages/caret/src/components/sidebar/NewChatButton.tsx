import React from 'react';
import { Plus } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';

export interface NewChatButtonProps {
  onClick: () => void;
}

export const NewChatButton: React.FC<NewChatButtonProps> = ({ onClick }) => (
  <button
    className="caret-new-chat-button"
    type="button"
    onClick={onClick}
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      width: '100%',
      minHeight: 36,
      padding: '8px 10px',
      background: 'transparent',
      color: caretTokens.textMuted,
      border: '1px solid transparent',
      borderRadius: caretRadius.md,
      fontSize: 13,
      fontWeight: 550,
      cursor: 'pointer',
      transition: 'background 0.12s ease, color 0.12s ease',
    }}
  >
    <Plus size={13} />
    New chat
  </button>
);

export default NewChatButton;
