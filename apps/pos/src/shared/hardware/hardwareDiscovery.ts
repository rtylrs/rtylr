import { invoke } from '@tauri-apps/api/core';

export interface PrinterInfo {
  name: string;
  isDefault: boolean;
  connectionType: 'USB' | 'Serial' | 'Network' | 'Parallel' | 'Unknown';
}

export interface SerialPortInfo {
  portName: string;
  manufacturer: string | null;
  product: string | null;
  vid: number | null;
  pid: number | null;
}

/** List all printers available to the OS (Windows: WinAPI, macOS/Linux: CUPS). */
export async function listPrinters(): Promise<PrinterInfo[]> {
  const raw = await invoke<Array<{
    name: string;
    is_default: boolean;
    connection_type: string;
  }>>('list_printers');
  return raw.map((p) => ({
    name: p.name,
    isDefault: p.is_default,
    connectionType: p.connection_type as PrinterInfo['connectionType'],
  }));
}

/** List all serial ports, including USB-to-serial adapters. */
export async function listSerialPorts(): Promise<SerialPortInfo[]> {
  const raw = await invoke<Array<{
    port_name: string;
    manufacturer: string | null;
    product: string | null;
    vid: number | null;
    pid: number | null;
  }>>('list_serial_ports');
  return raw.map((p) => ({
    portName: p.port_name,
    manufacturer: p.manufacturer,
    product: p.product,
    vid: p.vid,
    pid: p.pid,
  }));
}
