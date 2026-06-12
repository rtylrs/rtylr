import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { ShoppingBag, Clock, CheckCircle2, Loader2, Info } from 'lucide-react';
import { useOrders, type HubOrder } from '../../shared/hooks/queries/useOrders';
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
  Badge,
  Button,
} from '../../shared/components/ui';

type OrderStatus = 'all' | 'pending' | 'in_progress' | 'completed';

const statusBadgeVariant: Record<string, any> = {
  pending: 'warning',
  in_progress: 'default',
  completed: 'success',
};

const statusLabel: Record<string, string> = {
  pending: 'Pending',
  in_progress: 'In Progress',
  completed: 'Completed',
};

const tabIcons: Record<string, React.ReactNode> = {
  pending: <Clock size={13} />,
  in_progress: <Loader2 size={13} />,
  completed: <CheckCircle2 size={13} />,
};

const EmptyState: React.FC<{ label: string }> = ({ label }) => (
  <TableRow>
    <TableCell colSpan={7}>
      <div className="flex flex-col items-center justify-center py-12 text-center">
        <ShoppingBag size={32} className="mb-3" style={{ color: 'var(--dash-text-dim)' }} />
        <p className="text-sm font-medium" style={{ color: 'var(--dash-text-muted)' }}>
          No {label} orders
        </p>
        <p className="text-xs mt-1" style={{ color: 'var(--dash-text-dim)' }}>
          Orders will appear here once placed by customers
        </p>
      </div>
    </TableCell>
  </TableRow>
);

function formatTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

const OrdersTable: React.FC<{ orders: HubOrder[]; emptyLabel: string }> = ({
  orders,
  emptyLabel,
}) => (
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Order #</TableHead>
        <TableHead>Customer</TableHead>
        <TableHead>Subtotal</TableHead>
        <TableHead>Total</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Time</TableHead>
        <TableHead className="text-right">Actions</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {orders.length === 0 ? (
        <EmptyState label={emptyLabel} />
      ) : (
        orders.map((order) => (
          <TableRow key={order.id}>
            <TableCell className="font-mono text-xs">#{order.id}</TableCell>
            <TableCell>{order.customerName || order.customerEmail || '-'}</TableCell>
            <TableCell>${order.subtotal.toFixed(2)}</TableCell>
            <TableCell>${order.total.toFixed(2)}</TableCell>
            <TableCell>
              <Badge variant={statusBadgeVariant[order.status]}>{statusLabel[order.status]}</Badge>
            </TableCell>
            <TableCell className="text-gray-400 text-xs">{formatTime(order.createdAt)}</TableCell>
            <TableCell className="text-right">
              <Button variant="ghost" size="sm">
                View
              </Button>
            </TableCell>
          </TableRow>
        ))
      )}
    </TableBody>
  </Table>
);

const Orders: React.FC = () => {
  const [tab, setTab] = useState<OrderStatus>('all');
  const { data: allOrders = [], isLoading, isError } = useOrders();

  const pendingOrders = useMemo(() => allOrders.filter((o) => o.status === 'pending'), [allOrders]);
  const inProgressOrders = useMemo(
    () => allOrders.filter((o) => o.status === 'in_progress'),
    [allOrders]
  );
  const completedOrders = useMemo(
    () => allOrders.filter((o) => o.status === 'completed'),
    [allOrders]
  );

  const ordersForTab: Record<OrderStatus, HubOrder[]> = {
    all: allOrders,
    pending: pendingOrders,
    in_progress: inProgressOrders,
    completed: completedOrders,
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Orders
          </h1>
        </div>
        <div
          className="rounded-xl p-8 flex items-center justify-center"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <Loader2 size={20} className="animate-spin" style={{ color: 'var(--dash-text-dim)' }} />
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Orders
          </h1>
        </div>
        <div
          className="rounded-xl p-8 flex flex-col items-center justify-center gap-2"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <ShoppingBag size={28} style={{ color: 'var(--dash-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
            Failed to load orders. Please refresh.
          </p>
        </div>
      </div>
    );
  }

  return (
    <>
      <HelmetContainer title="Orders · rtylr" />
      <div className="space-y-6">
        {/* Header */}
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            Orders
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
            Incoming orders from your customer menus
          </p>
        </div>

        {/* POS info banner */}
        <div
          className="flex items-start gap-3 rounded-xl p-4"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          <Info size={16} className="shrink-0 mt-0.5" style={{ color: '#4583C4' }} />
          <div>
            <p className="text-sm font-medium" style={{ color: 'var(--dash-text)' }}>
              Orders are placed by customers
            </p>
            <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
              Orders come in when customers scan your QR code or visit your menu link. Staff orders
              are created through your POS system.
            </p>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4">
          {[
            {
              label: 'Pending',
              count: pendingOrders.length,
              color: '#f59e0b',
              bg: 'rgba(245,158,11,0.12)',
              border: 'rgba(245,158,11,0.2)',
              icon: Clock,
            },
            {
              label: 'In Progress',
              count: inProgressOrders.length,
              color: '#4583C4',
              bg: 'rgba(69,131,196,0.12)',
              border: 'rgba(69,131,196,0.2)',
              icon: Loader2,
            },
            {
              label: 'Completed',
              count: completedOrders.length,
              color: '#10b981',
              bg: 'rgba(16,185,129,0.12)',
              border: 'rgba(16,185,129,0.2)',
              icon: CheckCircle2,
            },
          ].map(({ label, count, color, bg, border, icon: Icon }) => (
            <div
              key={label}
              className="rounded-xl p-4 flex items-center justify-between"
              style={{ background: bg, border: `1px solid ${border}`, color }}
            >
              <div className="flex items-center gap-2">
                <Icon size={16} />
                <span className="text-sm font-semibold">{label}</span>
              </div>
              <span className="text-2xl font-bold">{count}</span>
            </div>
          ))}
        </div>

        {/* Tabs */}
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center justify-between">
              <CardTitle>All Orders</CardTitle>
              <Tabs value={tab} onValueChange={(v) => setTab(v as OrderStatus)}>
                <TabsList>
                  <TabsTrigger value="all">All</TabsTrigger>
                  <TabsTrigger value="pending">
                    <span className="flex items-center gap-1.5">{tabIcons.pending} Pending</span>
                  </TabsTrigger>
                  <TabsTrigger value="in_progress">
                    <span className="flex items-center gap-1.5">
                      {tabIcons.in_progress} In Progress
                    </span>
                  </TabsTrigger>
                  <TabsTrigger value="completed">
                    <span className="flex items-center gap-1.5">
                      {tabIcons.completed} Completed
                    </span>
                  </TabsTrigger>
                </TabsList>
              </Tabs>
            </div>
          </CardHeader>
          <CardContent className="pt-0">
            <OrdersTable
              orders={ordersForTab[tab]}
              emptyLabel={tab === 'all' ? '' : tab.replace('_', ' ')}
            />
          </CardContent>
        </Card>
      </div>
    </>
  );
};

export default Orders;
