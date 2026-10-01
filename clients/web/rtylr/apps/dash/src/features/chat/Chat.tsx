import React, { useRef, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { resolveApiBase } from '@voxire/config';
import { cn, EmptyState } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import { Send, Sparkles } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import {
  MessageActions,
  MessageBlocks,
  StreamingDot,
  useCaretStore,
  useCaretStream,
  type CaretMessageType,
  type ContentBlock,
} from '@voxire/caret';
import { useUserStore } from '../../shared/stores/userStore';
import { buildDashPageContext, buildDashSystemContext } from '../../shared/lib/dashCaretContext';

const SUGGESTIONS = [
  'How much revenue did we make today?',
  'Which products are low on stock?',
  'Show me our top 5 customers',
  'How many staff are clocked in?',
];

function blocksForMessage(message: CaretMessageType): ContentBlock[] {
  if (message.blocks.length > 0) return message.blocks;
  if (!message.content) return [];
  return [{ id: `${message.id}-text`, type: 'text', content: message.content }];
}

function blocksToText(blocks: ContentBlock[]): string {
  return blocks
    .map((block) => {
      if (block.type === 'text') return block.content;
      if (block.type === 'thinking') return '';
      return `[tool:${block.name}]`;
    })
    .filter(Boolean)
    .join('\n');
}

/**
 * Copy/share text for a message. Falls back to the rendered blocks when
 * content is empty (streaming/partial states), so the copy action never
 * receives a blank string while the bubble shows block text.
 */
export function messageCopyText(message: CaretMessageType): string {
  return message.content || blocksToText(blocksForMessage(message));
}

const Chat: React.FC = () => {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const userId = Number(user?.id) || 0;
  const apiBaseUrl = resolveApiBase();
  const systemContext = React.useMemo(
    () =>
      buildDashSystemContext({
        pathname: '/chat',
        businessName: user?.businessName,
        apps: user?.effectiveAccess?.apps
          .filter((app) => app.licensed && app.hasAccess)
          .map((app) => app.appId),
      }),
    [user?.businessName, user?.effectiveAccess]
  );
  const pageContext = React.useMemo(
    () =>
      buildDashPageContext({
        pathname: '/chat',
        businessName: user?.businessName,
        apps: user?.effectiveAccess?.apps
          .filter((app) => app.licensed && app.hasAccess)
          .map((app) => app.appId),
      }),
    [user?.businessName, user?.effectiveAccess]
  );

  const { messages, isStreaming } = useCaretStore();
  const { sendMessage } = useCaretStream({ userId, apiBaseUrl, systemContext, pageContext });

  const [input, setInput] = React.useState('');
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isStreaming]);

  const send = useCallback(
    (text: string): void => {
      const trimmed = text.trim();
      if (!trimmed || isStreaming) return;
      setInput('');
      void sendMessage(trimmed);
    },
    [isStreaming, sendMessage]
  );

  const isEmpty = messages.length === 0;

  return (
    <>
      <HelmetContainer title="Chat · rtylr" />
      <div className="flex flex-col overflow-hidden -m-6" style={{ height: 'calc(100vh - 44px)' }}>
        {/* Messages or empty state */}
        <div className="flex-1 overflow-y-auto" aria-live="polite" aria-atomic="false">
          {isEmpty ? (
            <div className="flex flex-col items-center justify-center h-full px-6">
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="text-center max-w-md"
              >
                <EmptyState
                  icon={Sparkles}
                  title={t('chat.welcomeTitle')}
                  description={t('chat.welcomeDescription')}
                  action={{
                    label: t('chat.startChatting'),
                    onClick: () => inputRef.current?.focus(),
                  }}
                />

                <div className="grid grid-cols-2 gap-2 mt-6">
                  {SUGGESTIONS.map((s, i) => (
                    <motion.button
                      key={i}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.1 + i * 0.05 }}
                      onClick={() => send(s)}
                      className="text-left px-4 py-3 rounded-xl text-sm transition-all"
                      style={{
                        background: 'var(--dash-card)',
                        border: '1px solid var(--dash-border)',
                        color: 'var(--dash-text)',
                      }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.borderColor = 'rgba(69,131,196,0.4)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.borderColor = 'var(--dash-border)';
                      }}
                    >
                      {s}
                    </motion.button>
                  ))}
                </div>
              </motion.div>
            </div>
          ) : (
            <div className="max-w-2xl mx-auto px-6 py-8 space-y-6">
              <AnimatePresence>
                {messages.map((msg, index) => {
                  const isLastStreaming =
                    index === messages.length - 1 && msg.role === 'assistant' && isStreaming;
                  const blocks = blocksForMessage(msg);
                  const text = messageCopyText(msg);

                  return (
                    <motion.div
                      key={msg.id}
                      initial={{ opacity: 0, y: 6 }}
                      animate={{ opacity: 1, y: 0 }}
                      className={cn('flex', msg.role === 'user' ? 'justify-end' : 'justify-start')}
                    >
                      {msg.role === 'assistant' && (
                        <div
                          className="h-7 w-7 rounded-lg flex items-center justify-center shrink-0 mr-3 mt-1"
                          style={{ background: 'rgba(69,131,196,0.1)' }}
                        >
                          <Sparkles size={14} style={{ color: '#4583C4' }} />
                        </div>
                      )}
                      <div
                        className={cn(
                          'rounded-2xl px-4 py-3',
                          msg.role === 'user' ? 'max-w-[80%]' : 'min-w-0 flex-1'
                        )}
                        style={{
                          background: msg.role === 'user' ? '#4583C4' : 'var(--dash-card)',
                          color: msg.role === 'user' ? '#fff' : 'var(--dash-text)',
                          border: msg.role === 'user' ? 'none' : '1px solid var(--dash-border)',
                        }}
                      >
                        {msg.role === 'user' ? (
                          <p className="text-sm leading-relaxed whitespace-pre-wrap">{text}</p>
                        ) : (
                          <>
                            <MessageBlocks blocks={blocks} />
                            {isLastStreaming && <StreamingDot />}
                            {!isLastStreaming && blocks.some((block) => block.type === 'text') && (
                              <MessageActions text={text} />
                            )}
                          </>
                        )}
                      </div>
                    </motion.div>
                  );
                })}
              </AnimatePresence>

              {isStreaming && messages[messages.length - 1]?.role !== 'assistant' && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="flex items-center gap-3"
                >
                  <div
                    className="h-7 w-7 rounded-lg flex items-center justify-center shrink-0"
                    style={{ background: 'rgba(69,131,196,0.1)' }}
                  >
                    <Sparkles size={14} style={{ color: '#4583C4' }} />
                  </div>
                  <div
                    className="flex items-center gap-1.5 px-4 py-3 rounded-2xl"
                    style={{
                      background: 'var(--dash-card)',
                      border: '1px solid var(--dash-border)',
                    }}
                  >
                    {[0, 1, 2].map((i) => (
                      <motion.div
                        key={i}
                        className="h-1.5 w-1.5 rounded-full"
                        style={{ background: '#4583C4' }}
                        animate={{ opacity: [0.2, 1, 0.2] }}
                        transition={{ duration: 1, repeat: Infinity, delay: i * 0.15 }}
                      />
                    ))}
                  </div>
                </motion.div>
              )}
              <div ref={bottomRef} />
            </div>
          )}
        </div>

        {/* Input bar */}
        <div className="px-6 pb-6 pt-2">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              send(input);
            }}
            className="max-w-2xl mx-auto relative"
          >
            <input
              ref={inputRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Message Caret..."
              aria-label="Message Caret"
              disabled={isStreaming}
              className="w-full h-12 rounded-xl pl-4 pr-12 text-sm outline-none transition-colors"
              style={{
                background: 'var(--dash-card)',
                border: '1px solid var(--dash-border)',
                color: 'var(--dash-text)',
                fontSize: '16px',
              }}
              onFocus={(e) => {
                (e.target as HTMLElement).style.borderColor = 'rgba(69,131,196,0.4)';
              }}
              onBlur={(e) => {
                (e.target as HTMLElement).style.borderColor = 'var(--dash-border)';
              }}
            />
            <button
              type="submit"
              disabled={!input.trim() || isStreaming}
              aria-label="Send message"
              className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8 rounded-lg flex items-center justify-center transition-colors"
              style={{
                background: input.trim() && !isStreaming ? '#4583C4' : 'transparent',
                color: input.trim() && !isStreaming ? '#fff' : 'var(--dash-text-dim)',
              }}
            >
              <Send size={15} />
            </button>
          </form>
        </div>
      </div>
    </>
  );
};

export default Chat;
