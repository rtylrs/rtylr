/** Per-message status indicator for the chat UI. */
export type MessageStatus = 'sending' | 'sent' | 'delivered' | 'failed' | 'cancelled';

export function isTerminalMessageStatus(s: MessageStatus): boolean {
  return s === 'delivered' || s === 'failed' || s === 'cancelled';
}
