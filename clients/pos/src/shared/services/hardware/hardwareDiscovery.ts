import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';

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
  if (!canUseNativeHardware()) return [];
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
  if (!canUseNativeHardware()) return [];
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

export interface DiscoveredDevice {
  host: string;
  port: number;
  deviceType: 'thermal_escpos' | 'network_lpr' | 'network_ipp' | 'kds' | 'unknown';
  label: string;
}

/**
 * Scan a /24 subnet for printers and KDS displays.
 * subnetPrefix: first three octets, e.g. "192.168.1"
 * timeoutMs: per-host per-port TCP connect timeout (50–5000 ms)
 */
export async function discoverNetworkHardware(
  subnetPrefix: string,
  timeoutMs = 150,
): Promise<DiscoveredDevice[]> {
  if (!canUseNativeHardware()) return [];
  const raw = await invoke<Array<{
    host: string;
    port: number;
    device_type: string;
    label: string;
  }>>('discover_network_hardware', {
    subnetPrefix,
    timeoutMs,
  });
  return raw.map((d) => ({
    host: d.host,
    port: d.port,
    deviceType: d.device_type as DiscoveredDevice['deviceType'],
    label: d.label,
  }));
}
