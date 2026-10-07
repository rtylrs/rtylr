import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';
import {
  Search,
  ChevronDown,
  ChevronUp,
  Phone,
  Mail,
  Plus,
  Upload,
  Users as UsersIcon,
  Tag as TagIcon,
} from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListCustomers, useUpdateCustomer } from '../../shared/hooks/queries/useCustomers';
import { useLoyaltyConfig } from '../../shared/hooks/queries/useLoyalty';
import { type Customer } from '../../shared/schemas/customers';
import { getCustomerTier, thresholdsFromConfig, Tier, tierColor } from '../../shared/lib/tiers';
import { Button, PageHeader, SkeletonList } from '../../shared/components/ui';
import { CustomerAvatar, TierBadge, TagChip } from '../../shared/components/crm';
import { customersToCsv } from './customerExport';
import { canCrmPermission } from '../../shared/lib/crmPermissions';
import { hasQueryFailed } from '../../../../shared/lib/queryFailureUtils';

const ACCENT = 'var(--crm-accent)';

const tierFilters: Array<Tier | 'All'> = ['All', 'Platinum', 'Gold', 'Silver', 'Bronze'];

function ExpandedDetail({ customer, onView }: { customer: Customer; onView: () => void }) {
  const { t } = useTranslation();
  const tier = getCustomerTier(customer.loyaltyPoints, customer.loyaltyTier);
  return (
    <motion.tr
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      transition={{ duration: 0.15 }}
    >
      <td colSpan={9} style={{ padding: 0 }}>
        <div
          style={{
            padding: '20px 24px',
            background: 'var(--dash-hover)',
            borderTop: '1px solid var(--hub-border)',
            borderBottom: '1px solid var(--hub-border)',
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr) auto',
            gap: '20px',
          }}
        >
          <div>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginBottom: '4px' }}>
              {t('crm.customers.colContactInfo')}
            </p>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px', marginBottom: '6px' }}>
              <Mail size={12} style={{ color: ACCENT }} />
              <span style={{ fontSize: '13px', color: 'var(--hub-text)' }}>{customer.email}</span>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Phone size={12} style={{ color: ACCENT }} />
              <span style={{ fontSize: '13px', color: 'var(--hub-text)' }}>{customer.phone}</span>
            </div>
          </div>
          <div>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginBottom: '4px' }}>
              {t('crmNav.loyalty')}
            </p>
            <TierBadge tier={tier} />
            <p style={{ fontSize: '13px', color: 'var(--hub-text)', marginTop: '6px' }}>
              {t('crm.customers.pointsCount', { count: customer.loyaltyPoints })}
            </p>
          </div>
          <div>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-dim)', marginBottom: '4px' }}>
              {t('crm.customers.colSummary')}
            </p>
            <p style={{ fontSize: '13px', color: 'var(--hub-text)' }}>
              {t('crm.customers.lifetimeSpend', {
                amount: customer.totalSpent.toLocaleString(),
              })}
            </p>
            <p style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {t('crm.customers.visitsCount', { count: customer.visits })}
            </p>
          </div>
          <div style={{ display: 'flex', alignItems: 'center' }}>
            <Button variant="primary" onClick={onView}>
              {t('crm.customers.viewProfile')}
            </Button>
          </div>
        </div>
      </td>
    </motion.tr>
  );
}

const Customers: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const canCreate = canCrmPermission('crm:customers:create');
  const canEdit = canCrmPermission('crm:customers:edit');
  const canExport = canCrmPermission('crm:customers:export');
  const [search, setSearch] = useState('');
  const [tierFilter, setTierFilter] = useState<Tier | 'All'>('All');
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [bulkTag, setBulkTag] = useState('');
  const pageSize = 100;
  const [offset, setOffset] = useState(0);
  const [loadedCustomers, setLoadedCustomers] = useState<Customer[]>([]);
  // Set once a "Load more" page comes back empty.
  const [exhausted, setExhausted] = useState(false);

  // NOTE: search is intentionally synchronous (no debounce) — the empty-state
  // UX reflects the query term immediately and is covered by tests.
  const customersQuery = useListCustomers({
    limit: pageSize,
    offset,
    search: search.trim() || undefined,
  });
  const { data: fetchedCustomers, isFetching, refetch } = customersQuery;
  const customersLoadFailed = hasQueryFailed(customersQuery);
  const customersLoading = customersQuery.isPending && !customersLoadFailed;
  const update = useUpdateCustomer();
  const { data: loyaltyConfig } = useLoyaltyConfig();
  const thresholds = thresholdsFromConfig(loyaltyConfig);

  useEffect(() => {
    if (!fetchedCustomers) return;
    if (offset > 0 && fetchedCustomers.length === 0) {
      // Exact page-size multiple: the extra "Load more" page came back
      // empty — stop offering it.
      setExhausted(true);
    }
    setLoadedCustomers((previous) => {
      if (offset === 0) return fetchedCustomers;
      const byId = new Map(previous.map((customer) => [customer.id, customer]));
      fetchedCustomers.forEach((customer) => byId.set(customer.id, customer));
      return Array.from(byId.values());
    });
  }, [fetchedCustomers, offset]);

  const customers = loadedCustomers;
  const hasMore = !exhausted && fetchedCustomers?.length === pageSize;

  if (customersLoadFailed && customers.length === 0) {
    return (
      <div role="alert" className="space-y-4 py-10">
        <p>{t('crm.customers.loadFailed')}</p>
        <Button onClick={() => void refetch()} disabled={isFetching}>
          {t('crmCommon.retry')}
        </Button>
      </div>
    );
  }

  const filtered = customers.filter((c) => {
    const matchTier =
      tierFilter === 'All' ||
      getCustomerTier(c.loyaltyPoints, c.loyaltyTier, thresholds) === tierFilter;
    return matchTier;
  });

  const allSelected = filtered.length > 0 && filtered.every((c) => selected.has(c.id));
  const toggleAll = () => {
    if (allSelected) setSelected(new Set());
    else setSelected(new Set(filtered.map((c) => c.id)));
  };
  const toggleOne = (id: number) => {
    const next = new Set(selected);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    setSelected(next);
  };

  const applyBulkTag = () => {
    const tagName = bulkTag.trim();
    if (!canEdit || !tagName || selected.size === 0) return;
    const ids = Array.from(selected).filter((id) => {
      const c = customers.find((x) => x.id === id);
      return c && !(c.tags || []).includes(tagName);
    });
    if (ids.length === 0) {
      toast.success(`All selected already tagged "${tagName}"`);
      setBulkTag('');
      setSelected(new Set());
      return;
    }
    let remaining = ids.length;
    let failed = 0;
    ids.forEach((id) => {
      const c = customers.find((x) => x.id === id);
      if (!c) return;
      const tags = [...(c.tags || []), tagName];
      update.mutate(
        { id, data: { tags } },
        {
          onError: () => {
            failed += 1;
          },
          onSettled: () => {
            remaining -= 1;
            if (remaining === 0) {
              if (failed === 0) {
                toast.success(`Tagged ${ids.length} customers with "${tagName}"`);
              } else {
                toast.error(`Failed to tag ${failed} of ${ids.length} customers`);
              }
              setBulkTag('');
              setSelected(new Set());
            }
          },
        }
      );
    });
  };

  const exportCsv = () => {
    if (!canExport) return;
    const csv = customersToCsv(filtered, thresholds);
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'customers.csv';
    a.click();
    URL.revokeObjectURL(url);
    toast.success(t('crm.customers.csvExported'));
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.customers.pageTitle')} · rtylr`} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: '20px', maxWidth: '1280px' }}>
        <PageHeader
          title={t('crm.customers.pageTitle')}
          subtitle={t('crm.customers.loadedCount', { count: customers.length })}
          actions={
            <>
              {canCreate && (
                <Button variant="outline" onClick={() => navigate({ to: '/crm/customers/import' })}>
                  <Upload size={13} />
                  {t('crm.customers.import')}
                </Button>
              )}
              {canExport && (
                <Button variant="outline" onClick={exportCsv}>
                  {t('crm.customers.exportCsv')}
                </Button>
              )}
              {canCreate && (
                <Button variant="primary" onClick={() => navigate({ to: '/crm/customers/new' })}>
                  <Plus size={13} />
                  {t('crm.customers.new')}
                </Button>
              )}
            </>
          }
        />

        {/* Filters */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
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
              minWidth: '220px',
              maxWidth: '340px',
            }}
          >
            <Search size={14} style={{ color: 'var(--hub-text-muted)', flexShrink: 0 }} />
            <input
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setOffset(0);
                setLoadedCustomers([]);
                setSelected(new Set());
                // A new query starts pagination over; without this a page that
                // was previously exhausted keeps "Load more" hidden forever.
                setExhausted(false);
              }}
              placeholder={t('crm.customers.searchPlaceholder')}
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
          <div
            style={{
              display: 'flex',
              gap: '4px',
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '8px',
              padding: '3px',
            }}
          >
            {tierFilters.map((t) => (
              <button
                type="button"
                key={t}
                onClick={() => setTierFilter(t)}
                style={{
                  padding: '4px 12px',
                  borderRadius: '6px',
                  fontSize: '12px',
                  fontWeight: 500,
                  cursor: 'pointer',
                  border: 'none',
                  background:
                    tierFilter === t
                      ? t === 'All'
                        ? ACCENT
                        : tierColor[t as Tier]
                      : 'transparent',
                  color: tierFilter === t ? '#fff' : 'var(--hub-text-muted)',
                  transition: 'all 0.15s',
                }}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        {/* Bulk actions */}
        {canEdit && selected.size > 0 && (
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              padding: '10px 14px',
              background: 'var(--dash-hover)',
              border: '1px solid var(--dash-border)',
              borderRadius: '10px',
            }}
          >
            <span style={{ fontSize: '12px', color: 'var(--hub-text)', fontWeight: 600 }}>
              {selected.size} selected
            </span>
            <div style={{ display: 'flex', gap: '6px', alignItems: 'center', marginLeft: 'auto' }}>
              <TagIcon size={12} style={{ color: 'var(--hub-text-muted)' }} />
              <input
                value={bulkTag}
                onChange={(e) => setBulkTag(e.target.value)}
                placeholder={t('crm.customers.addTagPlaceholder')}
                onKeyDown={(e) => e.key === 'Enter' && applyBulkTag()}
                style={{
                  height: '30px',
                  padding: '0 10px',
                  borderRadius: '6px',
                  background: 'var(--hub-input-bg)',
                  border: '1px solid var(--hub-border)',
                  color: 'var(--hub-text)',
                  fontSize: '12px',
                  outline: 'none',
                  width: '160px',
                }}
              />
              <Button variant="primary" onClick={applyBulkTag} style={{ padding: '6px 12px' }}>
                {t('crm.customers.apply')}
              </Button>
              <Button variant="ghost" onClick={() => setSelected(new Set())}>
                {t('crm.customers.clear')}
              </Button>
            </div>
          </div>
        )}

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
          {customersLoading && customers.length === 0 ? (
            <div style={{ padding: '16px' }}>
              <SkeletonList count={6} />
            </div>
          ) : filtered.length === 0 ? (
            <EmptyState
              icon={UsersIcon}
              title={t('crm.customers.empty')}
              description={
                search
                  ? t('crm.customers.emptySearchDescription', { search })
                  : t('crm.customers.createFirst')
              }
              action={
                search
                  ? {
                      label: t('crm.actions.clearSearch'),
                      onClick: () => {
                        setSearch('');
                        setOffset(0);
                        setLoadedCustomers([]);
                        setSelected(new Set());
                        setExhausted(false);
                      },
                    }
                  : canCreate
                    ? {
                        label: t('crm.customers.new'),
                        onClick: () => navigate({ to: '/crm/customers/new' }),
                      }
                    : undefined
              }
            />
          ) : (
            <>
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                    <th style={{ padding: '11px 12px 11px 16px', width: '28px' }}>
                      <input
                        type="checkbox"
                        disabled={!canEdit}
                        checked={allSelected}
                        onChange={toggleAll}
                        style={{ accentColor: ACCENT, cursor: 'pointer' }}
                      />
                    </th>
                    {[
                      'Customer',
                      'Email',
                      'Phone',
                      'Total Spent',
                      'Visits',
                      'Points',
                      'Tags',
                      'Last Visit',
                      '',
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
                    {filtered.map((c) => {
                      const tier = getCustomerTier(c.loyaltyPoints, c.loyaltyTier, thresholds);
                      const isExpanded = expandedId === c.id;
                      const isSelected = selected.has(c.id);
                      return (
                        <React.Fragment key={c.id}>
                          <tr
                            style={{
                              borderBottom: isExpanded ? 'none' : '1px solid var(--hub-border)',
                              cursor: 'pointer',
                              transition: 'background 0.12s',
                              background:
                                isExpanded || isSelected ? 'rgba(59,130,246,0.04)' : 'transparent',
                            }}
                            onMouseEnter={(e) => {
                              if (!isExpanded)
                                (e.currentTarget as HTMLElement).style.background =
                                  'var(--hub-hover)';
                            }}
                            onMouseLeave={(e) => {
                              if (!isExpanded)
                                (e.currentTarget as HTMLElement).style.background = isSelected
                                  ? 'rgba(59,130,246,0.04)'
                                  : 'transparent';
                            }}
                          >
                            <td
                              style={{ padding: '12px 12px 12px 16px' }}
                              onClick={(e) => e.stopPropagation()}
                            >
                              <input
                                type="checkbox"
                                disabled={!canEdit}
                                checked={isSelected}
                                onChange={() => toggleOne(c.id)}
                                style={{ accentColor: ACCENT, cursor: 'pointer' }}
                              />
                            </td>
                            <td
                              style={{ padding: '12px 16px' }}
                              onClick={() => setExpandedId(isExpanded ? null : c.id)}
                            >
                              <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                                <CustomerAvatar
                                  initials={c.initials}
                                  points={c.loyaltyPoints}
                                  size={32}
                                />
                                <div>
                                  <div
                                    style={{
                                      fontSize: '13px',
                                      fontWeight: 500,
                                      color: 'var(--hub-text)',
                                    }}
                                  >
                                    {c.name}
                                  </div>
                                  <TierBadge tier={tier} />
                                </div>
                              </div>
                            </td>
                            <td
                              style={{
                                padding: '12px 16px',
                                fontSize: '12px',
                                color: 'var(--hub-text-muted)',
                              }}
                            >
                              {c.email}
                            </td>
                            <td
                              style={{
                                padding: '12px 16px',
                                fontSize: '12px',
                                color: 'var(--hub-text-muted)',
                              }}
                            >
                              {c.phone}
                            </td>
                            <td
                              style={{
                                padding: '12px 16px',
                                fontSize: '13px',
                                fontWeight: 600,
                                color: 'var(--hub-text)',
                              }}
                            >
                              {c.totalSpent.toLocaleString()}
                            </td>
                            <td
                              style={{
                                padding: '12px 16px',
                                fontSize: '13px',
                                color: 'var(--hub-text)',
                              }}
                            >
                              {c.visits}
                            </td>
                            <td style={{ padding: '12px 16px' }}>
                              <span
                                style={{
                                  fontSize: '13px',
                                  color: 'var(--hub-text)',
                                  fontWeight: 500,
                                }}
                              >
                                {c.loyaltyPoints.toLocaleString()}
                              </span>
                            </td>
                            <td style={{ padding: '12px 16px' }}>
                              <div
                                style={{
                                  display: 'flex',
                                  gap: '4px',
                                  flexWrap: 'wrap',
                                  maxWidth: '160px',
                                }}
                              >
                                {(c.tags || []).slice(0, 2).map((t) => (
                                  <TagChip key={t} label={t} />
                                ))}
                              </div>
                            </td>
                            <td
                              style={{
                                padding: '12px 16px',
                                fontSize: '12px',
                                color: 'var(--hub-text-muted)',
                              }}
                            >
                              {c.lastVisit}
                            </td>
                            <td
                              style={{ padding: '12px 16px', textAlign: 'right' }}
                              onClick={() => setExpandedId(isExpanded ? null : c.id)}
                            >
                              {isExpanded ? (
                                <ChevronUp size={14} style={{ color: ACCENT }} />
                              ) : (
                                <ChevronDown size={14} style={{ color: 'var(--hub-text-dim)' }} />
                              )}
                            </td>
                          </tr>
                          <AnimatePresence>
                            {isExpanded && (
                              <ExpandedDetail
                                customer={c}
                                onView={() => navigate({ to: `/crm/customers/${c.id}` })}
                              />
                            )}
                          </AnimatePresence>
                        </React.Fragment>
                      );
                    })}
                  </AnimatePresence>
                </tbody>
              </table>
              {hasMore && filtered.length > 0 && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '16px' }}>
                  <Button
                    variant="outline"
                    onClick={() => setOffset((current) => current + pageSize)}
                    disabled={isFetching}
                  >
                    {isFetching ? 'Loading…' : 'Load more customers'}
                  </Button>
                </div>
              )}
            </>
          )}
        </motion.div>
      </div>
    </>
  );
};

export default Customers;
