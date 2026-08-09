import React from 'react';
import { Search, X } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';

export interface ChatSearchProps {
  value: string;
  onChange: (v: string) => void;
}

export const ChatSearch: React.FC<ChatSearchProps> = ({ value, onChange }) => (
  <div
    className="caret-chat-search"
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      minHeight: 36,
      padding: '7px 10px',
      background: caretTokens.surfaceHigh,
      borderRadius: caretRadius.lg,
      border: '1px solid transparent',
      transition: 'border-color 0.14s ease, box-shadow 0.14s ease',
    }}
  >
    <Search size={13} style={{ color: caretTokens.textDim }} />
    <input
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder="Search chats…"
      aria-label="Search chats"
      style={{
        flex: 1,
        background: 'transparent',
        border: 'none',
        outline: 'none',
        color: caretTokens.text,
        fontSize: 13,
      }}
    />
    {value && (
      <button
        type="button"
        aria-label="Clear search"
        onClick={() => onChange('')}
        style={{
          background: 'transparent',
          border: 'none',
          color: caretTokens.textDim,
          cursor: 'pointer',
          padding: 0,
          display: 'flex',
        }}
      >
        <X size={11} />
      </button>
    )}
  </div>
);

export default ChatSearch;
