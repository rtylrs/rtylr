import React, { useEffect, useMemo, useRef, useState } from 'react';
import { motion } from 'motion/react';
import { useNavigate } from '@tanstack/react-router';
import { PackageCheck, Check, X as XIcon, Plus, Minus } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import {
  usePurchaseOrder,
  usePurchaseOrders,
  useUpdatePurchaseOrderStatus,
} from './usePurchaseOrders';
import type { PurchaseOrder } from './usePurchaseOrders';
import { useSuppliers } from '../suppliers/useSuppliers';
import { useAllProducts } from '../products/useProducts';
import { displayPurchaseOrderStatus, purchaseOrderLabel } from './purchaseOrderStatusUtils';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatDate } from '../../shared/lib/format';
import { toast } from 'sonner';
import { apiErrorMessage } from '../../shared/lib/apiErrorMessage';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';

const ACCENT = 'var(--erp-accent)';

interface ReceiveLine {
  id?: string;
  productId: string;
  expected: number;
  received: number;
  // Already booked into stock; the server refuses to lower a line below it.
  previouslyReceived: number;
}

const Receiving: React.FC = () => {
  const navigate = useNavigate();
  const { data: ordersResult, isLoading } = usePurchaseOrders();
  const { data: supplierResult } = useSuppliers();
  const orders = ordersResult?.purchaseOrders ?? [];
  const supplierNameById = new Map((supplierResult?.suppliers ?? []).map((s) => [s.id, s.name]));
  const updateStatus = useUpdatePurchaseOrderStatus();
  const { t } = useTranslation('erp');
  const { data: products = [] } = useAllProducts();
  const productById = useMemo(
    () => new Map(products.map((product) => [product.id ?? '', product])),
    [products]
  );
  const open = orders.filter(
    (o) => o.status === 'sent' || o.status === 'approved' || o.status === 'partial'
  );
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [lines, setLines] = useState<ReceiveLine[]>([]);
  // The list payload has no line items, so the panel is seeded from the detail
  // fetch. Seed once per order so a background refetch cannot clobber edits.
  const seededOrderId = useRef<string | null>(null);
  const {
    data: orderDetail,
    isLoading: isDetailLoading,
    isError: isDetailError,
  } = usePurchaseOrder(selectedId ?? undefined);

  useEffect(() => {
    if (!selectedId || seededOrderId.current === selectedId) return;
    if (!orderDetail || orderDetail.id !== selectedId) return;
    seededOrderId.current = selectedId;
    setLines(
      (orderDetail.items ?? []).map((item) => ({
        id: item.id,
        productId: item.productId ?? '',
        expected: item.quantityOrdered ?? 0,
        received: item.quantityReceived ?? item.quantityOrdered ?? 0,
        previouslyReceived: item.quantityReceived ?? 0,
      }))
    );
  }, [selectedId, orderDetail]);

  // Names are looked up at render time so they appear once the catalog loads.
  const productName = (productId: string): string =>
    productById.get(productId)?.name || t('erp.purchasing.productFallback', { id: productId });
  const productSku = (productId: string): string => productById.get(productId)?.sku || productId;
  const selectedOrder = open.find((order) => order.id === selectedId) ?? orderDetail ?? undefined;

  const closeDetails = (): void => {
    seededOrderId.current = null;
    setLines([]);
    setSelectedId(null);
  };

  const handleSelectOrder = (order: PurchaseOrder): void => {
    seededOrderId.current = null;
    setLines([]);
    setSelectedId(order.id ?? null);
  };

  const adjust = (lineKey: string, delta: number) => {
    setLines((prev) =>
      prev.map((l) =>
        (l.id ?? l.productId) === lineKey
          ? {
              ...l,
              received: Math.min(l.expected, Math.max(l.previouslyReceived, l.received + delta)),
            }
          : l
      )
    );
  };

  const completeReceipt = () => {
    // The list rows carry no items. Submitting items: [] would let the backend
    // default every line to its ordered quantity and receive the whole order.
    if (lines.length === 0) {
      toast.error(t('erp.receiving.validation.noItems'));
      return;
    }
    const overReceived = lines.filter((l) => l.received > l.expected);
    if (overReceived.length > 0) {
      toast.error(t('erp.receiving.validation.overReceived'));
      return;
    }
    const hasReceivedItems = lines.some((line) => line.received > 0);
    // The backend derives partial vs received from the submitted quantities,
    // and only reads them for the received status.
    const status = hasReceivedItems ? 'received' : 'approved';
    const variance = lines.reduce((sum, l) => sum + (l.expected - l.received), 0);
    if (selectedId === null) return;
    updateStatus.mutate(
      {
        id: selectedId,
        body: {
          status,
          items: lines.map((line) => ({
            id: line.id,
            quantityOrdered: line.expected,
            quantityReceived: line.received,
          })),
        },
      },
      {
        onSuccess: () => {
          toast.success(
            variance === 0
              ? t('erp.receiving.toast.completed')
              : t('erp.receiving.toast.completedVariance', { count: variance })
          );
          closeDetails();
        },
        onError: (error) =>
          toast.error(apiErrorMessage(error, t('erp.receiving.toast.updateFailed'))),
      }
    );
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
              {t('erp.receiving.receiveFor', {
                label: selectedOrder ? purchaseOrderLabel(selectedOrder) : `PO-${selectedId}`,
              })}
            </h3>
            <button
              type="button"
              onClick={closeDetails}
              aria-label={t('erp.receiving.closeDetails')}
              className="h-11 w-11 rounded-lg flex items-center justify-center"
            >
              <XIcon size={14} style={{ color: 'var(--hub-text-muted)' }} />
            </button>
          </div>
          <div className="space-y-2 mb-4">
            {isDetailLoading ? (
              <Skeleton className="h-14 w-full rounded-lg" />
            ) : isDetailError || !orderDetail ? (
              <p className="text-xs" style={{ color: 'var(--erp-danger)' }}>
                {t('erp.receiving.error.loadFailed')}
              </p>
            ) : lines.length === 0 ? (
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                {t('erp.receiving.validation.noItems')}
              </p>
            ) : (
              lines.map((line) => {
                const lineKey = line.id ?? line.productId;
                const name = productName(line.productId);
                const variance = line.received - line.expected;
                return (
                  <div
                    key={lineKey}
                    className="flex items-center gap-3 px-3 py-2 rounded-lg"
                    style={{ background: 'var(--hub-hover)' }}
                  >
                    <div className="flex-1">
                      <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                        {name}
                      </p>
                      <p className="text-[10px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
                        {productSku(line.productId)} /{' '}
                        {t('erp.receiving.expected', { count: line.expected })}
                      </p>
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        type="button"
                        onClick={() => adjust(lineKey, -1)}
                        disabled={line.received <= line.previouslyReceived}
                        aria-label={t('erp.receiving.decrease', { name })}
                        className="h-11 w-11 rounded flex items-center justify-center disabled:opacity-40"
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
                        onClick={() => adjust(lineKey, 1)}
                        disabled={line.received >= line.expected}
                        aria-label={t('erp.receiving.increase', { name })}
                        className="h-11 w-11 rounded flex items-center justify-center disabled:opacity-40"
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
              })
            )}
          </div>
          <button
            type="button"
            onClick={completeReceipt}
            disabled={lines.length === 0 || updateStatus.isPending}
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
            style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
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
            onClick: () => navigate({ to: '/erp/purchasing' }),
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
                <PackageCheck size={14} style={{ color: ACCENT }} />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  {purchaseOrderLabel(o)} /{' '}
                  {supplierNameById.get(o.supplierId) ??
                    t('erp.purchasing.supplierFallback', { id: o.supplierId ?? '' })}
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
                  color: ACCENT,
                }}
              >
                {t(
                  `erp.purchasing.status.${displayPurchaseOrderStatus(o.status ?? '').toLowerCase()}`
                )}
              </span>
            </button>
          ))}
        </div>
      )}
    </motion.div>
  );
};

export default Receiving;
