/**
 * In-memory ring buffer that records the last N message votes (thumbs
 * up/down) so the analytics layer can flush them in the background
 * without round-tripping each click.
 */
export type Vote = 'up' | 'down';

export interface FeedbackEntry {
  messageId: string;
  vote: Vote;
  at: number;
}

const MAX = 50;
const buffer: FeedbackEntry[] = [];

export function recordVote(messageId: string, vote: Vote): void {
  buffer.push({ messageId, vote, at: Date.now() });
  if (buffer.length > MAX) buffer.shift();
}

export function drainVotes(): FeedbackEntry[] {
  const out = buffer.slice();
  buffer.length = 0;
  return out;
}
