import React, { useState } from 'react';
import { motion } from 'motion/react';
import {
  AlertTriangle,
  ArrowLeftRight,
  ArrowRight,
  CheckCircle2,
  Plus,
  Trash2,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import { useTranslation } from 'react-i18next';
import {
  useStockTransfers,
  useCreateStockTransfer,
  useUpdateStockTransferStatus,
} from './useStockTransfers';
import { canErpPermission } from '../../shared/lib/erpPermissions';
import {
  stockTransferStatusLabel,
  toStockTransferItems,
  type StockTransferDraftItem,
} from './stockTransferSchema';
import { useWarehouses } from '../warehouses/useWarehouses';
import { useProducts } from '../products/useProducts';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { formatDate } from '../../shared/lib/format';
import PaginationControls from '../../shared/components/PaginationControls';
import { toast } from 'sonner';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const ACCENT = 'var(--erp-accent)';
const PRODUCT_PAGE_SIZE = 100;

const inputStyle: React.CSSProperties = {
  padding: '8px 12px',
  borderRadius: '8px',
  border: '1px solid var(--hub-border)',
  background: 'var(--hub-input-bg)',
  color: 'var(--hub-text)',
  fontSize: '16px',
  outline: 'none',
  width: '100%',
  boxSizing: 'border-box',
};

const lineInputStyle: React.CSSProperties = { ...inputStyle, width: 'auto', flex: 1 };
const quantityInputStyle: React.CSSProperties = { ...inputStyle, width: '5rem' };

function statusConfig(status: string): { color: string; bg: string } {
  if (status === 'received')
    return {
      color: 'var(--erp-success)',
      bg: 'color-mix(in srgb, var(--erp-success) 12%, transparent)',
    };
  if (status === 'in_transit')
    return {
      color: ACCENT,
      bg: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
    };
  if (status === 'cancelled')
    return {
      color: 'var(--erp-danger)',
      bg: 'color-mix(in srgb, var(--erp-danger) 12%, transparent)',
    };
  return { color: 'var(--hub-text-dim)', bg: 'var(--hub-hover)' };
}

const ReceiveTransferAction: React.FC<{ transferId?: string }> = ({ transferId }) => {
  const { t } = useTranslation('erp');
  const updateMutation = useUpdateStockTransferStatus();
  if (!transferId) return null;

  return (
    <button
      type="button"
      aria-label={t('erp.stockTransfers.receiveAria', { id: transferId })}
      disabled={updateMutation.isPending}
      onClick={() =>
        updateMutation.mutate(
          { id: transferId, body: { status: 'received' } },
          {
            onSuccess: () => toast.success(t('erp.stockTransfers.received')),
            onError: () => toast.error(t('erp.stockTransfers.receiveFailed')),
          }
        )
      }
      className="flex items-center gap-1 rounded-md px-2 py-1 text-[10px] font-semibold disabled:opacity-40"
      style={{
        background: 'color-mix(in srgb, var(--erp-success) 12%, transparent)',
        color: 'var(--erp-success)',
      }}
    >
      <CheckCircle2 size={11} /> {t('erp.stockTransfers.receive')}
    </button>
  );
};

const StockTransfers: React.FC = () => {
  const { t } = useTranslation('erp');
  const [page, setPage] = useState(1);
  const pageSize = 50;
  const {
    data: transferResult,
    isLoading,
    isError: transfersError,
    refetch: refetchTransfers,
  } = useStockTransfers({ page, pageSize });
  const transfers = transferResult?.stockTransfers ?? [];
  const hasMore = transferResult?.hasMore ?? false;
  const {
    data: warehouses = [],
    isLoading: warehousesLoading,
    isError: warehousesError,
    refetch: refetchWarehouses,
  } = useWarehouses();
  const [productSearch, setProductSearch] = useState('');
  const [productPage, setProductPage] = useState(1);
  const {
    data: productResult,
    isLoading: productsLoading,
    isError: productsError,
    refetch: refetchProducts,
  } = useProducts({
    page: productPage,
    pageSize: PRODUCT_PAGE_SIZE,
    search: productSearch,
  });
  const products = productResult?.products ?? [];
  const productsHaveMore = productResult?.hasMore ?? false;
  const productsTotal = productResult?.total;
  const createMut = useCreateStockTransfer();
  const canCreateTransfer = canErpPermission('erp:inventory:transfer');
  const [showCreate, setShowCreate] = useState(false);
  const [fromWarehouseId, setFromWarehouseId] = useState('');
  const [toWarehouseId, setToWarehouseId] = useState('');
  const [notes, setNotes] = useState('');
  const [newItems, setNewItems] = useState<StockTransferDraftItem[]>([
    { productId: '', quantity: '' },
  ]);
  const validItems = toStockTransferItems(newItems);

  const warehouseName = (id: string): string =>
    warehouses.find((w) => w.id === id)?.name ?? `Warehouse #${id}`;

  const handleCreate = async (): Promise<void> => {
    if (warehousesError || productsError || warehousesLoading || productsLoading) {
      toast.error(t('erp.stockTransfers.reloadBeforeCreate'));
      return;
    }
    if (!fromWarehouseId || !toWarehouseId) {
      toast.error(t('erp.stockTransfers.warehousesRequired'));
      return;
    }
    if (fromWarehouseId === toWarehouseId) {
      toast.error(t('erp.stockTransfers.warehousesMustDiffer'));
      return;
    }
    if (validItems.length === 0) {
      toast.error(t('erp.stockTransfers.itemRequired'));
      return;
    }
    try {
      await createMut.mutateAsync({
        fromWarehouseId,
        toWarehouseId,
        notes: notes.trim(),
        items: validItems,
      });
      toast.success(t('erp.stockTransfers.draftCreated'));
      setShowCreate(false);
      setFromWarehouseId('');
      setToWarehouseId('');
      setNotes('');
      setNewItems([{ productId: '', quantity: '' }]);
    } catch {
      toast.error(t('erp.stockTransfers.createFailed'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/transfers'].title))} />
      <PageHeader
        title={t(routeLabels['/erp/transfers'].title)}
        description={t('erp.stockTransfers.subtitle')}
        actions={
          canCreateTransfer && (
            <button
              type="button"
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
            >
              <Plus size={14} /> New Transfer
            </button>
          )
        }
      />

      <div
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3 text-[10px] font-semibold uppercase tracking-wider"
          style={{
            gridTemplateColumns: '0.5fr 1.5fr 0.3fr 1.5fr 0.7fr 0.8fr 0.8fr 1fr',
            color: 'var(--hub-text-dim)',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          <span>{t('erp.stockTransfers.idColumn')}</span>
          <span>{t('erp.stockTransfers.from')}</span>
          <span></span>
          <span>{t('erp.stockTransfers.to')}</span>
          <span>{t('erp.stockTransfers.items')}</span>
          <span>{t('erp.stockTransfers.created')}</span>
          <span>{t('erpCommon.status')}</span>
          <span></span>
        </div>
        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={8} />)
          ) : transfersError ? (
            <div
              role="alert"
              className="flex flex-col items-center gap-3 p-8 text-center"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <AlertTriangle size={24} style={{ color: 'var(--erp-danger)' }} />
              <div>
                <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  Unable to load stock transfers
                </p>
                <p className="mt-1 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                  The transfer history could not be loaded from the tenant API.
                </p>
              </div>
              <button
                type="button"
                onClick={() => void refetchTransfers()}
                className="rounded-lg px-3 py-1.5 text-xs font-semibold"
                style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
              >
                {t('erp.stockTransfers.refresh')}
              </button>
            </div>
          ) : transfers.length === 0 ? (
            <EmptyState
              icon={ArrowLeftRight}
              title={t('erp.stockTransfers.empty.title')}
              description={t('erp.stockTransfers.empty.description')}
              action={
                canCreateTransfer
                  ? {
                      label: t('erp.stockTransfers.newTransfer'),
                      onClick: () => setShowCreate(true),
                    }
                  : {
                      label: t('erp.stockTransfers.refresh'),
                      onClick: () => void refetchTransfers(),
                    }
              }
            />
          ) : (
            transfers.map((transfer) => {
              const sc = statusConfig(transfer.status ?? '');
              return (
                <div
                  key={transfer.id}
                  className="grid px-4 py-3 items-center"
                  style={{ gridTemplateColumns: '0.5fr 1.5fr 0.3fr 1.5fr 0.7fr 0.8fr 0.8fr 1fr' }}
                >
                  <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                    #{transfer.id}
                  </span>
                  <span
                    className="text-xs font-medium truncate pr-2"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {warehouseName(transfer.fromWarehouseId ?? '')}
                  </span>
                  <ArrowRight size={12} style={{ color: 'var(--hub-text-dim)' }} />
                  <span
                    className="text-xs font-medium truncate pr-2"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {warehouseName(transfer.toWarehouseId ?? '')}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text)' }}>
                    {(transfer.items ?? []).length}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {formatDate(transfer.createdAt ?? '')}
                  </span>
                  <span>
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{ background: sc.bg, color: sc.color }}
                    >
                      {stockTransferStatusLabel(transfer.status ?? '')}
                    </span>
                  </span>
                  {canCreateTransfer &&
                  transfer.id &&
                  (transfer.status === 'draft' || transfer.status === 'in_transit') ? (
                    <ReceiveTransferAction transferId={transfer.id} />
                  ) : null}
                </div>
              );
            })
          )}
        </div>
      </div>
      <PaginationControls
        page={page}
        hasMore={hasMore}
        itemCount={transfers.length}
        onPageChange={setPage}
      />
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
              {t('erp.stockTransfers.newTransferTitle')}
            </h3>
            {warehousesError ? (
              <div
                role="alert"
                className="rounded-lg p-2 text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--erp-danger)' }}
              >
                <p>Unable to load warehouses.</p>
                <button
                  type="button"
                  onClick={() => void refetchWarehouses()}
                  className="mt-1 underline"
                >
                  Retry
                </button>
              </div>
            ) : null}
            <select
              value={fromWarehouseId}
              onChange={(e) => setFromWarehouseId(e.target.value)}
              disabled={warehousesLoading || warehousesError}
              style={inputStyle}
            >
              <option value="">
                {warehousesLoading
                  ? t('erp.stockTransfers.loadingWarehouses')
                  : warehousesError
                    ? t('erp.stockTransfers.warehousesUnavailable')
                    : t('erp.stockTransfers.fromWarehouse')}
              </option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name}
                </option>
              ))}
            </select>
            <select
              value={toWarehouseId}
              onChange={(e) => setToWarehouseId(e.target.value)}
              disabled={warehousesLoading || warehousesError}
              style={inputStyle}
            >
              <option value="">
                {warehousesLoading
                  ? t('erp.stockTransfers.loadingWarehouses')
                  : warehousesError
                    ? t('erp.stockTransfers.warehousesUnavailable')
                    : t('erp.stockTransfers.toWarehouse')}
              </option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name}
                </option>
              ))}
            </select>
            <input
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder={t('erpCommon.notesOptional')}
              style={inputStyle}
            />
            <p
              className="text-[11px] font-semibold uppercase tracking-wide"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              Items *
            </p>
            <input
              value={productSearch}
              onChange={(e) => {
                setProductSearch(e.target.value);
                setProductPage(1);
              }}
              placeholder={t('erp.stockTransfers.searchProducts')}
              aria-label={t('erp.stockTransfers.searchProducts')}
              disabled={productsError}
              style={inputStyle}
            />
            {productsError ? (
              <div
                role="alert"
                className="rounded-lg p-2 text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--erp-danger)' }}
              >
                <p>Unable to load products.</p>
                <button
                  type="button"
                  onClick={() => void refetchProducts()}
                  className="mt-1 underline"
                >
                  Retry
                </button>
              </div>
            ) : null}
            {newItems.map((item, idx) => (
              <div key={idx} className="flex gap-1 items-center">
                <select
                  value={item.productId}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) => (i === idx ? { ...it, productId: e.target.value } : it))
                    )
                  }
                  aria-label={`Product for transfer item ${idx + 1}`}
                  disabled={productsLoading || productsError || products.length === 0}
                  style={lineInputStyle}
                >
                  <option value="">
                    {productsLoading
                      ? t('erp.stockTransfers.loadingProducts')
                      : productsError
                        ? t('erp.stockTransfers.productsUnavailable')
                        : products.length === 0
                          ? t('erp.stockTransfers.noProducts')
                          : t('erp.stockTransfers.selectProduct')}
                  </option>
                  {products.map((product) => (
                    <option key={product.id} value={product.id}>
                      {product.name ?? `Product #${product.id}`}
                      {product.sku ? ` (${product.sku})` : ''}
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min="0.01"
                  step="0.01"
                  value={item.quantity}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) => (i === idx ? { ...it, quantity: e.target.value } : it))
                    )
                  }
                  aria-label={`Quantity for transfer item ${idx + 1}`}
                  placeholder="Qty"
                  style={quantityInputStyle}
                />
                <button
                  type="button"
                  onClick={() => setNewItems((prev) => prev.filter((_, i) => i !== idx))}
                  disabled={newItems.length === 1}
                  aria-label={`Remove transfer item ${idx + 1}`}
                  className="p-2 rounded-lg disabled:opacity-30"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
            {!productsError && (productPage > 1 || productsHaveMore) && (
              <PaginationControls
                page={productPage}
                hasMore={productsHaveMore}
                itemCount={products.length}
                pageSize={PRODUCT_PAGE_SIZE}
                total={productsTotal}
                onPageChange={setProductPage}
              />
            )}
            <button
              type="button"
              onClick={() => setNewItems((prev) => [...prev, { productId: '', quantity: '' }])}
              className="text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              + Add item
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
                onClick={handleCreate}
                disabled={
                  createMut.isPending ||
                  !fromWarehouseId ||
                  !toWarehouseId ||
                  validItems.length === 0
                }
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: ACCENT, color: 'var(--erp-accent-foreground)' }}
              >
                {createMut.isPending ? 'Creating…' : 'Create Draft'}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default StockTransfers;
