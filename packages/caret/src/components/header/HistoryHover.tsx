import React, { useRef, useState } from 'react';
import { History, MessageSquare, Trash2 } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useChatsStore } from '../../stores/chatsStore';
import { useCaretUIStore } from '../../stores/uiStore';

export interface HistoryHoverProps {
  onSelectChat?: (id: string) => void;
  onDeleteChat?: (id: string) => void | Promise<void>;
}

/** History icon button that reveals a hover dropdown of recent chats. */
export const HistoryHover: React.FC<HistoryHoverProps> = ({ onSelectChat, onDeleteChat }) => {
  const recent = useChatsStore((s) => s.recent);
  const setActive = useChatsStore((s) => s.setActiveChat);
  const setView = useCaretUIStore((s) => s.setView);
  const view = useCaretUIStore((s) => s.view);
  const [open, setOpen] = useState(false);
  const closeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const enter = (): void => {
    if (closeTimer.current) {
      clearTimeout(closeTimer.current);
      closeTimer.current = null;
    }
    setOpen(true);
  };
  const leave = (): void => {
    closeTimer.current = setTimeout(() => setOpen(false), 150);
  };

  return (
    <span
      style={{ position: 'relative', display: 'inline-flex' }}
      onMouseEnter={enter}
      onMouseLeave={leave}
    >
      <Tooltip label="History">
        <IconButton
          size="md"
          label="History"
          active={view === 'history'}
          onClick={() => setView('history')}
        >
          <History size={13} />
        </IconButton>
      </Tooltip>

      {open && recent.length > 0 && (
        <div
          role="menu"
          style={{
            position: 'absolute',
            top: 'calc(100% + 4px)',
            right: 0,
            minWidth: 240,
            maxHeight: 280,
            overflowY: 'auto',
            background: caretTokens.surface,
            border: `1px solid ${caretTokens.border}`,
            borderRadius: caretRadius.lg,
            padding: 4,
            zIndex: 200,
            boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
          }}
        >
          <div
            style={{
              padding: '4px 8px',
              fontSize: 10,
              fontWeight: 600,
              color: caretTokens.textDim,
              textTransform: 'uppercase',
              letterSpacing: 0.4,
            }}
          >
            Recent chats
          </div>
          {recent.map((c) => (
            <div
              key={c.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 2,
                width: '100%',
                padding: 0,
                background: 'transparent',
                borderRadius: caretRadius.sm,
              }}
            >
              <button
                type="button"
                onClick={() => {
                  setActive(c.id);
                  if (onSelectChat) onSelectChat(c.id);
                  else setView('chat');
                  setOpen(false);
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 6,
                  flex: 1,
                  minWidth: 0,
                  padding: '6px 8px',
                  background: 'transparent',
                  border: 'none',
                  borderRadius: caretRadius.sm,
                  color: caretTokens.text,
                  fontSize: 12.5,
                  cursor: 'pointer',
                  textAlign: 'left',
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = caretTokens.surfaceHigh;
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
                }}
              >
                <MessageSquare size={11} style={{ color: caretTokens.textDim, flexShrink: 0 }} />
                <span
                  style={{
                    flex: 1,
                    whiteSpace: 'nowrap',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                  }}
                >
                  {c.title || 'New chat'}
                </span>
              </button>
              {onDeleteChat && (
                <IconButton
                  size="sm"
                  label="Delete chat"
                  onClick={() => {
                    void onDeleteChat(c.id);
                  }}
                >
                  <Trash2 size={11} />
                </IconButton>
              )}
            </div>
          ))}
        </div>
      )}
    </span>
  );
};

export default HistoryHover;
