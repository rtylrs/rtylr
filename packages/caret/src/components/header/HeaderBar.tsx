import React from 'react';
import { Settings, X, Plus } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useCaretStore } from '../../stores/caretStore';
import { useCaretUIStore } from '../../stores/uiStore';
import CaretLogo from '../CaretLogo';
import HistoryHover from './HistoryHover';

export interface HeaderBarProps {
  onNewChat?: () => void;
  onSelectChat?: (id: string) => void;
  onDeleteChat?: (id: string) => void | Promise<void>;
}

export const HeaderBar: React.FC<HeaderBarProps> = ({ onNewChat, onSelectChat, onDeleteChat }) => {
  const close = useCaretStore((s) => s.close);
  const view = useCaretUIStore((s) => s.view);
  const setView = useCaretUIStore((s) => s.setView);

  return (
    <div
      style={{
        height: 44,
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        padding: '0 8px',
        borderBottom: `1px solid ${caretTokens.border}`,
        flexShrink: 0,
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          padding: '0 4px',
        }}
      >
        <span
          style={{
            width: 22,
            height: 22,
            borderRadius: caretRadius.md,
            background: caretTokens.accentSoft,
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            overflow: 'hidden',
          }}
        >
          <CaretLogo variant="teal" size={18} scale={1.8} />
        </span>
        <span style={{ color: caretTokens.text, fontSize: 13, fontWeight: 600 }}>Caret</span>
      </div>

      <span style={{ flex: 1 }} />

      <Tooltip label="Chat">
        <IconButton size="md" label="Chat" active={view === 'chat'} onClick={() => setView('chat')}>
          <CaretLogo variant="teal" size={15} scale={1.85} />
        </IconButton>
      </Tooltip>

      <HistoryHover onSelectChat={onSelectChat} onDeleteChat={onDeleteChat} />

      <Tooltip label="Settings">
        <IconButton
          size="md"
          label="Settings"
          active={view === 'settings'}
          onClick={() => setView('settings')}
        >
          <Settings size={13} />
        </IconButton>
      </Tooltip>

      {onNewChat && (
        <Tooltip label="New chat">
          <IconButton size="md" label="New chat" onClick={onNewChat}>
            <Plus size={13} />
          </IconButton>
        </Tooltip>
      )}

      <Tooltip label="Close">
        <IconButton size="md" label="Close Caret" onClick={close}>
          <X size={13} />
        </IconButton>
      </Tooltip>
    </div>
  );
};

export default HeaderBar;
