import React from 'react';
import type { Receipt } from '../types/pos';
import { useSettingsStore } from '../stores/settingsStore';
import { formatCurrency, formatDate } from '../utils/format';

interface ReceiptPreviewProps {
  receipt: Receipt;
  businessName?: string;
}

const ReceiptPreview: React.FC<ReceiptPreviewProps> = ({ receipt, businessName }) => {
  const footer = useSettingsStore((s) => s.receiptFooter);
  return (
    <div className="bg-white text-black w-[320px] mx-auto p-4 font-mono text-xs shadow-lg print:shadow-none">
      <div className="text-center mb-2">
        <div className="font-bold text-sm">{businessName ?? 'Receipt'}</div>
        <div className="text-[10px]">{formatDate(receipt.printedAt)}</div>
        <div className="text-[10px]">Order #{receipt.orderId}</div>
      </div>
      <div className="border-t border-dashed border-black my-2" />
      <div className="space-y-1">
        {receipt.items.map((it, idx) => (
          <div key={idx} className="flex justify-between">
            <span>
              {it.qty}× {it.name}
            </span>
            <span>{formatCurrency(it.price * it.qty)}</span>
          </div>
        ))}
      </div>
      <div className="border-t border-dashed border-black my-2" />
      <div className="flex justify-between font-bold">
        <span>Total</span>
        <span>{formatCurrency(receipt.total)}</span>
      </div>
      {receipt.payments.length > 0 && (
        <div className="mt-2 space-y-1">
          {receipt.payments.map((p) => (
            <div key={p.id} className="flex justify-between text-[10px]">
              <span>{p.method}</span>
              <span>{formatCurrency(p.amount)}</span>
            </div>
          ))}
        </div>
      )}
      <div className="border-t border-dashed border-black my-2" />
      <p className="text-center text-[10px]">{footer}</p>
    </div>
  );
};

export default ReceiptPreview;
