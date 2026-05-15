import React, { useCallback, useEffect, useRef, useState } from 'react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import ModePicker from './ModePicker';
import ModelPicker from './ModelPicker';
import MicButton from './MicButton';
import AttachButton from './AttachButton';
import SendButton from './SendButton';
import StopButton from './StopButton';
import { useAutosize } from '../../hooks/useAutosize';

export interface ComposerProps {
  isStreaming: boolean;
  onSubmit: (text: string) => void;
  onStop: () => void;
  onAttach?: (files: File[]) => void;
  placeholder?: string;
  autoFocus?: boolean;
}

export const Composer: React.FC<ComposerProps> = ({
  isStreaming,
  onSubmit,
  onStop,
  onAttach,
  placeholder = 'Ask Caret anything…',
  autoFocus = true,
}) => {
  const [value, setValue] = useState('');
  const taRef = useRef<HTMLTextAreaElement>(null);
  useAutosize(taRef, value);

  useEffect(() => {
    if (autoFocus) taRef.current?.focus();
  }, [autoFocus]);

  const send = useCallback((): void => {
    const trimmed = value.trim();
    if (!trimmed || isStreaming) return;
    onSubmit(trimmed);
    setValue('');
  }, [value, isStreaming, onSubmit]);

  const handleKey = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  };

  return (
    <div
      style={{
        margin: 12,
        background: caretTokens.surface,
        border: `1px solid ${caretTokens.border}`,
        borderRadius: caretRadius.lg,
        padding: 8,
        display: 'flex',
        flexDirection: 'column',
        gap: 6,
      }}
    >
      <textarea
        ref={taRef}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={handleKey}
        placeholder={placeholder}
        rows={1}
        aria-label="Message Caret"
        style={{
          resize: 'none',
          width: '100%',
          background: 'transparent',
          border: 'none',
          outline: 'none',
          color: caretTokens.text,
          fontSize: 14,
          lineHeight: 1.5,
          padding: '4px 6px',
          fontFamily: 'inherit',
        }}
      />
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
        }}
      >
        <ModePicker />
        <ModelPicker />
        <span style={{ flex: 1 }} />
        {onAttach && <AttachButton onFiles={onAttach} />}
        <MicButton
          onTranscript={(text) =>
            setValue((prev) => (prev.length && !prev.endsWith(' ') ? prev + ' ' : prev) + text)
          }
        />
        {isStreaming ? (
          <StopButton onStop={onStop} />
        ) : (
          <SendButton disabled={!value.trim()} onClick={send} />
        )}
      </div>
    </div>
  );
};

export default Composer;
