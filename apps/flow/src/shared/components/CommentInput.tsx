import React, { useRef, useState } from 'react';
import { AtSign, Send, Smile } from 'lucide-react';
import { useFlowStore } from '../stores/flowStore';

interface CommentInputProps {
  onSubmit: (body: string) => void;
  placeholder?: string;
  disabled?: boolean;
}

export const CommentInput: React.FC<CommentInputProps> = ({
  onSubmit,
  placeholder = 'Leave a comment… @mention to notify',
  disabled,
}) => {
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
        border: '1px solid var(--flow-border)',
      }}
    >
      <textarea
        ref={ref}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={handleKey}
        placeholder={placeholder}
        disabled={disabled}
        className="w-full bg-transparent outline-none resize-none p-3 text-[12px] disabled:opacity-60"
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
            className="h-6 w-6 rounded flex items-center justify-center hover:bg-[var(--flow-hover)] disabled:opacity-40"
            style={{ color: 'var(--flow-text-dim)' }}
            title="Mention someone"
          >
            <AtSign size={12} />
          </button>
          <button
            type="button"
            onClick={undefined}
            disabled
            className="h-6 w-6 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
            title="Emoji picker"
          >
            <Smile size={12} />
          </button>
        </div>
        <button
          type="button"
          onClick={handleSubmit}
          disabled={disabled || !value.trim()}
          className="flex items-center gap-1 px-2.5 py-1 rounded text-[11px] font-semibold transition-opacity disabled:opacity-40"
          style={{ background: 'var(--accent)', color: '#0a0a0a' }}
        >
          <Send size={11} />
          Comment
        </button>
      </div>
      {mentionOpen && (
        <div
          className="absolute left-2 bottom-full mb-1 rounded-lg py-1 z-20 w-48"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
            boxShadow: 'var(--flow-shadow)',
          }}
        >
          {members.slice(0, 5).map((m) => (
            <button
              key={m.id}
              onClick={() => insertMention('@' + m.name)}
              className="w-full text-left px-3 py-1.5 text-[12px] hover:bg-[var(--flow-hover)] flex items-center gap-2"
              style={{ color: 'var(--flow-text)' }}
            >
              <span
                className="h-5 w-5 rounded-full flex items-center justify-center text-[9px] font-bold"
                style={{ background: 'var(--accent-muted)', color: 'var(--accent)' }}
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
