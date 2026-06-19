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
        gap: 8,
        padding: 12,
        borderBottom: `1px solid ${caretTokens.border}`,
      }}
    >
      <NewChatButton onClick={onNew} />
      <ChatSearch value={query} onChange={setQuery} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
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
