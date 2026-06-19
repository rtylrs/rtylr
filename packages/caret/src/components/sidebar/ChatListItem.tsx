import React from 'react';
import { MessageSquare, Trash2, Pin, PinOff } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';
import type { CaretChatSummary } from '../../types/chat';

export interface ChatListItemProps {
  chat: CaretChatSummary;
  active?: boolean;
  onSelect: (id: string) => void;
  onTogglePin?: (id: string) => void;
  onDelete?: (id: string) => void | Promise<void>;
}

export const ChatListItem: React.FC<ChatListItemProps> = ({
  chat,
  active = false,
  onSelect,
  onTogglePin,
  onDelete,
}) => {
  return (
    <div
      className="caret-chat-row"
      onClick={() => onSelect(chat.id)}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '6px 8px',
        borderRadius: caretRadius.sm,
        background: active ? caretTokens.accentSoft : 'transparent',
        color: active ? caretTokens.text : caretTokens.textMuted,
        cursor: 'pointer',
        position: 'relative',
      }}
    >
      <MessageSquare size={13} style={{ flexShrink: 0 }} />
      <span
        style={{
          flex: 1,
          fontSize: 13,
          whiteSpace: 'nowrap',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {chat.title || 'New chat'}
      </span>
      <div
        className="caret-chat-actions"
        onClick={(e) => e.stopPropagation()}
        style={{ display: 'flex', gap: 2 }}
      >
        {onTogglePin && (
          <IconButton
            size="sm"
            label={chat.pinned ? 'Unpin chat' : 'Pin chat'}
            onClick={() => onTogglePin(chat.id)}
          >
            {chat.pinned ? <PinOff size={11} /> : <Pin size={11} />}
          </IconButton>
        )}
        {onDelete && (
          <IconButton
            size="sm"
            label="Delete chat"
            onClick={() => {
              void onDelete(chat.id);
            }}
          >
            <Trash2 size={11} />
          </IconButton>
        )}
      </div>
    </div>
  );
};

export default ChatListItem;
