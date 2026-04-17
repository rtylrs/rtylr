import React, { useState, useRef, useEffect } from 'react';
import { Send, Sparkles } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

const SUGGESTIONS = [
  'How much revenue did we make today?',
  'Which products are low on stock?',
  'Show me our top 5 customers',
  'How many staff are clocked in?',
];

const Chat: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isTyping]);

  const send = (text: string) => {
    if (!text.trim()) return;
    const userMsg: Message = { id: Date.now().toString(), role: 'user', content: text.trim() };
    setMessages((p) => [...p, userMsg]);
    setInput('');
    setIsTyping(true);

    setTimeout(() => {
      setMessages((p) => [
        ...p,
        {
          id: (Date.now() + 1).toString(),
          role: 'assistant',
          content:
            'Caret AI is coming soon. When ready, I will be able to answer questions about your revenue, inventory, staff, customers, and more by querying your live business data across all rtylr apps.',
        },
      ]);
      setIsTyping(false);
    }, 1500);
  };

  const isEmpty = messages.length === 0;

  return (
    <div className="flex flex-col overflow-hidden -m-6" style={{ height: 'calc(100vh - 44px)' }}>
      {/* Messages or empty state */}
      <div className="flex-1 overflow-y-auto">
        {isEmpty ? (
          <div className="flex flex-col items-center justify-center h-full px-6">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
              className="text-center max-w-md"
            >
              <div
                className="h-12 w-12 rounded-2xl flex items-center justify-center mx-auto mb-5"
                style={{ background: 'rgba(139,92,246,0.1)' }}
              >
                <Sparkles size={24} style={{ color: '#8b5cf6' }} />
              </div>
              <h1 className="text-xl font-semibold mb-2" style={{ color: 'var(--hub-text)' }}>
                Chat with Caret
              </h1>
              <p
                className="text-sm leading-relaxed mb-8"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                Your AI assistant with access to all your business data.
              </p>

              <div className="grid grid-cols-2 gap-2">
                {SUGGESTIONS.map((s, i) => (
                  <motion.button
                    key={i}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: 0.1 + i * 0.05 }}
                    onClick={() => send(s)}
                    className="text-left px-4 py-3 rounded-xl text-sm transition-all"
                    style={{
                      background: 'var(--hub-card)',
                      border: '1px solid var(--hub-border)',
                      color: 'var(--hub-text)',
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.borderColor = 'rgba(139,92,246,0.4)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
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
              {messages.map((msg) => (
                <motion.div
                  key={msg.id}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  {msg.role === 'assistant' && (
                    <div
                      className="h-7 w-7 rounded-lg flex items-center justify-center shrink-0 mr-3 mt-1"
                      style={{ background: 'rgba(139,92,246,0.1)' }}
                    >
                      <Sparkles size={14} style={{ color: '#8b5cf6' }} />
                    </div>
                  )}
                  <div
                    className="rounded-2xl px-4 py-3 max-w-[80%]"
                    style={{
                      background: msg.role === 'user' ? '#4583C4' : 'var(--hub-card)',
                      color: msg.role === 'user' ? '#fff' : 'var(--hub-text)',
                      border: msg.role === 'user' ? 'none' : '1px solid var(--hub-border)',
                    }}
                  >
                    <p className="text-sm leading-relaxed">{msg.content}</p>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>

            {isTyping && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="flex items-center gap-3"
              >
                <div
                  className="h-7 w-7 rounded-lg flex items-center justify-center shrink-0"
                  style={{ background: 'rgba(139,92,246,0.1)' }}
                >
                  <Sparkles size={14} style={{ color: '#8b5cf6' }} />
                </div>
                <div
                  className="flex items-center gap-1.5 px-4 py-3 rounded-2xl"
                  style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                >
                  {[0, 1, 2].map((i) => (
                    <motion.div
                      key={i}
                      className="h-1.5 w-1.5 rounded-full"
                      style={{ background: '#8b5cf6' }}
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
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Message Caret..."
            className="w-full h-12 rounded-xl pl-4 pr-12 text-sm outline-none transition-colors"
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
            onFocus={(e) => {
              (e.target as HTMLElement).style.borderColor = 'rgba(139,92,246,0.4)';
            }}
            onBlur={(e) => {
              (e.target as HTMLElement).style.borderColor = 'var(--hub-border)';
            }}
          />
          <button
            type="submit"
            disabled={!input.trim()}
            className="absolute right-2 top-1/2 -translate-y-1/2 h-8 w-8 rounded-lg flex items-center justify-center transition-colors"
            style={{
              background: input.trim() ? '#8b5cf6' : 'transparent',
              color: input.trim() ? '#fff' : 'var(--hub-text-dim)',
            }}
          >
            <Send size={15} />
          </button>
        </form>
      </div>
    </div>
  );
};

export default Chat;
