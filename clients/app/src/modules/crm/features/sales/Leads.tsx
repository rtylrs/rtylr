import React, { useEffect, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { AlertCircle, Plus, Search, Trash2, UserPlus, Eye } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { useLeads, useDeleteLead, useSalesReps } from '../../shared/hooks/queries/useSales';
import type { Lead, LeadStage } from '../../shared/hooks/queries/useSales';
import { Button, PageHeader, SkeletonList, ConfirmModal } from '../../shared/components/ui';
import LeadFormModal from './components/LeadFormModal';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';
import { followUpStatus, formatDay, type FollowUpStatus } from '../../shared/lib/dateUtils';

const STAGE_COLORS: Record<LeadStage, { bg: string; color: string }> = {
  cold: { bg: 'rgba(107,114,128,0.12)', color: '#6b7280' },
  contacted: { bg: 'rgba(59,130,246,0.12)', color: 'var(--crm-accent)' },
  demo_booked: { bg: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
  demo_done: { bg: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
  proposal_sent: { bg: 'rgba(249,115,22,0.12)', color: '#f97316' },
  contracted: { bg: 'rgba(16,185,129,0.12)', color: '#10b981' },
  lost: { bg: 'rgba(239,68,68,0.12)', color: '#ef4444' },
};

const ALL_STAGES: Array<LeadStage | 'all'> = [
  'all',
  'cold',
  'contacted',
  'demo_booked',
  'demo_done',
  'proposal_sent',
  'contracted',
  'lost',
];

const StageBadge: React.FC<{ stage: LeadStage }> = ({ stage }) => {
  const { t } = useTranslation();
  const colors = STAGE_COLORS[stage];
  return (
    <span
      style={{
        fontSize: '11px',
        fontWeight: 600,
        padding: '2px 8px',
        borderRadius: '6px',
        background: colors.bg,
        color: colors.color,
        whiteSpace: 'nowrap',
      }}
    >
      {t(`crm.leadStages.${stage}`)}
    </span>
  );
};

const FOLLOW_UP_COLORS: Record<FollowUpStatus, string> = {
  overdue: '#ef4444',
  today: '#f59e0b',
  upcoming: 'var(--hub-text-muted)',
};

const FollowUpCell: React.FC<{ date?: string }> = ({ date }) => {
  const { t, i18n } = useTranslation();
  if (!date) return <span style={{ color: 'var(--hub-text-dim)' }}>—</span>;
  const status = followUpStatus(date);
  return (
    <span
      style={{ color: FOLLOW_UP_COLORS[status], fontWeight: status === 'upcoming' ? 400 : 600 }}
    >
      {status === 'upcoming' ? formatDay(date, i18n.language) : t(`crm.leads.followUp.${status}`)}
    </span>
  );
};

const Leads: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const pageSize = 100;
  const [offset, setOffset] = useState(0);
  // Set once a "Load more" page comes back empty.
  const [exhausted, setExhausted] = useState(false);
  const [loadedLeads, setLoadedLeads] = useState<Lead[]>([]);
  const leadsQuery = useLeads(pageSize, offset);
  const { data: fetchedLeads, isFetching, refetch } = leadsQuery;
  const isError = hasQueryFailed(leadsQuery);
  const isLoading = leadsQuery.isPending && !isError;
  // Keep the first page mounted as a write sentinel. The create modal lives in
  // a child component, so this is how the list notices its cache invalidation
  // and rebuilds from page zero instead of merging a stale later page.
  const firstPageQuery = useLeads(pageSize, 0);
  const firstPageFailed = hasQueryFailed(firstPageQuery);
  const firstPageRevisionRef = useRef<number | string | null>(null);
  const deleteLead = useDeleteLead();

  const resetPagination = (): void => {
    setOffset(0);
    setExhausted(false);
    setLoadedLeads([]);
  };

  useEffect(() => {
    if (!fetchedLeads) return;
    if (offset > 0 && fetchedLeads.length === 0) {
      // Exact page-size multiple: the extra page came back empty — stop
      // offering "Load more".
      setExhausted(true);
    }
    setLoadedLeads((previous) => {
      if (offset === 0) return fetchedLeads;
      const byId = new Map(previous.map((lead) => [lead.id, lead]));
      fetchedLeads.forEach((lead) => byId.set(lead.id, lead));
      return Array.from(byId.values());
    });
  }, [fetchedLeads, offset]);

  const firstPageLeads = firstPageQuery.data;
  const firstPageSignature = firstPageLeads?.map((lead) => lead.id).join(',') ?? null;
  const firstPageRevision = firstPageQuery.dataUpdatedAt ?? firstPageSignature;
  useEffect(() => {
    if (!firstPageLeads || firstPageRevision === null || firstPageFailed) return;
    const previousRevision = firstPageRevisionRef.current;
    firstPageRevisionRef.current = firstPageRevision;
    if (offset > 0 && previousRevision !== null && previousRevision !== firstPageRevision) {
      resetPagination();
      setLoadedLeads(firstPageLeads);
    }
  }, [firstPageFailed, firstPageLeads, firstPageRevision, offset]);

  const leads = loadedLeads;
  const hasMore = !exhausted && !isError && fetchedLeads?.length === pageSize;
  const [search, setSearch] = useState('');
  const [stageFilter, setStageFilter] = useState<LeadStage | 'all'>('all');
  const [sourceFilter, setSourceFilter] = useState('all');
  const [dueOnly, setDueOnly] = useState(false);
  const { data: reps = [] } = useSalesReps();
  const repNames = new Map(reps.map((rep) => [rep.id, rep.name]));
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const sources = Array.from(new Set(leads.map((l) => l.source).filter(Boolean)));

  const filtered = leads.filter((l) => {
    const matchSearch = (l.companyName ?? l.company ?? '')
      .toLowerCase()
      .includes(search.toLowerCase());
    const matchStage = stageFilter === 'all' || l.stage === stageFilter;
    const matchSource = sourceFilter === 'all' || l.source === sourceFilter;
    const matchDue =
      !dueOnly || (!!l.nextFollowUp && followUpStatus(l.nextFollowUp) !== 'upcoming');
    return matchSearch && matchStage && matchSource && matchDue;
  });

  const deleteTarget = deleteId !== null ? leads.find((l) => l.id === deleteId) : null;

  const handleDelete = (): void => {
    if (deleteId === null) return;
    const deletedId = deleteId;
    deleteLead.mutate(deletedId, {
      onSuccess: () => {
        setLoadedLeads((previous) => previous.filter((lead) => lead.id !== deletedId));
        setOffset(0);
        setExhausted(false);
        toast.success(t('crm.leads.deleted'));
        setDeleteId(null);
      },
      onError: () => toast.error(t('crm.leadDetail.deleteFailed')),
    });
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.leads.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1280px' }}>
        <PageHeader
          title={t('crm.leads.pageTitle')}
          subtitle={`${leads.length} loaded`}
          actions={
            <Button variant="primary" onClick={() => setCreateOpen(true)}>
              <Plus size={13} />
              {t('crm.leads.new')}
            </Button>
          }
        />

        {/* Filters */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
          {/* Search */}
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              padding: '0 12px',
              height: '36px',
              borderRadius: '8px',
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              flex: '1',
              minWidth: '200px',
              maxWidth: '300px',
            }}
          >
            <Search size={14} style={{ color: 'var(--hub-text-muted)', flexShrink: 0 }} />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('crm.leads.searchPlaceholder')}
              style={{
                flex: 1,
                background: 'transparent',
                border: 'none',
                outline: 'none',
                fontSize: '13px',
                color: 'var(--hub-text)',
              }}
            />
          </div>

          {/* Stage filter */}
          <select
            value={stageFilter}
            onChange={(e) => setStageFilter(e.target.value as LeadStage | 'all')}
            style={{
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
            {ALL_STAGES.map((s) => (
              <option key={s} value={s}>
                {s === 'all' ? t('crm.leads.allStages') : t(`crm.leadStages.${s}`)}
              </option>
            ))}
          </select>

          {/* Source filter */}
          <select
            value={sourceFilter}
            onChange={(e) => setSourceFilter(e.target.value)}
            style={{
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
            <option value="all">{t('crm.leads.allSources')}</option>
            {sources.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>

          {/* Follow-ups due today or overdue */}
          <label
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              fontSize: '13px',
              color: 'var(--hub-text)',
              cursor: 'pointer',
            }}
          >
            <input
              type="checkbox"
              checked={dueOnly}
              onChange={(e) => setDueOnly(e.target.checked)}
            />
            {t('crm.leads.followUpsDue')}
          </label>
        </div>

        {/* Table */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.2, ease: 'easeOut' }}
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            borderRadius: '12px',
            overflow: 'hidden',
          }}
        >
          {isLoading && leads.length === 0 ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={5} />
            </div>
          ) : isError && leads.length === 0 ? (
            <EmptyState
              icon={AlertCircle}
              title={t('crm.leads.loadFailedTitle')}
              description={t('crmErrors.refreshAndTryAgain')}
              action={{ label: t('crm.actions.retry'), onClick: () => refetch() }}
            />
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={UserPlus}
              title={t('crm.leads.empty')}
              description={
                search
                  ? t('crm.leads.emptySearchDescription', { search })
                  : t('crm.leads.createFirstDescription')
              }
              action={
                search
                  ? { label: t('crm.actions.clearSearch'), onClick: () => setSearch('') }
                  : { label: t('crm.leads.new'), onClick: () => setCreateOpen(true) }
              }
            />
          ) : (
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  {[
                    t('crm.leads.columns.company'),
                    t('crm.leads.columns.industry'),
                    t('crm.leads.columns.stage'),
                    t('crm.leads.columns.source'),
                    t('crm.leads.columns.rep'),
                    t('crm.leads.columns.nextFollowUp'),
                    t('crm.leads.columns.value'),
                    t('crm.leads.columns.actions'),
                  ].map((h, i) => (
                    <th
                      key={i}
                      style={{
                        padding: '11px 16px',
                        textAlign: 'left',
                        fontSize: '11px',
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        letterSpacing: '0.05em',
                        color: 'var(--hub-text-dim)',
                        whiteSpace: 'nowrap',
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                <AnimatePresence>
                  {filtered.map((lead: Lead) => (
                    <motion.tr
                      key={lead.id}
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      style={{
                        borderBottom: '1px solid var(--hub-border)',
                        cursor: 'pointer',
                        transition: 'background 0.12s',
                      }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.background = 'transparent';
                      }}
                      onClick={() => navigate({ to: `/crm/sales/leads/${lead.id}` })}
                    >
                      <td style={{ padding: '12px 16px' }}>
                        <p
                          style={{
                            fontSize: '13px',
                            fontWeight: 600,
                            color: 'var(--hub-text)',
                            margin: 0,
                          }}
                        >
                          {lead.companyName ?? lead.company ?? '—'}
                        </p>
                      </td>
                      <td
                        style={{
                          padding: '12px 16px',
                          fontSize: '12px',
                          color: 'var(--hub-text-muted)',
                        }}
                      >
                        {lead.industry}
                      </td>
                      <td style={{ padding: '12px 16px' }}>
                        <StageBadge stage={lead.stage} />
                      </td>
                      <td
                        style={{
                          padding: '12px 16px',
                          fontSize: '12px',
                          color: 'var(--hub-text-muted)',
                        }}
                      >
                        {lead.source}
                      </td>
                      <td
                        style={{ padding: '12px 16px', fontSize: '12px', color: 'var(--hub-text)' }}
                      >
                        {lead.assignedTo != null
                          ? (repNames.get(lead.assignedTo) ?? t('crm.leads.formerRep'))
                          : '—'}
                      </td>
                      <td style={{ padding: '12px 16px', fontSize: '12px' }}>
                        <FollowUpCell date={lead.nextFollowUp} />
                      </td>
                      <td
                        style={{
                          padding: '12px 16px',
                          fontSize: '13px',
                          fontWeight: 600,
                          color: 'var(--hub-text)',
                        }}
                      >
                        ${(lead.mrrPotential ?? lead.dealValue ?? 0).toLocaleString()}
                      </td>
                      <td style={{ padding: '12px 16px' }} onClick={(e) => e.stopPropagation()}>
                        <div style={{ display: 'flex', gap: '6px' }}>
                          <button
                            type="button"
                            aria-label={t('crm.leads.viewLead')}
                            onClick={() => navigate({ to: `/crm/sales/leads/${lead.id}` })}
                            style={{
                              height: '28px',
                              width: '28px',
                              borderRadius: '6px',
                              background: 'transparent',
                              border: '1px solid var(--hub-border)',
                              cursor: 'pointer',
                              color: 'var(--hub-text-muted)',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                            }}
                          >
                            <Eye size={13} />
                          </button>
                          <button
                            type="button"
                            aria-label={t('crm.leadDetail.deleteLead')}
                            onClick={() => setDeleteId(lead.id)}
                            style={{
                              height: '28px',
                              width: '28px',
                              borderRadius: '6px',
                              background: 'transparent',
                              border: '1px solid var(--hub-border)',
                              cursor: 'pointer',
                              color: '#ef4444',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center',
                            }}
                          >
                            <Trash2 size={13} />
                          </button>
                        </div>
                      </td>
                    </motion.tr>
                  ))}
                </AnimatePresence>
              </tbody>
            </table>
          )}
          {isError && leads.length > 0 && (
            <div
              role="alert"
              style={{
                padding: '12px 16px',
                borderTop: '1px solid var(--hub-border)',
                color: '#ef4444',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
              }}
            >
              <AlertCircle size={15} />
              <span style={{ flex: 1, fontSize: '12px' }}>{t('crm.leads.loadFailedTitle')}</span>
              <Button variant="outline" onClick={() => void refetch()} disabled={isFetching}>
                {t('crm.actions.retry')}
              </Button>
            </div>
          )}
          {hasMore && (
            <div style={{ display: 'flex', justifyContent: 'center', padding: '16px' }}>
              <Button
                variant="outline"
                onClick={() => setOffset((current) => current + pageSize)}
                disabled={isFetching}
              >
                {isFetching ? 'Loading…' : 'Load more leads'}
              </Button>
            </div>
          )}
        </motion.div>

        <LeadFormModal open={createOpen} onClose={() => setCreateOpen(false)} />

        <ConfirmModal
          open={deleteId !== null}
          onClose={() => setDeleteId(null)}
          onConfirm={handleDelete}
          title={t('crm.leadDetail.deleteLead')}
          description={
            deleteTarget
              ? `Delete "${deleteTarget.companyName ?? deleteTarget.company ?? 'this lead'}"? This will also remove all associated touchpoints and contacts.`
              : 'Are you sure?'
          }
          confirmLabel={t('crmCommon.delete')}
          variant="danger"
        />
      </div>
    </>
  );
};

export default Leads;
