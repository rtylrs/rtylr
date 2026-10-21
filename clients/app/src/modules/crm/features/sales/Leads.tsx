import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Contact, Phone, Plus, Upload, X } from 'lucide-react';
import { toast } from 'sonner';
import { Button, ConfirmModal, Modal } from '../../shared/components/ui';
import {
  dateOf,
  formatRelative,
  LEAD_SORTS,
  LEAD_STATUSES,
  type Lead,
  type LeadStatus,
  leadTitle,
  personName,
  telHref,
  WORKING_STATUSES,
} from './data/model';
import { activeLeadFilterGroups, NO_LEAD_FILTERS, useLeadFilters } from './data/leadFilters';
import { useDeleteLeads, useLeadList, useLeadStats, useReps, useUpdateLeads } from './data/queries';
import {
  DueDate,
  EmptyBlock,
  LeadStatusBadge,
  ListFooter,
  LoadFailed,
  RepName,
  SearchBox,
  StatTile,
  useRepNames,
} from './components/common';
import { LeadFilterChips, LeadFilterPopover } from './components/LeadFilters';
import LeadFormModal from './components/LeadFormModal';
import { useVirtualRows } from '../../../../shared/hooks/useVirtualRows';
import { VirtualSpacerRow } from '../../../../shared/components/VirtualList';
import { ReasonPicker } from './components/ReasonPicker';

type Tab = 'working' | 'all' | LeadStatus;
const TABS: Tab[] = ['working', ...LEAD_STATUSES, 'all'];

function tabOf(statuses: LeadStatus[]): Tab {
  if (statuses.length === 0) return 'all';
  if (statuses.length === 1) return statuses[0];
  const working = WORKING_STATUSES.every((status) => statuses.includes(status));
  return working && statuses.length === WORKING_STATUSES.length ? 'working' : 'all';
}

function statusesOf(tab: Tab): LeadStatus[] {
  if (tab === 'all') return [];
  if (tab === 'working') return WORKING_STATUSES;
  return [tab];
}

const BULK_REASONS = [
  'notAFit',
  'noBudget',
  'noNeed',
  'unreachable',
  'wrongContact',
  'competitor',
  'doNotContact',
] as const;

const Leads: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { filters, setFilters } = useLeadFilters();
  const leads = useLeadList(filters);
  const stats = useLeadStats(filters);
  const { data: reps = [] } = useReps();
  const names = useRepNames();
  const updateLeads = useUpdateLeads();
  const deleteLeads = useDeleteLeads();
  const [creating, setCreating] = useState(false);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [disqualifying, setDisqualifying] = useState(false);
  const [reason, setReason] = useState('');

  const rows = useMemo(
    () => leads.data?.pages.flatMap((page) => page.leads ?? []) ?? [],
    [leads.data]
  );
  const total = leads.data?.pages[0]?.total;
  const virtual = useVirtualRows(rows.length, { estimateSize: 53, headerRows: 1 });
  const counts = useMemo(() => {
    const byStatus = new Map(
      (stats.data?.byStatus ?? []).map((item) => [item.status, item.count ?? 0])
    );
    const count = (tab: Tab) =>
      tab === 'all'
        ? [...byStatus.values()].reduce((sum, value) => sum + value, 0)
        : statusesOf(tab).reduce((sum, status) => sum + (byStatus.get(status) ?? 0), 0);
    return count;
  }, [stats.data]);
  const format = (value?: number) => new Intl.NumberFormat(i18n.language).format(value ?? 0);
  const activeTab = tabOf(filters.statuses);
  const filtered = filters.search !== '' || activeLeadFilterGroups(filters).length > 0;

  const allSelected = rows.length > 0 && rows.every((lead) => selected.has(lead.id ?? ''));
  const toggle = (id: string) =>
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  const clearSelection = () => setSelected(new Set());
  const ids = [...selected];
  const bulk = (change: Parameters<typeof updateLeads.mutate>[0]) =>
    updateLeads.mutate(change, {
      onSuccess: (result) => {
        clearSelection();
        setDisqualifying(false);
        toast.success(t('crmSales.leads.bulkUpdated', { count: result.updated ?? 0 }));
      },
    });

  const open = (lead: Lead) => navigate({ to: `/crm/sales/leads/${lead.id}` });

  return (
    <div className="sales-page">
      <HelmetContainer title={`${t('crmSales.leads.title')} · rtylr`} />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.leads.title')}</h1>
          <p>{t('crmSales.leads.subtitle')}</p>
        </div>
        <div className="sales-actions">
          <Button
            variant="outline"
            onClick={() => navigate({ to: '/crm/sales/leads/call' })}
            disabled={!total}
          >
            <Phone size={13} />
            {t('crmSales.leads.startCalling')}
          </Button>
          <Button variant="outline" onClick={() => navigate({ to: '/crm/sales/leads/import' })}>
            <Upload size={13} />
            {t('crmSales.leads.import')}
          </Button>
          <Button onClick={() => setCreating(true)}>
            <Plus size={13} />
            {t('crmSales.leads.add')}
          </Button>
        </div>
      </header>

      <div className="sales-stats">
        <StatTile
          label={t('crmSales.leads.stats.followUpsDue')}
          value={format((stats.data?.followUpsOverdue ?? 0) + (stats.data?.followUpsToday ?? 0))}
          hint={t('crmSales.leads.stats.overdueHint', { count: stats.data?.followUpsOverdue ?? 0 })}
          warning={(stats.data?.followUpsOverdue ?? 0) > 0}
          onClick={() => {
            setFilters({ statuses: WORKING_STATUSES, followUp: 'due' });
            clearSelection();
          }}
        />
        <StatTile
          label={t('crmSales.leads.stats.callsToday')}
          value={format(stats.data?.callsToday)}
          hint={t('crmSales.leads.stats.connectedHint', { count: stats.data?.connectedToday ?? 0 })}
        />
        <StatTile
          label={t('crmSales.leads.stats.working')}
          value={format(counts('working'))}
          onClick={() => {
            setFilters({ statuses: WORKING_STATUSES });
            clearSelection();
          }}
        />
        <StatTile
          label={t('crmSales.leads.stats.convertedThisMonth')}
          value={format(stats.data?.convertedThisMonth)}
        />
      </div>

      <div className="sales-listbar">
        <div className="sales-listbar-grid">
          <div className="sales-tabs" role="tablist" aria-label={t('crmSales.leads.statusFilter')}>
            {TABS.map((tab) => (
              <button
                key={tab}
                type="button"
                role="tab"
                className="sales-tab"
                aria-selected={activeTab === tab}
                onClick={() => {
                  setFilters({ statuses: statusesOf(tab) });
                  clearSelection();
                }}
              >
                {tab === 'working' || tab === 'all'
                  ? t(`crmSales.leads.tabs.${tab}`)
                  : t(`crmSales.leadStatus.${tab}`)}
                <span className="sales-tab-count">{format(counts(tab))}</span>
              </button>
            ))}
          </div>
          <div className="sales-toolbar-actions">
            <SearchBox
              value={filters.search}
              onChange={(search) => setFilters({ search })}
              placeholder={t('crmSales.leads.searchPlaceholder')}
            />
            <LeadFilterPopover
              filters={filters}
              onChange={(change) => {
                setFilters(change);
                clearSelection();
              }}
              total={total}
            />
            <select
              className="sales-select"
              aria-label={t('crmSales.common.sortBy')}
              value={filters.sort}
              onChange={(event) => setFilters({ sort: event.target.value as typeof filters.sort })}
            >
              {LEAD_SORTS.map((sort) => (
                <option key={sort} value={sort}>
                  {t(`crmSales.leads.sort.${sort}`)}
                </option>
              ))}
            </select>
          </div>
          {activeLeadFilterGroups(filters).length > 0 && (
            <div className="sales-listbar-chips">
              <LeadFilterChips
                filters={filters}
                onChange={(change) => {
                  setFilters(change);
                  clearSelection();
                }}
              />
              <Button
                variant="ghost"
                onClick={() => {
                  setFilters({ ...NO_LEAD_FILTERS, search: '' });
                  clearSelection();
                }}
              >
                <X size={13} />
                {t('crmSales.common.clearFilters')}
              </Button>
            </div>
          )}
        </div>
      </div>

      <div className="sales-card">
        {leads.isError && rows.length === 0 ? (
          <LoadFailed title={t('crmSales.leads.loadFailed')} onRetry={() => void leads.refetch()} />
        ) : leads.isPending ? (
          <p className="sales-empty" role="status">
            {t('crmSales.common.loading')}
          </p>
        ) : rows.length === 0 ? (
          filtered || activeTab !== 'working' ? (
            <EmptyBlock
              icon={Contact}
              title={t('crmSales.leads.noMatches')}
              description={t('crmSales.leads.noMatchesHint')}
            />
          ) : (
            <EmptyBlock
              icon={Contact}
              title={t('crmSales.leads.emptyTitle')}
              description={t('crmSales.leads.emptyHint')}
            >
              <Button variant="outline" onClick={() => navigate({ to: '/crm/sales/leads/import' })}>
                <Upload size={13} />
                {t('crmSales.leads.import')}
              </Button>
              <Button onClick={() => setCreating(true)}>
                <Plus size={13} />
                {t('crmSales.leads.add')}
              </Button>
            </EmptyBlock>
          )
        ) : (
          <>
            <div className="sales-table-wrap">
              <table className="sales-table">
                <thead>
                  <tr>
                    <th className="is-check">
                      <input
                        type="checkbox"
                        aria-label={t('crmSales.common.selectAll')}
                        checked={allSelected}
                        onChange={() =>
                          setSelected(
                            allSelected ? new Set() : new Set(rows.map((lead) => lead.id ?? ''))
                          )
                        }
                      />
                    </th>
                    <th>{t('crmSales.fields.lead')}</th>
                    <th>{t('crmSales.fields.phone')}</th>
                    <th>{t('crmSales.fields.status')}</th>
                    <th>{t('crmSales.fields.rep')}</th>
                    <th>{t('crmSales.fields.lastActivity')}</th>
                    <th>{t('crmSales.fields.followUp')}</th>
                    <th>{t('crmSales.fields.source')}</th>
                  </tr>
                </thead>
                <tbody ref={virtual.listRef}>
                  <VirtualSpacerRow height={virtual.before} columns={8} />
                  {virtual.indexes.map((index) => {
                    const lead = rows[index];
                    const id = lead.id ?? '';
                    const person = personName(lead);
                    const href = telHref(lead.phone);
                    return (
                      <tr
                        key={id}
                        {...virtual.rowProps(index)}
                        className={selected.has(id) ? 'is-selected' : undefined}
                        onClick={() => open(lead)}
                      >
                        <td className="is-check" onClick={(event) => event.stopPropagation()}>
                          <input
                            type="checkbox"
                            aria-label={t('crmSales.common.selectRow', { name: leadTitle(lead) })}
                            checked={selected.has(id)}
                            onChange={() => toggle(id)}
                          />
                        </td>
                        <td>
                          <div className="sales-cell-main">{leadTitle(lead)}</div>
                          <div className="sales-cell-sub">
                            {lead.companyName && person
                              ? [person, lead.jobTitle].filter(Boolean).join(' · ')
                              : lead.jobTitle || lead.email || ''}
                          </div>
                        </td>
                        <td onClick={(event) => event.stopPropagation()}>
                          {href ? (
                            <a className="sales-link" href={href}>
                              <bdi dir="ltr">{lead.phone}</bdi>
                            </a>
                          ) : (
                            <span className="sales-muted">—</span>
                          )}
                        </td>
                        <td>
                          <LeadStatusBadge status={lead.status} />
                        </td>
                        <td>
                          <RepName id={lead.assignedTo} names={names} />
                        </td>
                        <td className="sales-muted">
                          {lead.lastActivityAt
                            ? formatRelative(dateOf(lead.lastActivityAt), i18n.language)
                            : t('crmSales.leads.neverWorked')}
                          {(lead.attemptCount ?? 0) > 0 && (
                            <div className="sales-cell-sub">
                              {t('crmSales.leads.attempts', { count: lead.attemptCount ?? 0 })}
                            </div>
                          )}
                        </td>
                        <td>
                          <DueDate date={lead.nextFollowUp} />
                        </td>
                        <td className="sales-muted">{lead.source || '—'}</td>
                      </tr>
                    );
                  })}
                  <VirtualSpacerRow height={virtual.after} columns={8} />
                </tbody>
              </table>
            </div>
            <ListFooter shown={rows.length} total={total} pages={leads} />
          </>
        )}
      </div>

      {selected.size > 0 && (
        <div className="sales-bulk" role="region" aria-label={t('crmSales.leads.bulkActions')}>
          <strong>{t('crmSales.common.selected', { count: selected.size })}</strong>
          <select
            className="sales-select"
            aria-label={t('crmSales.leads.assignTo')}
            value=""
            onChange={(event) => {
              const value = event.target.value;
              if (!value) return;
              bulk(value === 'unassigned' ? { ids, unassign: true } : { ids, assignTo: value });
            }}
          >
            <option value="">{t('crmSales.leads.assignTo')}</option>
            <option value="unassigned">{t('crmSales.team.unassigned')}</option>
            {reps.map((rep) => (
              <option key={rep.id} value={rep.id}>
                {rep.name}
              </option>
            ))}
          </select>
          <select
            className="sales-select"
            aria-label={t('crmSales.leads.setStatus')}
            value=""
            onChange={(event) => {
              const status = event.target.value as LeadStatus;
              if (!status) return;
              if (status === 'disqualified') {
                setReason('');
                setDisqualifying(true);
              } else {
                bulk({ ids, status });
              }
            }}
          >
            <option value="">{t('crmSales.leads.setStatus')}</option>
            {LEAD_STATUSES.filter((status) => status !== 'qualified').map((status) => (
              <option key={status} value={status}>
                {t(`crmSales.leadStatus.${status}`)}
              </option>
            ))}
          </select>
          <Button variant="danger" onClick={() => setConfirmDelete(true)}>
            {t('crmSales.common.delete')}
          </Button>
          <span className="sales-spacer" style={{ flex: 1 }} />
          <Button variant="ghost" onClick={clearSelection}>
            {t('crmSales.common.clearSelection')}
          </Button>
        </div>
      )}

      <LeadFormModal open={creating} onClose={() => setCreating(false)} onCreated={open} />
      <ConfirmModal
        open={confirmDelete}
        onClose={() => setConfirmDelete(false)}
        onConfirm={() =>
          deleteLeads.mutate(ids, {
            onSuccess: (result) => {
              clearSelection();
              toast.success(t('crmSales.leads.bulkDeleted', { count: result.deleted ?? 0 }));
            },
          })
        }
        title={t('crmSales.leads.deleteSelected', { count: selected.size })}
        description={t('crmSales.leads.deleteSelectedHint')}
        confirmLabel={t('crmSales.common.delete')}
        variant="danger"
      />
      <Modal
        open={disqualifying}
        onClose={() => setDisqualifying(false)}
        title={t('crmSales.disqualify.bulkTitle', { count: selected.size })}
        footer={
          <>
            <Button variant="outline" onClick={() => setDisqualifying(false)}>
              {t('crmSales.common.cancel')}
            </Button>
            <Button
              variant="danger"
              disabled={!reason.trim() || updateLeads.isPending}
              onClick={() => bulk({ ids, status: 'disqualified', disqualifyReason: reason.trim() })}
            >
              {t('crmSales.disqualify.action')}
            </Button>
          </>
        }
      >
        <ReasonPicker
          reasons={BULK_REASONS}
          labelPrefix="crmSales.disqualify.reasons"
          value={reason}
          onChange={setReason}
          label={t('crmSales.disqualify.reason')}
        />
      </Modal>
    </div>
  );
};

export default Leads;
