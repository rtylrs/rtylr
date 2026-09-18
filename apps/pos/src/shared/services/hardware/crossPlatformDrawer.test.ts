import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';
import { openDrawerSerial, openDrawerNetwork } from './crossPlatformDrawer';

vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(),
}));

vi.mock('../../utils/platform', () => ({
  canUseNativeHardware: vi.fn(),
}));

describe('crossPlatformDrawer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('openDrawerSerial', () => {
    it('when in browser: skips native invoke and returns status skipped', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(false);

      const result = await openDrawerSerial('COM3', 9600, 2);

      expect(result).toEqual({ status: 'skipped' });
      expect(invoke).not.toHaveBeenCalled();
    });

    it('when port is empty: skips native invoke and returns status skipped', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(true);

      const result = await openDrawerSerial('   ');

      expect(result).toEqual({ status: 'skipped' });
      expect(invoke).not.toHaveBeenCalled();
    });

    it('when online: calls open_cash_drawer_serial and returns opened', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(true);
      (invoke as Mock).mockResolvedValue('COM3 opened');

      const result = await openDrawerSerial('COM3', 9600, 2);

      expect(invoke).toHaveBeenCalledWith('open_cash_drawer_serial', {
        portName: 'COM3',
        baudRate: 9600,
        pin: 2,
      });
      expect(result).toEqual({ status: 'opened', via: 'COM3 opened' });
    });

    it('when command fails: returns failed with the error message', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(true);
      (invoke as Mock).mockRejectedValue(new Error('Port busy'));

      const result = await openDrawerSerial('COM3', 9600, 2);

      expect(result).toEqual({ status: 'failed', error: 'Port busy' });
    });
  });

  describe('openDrawerNetwork', () => {
    it('when in browser: skips native invoke and returns status skipped', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(false);

      const result = await openDrawerNetwork('192.168.1.100', 9100, 2);

      expect(result).toEqual({ status: 'skipped' });
      expect(invoke).not.toHaveBeenCalled();
    });

    it('when host is empty: skips native invoke and returns status skipped', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(true);

      const result = await openDrawerNetwork('');

      expect(result).toEqual({ status: 'skipped' });
      expect(invoke).not.toHaveBeenCalled();
    });

    it('when online: calls open_cash_drawer_network and returns opened', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(true);
      (invoke as Mock).mockResolvedValue('Connection established');

      const result = await openDrawerNetwork('192.168.1.100', 9100, 5);

      expect(invoke).toHaveBeenCalledWith('open_cash_drawer_network', {
        host: '192.168.1.100',
        port: 9100,
        pin: 5,
      });
      expect(result).toEqual({ status: 'opened', via: 'Connection established' });
    });

    it('when command fails: returns failed with the error message', async () => {
      (canUseNativeHardware as Mock).mockReturnValue(true);
      (invoke as Mock).mockRejectedValue(new Error('Connection timed out'));

      const result = await openDrawerNetwork('192.168.1.100', 9100, 2);

      expect(result).toEqual({ status: 'failed', error: 'Connection timed out' });
    });
  });
});
