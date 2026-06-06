import React, { useCallback, useEffect, useState } from 'react';
import '../styles/index.css';
import { caretTokens } from '../ui/tokens';
import { useCaretStore } from '../stores/caretStore';
import { useCaretSettingsStore } from '../stores/settingsStore';
import { useCaretUIStore } from '../stores/uiStore';
import { useChatsStore } from '../stores/chatsStore';
import { useCaretStream } from '../hooks/useCaretStream';
import { useCaretSessions } from '../hooks/useCaretSessions';
import { useLoadCaretChat } from '../hooks/useLoadCaretChat';
import { usePanelResize } from '../hooks/usePanelResize';
import { useShortcuts } from '../hooks/useShortcuts';
import HeaderBar from './header/HeaderBar';
import MessageList from './MessageList';
import { Composer } from './composer';
import EmptyState from './EmptyState';
import SettingsPanel from './SettingsPanel';
import { ChatList } from './sidebar';

export interface CaretPanelProps {
  appId: string;
  systemContext?: string;
  userId: number;
  apiBaseUrl: string;
  /** 'inline' renders inside a flex column the parent controls. */
  variant?: 'inline';
}

const DEFAULT_SUGGESTIONS = [
  'How much revenue today?',
  'Which products are low on stock?',
  'Show my top 5 customers',
  'How many staff are clocked in?',
];

const CaretPanel: React.FC<CaretPanelProps> = ({ systemContext, userId, apiBaseUrl, variant }) => {
  useShortcuts();
  const { isOpen, sessionId, messages, isStreaming, setToolStatus, setSessionId, setMessages, reset } =
    useCaretStore();
  const showThinking = useCaretSettingsStore((s) => s.showThinking);
  const panelWidth = useCaretSettingsStore((s) => s.panelWidth);
  const view = useCaretUIStore((s) => s.view);
  const setView = useCaretUIStore((s) => s.setView);

  const { sendMessage, stopStreaming } = useCaretStream({
    userId,
    apiBaseUrl,
    systemContext,
  });
  const { sessions } = useCaretSessions({ userId, apiBaseUrl });
  const loadChat = useLoadCaretChat({ userId, apiBaseUrl });
  const setRecent = useChatsStore((s) => s.setRecent);
  const { onMouseDown } = usePanelResize();

  useEffect(() => {
    if (sessions.length > 0) setRecent(sessions);
  }, [sessions, setRecent]);

  const newChat = useCallback((): void => {
    reset();
    setView('chat');
  }, [reset, setView]);

  const selectChat = useCallback(
    async (id: string): Promise<void> => {
      setSessionId(id);
      setView('chat');
      try {
        const loaded = await loadChat(id);
        setMessages(loaded);
      } catch {
        setMessages([]);
      }
    },
    [loadChat, setSessionId, setMessages, setView],
  );

  const [introPlaying, setIntroPlaying] = useState(false);
  useEffect(() => {
    if (!isOpen) return;
    try {
      if (localStorage.getItem('caret:intro-shown') === '1') return;
      localStorage.setItem('caret:intro-shown', '1');
    } catch {
      return;
    }
    setIntroPlaying(true);
    const timer = window.setTimeout(() => setIntroPlaying(false), 1500);
    return () => window.clearTimeout(timer);
  }, [isOpen]);

  const onApprove = (id: string): void => setToolStatus(id, 'approved');
  const onDeny = (id: string): void => setToolStatus(id, 'denied');

  const rootStyle: React.CSSProperties =
    variant === 'inline'
      ? {
          position: 'relative',
          width: '100%',
          height: '100%',
          background: caretTokens.bg,
          borderLeft: `1px solid ${caretTokens.border}`,
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }
      : {
          position: 'fixed',
          top: 0,
          right: 0,
          width: panelWidth,
          height: '100dvh',
          zIndex: 9999,
          background: caretTokens.bg,
          borderLeft: `1px solid ${caretTokens.border}`,
          display: 'flex',
          flexDirection: 'column',
          transform: isOpen ? 'translateX(0)' : 'translateX(100%)',
          transition: 'transform 0.25s cubic-bezier(0.4, 0, 0.2, 1)',
          willChange: 'transform',
        };

  return (
    <aside
      role="complementary"
      aria-label="Caret AI assistant"
      className={introPlaying ? 'caret-intro' : undefined}
      style={rootStyle}
    >
      <span
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize Caret panel"
        onMouseDown={onMouseDown}
        style={{
          position: 'absolute',
          top: 0,
          left: -3,
          width: 6,
          height: '100%',
          cursor: 'col-resize',
          zIndex: 10,
        }}
      />

      <HeaderBar onNewChat={newChat} onSelectChat={selectChat} />

      {view === 'chat' && (
        <>
          {messages.length === 0 ? (
            <EmptyState suggestions={DEFAULT_SUGGESTIONS} onSuggestion={(s) => sendMessage(s)} />
          ) : (
            <MessageList
              messages={messages}
              isStreaming={isStreaming}
              showThinking={showThinking}
              onApproveTool={onApprove}
              onDenyTool={onDeny}
            />
          )}
          <Composer
            isStreaming={isStreaming}
            onSubmit={(t) => sendMessage(t)}
            onStop={stopStreaming}
          />
        </>
      )}

      {view === 'history' && (
        <ChatList
          chats={sessions}
          activeId={sessionId}
          onSelect={selectChat}
          onNew={newChat}
        />
      )}

      {view === 'settings' && <SettingsPanel />}
    </aside>
  );
};

export default CaretPanel;
