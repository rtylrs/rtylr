import type { AlarmState } from '../../../../shared/alarm/alarm';
import { isOpenStatus } from './ticketPoller';

export interface AlarmTicket {
  status?: string | undefined;
  readyAt?: string | undefined;
  remindedAt?: string | undefined;
  acknowledgedAt?: string | undefined;
}

/**
 * The key changes when the order becomes ready and again on every "call
 * again", so each of those rings once. An acknowledged ticket is silent until
 * the staff remind it, which clears the acknowledgement and moves the key.
 */
export function ticketRingKey(ticket: AlarmTicket | null): string | null {
  if (!ticket || ticket.status !== 'ready' || ticket.acknowledgedAt) return null;
  return `${ticket.readyAt ?? ''}|${ticket.remindedAt ?? ''}`;
}

export function ticketAlarmState(ticket: AlarmTicket | null, title?: string): AlarmState {
  return {
    keepAwake: isOpenStatus(ticket?.status),
    ringKey: ticketRingKey(ticket),
    title,
  };
}
