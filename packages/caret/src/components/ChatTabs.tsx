import React from 'react';
import { Plus, X } from 'lucide-react';
import type { CaretTab } from '../stores/caretStore';
import { caretRadius, caretTokens } from '../ui/tokens';
import IconButton from '../ui/IconButton';

export interface ChatTabsProps {
  tabs: CaretTab[];
  activeTabId: string;
  onSelect: (id: string) => void;
  onNew: () => void;
  onClose: (id: string) => void;
}

function selectWithKeyboard(
  event: React.KeyboardEvent<HTMLDivElement>,
  onSelect: () => void
): void {
  if (event.key !== 'Enter' && event.key !== ' ') return;
  event.preventDefault();
  onSelect();
}

/** The open-chat strip. Each tab owns a separate Caret session and stream. */
export const ChatTabs: React.FC<ChatTabsProps> = ({
  tabs,
  activeTabId,
  onSelect,
  onNew,
  onClose,
}) => (
  <div
    role="tablist"
    aria-label="Open Caret chats"
    style={{
      display: 'flex',
      alignItems: 'stretch',
      gap: 4,
      minHeight: 40,
      padding: '4px 8px 0',
      borderBottom: `1px solid ${caretTokens.border}`,
      overflowX: 'auto',
      overflowY: 'hidden',
      flexShrink: 0,
      scrollbarWidth: 'thin',
    }}
  >
    {tabs.map((tab) => {
      const active = tab.id === activeTabId;
      return (
        <div
          key={tab.id}
          role="tab"
          aria-selected={active}
          aria-label={tab.title}
          tabIndex={active ? 0 : -1}
          onClick={() => onSelect(tab.id)}
          onKeyDown={(event) => selectWithKeyboard(event, () => onSelect(tab.id))}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 2,
            minWidth: 104,
            maxWidth: 190,
            minHeight: 36,
            padding: '0 3px 0 10px',
            border: `1px solid ${active ? caretTokens.borderStrong : 'transparent'}`,
            borderBottomColor: active ? caretTokens.bg : 'transparent',
            borderRadius: `${caretRadius.sm} ${caretRadius.sm} 0 0`,
            background: active ? caretTokens.bg : 'transparent',
            color: active ? caretTokens.text : caretTokens.textMuted,
            cursor: 'pointer',
            flex: '0 0 auto',
            boxSizing: 'border-box',
            outline: 'none',
          }}
        >
          <span
            style={{
              flex: 1,
              minWidth: 0,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              fontSize: 11.5,
              fontWeight: active ? 600 : 500,
            }}
            title={tab.title}
          >
            {tab.title}
          </span>
          {tab.isStreaming && (
            <span
              aria-label="Streaming"
              title="Streaming"
              style={{
                width: 5,
                height: 5,
                borderRadius: '50%',
                background: caretTokens.accent,
                flexShrink: 0,
              }}
            />
          )}
          {tabs.length > 1 && (
            <IconButton
              size="sm"
              label={`Close ${tab.title}`}
              onClick={(event) => {
                event.stopPropagation();
                onClose(tab.id);
              }}
              style={{
                color: caretTokens.textDim,
                width: 24,
                height: 24,
              }}
            >
              <X size={11} />
            </IconButton>
          )}
        </div>
      );
    })}
    <IconButton
      size="md"
      label="New chat tab"
      onClick={onNew}
      style={{
        marginLeft: 'auto',
        alignSelf: 'center',
        color: caretTokens.textMuted,
      }}
    >
      <Plus size={13} />
    </IconButton>
  </div>
);

export default ChatTabs;
