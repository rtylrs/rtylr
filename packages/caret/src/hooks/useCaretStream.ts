import { useCaretStore } from '../stores/caretStore';

interface StreamEvent {
  text: string;
  done: boolean;
  session_id?: string;
}

interface UseCaretStreamParams {
  userId: number;
  apiBaseUrl: string;
  systemContext?: string;
}

interface UseCaretStreamResult {
  sendMessage: (prompt: string) => Promise<void>;
}

export function useCaretStream({
  userId,
  apiBaseUrl,
  systemContext,
}: UseCaretStreamParams): UseCaretStreamResult {
  const store = useCaretStore();

  const sendMessage = async (prompt: string): Promise<void> => {
    if (!prompt.trim() || store.isStreaming) return;

    const userMessageId = `user-${Date.now()}`;
    store.addMessage({
      id: userMessageId,
      role: 'user',
      content: prompt.trim(),
      createdAt: new Date().toISOString(),
    });

    store.startAssistantMessage();

    const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
    const url = `${baseUrl}ai/generate-text-stream`;

    try {
      const response = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          prompt: prompt.trim(),
          session_id: store.sessionId ?? '',
          user_id: userId,
          system_context: systemContext ?? '',
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      if (!response.body) {
        throw new Error('Response body is null');
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed.startsWith('data:')) continue;

          const jsonStr = trimmed.slice(5).trim();
          if (!jsonStr) continue;

          let event: StreamEvent;
          try {
            event = JSON.parse(jsonStr) as StreamEvent;
          } catch {
            continue;
          }

          if (event.session_id && !store.sessionId) {
            store.setSessionId(event.session_id);
          }

          if (event.done) {
            store.setStreaming(false);
            return;
          }

          if (event.text) {
            store.appendToLastAssistantMessage(event.text);
          }
        }
      }

      store.setStreaming(false);
    } catch (error) {
      store.setStreaming(false);
      const message =
        error instanceof Error ? error.message : 'An unexpected error occurred. Please try again.';
      store.addMessage({
        id: `error-${Date.now()}`,
        role: 'assistant',
        content: `Error: ${message}`,
        createdAt: new Date().toISOString(),
      });
    }
  };

  return { sendMessage };
}
