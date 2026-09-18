import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';
import { checkHardwareHealth } from './hardwareHealth';
import type { PrinterSettings, KdsHardwareSettings } from '../../stores/settingsStore';

vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(),
}));

vi.mock('../../utils/platform', () => ({
  canUseNativeHardware: vi.fn(),
}));

describe('hardwareHealth', () => {
  const mockPrinterSettings: PrinterSettings = {
    name: 'Thermal Printer',
    copies: 1,
    cashDrawerKick: true,
    drawerPin: 2,
    transport: 'serial',
    serialPort: 'COM3',
    serialBaudRate: 9600,
    networkHost: '',
    networkPort: 9100,
  };

  const mockKdsSettings: KdsHardwareSettings = {
    transport: 'network',
    serialPort: '',
    serialBaudRate: 9600,
    networkHost: '192.168.1.120',
    networkPort: 9101,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('when in browser (canUseNativeHardware is false): returns skipped report without invoking Tauri commands', async () => {
    (canUseNativeHardware as Mock).mockReturnValue(false);

    const report = await checkHardwareHealth(mockPrinterSettings, mockKdsSettings);

    expect(report.printer.status).toBe('not_configured');
    expect(report.printer.message).toContain('desktop app');
    expect(report.drawer.status).toBe('not_configured');
    expect(report.kds.status).toBe('not_configured');
    expect(invoke).not.toHaveBeenCalled();
  });

  it('when in native app (canUseNativeHardware is true): constructs correct config and invokes check_hardware', async () => {
    (canUseNativeHardware as Mock).mockReturnValue(true);

    const mockInvokeResponse = {
      printer: { name: 'Receipt Printer', status: 'ok', message: 'Online' },
      drawer: { name: 'Cash Drawer', status: 'ok', message: 'Ready' },
      kds: { name: 'Kitchen Screen', status: 'unreachable', message: 'Ping timeout' },
    };
    (invoke as Mock).mockResolvedValue(mockInvokeResponse);

    const report = await checkHardwareHealth(mockPrinterSettings, mockKdsSettings);

    expect(invoke).toHaveBeenCalledWith('check_hardware', {
      config: {
        printer_transport: 'serial',
        printer_serial_port: 'COM3',
        printer_baud_rate: 9600,
        printer_host: null,
        printer_port: null,
        printer_name: null,
        drawer_transport: 'printer',
        drawer_serial_port: null,
        drawer_baud_rate: null,
        drawer_host: null,
        drawer_port: null,
        kds_transport: 'network',
        kds_serial_port: null,
        kds_baud_rate: null,
        kds_host: '192.168.1.120',
        kds_port: 9101,
      },
    });

    expect(report).toEqual(mockInvokeResponse);
  });
});
