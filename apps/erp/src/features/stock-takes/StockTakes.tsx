import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, ClipboardList } from 'lucide-react';
import { useStockTakes, useCreateStockTake } from './useStockTakes';
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
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newWarehouseId, setNewWarehouseId] = useState<number>(0);

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) {
      toast.error('Stock take name is required');
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        warehouseId: newWarehouseId || undefined,
        status: 'Draft',
      });
      toast.success('Stock take created');
      setShowCreate(false);
      setNewName('');
    } catch {
      toast.error('Failed to create stock take');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
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
            const pct = st.totalItems > 0 ? Math.round((st.itemsCounted / st.totalItems) * 100) : 0;
            const statusColor =
              st.status === 'Completed'
                ? '#10b981'
                : st.status === 'In Progress'
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
                        {st.name}
                      </p>
                      <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                        {st.warehouseName} / Created {formatDate(st.createdAt)}
                      </p>
                    </div>
                  </div>
                  <span
                    className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                    style={{ background: `${statusColor}20`, color: statusColor }}
                  >
                    {st.status}
                  </span>
                </div>
                <div className="flex items-center gap-4 text-xs">
                  <div className="flex-1">
                    <div className="flex justify-between mb-1">
                      <span style={{ color: 'var(--hub-text-muted)' }}>
                        Progress: {st.itemsCounted} / {st.totalItems}
                      </span>
                      <span style={{ color: 'var(--hub-text)' }}>{pct}%</span>
                    </div>
                    <div className="h-1.5 rounded-full" style={{ background: 'var(--hub-hover)' }}>
                      <div
                        className="h-full rounded-full"
                        style={{ width: `${pct}%`, background: AMBER }}
                      />
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-[10px]" style={{ color: 'var(--hub-text-dim)' }}>
                      VARIANCE
                    </p>
                    <p
                      className="text-sm font-bold"
                      style={{ color: st.variance > 0 ? '#f87171' : '#10b981' }}
                    >
                      {st.variance > 0 ? `${st.variance} items` : 'None'}
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
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Name (e.g. Q2 Count)"
              style={inputStyle}
            />
            {warehouses.length > 0 && (
              <select
                value={newWarehouseId}
                onChange={(e) => setNewWarehouseId(Number(e.target.value))}
                style={inputStyle}
              >
                <option value={0}>All warehouses</option>
                {warehouses.map((w) => (
                  <option key={w.id} value={w.id}>
                    {w.name}
                  </option>
                ))}
              </select>
            )}
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
                disabled={createMut.isPending || !newName.trim()}
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
