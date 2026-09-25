import { describe, expect, it, vi, beforeEach } from 'vitest';
import { sendKdsTicket } from './kdsHardware';
import type { KdsHardwareSettings } from '../../stores/settingsStore';
import type { KdsTicket } from './kdsHardware';

vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(),
}));

import { invoke } from '@tauri-apps/api/core';

const ticket: KdsTicket = {
  orderNumber: 'ORD-42',
  station: 'grill',
  items: [
    { name: 'Burger', quantity: 2, modifiers: ['no onion'], notes: 'well done' },
    { name: 'Fries', quantity: 1, modifiers: [] },
  ],
  createdAt: '2026-05-13T10:00:00.000Z',
};

function kdsConfig(overrides: Partial<KdsHardwareSettings> = {}): KdsHardwareSettings {
  return {
    transport: 'none',
    serialPort: '',
    serialBaudRate: 9600,
    networkHost: '',
    networkPort: 9100,
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('sendKdsTicket - none transport', () => {
  it('returns skipped immediately without invoking Tauri', async () => {
    const result = await sendKdsTicket(kdsConfig({ transport: 'none' }), ticket);
    expect(result).toEqual({ status: 'skipped' });
    expect(invoke).not.toHaveBeenCalled();
  });
});

describe('sendKdsTicket - serial transport', () => {
  it('invokes send_kds_ticket_serial with correct payload and returns sent', async () => {
    vi.mocked(invoke).mockResolvedValue('Ticket sent via serial COM3');

    const result = await sendKdsTicket(
      kdsConfig({ transport: 'serial', serialPort: 'COM3', serialBaudRate: 9600 }),
      ticket,
    );

    expect(result).toEqual({ status: 'sent', via: 'Ticket sent via serial COM3' });
    expect(invoke).toHaveBeenCalledWith('send_kds_ticket_serial', {
      portName: 'COM3',
      baudRate: 9600,
      ticket: expect.objectContaining({
        order_number: 'ORD-42',
        station: 'grill',
        items: expect.arrayContaining([
          expect.objectContaining({ name: 'Burger', quantity: 2 }),
        ]),
      }),
    });
  });

  it('returns failed when Tauri throws', async () => {
    vi.mocked(invoke).mockRejectedValue(new Error('port not found'));

    const result = await sendKdsTicket(
      kdsConfig({ transport: 'serial', serialPort: 'COM99' }),
      ticket,
    );

    expect(result).toEqual({ status: 'failed', error: 'port not found' });
  });

  it('returns skipped when serial port is blank', async () => {
    const result = await sendKdsTicket(
      kdsConfig({ transport: 'serial', serialPort: '   ' }),
      ticket,
    );
    expect(result).toEqual({ status: 'skipped' });
    expect(invoke).not.toHaveBeenCalled();
  });
});

describe('sendKdsTicket - network transport', () => {
  it('invokes send_kds_ticket_network with correct host and port', async () => {
    vi.mocked(invoke).mockResolvedValue('Ticket sent via TCP 192.168.1.50:9100');

    const result = await sendKdsTicket(
      kdsConfig({ transport: 'network', networkHost: '192.168.1.50', networkPort: 9100 }),
      ticket,
    );

    expect(result).toEqual({ status: 'sent', via: 'Ticket sent via TCP 192.168.1.50:9100' });
    expect(invoke).toHaveBeenCalledWith('send_kds_ticket_network', {
      host: '192.168.1.50',
      port: 9100,
      ticket: expect.objectContaining({ order_number: 'ORD-42' }),
    });
  });

  it('returns failed when TCP connection fails', async () => {
    vi.mocked(invoke).mockRejectedValue(new Error('connection refused'));

    const result = await sendKdsTicket(
      kdsConfig({ transport: 'network', networkHost: '10.0.0.1' }),
      ticket,
    );

    expect(result).toEqual({ status: 'failed', error: 'connection refused' });
  });

  it('returns skipped when host is blank', async () => {
    const result = await sendKdsTicket(
      kdsConfig({ transport: 'network', networkHost: '' }),
      ticket,
    );
    expect(result).toEqual({ status: 'skipped' });
    expect(invoke).not.toHaveBeenCalled();
  });
});
