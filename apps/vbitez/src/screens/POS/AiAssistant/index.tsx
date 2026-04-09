import React, { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useCartStore } from '../../../stores/cartStore';
import styles from './AiAssistant.module.scss';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

interface AiAssistantProps {
  onSendMessage?: (message: string, history: Message[]) => Promise<string>;
}

const AiAssistant: React.FC<AiAssistantProps> = ({ onSendMessage }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 'greeting',
      role: 'assistant',
      content: 'Hi! I\'m your POS AI assistant. I can help you with order suggestions, upsells, or answer questions about menu items.',
    },
  ]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const cartItems = useCartStore((s) => s.items);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, loading]);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  }, [isOpen]);

  const send = async () => {
    const text = input.trim();
    if (!text || loading) return;

    const userMsg: Message = { id: Date.now().toString(), role: 'user', content: text };
    setMessages((prev) => [...prev, userMsg]);
    setInput('');
    setLoading(true);

    try {
      let reply: string;
      if (onSendMessage) {
        reply = await onSendMessage(text, messages);
      } else {
        // Placeholder response until backend is wired
        reply = `I understand you're asking about "${text}". Once the backend is connected, I'll be able to provide real-time menu insights, suggest upsells, and help manage orders.`;
      }
      setMessages((prev) => [
        ...prev,
        { id: (Date.now() + 1).toString(), role: 'assistant', content: reply },
      ]);
    } catch {
      setMessages((prev) => [
        ...prev,
        { id: (Date.now() + 1).toString(), role: 'assistant', content: 'Sorry, something went wrong. Please try again.' },
      ]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.ai}>
      {/* FAB */}
      <button
        className={styles.ai__fab}
        onClick={() => setIsOpen((v) => !v)}
        aria-label={isOpen ? 'Close AI assistant' : 'Open AI assistant'}
        title="AI Assistant"
      >
        <AnimatePresence mode="wait">
          {isOpen ? (
            <motion.span key="x" initial={{ rotate: -90, opacity: 0 }} animate={{ rotate: 0, opacity: 1 }} exit={{ rotate: 90, opacity: 0 }} transition={{ duration: 0.12 }}>✕</motion.span>
          ) : (
            <motion.span key="ai" initial={{ rotate: 90, opacity: 0 }} animate={{ rotate: 0, opacity: 1 }} exit={{ rotate: -90, opacity: 0 }} transition={{ duration: 0.12 }}>✦</motion.span>
          )}
        </AnimatePresence>
        {!isOpen && <span className={styles.ai__fabLabel}>AI</span>}
      </button>

      {/* Panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            className={styles.ai__panel}
            initial={{ opacity: 0, y: 12, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 12, scale: 0.96 }}
            transition={{ type: 'spring', damping: 28, stiffness: 320 }}
          >
            <div className={styles.ai__header}>
              <span className={styles.ai__headerTitle}>✦ POS Assistant</span>
              {cartItems.length > 0 && (
                <span className={styles.ai__headerCtx}>{cartItems.length} item{cartItems.length > 1 ? 's' : ''} in cart</span>
              )}
            </div>

            <div className={styles.ai__messages}>
              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`${styles.ai__bubble} ${msg.role === 'user' ? styles['ai__bubble--user'] : styles['ai__bubble--assistant']}`}
                >
                  {msg.content}
                </div>
              ))}
              {loading && (
                <div className={`${styles.ai__bubble} ${styles['ai__bubble--assistant']} ${styles['ai__bubble--typing']}`}>
                  <span /><span /><span />
                </div>
              )}
              <div ref={bottomRef} />
            </div>

            <div className={styles.ai__inputRow}>
              <input
                ref={inputRef}
                className={styles.ai__input}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), send())}
                placeholder="Ask about upsells, items…"
                disabled={loading}
              />
              <button
                className={styles.ai__sendBtn}
                onClick={send}
                disabled={!input.trim() || loading}
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

export default AiAssistant;
