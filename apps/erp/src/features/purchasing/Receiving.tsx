import React, { useState } from 'react';
import { motion } from 'motion/react';
import { PackageCheck, Check, X as XIcon, Plus, Minus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { usePurchaseOrders, useUpdatePurchaseOrderStatus } from './usePurchaseOrders';
import type { PurchaseOrder } from './usePurchaseOrders';
import { useSuppliers } from '../suppliers/useSuppliers';
import PageHeader from '../../shared/components/PageHeader';
import EmptyState from '../../shared/components/EmptyState';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatDate } from '../../shared/lib/format';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

interface ReceiveLine {
  sku: string;
  name: string;
  expected: number;
  received: number;
}

const Receiving: React.FC = () => {
  const { data: ordersResult, isLoading } = usePurchaseOrders();
  const { data: supplierResult } = useSuppliers();
  const orders = ordersResult?.purchaseOrders ?? [];
  const supplierNameById = new Map((supplierResult?.suppliers ?? []).map((s) => [s.id, s.name]));
  const updateStatus = useUpdatePurchaseOrderStatus();
  const open = orders.filter((o) => o.status === 'sent' || o.status === 'approved');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [lines, setLines] = useState<ReceiveLine[]>([]);

  const handleSelectOrder = (order: PurchaseOrder): void => {
    setSelectedId(order.id);
    setLines(
      order.items.map((item) => ({
        sku: item.productId,
        name: `Product #${item.productId}`,
        expected: item.quantityOrdered,
        received: item.quantityReceived || item.quantityOrdered,
      }))
    );
  };

  const adjust = (sku: string, delta: number) => {
    setLines((prev) =>
      prev.map((l) =>
        l.sku === sku
          ? { ...l, received: Math.min(l.expected, Math.max(0, l.received + delta)) }
          : l
      )
    );
  };

  const completeReceipt = () => {
    const overReceived = lines.filter((l) => l.received > l.expected);
    if (overReceived.length > 0) {
      toast.error('Cannot receive more units than ordered');
      return;
    }
    const variance = lines.reduce((sum, l) => sum + (l.expected - l.received), 0);
    toast.success(
      variance === 0
        ? 'Receipt completed, all items received'
        : `Receipt completed with ${variance} unit variance`
    );
    if (selectedId !== null) {
      // Backend expects the lowercase status enum (draft|sent|approved|partial|received|cancelled).
      updateStatus.mutate(
        { id: selectedId, body: { status: 'received' } },
        { onError: () => toast.error('Failed to update receipt status') }
      );
    }
    setSelectedId(null);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Goods receiving · rtylr" />
      <PageHeader
        title="Goods Receiving"
        description="Receive items against open purchase orders"
      />

      {isLoading ? (
        <Skeleton className="h-40 w-full rounded-xl" />
      ) : selectedId ? (
        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Receive items for PO-{selectedId}
            </h3>
            <button onClick={() => setSelectedId(null)}>
              <XIcon size={14} style={{ color: 'var(--hub-text-muted)' }} />
            </button>
          </div>
          <div className="space-y-2 mb-4">
            {lines.map((line) => {
              const variance = line.received - line.expected;
              return (
                <div
                  key={line.sku}
                  className="flex items-center gap-3 px-3 py-2 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <div className="flex-1">
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {line.name}
                    </p>
                    <p className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                      {line.sku} / Expected {line.expected}
                    </p>
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => adjust(line.sku, -1)}
                      className="h-6 w-6 rounded flex items-center justify-center"
                      style={{ background: 'var(--hub-card)', color: 'var(--hub-text-muted)' }}
                    >
                      <Minus size={12} />
                    </button>
                    <span
                      className="text-sm font-semibold min-w-[40px] text-center"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      {line.received}
                    </span>
                    <button
                      onClick={() => adjust(line.sku, 1)}
                      className="h-6 w-6 rounded flex items-center justify-center"
                      style={{ background: 'var(--hub-card)', color: 'var(--hub-text-muted)' }}
                    >
                      <Plus size={12} />
                    </button>
                  </div>
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full min-w-[48px] text-center"
                    style={{
                      background: variance === 0 ? 'rgba(16,185,129,0.1)' : 'rgba(248,113,113,0.1)',
                      color: variance === 0 ? '#10b981' : '#f87171',
                    }}
                  >
                    {variance === 0 ? 'OK' : `${variance > 0 ? '+' : ''}${variance}`}
                  </span>
                </div>
              );
            })}
          </div>
          <button
            onClick={completeReceipt}
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Check size={14} /> Mark Received
          </button>
        </div>
      ) : open.length === 0 ? (
        <EmptyState
          icon={PackageCheck}
          title="No open orders to receive"
          description="All orders have been received"
        />
      ) : (
        <div className="space-y-2">
          {open.map((o) => (
            <button
              key={o.id}
              onClick={() => handleSelectOrder(o)}
              className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-left transition-colors"
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor = 'rgba(245,158,11,0.4)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
              }}
            >
              <div
                className="h-9 w-9 rounded-lg flex items-center justify-center"
                style={{ background: 'rgba(245,158,11,0.12)' }}
              >
                <PackageCheck size={14} style={{ color: AMBER }} />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  PO-{o.id} / {supplierNameById.get(o.supplierId) ?? `Supplier #${o.supplierId}`}
                </p>
                <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                  {o.items.length} items / Expected {formatDate(o.expectedDelivery ?? '')}
                </p>
              </div>
              <span
                className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                style={{ background: 'rgba(245,158,11,0.1)', color: AMBER }}
              >
                {o.status}
              </span>
            </button>
          ))}
        </div>
      )}
    </motion.div>
  );
};

export default Receiving;
