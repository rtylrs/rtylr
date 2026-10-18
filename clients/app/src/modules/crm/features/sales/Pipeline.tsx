import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { AlertCircle, ChevronRight, Plus, TrendingUp, X } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { motion } from 'motion/react';
import {
  usePipeline,
  useLeads,
  useSalesReps,
  useUpdateLead,
} from '../../shared/hooks/queries/useSales';
import type { Lead, LeadStage } from '../../shared/hooks/queries/useSales';
import { Button, PageHeader, SkeletonList } from '../../shared/components/ui';
import LeadFormModal from './components/LeadFormModal';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';
import { followUpStatus } from '../../shared/lib/dateUtils';

// Ids only: evaluated at module load, before i18next has a language. Labels
// come from crm.leadStages.<id>, shared with LeadDetail and LeadFormModal.
const STAGES: LeadStage[] = [
  'cold',
  'contacted',
  'demo_booked',
  'demo_done',
  'proposal_sent',
  'contracted',
  'lost',
];

const LEADS_PAGE_SIZE = 100;

function pipelineCutoff(): number {
  const cutoff = new Date();
  cutoff.setUTCMonth(cutoff.getUTCMonth() - 12);
  return cutoff.getTime();
}

function isInPipelineWindow(lead: Lead, cutoff: number): boolean {
  const createdAt = Date.parse(lead.createdAt);
  return Number.isFinite(createdAt) && createdAt >= cutoff;
}

const STAGE_COLORS: Record<LeadStage, string> = {
  cold: 'var(--crm-sales-cold, #6b7280)',
  contacted: 'var(--crm-sales-contacted, #3b82f6)',
  demo_booked: 'var(--crm-sales-demo-booked, #8b5cf6)',
  demo_done: 'var(--crm-sales-demo-done, #f59e0b)',
  proposal_sent: 'var(--crm-sales-proposal, #f97316)',
  contracted: 'var(--crm-sales-contracted, #10b981)',
  lost: 'var(--crm-sales-lost, #ef4444)',
};

function initials(name: string | null | undefined): string {
  if (!name) return '?';
  return name
    .split(' ')
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('');
}

interface LeadCardProps {
  lead: Lead;
  /** Assigned rep's name; the avatar is left out when unassigned. */
  repName?: string;
  onClick: () => void;
  onMoveStage?: (lead: Lead, nextStage: LeadStage) => void;
  isUpdating?: boolean;
}

const LeadCard: React.FC<LeadCardProps> = ({ lead, repName, onClick, onMoveStage, isUpdating }) => {
  const { t } = useTranslation();
  const followUp = lead.nextFollowUp ? followUpStatus(lead.nextFollowUp) : null;
  const currentIndex = STAGES.indexOf(lead.stage);
  const nextStage =
    currentIndex >= 0 && currentIndex < STAGES.length - 2 ? STAGES[currentIndex + 1] : null;
  const isTerminal = lead.stage === 'contracted' || lead.stage === 'lost';

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -4 }}
      onClick={onClick}
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        borderRadius: '10px',
        padding: '12px',
        cursor: 'pointer',
        transition: 'border-color 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-text-dim)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          justifyContent: 'space-between',
          gap: '8px',
        }}
      >
        <div style={{ flex: 1, minWidth: 0 }}>
          <p
            style={{
              fontSize: '13px',
              fontWeight: 600,
              color: 'var(--hub-text)',
              margin: 0,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          >
            {lead.companyName ?? lead.company ?? '—'}
          </p>
          <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)', margin: '2px 0 0' }}>
            {lead.industry}
          </p>
          {followUp && followUp !== 'upcoming' && (
            <p
              style={{
                fontSize: '11px',
                fontWeight: 600,
                margin: '4px 0 0',
                color: followUp === 'overdue' ? '#ef4444' : '#f59e0b',
              }}
            >
              {t(`crm.leads.followUp.${followUp}`)}
            </p>
          )}
        </div>
        {repName && (
          <div
            title={repName}
            style={{
              width: '24px',
              height: '24px',
              borderRadius: '6px',
              background: 'rgba(59,130,246,0.15)',
              color: 'var(--accent-fg)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '10px',
              fontWeight: 700,
              flexShrink: 0,
            }}
          >
            {initials(repName)}
          </div>
        )}
      </div>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginTop: '10px',
        }}
      >
        <span
          style={{
            fontSize: '10px',
            fontWeight: 600,
            padding: '2px 7px',
            borderRadius: '4px',
            background: 'var(--hub-hover)',
            color: 'var(--hub-text-muted)',
          }}
        >
          {lead.source}
        </span>
        <span style={{ fontSize: '12px', fontWeight: 600, color: 'var(--hub-text)' }}>
          ${(lead.mrrPotential ?? lead.dealValue ?? 0).toLocaleString()}/mo
        </span>
      </div>
      {!isTerminal && onMoveStage && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            gap: '6px',
            marginTop: '10px',
            paddingTop: '8px',
            borderTop: '1px solid var(--hub-border)',
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <button
            type="button"
            disabled={isUpdating}
            onClick={(e) => {
              e.stopPropagation();
              onMoveStage(lead, 'lost');
            }}
            title={t('crm.pipeline.markLost')}
            aria-label={t('crm.pipeline.markLost')}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '4px',
              padding: '3px 7px',
              fontSize: '11px',
              borderRadius: '5px',
              background: 'transparent',
              border: '1px solid var(--hub-border)',
              color: 'var(--crm-sales-lost, #ef4444)',
              cursor: 'pointer',
            }}
          >
            <X size={11} />
            <span>{t('crm.leadStages.lost')}</span>
          </button>
          {nextStage && (
            <button
              type="button"
              disabled={isUpdating}
              onClick={(e) => {
                e.stopPropagation();
                onMoveStage(lead, nextStage);
              }}
              title={`${t('crm.pipeline.nextStage')}: ${t(`crm.leadStages.${nextStage}`)}`}
              aria-label={`${t('crm.pipeline.nextStage')}: ${t(`crm.leadStages.${nextStage}`)}`}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '4px',
                padding: '3px 8px',
                fontSize: '11px',
                fontWeight: 600,
                borderRadius: '5px',
                background: 'var(--hub-hover)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
                cursor: 'pointer',
              }}
            >
              <span>{t(`crm.leadStages.${nextStage}`)}</span>
              <ChevronRight size={11} />
            </button>
          )}
        </div>
      )}
    </motion.div>
  );
};

const Pipeline: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const pipelineQuery = usePipeline();
  const { data: pipelineCols, refetch: refetchPipeline } = pipelineQuery;
  const pipelineError = hasQueryFailed(pipelineQuery);
  const pipelineLoading = pipelineQuery.isPending && !pipelineError;
  const [leadOffset, setLeadOffset] = useState(0);
  const [exhausted, setExhausted] = useState(false);
  const [loadedLeads, setLoadedLeads] = useState<Lead[]>([]);
  const leadsQuery = useLeads(LEADS_PAGE_SIZE, leadOffset);
  const { data: fetchedLeads, isFetching: leadsFetching, refetch: refetchLeads } = leadsQuery;
  const leadsError = hasQueryFailed(leadsQuery);
  const leadsLoading = leadsQuery.isPending && !leadsError;
  const [createOpen, setCreateOpen] = useState(false);
  const updateLead = useUpdateLead();
  const { data: reps = [] } = useSalesReps();
  const repNames = useMemo(() => new Map(reps.map((rep) => [rep.id, rep.name])), [reps]);

  const handleMoveStage = (lead: Lead, nextStage: LeadStage): void => {
    setLoadedLeads((prev) => prev.map((l) => (l.id === lead.id ? { ...l, stage: nextStage } : l)));
    updateLead.mutate(
      {
        id: lead.id,
        data: {
          stage: nextStage,
          companyName: lead.companyName ?? lead.company,
        },
      },
      {
        onError: () => {
          setLoadedLeads((prev) =>
            prev.map((l) => (l.id === lead.id ? { ...l, stage: lead.stage } : l))
          );
        },
      }
    );
  };

  useEffect(() => {
    if (!fetchedLeads) return;
    if (leadOffset > 0 && fetchedLeads.length === 0) {
      setExhausted(true);
    }
    setLoadedLeads((previous) => {
      if (leadOffset === 0) return fetchedLeads;
      const byId = new Map(previous.map((lead) => [lead.id, lead]));
      fetchedLeads.forEach((lead) => byId.set(lead.id, lead));
      return Array.from(byId.values());
    });
  }, [fetchedLeads, leadOffset]);

  const cutoff = useMemo(pipelineCutoff, []);
  const leadsByStage = useMemo((): Record<LeadStage, Lead[]> => {
    const map = {} as Record<LeadStage, Lead[]>;
    STAGES.forEach((key) => {
      map[key] = loadedLeads.filter(
        (lead) => lead.stage === key && isInPipelineWindow(lead, cutoff)
      );
    });
    return map;
  }, [cutoff, loadedLeads]);

  const colSummary = useMemo((): Record<LeadStage, { count: number; mrr: number }> => {
    const result = {} as Record<LeadStage, { count: number; mrr: number }>;
    STAGES.forEach((key) => {
      const fromApi = (pipelineCols ?? []).find((column) => column.stage === key);
      result[key] = fromApi
        ? { count: fromApi.count, mrr: fromApi.totalMrr }
        : { count: 0, mrr: 0 };
    });
    return result;
  }, [pipelineCols]);

  const isInitialLoading = pipelineLoading || (leadsLoading && loadedLeads.length === 0);
  const queryFailed = pipelineError || leadsError;
  const hasMore = !exhausted && fetchedLeads?.length === LEADS_PAGE_SIZE;
  const retryQueries = (): void => {
    void Promise.all([refetchPipeline(), refetchLeads()]);
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.pipeline.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', minWidth: 0 }}>
        <PageHeader
          title={t('crm.pipeline.title')}
          subtitle={t('crm.pipeline.subtitle')}
          actions={
            <Button variant="primary" onClick={() => setCreateOpen(true)}>
              <Plus size={13} />
              {t('crm.pipeline.addLead')}
            </Button>
          }
        />

        {queryFailed ? (
          <EmptyState
            icon={AlertCircle}
            title={t('crm.leads.loadFailedTitle')}
            description={t('crmErrors.refreshAndTryAgain')}
            action={{ label: t('crm.actions.retry'), onClick: retryQueries }}
          />
        ) : isInitialLoading ? (
          <div style={{ padding: '16px' }}>
            <SkeletonList count={4} />
          </div>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(7, minmax(200px, 1fr))',
              gap: '12px',
              overflowX: 'auto',
              paddingBottom: '8px',
            }}
          >
            {STAGES.map((key) => {
              const label = t(`crm.leadStages.${key}`);
              const summary = colSummary[key];
              const columnLeads = leadsByStage[key] ?? [];
              const stageColor = STAGE_COLORS[key];
              return (
                <div
                  key={key}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '8px',
                    minWidth: '200px',
                  }}
                >
                  {/* Column header */}
                  <div
                    style={{
                      padding: '10px 12px',
                      background: 'var(--hub-card)',
                      border: '1px solid var(--hub-border)',
                      borderRadius: '10px',
                      borderTop: `3px solid ${stageColor}`,
                    }}
                  >
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        marginBottom: '4px',
                      }}
                    >
                      <p
                        style={{
                          fontSize: '12px',
                          fontWeight: 600,
                          color: 'var(--hub-text)',
                          margin: 0,
                        }}
                      >
                        {label}
                      </p>
                      <span
                        style={{
                          fontSize: '11px',
                          fontWeight: 700,
                          padding: '1px 6px',
                          borderRadius: '4px',
                          background: 'var(--hub-hover)',
                          color: 'var(--hub-text-muted)',
                        }}
                      >
                        {summary.count}
                      </span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                      <TrendingUp size={10} style={{ color: stageColor }} />
                      <span style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
                        ${summary.mrr.toLocaleString()}/mo
                      </span>
                    </div>
                  </div>

                  {/* Lead cards */}
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', flex: 1 }}>
                    {columnLeads.map((lead) => (
                      <LeadCard
                        key={lead.id}
                        lead={lead}
                        repName={
                          lead.assignedTo != null ? repNames.get(lead.assignedTo) : undefined
                        }
                        onClick={() => navigate({ to: `/crm/sales/leads/${lead.id}` })}
                        onMoveStage={handleMoveStage}
                        isUpdating={updateLead.isPending}
                      />
                    ))}
                  </div>

                  {/* Add button - only on Cold column */}
                  {key === 'cold' && (
                    <button
                      type="button"
                      onClick={() => setCreateOpen(true)}
                      aria-label={t('crm.pipeline.addNewLead')}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '6px',
                        padding: '8px',
                        borderRadius: '8px',
                        background: 'transparent',
                        border: '1px dashed var(--hub-border)',
                        cursor: 'pointer',
                        color: 'var(--hub-text-dim)',
                        fontSize: '12px',
                        transition: 'all 0.15s',
                      }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                        (e.currentTarget as HTMLElement).style.color = 'var(--hub-text)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'transparent';
                        (e.currentTarget as HTMLElement).style.color = 'var(--hub-text-dim)';
                      }}
                    >
                      <Plus size={12} />
                      {t('crm.pipeline.addLead')}
                    </button>
                  )}
                </div>
              );
            })}
            {leadsFetching && loadedLeads.length > 0 && (
              <p
                role="status"
                style={{
                  gridColumn: '1 / -1',
                  margin: 0,
                  textAlign: 'center',
                  color: 'var(--hub-text-muted)',
                }}
              >
                {t('crmCommon.loading')}
              </p>
            )}
            {hasMore && (
              <div style={{ gridColumn: '1 / -1', display: 'flex', justifyContent: 'center' }}>
                <Button
                  variant="outline"
                  onClick={() => setLeadOffset((current) => current + LEADS_PAGE_SIZE)}
                  disabled={leadsFetching}
                >
                  {leadsFetching ? t('crmCommon.loading') : t('crm.pipeline.loadMore')}
                </Button>
              </div>
            )}
          </div>
        )}

        <LeadFormModal open={createOpen} onClose={() => setCreateOpen(false)} />
      </div>
    </>
  );
};

export default Pipeline;
