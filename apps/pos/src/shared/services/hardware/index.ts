// Cash drawer - legacy Windows-spooler path
export { openConfiguredCashDrawer } from './cashDrawerHardware';

// Cash drawer - cross-platform (serial + network)
export { openDrawerSerial, openDrawerNetwork } from './crossPlatformDrawer';
export type { DrawerResult, DrawerPin } from './crossPlatformDrawer';

// Hardware discovery
export { listPrinters, listSerialPorts } from './hardwareDiscovery';
export type { PrinterInfo, SerialPortInfo } from './hardwareDiscovery';

// ESC/POS receipt printing
export { printReceipt } from './receiptPrinter';
export type {
  ReceiptData,
  ReceiptItem,
  PrinterTransport,
  PrintResult,
} from './receiptPrinter';

// KDS hardware display
export { sendKdsTicket } from './kdsHardware';
export type { KdsTicket, KdsTicketItem, KdsHardwareResult } from './kdsHardware';

// Hardware health check
export { checkHardwareHealth } from './hardwareHealth';
export type { DeviceStatus, HardwareHealthReport } from './hardwareHealth';
