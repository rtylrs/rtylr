import React, { useEffect, useMemo, useRef, useState } from 'react';
import { AtSign, Send, Smile } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export interface MentionCandidate {
  id: string;
  name: string;
  initials: string;
}

interface CommentInputProps {
  onSubmit: (body: string) => void;
  /** People the comment can mention. Picking one inserts "@Full Name". */
  mentionCandidates?: MentionCandidate[];
  placeholder?: string;
  disabled?: boolean;
  /** Starting text, for rewording an existing comment. */
  initialValue?: string;
  /** Label of the submit button; defaults to "Comment". */
  submitLabel?: string;
  /** Shows a Cancel button when set, for leaving an edit. */
  onCancel?: () => void;
  autoFocus?: boolean;
}

const MAX_SUGGESTIONS = 6;
// A name runs to a few words; past this the "@" is not a mention being typed.
const MAX_QUERY_LENGTH = 40;

interface ActiveMention {
  /** Index of the "@" that opened the mention. */
  start: number;
  query: string;
}

/**
 * The mention being typed at the caret: an "@" at the start of the text or
 * after whitespace, followed by up to a few words on the same line.
 */
export function findActiveMention(text: string, caret: number): ActiveMention | null {
  const before = text.slice(0, caret);
  const at = before.lastIndexOf('@');
  if (at < 0) return null;
  if (at > 0 && !/\s/.test(before[at - 1])) return null;
  const query = before.slice(at + 1);
  if (query.length > MAX_QUERY_LENGTH || /[\n@]/.test(query) || /^\s/.test(query)) return null;
  return { start: at, query };
}

/** Candidates whose name, or any word of it, starts with what was typed. */
export function matchMentionCandidates(
  candidates: MentionCandidate[],
  query: string
): MentionCandidate[] {
  const q = query.trim().toLocaleLowerCase();
  if (!q) return candidates.slice(0, MAX_SUGGESTIONS);
  return candidates
    .filter((c) => {
      const name = c.name.toLocaleLowerCase();
      return name.startsWith(q) || name.split(/\s+/).some((word) => word.startsWith(q));
    })
    .slice(0, MAX_SUGGESTIONS);
}

export const CommentInput: React.FC<CommentInputProps> = ({
  onSubmit,
  mentionCandidates = [],
  placeholder,
  disabled,
  initialValue = '',
  submitLabel,
  onCancel,
  autoFocus,
}) => {
  const { t } = useTranslation();
  const [value, setValue] = useState(initialValue);
  const [mention, setMention] = useState<ActiveMention | null>(null);
  const [highlight, setHighlight] = useState(0);
  const ref = useRef<HTMLTextAreaElement>(null);

  // Opening an edit puts the caret after the existing text, where a reworded
  // comment is usually continued, rather than before its first word.
  useEffect(() => {
    if (!autoFocus || !ref.current) return;
    const end = ref.current.value.length;
    ref.current.focus();
    ref.current.setSelectionRange(end, end);
    // Only on mount: later renders must not move the caret.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const suggestions = useMemo(
    () => (mention ? matchMentionCandidates(mentionCandidates, mention.query) : []),
    [mention, mentionCandidates]
  );
  // A query with a space that matches nobody is ordinary prose after an "@",
  // so the picker gets out of the way rather than saying "no matches".
  const pickerOpen = mention !== null && (suggestions.length > 0 || !mention.query.includes(' '));

  const syncMention = (text: string, caret: number) => {
    const next = findActiveMention(text, caret);
    setMention(next);
    if (next?.query !== mention?.query) setHighlight(0);
  };

  const handleSubmit = () => {
    if (disabled || !value.trim()) return;
    onSubmit(value.trim());
    if (!onCancel) setValue('');
    setMention(null);
  };

  const insertMention = (candidate: MentionCandidate) => {
    if (!mention) return;
    const caret = ref.current?.selectionStart ?? value.length;
    const inserted = `@${candidate.name} `;
    const next = value.slice(0, mention.start) + inserted + value.slice(caret);
    setValue(next);
    setMention(null);
    const position = mention.start + inserted.length;
    requestAnimationFrame(() => {
      ref.current?.focus();
      ref.current?.setSelectionRange(position, position);
    });
  };

  const startMention = () => {
    const el = ref.current;
    const caret = el?.selectionStart ?? value.length;
    const needsSpace = caret > 0 && !/\s/.test(value[caret - 1]);
    const prefix = needsSpace ? ' @' : '@';
    const next = value.slice(0, caret) + prefix + value.slice(caret);
    const position = caret + prefix.length;
    setValue(next);
    setMention({ start: position - 1, query: '' });
    setHighlight(0);
    requestAnimationFrame(() => {
      el?.focus();
      el?.setSelectionRange(position, position);
    });
  };

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (pickerOpen && suggestions.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setHighlight((h) => (h + 1) % suggestions.length);
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setHighlight((h) => (h - 1 + suggestions.length) % suggestions.length);
        return;
      }
      if ((e.key === 'Enter' && !e.metaKey && !e.ctrlKey && !e.shiftKey) || e.key === 'Tab') {
        e.preventDefault();
        insertMention(suggestions[Math.min(highlight, suggestions.length - 1)]);
        return;
      }
    }
    if (e.key === 'Escape') {
      if (pickerOpen) {
        e.preventDefault();
        e.stopPropagation();
        setMention(null);
      } else if (onCancel) {
        e.preventDefault();
        e.stopPropagation();
        onCancel();
      }
    } else if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleSubmit();
    }
  };

  const listId = 'flow-comment-mentions';

  return (
    <div
      className="flow-comment-box relative rounded-lg"
      style={{ background: 'var(--flow-input-bg)' }}
    >
      <textarea
        ref={ref}
        value={value}
        onChange={(e) => {
          setValue(e.target.value);
          syncMention(e.target.value, e.target.selectionStart);
        }}
        onSelect={(e) =>
          syncMention(e.currentTarget.value, e.currentTarget.selectionStart ?? value.length)
        }
        onBlur={() => setMention(null)}
        onKeyDown={handleKey}
        placeholder={placeholder ?? t('flow.taskDetail.commentPlaceholder')}
        disabled={disabled}
        className="w-full bg-transparent resize-none p-3 text-flow-meta disabled:opacity-60"
        rows={3}
        style={{ color: 'var(--flow-text)' }}
        role="combobox"
        aria-expanded={pickerOpen}
        aria-controls={pickerOpen ? listId : undefined}
        aria-autocomplete="list"
        aria-activedescendant={
          pickerOpen && suggestions.length > 0
            ? `${listId}-${suggestions[Math.min(highlight, suggestions.length - 1)].id}`
            : undefined
        }
      />
      <div
        className="flex items-center justify-between px-2 py-1.5"
        style={{ borderTop: '1px solid var(--flow-border)' }}
      >
        <div className="flex items-center gap-1">
          <button
            type="button"
            onMouseDown={(e) => e.preventDefault()}
            onClick={startMention}
            disabled={disabled}
            className="h-6 w-6 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
            style={{ color: 'var(--flow-text-dim)' }}
            title={t('flow.taskDetail.mentionSomeone')}
            aria-label={t('flow.taskDetail.mentionSomeone')}
          >
            <AtSign size={12} />
          </button>
          <button
            type="button"
            disabled
            className="h-6 w-6 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
            style={{ color: 'var(--flow-text-dim)' }}
            title={t('flowCommon.emojiPickerComingSoon')}
            aria-label={t('flowCommon.emojiPickerComingSoon')}
          >
            <Smile size={12} />
          </button>
        </div>
        <div className="flex items-center gap-1.5">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="px-2.5 py-1 rounded-sm text-flow-meta font-medium hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.taskDetail.cancelEdit')}
            </button>
          )}
          <button
            type="button"
            onClick={handleSubmit}
            disabled={disabled || !value.trim()}
            className="flex items-center gap-1 px-2.5 py-1 rounded-sm text-flow-meta font-semibold transition-opacity disabled:opacity-40"
            style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
          >
            {!onCancel && <Send size={11} />}
            {submitLabel ?? t('flow.taskDetail.comment')}
          </button>
        </div>
      </div>
      {pickerOpen && (
        <div
          id={listId}
          role="listbox"
          aria-label={t('flow.taskDetail.mentionSomeone')}
          className="absolute start-2 bottom-full mb-1 rounded-lg py-1 z-20 w-[240px] max-w-[calc(100%-1rem)]"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
            boxShadow: 'var(--flow-shadow)',
          }}
        >
          {suggestions.length === 0 ? (
            <p className="px-3 py-1.5 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {mentionCandidates.length === 0
                ? t('flow.taskDetail.noOneToMention')
                : t('flow.taskDetail.noMentionMatches')}
            </p>
          ) : (
            suggestions.map((m, index) => {
              const active = index === Math.min(highlight, suggestions.length - 1);
              return (
                <button
                  type="button"
                  key={m.id}
                  id={`${listId}-${m.id}`}
                  role="option"
                  aria-selected={active}
                  // Keep focus in the textarea so its blur does not close the
                  // list before the click lands.
                  onMouseDown={(e) => e.preventDefault()}
                  onMouseEnter={() => setHighlight(index)}
                  onClick={() => insertMention(m)}
                  className="w-full text-start px-3 py-1.5 text-flow-meta flex items-center gap-2"
                  style={{
                    color: 'var(--flow-text)',
                    background: active ? 'var(--flow-hover)' : undefined,
                  }}
                >
                  <span
                    className="h-5 w-5 shrink-0 rounded-full flex items-center justify-center text-flow-meta font-bold"
                    style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
                  >
                    {m.initials}
                  </span>
                  <span className="truncate">{m.name}</span>
                </button>
              );
            })
          )}
        </div>
      )}
    </div>
  );
};

export default CommentInput;
