import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, Send, Lock, Circle, User } from 'lucide-react';
import { useTicket, useTicketMessages, useSendTicketMessage, useUpdateTicketStatus } from './useTickets';

const statusColors: Record<string, string> = {
  open: '#3b82f6',
  pending: '#f59e0b',
  resolved: '#10b981',
  closed: '#6b7280',
};

const TicketDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const ticketId = Number(id);
  const { data: ticket, isLoading: ticketLoading, isError: ticketError } = useTicket(ticketId);
  const { data: messagesData, isLoading: messagesLoading } = useTicketMessages(ticketId);
  const sendMessage = useSendTicketMessage();
  const updateStatus = useUpdateTicketStatus();

  const [reply, setReply] = useState('');
  const [isInternal, setIsInternal] = useState(false);

  const messages: any[] = messagesData ?? [];

  const handleSend = () => {
    if (!reply.trim()) return;
    sendMessage.mutate(
      { ticketId, data: { body: reply, isInternal } },
      { onSuccess: () => setReply('') }
    );
  };

  const handleStatusChange = (status: string) => {
    updateStatus.mutate({ id: ticketId, status });
  };

  if (ticketLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 rounded animate-pulse" style={{ background: 'var(--hub-hover)' }} />
        <div className="h-64 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
      </div>
    );
  }

  if (ticketError || !ticket) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-sm" style={{ color: 'var(--hub-text-muted)' }}>
        <p>Ticket not found</p>
        <Link to="/tickets" className="mt-2 text-xs" style={{ color: '#4583C4' }}>Back to tickets</Link>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <Link
          to="/tickets"
          className="h-7 w-7 rounded-lg flex items-center justify-center"
          style={{ color: 'var(--hub-text-muted)', border: '1px solid var(--hub-border)' }}
        >
          <ArrowLeft size={14} />
        </Link>
        <div className="flex-1">
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            {ticket.subject}
          </h2>
          <div className="flex items-center gap-3 mt-1">
            <span className="text-xs font-mono" style={{ color: 'var(--hub-text-dim)' }}>#{ticket.id}</span>
            <span
              className="inline-flex items-center gap-1 text-[11px] font-medium px-2 py-0.5 rounded-full"
              style={{
                background: `${statusColors[ticket.status] || '#6b7280'}15`,
                color: statusColors[ticket.status] || '#6b7280',
              }}
            >
              <Circle size={5} fill="currentColor" stroke="none" />
              {ticket.status}
            </span>
            <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              {ticket.customerName}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-1">
          {['open', 'pending', 'resolved', 'closed'].map((s) => (
            <button
              key={s}
              onClick={() => handleStatusChange(s)}
              disabled={ticket.status === s}
              className="h-7 px-2.5 rounded-lg text-[11px] font-medium transition-colors disabled:opacity-30"
              style={{
                background: ticket.status === s ? `${statusColors[s]}15` : 'var(--hub-hover)',
                color: ticket.status === s ? statusColors[s] : 'var(--hub-text-muted)',
                border: '1px solid var(--hub-border)',
              }}
            >
              {s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>
      </div>

      <div
        className="rounded-xl overflow-hidden mb-4"
        style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
      >
        {messagesLoading && (
          <div className="p-6 space-y-4">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-16 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
            ))}
          </div>
        )}

        {!messagesLoading && messages.length === 0 && (
          <div className="flex items-center justify-center py-12 text-sm" style={{ color: 'var(--hub-text-muted)' }}>
            No messages yet
          </div>
        )}

        {!messagesLoading && messages.length > 0 && (
          <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
            {messages.map((msg: any) => (
              <div
                key={msg.id}
                className="px-5 py-4"
                style={{
                  background: msg.isInternal ? 'rgba(245,158,11,0.04)' : 'transparent',
                  borderLeft: msg.isInternal ? '3px solid #f59e0b' : '3px solid transparent',
                }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <div
                    className="h-6 w-6 rounded-full flex items-center justify-center text-[10px] font-bold"
                    style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                  >
                    <User size={12} />
                  </div>
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {msg.senderName}
                  </span>
                  {msg.isInternal && (
                    <span className="flex items-center gap-1 text-[10px] px-1.5 py-0.5 rounded" style={{ background: 'rgba(245,158,11,0.1)', color: '#f59e0b' }}>
                      <Lock size={8} />
                      Internal note
                    </span>
                  )}
                  <span className="text-[10px] ml-auto" style={{ color: 'var(--hub-text-dim)' }}>
                    {msg.createdAt}
                  </span>
                </div>
                <p className="text-sm leading-relaxed" style={{ color: 'var(--hub-text)' }}>
                  {msg.body}
                </p>
              </div>
            ))}
          </div>
        )}
      </div>

      <div
        className="rounded-xl p-4"
        style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
      >
        <div className="flex items-center gap-2 mb-3">
          <button
            onClick={() => setIsInternal(false)}
            className="text-xs font-medium px-3 py-1 rounded-lg transition-colors"
            style={{
              background: !isInternal ? 'var(--hub-active)' : 'transparent',
              color: !isInternal ? 'var(--hub-text)' : 'var(--hub-text-muted)',
            }}
          >
            Reply
          </button>
          <button
            onClick={() => setIsInternal(true)}
            className="flex items-center gap-1 text-xs font-medium px-3 py-1 rounded-lg transition-colors"
            style={{
              background: isInternal ? 'rgba(245,158,11,0.12)' : 'transparent',
              color: isInternal ? '#f59e0b' : 'var(--hub-text-muted)',
            }}
          >
            <Lock size={10} />
            Internal note
          </button>
        </div>
        <textarea
          placeholder={isInternal ? 'Write an internal note...' : 'Type your reply...'}
          value={reply}
          onChange={(e) => setReply(e.target.value)}
          rows={3}
          className="w-full bg-transparent outline-none text-sm resize-none mb-3"
          style={{
            color: 'var(--hub-text)',
            borderBottom: '1px solid var(--hub-border)',
            paddingBottom: '12px',
          }}
        />
        <div className="flex justify-end">
          <button
            onClick={handleSend}
            disabled={!reply.trim() || sendMessage.isPending}
            className="flex items-center gap-2 h-8 px-4 rounded-lg text-xs font-medium text-white transition-colors disabled:opacity-40"
            style={{ background: isInternal ? '#f59e0b' : '#4583C4' }}
          >
            <Send size={12} />
            {isInternal ? 'Add note' : 'Send reply'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default TicketDetail;
