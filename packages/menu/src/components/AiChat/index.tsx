import React, { useState, useRef, useEffect } from 'react';
import classNames from 'classnames';
import { motion, AnimatePresence } from 'motion/react';
import styles from './AiChat.module.scss';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
}

interface AiChatProps {
  /** Called when the user sends a message. Resolve with the assistant reply. */
  onSendMessage: (message: string, history: ChatMessage[]) => Promise<string>;
  /** Placeholder text in the input */
  placeholder?: string;
  /** Label shown above the chat bubble button */
  buttonLabel?: string;
  /** Greeting message shown on first open */
  greeting?: string;
  /** Business name to personalise the greeting */
  businessName?: string;
  className?: string;
}

const AiChat: React.FC<AiChatProps> = ({
  onSendMessage,
  placeholder = 'Ask about our menu…',
  buttonLabel,
  greeting,
  businessName,
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const defaultGreeting =
    greeting ??
    (businessName
      ? `Hi! I'm the AI assistant for ${businessName}. Ask me anything about the menu, ingredients, or specials!`
      : "Hi! I'm your AI menu assistant. Ask me anything about the menu!");

  // Show greeting when first opened
  useEffect(() => {
    if (isOpen && messages.length === 0) {
      setMessages([
        {
          id: 'greeting',
          role: 'assistant',
          content: defaultGreeting,
          timestamp: new Date(),
        },
      ]);
    }
    if (isOpen) {
      const timeoutId = window.setTimeout(() => inputRef.current?.focus(), 150);
      return () => window.clearTimeout(timeoutId);
    }
  }, [defaultGreeting, isOpen, messages.length]);

  // Scroll to bottom on new message
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, loading]);

  const sendMessage = async () => {
    const text = input.trim();
    if (!text || loading) return;

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: text,
      timestamp: new Date(),
    };

    const nextHistory = [...messages, userMsg];
    setMessages(nextHistory);
    setInput('');
    setLoading(true);

    try {
      const reply = await onSendMessage(text, nextHistory);
      setMessages((prev) => [
        ...prev,
        {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: reply,
          timestamp: new Date(),
        },
      ]);
    } catch {
      setMessages((prev) => [
        ...prev,
        {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: 'Sorry, I had trouble responding. Please try again.',
          timestamp: new Date(),
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <div className={classNames(styles.aiChat, className)}>
      {/* Floating toggle button */}
      <button
        className={styles.aiChat__fab}
        onClick={() => setIsOpen((v) => !v)}
        aria-label={isOpen ? 'Close AI chat' : 'Open AI chat'}
      >
        <AnimatePresence mode="wait">
          {isOpen ? (
            <motion.span
              key="close"
              initial={{ rotate: -90, opacity: 0 }}
              animate={{ rotate: 0, opacity: 1 }}
              exit={{ rotate: 90, opacity: 0 }}
              transition={{ duration: 0.15 }}
            >
              ✕
            </motion.span>
          ) : (
            <motion.span
              key="open"
              initial={{ rotate: 90, opacity: 0 }}
              animate={{ rotate: 0, opacity: 1 }}
              exit={{ rotate: -90, opacity: 0 }}
              transition={{ duration: 0.15 }}
              className={styles.aiChat__fabIcon}
            >
              ✦
            </motion.span>
          )}
        </AnimatePresence>
        {!isOpen && buttonLabel && <span className={styles.aiChat__fabLabel}>{buttonLabel}</span>}
      </button>

      {/* Chat panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            role="dialog"
            aria-modal="true"
            aria-label="AI Assistant"
            className={styles.aiChat__panel}
            initial={{ opacity: 0, y: 20, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 20, scale: 0.95 }}
            transition={{ type: 'spring', damping: 26, stiffness: 340 }}
          >
            <div className={styles.aiChat__panelHeader}>
              <span className={styles.aiChat__panelTitle}>✦ AI Assistant</span>
              <span className={styles.aiChat__panelBadge}>Online</span>
            </div>

            <div className={styles.aiChat__messages}>
              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={classNames(styles.aiChat__bubble, {
                    [styles['aiChat__bubble--user']]: msg.role === 'user',
                    [styles['aiChat__bubble--assistant']]: msg.role === 'assistant',
                  })}
                >
                  {msg.content}
                </div>
              ))}

              {loading && (
                <div
                  className={classNames(
                    styles.aiChat__bubble,
                    styles['aiChat__bubble--assistant'],
                    styles['aiChat__bubble--typing']
                  )}
                >
                  <span />
                  <span />
                  <span />
                </div>
              )}

              <div ref={bottomRef} />
            </div>

            <div className={styles.aiChat__inputRow}>
              <input
                ref={inputRef}
                className={styles.aiChat__input}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={placeholder}
                disabled={loading}
              />
              <button
                className={styles.aiChat__sendBtn}
                onClick={sendMessage}
                disabled={!input.trim() || loading}
                aria-label="Send message"
              >
                ↑
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default AiChat;
