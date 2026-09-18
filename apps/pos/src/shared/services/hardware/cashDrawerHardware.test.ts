import { describe, expect, it, vi } from 'vitest';
import { openConfiguredCashDrawer } from './cashDrawerHardware';

describe('openConfiguredCashDrawer', () => {
  it('invokes the native drawer command with the configured printer name', async () => {
    const invokeCommand = vi.fn(async () => 'ok');

    const result = await openConfiguredCashDrawer({
      enabled: true,
      printerName: ' Receipt Printer ',
      invokeCommand,
    });

    expect(result).toEqual({ status: 'opened' });
    expect(invokeCommand).toHaveBeenCalledWith('open_cash_drawer', {
      printerName: 'Receipt Printer',
    });
  });

  it('skips the native command when drawer kick is disabled', async () => {
    const invokeCommand = vi.fn(async () => 'ok');

    const result = await openConfiguredCashDrawer({
      enabled: false,
      printerName: 'Receipt Printer',
      invokeCommand,
    });

    expect(result).toEqual({ status: 'skipped' });
    expect(invokeCommand).not.toHaveBeenCalled();
  });

  it('returns a failed result when the native command rejects', async () => {
    const invokeCommand = vi.fn(async () => {
      throw new Error('printer missing');
    });

    const result = await openConfiguredCashDrawer({
      enabled: true,
      printerName: 'Receipt Printer',
      invokeCommand,
    });

    expect(result).toEqual({ status: 'failed', error: 'printer missing' });
  });
});
