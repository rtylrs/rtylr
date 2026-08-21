import React, { useEffect, useRef, useState } from 'react';
import { Check, Printer } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import POSLayout from '../../layouts/POSLayout';
import ReceiptPreview from '../../shared/components/ReceiptPreview';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { printReceipt } from '../../shared/services/hardware/receiptPrinter';
import type { PrinterTransport } from '../../shared/services/hardware/receiptPrinter';

const Receipt: React.FC = () => {
  const navigate = useNavigate();
  const receipt = useReceiptsStore((s) => s.receipts[0]);
  const printer = useSettingsStore((s) => s.printer);
  const business = useSettingsStore((s) => s.business);
  const receiptFooter = useSettingsStore((s) => s.receiptFooter);
  const orderRef = useRef(receipt?.orderId ?? `ORD-${Date.now().toString(36).toUpperCase()}`);
  const [printing, setPrinting] = useState(false);

  const handlePrint = async () => {
    if (!receipt) return;
    setPrinting(true);
    try {
      const transport: PrinterTransport =
        printer.transport === 'serial'
          ? { type: 'serial', portName: printer.serialPort, baudRate: printer.serialBaudRate }
          : printer.transport === 'network'
            ? { type: 'network', host: printer.networkHost, port: printer.networkPort }
            : { type: 'windows', printerName: printer.name };

      const result = await printReceipt(transport, {
        businessName: business.name,
        businessAddress: business.address || undefined,
        businessPhone: business.phone || undefined,
        orderNumber: receipt.orderId,
        items: receipt.items.map((it) => ({
          name: it.name,
          quantity: it.qty,
          unitPrice: it.price,
          total: it.price * it.qty,
          modifiers: [],
        })),
        subtotal: receipt.total,
        taxAmount: 0,
        discountAmount: 0,
        tipAmount: 0,
        total: receipt.total,
        paymentMethod: receipt.payments[0]?.method ?? 'cash',
        amountPaid: receipt.payments.reduce((s, p) => s + p.amount, 0),
        changeDue: 0,
        footerMessage: receiptFooter,
        openDrawer: false,
        drawerPin: printer.drawerPin,
      });

      if (result.status === 'failed') {
        toast.error(`Print failed: ${result.error}`);
      } else {
        toast.success('Receipt printed');
      }
    } finally {
      setPrinting(false);
    }
  };

  // Redirect in an effect: navigating during render can update a component
  // while React is rendering another one.
  useEffect(() => {
    if (!receipt) navigate('/');
  }, [receipt, navigate]);

  if (!receipt) {
    return null;
  }

  return (
    <POSLayout>
      <div className="h-full flex items-center justify-center p-6">
        <motion.div
          className="w-full max-w-sm bg-pos-card border border-pos-border rounded-2xl p-6 space-y-5"
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          {/* Check */}
          <div className="flex justify-center">
            <motion.div
              className="w-16 h-16 rounded-full bg-success/15 border-2 border-success/30 flex items-center justify-center text-success text-2xl font-bold"
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              transition={{ type: 'spring', damping: 14, delay: 0.15 }}
            >
              <Check className="h-8 w-8" aria-hidden="true" strokeWidth={2.4} />
            </motion.div>
          </div>

          <div className="text-center">
            <h1 className="text-xl font-bold text-pos-text">Order Placed</h1>
            <p className="text-xs text-pos-muted mt-1">
              {orderRef.current} ·{' '}
              {new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
            </p>
          </div>

          <ReceiptPreview receipt={receipt} businessName={business.name} />

          <div className="flex gap-3">
            <button
              className="flex-1 py-2.5 rounded-xl border border-pos-border text-pos-muted text-sm hover:border-white/20 transition-colors disabled:opacity-50"
              onClick={handlePrint}
              disabled={printing}
            >
              {printing ? (
                '...'
              ) : (
                <span className="inline-flex items-center justify-center gap-2">
                  <Printer className="h-4 w-4" aria-hidden="true" />
                  Print
                </span>
              )}
            </button>
            <button
              className="flex-1 py-2.5 rounded-xl bg-pos-text text-pos-bg text-sm font-semibold hover:opacity-90 transition-opacity"
              onClick={() => {
                navigate('/');
              }}
            >
              New Order
            </button>
          </div>
        </motion.div>
      </div>
    </POSLayout>
  );
};

export default Receipt;
