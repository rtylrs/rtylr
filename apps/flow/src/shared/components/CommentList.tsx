import React from 'react';
import { Comment } from '../stores/flowStore';
import AssigneeAvatar from './AssigneeAvatar';

interface CommentListProps {
  comments: Comment[];
}

function timeAgo(ts: string): string {
  const d = new Date(ts);
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function renderBody(body: string): React.ReactNode {
  // simple @mention highlight + basic markdown (bold/italic/code)
  const parts = body.split(/(@\S+|\*\*[^*]+\*\*|`[^`]+`|\*[^*]+\*)/g).filter(Boolean);
  return parts.map((p, i) => {
    if (p.startsWith('@')) {
      return (
        <span
          key={i}
          className="rounded px-1 font-medium"
          style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
        >
          {p}
        </span>
      );
    }
    if (p.startsWith('**') && p.endsWith('**')) {
      return <strong key={i}>{p.slice(2, -2)}</strong>;
    }
    if (p.startsWith('*') && p.endsWith('*')) {
      return <em key={i}>{p.slice(1, -1)}</em>;
    }
    if (p.startsWith('`') && p.endsWith('`')) {
      return (
        <code
          key={i}
          className="rounded px-1 py-0.5 font-mono text-[11px]"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
        >
          {p.slice(1, -1)}
        </code>
      );
    }
    return <React.Fragment key={i}>{p}</React.Fragment>;
  });
}

export const CommentList: React.FC<CommentListProps> = ({ comments }) => {
  if (comments.length === 0) {
    return (
      <div className="text-[12px] text-center py-6" style={{ color: 'var(--flow-text-dim)' }}>
        No comments yet. Start the conversation.
      </div>
    );
  }
  return (
    <div className="space-y-4">
      {comments.map((c) => (
        <div key={c.id} className="flex gap-3">
          <AssigneeAvatar name={c.author} initials={c.authorInitials} size={28} />
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                {c.author}
              </span>
              <span className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
                {timeAgo(c.createdAt)}
              </span>
            </div>
            <div
              className="text-[12px] leading-relaxed"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {renderBody(c.body)}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};

export default CommentList;
