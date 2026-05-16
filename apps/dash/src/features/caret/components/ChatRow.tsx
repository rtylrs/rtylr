import React from 'react';
import { MessageSquare } from 'lucide-react';
import type { ChatSummary } from '../../../shared/api/caret';
import { relativeTime } from '../../../shared/lib/caret';

export interface ChatRowProps {
  chat: ChatSummary;
  onSelect: (id: string) => void;
}

/** A single chat in the dedicated history page. */
export const ChatRow: React.FC<ChatRowProps> = ({ chat, onSelect }) => (
  <button
    type="button"
    onClick={() => onSelect(chat.id)}
    style={{
      display: 'flex',
      alignItems: 'flex-start',
      gap: 10,
      padding: 10,
      borderRadius: 8,
      background: 'var(--dash-surface)',
      border: '1px solid var(--dash-border)',
      width: '100%',
      cursor: 'pointer',
      textAlign: 'left',
    }}
  >
    <MessageSquare size={14} style={{ color: 'var(--dash-text-muted)', marginTop: 2 }} />
    <span style={{ flex: 1, minWidth: 0 }}>
      <span
        style={{
          display: 'block',
          fontSize: 13,
          fontWeight: 600,
          color: 'var(--dash-text)',
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {chat.title || 'Untitled chat'}
      </span>
      {chat.summary && (
        <span
          style={{
            display: 'block',
            fontSize: 12,
            color: 'var(--dash-text-muted)',
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            marginTop: 2,
          }}
        >
          {chat.summary}
        </span>
      )}
    </span>
    <span style={{ fontSize: 11, color: 'var(--dash-text-dim)', flexShrink: 0 }}>
      {relativeTime(chat.updatedAt)}
    </span>
  </button>
);

export default ChatRow;
