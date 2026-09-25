import { openConfiguredCashDrawer } from '../../shared/services/hardware/cashDrawerHardware';
import { openDrawerSerial, openDrawerNetwork } from '../../shared/services/hardware/crossPlatformDrawer';
import { printReceipt } from '../../shared/services/hardware/receiptPrinter';
import type { ReceiptData, PrinterTransport } from '../../shared/services/hardware/receiptPrinter';
import type { PrinterSettings } from '../../shared/stores/settingsStore';
import type { CartItem } from '../../shared/stores/cartStore';

export interface PrintReceiptParams {
  printer: PrinterSettings;
  items: CartItem[];
  subtotal: number;
  taxAmount: number;
  discountAmount: number;
  tipAmount: number;
  total: number;
  paymentMethod: string;
  amountPaid: number;
  changeDue: number;
  orderNumber: string;
  orderType?: string;
  cashierName?: string;
  tableNumber?: string;
  businessName: string;
  footerMessage?: string;
  kickDrawer: boolean;
}

export type PrintOutcome =
  | { kind: 'ok' }
  | { kind: 'print_failed'; error: string }
  | { kind: 'drawer_failed'; error: string };

/** Print receipt and open cash drawer using the configured transport. */
export async function printReceiptAndOpenDrawer(
  params: PrintReceiptParams,
): Promise<PrintOutcome> {
  const transport = resolveTransport(params.printer);
  const receiptData = buildReceiptData(params, transport);

  // Print (and embedded drawer kick for serial/network)
  const result = await printReceipt(transport, receiptData);
  // 'skipped' (browser POS) is not a successful print, but it must not
  // block checkout either - only failures do.
  if (result.status === 'failed') {
    return { kind: 'print_failed', error: result.error };
  }

  // Windows: kick drawer separately via spooler API (more reliable than embedded ESC/POS)
  if (transport.type === 'windows' && params.kickDrawer) {
    const drawerResult = await openConfiguredCashDrawer({
      enabled: true,
      printerName: params.printer.name,
    });
    if (drawerResult.status === 'failed') {
      return { kind: 'drawer_failed', error: drawerResult.error };
    }
  }

  return { kind: 'ok' };
}

/** Open cash drawer only (no receipt), routing through the correct transport. */
export async function openDrawer(printer: PrinterSettings): Promise<PrintOutcome> {
  if (!printer.cashDrawerKick) return { kind: 'ok' };

  switch (printer.transport) {
    case 'windows': {
      const r = await openConfiguredCashDrawer({
        enabled: true,
        printerName: printer.name,
      });
      return r.status === 'failed' ? { kind: 'drawer_failed', error: r.error } : { kind: 'ok' };
    }
    case 'serial': {
      const r = await openDrawerSerial(printer.serialPort, printer.serialBaudRate, printer.drawerPin);
      return r.status === 'failed' ? { kind: 'drawer_failed', error: r.error } : { kind: 'ok' };
    }
    case 'network': {
      const r = await openDrawerNetwork(printer.networkHost, printer.networkPort, printer.drawerPin);
      return r.status === 'failed' ? { kind: 'drawer_failed', error: r.error } : { kind: 'ok' };
    }
  }
}

// ── Internal helpers ─────────────────────────────────────────────────────────

function resolveTransport(printer: PrinterSettings): PrinterTransport {
  switch (printer.transport) {
    case 'serial':
      return { type: 'serial', portName: printer.serialPort, baudRate: printer.serialBaudRate };
    case 'network':
      return { type: 'network', host: printer.networkHost, port: printer.networkPort };
    case 'windows':
    default:
      return { type: 'windows', printerName: printer.name };
  }
}

function buildReceiptData(
  params: PrintReceiptParams,
  transport: PrinterTransport,
): ReceiptData {
  // For serial/network, embed the drawer kick in the ESC/POS stream
  const embedDrawer = params.kickDrawer && transport.type !== 'windows';

  return {
    businessName: params.businessName,
    orderNumber: params.orderNumber,
    orderType: params.orderType,
    cashierName: params.cashierName,
    tableNumber: params.tableNumber,
    items: params.items.map((item) => ({
      name: item.name,
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      total: item.unitPrice * item.quantity,
      modifiers: item.modifiers.map((m) => m.name),
      notes: item.notes,
    })),
    subtotal: params.subtotal,
    taxAmount: params.taxAmount,
    discountAmount: params.discountAmount,
    tipAmount: params.tipAmount,
    total: params.total,
    paymentMethod: params.paymentMethod,
    amountPaid: params.amountPaid,
    changeDue: params.changeDue,
    footerMessage: params.footerMessage,
    openDrawer: embedDrawer,
    drawerPin: params.printer.drawerPin,
  };
}
