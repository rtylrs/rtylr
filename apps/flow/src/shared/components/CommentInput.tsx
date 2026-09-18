import React, { useRef, useState } from 'react';
import { AtSign, Send, Smile } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useFlowStore } from '../stores/flowStore';

interface CommentInputProps {
  onSubmit: (body: string) => void;
  placeholder?: string;
  disabled?: boolean;
}

export const CommentInput: React.FC<CommentInputProps> = ({ onSubmit, placeholder, disabled }) => {
  const { t } = useTranslation();
  const [value, setValue] = useState('');
  const [mentionOpen, setMentionOpen] = useState(false);
  const ref = useRef<HTMLTextAreaElement>(null);
  const { members } = useFlowStore();

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === '@') {
      setMentionOpen(true);
    } else if (e.key === 'Escape') {
      setMentionOpen(false);
    } else if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit();
    }
  };

  const handleSubmit = () => {
    if (disabled || !value.trim()) return;
    onSubmit(value.trim());
    setValue('');
    setMentionOpen(false);
  };

  const insertMention = (name: string) => {
    setValue((v) => v + name + ' ');
    setMentionOpen(false);
    ref.current?.focus();
  };

  return (
    <div
      className="relative rounded-lg"
      style={{
        background: 'var(--flow-input-bg)',
        border: '1px solid var(--flow-border-control)',
      }}
    >
      <textarea
        ref={ref}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={handleKey}
        placeholder={placeholder ?? t('flow.taskDetail.commentPlaceholder')}
        disabled={disabled}
        className="w-full bg-transparent resize-none p-3 text-meta disabled:opacity-60"
        rows={3}
        style={{ color: 'var(--flow-text)' }}
      />
      <div
        className="flex items-center justify-between px-2 py-1.5"
        style={{ borderTop: '1px solid var(--flow-border)' }}
      >
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setMentionOpen((o) => !o)}
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
            title={t('common.emojiPickerComingSoon')}
            aria-label={t('common.emojiPickerComingSoon')}
          >
            <Smile size={12} />
          </button>
        </div>
        <button
          type="button"
          onClick={handleSubmit}
          disabled={disabled || !value.trim()}
          className="flex items-center gap-1 px-2.5 py-1 rounded-sm text-meta font-semibold transition-opacity disabled:opacity-40"
          style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
        >
          <Send size={11} />
          {t('flow.taskDetail.comment')}
        </button>
      </div>
      {mentionOpen && (
        <div
          className="absolute start-2 bottom-full mb-1 rounded-lg py-1 z-20 w-[192px]"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
            boxShadow: 'var(--flow-shadow)',
          }}
        >
          {members.slice(0, 5).map((m) => (
            <button
              type="button"
              key={m.id}
              onClick={() => insertMention('@' + m.name)}
              className="w-full text-start px-3 py-1.5 text-meta hover:bg-[var(--flow-hover)] flex items-center gap-2"
              style={{ color: 'var(--flow-text)' }}
            >
              <span
                className="h-5 w-5 rounded-full flex items-center justify-center text-meta font-bold"
                style={{ background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }}
              >
                {m.initials}
              </span>
              {m.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default CommentInput;
