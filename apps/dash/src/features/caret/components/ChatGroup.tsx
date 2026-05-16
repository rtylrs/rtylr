import React from 'react';
import ChatRow from './ChatRow';
import type { ChatSummary } from '../../../shared/api/caret';

export interface ChatGroupProps {
  label: string;
  chats: ChatSummary[];
  onSelect: (id: string) => void;
}

/** A labeled group of chats (e.g. "Today", "Last 7 days"). */
export const ChatGroup: React.FC<ChatGroupProps> = ({ label, chats, onSelect }) => (
  <section style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
    <h3
      style={{
        margin: 0,
        fontSize: 11,
        fontWeight: 600,
        color: 'var(--dash-text-dim)',
        textTransform: 'uppercase',
        letterSpacing: 0.4,
      }}
    >
      {label}
    </h3>
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      {chats.map((c) => (
        <ChatRow key={c.id} chat={c} onSelect={onSelect} />
      ))}
    </div>
  </section>
);

export default ChatGroup;
