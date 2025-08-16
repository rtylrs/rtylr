import React, { useState, useEffect, useRef, useCallback } from 'react';
import './Chat.scss';
import api from '../../lib/api';

interface Message {
  id: string;
  content: string;
  isUser: boolean;
  isSystem?: boolean;
  hasMemory?: boolean;
}

interface Session {
  session_id: string;
}

const Chat: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isStreaming, setIsStreaming] = useState(false);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const chatMessagesRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const currentUserId = 1; // Default user ID

  // Initialize session on component mount
  useEffect(() => {
    initializeSession();
  }, []);

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    if (chatMessagesRef.current) {
      chatMessagesRef.current.scrollTop = chatMessagesRef.current.scrollHeight;
    }
  }, [messages]);

  // Focus input on mount
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  const initializeSession = async () => {
    try {
      console.log('Attempting to create session...');
      const response = await api.post('/api/memory/sessions', {
        user_id: currentUserId,
        title: `Chat Session - ${new Date().toLocaleString()}`,
      });

      const data = response.data;
      console.log('Response data:', data);

      if (data.success) {
        setCurrentSession(data.session);
        addSystemMessage(`Session created: ${data.session.session_id}`);
      } else {
        const errorMsg =
          typeof data.error === 'object'
            ? JSON.stringify(data.error)
            : data.error || 'Unknown error';
        addSystemMessage(
          `Failed to create session: ${errorMsg}. Memory features disabled.`
        );
      }
    } catch (error) {
      console.error('Failed to initialize session:', error);
      addSystemMessage(
        `Failed to initialize memory: ${
          error instanceof Error ? error.message : 'Unknown error'
        }. Continuing without memory features.`
      );
    }
  };

  const addMessage = useCallback(
    (content: string, isUser: boolean, hasMemory = false) => {
      const newMessage: Message = {
        id: Date.now().toString(),
        content,
        isUser,
        hasMemory,
      };
      setMessages((prev) => [...prev, newMessage]);
    },
    []
  );

  const addSystemMessage = useCallback((content: string) => {
    const newMessage: Message = {
      id: Date.now().toString(),
      content,
      isUser: false,
      isSystem: true,
      hasMemory: false,
    };
    setMessages((prev) => [...prev, newMessage]);
  }, []);

  const getContext = async (): Promise<string> => {
    if (!currentSession) return '';

    try {
      const response = await api.post('/api/memory/context', {
        session_id: currentSession.session_id,
        user_id: currentUserId,
      });

      const data = response.data;
      console.log('Context response:', data);

      if (data.success && data.context) {
        let contextString = '\n\n[Previous conversation context]:\n';

        if (data.context.messages && data.context.messages.length > 0) {
          data.context.messages.forEach((message: any) => {
            contextString += `${message.role}: ${message.content}\n`;
          });
        }

        if (data.context.memories && data.context.memories.length > 0) {
          data.context.memories.forEach((memory: any) => {
            contextString += `[Memory] ${memory.memory_type}: ${memory.value}\n`;
          });
        }

        return contextString;
      }
    } catch (error) {
      console.error('Failed to get context:', error);
    }
    return '';
  };

  const parseMarkdown = (text: string): string => {
    return text
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/`([^`]+)`/g, '<code>$1</code>')
      .replace(/\n/g, '<br>')
      .replace(/^\* (.+)$/gm, '<li>$1</li>')
      .replace(/(<li>.*<\/li>)/, '<ul>$1</ul>');
  };

  const sendMessage = async () => {
    const prompt = inputValue.trim();
    if (!prompt || isStreaming) return;

    // Add user message to chat
    addMessage(prompt, true);
    setInputValue('');
    setIsStreaming(true);
    setIsLoading(true);

    try {
      // Get conversation context
      const context = await getContext();
      const enhancedPrompt = prompt + context;

      console.log('Sending AI request:', {
        prompt: enhancedPrompt,
        session_id: currentSession ? currentSession.session_id : '',
        user_id: currentUserId,
      });

      // Create AI response message
      const messageId = Date.now().toString();
      const newMessage: Message = {
        id: messageId,
        content: '',
        isUser: false,
      };
      setMessages((prev) => [...prev, newMessage]);

      // Stream response from AI using Server-Sent Events
      const response = await fetch(
        `${api.defaults.baseURL}api/ai/generate-text-stream`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            prompt: enhancedPrompt,
            session_id: currentSession ? currentSession.session_id : '',
            user_id: currentUserId,
          }),
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Handle SSE response
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let fullResponse = '';

      if (reader) {
        const readStream = async () => {
          try {
            while (true) {
              const { done, value } = await reader.read();
              if (done) break;

              buffer += decoder.decode(value, { stream: true });
              const lines = buffer.split('\n');
              buffer = lines.pop() || '';

              for (const line of lines) {
                if (line.trim() === '') continue;

                if (line.startsWith('data: ')) {
                  const data = line.substring(6);
                  if (data.trim() === '') continue;

                  try {
                    const token = JSON.parse(data);

                    if (token.error) {
                      setMessages((prev) =>
                        prev.map((msg) =>
                          msg.id === messageId
                            ? {
                                ...msg,
                                content:
                                  msg.content + `\n[ERROR] ${token.error}`,
                              }
                            : msg
                        )
                      );
                      continue;
                    }

                    if (token.done) continue;

                    if (token.text) {
                      fullResponse += token.text;
                      setMessages((prev) =>
                        prev.map((msg) =>
                          msg.id === messageId
                            ? { ...msg, content: fullResponse }
                            : msg
                        )
                      );
                    }
                  } catch (error) {
                    console.warn('Failed to parse SSE data:', data, error);
                  }
                }
              }
            }
          } catch (error) {
            console.error('Error reading stream:', error);
          }
        };

        await readStream();
      }
    } catch (error) {
      if (
        error instanceof Error &&
        error.name === 'TypeError' &&
        error.message.includes('Failed to fetch')
      ) {
        addMessage(
          '❌ Connection failed. Please check if the server is running.',
          false
        );
      } else {
        addMessage(
          `❌ Error: ${
            error instanceof Error ? error.message : 'Unknown error'
          }`,
          false
        );
      }
    } finally {
      setIsStreaming(false);
      setIsLoading(false);
      if (inputRef.current) {
        inputRef.current.focus();
      }
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const renderMessage = (message: Message) => {
    let messageClass = 'assistant';
    if (message.isUser) {
      messageClass = 'user';
    } else if (message.isSystem) {
      messageClass = 'system';
    }

    const content = message.isUser
      ? message.content
      : parseMarkdown(message.content);

    return (
      <div key={message.id} className={`message ${messageClass}`}>
        <div
          className="message-content"
          dangerouslySetInnerHTML={{ __html: content }}
        />
        {message.hasMemory && <span className="memory-badge">Memory</span>}
      </div>
    );
  };

  return (
    <div className="chat-page">
      <div className="header">
        <div className="header-content">
          <div className="header-logo">
            <img src="/logo.png" alt="Vbitez Logo" />
          </div>
          <div className="header-right">
            <div className="memory-status">
              <div className="memory-indicator"></div>
              Memory Active
            </div>
            <div className="session-info">
              Session:{' '}
              {currentSession
                ? currentSession.session_id.slice(-8)
                : 'Not Started'}
            </div>
          </div>
        </div>
      </div>

      <div className="main-container">
        <div className="chat-container">
          <div className="chat-messages" ref={chatMessagesRef}>
            {messages.map(renderMessage)}
            {isLoading && (
              <div className="message assistant">
                <div className="typing-indicator">
                  <div className="typing-dot"></div>
                  <div className="typing-dot"></div>
                  <div className="typing-dot"></div>
                </div>
              </div>
            )}
          </div>

          <div className="input-container">
            <div className="input-wrapper">
              <textarea
                ref={inputRef}
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Ask anything..."
                rows={1}
                disabled={isStreaming}
              />
              <button
                className="send-button"
                onClick={sendMessage}
                disabled={isStreaming || !inputValue.trim()}
              >
                {isStreaming ? (
                  <div className="loading-spinner"></div>
                ) : (
                  <svg
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                  >
                    <path d="M12 5l0 14M5 12l7 -7 7 7" />
                  </svg>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Chat;
