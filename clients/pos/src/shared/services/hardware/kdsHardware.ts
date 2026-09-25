import { invoke } from '@tauri-apps/api/core';
import type { KdsHardwareSettings } from '../../stores/settingsStore';
import { canUseNativeHardware } from '../../utils/platform';

export interface KdsTicketItem {
  name: string;
  quantity: number;
  notes?: string;
  modifiers: string[];
}

export interface KdsTicket {
  orderNumber: string;
  station?: string;
  orderType?: string;
  tableNumber?: string;
  items: KdsTicketItem[];
  createdAt?: string;
}

export type KdsHardwareResult =
  | { status: 'sent'; via: string }
  | { status: 'skipped' }
  | { status: 'failed'; error: string };

/**
 * Send a kitchen ticket to a hardware KDS display.
 * Routes through serial or network based on the configured transport.
 * If transport is 'none', returns 'skipped' (software-only KDS).
 */
export async function sendKdsTicket(
  config: KdsHardwareSettings,
  ticket: KdsTicket,
): Promise<KdsHardwareResult> {
  if (!canUseNativeHardware()) return { status: 'skipped' };

  const payload = {
    order_number: ticket.orderNumber,
    station: ticket.station ?? null,
    order_type: ticket.orderType ?? null,
    table_number: ticket.tableNumber ?? null,
    items: ticket.items.map((it) => ({
      name: it.name,
      quantity: it.quantity,
      notes: it.notes ?? null,
      modifiers: it.modifiers,
    })),
    created_at: ticket.createdAt ?? null,
  };

  switch (config.transport) {
    case 'none':
      return { status: 'skipped' };

    case 'serial':
      if (!config.serialPort.trim()) return { status: 'skipped' };
      try {
        const msg = await invoke<string>('send_kds_ticket_serial', {
          portName: config.serialPort.trim(),
          baudRate: config.serialBaudRate,
          ticket: payload,
        });
        return { status: 'sent', via: msg };
      } catch (err) {
        return { status: 'failed', error: err instanceof Error ? err.message : String(err) };
      }

    case 'network':
      if (!config.networkHost.trim()) return { status: 'skipped' };
      try {
        const msg = await invoke<string>('send_kds_ticket_network', {
          host: config.networkHost.trim(),
          port: config.networkPort,
          ticket: payload,
        });
        return { status: 'sent', via: msg };
      } catch (err) {
        return { status: 'failed', error: err instanceof Error ? err.message : String(err) };
      }
    default:
      return { status: 'skipped' };
  }
}
