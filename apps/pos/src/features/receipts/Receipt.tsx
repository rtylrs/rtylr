import React, { useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'motion/react';
import POSLayout from '../../layouts/POSLayout';
import ReceiptPreview from '../../shared/components/ReceiptPreview';
import { useReceiptsStore } from '../../shared/stores/receiptsStore';

const Receipt: React.FC = () => {
  const navigate = useNavigate();
  const receipt = useReceiptsStore((s) => s.receipts[0]);
  const orderRef = useRef(receipt?.orderId ?? `ORD-${Date.now().toString(36).toUpperCase()}`);

  if (!receipt) {
    navigate('/');
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
              ✓
            </motion.div>
          </div>

          <div className="text-center">
            <h1 className="text-xl font-bold text-pos-text">Order Placed</h1>
            <p className="text-xs text-pos-muted mt-1">
              {orderRef.current} ·{' '}
              {new Date().toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
            </p>
          </div>

          <ReceiptPreview receipt={receipt} businessName="Vbitez POS" />

          <div className="flex gap-3">
            <button
              className="flex-1 py-2.5 rounded-xl border border-pos-border text-pos-muted text-sm hover:border-white/20 transition-colors"
              onClick={() => window.print()}
            >
              🖨 Print
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
