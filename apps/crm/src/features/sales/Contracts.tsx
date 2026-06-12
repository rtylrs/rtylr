import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { ChevronDown, ChevronUp, Plus, FileText } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { useContracts, useLeads, useCreateContract } from '../../shared/hooks/queries/useSales';
import type { Contract, ContractPlan } from '../../shared/hooks/queries/useSales';
import {
  Button,
  PageHeader,
  EmptyState,
  SkeletonList,
  Modal,
  Input,
  Select,
  Textarea,
} from '../../shared/components/ui';

const PLAN_COLORS: Record<ContractPlan, { bg: string; color: string }> = {
  free: { bg: 'rgba(107,114,128,0.12)', color: '#6b7280' },
  pro: { bg: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
  scale: { bg: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
};

function formatDate(iso: string): string {
  if (!iso) return '-';
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

const PlanBadge: React.FC<{ plan: ContractPlan }> = ({ plan }) => {
  const colors = PLAN_COLORS[plan];
  return (
    <span
      style={{
        fontSize: '11px',
        fontWeight: 600,
        padding: '2px 8px',
        borderRadius: '6px',
        background: colors.bg,
        color: colors.color,
        textTransform: 'capitalize',
      }}
    >
      {plan}
    </span>
  );
};

const Contracts: React.FC = () => {
  const { data: allContracts, isLoading } = useContracts();
  const { data: leads } = useLeads();
  const createContract = useCreateContract();

  const contracts = allContracts ?? [];
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [form, setForm] = useState({
    leadId: '',
    plan: 'pro' as ContractPlan,
    mrr: '',
    seats: '',
    signedAt: '',
    signedBy: '',
    closedBy: '',
    notes: '',
  });

  const totalMrr = contracts.reduce((sum, c) => sum + c.mrr, 0);

  const companyName = (contract: Contract): string => {
    const lead = (leads ?? []).find((l) => l.id === contract.leadId);
    return lead?.company ?? contract.company ?? `Lead #${contract.leadId}`;
  };

  const handleSubmit = (): void => {
    if (!form.leadId || !form.signedAt || !form.signedBy) {
      toast.error('Lead, signed date, and signed by are required');
      return;
    }
    createContract.mutate(
      {
        leadId: Number(form.leadId),
        plan: form.plan,
        mrr: Number(form.mrr),
        seats: Number(form.seats),
        signedAt: form.signedAt,
        signedBy: form.signedBy,
        closedBy: form.closedBy,
        notes: form.notes,
      },
      {
        onSuccess: () => {
          setCreateOpen(false);
          setForm({
            leadId: '',
            plan: 'pro',
            mrr: '',
            seats: '',
            signedAt: '',
            signedBy: '',
            closedBy: '',
            notes: '',
          });
        },
        onError: () => toast.error('Failed to create contract'),
      }
    );
  };

  return (
    <>
      <HelmetContainer title="Contracts · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1100px' }}>
        <PageHeader
          title="Contracts"
          subtitle="All signed contracts and their details"
          actions={
            <Button variant="primary" onClick={() => setCreateOpen(true)}>
              <Plus size={13} />
              New contract
            </Button>
          }
        />

        {/* Total MRR banner */}
        {contracts.length > 0 && (
          <div
            style={{
              padding: '14px 20px',
              background: 'rgba(16,185,129,0.08)',
              border: '1px solid rgba(16,185,129,0.2)',
              borderRadius: '10px',
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
            }}
          >
            <div
              style={{
                width: '32px',
                height: '32px',
                borderRadius: '8px',
                background: 'rgba(16,185,129,0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <FileText size={16} style={{ color: '#10b981' }} />
            </div>
            <div>
              <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: 0 }}>
                Total contracted MRR
              </p>
              <p style={{ fontSize: '20px', fontWeight: 700, color: '#10b981', margin: 0 }}>
                ${totalMrr.toLocaleString()}
                <span style={{ fontSize: '12px', fontWeight: 400, color: 'var(--hub-text-muted)' }}>
                  /mo
                </span>
              </p>
            </div>
            <div style={{ marginLeft: 'auto' }}>
              <p
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-muted)',
                  margin: 0,
                  textAlign: 'right',
                }}
              >
                {contracts.length} active contract{contracts.length !== 1 ? 's' : ''}
              </p>
            </div>
          </div>
        )}

        {/* Contracts list */}
        <div
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '12px',
            overflow: 'hidden',
          }}
        >
          {isLoading ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={4} />
            </div>
          ) : contracts.length === 0 ? (
            <EmptyState
              icon={FileText}
              title="No contracts yet"
              description="Signed contracts will appear here once you close a deal."
              action={
                <Button variant="primary" onClick={() => setCreateOpen(true)}>
                  <Plus size={13} />
                  New contract
                </Button>
              }
            />
          ) : (
            <div>
              {contracts.map((c: Contract) => {
                const isExpanded = expandedId === c.id;
                const name = companyName(c);
                return (
                  <React.Fragment key={c.id}>
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        padding: '14px 20px',
                        borderBottom: isExpanded ? 'none' : '1px solid var(--hub-border)',
                        cursor: 'pointer',
                        transition: 'background 0.12s',
                        background: isExpanded ? 'rgba(59,130,246,0.03)' : 'transparent',
                      }}
                      onClick={() => setExpandedId(isExpanded ? null : c.id)}
                      onMouseEnter={(e) => {
                        if (!isExpanded)
                          (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                      }}
                      onMouseLeave={(e) => {
                        if (!isExpanded)
                          (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                    >
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                          <p
                            style={{
                              fontSize: '14px',
                              fontWeight: 600,
                              color: 'var(--hub-text)',
                              margin: 0,
                            }}
                          >
                            {name}
                          </p>
                          <PlanBadge plan={c.plan} />
                        </div>
                        <p
                          style={{
                            fontSize: '12px',
                            color: 'var(--hub-text-muted)',
                            margin: '2px 0 0',
                          }}
                        >
                          Closed by {c.closedBy} · Signed {formatDate(c.signedAt)}
                        </p>
                      </div>
                      <div
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '20px',
                          flexShrink: 0,
                        }}
                      >
                        <div style={{ textAlign: 'right' }}>
                          <p
                            style={{
                              fontSize: '14px',
                              fontWeight: 700,
                              color: 'var(--hub-text)',
                              margin: 0,
                            }}
                          >
                            ${c.mrr.toLocaleString()}/mo
                          </p>
                          <p
                            style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: 0 }}
                          >
                            {c.seats} seats
                          </p>
                        </div>
                        {isExpanded ? (
                          <ChevronUp size={16} style={{ color: 'var(--hub-text-dim)' }} />
                        ) : (
                          <ChevronDown size={16} style={{ color: 'var(--hub-text-dim)' }} />
                        )}
                      </div>
                    </div>

                    <AnimatePresence>
                      {isExpanded && (
                        <motion.div
                          initial={{ opacity: 0, height: 0 }}
                          animate={{ opacity: 1, height: 'auto' }}
                          exit={{ opacity: 0, height: 0 }}
                          style={{
                            overflow: 'hidden',
                            borderBottom: '1px solid var(--hub-border)',
                          }}
                        >
                          <div
                            style={{
                              padding: '16px 20px 20px',
                              background: 'rgba(59,130,246,0.03)',
                              display: 'grid',
                              gridTemplateColumns: 'repeat(3, 1fr)',
                              gap: '16px',
                            }}
                          >
                            <DetailItem label="Plan" value={c.plan} />
                            <DetailItem label="MRR" value={`$${c.mrr.toLocaleString()}/mo`} />
                            <DetailItem label="Seats" value={String(c.seats)} />
                            <DetailItem label="Signed At" value={formatDate(c.signedAt)} />
                            <DetailItem label="Signed By" value={c.signedBy} />
                            <DetailItem label="Closed By" value={c.closedBy} />
                            {c.notes && (
                              <div style={{ gridColumn: '1 / -1' }}>
                                <p
                                  style={{
                                    fontSize: '10px',
                                    color: 'var(--hub-text-dim)',
                                    textTransform: 'uppercase',
                                    letterSpacing: '0.05em',
                                    fontWeight: 600,
                                    margin: '0 0 4px',
                                  }}
                                >
                                  Notes
                                </p>
                                <p
                                  style={{
                                    fontSize: '13px',
                                    color: 'var(--hub-text)',
                                    margin: 0,
                                    lineHeight: 1.5,
                                  }}
                                >
                                  {c.notes}
                                </p>
                              </div>
                            )}
                          </div>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </React.Fragment>
                );
              })}
            </div>
          )}
        </div>

        {/* Create contract modal */}
        <Modal
          open={createOpen}
          onClose={() => setCreateOpen(false)}
          title="New Contract"
          size="md"
          footer={
            <>
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                Cancel
              </Button>
              <Button
                variant="primary"
                onClick={handleSubmit}
                disabled={
                  createContract.isPending || !form.leadId || !form.signedAt || !form.signedBy
                }
              >
                {createContract.isPending ? 'Saving...' : 'Save contract'}
              </Button>
            </>
          }
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            <div>
              <p
                style={{
                  fontSize: '11px',
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
                  fontWeight: 600,
                  margin: '0 0 4px',
                }}
              >
                Lead
              </p>
              <select
                value={form.leadId}
                onChange={(e) => setForm((f) => ({ ...f, leadId: e.target.value }))}
                style={{
                  width: '100%',
                  height: '36px',
                  padding: '0 10px',
                  borderRadius: '8px',
                  background: 'var(--hub-input-bg)',
                  border: '1px solid var(--hub-border)',
                  color: 'var(--hub-text)',
                  fontSize: '13px',
                  outline: 'none',
                  cursor: 'pointer',
                }}
              >
                <option value="">Select a lead...</option>
                {(leads ?? []).map((l) => (
                  <option key={l.id} value={String(l.id)}>
                    {l.company}
                  </option>
                ))}
              </select>
            </div>
            <Select
              label="Plan"
              value={form.plan}
              onChange={(e) => setForm((f) => ({ ...f, plan: e.target.value as ContractPlan }))}
            >
              <option value="free">Free</option>
              <option value="pro">Pro</option>
              <option value="scale">Scale</option>
            </Select>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
              <Input
                label="MRR ($)"
                type="number"
                value={form.mrr}
                onChange={(e) => setForm((f) => ({ ...f, mrr: e.target.value }))}
                placeholder="500"
              />
              <Input
                label="Seats"
                type="number"
                value={form.seats}
                onChange={(e) => setForm((f) => ({ ...f, seats: e.target.value }))}
                placeholder="5"
              />
            </div>
            <Input
              label="Signed At"
              type="date"
              value={form.signedAt}
              onChange={(e) => setForm((f) => ({ ...f, signedAt: e.target.value }))}
            />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
              <Input
                label="Signed By"
                value={form.signedBy}
                onChange={(e) => setForm((f) => ({ ...f, signedBy: e.target.value }))}
                placeholder="John Smith"
              />
              <Input
                label="Closed By (rep)"
                value={form.closedBy}
                onChange={(e) => setForm((f) => ({ ...f, closedBy: e.target.value }))}
                placeholder="Sales rep name"
              />
            </div>
            <Textarea
              label="Notes"
              value={form.notes}
              onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
              placeholder="Additional notes..."
            />
          </div>
        </Modal>
      </div>
    </>
  );
};

const DetailItem: React.FC<{ label: string; value: string }> = ({ label, value }) => (
  <div>
    <p
      style={{
        fontSize: '10px',
        color: 'var(--hub-text-dim)',
        textTransform: 'uppercase',
        letterSpacing: '0.05em',
        fontWeight: 600,
        margin: '0 0 2px',
      }}
    >
      {label}
    </p>
    <p
      style={{ fontSize: '13px', color: 'var(--hub-text)', margin: 0, textTransform: 'capitalize' }}
    >
      {value || '-'}
    </p>
  </div>
);

export default Contracts;
