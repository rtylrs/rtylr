import { invoke } from '@tauri-apps/api/core';
import { canUseNativeHardware } from '../../utils/platform';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface ReceiptItem {
  name: string;
  quantity: number;
  unitPrice: number;
  total: number;
  modifiers: string[];
  notes?: string;
}

export interface ReceiptData {
  businessName: string;
  businessAddress?: string;
  businessPhone?: string;
  orderNumber: string;
  orderType?: string;
  cashierName?: string;
  tableNumber?: string;
  items: ReceiptItem[];
  subtotal: number;
  taxAmount: number;
  discountAmount: number;
  tipAmount: number;
  total: number;
  paymentMethod: string;
  amountPaid: number;
  changeDue: number;
  footerMessage?: string;
  /** Kick cash drawer immediately before cut */
  openDrawer?: boolean;
  /** 2 or 5 (default 2) */
  drawerPin?: 2 | 5;
}

/** Transport for sending ESC/POS data to the printer. */
export type PrinterTransport =
  | { type: 'windows'; printerName: string }
  | { type: 'serial'; portName: string; baudRate?: number }
  | { type: 'network'; host: string; port?: number };

export type PrintResult =
  | { status: 'printed'; message: string }
  | { status: 'skipped'; message: string }
  | { status: 'failed'; error: string };

// ── Command ───────────────────────────────────────────────────────────────────

/**
 * Print a receipt via the Tauri backend.
 * Supports Windows spooler, serial port, and TCP/IP transports.
 */
export async function printReceipt(
  transport: PrinterTransport,
  receipt: ReceiptData,
): Promise<PrintResult> {
  if (!canUseNativeHardware()) {
    // Report the skip honestly - callers must not treat it as a successful
    // print (matches the drawer/KDS modules' 'skipped' convention).
    return { status: 'skipped', message: 'Receipt print skipped in browser POS.' };
  }
  const config = buildConfig(transport);
  const receiptPayload = toSnakeCase(receipt);

  try {
    const message = await invoke<string>('print_receipt', {
      config,
      receipt: receiptPayload,
    });
    return { status: 'printed', message };
  } catch (err) {
    return {
      status: 'failed',
      error: err instanceof Error ? err.message : String(err),
    };
  }
}

function buildConfig(transport: PrinterTransport): Record<string, unknown> {
  switch (transport.type) {
    case 'windows':
      return { transport: 'windows', target: transport.printerName, port: null };
    case 'serial':
      return {
        transport: 'serial',
        target: transport.portName,
        port: transport.baudRate ?? 9600,
      };
    case 'network':
      return {
        transport: 'network',
        target: transport.host,
        port: transport.port ?? 9100,
      };
  }
}

function toSnakeCase(receipt: ReceiptData): Record<string, unknown> {
  return {
    business_name: receipt.businessName,
    business_address: receipt.businessAddress ?? null,
    business_phone: receipt.businessPhone ?? null,
    order_number: receipt.orderNumber,
    order_type: receipt.orderType ?? null,
    cashier_name: receipt.cashierName ?? null,
    table_number: receipt.tableNumber ?? null,
    items: receipt.items.map((item) => ({
      name: item.name,
      quantity: item.quantity,
      unit_price: item.unitPrice,
      total: item.total,
      modifiers: item.modifiers,
      notes: item.notes ?? null,
    })),
    subtotal: receipt.subtotal,
    tax_amount: receipt.taxAmount,
    discount_amount: receipt.discountAmount,
    tip_amount: receipt.tipAmount,
    total: receipt.total,
    payment_method: receipt.paymentMethod,
    amount_paid: receipt.amountPaid,
    change_due: receipt.changeDue,
    footer_message: receipt.footerMessage ?? null,
    open_drawer: receipt.openDrawer ?? false,
    drawer_pin: receipt.drawerPin ?? 2,
  };
}
