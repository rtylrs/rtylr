import React, { useMemo, useState } from 'react';
import { caretTokens } from '../../ui/tokens';
import ChatListItem from './ChatListItem';
import ChatSearch from './ChatSearch';
import NewChatButton from './NewChatButton';
import { useDebouncedValue } from '../../hooks/useDebouncedValue';
import type { CaretChatSummary } from '../../types/chat';

export interface ChatListProps {
  chats: CaretChatSummary[];
  activeId?: string | null;
  onSelect: (id: string) => void;
  onNew: () => void;
  onDelete?: (id: string) => void | Promise<void>;
  onTogglePin?: (id: string) => void;
}

export const ChatList: React.FC<ChatListProps> = ({
  chats,
  activeId,
  onSelect,
  onNew,
  onDelete,
  onTogglePin,
}) => {
  const [query, setQuery] = useState('');
  const debounced = useDebouncedValue(query, 150);

  const filtered = useMemo(() => {
    const q = debounced.toLowerCase().trim();
    const ordered = [...chats].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
      return b.updatedAt.localeCompare(a.updatedAt);
    });
    if (!q) return ordered;
    return ordered.filter(
      (c) => c.title.toLowerCase().includes(q) || (c.summary ?? '').toLowerCase().includes(q)
    );
  }, [chats, debounced]);

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 9,
        padding: '18px 14px 14px',
        flex: 1,
        minHeight: 0,
        overflow: 'hidden',
        boxSizing: 'border-box',
      }}
    >
      <div style={{ margin: '0 2px 7px' }}>
        <h2
          style={{
            margin: 0,
            color: caretTokens.text,
            fontSize: 15,
            fontWeight: 650,
            lineHeight: 1.4,
          }}
        >
          History
        </h2>
        <p style={{ margin: '3px 0 0', color: caretTokens.textMuted, fontSize: 11.5 }}>
          Open a past chat or start another.
        </p>
      </div>
      <NewChatButton onClick={onNew} />
      <ChatSearch value={query} onChange={setQuery} />
      <div
        className="caret-chat-list-scroll"
        style={{
          display: 'flex',
          flex: 1,
          flexDirection: 'column',
          gap: 2,
          minHeight: 0,
          overflowY: 'auto',
          paddingRight: 2,
        }}
      >
        {filtered.length === 0 && (
          <p
            style={{
              fontSize: 12,
              color: caretTokens.textDim,
              padding: '8px 4px',
              margin: 0,
            }}
          >
            {query ? 'No chats match your search.' : 'No saved chats yet.'}
          </p>
        )}
        {filtered.map((c) => (
          <ChatListItem
            key={c.id}
            chat={c}
            active={c.id === activeId}
            onSelect={onSelect}
            onDelete={onDelete}
            onTogglePin={onTogglePin}
          />
        ))}
      </div>
    </div>
  );
};

export default ChatList;
