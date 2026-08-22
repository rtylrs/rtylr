import React, { useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { Gift, Package, Plus, X } from 'lucide-react';
import { useLoyaltyRewards, useUpsertReward } from '../../shared/hooks/queries/useLoyalty';
import { PageHeader, Card, EmptyState, Button, SkeletonList } from '../../shared/components/ui';
import { useDialogFocus } from '../../shared/lib/keyboard';

interface RewardForm {
  name: string;
  description: string;
  pointsCost: number;
  stock: number;
  category: string;
}

const AddRewardModal: React.FC<{ onClose: () => void }> = ({ onClose }) => {
  const upsert = useUpsertReward();
  const dialogRef = useRef<HTMLDivElement>(null);
  useDialogFocus(true, dialogRef, onClose);
  const [form, setForm] = useState<RewardForm>({
    name: '',
    description: '',
    pointsCost: 100,
    stock: 10,
    category: 'general',
  });

  const handleSubmit = (): void => {
    if (upsert.isPending) return;
    if (!form.name.trim()) {
      toast.error('Reward name is required');
      return;
    }
    if (
      !Number.isInteger(form.pointsCost) ||
      form.pointsCost < 0 ||
      !Number.isInteger(form.stock) ||
      form.stock < 0
    ) {
      toast.error('Points cost and stock must be whole numbers of zero or more');
      return;
    }
    upsert.mutate(
      { data: { ...form, active: true } },
      {
        onSuccess: () => onClose(),
      }
    );
  };

  const inputStyle: React.CSSProperties = {
    padding: '8px 12px',
    borderRadius: '8px',
    border: '1px solid var(--hub-border)',
    background: 'var(--hub-input-bg, var(--hub-card))',
    color: 'var(--hub-text)',
    fontSize: '13px',
    outline: 'none',
    width: '100%',
    boxSizing: 'border-box',
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 9999,
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-label="New reward"
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          borderRadius: '16px',
          padding: '24px',
          width: '420px',
          maxWidth: 'calc(100vw - 32px)',
          maxHeight: 'calc(100vh - 32px)',
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: '12px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h3 style={{ fontSize: '15px', fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
            New reward
          </h3>
          <button
            type="button"
            aria-label="Close"
            onClick={onClose}
            style={{
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--hub-text-muted)',
            }}
          >
            <X size={16} />
          </button>
        </div>
        <input
          aria-label="Reward name"
          placeholder="Reward name"
          value={form.name}
          onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
          style={inputStyle}
        />
        <input
          aria-label="Reward description"
          placeholder="Description"
          value={form.description}
          onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
          style={inputStyle}
        />
        <input
          aria-label="Reward category"
          placeholder="Category (e.g. discount, freebie)"
          value={form.category}
          onChange={(e) => setForm((f) => ({ ...f, category: e.target.value }))}
          style={inputStyle}
        />
        <div style={{ display: 'flex', gap: '8px' }}>
          <div style={{ flex: 1 }}>
            <label
              style={{
                fontSize: '11px',
                color: 'var(--hub-text-muted)',
                display: 'block',
                marginBottom: '4px',
              }}
            >
              Points cost
            </label>
            <input
              aria-label="Points cost"
              type="number"
              min={0}
              value={form.pointsCost}
              onChange={(e) => setForm((f) => ({ ...f, pointsCost: Number(e.target.value) }))}
              style={inputStyle}
            />
          </div>
          <div style={{ flex: 1 }}>
            <label
              style={{
                fontSize: '11px',
                color: 'var(--hub-text-muted)',
                display: 'block',
                marginBottom: '4px',
              }}
            >
              Stock
            </label>
            <input
              aria-label="Stock"
              type="number"
              min={0}
              value={form.stock}
              onChange={(e) => setForm((f) => ({ ...f, stock: Number(e.target.value) }))}
              style={inputStyle}
            />
          </div>
        </div>
        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            disabled={
              upsert.isPending ||
              !form.name.trim() ||
              !Number.isInteger(form.pointsCost) ||
              form.pointsCost < 0 ||
              !Number.isInteger(form.stock) ||
              form.stock < 0
            }
          >
            {upsert.isPending ? 'Saving…' : 'Create reward'}
          </Button>
        </div>
      </div>
    </div>
  );
};

const LoyaltyRewards: React.FC = () => {
  const { data: fetched, isLoading, isError } = useLoyaltyRewards();
  const rewards = fetched ?? [];
  const [showAdd, setShowAdd] = useState(false);

  return (
    <>
      <HelmetContainer title="Rewards catalog · rtylr" />
      <div style={{ maxWidth: '1100px' }}>
        <PageHeader
          title="Rewards catalog"
          subtitle="What customers can redeem their points for"
          actions={
            <Button variant="primary" aria-label="Add reward" onClick={() => setShowAdd(true)}>
              <Plus size={13} />
              Add reward
            </Button>
          }
        />

        {showAdd && <AddRewardModal onClose={() => setShowAdd(false)} />}

        {isLoading ? (
          <Card>
            <SkeletonList count={4} />
          </Card>
        ) : isError ? (
          <Card>
            <EmptyState
              icon={Gift}
              title="Unable to load rewards"
              description="Refresh and try again."
            />
          </Card>
        ) : rewards.length === 0 ? (
          <Card>
            <EmptyState
              icon={Gift}
              title="No rewards yet"
              description="Add your first reward to the catalog."
              action={
                <Button variant="primary" onClick={() => setShowAdd(true)}>
                  <Plus size={13} />
                  Add reward
                </Button>
              }
            />
          </Card>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
              gap: '14px',
            }}
          >
            {rewards.map((r) => (
              <div
                key={r.id}
                style={{
                  background: 'var(--hub-card)',
                  border: '1px solid var(--hub-border)',
                  borderRadius: '14px',
                  padding: '16px',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '10px',
                  position: 'relative',
                }}
              >
                <div
                  style={{
                    width: '44px',
                    height: '44px',
                    borderRadius: '12px',
                    background: 'rgba(232,121,249,0.14)',
                    color: '#e879f9',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Gift size={20} />
                </div>
                <div>
                  <div style={{ fontSize: '14px', fontWeight: 700, color: 'var(--hub-text)' }}>
                    {r.name}
                  </div>
                  <div
                    style={{ fontSize: '11px', color: 'var(--hub-text-muted)', marginTop: '2px' }}
                  >
                    {r.description}
                  </div>
                </div>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    paddingTop: '10px',
                    borderTop: '1px solid var(--hub-border)',
                  }}
                >
                  <span
                    style={{
                      fontSize: '16px',
                      fontWeight: 700,
                      color: 'var(--crm-accent, #3b82f6)',
                    }}
                  >
                    {r.pointsCost.toLocaleString()} pts
                  </span>
                  <span
                    style={{
                      fontSize: '10px',
                      padding: '2px 8px',
                      borderRadius: '9999px',
                      background: 'var(--hub-hover)',
                      color: 'var(--hub-text-muted)',
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                  >
                    <Package size={9} />
                    {r.stock === null ? 'Unlimited' : `${r.stock} left`}
                  </span>
                </div>
                <div
                  style={{
                    fontSize: '10px',
                    color: 'var(--hub-text-dim)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                  }}
                >
                  {r.category}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default LoyaltyRewards;
