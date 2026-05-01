import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Tags } from 'lucide-react';
import { usePriceLists, useCreatePriceList } from './usePriceLists';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { toast } from 'sonner';
import { PriceList } from '../../shared/services/erpApi';

const AMBER = '#f59e0b';

const tierConfig: Record<string, { color: string; bg: string }> = {
  Retail: { color: '#10b981', bg: 'rgba(16,185,129,0.12)' },
  Wholesale: { color: AMBER, bg: 'rgba(245,158,11,0.12)' },
  VIP: { color: '#a78bfa', bg: 'rgba(167,139,250,0.12)' },
  Custom: { color: '#60a5fa', bg: 'rgba(96,165,250,0.12)' },
};

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

const TIERS: PriceList['tier'][] = ['Retail', 'Wholesale', 'VIP', 'Custom'];

const PriceLists: React.FC = () => {
  const { data: lists = [], isLoading } = usePriceLists();
  const createMut = useCreatePriceList();
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDescription, setNewDescription] = useState('');
  const [newTier, setNewTier] = useState<PriceList['tier']>('Retail');
  const [newMultiplier, setNewMultiplier] = useState('1.0');

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim()) return;
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        description: newDescription.trim(),
        tier: newTier,
        multiplier: parseFloat(newMultiplier) || 1,
      });
      toast.success('Price list created');
      setShowCreate(false);
      setNewName('');
      setNewDescription('');
      setNewTier('Retail');
      setNewMultiplier('1.0');
    } catch {
      toast.error('Failed to create price list');
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-5xl"
    >
      <PageHeader
        title="Price Lists"
        description="Customer-tier pricing strategies"
        actions={
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} /> New Price List
          </button>
        }
      />

      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-32 rounded-xl" />
          ))}
        </div>
      ) : lists.length === 0 ? (
        <EmptyState
          icon={Tags}
          title="No price lists"
          description="Create your first price list to enable tier pricing"
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {lists.map((l) => {
            const cfg = tierConfig[l.tier] ?? tierConfig.Custom;
            return (
              <div
                key={l.id}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="flex items-center gap-3">
                    <div
                      className="h-10 w-10 rounded-lg flex items-center justify-center"
                      style={{ background: cfg.bg }}
                    >
                      <Tags size={16} style={{ color: cfg.color }} />
                    </div>
                    <div>
                      <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                        {l.name}
                      </p>
                      <span
                        className="text-[10px] font-medium px-1.5 py-0.5 rounded"
                        style={{ background: cfg.bg, color: cfg.color }}
                      >
                        {l.tier}
                      </span>
                    </div>
                  </div>
                  <span className="text-xl font-bold" style={{ color: cfg.color }}>
                    {Math.round(l.multiplier * 100)}%
                  </span>
                </div>
                <p className="text-xs mb-3" style={{ color: 'var(--hub-text-muted)' }}>
                  {l.description}
                </p>
                <div
                  className="flex items-center justify-between pt-3"
                  style={{ borderTop: '1px solid var(--hub-border)' }}
                >
                  <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                    Active products
                  </span>
                  <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {l.activeCount}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}

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
              New Price List
            </h3>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Name"
              style={inputStyle}
            />
            <input
              value={newDescription}
              onChange={(e) => setNewDescription(e.target.value)}
              placeholder="Description (optional)"
              style={inputStyle}
            />
            <select
              value={newTier}
              onChange={(e) => setNewTier(e.target.value as PriceList['tier'])}
              style={inputStyle}
            >
              {TIERS.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
            <input
              type="number"
              min="0"
              step="0.01"
              value={newMultiplier}
              onChange={(e) => setNewMultiplier(e.target.value)}
              placeholder="Multiplier (e.g. 0.9 = 90%)"
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

export default PriceLists;
