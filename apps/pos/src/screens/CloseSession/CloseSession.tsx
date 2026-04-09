import React, { useState } from 'react';
import { motion } from 'framer-motion';
import {
  TrendingUp,
  ShoppingCart,
  Banknote,
  CreditCard,
  Wifi,
  Printer,
  LogOut,
  AlertTriangle,
  CheckCircle2,
  DollarSign,
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { usePOSStore } from '../../stores/posStore';

const CloseSession: React.FC = () => {
  const navigate = useNavigate();
  const { orders, staffName, sessionStart } = usePOSStore();

  const [cashActual, setCashActual] = useState('');
  const [sessionClosed, setSessionClosed] = useState(false);

  // Session duration
  const elapsed = Math.floor((Date.now() - sessionStart.getTime()) / 1000);
  const h = Math.floor(elapsed / 3600);
  const m = Math.floor((elapsed % 3600) / 60);
  const sessionDuration = h > 0 ? `${h}h ${m}m` : `${m}m`;

  // Compute totals from paid orders
  const paidOrders = orders.filter((o) => o.status === 'paid');
  const totalRevenue = paidOrders.reduce((s, o) => s + o.total, 0);
  const totalOrders  = paidOrders.length;

  const cashTotal   = paidOrders.filter((o) => o.paymentMethod === 'cash').reduce((s, o) => s + o.total, 0);
  const cardTotal   = paidOrders.filter((o) => o.paymentMethod === 'card').reduce((s, o) => s + o.total, 0);
  const onlineTotal = paidOrders.filter((o) => o.paymentMethod === 'online').reduce((s, o) => s + o.total, 0);

  const cashActualNum = parseFloat(cashActual) || 0;
  const cashVariance  = cashActualNum - cashTotal;
  const hasVariance   = cashActual !== '' && Math.abs(cashVariance) > 0.01;

  const handlePrint = () => {
    toast.success('Report sent to printer');
  };

  const handleClose = () => {
    setSessionClosed(true);
    toast.success('Session closed. Goodbye!');
    setTimeout(() => {
      window.location.href = 'http://localhost:5175';
    }, 2000);
  };

  if (sessionClosed) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-4" style={{ background: 'var(--hub-bg)' }}>
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: 'spring', stiffness: 300, damping: 20 }}
          className="flex items-center justify-center h-16 w-16 rounded-full"
          style={{ background: 'rgba(16,185,129,0.15)' }}
        >
          <CheckCircle2 size={32} style={{ color: '#10b981' }} />
        </motion.div>
        <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>Session Closed</p>
        <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>Redirecting to Hub…</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full overflow-y-auto">
      {/* Header */}
      <div
        className="shrink-0 px-6 py-4"
        style={{ borderBottom: '1px solid var(--hub-border)', background: 'var(--hub-surface)' }}
      >
        <div className="flex items-center gap-3">
          <span
            className="flex items-center justify-center h-9 w-9 rounded-xl"
            style={{ background: 'rgba(239,68,68,0.12)' }}
          >
            <LogOut size={16} style={{ color: '#ef4444' }} />
          </span>
          <div>
            <h2 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              End of Session
            </h2>
            <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              Staff: {staffName} · Session duration: {sessionDuration}
            </p>
          </div>
        </div>
      </div>

      <div className="flex-1 px-6 py-6 space-y-6 max-w-3xl mx-auto w-full">
        {/* Summary stats */}
        <section>
          <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--hub-text-dim)' }}>
            Session Summary
          </h3>
          <div className="grid grid-cols-2 gap-3">
            {[
              { label: 'Total Sales',    value: `$${totalRevenue.toFixed(2)}`, icon: TrendingUp,  color: '#10b981' },
              { label: 'Orders Closed',  value: String(totalOrders),            icon: ShoppingCart, color: '#4583C4' },
            ].map(({ label, value, icon: Icon, color }) => (
              <motion.div
                key={label}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                className="flex items-center gap-4 px-5 py-4 rounded-xl"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <span
                  className="flex items-center justify-center h-10 w-10 rounded-xl shrink-0"
                  style={{ background: color + '18' }}
                >
                  <Icon size={18} style={{ color }} />
                </span>
                <div>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{label}</p>
                  <p className="text-2xl font-bold tabular-nums" style={{ color: 'var(--hub-text)' }}>{value}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </section>

        {/* Payment breakdown */}
        <section>
          <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--hub-text-dim)' }}>
            Payment Breakdown
          </h3>
          <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--hub-border)' }}>
            {[
              { method: 'Cash',   icon: Banknote,   total: cashTotal,   color: '#f59e0b' },
              { method: 'Card',   icon: CreditCard, total: cardTotal,   color: '#4583C4' },
              { method: 'Online', icon: Wifi,        total: onlineTotal, color: '#a78bfa' },
            ].map(({ method, icon: Icon, total, color }, idx) => (
              <div
                key={method}
                className="flex items-center gap-4 px-5 py-3.5"
                style={{
                  borderBottom: idx < 2 ? '1px solid var(--hub-border)' : 'none',
                  background: idx % 2 === 0 ? 'transparent' : 'var(--hub-hover)',
                }}
              >
                <span
                  className="flex items-center justify-center h-8 w-8 rounded-lg shrink-0"
                  style={{ background: color + '18' }}
                >
                  <Icon size={14} style={{ color }} />
                </span>
                <span className="flex-1 text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  {method}
                </span>
                <span className="text-sm font-bold tabular-nums" style={{ color: 'var(--hub-text)' }}>
                  ${total.toFixed(2)}
                </span>
              </div>
            ))}
          </div>
        </section>

        {/* Cash drawer reconciliation */}
        <section>
          <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--hub-text-dim)' }}>
            Cash Drawer Reconciliation
          </h3>
          <div
            className="rounded-xl p-5 space-y-3"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center justify-between text-sm">
              <span style={{ color: 'var(--hub-text-muted)' }}>Expected cash in drawer</span>
              <span className="font-bold tabular-nums" style={{ color: 'var(--hub-text)' }}>
                ${cashTotal.toFixed(2)}
              </span>
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm shrink-0" style={{ color: 'var(--hub-text-muted)' }}>
                Actual cash counted
              </label>
              <div className="relative flex-1">
                <span
                  className="absolute left-3 top-1/2 -translate-y-1/2 text-sm font-semibold"
                  style={{ color: 'var(--hub-text-dim)' }}
                >
                  $
                </span>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="0.00"
                  value={cashActual}
                  onChange={(e) => setCashActual(e.target.value)}
                  className="w-full pl-7 pr-3 py-2 rounded-lg text-sm outline-none tabular-nums"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                  onFocus={(e) => { e.currentTarget.style.borderColor = 'var(--accent-border)'; }}
                  onBlur={(e)  => { e.currentTarget.style.borderColor = 'var(--hub-border)'; }}
                />
              </div>
            </div>

            {cashActual !== '' && (
              <motion.div
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                className="flex items-center gap-2 px-3 py-2.5 rounded-lg"
                style={{
                  background: hasVariance ? 'rgba(239,68,68,0.08)' : 'rgba(16,185,129,0.08)',
                  border: `1px solid ${hasVariance ? 'rgba(239,68,68,0.2)' : 'rgba(16,185,129,0.2)'}`,
                }}
              >
                {hasVariance ? (
                  <AlertTriangle size={14} style={{ color: '#ef4444' }} />
                ) : (
                  <CheckCircle2 size={14} style={{ color: '#10b981' }} />
                )}
                <span
                  className="text-sm font-semibold"
                  style={{ color: hasVariance ? '#ef4444' : '#10b981' }}
                >
                  {hasVariance
                    ? `Variance: ${cashVariance > 0 ? '+' : ''}$${cashVariance.toFixed(2)}`
                    : 'Cash balanced — no variance'}
                </span>
              </motion.div>
            )}
          </div>
        </section>

        {/* Actions */}
        <section className="flex gap-3 pb-6">
          <button
            onClick={handlePrint}
            className="flex items-center gap-2 px-5 py-3 rounded-xl text-sm font-semibold transition-colors hover:bg-[var(--hub-hover)]"
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text-muted)',
            }}
          >
            <Printer size={15} />
            Print Report
          </button>

          <motion.button
            whileTap={{ scale: 0.98 }}
            onClick={handleClose}
            className="flex items-center gap-2 px-6 py-3 rounded-xl text-sm font-semibold transition-opacity hover:opacity-90"
            style={{ background: '#ef4444', color: '#fff' }}
          >
            <LogOut size={15} />
            Close Session
          </motion.button>
        </section>
      </div>
    </div>
  );
};

export default CloseSession;
