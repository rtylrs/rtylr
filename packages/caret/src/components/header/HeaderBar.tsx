import React from 'react';
import { ArrowLeft, Settings, SquarePen, X } from 'lucide-react';
import { caretTokens } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useCaretStore } from '../../stores/caretStore';
import { useCaretUIStore } from '../../stores/uiStore';
import HistoryHover from './HistoryHover';

export interface HeaderBarProps {
  onNewChat?: () => void;
}

export const HeaderBar: React.FC<HeaderBarProps> = ({ onNewChat }) => {
  const close = useCaretStore((s) => s.close);
  const view = useCaretUIStore((s) => s.view);
  const setView = useCaretUIStore((s) => s.setView);

  return (
    <div
      className="caret-header-bar"
      style={{
        height: 44,
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        padding: '0 10px',
        background: caretTokens.bg,
        flexShrink: 0,
      }}
    >
      {view === 'chat' ? (
        onNewChat && (
          <Tooltip label="New chat">
            <IconButton
              className="caret-chrome-button"
              size="md"
              label="New chat"
              onClick={onNewChat}
            >
              <SquarePen size={15} />
            </IconButton>
          </Tooltip>
        )
      ) : (
        <Tooltip label="Back to chat">
          <IconButton
            className="caret-chrome-button"
            size="md"
            label="Back to chat"
            onClick={() => setView('chat')}
          >
            <ArrowLeft size={15} />
          </IconButton>
        </Tooltip>
      )}

      <span style={{ flex: 1 }} />

      <HistoryHover />

      <Tooltip label="Settings">
        <IconButton
          className="caret-chrome-button"
          size="md"
          label="Settings"
          active={view === 'settings'}
          onClick={() => setView('settings')}
        >
          <Settings size={13} />
        </IconButton>
      </Tooltip>

      <Tooltip label="Close">
        <IconButton className="caret-chrome-button" size="md" label="Close Caret" onClick={close}>
          <X size={15} />
        </IconButton>
      </Tooltip>
    </div>
  );
};

export default HeaderBar;
