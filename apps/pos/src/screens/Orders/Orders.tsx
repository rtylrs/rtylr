import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import {
  ClipboardList,
  CheckCircle2,
  Clock,
  XCircle,
  CreditCard,
  Banknote,
  Wifi,
  TrendingUp,
  ShoppingCart,
  DollarSign,
} from 'lucide-react';
import { usePOSStore } from '../../stores/posStore';
import type { CompletedOrder } from '../../stores/posStore';
import { posApi } from '../../services/posApi';

type StatusFilter = 'all' | 'paid' | 'open' | 'voided';

const STATUS_META: Record<
  CompletedOrder['status'],
  { label: string; color: string; bg: string; icon: React.ComponentType<{ size?: number }> }
> = {
  paid:   { label: 'Paid',   color: '#10b981', bg: 'rgba(16,185,129,0.12)',  icon: CheckCircle2 },
  open:   { label: 'Open',   color: '#f59e0b', bg: 'rgba(245,158,11,0.12)',  icon: Clock        },
  voided: { label: 'Voided', color: '#ef4444', bg: 'rgba(239,68,68,0.12)',   icon: XCircle      },
};

const PAYMENT_META: Record<string, { label: string; icon: React.ComponentType<{ size?: number }> }> = {
  cash:   { label: 'Cash',   icon: Banknote   },
  card:   { label: 'Card',   icon: CreditCard },
  online: { label: 'Online', icon: Wifi       },
};

function formatTime(date: Date): string {
  return date.toLocaleTimeString('en-AU', { hour: '2-digit', minute: '2-digit' });
}

function formatRelative(date: Date): string {
  const diffMs = Date.now() - date.getTime();
  const mins = Math.floor(diffMs / 60_000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  return `${hours}h ${mins % 60}m ago`;
}

const Orders: React.FC = () => {
  const { orders, setOrders } = usePOSStore();
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');

  useEffect(() => {
    posApi.getOrders().then((apiOrders) => {
      setOrders(apiOrders.map((o) => ({
        id: o.id,
        tableId: o.tableId,
        tableLabel: o.tableLabel,
        items: o.items,
        subtotal: o.subtotal,
        tax: o.tax,
        total: o.total,
        paymentMethod: o.paymentMethod,
        status: o.status,
        notes: o.notes,
        createdAt: new Date(o.createdAt),
      })));
    }).catch(() => {});
  }, []);

  const filtered =
    statusFilter === 'all' ? orders : orders.filter((o) => o.status === statusFilter);

  // Today's stats
  const todayOrders = orders.filter((o) => {
    const d = new Date(o.createdAt);
    const now = new Date();
    return (
      d.getFullYear() === now.getFullYear() &&
      d.getMonth() === now.getMonth() &&
      d.getDate() === now.getDate()
    );
  });

  const totalRevenue = todayOrders.filter((o) => o.status === 'paid').reduce((s, o) => s + o.total, 0);
  const cashTotal = todayOrders.filter((o) => o.status === 'paid' && o.paymentMethod === 'cash').reduce((s, o) => s + o.total, 0);
  const cardTotal = todayOrders.filter((o) => o.status === 'paid' && o.paymentMethod === 'card').reduce((s, o) => s + o.total, 0);
  const onlineTotal = todayOrders.filter((o) => o.status === 'paid' && o.paymentMethod === 'online').reduce((s, o) => s + o.total, 0);

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Summary stats */}
      <div
        className="shrink-0 px-6 py-4"
        style={{ borderBottom: '1px solid var(--hub-border)', background: 'var(--hub-surface)' }}
      >
        <div className="grid grid-cols-4 gap-3">
          {[
            { label: "Today's Revenue", value: `$${totalRevenue.toFixed(2)}`, icon: TrendingUp, color: '#10b981' },
            { label: 'Orders Today',    value: String(todayOrders.length),       icon: ShoppingCart, color: '#4583C4' },
            { label: 'Cash Sales',      value: `$${cashTotal.toFixed(2)}`,       icon: Banknote,     color: '#f59e0b' },
            { label: 'Card + Online',   value: `$${(cardTotal + onlineTotal).toFixed(2)}`, icon: CreditCard, color: '#a78bfa' },
          ].map(({ label, value, icon: Icon, color }) => (
            <motion.div
              key={label}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.2 }}
              className="flex items-center gap-3 px-4 py-3 rounded-xl"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <span
                className="flex items-center justify-center h-9 w-9 rounded-lg shrink-0"
                style={{ background: color + '18' }}
              >
                <Icon size={16} style={{ color }} />
              </span>
              <div>
                <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{label}</p>
                <p className="text-lg font-bold tabular-nums" style={{ color: 'var(--hub-text)' }}>{value}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>

      {/* Filters + table */}
      <div className="flex-1 overflow-hidden flex flex-col">
        {/* Filter bar */}
        <div
          className="flex items-center gap-2 px-6 py-3 shrink-0"
          style={{ borderBottom: '1px solid var(--hub-border)' }}
        >
          <ClipboardList size={14} style={{ color: 'var(--hub-text-muted)' }} />
          <span className="text-xs font-semibold mr-2" style={{ color: 'var(--hub-text-muted)' }}>
            Filter by status:
          </span>
          {(['all', 'paid', 'open', 'voided'] as StatusFilter[]).map((f) => (
            <button
              key={f}
              onClick={() => setStatusFilter(f)}
              className="px-3 py-1 rounded-lg text-xs font-semibold capitalize transition-all"
              style={{
                background: statusFilter === f ? '#4583C4' : 'var(--hub-card)',
                color: statusFilter === f ? '#fff' : 'var(--hub-text-muted)',
                border: `1px solid ${statusFilter === f ? '#4583C4' : 'var(--hub-border)'}`,
              }}
            >
              {f === 'all' ? 'All' : f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
          <span className="ml-auto text-xs" style={{ color: 'var(--hub-text-dim)' }}>
            {filtered.length} order{filtered.length !== 1 ? 's' : ''}
          </span>
        </div>

        {/* Table */}
        <div className="flex-1 overflow-y-auto px-6 py-4">
          {filtered.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-40 gap-2">
              <ClipboardList size={24} style={{ color: 'var(--hub-text-dim)' }} />
              <p className="text-sm" style={{ color: 'var(--hub-text-dim)' }}>No orders found</p>
            </div>
          ) : (
            <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--hub-border)' }}>
              {/* Table header */}
              <div
                className="grid text-[11px] font-semibold uppercase tracking-wider px-4 py-2.5"
                style={{
                  background: 'var(--hub-surface)',
                  color: 'var(--hub-text-dim)',
                  gridTemplateColumns: '80px 100px 1fr 90px 90px 110px 80px',
                  borderBottom: '1px solid var(--hub-border)',
                }}
              >
                <span>Order #</span>
                <span>Table</span>
                <span>Items</span>
                <span>Total</span>
                <span>Payment</span>
                <span>Time</span>
                <span>Status</span>
              </div>

              {/* Rows */}
              {filtered.map((order, idx) => {
                const status = STATUS_META[order.status];
                const StatusIcon = status.icon;
                const payment = PAYMENT_META[order.paymentMethod];
                const PayIcon = payment.icon;

                return (
                  <motion.div
                    key={order.id}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: idx * 0.03 }}
                    className="grid items-center px-4 py-3 text-xs"
                    style={{
                      gridTemplateColumns: '80px 100px 1fr 90px 90px 110px 80px',
                      borderBottom: idx < filtered.length - 1 ? '1px solid var(--hub-border)' : 'none',
                      background: idx % 2 === 0 ? 'transparent' : 'var(--hub-hover)',
                    }}
                  >
                    <span className="font-mono font-semibold" style={{ color: '#4583C4' }}>
                      {order.id}
                    </span>
                    <span style={{ color: 'var(--hub-text-muted)' }}>{order.tableLabel}</span>
                    <span style={{ color: 'var(--hub-text)' }} className="truncate pr-4">
                      {order.items.map((i) => `${i.name} ×${i.qty}`).join(', ')}
                    </span>
                    <span className="font-bold tabular-nums" style={{ color: 'var(--hub-text)' }}>
                      ${order.total.toFixed(2)}
                    </span>
                    <span className="flex items-center gap-1" style={{ color: 'var(--hub-text-muted)' }}>
                      <PayIcon size={12} />
                      {payment.label}
                    </span>
                    <div>
                      <p style={{ color: 'var(--hub-text-muted)' }}>{formatTime(new Date(order.createdAt))}</p>
                      <p style={{ color: 'var(--hub-text-dim)', fontSize: '10px' }}>
                        {formatRelative(new Date(order.createdAt))}
                      </p>
                    </div>
                    <span
                      className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold"
                      style={{ background: status.bg, color: status.color }}
                    >
                      <StatusIcon size={10} />
                      {status.label}
                    </span>
                  </motion.div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Orders;
