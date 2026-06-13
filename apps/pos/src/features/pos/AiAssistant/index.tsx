import React, { useState, useRef, useEffect } from 'react';
import { Sparkles, X } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';
import { useCartStore } from '../../../shared/stores/cartStore';

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
      content:
        "Hi! I'm your POS AI assistant. I can help with order suggestions, upsells, or questions about menu items.",
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
    if (isOpen) setTimeout(() => inputRef.current?.focus(), 100);
  }, [isOpen]);

  const send = async () => {
    const text = input.trim();
    if (!text || loading) return;
    const userMsg: Message = { id: Date.now().toString(), role: 'user', content: text };
    setMessages((p) => [...p, userMsg]);
    setInput('');
    setLoading(true);
    try {
      const reply = onSendMessage
        ? await onSendMessage(text, messages)
        : `I understand you're asking about "${text}". Once the backend is connected, I'll provide real-time insights.`;
      setMessages((p) => [
        ...p,
        { id: (Date.now() + 1).toString(), role: 'assistant', content: reply },
      ]);
    } catch {
      setMessages((p) => [
        ...p,
        {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content: 'Sorry, something went wrong.',
        },
      ]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed bottom-5 right-5 z-50 flex flex-col items-end gap-3">
      {/* FAB */}
      <button
        className="w-12 h-12 rounded-full bg-pos-text text-pos-bg flex items-center justify-center shadow-modal hover:opacity-90 transition-opacity"
        onClick={() => setIsOpen((v) => !v)}
        aria-label={isOpen ? 'Close AI' : 'Open AI'}
      >
        <AnimatePresence mode="wait">
          {isOpen ? (
            <motion.span
              key="x"
              initial={{ rotate: -90, opacity: 0 }}
              animate={{ rotate: 0, opacity: 1 }}
              exit={{ rotate: 90, opacity: 0 }}
              transition={{ duration: 0.12 }}
            >
              <X className="h-5 w-5" aria-hidden="true" />
            </motion.span>
          ) : (
            <motion.span
              key="ai"
              initial={{ rotate: 90, opacity: 0 }}
              animate={{ rotate: 0, opacity: 1 }}
              exit={{ rotate: -90, opacity: 0 }}
              transition={{ duration: 0.12 }}
            >
              <Sparkles className="h-5 w-5" aria-hidden="true" />
            </motion.span>
          )}
        </AnimatePresence>
      </button>

      {/* Panel */}
      <AnimatePresence>
        {isOpen && (
          <motion.div
            className="w-80 h-96 flex flex-col bg-pos-card border border-pos-border rounded-2xl shadow-modal overflow-hidden"
            initial={{ opacity: 0, y: 12, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 12, scale: 0.96 }}
            transition={{ type: 'spring', damping: 28, stiffness: 320 }}
          >
            <div className="flex items-center justify-between px-4 py-3 border-b border-pos-border shrink-0">
              <span className="inline-flex items-center gap-2 text-sm font-semibold text-pos-text">
                <Sparkles className="h-4 w-4" aria-hidden="true" />
                POS Assistant
              </span>
              {cartItems.length > 0 && (
                <span className="text-xs text-pos-muted">
                  {cartItems.length} item{cartItems.length > 1 ? 's' : ''} in cart
                </span>
              )}
            </div>

            <div className="flex-1 overflow-y-auto p-3 space-y-2">
              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`max-w-[85%] px-3 py-2 rounded-xl text-sm leading-relaxed ${
                    msg.role === 'user'
                      ? 'ml-auto bg-pos-text text-pos-bg rounded-br-sm'
                      : 'bg-white/8 text-pos-text rounded-bl-sm'
                  }`}
                >
                  {msg.content}
                </div>
              ))}
              {loading && (
                <div className="bg-white/8 rounded-xl rounded-bl-sm px-3 py-2 w-fit flex gap-1">
                  {[0, 150, 300].map((d) => (
                    <span
                      key={d}
                      className="w-1.5 h-1.5 rounded-full bg-pos-muted animate-bounce"
                      style={{ animationDelay: `${d}ms` }}
                    />
                  ))}
                </div>
              )}
              <div ref={bottomRef} />
            </div>

            <div className="flex gap-2 p-3 border-t border-pos-border shrink-0">
              <input
                ref={inputRef}
                className="flex-1 bg-white/6 border border-pos-border rounded-lg px-3 py-2 text-sm text-pos-text placeholder:text-pos-muted outline-none focus:border-white/20 transition-colors"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && (e.preventDefault(), send())}
                placeholder="Ask about upsells, items…"
                disabled={loading}
              />
              <button
                className="w-9 h-9 flex items-center justify-center rounded-lg bg-pos-text text-pos-bg font-bold hover:opacity-90 disabled:opacity-40 transition-opacity"
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
