import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, ArrowRight } from 'lucide-react';
import { useStockTransfers, useCreateStockTransfer } from './useStockTransfers';
import PageHeader from '../../shared/components/PageHeader';
import EmptyState from '../../shared/components/EmptyState';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { formatDate } from '../../shared/lib/format';
import { StockTransfer } from '../../shared/services/erpApi';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

function statusConfig(status: StockTransfer['status']) {
  if (status === 'Received') return { color: '#10b981', bg: 'rgba(16,185,129,0.12)' };
  if (status === 'In Transit') return { color: AMBER, bg: 'rgba(245,158,11,0.12)' };
  if (status === 'Cancelled') return { color: '#f87171', bg: 'rgba(248,113,113,0.12)' };
  return { color: 'var(--hub-text-dim)', bg: 'var(--hub-hover)' };
}

const StockTransfers: React.FC = () => {
  const { data: transferResult, isLoading } = useStockTransfers();
  const transfers = transferResult?.transfers ?? [];
  const createMut = useCreateStockTransfer();
  const [showCreate, setShowCreate] = useState(false);
  const [fromName, setFromName] = useState('');
  const [toName, setToName] = useState('');

  const handleCreate = async (): Promise<void> => {
    if (!fromName.trim() || !toName.trim()) return;
    try {
      await createMut.mutateAsync({
        fromName: fromName.trim(),
        toName: toName.trim(),
        status: 'Draft',
      });
      toast.success('Transfer draft created');
      setShowCreate(false);
      setFromName('');
      setToName('');
    } catch {
      toast.error('Failed to create transfer');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-6xl"
    >
      <PageHeader
        title="Stock Transfers"
        description="Move inventory between warehouses"
        actions={
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Transfer
          </button>
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
          <span>ID</span>
          <span>From</span>
          <span></span>
          <span>To</span>
          <span>Items</span>
          <span>Created</span>
          <span>Status</span>
        </div>
        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading ? (
            Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={7} />)
          ) : transfers.length === 0 ? (
            <EmptyState title="No transfers" description="Create your first stock transfer" />
          ) : (
            transfers.map((t) => {
              const sc = statusConfig(t.status);
              return (
                <div
                  key={t.id}
                  className="grid px-4 py-3 items-center"
                  style={{ gridTemplateColumns: '0.5fr 1.5fr 0.3fr 1.5fr 0.7fr 0.8fr 0.8fr' }}
                >
                  <span className="text-xs font-mono" style={{ color: 'var(--hub-text-muted)' }}>
                    #{t.id}
                  </span>
                  <span
                    className="text-xs font-medium truncate pr-2"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {t.fromName}
                  </span>
                  <ArrowRight size={12} style={{ color: 'var(--hub-text-dim)' }} />
                  <span
                    className="text-xs font-medium truncate pr-2"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {t.toName}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text)' }}>
                    {t.items}
                  </span>
                  <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {formatDate(t.createdAt)}
                  </span>
                  <span>
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                      style={{ background: sc.bg, color: sc.color }}
                    >
                      {t.status}
                    </span>
                  </span>
                </div>
              );
            })
          )}
        </div>
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
              New Stock Transfer
            </h3>
            <input
              value={fromName}
              onChange={(e) => setFromName(e.target.value)}
              placeholder="From warehouse"
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <input
              value={toName}
              onChange={(e) => setToName(e.target.value)}
              placeholder="To warehouse"
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
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
                disabled={createMut.isPending}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: '#000' }}
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
