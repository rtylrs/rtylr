import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Mail, MessageCircle, Circle, User, Search, Filter } from 'lucide-react';
import { useConversations } from './useConversations';

const channelIcon = (channel: string) => {
  if (channel === 'chat') return MessageCircle;
  return Mail;
};

const InboxScreen: React.FC = () => {
  const [search, setSearch] = useState('');
  const [channelFilter, setChannelFilter] = useState<string>('');
  const { data, isLoading, isError } = useConversations({
    channel: channelFilter || undefined,
  });

  const conversations: any[] = data ?? [];
  const filtered = search
    ? conversations.filter(
        (c: any) =>
          c.subject?.toLowerCase().includes(search.toLowerCase()) ||
          c.customerName?.toLowerCase().includes(search.toLowerCase())
      )
    : conversations;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Inbox
          </h2>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            All conversations across email and chat
          </p>
        </div>
      </div>

      <div className="flex items-center gap-3 mb-4">
        <div
          className="flex items-center gap-2 flex-1 h-9 px-3 rounded-lg text-sm"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={14} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search conversations..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <select
          value={channelFilter}
          onChange={(e) => setChannelFilter(e.target.value)}
          className="h-9 px-3 rounded-lg text-sm outline-none"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)', color: 'var(--hub-text)' }}
        >
          <option value="">All channels</option>
          <option value="email">Email</option>
          <option value="chat">Chat</option>
        </select>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className="h-16 rounded-lg animate-pulse"
              style={{ background: 'var(--hub-hover)' }}
            />
          ))}
        </div>
      )}

      {isError && (
        <div
          className="flex items-center justify-center py-16 rounded-xl text-sm"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
        >
          Failed to load conversations. Please try again.
        </div>
      )}

      {!isLoading && !isError && filtered.length === 0 && (
        <div
          className="flex flex-col items-center justify-center py-16 rounded-xl text-sm"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
        >
          <Mail size={32} className="mb-3" style={{ color: 'var(--hub-text-dim)' }} />
          <p>No conversations found</p>
        </div>
      )}

      {!isLoading && !isError && filtered.length > 0 && (
        <div className="space-y-1">
          {filtered.map((conv: any) => {
            const Icon = channelIcon(conv.channel);
            return (
              <motion.div
                key={conv.id}
                initial={{ opacity: 0, y: 4 }}
                animate={{ opacity: 1, y: 0 }}
                className="flex items-center gap-3 px-4 py-3 rounded-lg cursor-pointer transition-colors"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-card)'; }}
              >
                <div className="relative">
                  <Icon size={16} style={{ color: 'var(--hub-text-muted)' }} />
                  {conv.unread && (
                    <Circle size={6} fill="#4583C4" stroke="none" className="absolute -top-0.5 -right-0.5" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                      {conv.customerName || 'Unknown'}
                    </span>
                    <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {conv.lastMessageAt}
                    </span>
                  </div>
                  <p className="text-xs truncate mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
                    {conv.subject || conv.lastMessage || 'No subject'}
                  </p>
                </div>
                {conv.assigneeName && (
                  <div className="flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full" style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}>
                    <User size={10} />
                    {conv.assigneeName}
                  </div>
                )}
              </motion.div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default InboxScreen;
