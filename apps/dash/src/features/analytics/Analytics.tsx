import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { DollarSign, ShoppingBag, TrendingUp, Star, BarChart2, Loader2 } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Tabs,
  TabsList,
  TabsTrigger,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '../../shared/components/ui';
import { cn } from '../../shared/lib/utils';
import { useKpis } from '../../shared/hooks/queries/useKpis';
import { insightsProductClient } from '../../shared/lib/clients';

interface BestSeller {
  productId: number;
  name: string;
  unitsSold: number;
  revenueMinorUnits?: number;
  revenue?: number;
}

type RawBestSeller = {
  productId?: string | number;
  name?: string;
  unitsSold?: string | number;
  revenueMinorUnits?: string | number;
  revenue?: string | number;
  product_id?: string | number;
  units_sold?: string | number;
  revenue_minor_units?: string | number;
};

function finiteNumber(value: unknown): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function normalizeBestSeller(item: RawBestSeller): BestSeller {
  return {
    productId: finiteNumber(item.productId ?? item.product_id),
    name: item.name ?? '',
    unitsSold: finiteNumber(item.unitsSold ?? item.units_sold),
    revenueMinorUnits:
      item.revenueMinorUnits !== undefined || item.revenue_minor_units !== undefined
        ? finiteNumber(item.revenueMinorUnits ?? item.revenue_minor_units)
        : undefined,
    revenue: item.revenue !== undefined ? finiteNumber(item.revenue) : undefined,
  };
}

function revenueDollars(item: BestSeller): number {
  if (typeof item.revenueMinorUnits === 'number') return item.revenueMinorUnits / 100;
  return item.revenue ?? 0;
}

// No range parameter: insights.product.v1.GetBestSellersRequest carries only
// `limit`, so there was nothing to send it in. It previously took one and used
// it in the queryKey alone, which refetched byte-identical data under a new key.
const useBestSellers = () =>
  useQuery({
    queryKey: ['dash-best-sellers'],
    queryFn: async (): Promise<BestSeller[]> => {
      const res = await insightsProductClient.getBestSellers({ limit: 5 });
      return (res.products ?? []).map((item) =>
        normalizeBestSeller({
          productId: item.productId,
          name: item.name,
          unitsSold: item.unitsSold,
          revenueMinorUnits: item.revenueMinorUnits,
        })
      );
    },
    staleTime: 60_000,
  });

const Analytics: React.FC = () => {
  const { t } = useTranslation();
  const { data: kpis, isLoading } = useKpis();
  const {
    data: bestSellers = [],
    isLoading: sellersLoading,
    refetch: refetchBestSellers,
  } = useBestSellers();

  // revenue and orderCount are today-scoped server-side: onion's kpis.sql
  // defines them as SumRevenueToday and CountOrdersToday, both filtered to
  // created_at >= CURDATE(). The card labels below are accurate; the field
  // names this screen used (revenueToday, ordersToday) never existed on
  // GetKPIsResponse, so every card read a transform default of 0.
  const avgOrderValue = kpis && kpis.orderCount > 0 ? kpis.revenue / kpis.orderCount : 0;

  const statCards = [
    {
      label: 'Revenue Today',
      value: isLoading ? '-' : `$${(kpis?.revenue ?? 0).toFixed(2)}`,
      icon: DollarSign,
      iconClass: 'text-emerald-600 bg-emerald-500/15',
    },
    {
      label: 'Orders Today',
      value: isLoading ? '-' : String(kpis?.orderCount ?? 0),
      icon: ShoppingBag,
      iconClass: 'text-blue-600 bg-blue-500/15',
    },
    {
      label: 'Avg Order Value',
      value: isLoading ? '-' : `$${avgOrderValue.toFixed(2)}`,
      icon: TrendingUp,
      iconClass: 'text-violet-600 bg-violet-500/15',
    },
    {
      // 'Customers', not 'New Customers': onion's CountCustomers has no date
      // filter, so this is the total. The old label described data that does
      // not exist on the contract.
      label: 'Customers',
      value: isLoading ? '-' : String(kpis?.customerCount ?? 0),
      icon: Star,
      iconClass: 'text-amber-600 bg-amber-500/15',
    },
  ];

  return (
    <>
      <HelmetContainer title="Analytics · rtylr" />
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-primary">Analytics</h1>
            <p className="text-sm text-gray-400 mt-0.5">Track your business performance</p>
          </div>
          {/* Disabled rather than removed, and never drawn as though it works.
              A date range is a real analytics concept, but there is nowhere to
              send one: hub.dashboard.v1.GetKPIsRequest is an empty message and
              insights.product.v1.GetBestSellersRequest carries only `limit`. The
              KPI figures are today-only server-side regardless (onion's
              kpis.sql filters on CURDATE()).

              aria-disabled rather than the native disabled attribute, so the
              reason below stays reachable to a screen reader. onClick is passed
              after TabsTrigger's own, which is what neutralises it. */}
          <div className="flex flex-col items-end gap-1">
            <Tabs value="" onValueChange={() => {}}>
              <TabsList aria-describedby="analytics-range-note">
                {[
                  ['7d', 'Last 7 days'],
                  ['30d', 'Last 30 days'],
                  ['90d', 'Last 90 days'],
                ].map(([value, label]) => (
                  <TabsTrigger
                    key={value}
                    value={value}
                    aria-disabled="true"
                    tabIndex={-1}
                    className="opacity-50 cursor-not-allowed"
                    onClick={(event) => event.preventDefault()}
                  >
                    {label}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
            <p id="analytics-range-note" className="text-xs text-gray-400">
              {t('analytics.rangeUnavailable')}
            </p>
          </div>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-4 gap-4">
          {statCards.map((stat) => {
            const Icon = stat.icon;
            return (
              <Card key={stat.label}>
                <CardContent className="pt-5">
                  <div className="flex items-start justify-between">
                    <div>
                      <p className="text-xs font-medium text-gray-400 uppercase tracking-wide">
                        {stat.label}
                      </p>
                      <div className="flex items-center gap-2 mt-1.5">
                        {isLoading ? (
                          <Loader2 size={16} className="animate-spin text-gray-400" />
                        ) : (
                          <p className="text-2xl font-bold text-primary">{stat.value}</p>
                        )}
                      </div>
                    </div>
                    <div
                      className={cn(
                        'h-10 w-10 rounded-xl flex items-center justify-center',
                        stat.iconClass
                      )}
                    >
                      <Icon size={20} />
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Top Items</CardTitle>
          </CardHeader>
          <CardContent className="pt-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Units Sold</TableHead>
                  <TableHead>Revenue</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {sellersLoading ? (
                  <TableRow>
                    <TableCell colSpan={3}>
                      <div className="flex items-center justify-center py-6">
                        <Loader2 size={18} className="animate-spin text-gray-300" />
                      </div>
                    </TableCell>
                  </TableRow>
                ) : bestSellers.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={3}>
                      <EmptyState
                        icon={BarChart2}
                        title={t('analytics.emptyNoDataTitle')}
                        description={t('analytics.emptyNoDataDescription')}
                        action={{
                          label: t('common.refresh'),
                          onClick: () => void refetchBestSellers(),
                        }}
                        size="compact"
                      />
                    </TableCell>
                  </TableRow>
                ) : (
                  bestSellers.map((item) => (
                    <TableRow key={item.productId}>
                      <TableCell className="font-medium">{item.name}</TableCell>
                      <TableCell>{item.unitsSold.toLocaleString()}</TableCell>
                      <TableCell>${revenueDollars(item).toFixed(2)}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default Analytics;
