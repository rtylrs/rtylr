import { invoke } from '@tauri-apps/api/core';
import type { PrinterSettings, KdsHardwareSettings } from '../../stores/settingsStore';
import { canUseNativeHardware } from '../../utils/platform';

export interface DeviceStatus {
  name: string;
  status: 'ok' | 'unreachable' | 'not_configured';
  message: string;
}

export interface HardwareHealthReport {
  printer: DeviceStatus;
  drawer: DeviceStatus;
  kds: DeviceStatus;
}

interface HealthCheckConfig {
  printer_transport: string;
  printer_serial_port: string | null;
  printer_baud_rate: number | null;
  printer_host: string | null;
  printer_port: number | null;
  printer_name: string | null;
  drawer_transport: string | null;
  drawer_serial_port: string | null;
  drawer_baud_rate: number | null;
  drawer_host: string | null;
  drawer_port: number | null;
  kds_transport: string;
  kds_serial_port: string | null;
  kds_baud_rate: number | null;
  kds_host: string | null;
  kds_port: number | null;
}

function buildConfig(
  printer: PrinterSettings,
  kds: KdsHardwareSettings,
): HealthCheckConfig {
  return {
    printer_transport: printer.transport,
    printer_serial_port: printer.transport === 'serial' ? printer.serialPort : null,
    printer_baud_rate: printer.transport === 'serial' ? printer.serialBaudRate : null,
    printer_host: printer.transport === 'network' ? printer.networkHost : null,
    printer_port: printer.transport === 'network' ? printer.networkPort : null,
    printer_name: printer.transport === 'windows' ? printer.name : null,
    // Drawer always shares the printer connection in current design
    drawer_transport: 'printer',
    drawer_serial_port: null,
    drawer_baud_rate: null,
    drawer_host: null,
    drawer_port: null,
    kds_transport: kds.transport,
    kds_serial_port: kds.transport === 'serial' ? kds.serialPort : null,
    kds_baud_rate: kds.transport === 'serial' ? kds.serialBaudRate : null,
    kds_host: kds.transport === 'network' ? kds.networkHost : null,
    kds_port: kds.transport === 'network' ? kds.networkPort : null,
  };
}

export async function checkHardwareHealth(
  printer: PrinterSettings,
  kds: KdsHardwareSettings,
): Promise<HardwareHealthReport> {
  if (!canUseNativeHardware()) {
    return {
      printer: {
        name: 'Browser POS',
        status: 'not_configured',
        message: 'Native hardware checks run only in the desktop app.',
      },
      drawer: {
        name: 'Browser POS',
        status: 'not_configured',
        message: 'Native cash drawer checks run only in the desktop app.',
      },
      kds: {
        name: 'Browser POS',
        status: 'not_configured',
        message: 'Native KDS checks run only in the desktop app.',
      },
    };
  }
  const config = buildConfig(printer, kds);
  const raw = await invoke<{
    printer: { name: string; status: string; message: string };
    drawer: { name: string; status: string; message: string };
    kds: { name: string; status: string; message: string };
  }>('check_hardware', { config });
  return {
    printer: raw.printer as DeviceStatus,
    drawer: raw.drawer as DeviceStatus,
    kds: raw.kds as DeviceStatus,
  };
}
