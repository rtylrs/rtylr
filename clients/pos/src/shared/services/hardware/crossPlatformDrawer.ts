import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';

export type DrawerPin = 2 | 5;

export type DrawerResult =
  | { status: 'opened'; via: string }
  | { status: 'skipped' }
  | { status: 'failed'; error: string };

/** Open drawer via a serial port (RJ11/RJ12 or USB-to-serial adapter). */
export async function openDrawerSerial(
  portName: string,
  baudRate: number = 9600,
  pin: DrawerPin = 2,
): Promise<DrawerResult> {
  if (!portName.trim() || !canUseNativeHardware()) return { status: 'skipped' };
  try {
    const msg = await invoke<string>('open_cash_drawer_serial', {
      portName: portName.trim(),
      baudRate,
      pin,
    });
    return { status: 'opened', via: msg };
  } catch (err) {
    return { status: 'failed', error: err instanceof Error ? err.message : String(err) };
  }
}

/** Open drawer over TCP (e.g. network-connected thermal printer). */
export async function openDrawerNetwork(
  host: string,
  port: number = 9100,
  pin: DrawerPin = 2,
): Promise<DrawerResult> {
  if (!host.trim() || !canUseNativeHardware()) return { status: 'skipped' };
  try {
    const msg = await invoke<string>('open_cash_drawer_network', {
      host: host.trim(),
      port,
      pin,
    });
    return { status: 'opened', via: msg };
  } catch (err) {
    return { status: 'failed', error: err instanceof Error ? err.message : String(err) };
  }
}
