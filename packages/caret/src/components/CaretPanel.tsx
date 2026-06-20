import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
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
import { authHeaders } from '../lib/auth';
import { buildCaretEmptyState } from '../lib/suggestions';
import type { CaretChatSummary } from '../types/chat';
import type { CaretPageContext } from '../types/context';
import type { ToolCallBlock } from '../types/message';
import HeaderBar from './header/HeaderBar';
import MessageList from './MessageList';
import { Composer } from './composer';
import EmptyState from './EmptyState';
import SettingsPanel from './SettingsPanel';
import { ChatList } from './sidebar';

export interface CaretPanelProps {
  appId: string;
  systemContext?: string;
  pageContext?: CaretPageContext;
  userId: number;
  apiBaseUrl: string;
  /** 'inline' renders inside a flex column the parent controls. */
  variant?: 'inline';
}

type ToolDecision = 'approve' | 'deny';

interface ToolDecisionResponse {
  tool_call_id: string;
  status: string;
  message: string;
  executed: boolean;
  preview?: string;
  output?: unknown;
}

function normalizeToolDecisionStatus(status: string): ToolCallBlock['status'] {
  switch (status) {
    case 'pending':
    case 'approved':
    case 'denied':
    case 'running':
    case 'success':
    case 'error':
      return status;
    default:
      return 'error';
  }
}

function outputMessage(output: unknown): string | null {
  if (!output || typeof output !== 'object') return null;
  const message = (output as { message?: unknown }).message;
  return typeof message === 'string' && message.trim() ? message.trim() : null;
}

function cleanDecisionMessage(message: string | null | undefined): string {
  const text = (message ?? '').trim();
  if (!text || text === 'Tool executed successfully.') return 'the action is complete.';
  return text.replace(/\.$/, '') + '.';
}

function decisionFollowUp(
  decision: ToolDecision,
  status: ToolCallBlock['status'],
  result?: ToolDecisionResponse
): string | null {
  if (decision === 'deny' || status === 'denied') {
    return "Okay, I didn't make any changes.";
  }
  if (status === 'success' && result?.executed) {
    return `Done. ${cleanDecisionMessage(outputMessage(result.output) ?? result.message)}`;
  }
  if (status === 'error') {
    return "I couldn't finish that. No confirmed changes were made.";
  }
  return null;
}

const CaretPanel: React.FC<CaretPanelProps> = ({
  appId,
  systemContext,
  pageContext,
  userId,
  apiBaseUrl,
  variant,
}) => {
  useShortcuts();
  const {
    isOpen,
    sessionId,
    messages,
    isStreaming,
    setToolStatus,
    appendBlockToLastAssistant,
    setSessionId,
    setMessages,
    reset,
  } = useCaretStore();
  const showThinking = useCaretSettingsStore((s) => s.showThinking);
  const panelWidth = useCaretSettingsStore((s) => s.panelWidth);
  const view = useCaretUIStore((s) => s.view);
  const setView = useCaretUIStore((s) => s.setView);
  const queryClient = useQueryClient();
  const mode = useCaretSettingsStore((s) => s.mode);

  const { sendMessage, stopStreaming } = useCaretStream({
    userId,
    apiBaseUrl,
    systemContext,
    pageContext,
  });
  const { sessions } = useCaretSessions({ userId, apiBaseUrl });
  const loadChat = useLoadCaretChat({ userId, apiBaseUrl });
  const setRecent = useChatsStore((s) => s.setRecent);
  const removeRecent = useChatsStore((s) => s.removeRecent);
  const { onMouseDown } = usePanelResize();
  const emptyState = useMemo(
    () => buildCaretEmptyState({ appId, pageContext }),
    [appId, pageContext]
  );

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
    [loadChat, setSessionId, setMessages, setView]
  );

  const deleteChat = useCallback(
    async (id: string): Promise<void> => {
      const queryKey = ['caret', 'sessions', userId] as const;
      const previous = queryClient.getQueryData<CaretChatSummary[]>(queryKey);

      removeRecent(id);
      queryClient.setQueryData<CaretChatSummary[]>(queryKey, (old) =>
        old ? old.filter((chat) => chat.id !== id) : []
      );

      const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
      const response = await fetch(`${baseUrl}memory/chats/${encodeURIComponent(id)}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: authHeaders(),
      }).catch(() => null);

      if (!response || (!response.ok && response.status !== 404)) {
        if (previous) {
          queryClient.setQueryData<CaretChatSummary[]>(queryKey, previous);
          setRecent(previous);
        }
        return;
      }

      if (sessionId === id) {
        reset();
        setView('chat');
      }

      void queryClient.invalidateQueries({ queryKey });
    },
    [apiBaseUrl, queryClient, removeRecent, reset, sessionId, setRecent, setView, userId]
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

  const findToolCall = useCallback(
    (id: string): ToolCallBlock | null => {
      for (const message of messages) {
        for (const block of message.blocks) {
          if (block.type === 'tool_call' && block.id === id) return block;
        }
      }
      return null;
    },
    [messages]
  );

  const sendToolDecision = useCallback(
    async (id: string, decision: ToolDecision): Promise<void> => {
      const call = findToolCall(id);
      if (!call) {
        setToolStatus(id, 'error', { error: 'Tool call no longer exists.' });
        return;
      }

      const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
      setToolStatus(id, 'running');

      try {
        const response = await fetch(`${baseUrl}caret/tools/decision`, {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            ...authHeaders(),
          },
          body: JSON.stringify({
            tool_call_id: id,
            name: call.name,
            input: call.input ?? {},
            decision,
            mode: call.mode ?? mode,
          }),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const result = (await response.json()) as ToolDecisionResponse;
        const status = normalizeToolDecisionStatus(result.status);
        const extras: Partial<ToolCallBlock> =
          status === 'error'
            ? { error: result.message || 'Tool decision failed.', preview: result.preview }
            : {
                output: result.output ?? { message: result.message, executed: result.executed },
                preview: result.preview,
              };
        setToolStatus(id, status, extras);
        const followUp = decisionFollowUp(decision, status, result);
        if (followUp) {
          appendBlockToLastAssistant({
            id: `tool-followup-${id}-${Date.now()}`,
            type: 'text',
            content: followUp,
          });
        }
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Tool decision failed.';
        setToolStatus(id, 'error', { error: message });
        appendBlockToLastAssistant({
          id: `tool-followup-${id}-${Date.now()}`,
          type: 'text',
          content: "I couldn't finish that. No confirmed changes were made.",
        });
      }
    },
    [apiBaseUrl, appendBlockToLastAssistant, findToolCall, mode, setToolStatus]
  );

  const onApprove = useCallback(
    (id: string): void => {
      void sendToolDecision(id, 'approve');
    },
    [sendToolDecision]
  );
  const onDeny = useCallback(
    (id: string): void => {
      void sendToolDecision(id, 'deny');
    },
    [sendToolDecision]
  );

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

      <HeaderBar onNewChat={newChat} onSelectChat={selectChat} onDeleteChat={deleteChat} />

      {view === 'chat' && (
        <>
          {messages.length === 0 ? (
            <EmptyState
              title={emptyState.title}
              subtitle={emptyState.subtitle}
              suggestions={emptyState.suggestions}
              onSuggestion={(s) => sendMessage(s)}
            />
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
          onDelete={deleteChat}
        />
      )}

      {view === 'settings' && <SettingsPanel />}
    </aside>
  );
};

export default CaretPanel;
