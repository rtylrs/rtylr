import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { AlertCircle, ChevronDown, ChevronUp, Plus, FileText } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { useContracts, useLeads, useCreateContract } from '../../shared/hooks/queries/useSales';
import type { Contract, Lead } from '../../shared/hooks/queries/useSales';
import { Button, PageHeader, SkeletonList, Modal } from '../../shared/components/ui';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';
import { useBusinessCurrency } from '../../../../shared/hooks/queries/useBusinesses';
import { formatMoney } from '../../../../shared/lib/money';
import {
  ContractFields,
  ContractStatusBadge,
  ContractSummary,
  EMPTY_CONTRACT,
  contractDraftIssue,
  contractStatus,
  formatContractValue,
  toContractRequest,
  type ContractDraft,
} from './components/ContractForm';

function leadDisplayName(lead: Lead, fallback: string): string {
  return lead.companyName?.trim() || lead.company?.trim() || fallback;
}

const Contracts: React.FC = () => {
  const { t } = useTranslation();
  const currency = useBusinessCurrency();
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
  const [leadId, setLeadId] = useState('');
  const [draft, setDraft] = useState<ContractDraft>(EMPTY_CONTRACT);
  const draftIssue = contractDraftIssue(draft);

  // Monthly recurring revenue from contracts in force: signed and not ended.
  const inForce = contracts.filter((c) => contractStatus(c) === 'signed');
  const monthlyRecurring = inForce.reduce((sum, c) => sum + c.mrr, 0);

  const companyName = (contract: Contract): string => {
    const lead = loadedLeads.find((l) => l.id === contract.leadId);
    const fallback = t('crm.contracts.leadFallback', { id: contract.leadId });
    return lead ? leadDisplayName(lead, fallback) : contract.company || fallback;
  };

  const closeCreate = (): void => {
    setCreateOpen(false);
    setLeadId('');
    setDraft(EMPTY_CONTRACT);
  };

  const handleSubmit = (): void => {
    if (!leadId || draftIssue) {
      toast.error(t(draftIssue ?? 'crm.contracts.requiredFields'));
      return;
    }
    createContract.mutate(
      { leadId: Number(leadId), ...toContractRequest(draft) },
      {
        onSuccess: () => {
          setOffset(0);
          setExhausted(false);
          setLoadedContracts([]);
          closeCreate();
        },
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
        {inForce.length > 0 && hasCompleteContractSet && (
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
                {t('crm.contractForm.valueBy.monthly', {
                  amount: formatMoney(monthlyRecurring, currency),
                })}
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
                {t('crm.contracts.inForce', { count: inForce.length })}
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
                          <ContractStatusBadge contract={c} />
                        </div>
                        <p
                          style={{
                            fontSize: '12px',
                            color: 'var(--hub-text-muted)',
                            margin: '2px 0 0',
                          }}
                        >
                          {c.plan}
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
                            {formatContractValue(c, currency, t)}
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
                            }}
                          >
                            <ContractSummary contract={c} currency={currency} columns={3} />
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
                    {isFetching ? t('crmCommon.loading') : t('crm.contracts.loadMore')}
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Create contract modal */}
        <Modal
          open={createOpen}
          onClose={closeCreate}
          title={t('crm.contracts.new')}
          size="md"
          footer={
            <>
              <Button variant="outline" onClick={closeCreate}>
                {t('crmCommon.cancel')}
              </Button>
              <Button
                variant="primary"
                onClick={handleSubmit}
                disabled={createContract.isPending || !leadId || draftIssue !== null}
              >
                {createContract.isPending ? t('crm.states.saving') : t('crm.contracts.save')}
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
                value={leadId}
                onChange={(e) => setLeadId(e.target.value)}
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
                    {leadDisplayName(l, t('crm.contracts.leadFallback', { id: l.id }))}
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
                    {isFetchingLeadOptions
                      ? t('crmCommon.loading')
                      : t('crm.contracts.loadMoreLeads')}
                  </Button>
                </div>
              )}
            </div>
            <ContractFields
              draft={draft}
              onChange={(patch) => setDraft((d) => ({ ...d, ...patch }))}
              currency={currency}
            />
          </div>
        </Modal>
      </div>
    </>
  );
};

export default Contracts;
