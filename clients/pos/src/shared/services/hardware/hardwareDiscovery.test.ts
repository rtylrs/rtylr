import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';
import {
  listPrinters,
  listSerialPorts,
  discoverNetworkHardware,
} from './hardwareDiscovery';

vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(),
}));

vi.mock('../../utils/platform', () => ({
  canUseNativeHardware: vi.fn(),
}));

describe('hardwareDiscovery', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('when in browser (canUseNativeHardware is false)', () => {
    beforeEach(() => {
      (canUseNativeHardware as Mock).mockReturnValue(false);
    });

    it('listPrinters returns empty array and skips invoke', async () => {
      const result = await listPrinters();
      expect(result).toEqual([]);
      expect(invoke).not.toHaveBeenCalled();
    });

    it('listSerialPorts returns empty array and skips invoke', async () => {
      const result = await listSerialPorts();
      expect(result).toEqual([]);
      expect(invoke).not.toHaveBeenCalled();
    });

    it('discoverNetworkHardware returns empty array and skips invoke', async () => {
      const result = await discoverNetworkHardware('192.168.1');
      expect(result).toEqual([]);
      expect(invoke).not.toHaveBeenCalled();
    });
  });

  describe('when in native desktop app (canUseNativeHardware is true)', () => {
    beforeEach(() => {
      (canUseNativeHardware as Mock).mockReturnValue(true);
    });

    it('listPrinters invokes native commands and returns camelCase mapped printer list', async () => {
      const mockRawPrinters = [
        { name: 'Epson TM-T88VI', is_default: true, connection_type: 'Network' },
        { name: 'Star TSP100', is_default: false, connection_type: 'USB' },
      ];
      (invoke as Mock).mockResolvedValue(mockRawPrinters);

      const result = await listPrinters();

      expect(invoke).toHaveBeenCalledWith('list_printers');
      expect(result).toEqual([
        { name: 'Epson TM-T88VI', isDefault: true, connectionType: 'Network' },
        { name: 'Star TSP100', isDefault: false, connectionType: 'USB' },
      ]);
    });

    it('listSerialPorts invokes native commands and returns camelCase mapped ports list', async () => {
      const mockRawPorts = [
        {
          port_name: 'COM3',
          manufacturer: 'FTDI',
          product: 'USB-to-Serial',
          vid: 1027,
          pid: 24577,
        },
      ];
      (invoke as Mock).mockResolvedValue(mockRawPorts);

      const result = await listSerialPorts();

      expect(invoke).toHaveBeenCalledWith('list_serial_ports');
      expect(result).toEqual([
        {
          portName: 'COM3',
          manufacturer: 'FTDI',
          product: 'USB-to-Serial',
          vid: 1027,
          pid: 24577,
        },
      ]);
    });

    it('discoverNetworkHardware invokes native scan and returns device structures', async () => {
      const mockRawDevices = [
        {
          host: '192.168.1.150',
          port: 9100,
          device_type: 'thermal_escpos',
          label: 'Epson Printer',
        },
      ];
      (invoke as Mock).mockResolvedValue(mockRawDevices);

      const result = await discoverNetworkHardware('192.168.1', 250);

      expect(invoke).toHaveBeenCalledWith('discover_network_hardware', {
        subnetPrefix: '192.168.1',
        timeoutMs: 250,
      });
      expect(result).toEqual([
        {
          host: '192.168.1.150',
          port: 9100,
          deviceType: 'thermal_escpos',
          label: 'Epson Printer',
        },
      ]);
    });
  });
});
