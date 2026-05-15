import React from 'react';
import { Plus } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';

export interface NewChatButtonProps {
  onClick: () => void;
}

export const NewChatButton: React.FC<NewChatButtonProps> = ({ onClick }) => (
  <button
    type="button"
    onClick={onClick}
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      width: '100%',
      padding: '8px 10px',
      background: caretTokens.accent,
      color: '#ffffff',
      border: 'none',
      borderRadius: caretRadius.md,
      fontSize: 13,
      fontWeight: 500,
      cursor: 'pointer',
    }}
  >
    <Plus size={13} />
    New chat
  </button>
);

export default NewChatButton;
