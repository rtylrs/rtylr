import React, { useState } from 'react';
import { motion } from 'motion/react';
import { useNavigate } from '@tanstack/react-router';
import { PackageCheck, Check, X as XIcon, Plus, Minus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { usePurchaseOrders, useUpdatePurchaseOrderStatus } from './usePurchaseOrders';
import type { PurchaseOrder } from './usePurchaseOrders';
import { useSuppliers } from '../suppliers/useSuppliers';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatDate } from '../../shared/lib/format';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';

const AMBER = 'var(--erp-accent)';

interface ReceiveLine {
  id?: string;
  sku: string;
  name: string;
  expected: number;
  received: number;
}

const Receiving: React.FC = () => {
  const navigate = useNavigate();
  const { data: ordersResult, isLoading } = usePurchaseOrders();
  const { data: supplierResult } = useSuppliers();
  const orders = ordersResult?.purchaseOrders ?? [];
  const supplierNameById = new Map((supplierResult?.suppliers ?? []).map((s) => [s.id, s.name]));
  const updateStatus = useUpdatePurchaseOrderStatus();
  const { t } = useTranslation('erp');
  const open = orders.filter(
    (o) => o.status === 'sent' || o.status === 'approved' || o.status === 'partial'
  );
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [lines, setLines] = useState<ReceiveLine[]>([]);

  const handleSelectOrder = (order: PurchaseOrder): void => {
    setSelectedId(order.id ?? null);
    setLines(
      (order.items ?? []).map((item) => ({
        id: item.id,
        sku: item.productId ?? '',
        name: `Product #${item.productId ?? ''}`,
        expected: item.quantityOrdered ?? 0,
        received: item.quantityReceived ?? item.quantityOrdered ?? 0,
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
      toast.error(t('erp.receiving.validation.overReceived'));
      return;
    }
    const hasReceivedItems = lines.some((line) => line.received > 0);
    const isFullyReceived = lines.every((line) => line.received === line.expected);
    const status = isFullyReceived ? 'received' : hasReceivedItems ? 'partial' : 'approved';
    const variance = lines.reduce((sum, l) => sum + (l.expected - l.received), 0);
    toast.success(
      variance === 0
        ? t('erp.receiving.toast.completed')
        : t('erp.receiving.toast.completedVariance', { count: variance })
    );
    if (selectedId !== null) {
      updateStatus.mutate(
        {
          id: selectedId,
          body: {
            status,
            items: lines.map((line) => ({
              id: line.id,
              quantityReceived: line.received,
            })),
          },
        },
        { onError: () => toast.error(t('erp.receiving.toast.updateFailed')) }
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
      <HelmetContainer title={pageTitle(t('erp.receiving.title'))} />
      <PageHeader title={t('erp.receiving.title')} description={t('erp.receiving.description')} />

      {isLoading ? (
        <Skeleton className="h-40 w-full rounded-xl" />
      ) : selectedId ? (
        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.receiving.receiveFor', { id: selectedId })}
            </h3>
            <button
              type="button"
              onClick={() => setSelectedId(null)}
              aria-label={t('erp.receiving.closeDetails')}
              className="h-11 w-11 rounded-lg flex items-center justify-center"
            >
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
                      {line.sku} / {t('erp.receiving.expected', { count: line.expected })}
                    </p>
                  </div>
                  <div className="flex items-center gap-1">
                    <button
                      type="button"
                      onClick={() => adjust(line.sku, -1)}
                      aria-label={`Decrease received quantity for ${line.name}`}
                      className="h-11 w-11 rounded flex items-center justify-center"
                      style={{ background: 'var(--hub-card)', color: 'var(--hub-text-muted)' }}
                    >
                      <Minus size={16} />
                    </button>
                    <span
                      className="text-sm font-semibold min-w-[40px] text-center"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      {line.received}
                    </span>
                    <button
                      type="button"
                      onClick={() => adjust(line.sku, 1)}
                      aria-label={`Increase received quantity for ${line.name}`}
                      className="h-11 w-11 rounded flex items-center justify-center"
                      style={{ background: 'var(--hub-card)', color: 'var(--hub-text-muted)' }}
                    >
                      <Plus size={16} />
                    </button>
                  </div>
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full min-w-[48px] text-center"
                    style={{
                      background:
                        variance === 0
                          ? 'color-mix(in srgb, var(--erp-success) 10%, transparent)'
                          : 'color-mix(in srgb, var(--erp-danger) 10%, transparent)',
                      color: variance === 0 ? 'var(--erp-success)' : 'var(--erp-danger)',
                    }}
                  >
                    {variance === 0
                      ? t('erp.receiving.ok')
                      : `${variance > 0 ? '+' : ''}${variance}`}
                  </span>
                </div>
              );
            })}
          </div>
          <button
            type="button"
            onClick={completeReceipt}
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
          >
            <Check size={14} /> {t('erp.receiving.markReceived')}
          </button>
        </div>
      ) : open.length === 0 ? (
        <EmptyState
          icon={PackageCheck}
          title={t('erp.receiving.empty.title')}
          description={t('erp.receiving.empty.description')}
          action={{
            label: t('erp.receiving.goToPurchasing'),
            onClick: () => navigate({ to: '/purchasing' }),
          }}
        />
      ) : (
        <div className="space-y-2">
          {open.map((o) => (
            <button
              type="button"
              key={o.id}
              onClick={() => handleSelectOrder(o)}
              className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-left transition-colors"
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor =
                  'color-mix(in srgb, var(--erp-accent) 40%, transparent)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
              }}
            >
              <div
                className="h-9 w-9 rounded-lg flex items-center justify-center"
                style={{ background: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)' }}
              >
                <PackageCheck size={14} style={{ color: AMBER }} />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  PO-{o.id} / {supplierNameById.get(o.supplierId) ?? `Supplier #${o.supplierId}`}
                </p>
                <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                  {t('erp.receiving.orderSummary', {
                    count: (o.items ?? []).length,
                    date: formatDate(o.expectedDelivery ?? ''),
                  })}
                </p>
              </div>
              <span
                className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                style={{
                  background: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)',
                  color: AMBER,
                }}
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
