import React from 'react';
import type { CaretMessage as CaretMessageType } from '../stores/caretStore';

interface CaretMessageProps {
  message: CaretMessageType;
  isStreaming?: boolean;
}

const CaretMessage = ({ message, isStreaming = false }: CaretMessageProps): React.JSX.Element => {
  const isUser = message.role === 'user';

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: isUser ? 'flex-end' : 'flex-start',
        marginBottom: '12px',
      }}
    >
      <div
        style={{
          maxWidth: '80%',
          padding: '10px 14px',
          borderRadius: isUser ? '18px 18px 4px 18px' : '18px 18px 18px 4px',
          background: isUser ? 'var(--caret-accent, #8b5cf6)' : 'var(--caret-card, #1e1e1e)',
          color: '#ffffff',
          border: isUser ? 'none' : '1px solid var(--caret-border, #2a2a2a)',
          fontSize: '14px',
          lineHeight: '1.6',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
        }}
      >
        {message.content}
        {isStreaming && (
          <span
            style={{
              display: 'inline-block',
              width: '2px',
              height: '14px',
              background: 'var(--caret-accent, #8b5cf6)',
              marginLeft: '2px',
              verticalAlign: 'text-bottom',
              animation: 'caret-blink 1s step-end infinite',
            }}
            aria-hidden="true"
          />
        )}
      </div>
      <style>{`
        @keyframes caret-blink {
          0%, 100% { opacity: 1; }
          50% { opacity: 0; }
        }
      `}</style>
    </div>
  );
};

export default CaretMessage;
