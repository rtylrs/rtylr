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
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '../../shared/components/ui';

type OrderStatus = 'all' | 'pending' | 'in_progress' | 'completed';

const statusBadgeVariant: Record<string, any> = {
  pending: 'warning',
  in_progress: 'default',
  completed: 'success',
  cancelled: 'destructive',
};

const statusLabel: Record<string, string> = {
  pending: 'Pending',
  in_progress: 'In Progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
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

const OrdersTable: React.FC<{
  orders: HubOrder[];
  emptyLabel: string;
  onSelectOrder: (order: HubOrder) => void;
}> = ({ orders, emptyLabel, onSelectOrder }) => (
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
              <Badge variant={statusBadgeVariant[order.status] || 'outline'}>
                {statusLabel[order.status] || order.status.replace('_', ' ')}
              </Badge>
            </TableCell>
            <TableCell className="text-gray-400 text-xs">{formatTime(order.createdAt)}</TableCell>
            <TableCell className="text-right">
              <Button variant="ghost" size="sm" onClick={() => onSelectOrder(order)}>
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
  const [selectedOrder, setSelectedOrder] = useState<HubOrder | null>(null);
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
              icon: Clock,
            },
            {
              label: 'In Progress',
              count: inProgressOrders.length,
              icon: Loader2,
            },
            {
              label: 'Completed',
              count: completedOrders.length,
              icon: CheckCircle2,
            },
          ].map(({ label, count, icon: Icon }) => (
            <div
              key={label}
              className="rounded-xl p-4 flex items-center justify-between"
              style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
            >
              <div className="flex items-center gap-2">
                <Icon size={16} style={{ color: 'var(--dash-text-muted)' }} />
                <span className="text-sm font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                  {label}
                </span>
              </div>
              <span className="text-2xl font-bold" style={{ color: 'var(--dash-text)' }}>
                {count}
              </span>
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
              onSelectOrder={setSelectedOrder}
            />
          </CardContent>
        </Card>
      </div>

      {/* Order Detail Modal */}
      <Dialog open={!!selectedOrder} onOpenChange={(open) => !open && setSelectedOrder(null)}>
        <DialogContent className="max-w-md">
          {selectedOrder && (
            <div className="space-y-4">
              <DialogHeader>
                <div className="flex items-center justify-between pr-6">
                  <DialogTitle>Order #{selectedOrder.id}</DialogTitle>
                  <Badge variant={statusBadgeVariant[selectedOrder.status]}>
                    {statusLabel[selectedOrder.status]}
                  </Badge>
                </div>
              </DialogHeader>

              <div className="space-y-2 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                <div
                  className="flex justify-between border-b pb-2"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <span>Customer</span>
                  <span className="font-medium" style={{ color: 'var(--dash-text)' }}>
                    {selectedOrder.customerName || selectedOrder.customerEmail || 'Guest'}
                  </span>
                </div>
                <div
                  className="flex justify-between border-b pb-2"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <span>Subtotal</span>
                  <span className="font-medium" style={{ color: 'var(--dash-text)' }}>
                    ${selectedOrder.subtotal.toFixed(2)}
                  </span>
                </div>
                <div
                  className="flex justify-between border-b pb-2"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <span>Total</span>
                  <span className="font-bold text-base" style={{ color: 'var(--dash-text)' }}>
                    ${selectedOrder.total.toFixed(2)}
                  </span>
                </div>
                <div className="flex justify-between pt-1 text-xs">
                  <span>Placed At</span>
                  <span>{new Date(selectedOrder.createdAt).toLocaleString()}</span>
                </div>
              </div>

              <div className="flex justify-end pt-2">
                <Button variant="outline" onClick={() => setSelectedOrder(null)}>
                  Close
                </Button>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
};

export default Orders;
