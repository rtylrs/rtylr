import React, { useRef, useEffect, useState, useCallback } from 'react';
import { Sparkles, X, Send } from 'lucide-react';
import { useCaretStore } from '../stores/caretStore';
import { useCaretStream } from '../hooks/useCaretStream';
import CaretMessage from './CaretMessage';

export interface CaretPanelProps {
  appId: string;
  systemContext?: string;
  userId: number;
  apiBaseUrl: string;
}

const SUGGESTIONS = [
  'How much revenue today?',
  'Which products are low on stock?',
  'Show my top 5 customers',
  'How many staff are clocked in?',
];

const CaretPanel = ({ systemContext, userId, apiBaseUrl }: CaretPanelProps): React.JSX.Element => {
  const { isOpen, close, messages, isStreaming } = useCaretStore();
  const { sendMessage } = useCaretStream({ userId, apiBaseUrl, systemContext });
  const [input, setInput] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isStreaming]);

  useEffect(() => {
    if (isOpen) {
      inputRef.current?.focus();
    }
  }, [isOpen]);

  const handleSubmit = useCallback(
    (text: string): void => {
      const trimmed = text.trim();
      if (!trimmed || isStreaming) return;
      setInput('');
      void sendMessage(trimmed);
    },
    [isStreaming, sendMessage]
  );

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>): void => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(input);
    }
  };

  const handleSuggestion = (suggestion: string): void => {
    handleSubmit(suggestion);
  };

  return (
    <div
      role="complementary"
      aria-label="Caret AI assistant"
      style={
        {
          '--caret-accent': '#8b5cf6',
          '--caret-card': '#1e1e1e',
          '--caret-border': '#2a2a2a',
          '--caret-text': '#ffffff',
          '--caret-text-muted': '#888888',
          '--caret-input-bg': '#2a2a2a',

          position: 'fixed',
          top: 0,
          right: 0,
          width: '380px',
          height: '100dvh',
          zIndex: 9999,
          background: '#141414',
          borderLeft: '1px solid #2a2a2a',
          display: 'flex',
          flexDirection: 'column',
          transform: isOpen ? 'translateX(0)' : 'translateX(100%)',
          transition: 'transform 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
          willChange: 'transform',
        } as React.CSSProperties
      }
    >
      {/* Header */}
      <div
        style={{
          height: '48px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 16px',
          borderBottom: '1px solid #2a2a2a',
          flexShrink: 0,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Sparkles size={16} style={{ color: 'var(--caret-accent)' }} aria-hidden="true" />
          <span
            style={{
              color: 'var(--caret-text)',
              fontSize: '14px',
              fontWeight: 600,
            }}
          >
            Caret
          </span>
        </div>
        <button
          onClick={close}
          aria-label="Close Caret"
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '28px',
            height: '28px',
            borderRadius: '6px',
            background: 'transparent',
            border: 'none',
            color: 'var(--caret-text-muted)',
            cursor: 'pointer',
            transition: 'background 0.15s, color 0.15s',
          }}
          onMouseEnter={(e): void => {
            (e.currentTarget as HTMLButtonElement).style.background = '#2a2a2a';
            (e.currentTarget as HTMLButtonElement).style.color = '#ffffff';
          }}
          onMouseLeave={(e): void => {
            (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
            (e.currentTarget as HTMLButtonElement).style.color = 'var(--caret-text-muted)';
          }}
        >
          <X size={15} aria-hidden="true" />
        </button>
      </div>

      {/* Messages area */}
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '16px',
          display: 'flex',
          flexDirection: 'column',
        }}
        aria-live="polite"
        aria-atomic="false"
      >
        {messages.length === 0 ? (
          <div
            style={{
              flex: 1,
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              padding: '24px 0',
            }}
          >
            <div
              style={{
                width: '44px',
                height: '44px',
                borderRadius: '12px',
                background: 'rgba(139, 92, 246, 0.12)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                marginBottom: '12px',
              }}
            >
              <Sparkles size={20} style={{ color: 'var(--caret-accent)' }} aria-hidden="true" />
            </div>
            <p
              style={{
                color: 'var(--caret-text)',
                fontSize: '14px',
                fontWeight: 600,
                margin: '0 0 4px',
              }}
            >
              Ask Caret anything
            </p>
            <p
              style={{
                color: 'var(--caret-text-muted)',
                fontSize: '13px',
                margin: '0 0 20px',
                textAlign: 'center',
              }}
            >
              Your AI assistant with access to all your business data.
            </p>
            <div
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: '8px',
                width: '100%',
              }}
            >
              {SUGGESTIONS.map((suggestion) => (
                <button
                  key={suggestion}
                  onClick={(): void => handleSuggestion(suggestion)}
                  style={{
                    padding: '8px 12px',
                    borderRadius: '8px',
                    background: '#1e1e1e',
                    border: '1px solid #2a2a2a',
                    color: 'var(--caret-text)',
                    fontSize: '13px',
                    textAlign: 'left',
                    cursor: 'pointer',
                    transition: 'border-color 0.15s',
                  }}
                  onMouseEnter={(e): void => {
                    (e.currentTarget as HTMLButtonElement).style.borderColor =
                      'rgba(139,92,246,0.5)';
                  }}
                  onMouseLeave={(e): void => {
                    (e.currentTarget as HTMLButtonElement).style.borderColor = '#2a2a2a';
                  }}
                >
                  {suggestion}
                </button>
              ))}
            </div>
          </div>
        ) : (
          <>
            {messages.map((msg, index) => {
              const isLastAssistant =
                index === messages.length - 1 && msg.role === 'assistant' && isStreaming;
              return <CaretMessage key={msg.id} message={msg} isStreaming={isLastAssistant} />;
            })}
            <div ref={bottomRef} />
          </>
        )}
      </div>

      {/* Input area */}
      <div
        style={{
          height: '72px',
          display: 'flex',
          alignItems: 'center',
          padding: '0 12px',
          borderTop: '1px solid #2a2a2a',
          gap: '8px',
          flexShrink: 0,
        }}
      >
        <input
          ref={inputRef}
          value={input}
          onChange={(e): void => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask Caret anything…"
          aria-label="Message Caret"
          disabled={isStreaming}
          style={{
            flex: 1,
            height: '40px',
            padding: '0 12px',
            borderRadius: '8px',
            background: 'var(--caret-input-bg)',
            border: '1px solid #3a3a3a',
            color: 'var(--caret-text)',
            fontSize: '16px',
            outline: 'none',
            transition: 'border-color 0.15s',
          }}
          onFocus={(e): void => {
            (e.target as HTMLInputElement).style.borderColor = 'rgba(139,92,246,0.5)';
          }}
          onBlur={(e): void => {
            (e.target as HTMLInputElement).style.borderColor = '#3a3a3a';
          }}
        />
        <button
          onClick={(): void => handleSubmit(input)}
          disabled={!input.trim() || isStreaming}
          aria-label="Send message"
          style={{
            width: '36px',
            height: '36px',
            borderRadius: '8px',
            background:
              input.trim() && !isStreaming ? 'var(--caret-accent)' : 'rgba(139,92,246,0.2)',
            border: 'none',
            color: '#ffffff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: input.trim() && !isStreaming ? 'pointer' : 'not-allowed',
            flexShrink: 0,
            transition: 'background 0.15s',
          }}
        >
          <Send size={15} aria-hidden="true" />
        </button>
      </div>
    </div>
  );
};

export default CaretPanel;
