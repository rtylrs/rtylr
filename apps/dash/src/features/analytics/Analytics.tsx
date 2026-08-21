import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
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
import { safe } from '../../shared/lib/api';

interface BestSeller {
  productId: number;
  name: string;
  unitsSold: number;
  revenueMinorUnits?: number;
  revenue?: number;
}

type RawBestSeller = Partial<BestSeller> & {
  product_id?: string | number;
  units_sold?: string | number;
  revenue_minor_units?: string | number;
  revenue?: string | number;
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

const useBestSellers = (range: string = '7d') =>
  useQuery({
    queryKey: ['dash-best-sellers', range],
    queryFn: async (): Promise<BestSeller[]> => {
      const res = await unwrap(
        safe.get<{ products?: RawBestSeller[] } | RawBestSeller[]>(
          '/insights/product/v1/best-sellers',
          {
            params: { limit: 5, range },
          }
        )
      );
      return ((Array.isArray(res) ? res : res.products) ?? []).map(normalizeBestSeller);
    },
    staleTime: 60_000,
  });

const Analytics: React.FC = () => {
  const { t } = useTranslation();
  const [range, setRange] = useState('7d');
  const { data: kpis, isLoading } = useKpis(range);
  const {
    data: bestSellers = [],
    isLoading: sellersLoading,
    refetch: refetchBestSellers,
  } = useBestSellers(range);

  const avgOrderValue = kpis && kpis.ordersToday > 0 ? kpis.revenueToday / kpis.ordersToday : 0;

  const statCards = [
    {
      label: 'Revenue Today',
      value: isLoading ? '-' : `$${(kpis?.revenueToday ?? 0).toFixed(2)}`,
      icon: DollarSign,
      iconClass: 'text-emerald-600 bg-emerald-500/15',
    },
    {
      label: 'Orders Today',
      value: isLoading ? '-' : String(kpis?.ordersToday ?? 0),
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
      label: 'New Customers',
      value: isLoading ? '-' : String(kpis?.newCustomers ?? 0),
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
          <Tabs value={range} onValueChange={setRange}>
            <TabsList>
              <TabsTrigger value="7d">Last 7 days</TabsTrigger>
              <TabsTrigger value="30d">Last 30 days</TabsTrigger>
              <TabsTrigger value="90d">Last 90 days</TabsTrigger>
            </TabsList>
          </Tabs>
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

        {/* Revenue sparkline */}
        {kpis && kpis.revenueSeries.length > 1 && (
          <Card>
            <CardHeader>
              <CardTitle>Revenue Trend</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <div style={{ height: 80, position: 'relative' }}>
                <svg
                  width="100%"
                  height="80"
                  viewBox={`0 0 ${kpis.revenueSeries.length - 1} 100`}
                  preserveAspectRatio="none"
                >
                  {(() => {
                    const vals = kpis.revenueSeries;
                    const max = Math.max(...vals);
                    const min = Math.min(...vals);
                    const range = max - min || 1;
                    const pts = vals
                      .map((v, i) => `${i},${100 - ((v - min) / range) * 90}`)
                      .join(' ');
                    return (
                      <>
                        <polyline
                          fill="none"
                          stroke="#10b981"
                          strokeWidth="2"
                          vectorEffect="non-scaling-stroke"
                          points={pts}
                        />
                        <polyline
                          fill="rgba(16,185,129,0.12)"
                          stroke="none"
                          points={`0,100 ${pts} ${vals.length - 1},100`}
                        />
                      </>
                    );
                  })()}
                </svg>
              </div>
            </CardContent>
          </Card>
        )}

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
