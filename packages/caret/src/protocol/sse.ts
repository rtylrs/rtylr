import { parseStreamLine, type CaretStreamEvent } from './events';

/** Consume an SSE response body. Yields a parsed event per `data:` line. */
export async function* consumeSse(
  res: Response,
  fallbackMessageId: string
): AsyncGenerator<CaretStreamEvent> {
  if (!res.body) return;
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() ?? '';
      for (const rawLine of lines) {
        const line = rawLine.startsWith('data: ') ? rawLine.slice(6) : rawLine;
        const ev = parseStreamLine(line, fallbackMessageId);
        if (ev) yield ev;
      }
    }
    // flush remainder
    if (buffer.trim()) {
      const line = buffer.startsWith('data: ') ? buffer.slice(6) : buffer;
      const ev = parseStreamLine(line, fallbackMessageId);
      if (ev) yield ev;
    }
  } finally {
    try {
      reader.releaseLock();
    } catch {
      // ignore - stream may already be closed
    }
  }
}
