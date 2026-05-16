import React, { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCaretChats } from '../../shared/hooks/queries/useCaretChats';
import { useCaretStore } from '@voxire/caret';
import { groupChatsByDate, scoreChat } from '../../shared/lib/caret';
import { PageHeader, EmptyHistory, ChatGroup, SearchBar } from './components';

/** Full-page Caret chat history (separate from the side-panel view). */
const CaretHistoryPage: React.FC = () => {
  const navigate = useNavigate();
  const setSessionId = useCaretStore((s) => s.setSessionId);
  const openPanel = useCaretStore((s) => s.open);
  const { data: chats = [], isLoading } = useCaretChats();
  const [query, setQuery] = useState('');

  const filtered = useMemo(() => {
    if (!query.trim()) return chats;
    return [...chats]
      .map((c) => ({ chat: c, score: scoreChat(c, query) }))
      .filter((x) => x.score > 0)
      .sort((a, b) => b.score - a.score)
      .map((x) => x.chat);
  }, [chats, query]);

  const grouped = useMemo(() => groupChatsByDate(filtered), [filtered]);

  const onSelect = (id: string): void => {
    setSessionId(id);
    openPanel();
    navigate('/');
  };

  return (
    <div className="space-y-4 max-w-3xl">
      <PageHeader title="Chats" subtitle="Everything you've asked Caret." />
      <SearchBar value={query} onChange={setQuery} />
      {isLoading && <p style={{ color: 'var(--dash-text-muted)' }}>Loading…</p>}
      {!isLoading && chats.length === 0 && <EmptyHistory />}
      {!isLoading && chats.length > 0 && filtered.length === 0 && (
        <p style={{ color: 'var(--dash-text-muted)' }}>No chats match "{query}".</p>
      )}
      {grouped.map((g) => (
        <ChatGroup key={g.group} label={g.label} chats={g.chats} onSelect={onSelect} />
      ))}
    </div>
  );
};

export default CaretHistoryPage;
