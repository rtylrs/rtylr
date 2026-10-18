import React, { useCallback, useMemo, useState } from 'react';
import { motion, type Variants } from 'motion/react';
import {
  Plus,
  FileText,
  Send,
  CheckCircle,
  Clock as ClockIcon,
  ThumbsUp,
  Trash2,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../../../shared/stores/userStore';
import {
  type PurchaseOrder,
  usePurchaseOrders,
  useApprovePurchaseOrder,
  useUpdatePurchaseOrderStatus,
  useCreatePurchaseOrder,
  fetchAllPurchaseOrders,
  useAllPurchaseOrders,
} from './usePurchaseOrders';
import * as purchaseOrderQueries from './usePurchaseOrders';
import {
  displayPurchaseOrderStatus,
  getPurchaseOrderStatusCounts,
  purchaseOrderLabel,
  purchaseOrderStatusFilter,
  type PurchaseOrderFilter,
  type PurchaseOrderStatusCard,
} from './purchaseOrderStatusUtils';
import { useSuppliers } from '../suppliers/useSuppliers';
import { useAllProducts } from '../products/useProducts';
import PrintPOButton from '../../shared/components/PrintPOButton';
import ExportButton from '../../shared/components/ExportButton';
import PaginationControls from '../../shared/components/PaginationControls';
import { canErpPermission } from '../../shared/lib/erpPermissions';
import { toast } from 'sonner';
import { apiErrorMessage } from '../../shared/lib/apiErrorMessage';
import { toIsoDate } from '../../shared/lib/dateUtils';
import { formatCurrency, formatDate } from '../../shared/lib/format';
import { useErpCurrency } from '../../shared/hooks/useErpCurrency';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';

const ACCENT = 'var(--erp-accent)';

type DisplayStatus = ReturnType<typeof displayPurchaseOrderStatus>;

const statusConfig: Record<
  DisplayStatus,
  {
    color: string;
    bg: string;
    icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  }
> = {
  Draft: { color: 'var(--hub-text-muted)', bg: 'var(--hub-hover)', icon: FileText },
  Partial: {
    color: 'var(--erp-purple)',
    bg: 'color-mix(in srgb, var(--erp-purple) 12%, transparent)',
    icon: ClockIcon,
  },
  Approved: {
    color: 'var(--erp-info)',
    bg: 'color-mix(in srgb, var(--erp-info) 12%, transparent)',
    icon: ThumbsUp,
  },
  Sent: {
    color: ACCENT,
    bg: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)',
    icon: Send,
  },
  Received: {
    color: 'var(--erp-success)',
    bg: 'color-mix(in srgb, var(--erp-success) 10%, transparent)',
    icon: CheckCircle,
  },
  Cancelled: {
    color: 'var(--erp-danger)',
    bg: 'color-mix(in srgb, var(--erp-danger) 12%, transparent)',
    icon: Trash2,
  },
};

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const columns = ['order', 'supplier', 'expected', 'items', 'total', 'status', 'actions'] as const;

export function purchaseOrderStatusLabel(t: TFunction, status: DisplayStatus): string {
  return t(`erp.purchasing.status.${status.toLowerCase()}`);
}

const SkeletonTableRow: React.FC = () => (
  <div
    className="grid px-4 py-3 items-center gap-2"
    style={{ gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr' }}
  >
    {Array.from({ length: 7 }).map((_, i) => (
      <div
        key={i}
        className="h-3 rounded animate-pulse"
        style={{ background: 'var(--hub-border)', opacity: 0.6 }}
      />
    ))}
  </div>
);

const emptyPurchaseOrderDetail = {
  data: null,
  isLoading: false,
  isError: false,
} as ReturnType<typeof purchaseOrderQueries.usePurchaseOrder>;

function usePurchaseOrderIfAvailable(
  id: string | undefined
): ReturnType<typeof purchaseOrderQueries.usePurchaseOrder> {
  // List-only consumers may provide the list hook without the detail hook.
  const detailHook = Object.prototype.hasOwnProperty.call(purchaseOrderQueries, 'usePurchaseOrder')
    ? purchaseOrderQueries.usePurchaseOrder
    : undefined;
  return detailHook ? detailHook(id) : emptyPurchaseOrderDetail;
}

const PurchaseOrderItemCount: React.FC<{ order: PurchaseOrder }> = ({ order }) => {
  const { t } = useTranslation();
  const detailQuery = usePurchaseOrderIfAvailable(order.id ?? undefined);
  const detailItems = detailQuery.data?.items;
  const listItems = order.items;

  if (detailQuery.isLoading && !listItems?.length) {
    return <span style={{ color: 'var(--hub-text-dim)' }}>…</span>;
  }
  if (detailQuery.isError || (!detailQuery.data && !listItems?.length)) {
    return (
      <span title={t('erp.purchasing.itemsUnavailable')} style={{ color: 'var(--hub-text-dim)' }}>
        —
      </span>
    );
  }

  return <span>{(detailItems ?? listItems ?? []).length}</span>;
};

const PurchaseOrderPrintAction: React.FC<{
  order: PurchaseOrder;
  supplierName: string;
  productNameById: ReadonlyMap<string, string>;
  currency: string;
}> = ({ order, supplierName, productNameById, currency }) => {
  const { t } = useTranslation();
  const detailQuery = usePurchaseOrderIfAvailable(order.id ?? undefined);
  const printableOrder = detailQuery.data ?? (order.items?.length ? order : null);

  if (detailQuery.isLoading && !printableOrder) {
    return (
      <span className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
        {t('erpCommon.loading')}
      </span>
    );
  }
  if (!printableOrder) {
    return (
      <span
        className="text-[10px]"
        title={t('erp.purchasing.print.unavailableHint')}
        style={{ color: 'var(--hub-text-dim)' }}
      >
        {t('erp.purchasing.print.unavailable')}
      </span>
    );
  }

  return (
    <PrintPOButton
      order={printableOrder}
      supplierName={supplierName}
      productNameById={productNameById}
      currency={currency}
    />
  );
};

const Purchasing: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const [filter, setFilter] = useState<PurchaseOrderFilter>('All');
  const status = purchaseOrderStatusFilter(filter);
  const { data: ordersResult, isLoading } = usePurchaseOrders({ page, status });
  const { data: allOrders = [], isLoading: isLoadingAllOrders } = useAllPurchaseOrders();
  const orders = ordersResult?.purchaseOrders ?? [];
  const hasMore = ordersResult?.hasMore ?? false;
  const approveMut = useApprovePurchaseOrder();
  const updateStatusMut = useUpdatePurchaseOrderStatus();
  const createMut = useCreatePurchaseOrder();
  const { data: supplierResult } = useSuppliers();
  const suppliers = supplierResult?.suppliers ?? [];
  // The picker must offer the whole catalog, not the first page of it.
  const { data: products = [] } = useAllProducts();
  const productNameById = useMemo(
    () => new Map(products.map((product) => [product.id ?? '', product.name ?? ''])),
    [products]
  );
  const currency = useErpCurrency();
  const canApprovePO = canErpPermission('erp:purchasing:approve');
  const [showCreate, setShowCreate] = useState(false);
  const [newSupplierId, setNewSupplierId] = useState('');
  const [newPoNumber, setNewPoNumber] = useState('');
  const [newExpectedDelivery, setNewExpectedDelivery] = useState(
    toIsoDate(new Date(Date.now() + 7 * 86400000))
  );
  const [newNotes, setNewNotes] = useState('');
  const [newItems, setNewItems] = useState<
    Array<{ productId: string; quantityOrdered: string; unitCost: string }>
  >([{ productId: '', quantityOrdered: '', unitCost: '' }]);

  const supplierNameById = useMemo(
    () =>
      new Map((supplierResult?.suppliers ?? []).map((supplier) => [supplier.id, supplier.name])),
    [supplierResult?.suppliers]
  );
  const supplierName = useCallback(
    (supplierId: string | undefined) =>
      supplierNameById.get(supplierId) ??
      t('erp.purchasing.supplierFallback', { id: supplierId ?? '' }),
    [supplierNameById, t]
  );

  const filteredOrders = orders;

  const toExportRows = useCallback(
    (purchaseOrders: PurchaseOrder[]) =>
      purchaseOrders.map((o) => ({
        id: o.id,
        supplier: supplierName(o.supplierId),
        createdAt: o.createdAt,
        expectedDelivery: o.expectedDelivery ?? '',
        // List rows intentionally omit line items; leave the count unknown
        // rather than exporting a misleading zero.
        items: o.items?.length ? o.items.length : '',
        total: o.total,
        status: displayPurchaseOrderStatus(o.status ?? ''),
      })),
    [supplierName]
  );
  const exportable = useMemo(() => toExportRows(filteredOrders), [filteredOrders, toExportRows]);
  const getAllExportable = useCallback(async () => {
    const allFilteredOrders = await fetchAllPurchaseOrders({ status });
    return toExportRows(allFilteredOrders);
  }, [status, toExportRows]);

  const statusCounts = getPurchaseOrderStatusCounts(allOrders);

  const handleApprove = (order: PurchaseOrder): void => {
    const label = purchaseOrderLabel(order);
    approveMut.mutate(order.id ?? '', {
      onSuccess: () => toast.success(t('erp.purchasing.toast.approved', { label })),
      onError: (error) =>
        toast.error(apiErrorMessage(error, t('erp.purchasing.toast.approveFailed', { label }))),
    });
  };

  const handleSend = (order: PurchaseOrder): void => {
    const label = purchaseOrderLabel(order);
    updateStatusMut.mutate(
      { id: order.id ?? '', body: { status: 'sent' } },
      {
        onSuccess: () => toast.success(t('erp.purchasing.toast.sent', { label })),
        onError: (error) =>
          toast.error(apiErrorMessage(error, t('erp.purchasing.toast.sendFailed', { label }))),
      }
    );
  };

  const handleCreatePO = async (): Promise<void> => {
    if (!newSupplierId || !newPoNumber.trim()) {
      toast.error(t('erp.purchasing.supplierAndPoRequired'));
      return;
    }
    const validItems = newItems.filter((it) => it.productId && Number(it.quantityOrdered) > 0);
    if (validItems.length === 0) {
      toast.error(t('erp.purchasing.itemRequired'));
      return;
    }
    try {
      await createMut.mutateAsync({
        supplierId: newSupplierId,
        poNumber: newPoNumber.trim(),
        taxAmount: 0,
        expectedDelivery: newExpectedDelivery,
        notes: newNotes.trim(),
        items: validItems.map((it) => ({
          productId: it.productId,
          quantityOrdered: Number(it.quantityOrdered),
          unitCost: it.unitCost ? Number(it.unitCost) : 0,
        })),
      });
      toast.success(t('erp.purchasing.draftCreated'));
      setShowCreate(false);
      setNewSupplierId('');
      setNewPoNumber('');
      setNewExpectedDelivery(toIsoDate(new Date(Date.now() + 7 * 86400000)));
      setNewNotes('');
      setNewItems([{ productId: '', quantityOrdered: '', unitCost: '' }]);
    } catch (error) {
      toast.error(apiErrorMessage(error, t('erp.purchasing.createFailed')));
    }
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/purchasing'].title))} />
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t(routeLabels['/erp/purchasing'].title)}
        </h2>
        <div className="flex items-center gap-2">
          <ExportButton
            data={exportable}
            getData={getAllExportable}
            filename="purchase-orders.csv"
          />
          <button
            type="button"
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
          >
            <Plus size={14} />
            {t('erp.purchasing.newOrder')}
          </button>
        </div>
      </motion.div>

      <motion.div variants={itemVariants} className="grid grid-cols-2 lg:grid-cols-5 gap-3">
        {(Object.entries(statusCounts) as [PurchaseOrderStatusCard, number][]).map(
          ([status, count]) => {
            const sc = statusConfig[status];
            if (!sc) return null;
            const Icon = sc.icon;
            const active = filter === status;
            return (
              <button
                type="button"
                key={status}
                onClick={() => {
                  setFilter(active ? 'All' : status);
                  setPage(1);
                }}
                className="flex items-center gap-3 px-4 py-3 rounded-xl transition-all text-left"
                style={{
                  background: 'var(--hub-card)',
                  border: active
                    ? '1px solid color-mix(in srgb, var(--erp-accent) 50%, transparent)'
                    : '1px solid var(--hub-border)',
                }}
              >
                <span
                  className="flex items-center justify-center h-8 w-8 rounded-lg shrink-0"
                  style={{ background: sc.bg }}
                >
                  <Icon size={15} style={{ color: sc.color }} />
                </span>
                <div>
                  <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>
                    {isLoadingAllOrders ? '…' : count}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {purchaseOrderStatusLabel(t, status)}
                  </p>
                </div>
              </button>
            );
          }
        )}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3"
          style={{
            gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          {columns.map((col) => (
            <span
              key={col}
              className="text-[10px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--hub-text-dim)' }}
            >
              {t(`erp.purchasing.columns.${col}`)}
            </span>
          ))}
        </div>

        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading
            ? Array.from({ length: 6 }).map((_, i) => <SkeletonTableRow key={i} />)
            : filteredOrders.map((order) => {
                const status = displayPurchaseOrderStatus(order.status ?? '');
                const sc = statusConfig[status] || statusConfig['Draft'];
                return (
                  <div
                    key={order.id}
                    className="grid px-4 py-3 items-center transition-colors"
                    style={{ gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <span
                      className="text-xs font-mono font-medium"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      {purchaseOrderLabel(order)}
                    </span>
                    <span
                      className="text-sm truncate pr-2"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      {supplierName(order.supplierId)}
                    </span>
                    <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                      {order.expectedDelivery ? formatDate(order.expectedDelivery) : ''}
                    </span>
                    <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      <PurchaseOrderItemCount order={order} />
                    </span>
                    <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                      {formatCurrency(order.total ?? 0, currency)}
                    </span>
                    <span>
                      <span
                        className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                        style={{ background: sc.bg, color: sc.color }}
                      >
                        {purchaseOrderStatusLabel(t, status)}
                      </span>
                    </span>
                    <div className="flex items-center gap-1">
                      {status === 'Sent' && canApprovePO && (
                        <button
                          type="button"
                          onClick={() => handleApprove(order)}
                          className="text-[10px] font-semibold px-2 py-1 rounded-md"
                          style={{
                            background: 'color-mix(in srgb, var(--erp-info) 15%, transparent)',
                            color: 'var(--erp-info)',
                          }}
                        >
                          {t('erp.purchasing.approve')}
                        </button>
                      )}
                      {status === 'Draft' && (
                        <button
                          type="button"
                          onClick={() => handleSend(order)}
                          className="text-[10px] font-semibold px-2 py-1 rounded-md"
                          style={{
                            background: 'color-mix(in srgb, var(--erp-accent) 15%, transparent)',
                            color: ACCENT,
                          }}
                        >
                          {t('erp.purchasing.send')}
                        </button>
                      )}
                      <PurchaseOrderPrintAction
                        order={order}
                        supplierName={supplierName(order.supplierId)}
                        productNameById={productNameById}
                        currency={currency}
                      />
                    </div>
                  </div>
                );
              })}
        </div>
      </motion.div>

      {showCreate && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ background: 'var(--hub-overlay)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.purchasing.newPurchaseOrder')}
            </h3>
            <select
              value={newSupplierId}
              onChange={(e) => setNewSupplierId(e.target.value)}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            >
              <option value="">{t('erp.purchasing.form.selectSupplier')}</option>
              {suppliers.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
            <input
              value={newPoNumber}
              onChange={(e) => setNewPoNumber(e.target.value)}
              placeholder={t('erp.purchasing.form.poNumber')}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <input
              type="date"
              value={newExpectedDelivery}
              onChange={(e) => setNewExpectedDelivery(e.target.value)}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <input
              value={newNotes}
              onChange={(e) => setNewNotes(e.target.value)}
              placeholder={t('erpCommon.notesOptional')}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <p
              className="text-[11px] font-semibold uppercase tracking-wide"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              {t('erp.purchasing.form.items')}
            </p>
            {newItems.map((item, idx) => (
              <div key={idx} className="flex gap-1 items-center">
                <select
                  value={item.productId}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) => (i === idx ? { ...it, productId: e.target.value } : it))
                    )
                  }
                  className="flex-1 px-2 py-1.5 rounded-lg text-xs outline-none"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                >
                  <option value="">{t('erpCommon.product')}</option>
                  {products.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min="1"
                  value={item.quantityOrdered}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) =>
                        i === idx ? { ...it, quantityOrdered: e.target.value } : it
                      )
                    )
                  }
                  placeholder={t('erpCommon.qtyShort')}
                  className="w-14 px-2 py-1.5 rounded-lg text-xs outline-none"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                />
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={item.unitCost}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) => (i === idx ? { ...it, unitCost: e.target.value } : it))
                    )
                  }
                  placeholder={t('erp.purchasing.cost')}
                  className="w-16 px-2 py-1.5 rounded-lg text-xs outline-none"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                />
                <button
                  type="button"
                  onClick={() => setNewItems((prev) => prev.filter((_, i) => i !== idx))}
                  disabled={newItems.length === 1}
                  aria-label={t('erp.purchasing.form.removeItem')}
                  className="p-1 rounded disabled:opacity-30"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  <Trash2 size={12} />
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() =>
                setNewItems((prev) => [
                  ...prev,
                  { productId: '', quantityOrdered: '', unitCost: '' },
                ])
              }
              className="text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              {t('erp.purchasing.form.addItem')}
            </button>
            <div className="flex gap-2 justify-end">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                {t('erpCommon.cancel')}
              </button>
              <button
                type="button"
                onClick={handleCreatePO}
                disabled={createMut.isPending || !newSupplierId || !newPoNumber.trim()}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
              >
                {createMut.isPending
                  ? t('erp.common.creating')
                  : t('erp.purchasing.form.createDraft')}
              </button>
            </div>
          </div>
        </div>
      )}

      <PaginationControls
        page={page}
        hasMore={hasMore}
        itemCount={orders.length}
        onPageChange={setPage}
      />
    </motion.div>
  );
};

export default Purchasing;
