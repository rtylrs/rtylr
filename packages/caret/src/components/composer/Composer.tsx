import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import ModePicker from './ModePicker';
import MicButton from './MicButton';
import AttachButton from './AttachButton';
import SendButton from './SendButton';
import StopButton from './StopButton';
import { useAutosize } from '../../hooks/useAutosize';
import { CARET_FOCUS_COMPOSER_EVENT } from '../../lib/focusComposer';
import { inferCaretModeForPrompt } from '../../lib/modeIntent';
import { useCaretSettingsStore } from '../../stores/settingsStore';
import { CARET_MODES } from '../../types/mode';

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
  const mode = useCaretSettingsStore((s) => s.mode);
  const voiceEnabled = useCaretSettingsStore((s) => s.voiceEnabled);
  useAutosize(taRef, value);
  const inferredMode = useMemo(() => inferCaretModeForPrompt(value, mode), [mode, value]);
  const showModeHint = value.trim().length > 0 && inferredMode !== mode && !isStreaming;

  const focusComposer = useCallback((): void => {
    const el = taRef.current;
    if (!el) return;
    el.focus();
    const cursor = el.value.length;
    el.setSelectionRange(cursor, cursor);
  }, []);

  useEffect(() => {
    if (autoFocus) focusComposer();
  }, [autoFocus, focusComposer]);

  useEffect(() => {
    window.addEventListener(CARET_FOCUS_COMPOSER_EVENT, focusComposer);
    return (): void => window.removeEventListener(CARET_FOCUS_COMPOSER_EVENT, focusComposer);
  }, [focusComposer]);

  const send = useCallback((): void => {
    const trimmed = value.trim();
    if (!trimmed || isStreaming) return;
    onSubmit(trimmed);
    setValue('');
  }, [value, isStreaming, onSubmit]);

  const appendTranscript = useCallback((text: string): void => {
    const normalized = text.trim();
    if (!normalized) return;
    setValue((prev) => (prev.length && !prev.endsWith(' ') ? `${prev} ` : prev) + normalized);
  }, []);

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
          lineHeight: '20px',
          padding: '4px 6px',
          fontFamily: 'inherit',
          height: 28,
          minHeight: 28,
          maxHeight: 140,
          overflowY: 'auto',
          boxSizing: 'border-box',
          display: 'block',
          flex: 'none',
        }}
      />
      {showModeHint && (
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            alignSelf: 'flex-start',
            gap: 5,
            maxWidth: '100%',
            marginLeft: 4,
            padding: '3px 7px',
            border: `1px solid ${caretTokens.border}`,
            borderRadius: caretRadius.sm,
            background: 'rgba(255, 255, 255, 0.02)',
            color: caretTokens.textMuted,
            fontSize: 11.5,
            lineHeight: '15px',
          }}
        >
          <span style={{ color: caretTokens.textDim }}>Mode</span>
          <span
            style={{
              color: caretTokens.text,
              fontWeight: 600,
            }}
          >
            {CARET_MODES[inferredMode].label}
          </span>
        </div>
      )}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
        }}
      >
        <ModePicker />
        <span style={{ flex: 1 }} />
        {onAttach && <AttachButton onFiles={onAttach} />}
        {voiceEnabled && <MicButton onTranscript={appendTranscript} />}
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
