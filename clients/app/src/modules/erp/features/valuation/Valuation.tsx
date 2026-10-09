import React, { useEffect, useMemo, useRef, useState } from 'react';
import { motion } from 'motion/react';
import { DollarSign, BarChart3, Boxes } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useTranslation } from 'react-i18next';
import { createOneKitClient, safeRatio, toFiniteNumber } from '@voxire/api';
import { InventoryServiceClient } from '@voxire/voxie/erp/inventory/v1/client';
import { SettingsServiceClient } from '@voxire/voxie/erp/settings/v1/client';
import type { GetValuationResponse, ValuationItem } from '@voxire/voxie/erp/inventory/v1/types';
import { useValuation } from './useValuation';
import PageHeader from '../../shared/components/PageHeader';
import { Skeleton } from '../../shared/components/Skeleton';
import { formatCurrency, formatNumber } from '../../shared/lib/format';
import { useSettings } from '../../shared/lib/settingsStore';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const ACCENT = 'var(--erp-accent)';
const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');
const settingsClient = createOneKitClient(SettingsServiceClient, 'erp');
const VALUATION_PAGE_LIMIT = 500;

async function loadAllValuationItems(firstPage: GetValuationResponse): Promise<ValuationItem[]> {
  const firstItems = firstPage.items ?? [];
  const items = [...firstItems];
  let offset = (firstPage.offset ?? 0) + firstItems.length;
  const hasMore = firstPage.hasMore === true;
  const visitedOffsets = new Set<number>();

  while (hasMore) {
    if (!Number.isSafeInteger(offset) || offset < 0 || visitedOffsets.has(offset)) {
      throw new Error('Invalid valuation pagination cursor');
    }
    visitedOffsets.add(offset);

    const page = await inventoryClient.getValuation({
      limit: VALUATION_PAGE_LIMIT,
      offset,
    });
    const pageItems = page.items ?? [];
    items.push(...pageItems);

    if (page.hasMore !== true) break;
    if (pageItems.length === 0) throw new Error('Valuation pagination returned no rows');
    const nextOffset = (page.offset ?? offset) + pageItems.length;
    if (nextOffset <= offset) throw new Error('Valuation pagination did not advance');
    offset = nextOffset;
  }

  return items;
}

const Valuation: React.FC = () => {
  const { t } = useTranslation('erp');
  const { data, isLoading, isFetching, isError, refetch } = useValuation();
  const storedCurrency = useSettings((state) => state.currency || 'USD');
  const [serverCurrency, setServerCurrency] = useState<string | null>(null);
  const [loadedItems, setLoadedItems] = useState<ValuationItem[] | null>(null);
  const [isLoadingPages, setIsLoadingPages] = useState(false);
  const [pagesError, setPagesError] = useState(false);
  const hasRevalidated = useRef(false);
  const currency = serverCurrency ?? storedCurrency;

  useEffect(() => {
    let active = true;
    void settingsClient
      .getErpSettings({})
      .then((response) => {
        const nextCurrency = response.settings?.currency;
        if (!active || !nextCurrency) return;
        setServerCurrency(nextCurrency);
        useSettings.getState().setCurrency(nextCurrency);
      })
      .catch(() => undefined);

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    // Only current paginated responses need an explicit mount revalidation; a
    // legacy cached response without page metadata has no reliable cursor to refresh.
    const hasPageMetadata =
      data?.hasMore !== undefined ||
      data?.limit !== undefined ||
      data?.offset !== undefined ||
      data?.total !== undefined;
    if (hasRevalidated.current || !data || !hasPageMetadata) return;

    hasRevalidated.current = true;
    void refetch();
  }, [data, refetch]);

  useEffect(() => {
    let active = true;

    if (!data || data.hasMore !== true) {
      setLoadedItems(data?.items ?? []);
      setIsLoadingPages(false);
      setPagesError(false);
      return () => {
        active = false;
      };
    }

    setLoadedItems(null);
    setIsLoadingPages(true);
    setPagesError(false);
    void loadAllValuationItems(data)
      .then((items) => {
        if (!active) return;
        setLoadedItems(items);
        setIsLoadingPages(false);
      })
      .catch(() => {
        if (!active) return;
        setPagesError(true);
        setIsLoadingPages(false);
      });

    return () => {
      active = false;
    };
  }, [data]);

  const items = loadedItems ?? data?.items ?? [];
  const totalCost = useMemo(
    () => items.reduce((sum, item) => sum + toFiniteNumber(item.totalCost), 0),
    [items]
  );
  const totalItems = useMemo(
    () => items.reduce((sum, item) => sum + toFiniteNumber(item.totalItems), 0),
    [items]
  );

  const categories = useMemo(() => {
    const totals = new Map<string, number>();
    items.forEach((item) => {
      const name = item.categoryName?.trim() || 'Uncategorized';
      totals.set(name, (totals.get(name) ?? 0) + toFiniteNumber(item.totalValue));
    });
    return Array.from(totals, ([name, value]) => ({ name, value })).sort(
      (a, b) => b.value - a.value
    );
  }, [items]);
  const maxCat = categories.length > 0 ? categories[0].value : 1;
  const queryFailed = Boolean(isError || pagesError);
  const isBusy = Boolean(
    isLoading || isFetching || isLoadingPages || (data?.hasMore === true && loadedItems === null)
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/valuation'].title))} />
      <PageHeader
        title={t(routeLabels['/erp/valuation'].title)}
        description="Total value of inventory on hand"
      />

      {queryFailed ? (
        <div
          role="alert"
          className="rounded-xl p-5 space-y-3"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <p className="text-sm" style={{ color: 'var(--hub-text)' }}>
            {t('erpErrors.generic')}
          </p>
          <button
            type="button"
            onClick={() => void refetch()}
            className="rounded-lg px-3 py-1.5 text-xs font-semibold"
            style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
          >
            {t('erp.valuation.refresh')}
          </button>
        </div>
      ) : isBusy ? (
        <Skeleton className="h-32 rounded-xl" />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <DollarSign size={14} style={{ color: ACCENT }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Stock Value
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {formatCurrency(data?.totalValue ?? 0, currency)}
            </p>
          </div>
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <DollarSign size={14} style={{ color: 'var(--erp-success)' }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Cost Basis
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--erp-success)' }}>
              {formatCurrency(totalCost, currency)}
            </p>
          </div>
          <div
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 size={14} style={{ color: 'var(--erp-info)' }} />
              <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                Total Items
              </span>
            </div>
            <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
              {formatNumber(totalItems)}
            </p>
          </div>
        </div>
      )}

      {!queryFailed && (
        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            Value by Category
          </h3>
          {isBusy ? (
            <Skeleton className="h-24 rounded-lg" />
          ) : categories.length === 0 ? (
            <EmptyState
              icon={Boxes}
              title={t('erp.valuation.empty.title')}
              description={t('erp.valuation.empty.description')}
              size="compact"
              action={{ label: t('erp.valuation.refresh'), onClick: () => void refetch() }}
            />
          ) : (
            <div className="space-y-2">
              {categories.map((cat) => (
                <div key={cat.name}>
                  <div className="flex justify-between text-xs mb-1">
                    <span style={{ color: 'var(--hub-text-muted)' }}>{cat.name}</span>
                    <span style={{ color: 'var(--hub-text)' }}>
                      {formatCurrency(cat.value, currency)}
                    </span>
                  </div>
                  <div className="h-1.5 rounded-full" style={{ background: 'var(--hub-hover)' }}>
                    <div
                      className="h-full rounded-full"
                      style={{
                        width: `${safeRatio(cat.value, maxCat) * 100}%`,
                        background: ACCENT,
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </motion.div>
  );
};

export default Valuation;
