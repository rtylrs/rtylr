/** UI streaming-status shorthand used in lots of components. */
export type StreamingState = 'idle' | 'thinking' | 'streaming' | 'tool' | 'awaiting-approval';

export function streamingLabel(s: StreamingState): string {
  switch (s) {
    case 'idle':
      return '';
    case 'thinking':
      return 'Thinking…';
    case 'streaming':
      return 'Generating…';
    case 'tool':
      return 'Running tool…';
    case 'awaiting-approval':
      return 'Waiting for approval';
  }
}
