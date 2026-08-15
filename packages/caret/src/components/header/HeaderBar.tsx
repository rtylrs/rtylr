import React from 'react';
import { ArrowLeft, ChevronDown, Maximize2, Minimize2, Settings, SquarePen, X } from 'lucide-react';
import { caretTokens } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';
import Tooltip from '../../ui/Tooltip';
import { useCaretStore } from '../../stores/caretStore';
import { useCaretUIStore } from '../../stores/uiStore';
import HistoryHover from './HistoryHover';

export interface HeaderBarProps {
  onNewChat?: () => void;
  /** Hide the close button - used in 'fullscreen' variant where there's nothing to close. */
  hideClose?: boolean;
  /** Current conversation title, shown next to a history shortcut. */
  title?: string;
  /** Whether the overlay is maximized to full-screen (with the chat sidebar). */
  expanded?: boolean;
  /** Present only for the 'overlay' variant - toggles between floating and maximized. */
  onToggleExpand?: () => void;
}

export const HeaderBar: React.FC<HeaderBarProps> = ({
  onNewChat,
  hideClose = false,
  title,
  expanded,
  onToggleExpand,
}) => {
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
        <>
          {title && (
            <button
              type="button"
              onClick={() => setView('history')}
              className="caret-chrome-button"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 4,
                border: 'none',
                background: 'transparent',
                color: caretTokens.textMuted,
                fontSize: 13,
                fontWeight: 600,
                cursor: 'pointer',
                padding: '4px 6px',
                borderRadius: 6,
              }}
            >
              {title}
              <ChevronDown size={13} />
            </button>
          )}
          {onNewChat && (
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
          )}
        </>
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

      {onToggleExpand && (
        <Tooltip label={expanded ? 'Restore' : 'Maximize'}>
          <IconButton
            className="caret-chrome-button"
            size="md"
            label={expanded ? 'Restore' : 'Maximize'}
            onClick={onToggleExpand}
          >
            {expanded ? <Minimize2 size={13} /> : <Maximize2 size={13} />}
          </IconButton>
        </Tooltip>
      )}

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

      {!hideClose && (
        <Tooltip label="Close">
          <IconButton className="caret-chrome-button" size="md" label="Close Caret" onClick={close}>
            <X size={15} />
          </IconButton>
        </Tooltip>
      )}
    </div>
  );
};

export default HeaderBar;
