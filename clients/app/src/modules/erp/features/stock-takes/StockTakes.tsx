import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, ClipboardList, CheckCircle2 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import { useTranslation } from 'react-i18next';
import {
  useStockTakes,
  useCreateStockTake,
  useUpdateStockTake,
  type StockTake,
} from './useStockTakes';
import { canErpPermission } from '../../shared/lib/erpPermissions';
import { stockTakeStatusLabel, stockTakeProgress } from './stockTakeSchema';
import type { UpdateStockTakeItemBody } from '@voxire/voxie/erp/stock_take/v1/types';
import { useWarehouses } from '../warehouses/useWarehouses';
import { useInventoryReport } from '../inventory/useInventory';
import { buildStockTakeItems } from './stockTakeUtils';
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

type StockTakeItem = NonNullable<NonNullable<StockTake['items']>[number]>;

function stockTakeItemKey(stockTakeId: string | undefined, productId: string | undefined): string {
  return `${stockTakeId ?? ''}:${productId ?? ''}`;
}

function parseCountedQuantity(value: string | number | null | undefined): number | null {
  if (value == null || (typeof value === 'string' && value.trim() === '')) return null;
  const count = Number(value);
  return Number.isFinite(count) && count >= 0 ? count : null;
}

const StockTakes: React.FC = () => {
  const { t } = useTranslation('erp');
  const [page, setPage] = useState(1);
  const pageSize = 50;
  const {
    data: stockTakeResult,
    isLoading,
    isFetching: stockTakesFetching,
    isError: stockTakesError,
    refetch,
  } = useStockTakes({ page, pageSize });
  const stockTakes = stockTakeResult?.stockTakes ?? [];
  const hasMore = stockTakeResult?.hasMore ?? false;
  const { data: warehouses = [] } = useWarehouses();
  const {
    data: inventoryItems = [],
    isLoading: inventoryLoading,
    isError: inventoryError,
  } = useInventoryReport();
  const createMut = useCreateStockTake();
  const updateMut = useUpdateStockTake();
  const effectiveAccess = useUserStore((state) => state.user?.effectiveAccess);
  const canCreateStockTakes = canErpPermission('erp:inventory:adjust');
  const canCompleteStockTakes = canErpPermission('erp:inventory:approve', effectiveAccess);
  const [showCreate, setShowCreate] = useState(false);
  const [newWarehouseId, setNewWarehouseId] = useState('');
  const [newNotes, setNewNotes] = useState('');
  const [countInputs, setCountInputs] = useState<Record<string, string>>({});
  const itemsToCount = buildStockTakeItems(inventoryItems).map(({ productId, expectedQty }) => ({
    productId,
    expectedQty,
  }));

  const warehouseName = (id: string): string =>
    warehouses.find((w) => w.id === id)?.name ?? `Warehouse #${id}`;

  const getCountValue = (st: StockTake, item: StockTakeItem): number | null => {
    const key = stockTakeItemKey(st.id, item.productId);
    if (Object.prototype.hasOwnProperty.call(countInputs, key)) {
      return parseCountedQuantity(countInputs[key]);
    }
    if (item.countedQty == null) return null;
    if (st.status === 'completed' || item.countedQty !== 0) {
      return parseCountedQuantity(item.countedQty);
    }
    return null;
  };

  const getCompletionItems = (st: StockTake): UpdateStockTakeItemBody[] | null => {
    const sourceItems = st.items ?? [];
    if (sourceItems.length === 0) return null;

    const items: UpdateStockTakeItemBody[] = [];
    for (const item of sourceItems) {
      const countedQty = getCountValue(st, item);
      if (!item.productId || item.expectedQty == null || countedQty == null) return null;
      items.push({ productId: item.productId, expectedQty: item.expectedQty, countedQty });
    }
    return items;
  };

  const updateCount = (st: StockTake, item: StockTakeItem, value: string): void => {
    const key = stockTakeItemKey(st.id, item.productId);
    setCountInputs((current) => ({ ...current, [key]: value }));
  };

  const handleCreate = async (): Promise<void> => {
    if (!newWarehouseId) {
      toast.error(t('erp.stockTakes.warehouseRequired'));
      return;
    }
    if (inventoryLoading) {
      toast.error(t('erp.stockTakes.stillLoading'));
      return;
    }
    if (itemsToCount.length === 0) {
      toast.error(
        inventoryError
          ? t('erp.stockTakes.inventoryRequiredForCreate')
          : t('erp.stockTakes.noInventoryToCount')
      );
      return;
    }
    try {
      await createMut.mutateAsync({
        warehouseId: newWarehouseId,
        notes: newNotes.trim(),
        items: itemsToCount,
      });
      toast.success(t('erp.stockTakes.created'));
      setShowCreate(false);
      setNewWarehouseId('');
      setNewNotes('');
    } catch {
      toast.error(t('erp.stockTakes.createFailed'));
    }
  };

  const handleComplete = async (st: StockTake): Promise<void> => {
    const items = getCompletionItems(st);
    if (!st.id || !items) {
      toast.error(t('erp.stockTakes.completeFailed'));
      return;
    }
    try {
      await updateMut.mutateAsync({
        id: st.id,
        body: { status: 'completed', notes: st.notes ?? '', items },
      });
      toast.success(t('erp.stockTakes.completed'));
    } catch {
      toast.error(t('erp.stockTakes.completeFailed'));
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/stock-takes'].title))} />
      <PageHeader
        title={t(routeLabels['/erp/stock-takes'].title)}
        description={t('erp.stockTakes.subtitle')}
        actions={
          canCreateStockTakes &&
          !stockTakesError && (
            <button
              type="button"
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
              style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
            >
              <Plus size={14} /> New Stock Take
            </button>
          )
        }
      />

      <div className="space-y-3">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={5} />)
        ) : stockTakesError ? (
          <div
            role="alert"
            className="rounded-xl p-5 space-y-3"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
              Unable to load stock takes
            </p>
            <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
              Stock-take history could not be loaded from the tenant API.
            </p>
            <button
              type="button"
              onClick={() => void refetch()}
              disabled={stockTakesFetching}
              className="rounded-lg px-3 py-1.5 text-xs font-semibold disabled:opacity-40"
              style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
            >
              {stockTakesFetching ? 'Retrying…' : 'Retry'}
            </button>
          </div>
        ) : stockTakes.length === 0 ? (
          <EmptyState
            icon={ClipboardList}
            title={t('erp.stockTakes.empty.title')}
            description={t('erp.stockTakes.empty.description')}
            action={
              canCreateStockTakes
                ? { label: t('erp.stockTakes.newStockTake'), onClick: () => setShowCreate(true) }
                : { label: t('erp.stockTakes.refresh'), onClick: () => void refetch() }
            }
          />
        ) : (
          stockTakes.map((st) => {
            const items = st.items ?? [];
            const completionItems = getCompletionItems(st);
            const progress = stockTakeProgress(
              items.map((item) => ({ countedQty: getCountValue(st, item) == null ? -1 : 0 }))
            );
            const variance =
              items.length > 0 && items.every((item) => getCountValue(st, item) != null)
                ? items.reduce(
                    (sum, item) => sum + (getCountValue(st, item) ?? 0) - (item.expectedQty ?? 0),
                    0
                  )
                : null;
            const canComplete = completionItems !== null;
            const statusColor =
              st.status === 'completed'
                ? 'var(--erp-success)'
                : st.status === 'in_progress'
                  ? AMBER
                  : 'var(--hub-text-dim)';
            return (
              <div
                key={st.id}
                className="rounded-xl p-4"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div
                      className="h-10 w-10 rounded-lg flex items-center justify-center"
                      style={{
                        background: 'color-mix(in srgb, var(--erp-accent) 10%, transparent)',
                      }}
                    >
                      <ClipboardList size={16} style={{ color: AMBER }} />
                    </div>
                    <div>
                      <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                        {warehouseName(st.warehouseId ?? '')}
                      </p>
                      <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                        #{st.id} / Created {formatDate(st.createdAt ?? '')}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {st.status !== 'completed' && canCompleteStockTakes && (
                      <button
                        type="button"
                        aria-label={t('erp.stockTakes.completeAria')}
                        onClick={() => void handleComplete(st)}
                        disabled={updateMut.isPending || !canComplete}
                        title={!canComplete ? 'Enter a count for every item first' : undefined}
                        className="flex items-center gap-1 text-[10px] font-medium px-2 py-1 rounded-lg disabled:opacity-40"
                        style={{
                          background: 'color-mix(in srgb, var(--erp-success) 12%, transparent)',
                          color: 'var(--erp-success)',
                        }}
                      >
                        <CheckCircle2 size={12} /> Complete
                      </button>
                    )}
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{
                        background: `color-mix(in srgb, ${statusColor} 13%, transparent)`,
                        color: statusColor,
                      }}
                    >
                      {stockTakeStatusLabel(st.status ?? '')}
                    </span>
                  </div>
                </div>
                {st.status !== 'completed' && items.length > 0 && (
                  <div className="space-y-2 mb-4">
                    {items.map((item) => {
                      const key = stockTakeItemKey(st.id, item.productId);
                      const productName =
                        inventoryItems.find(
                          (inventoryItem) => inventoryItem.productId === item.productId
                        )?.productName ?? `Product #${item.productId ?? ''}`;
                      const inputValue =
                        countInputs[key] ??
                        (item.countedQty != null && item.countedQty !== 0
                          ? String(item.countedQty)
                          : '');
                      return (
                        <label
                          key={key}
                          className="flex items-center gap-3 text-xs"
                          style={{ color: 'var(--hub-text-muted)' }}
                        >
                          <span className="min-w-0 flex-1 truncate">{productName}</span>
                          <span
                            className="shrink-0 text-[10px]"
                            style={{ color: 'var(--hub-text-dim)' }}
                          >
                            Expected {item.expectedQty ?? 0}
                          </span>
                          <input
                            type="number"
                            min="0"
                            step="any"
                            inputMode="decimal"
                            value={inputValue}
                            onChange={(event) => updateCount(st, item, event.target.value)}
                            disabled={!canCompleteStockTakes}
                            aria-label={`Counted quantity for ${productName}`}
                            className="w-24 rounded-lg px-2 py-1.5 text-xs outline-none disabled:opacity-50"
                            style={{
                              background: 'var(--hub-input-bg)',
                              border: '1px solid var(--hub-border)',
                              color: 'var(--hub-text)',
                            }}
                          />
                        </label>
                      );
                    })}
                  </div>
                )}
                <div className="flex items-center gap-4 text-xs">
                  <div className="flex-1">
                    <div className="flex justify-between mb-1">
                      <span style={{ color: 'var(--hub-text-muted)' }}>
                        Progress: {progress.counted} / {progress.total}
                      </span>
                      <span style={{ color: 'var(--hub-text)' }}>{progress.pct}%</span>
                    </div>
                    <div className="h-1.5 rounded-full" style={{ background: 'var(--hub-hover)' }}>
                      <div
                        className="h-full rounded-full"
                        style={{ width: `${progress.pct}%`, background: AMBER }}
                      />
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {t('erp.stockTakes.variance')}
                    </p>
                    <p
                      className="text-sm font-bold"
                      style={{
                        color:
                          variance === null
                            ? 'var(--hub-text-dim)'
                            : variance !== 0
                              ? 'var(--erp-danger)'
                              : 'var(--erp-success)',
                      }}
                    >
                      {variance === null ? '—' : variance !== 0 ? `${variance} items` : 'None'}
                    </p>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      <PaginationControls
        page={page}
        hasMore={hasMore}
        itemCount={stockTakes.length}
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
              {t('erp.stockTakes.newStockTake')}
            </h3>
            <select
              value={newWarehouseId}
              onChange={(e) => setNewWarehouseId(e.target.value)}
              style={inputStyle}
            >
              <option value="">{t('erp.stockTakes.selectWarehouse')}</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name}
                </option>
              ))}
            </select>
            <input
              value={newNotes}
              onChange={(e) => setNewNotes(e.target.value)}
              placeholder={t('erpCommon.notesOptional')}
              style={inputStyle}
            />
            <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }} aria-live="polite">
              {inventoryLoading
                ? t('erp.stockTakes.loadingInventory')
                : inventoryError
                  ? t('erp.stockTakes.inventoryLoadFailed')
                  : t('erp.stockTakes.itemsWillBeIncluded')}
            </p>
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
                  !newWarehouseId ||
                  inventoryLoading ||
                  itemsToCount.length === 0
                }
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: 'var(--erp-accent-foreground)' }}
              >
                {createMut.isPending ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default StockTakes;
