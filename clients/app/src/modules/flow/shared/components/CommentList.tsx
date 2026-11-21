import React, { useState } from 'react';
import { Pencil } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Comment } from '../stores/flowStore';
import AssigneeAvatar from './AssigneeAvatar';
import CommentInput, { type MentionCandidate } from './CommentInput';

interface CommentListProps {
  comments: Comment[];
  /** The signed-in user; only their own comments offer Edit. */
  currentUserId?: string;
  /** Saves a reworded comment. Without it no comment offers Edit. */
  onEdit?: (commentId: string, body: string) => void;
  /** People the edit box can mention. */
  mentionCandidates?: MentionCandidate[];
  /**
   * Every name a comment may mention, including the signed-in user's, so
   * "@Full Name" highlights as one mention. Defaults to the candidates' names.
   */
  mentionNames?: string[];
}

function escapeRegExp(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function timeAgo(ts: string, t: ReturnType<typeof useTranslation>['t']): string {
  const d = new Date(ts);
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return t('flow.taskDetail.justNow');
  if (mins < 60) return t('flow.taskDetail.minutesAgo', { count: mins });
  const hours = Math.floor(mins / 60);
  if (hours < 24) return t('flow.taskDetail.hoursAgo', { count: hours });
  const days = Math.floor(hours / 24);
  return t('flow.taskDetail.daysAgo', { count: days });
}

/**
 * A comment's text with its @mentions and inline bold, italic and code, as
 * the thread shows it. The task print view reuses it so paper matches screen.
 */
export function renderCommentBody(body: string, mentionNames: string[]): React.ReactNode {
  // @mention highlight + basic markdown (bold/italic/code). Known names go
  // first, longest first, so "@Mohamad Issa" highlights as one mention rather
  // than stopping at the first space.
  const names = [...mentionNames]
    .filter(Boolean)
    .sort((a, b) => b.length - a.length)
    .map((name) => `@${escapeRegExp(name)}(?![\\p{L}\\p{N}])`);
  const pattern = new RegExp(
    `(${[...names, '@\\S+', '\\*\\*[^*]+\\*\\*', '`[^`]+`', '\\*[^*]+\\*'].join('|')})`,
    'giu'
  );
  const parts = body.split(pattern).filter(Boolean);
  return parts.map((p, i) => {
    if (p.startsWith('@')) {
      return (
        <span
          key={i}
          className="rounded-sm px-1 font-medium"
          style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
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
          className="rounded-sm px-1 py-0.5 font-mono text-flow-meta"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
        >
          {p.slice(1, -1)}
        </code>
      );
    }
    return <React.Fragment key={i}>{p}</React.Fragment>;
  });
}

export const CommentList: React.FC<CommentListProps> = ({
  comments,
  currentUserId,
  onEdit,
  mentionCandidates = [],
  mentionNames = mentionCandidates.map((c) => c.name),
}) => {
  const { t } = useTranslation();
  const [editingId, setEditingId] = useState<string | null>(null);
  if (comments.length === 0) {
    return (
      <div className="text-flow-meta text-center py-6" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.taskDetail.noComments')}
      </div>
    );
  }
  return (
    <div className="space-y-4">
      {comments.map((c) => {
        const canEdit =
          !!onEdit && currentUserId !== undefined && String(c.authorId) === currentUserId;
        const editing = canEdit && editingId === c.id;
        return (
          <div key={c.id} className="flex gap-3 group">
            <AssigneeAvatar
              name={c.author}
              initials={c.authorInitials}
              seed={c.authorId ? String(c.authorId) : undefined}
              size={28}
            />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <span
                  className="text-flow-meta font-semibold"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {c.author}
                </span>
                <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {timeAgo(c.createdAt, t)}
                </span>
                {c.edited && (
                  <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                    {t('flow.taskDetail.edited')}
                  </span>
                )}
                {canEdit && !editing && (
                  <button
                    type="button"
                    onClick={() => setEditingId(c.id)}
                    className="ms-auto flex items-center gap-1 px-1.5 py-0.5 rounded-sm text-flow-meta opacity-100 sm:opacity-0 sm:group-hover:opacity-100 focus-visible:opacity-100 hover:bg-[var(--flow-hover)]"
                    style={{ color: 'var(--flow-text-dim)' }}
                    aria-label={t('flow.taskDetail.editCommentLabel')}
                  >
                    <Pencil size={11} />
                    {t('flow.taskDetail.editComment')}
                  </button>
                )}
              </div>
              {editing ? (
                <CommentInput
                  initialValue={c.body}
                  mentionCandidates={mentionCandidates}
                  submitLabel={t('flow.taskDetail.saveComment')}
                  autoFocus
                  onCancel={() => setEditingId(null)}
                  onSubmit={(body) => {
                    if (body !== c.body) onEdit?.(c.id, body);
                    setEditingId(null);
                  }}
                />
              ) : (
                <div
                  className="text-flow-meta leading-relaxed whitespace-pre-wrap break-words"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {renderCommentBody(c.body, mentionNames)}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default CommentList;
