import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, ClipboardList, CheckCircle2 } from 'lucide-react';
import { useStockTakes, useCreateStockTake, useUpdateStockTake } from './useStockTakes';
import {
  stockTakeStatusLabel,
  stockTakeProgress,
  stockTakeVariance,
  type CreateStockTakeItemInput,
} from './stockTakeSchema';
import { useWarehouses } from '../warehouses/useWarehouses';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { formatDate } from '../../shared/lib/format';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const inputStyle: React.CSSProperties = {
  padding: '8px 12px',
  borderRadius: '8px',
  border: '1px solid var(--hub-border)',
  background: 'var(--hub-input-bg)',
  color: 'var(--hub-text)',
  fontSize: '13px',
  outline: 'none',
  width: '100%',
  boxSizing: 'border-box',
};

const StockTakes: React.FC = () => {
  const { data: stockTakeResult, isLoading } = useStockTakes();
  const stockTakes = stockTakeResult?.stockTakes ?? [];
  const { data: warehouses = [] } = useWarehouses();
  const createMut = useCreateStockTake();
  const updateMut = useUpdateStockTake();
  const [showCreate, setShowCreate] = useState(false);
  const [newWarehouseId, setNewWarehouseId] = useState<number>(0);
  const [newNotes, setNewNotes] = useState('');

  const warehouseName = (id: number): string =>
    warehouses.find((w) => w.id === id)?.name ?? `Warehouse #${id}`;

  const handleCreate = async (): Promise<void> => {
    if (!newWarehouseId) {
      toast.error('Warehouse is required');
      return;
    }
    try {
      await createMut.mutateAsync({
        warehouseId: newWarehouseId,
        notes: newNotes.trim(),
        items: [],
      });
      toast.success('Stock take created');
      setShowCreate(false);
      setNewWarehouseId(0);
      setNewNotes('');
    } catch {
      toast.error('Failed to create stock take');
    }
  };

  const handleComplete = async (id: number, items: CreateStockTakeItemInput[]): Promise<void> => {
    const uncounted = items.filter((item) => item.expectedQty > 0 && !item.countedQty).length;
    if (
      uncounted > 0 &&
      !window.confirm(
        `You have ${uncounted} uncounted item${uncounted === 1 ? '' : 's'}. Complete anyway?`
      )
    ) {
      return;
    }
    try {
      await updateMut.mutateAsync({ id, status: 'completed', items });
      toast.success('Stock take completed');
    } catch {
      toast.error('Failed to complete stock take');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <PageHeader
        title="Stock Takes"
        description="Periodic inventory counts and variance reports"
        actions={
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Stock Take
          </button>
        }
      />

      <div className="space-y-3">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={5} />)
        ) : stockTakes.length === 0 ? (
          <EmptyState
            icon={ClipboardList}
            title="No stock takes"
            description="Start a count to find discrepancies"
          />
        ) : (
          stockTakes.map((st) => {
            const progress = stockTakeProgress(st.items);
            const variance = stockTakeVariance(st.items);
            const statusColor =
              st.status === 'completed'
                ? '#10b981'
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
                      style={{ background: 'rgba(245,158,11,0.1)' }}
                    >
                      <ClipboardList size={16} style={{ color: AMBER }} />
                    </div>
                    <div>
                      <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                        {warehouseName(st.warehouseId)}
                      </p>
                      <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                        #{st.id} / Created {formatDate(st.createdAt)}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {st.status !== 'completed' && (
                      <button
                        aria-label="Complete stock take"
                        onClick={() =>
                          handleComplete(
                            st.id,
                            st.items.map((item) => ({
                              productId: item.productId,
                              expectedQty: item.expectedQty,
                              countedQty: item.countedQty,
                            }))
                          )
                        }
                        disabled={updateMut.isPending}
                        className="flex items-center gap-1 text-[10px] font-medium px-2 py-1 rounded-lg disabled:opacity-40"
                        style={{ background: 'rgba(16,185,129,0.12)', color: '#10b981' }}
                      >
                        <CheckCircle2 size={12} /> Complete
                      </button>
                    )}
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{ background: `${statusColor}20`, color: statusColor }}
                    >
                      {stockTakeStatusLabel(st.status)}
                    </span>
                  </div>
                </div>
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
                      VARIANCE
                    </p>
                    <p
                      className="text-sm font-bold"
                      style={{ color: variance !== 0 ? '#f87171' : '#10b981' }}
                    >
                      {variance !== 0 ? `${variance} items` : 'None'}
                    </p>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>

      {showCreate && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ background: 'rgba(0,0,0,0.5)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              New Stock Take
            </h3>
            <select
              value={newWarehouseId}
              onChange={(e) => setNewWarehouseId(Number(e.target.value))}
              style={inputStyle}
            >
              <option value={0}>Select warehouse…</option>
              {warehouses.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name}
                </option>
              ))}
            </select>
            <input
              value={newNotes}
              onChange={(e) => setNewNotes(e.target.value)}
              placeholder="Notes (optional)"
              style={inputStyle}
            />
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                Cancel
              </button>
              <button
                onClick={handleCreate}
                disabled={createMut.isPending || !newWarehouseId}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: '#000' }}
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
