import React, { useState } from 'react';
import { DollarSign, ShoppingBag, TrendingUp, Star, BarChart2 } from 'lucide-react';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '../../../components/ui';
import { cn } from '../../../lib/utils';

const statCards = [
  { label: 'Total Revenue', value: '$0.00', icon: DollarSign, iconClass: 'text-emerald-600 bg-emerald-500/15' },
  { label: 'Total Orders', value: '0', icon: ShoppingBag, iconClass: 'text-blue-600 bg-blue-500/15' },
  { label: 'Avg Order Value', value: '$0.00', icon: TrendingUp, iconClass: 'text-violet-600 bg-violet-500/15' },
  { label: 'Top Item', value: '—', icon: Star, iconClass: 'text-amber-600 bg-amber-500/15' },
];

const Analytics: React.FC = () => {
  const [range, setRange] = useState('7d');

  return (
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
                    <p className="text-xs font-medium text-gray-400 uppercase tracking-wide">{stat.label}</p>
                    <p className="text-2xl font-bold text-primary mt-1.5">{stat.value}</p>
                  </div>
                  <div className={cn('h-10 w-10 rounded-xl flex items-center justify-center', stat.iconClass)}>
                    <Icon size={20} />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Chart Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle>Revenue Over Time</CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <div className="flex flex-col items-center justify-center h-52 rounded-xl" style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}>
            <BarChart2 size={36} className="text-gray-200 mb-3" />
            <p className="text-sm font-medium text-gray-400">Chart coming soon</p>
            <p className="text-xs text-gray-300 mt-1">Revenue visualization will appear here</p>
          </div>
        </CardContent>
      </Card>

      {/* Top Items */}
      <Card>
        <CardHeader>
          <CardTitle>Top Items</CardTitle>
        </CardHeader>
        <CardContent className="pt-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Orders</TableHead>
                <TableHead>Revenue</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow>
                <TableCell colSpan={3}>
                  <div className="flex flex-col items-center justify-center py-8 text-center">
                    <Star size={28} className="text-gray-200 mb-3" />
                    <p className="text-sm font-medium text-gray-400">No data yet</p>
                    <p className="text-xs text-gray-300 mt-1">Top performing items will appear here</p>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
};

export default Analytics;
