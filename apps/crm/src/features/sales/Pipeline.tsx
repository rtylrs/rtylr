import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Plus, TrendingUp } from 'lucide-react';
import { motion } from 'motion/react';
import { usePipeline, useLeads } from '../../shared/hooks/queries/useSales';
import type { Lead, LeadStage } from '../../shared/hooks/queries/useSales';
import { Button, PageHeader, SkeletonList } from '../../shared/components/ui';
import CreateLeadModal from './components/CreateLeadModal';

const STAGES: Array<{ key: LeadStage; label: string }> = [
  { key: 'cold', label: 'Cold' },
  { key: 'contacted', label: 'Contacted' },
  { key: 'demo_booked', label: 'Demo Booked' },
  { key: 'demo_done', label: 'Demo Done' },
  { key: 'proposal_sent', label: 'Proposal Sent' },
  { key: 'contracted', label: 'Contracted' },
  { key: 'lost', label: 'Lost' },
];

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
  onClick: () => void;
}

const LeadCard: React.FC<LeadCardProps> = ({ lead, onClick }) => (
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
      </div>
      <div
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
        {initials(lead.assignedTo != null ? String(lead.assignedTo) : undefined)}
      </div>
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
  </motion.div>
);

const Pipeline: React.FC = () => {
  const navigate = useNavigate();
  const { data: pipelineCols, isLoading: pipelineLoading } = usePipeline();
  const { data: leads, isLoading: leadsLoading } = useLeads();
  const [createOpen, setCreateOpen] = useState(false);

  const isLoading = pipelineLoading || leadsLoading;

  const leadsByStage = useMemo((): Record<LeadStage, Lead[]> => {
    const all = leads ?? [];
    const map = {} as Record<LeadStage, Lead[]>;
    STAGES.forEach(({ key }) => {
      map[key] = all.filter((l) => l.stage === key);
    });
    return map;
  }, [leads]);

  const colSummary = useMemo((): Record<LeadStage, { count: number; mrr: number }> => {
    const byStage = leadsByStage;
    const result = {} as Record<LeadStage, { count: number; mrr: number }>;
    STAGES.forEach(({ key }) => {
      const fromApi = (pipelineCols ?? []).find((c) => c.stage === key);
      if (fromApi) {
        result[key] = { count: fromApi.count, mrr: fromApi.totalMrr };
      } else {
        const stageLeads = byStage[key] ?? [];
        result[key] = {
          count: stageLeads.length,
          mrr: stageLeads.reduce((s, l) => s + (l.mrrPotential ?? l.dealValue ?? 0), 0),
        };
      }
    });
    return result;
  }, [pipelineCols, leadsByStage]);

  return (
    <>
      <HelmetContainer title="Sales pipeline · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', minWidth: 0 }}>
        <PageHeader
          title="Sales Pipeline"
          subtitle="Track deals from cold outreach to closed contracts"
          actions={
            <Button variant="primary" onClick={() => setCreateOpen(true)}>
              <Plus size={13} />
              Add lead
            </Button>
          }
        />

        {isLoading ? (
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
            {STAGES.map(({ key, label }) => {
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
                        onClick={() => navigate({ to: `/sales/leads/${lead.id}` })}
                      />
                    ))}
                  </div>

                  {/* Add button - only on Cold column */}
                  {key === 'cold' && (
                    <button
                      type="button"
                      onClick={() => setCreateOpen(true)}
                      aria-label="Add new lead"
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
                      Add lead
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        )}

        <CreateLeadModal open={createOpen} onClose={() => setCreateOpen(false)} />
      </div>
    </>
  );
};

export default Pipeline;
