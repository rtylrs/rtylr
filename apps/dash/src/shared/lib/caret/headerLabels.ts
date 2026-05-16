/** Friendly labels for the standard HTTP headers we set on Caret requests. */
export const CARET_HEADERS = {
  accept: 'application/json, text/plain, */*',
  contentType: 'application/json',
  /** Sent so the backend can log which UI surface initiated a chat. */
  surface: 'X-Caret-Surface',
  /** Sent so the backend can stitch per-message correlation across spans. */
  trace: 'X-Caret-Trace-Id',
} as const;
