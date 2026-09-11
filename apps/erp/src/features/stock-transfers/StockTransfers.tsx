import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, ArrowLeftRight, ArrowRight, Trash2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '@voxire/auth';
import { useTranslation } from 'react-i18next';
import { useStockTransfers, useCreateStockTransfer } from './useStockTransfers';
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

const AMBER = 'var(--erp-accent)';

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
      color: AMBER,
      bg: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
    };
  if (status === 'cancelled')
    return {
      color: 'var(--erp-danger)',
      bg: 'color-mix(in srgb, var(--erp-danger) 12%, transparent)',
    };
  return { color: 'var(--hub-text-dim)', bg: 'var(--hub-hover)' };
}

const StockTransfers: React.FC = () => {
  const { t } = useTranslation('erp');
  const [page, setPage] = useState(1);
  const pageSize = 50;
  const { data: transferResult, isLoading, refetch } = useStockTransfers({ page, pageSize });
  const transfers = transferResult?.stockTransfers ?? [];
  const hasMore = transferResult?.hasMore ?? false;
  const { data: warehouses = [] } = useWarehouses();
  const { data: productResult, isLoading: productsLoading } = useProducts({ pageSize: 100 });
  const products = productResult?.products ?? [];
  const createMut = useCreateStockTransfer();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canCreateTransfer = canErpPermission(role, 'erp:inventory:transfer');
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
      <HelmetContainer title={pageTitle(t(routeLabels['/transfers'].title))} />
      <PageHeader
        title={t(routeLabels['/transfers'].title)}
        description={t('erp.stockTransfers.subtitle')}
        actions={
          canCreateTransfer && (
            <button
              type="button"
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
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
            gridTemplateColumns: '0.5fr 1.5fr 0.3fr 1.5fr 0.7fr 0.8fr 0.8fr',
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
          <span>{t('common.status')}</span>
        </div>
        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={7} />)
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
                  : { label: t('erp.stockTransfers.refresh'), onClick: () => void refetch() }
              }
            />
          ) : (
            transfers.map((transfer) => {
              const sc = statusConfig(transfer.status ?? '');
              return (
                <div
                  key={transfer.id}
                  className="grid px-4 py-3 items-center"
                  style={{ gridTemplateColumns: '0.5fr 1.5fr 0.3fr 1.5fr 0.7fr 0.8fr 0.8fr' }}
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
            <select
              value={fromWarehouseId}
              onChange={(e) => setFromWarehouseId(e.target.value)}
              style={inputStyle}
            >
              <option value="">{t('erp.stockTransfers.fromWarehouse')}</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name}
                </option>
              ))}
            </select>
            <select
              value={toWarehouseId}
              onChange={(e) => setToWarehouseId(e.target.value)}
              style={inputStyle}
            >
              <option value="">{t('erp.stockTransfers.toWarehouse')}</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name}
                </option>
              ))}
            </select>
            <input
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder={t('common.notesOptional')}
              style={inputStyle}
            />
            <p
              className="text-[11px] font-semibold uppercase tracking-wide"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              Items *
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
                  aria-label={`Product for transfer item ${idx + 1}`}
                  disabled={productsLoading || products.length === 0}
                  style={lineInputStyle}
                >
                  <option value="">
                    {productsLoading
                      ? 'Loading products…'
                      : products.length === 0
                        ? 'No products available'
                        : 'Select product *'}
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
                {t('common.cancel')}
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
                style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
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
