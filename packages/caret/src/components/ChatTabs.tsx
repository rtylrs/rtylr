import React, { useRef } from 'react';
import { LoaderCircle, X } from 'lucide-react';
import type { CaretTab } from '../stores/caretStore';
import { caretTokens } from '../ui/tokens';
import IconButton from '../ui/IconButton';

export interface ChatTabsProps {
  tabs: CaretTab[];
  activeTabId: string;
  onSelect: (id: string) => void;
  onClose: (id: string) => void;
}

/** The open-chat strip. Each tab owns a separate Caret session and stream. */
export const ChatTabs: React.FC<ChatTabsProps> = ({ tabs, activeTabId, onSelect, onClose }) => {
  const tabRefs = useRef<Record<string, HTMLButtonElement | null>>({});

  const focusTabElement = (id: string): void => {
    const focus = (): void => tabRefs.current[id]?.focus();
    focus();
    // The active chat remounts its composer, whose initial focus runs after
    // selection. Restore focus to the tab so arrow-key navigation stays in
    // the tablist as required by the tabs interaction pattern.
    window.setTimeout(focus, 0);
    window.setTimeout(focus, 50);
  };

  const focusTab = (id: string): void => {
    onSelect(id);
    focusTabElement(id);
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLButtonElement>, index: number): void => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onSelect(tabs[index].id);
      return;
    }

    if (event.key === 'Delete' && tabs.length > 1) {
      event.preventDefault();
      const nextTab = tabs[index + 1] ?? tabs[index - 1];
      onClose(tabs[index].id);
      focusTabElement(nextTab.id);
      return;
    }

    let nextIndex: number | null = null;
    if (event.key === 'ArrowRight') nextIndex = (index + 1) % tabs.length;
    if (event.key === 'ArrowLeft') nextIndex = (index - 1 + tabs.length) % tabs.length;
    if (event.key === 'Home') nextIndex = 0;
    if (event.key === 'End') nextIndex = tabs.length - 1;
    if (nextIndex === null) return;

    event.preventDefault();
    focusTab(tabs[nextIndex].id);
  };

  return (
    <div role="tablist" aria-label="Open Caret chats" className="caret-tabs-scroll">
      {tabs.map((tab, index) => {
        const active = tab.id === activeTabId;
        return (
          <div key={tab.id} className={active ? 'caret-tab-shell is-active' : 'caret-tab-shell'}>
            <button
              ref={(element) => {
                tabRefs.current[tab.id] = element;
              }}
              id={`caret-chat-tab-${tab.id}`}
              type="button"
              role="tab"
              aria-controls={`caret-chat-panel-${tab.id}`}
              aria-selected={active}
              aria-label={tab.title}
              tabIndex={active ? 0 : -1}
              className="caret-tab-button"
              onClick={() => onSelect(tab.id)}
              onKeyDown={(event) => handleKeyDown(event, index)}
            >
              <span className="caret-tab-title">{tab.title}</span>
              {tab.isLoading ? (
                <LoaderCircle
                  className="caret-tab-spinner"
                  aria-label="Loading chat"
                  size={12}
                  style={{ color: caretTokens.textMuted, flexShrink: 0 }}
                />
              ) : tab.isStreaming ? (
                <span
                  className="caret-tab-status"
                  role="img"
                  aria-label="Chat is running"
                  title="Chat is running"
                  style={{
                    width: 6,
                    height: 6,
                    borderRadius: '50%',
                    background: caretTokens.accent,
                    boxShadow: `0 0 0 3px ${caretTokens.accentSoft}`,
                    flexShrink: 0,
                  }}
                />
              ) : null}
            </button>
            {tabs.length > 1 && (
              <IconButton
                className="caret-tab-close"
                size="sm"
                label={`Close ${tab.title}`}
                title={undefined}
                onClick={(event) => {
                  event.stopPropagation();
                  onClose(tab.id);
                }}
              >
                <X size={11} />
              </IconButton>
            )}
          </div>
        );
      })}
    </div>
  );
};

export default ChatTabs;
