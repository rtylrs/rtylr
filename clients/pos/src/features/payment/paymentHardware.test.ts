import { describe, expect, it, vi, beforeEach } from 'vitest';
import { printReceiptAndOpenDrawer, openDrawer } from './paymentHardware';
import type { CartItem } from '../../shared/stores/cartStore';
import type { PrinterSettings } from '../../shared/stores/settingsStore';

// Mock all hardware modules so tests never hit Tauri invoke
vi.mock('../../shared/services/hardware/receiptPrinter', () => ({
  printReceipt: vi.fn(),
}));
vi.mock('../../shared/services/hardware/cashDrawerHardware', () => ({
  openConfiguredCashDrawer: vi.fn(),
}));
vi.mock('../../shared/services/hardware/crossPlatformDrawer', () => ({
  openDrawerSerial: vi.fn(),
  openDrawerNetwork: vi.fn(),
}));

import { printReceipt } from '../../shared/services/hardware/receiptPrinter';
import { openConfiguredCashDrawer } from '../../shared/services/hardware/cashDrawerHardware';
import { openDrawerSerial, openDrawerNetwork } from '../../shared/services/hardware/crossPlatformDrawer';

const item: CartItem = {
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Burger',
  unitPrice: 12,
  price: 12,
  quantity: 1,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
};

const baseParams = {
  items: [item],
  subtotal: 12,
  taxAmount: 0.96,
  discountAmount: 0,
  tipAmount: 0,
  total: 12.96,
  paymentMethod: 'cash',
  amountPaid: 15,
  changeDue: 2.04,
  orderNumber: 'ORD-001',
  businessName: 'Test Cafe',
  kickDrawer: true,
};

function makePrinter(overrides: Partial<PrinterSettings> = {}): PrinterSettings {
  return {
    name: 'EPSON TM-T88',
    copies: 1,
    cashDrawerKick: true,
    drawerPin: 2,
    transport: 'windows',
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

// ── printReceiptAndOpenDrawer ─────────────────────────────────────────────────

describe('printReceiptAndOpenDrawer - Windows transport', () => {
  it('prints then kicks drawer via Windows spooler and returns ok', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });

    const result = await printReceiptAndOpenDrawer({
      ...baseParams,
      printer: makePrinter({ transport: 'windows' }),
    });

    expect(result).toEqual({ kind: 'ok' });
    expect(printReceipt).toHaveBeenCalledOnce();
    expect(openConfiguredCashDrawer).toHaveBeenCalledOnce();
  });

  it('returns print_failed when printing fails', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'failed', error: 'printer offline' });

    const result = await printReceiptAndOpenDrawer({
      ...baseParams,
      printer: makePrinter({ transport: 'windows' }),
    });

    expect(result).toEqual({ kind: 'print_failed', error: 'printer offline' });
    expect(openConfiguredCashDrawer).not.toHaveBeenCalled();
  });

  it('returns drawer_failed when drawer kick fails after successful print', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({
      status: 'failed',
      error: 'access denied',
    });

    const result = await printReceiptAndOpenDrawer({
      ...baseParams,
      printer: makePrinter({ transport: 'windows' }),
    });

    expect(result).toEqual({ kind: 'drawer_failed', error: 'access denied' });
  });

  it('skips drawer kick when kickDrawer=false', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });

    const result = await printReceiptAndOpenDrawer({
      ...baseParams,
      kickDrawer: false,
      printer: makePrinter({ transport: 'windows' }),
    });

    expect(result).toEqual({ kind: 'ok' });
    expect(openConfiguredCashDrawer).not.toHaveBeenCalled();
  });
});

describe('printReceiptAndOpenDrawer - Serial transport', () => {
  it('embeds drawer kick in ESC/POS stream (no separate drawer call)', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });

    const result = await printReceiptAndOpenDrawer({
      ...baseParams,
      printer: makePrinter({ transport: 'serial', serialPort: '/dev/ttyUSB0' }),
    });

    expect(result).toEqual({ kind: 'ok' });
    // Drawer kick is embedded in the receipt stream - no separate spooler call
    expect(openConfiguredCashDrawer).not.toHaveBeenCalled();
    // Receipt data should have openDrawer: true
    const receiptArg = vi.mocked(printReceipt).mock.calls[0][1];
    expect(receiptArg.openDrawer).toBe(true);
  });
});

describe('printReceiptAndOpenDrawer - Network transport', () => {
  it('embeds drawer kick in ESC/POS stream over TCP', async () => {
    vi.mocked(printReceipt).mockResolvedValue({ status: 'printed', message: 'receipt printed' });

    const result = await printReceiptAndOpenDrawer({
      ...baseParams,
      printer: makePrinter({ transport: 'network', networkHost: '192.168.1.100' }),
    });

    expect(result).toEqual({ kind: 'ok' });
    expect(openConfiguredCashDrawer).not.toHaveBeenCalled();
    const receiptArg = vi.mocked(printReceipt).mock.calls[0][1];
    expect(receiptArg.openDrawer).toBe(true);
  });
});

// ── openDrawer ───────────────────────────────────────────────────────────────

describe('openDrawer', () => {
  it('skips when cashDrawerKick is disabled', async () => {
    const result = await openDrawer(makePrinter({ cashDrawerKick: false }));
    expect(result).toEqual({ kind: 'ok' });
    expect(openConfiguredCashDrawer).not.toHaveBeenCalled();
    expect(openDrawerSerial).not.toHaveBeenCalled();
    expect(openDrawerNetwork).not.toHaveBeenCalled();
  });

  it('uses Windows spooler for windows transport', async () => {
    vi.mocked(openConfiguredCashDrawer).mockResolvedValue({ status: 'opened' });
    const result = await openDrawer(makePrinter({ transport: 'windows' }));
    expect(result).toEqual({ kind: 'ok' });
    expect(openConfiguredCashDrawer).toHaveBeenCalledOnce();
  });

  it('uses serial for serial transport', async () => {
    vi.mocked(openDrawerSerial).mockResolvedValue({ status: 'opened', via: 'serial' });
    const result = await openDrawer(makePrinter({ transport: 'serial', serialPort: 'COM3' }));
    expect(result).toEqual({ kind: 'ok' });
    expect(openDrawerSerial).toHaveBeenCalledWith('COM3', 9600, 2);
  });

  it('uses network for network transport', async () => {
    vi.mocked(openDrawerNetwork).mockResolvedValue({ status: 'opened', via: 'network' });
    const result = await openDrawer(
      makePrinter({ transport: 'network', networkHost: '10.0.0.1', networkPort: 9100 }),
    );
    expect(result).toEqual({ kind: 'ok' });
    expect(openDrawerNetwork).toHaveBeenCalledWith('10.0.0.1', 9100, 2);
  });

  it('returns drawer_failed on serial error', async () => {
    vi.mocked(openDrawerSerial).mockResolvedValue({ status: 'failed', error: 'port busy' });
    const result = await openDrawer(makePrinter({ transport: 'serial', serialPort: 'COM3' }));
    expect(result).toEqual({ kind: 'drawer_failed', error: 'port busy' });
  });
});
