/**
 * Buffer streaming tokens and flush in animation-frame batches so the
 * renderer doesn't thrash when the model produces fast tokens.
 */
export interface StreamBuffer {
  push(token: string): void;
  stop(): void;
}

export function createStreamBuffer(onFlush: (chunk: string) => void): StreamBuffer {
  let buf = '';
  let raf: number | null = null;

  const flush = (): void => {
    if (!buf) {
      raf = null;
      return;
    }
    const chunk = buf;
    buf = '';
    raf = null;
    onFlush(chunk);
  };

  return {
    push(token: string): void {
      buf += token;
      if (raf == null && typeof requestAnimationFrame !== 'undefined') {
        raf = requestAnimationFrame(flush);
      } else if (raf == null) {
        // SSR path - flush synchronously
        flush();
      }
    },
    stop(): void {
      if (raf != null && typeof cancelAnimationFrame !== 'undefined') {
        cancelAnimationFrame(raf);
        raf = null;
      }
      if (buf) flush();
    },
  };
}
