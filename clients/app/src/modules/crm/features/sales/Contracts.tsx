import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { AlertCircle, ChevronDown, ChevronUp, Plus, FileText } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { useContracts, useLeads, useCreateContract } from '../../shared/hooks/queries/useSales';
import type { Contract, ContractPlan, Lead } from '../../shared/hooks/queries/useSales';
import {
  Button,
  PageHeader,
  SkeletonList,
  Modal,
  Input,
  Select,
  Textarea,
} from '../../shared/components/ui';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';

const PLAN_COLORS: Record<ContractPlan, { bg: string; color: string }> = {
  free: { bg: 'rgba(107,114,128,0.12)', color: '#6b7280' },
  pro: { bg: 'rgba(59,130,246,0.12)', color: 'var(--crm-accent)' },
  scale: { bg: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
};

const DEFAULT_PLAN_COLORS = PLAN_COLORS.pro;

function formatDate(value: string | null | undefined): string {
  const normalized = value?.trim() ?? '';
  if (!normalized) return '-';

  // The API deliberately returns DATE columns as YYYY-MM-DD. Parsing that
  // shape with `new Date(value)` treats it as UTC and shifts it backwards in
  // western time zones, so construct a local calendar date for date-only input.
  const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(normalized);
  const date = dateOnly
    ? new Date(Number(dateOnly[1]), Number(dateOnly[2]) - 1, Number(dateOnly[3]))
    : new Date(normalized);
  if (Number.isNaN(date.getTime())) return normalized;

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

function getContractSigner(contract: Contract): string {
  const wireContract = contract as Contract & {
    signedByName?: string | null;
    signed_by_name?: string | null;
  };
  return (
    wireContract.signedByName?.trim() ||
    wireContract.signed_by_name?.trim() ||
    contract.signedBy ||
    ''
  );
}

function leadDisplayName(lead: Lead): string {
  return lead.companyName?.trim() || lead.company?.trim() || `Lead #${lead.id}`;
}

const PlanBadge: React.FC<{ plan: ContractPlan | string }> = ({ plan }) => {
  const colors = Object.prototype.hasOwnProperty.call(PLAN_COLORS, plan)
    ? PLAN_COLORS[plan as ContractPlan]
    : DEFAULT_PLAN_COLORS;
  const label = plan || 'unknown';
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
      {label}
    </span>
  );
};

const Contracts: React.FC = () => {
  const { t } = useTranslation();
  const pageSize = 100;
  const [offset, setOffset] = useState(0);
  // Set once a "Load more" page comes back empty.
  const [exhausted, setExhausted] = useState(false);
  const [loadedContracts, setLoadedContracts] = useState<Contract[]>([]);
  const contractsQuery = useContracts(pageSize, offset);
  const { data: fetchedContracts, isFetching, refetch } = contractsQuery;
  const isError = hasQueryFailed(contractsQuery);
  const isLoading = contractsQuery.isPending && !isError;
  const [leadOffset, setLeadOffset] = useState(0);
  const [leadExhausted, setLeadExhausted] = useState(false);
  const [loadedLeads, setLoadedLeads] = useState<Lead[]>([]);
  const leadOptionsQuery = useLeads(pageSize, leadOffset);
  const {
    data: fetchedLeadOptions,
    isFetching: isFetchingLeadOptions,
    refetch: refetchLeadOptions,
  } = leadOptionsQuery;
  const leadOptionsError = hasQueryFailed(leadOptionsQuery);
  const leadOptionsLoading = leadOptionsQuery.isPending && !leadOptionsError;
  const createContract = useCreateContract();

  useEffect(() => {
    if (!fetchedContracts) return;
    if (offset > 0 && fetchedContracts.length === 0) {
      // Exact page-size multiple: the extra page came back empty — stop
      // offering "Load more".
      setExhausted(true);
    }
    setLoadedContracts((previous) => {
      if (offset === 0) return fetchedContracts;
      const byId = new Map(previous.map((contract) => [contract.id, contract]));
      fetchedContracts.forEach((contract) => byId.set(contract.id, contract));
      return Array.from(byId.values());
    });
  }, [fetchedContracts, offset]);

  useEffect(() => {
    if (!fetchedLeadOptions) return;
    if (leadOffset === 0) {
      setLeadExhausted(false);
    } else if (fetchedLeadOptions.length === 0) {
      setLeadExhausted(true);
    }
    setLoadedLeads((previous) => {
      if (leadOffset === 0) return fetchedLeadOptions;
      const byId = new Map(previous.map((lead) => [lead.id, lead]));
      fetchedLeadOptions.forEach((lead) => byId.set(lead.id, lead));
      return Array.from(byId.values());
    });
  }, [fetchedLeadOptions, leadOffset]);

  const contracts = loadedContracts;
  const hasMore = !exhausted && !isError && fetchedContracts?.length === pageSize;
  const hasCompleteContractSet = !isLoading && !isFetching && !isError && !hasMore;
  const hasMoreLeadOptions = !leadExhausted && fetchedLeadOptions?.length === pageSize;
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
  const trimmedSigner = form.signedBy.trim();
  const hasRequiredFields = Boolean(form.leadId && form.signedAt.trim() && trimmedSigner);

  const companyName = (contract: Contract): string => {
    const lead = loadedLeads.find((l) => l.id === contract.leadId);
    return lead ? leadDisplayName(lead) : contract.company || `Lead #${contract.leadId}`;
  };

  const handleSubmit = (): void => {
    if (!hasRequiredFields) {
      toast.error(t('crm.contracts.requiredFields'));
      return;
    }
    createContract.mutate(
      {
        leadId: Number(form.leadId),
        plan: form.plan,
        mrr: Number(form.mrr),
        seats: Number(form.seats),
        signedAt: form.signedAt,
        signedBy: trimmedSigner,
        closedBy: form.closedBy,
        notes: form.notes,
      },
      {
        onSuccess: () => {
          setOffset(0);
          setExhausted(false);
          setLoadedContracts([]);
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
        onError: () => toast.error(t('crm.contracts.createFailed')),
      }
    );
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.contracts.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1100px' }}>
        <PageHeader
          title={t('crm.contracts.pageTitle')}
          subtitle={t('crm.contracts.subtitle')}
          actions={
            <Button variant="primary" onClick={() => setCreateOpen(true)}>
              <Plus size={13} />
              {t('crm.contracts.new')}
            </Button>
          }
        />

        {/* Total MRR banner */}
        {contracts.length > 0 && hasCompleteContractSet && (
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
                {t('crm.contracts.totalMrr')}
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
          {isLoading && contracts.length === 0 ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={4} />
            </div>
          ) : isError && contracts.length === 0 ? (
            <EmptyState
              icon={AlertCircle}
              title={t('crm.contracts.loadFailedTitle')}
              description={t('crmErrors.refreshAndTryAgain')}
              action={{ label: t('crm.actions.retry'), onClick: () => refetch() }}
            />
          ) : contracts.length === 0 ? (
            <EmptyState
              icon={FileText}
              title={t('crm.contracts.empty')}
              description={t('crm.contracts.emptyDescription')}
              action={{ label: t('crm.contracts.new'), onClick: () => setCreateOpen(true) }}
            />
          ) : (
            <div>
              {isError && (
                <div
                  role="alert"
                  style={{
                    padding: '12px 16px',
                    borderBottom: '1px solid var(--hub-border)',
                    color: '#ef4444',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                  }}
                >
                  <AlertCircle size={15} />
                  <span style={{ flex: 1, fontSize: '12px' }}>
                    {t('crm.contracts.loadFailedTitle')}
                  </span>
                  <Button variant="outline" onClick={() => void refetch()} disabled={isFetching}>
                    {t('crm.actions.retry')}
                  </Button>
                </div>
              )}
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
                            <DetailItem label={t('crm.leadDetail.planLabel')} value={c.plan} />
                            <DetailItem
                              label={t('crm.leadDetail.mrr')}
                              value={`$${c.mrr.toLocaleString()}/mo`}
                            />
                            <DetailItem label={t('crm.leadDetail.seats')} value={String(c.seats)} />
                            <DetailItem
                              label={t('crm.leadDetail.signedAt')}
                              value={formatDate(c.signedAt)}
                            />
                            <DetailItem
                              label={t('crm.leadDetail.signedBy')}
                              value={getContractSigner(c)}
                            />
                            <DetailItem label={t('crm.leadDetail.closedBy')} value={c.closedBy} />
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
                                  {t('crm.leadDetail.notesLabel')}
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
              {hasMore && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '16px' }}>
                  <Button
                    variant="outline"
                    onClick={() => setOffset((current) => current + pageSize)}
                    disabled={isFetching}
                  >
                    {isFetching ? 'Loading…' : 'Load more contracts'}
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Create contract modal */}
        <Modal
          open={createOpen}
          onClose={() => setCreateOpen(false)}
          title={t('crm.contracts.new')}
          size="md"
          footer={
            <>
              <Button variant="outline" onClick={() => setCreateOpen(false)}>
                {t('crmCommon.cancel')}
              </Button>
              <Button
                variant="primary"
                onClick={handleSubmit}
                disabled={createContract.isPending || !hasRequiredFields}
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
                {t('crm.contracts.leadLabel')}
              </p>
              <select
                value={form.leadId}
                onChange={(e) => setForm((f) => ({ ...f, leadId: e.target.value }))}
                disabled={leadOptionsLoading}
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
                <option value="">
                  {leadOptionsLoading ? t('crmCommon.loading') : t('crm.contracts.selectLead')}
                </option>
                {loadedLeads.map((l) => (
                  <option key={l.id} value={String(l.id)}>
                    {leadDisplayName(l)}
                  </option>
                ))}
              </select>
              {leadOptionsError && (
                <div
                  role="alert"
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '8px',
                    color: '#ef4444',
                    fontSize: '12px',
                  }}
                >
                  <span style={{ flex: 1 }}>{t('crm.leads.loadFailedTitle')}</span>
                  <Button
                    variant="outline"
                    onClick={() => void refetchLeadOptions()}
                    disabled={isFetchingLeadOptions}
                  >
                    {t('crm.actions.retry')}
                  </Button>
                </div>
              )}
              {hasMoreLeadOptions && !leadOptionsError && (
                <div style={{ display: 'flex', justifyContent: 'center', marginTop: '8px' }}>
                  <Button
                    variant="outline"
                    onClick={() => setLeadOffset((current) => current + pageSize)}
                    disabled={isFetchingLeadOptions}
                  >
                    {isFetchingLeadOptions ? 'Loading…' : 'Load more leads'}
                  </Button>
                </div>
              )}
            </div>
            <Select
              label={t('crm.leadDetail.planLabel')}
              value={form.plan}
              onChange={(e) => setForm((f) => ({ ...f, plan: e.target.value as ContractPlan }))}
            >
              <option value="free">{t('crm.leadDetail.planFree')}</option>
              <option value="pro">{t('crm.leadDetail.planPro')}</option>
              <option value="scale">{t('crm.leadDetail.planScale')}</option>
            </Select>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
              <Input
                label={t('crm.leadDetail.mrrUsd')}
                type="number"
                value={form.mrr}
                onChange={(e) => setForm((f) => ({ ...f, mrr: e.target.value }))}
                placeholder="500"
              />
              <Input
                label={t('crm.leadDetail.seats')}
                type="number"
                value={form.seats}
                onChange={(e) => setForm((f) => ({ ...f, seats: e.target.value }))}
                placeholder="5"
              />
            </div>
            <Input
              label={t('crm.leadDetail.signedAt')}
              type="date"
              value={form.signedAt}
              onChange={(e) => setForm((f) => ({ ...f, signedAt: e.target.value }))}
            />
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
              <Input
                label={t('crm.leadDetail.signedBy')}
                value={form.signedBy}
                onChange={(e) => setForm((f) => ({ ...f, signedBy: e.target.value }))}
                placeholder={t('crm.leadDetail.signedByPlaceholder')}
              />
              <Input
                label={t('crm.leadDetail.closedByLabel')}
                value={form.closedBy}
                onChange={(e) => setForm((f) => ({ ...f, closedBy: e.target.value }))}
                placeholder={t('crm.contracts.repUserIdPlaceholder')}
              />
            </div>
            <Textarea
              label={t('crm.leadDetail.notesLabel')}
              value={form.notes}
              onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
              placeholder={t('crm.leadDetail.contractNotesPlaceholder')}
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
