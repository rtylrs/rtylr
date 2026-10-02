import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { ShoppingBag, Clock, CheckCircle2, Loader2, Info } from 'lucide-react';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import {
  useOrders,
  useUpdateOrderStatus,
  type HubOrder,
} from '../../shared/hooks/queries/useOrders';
import { formatMoney } from '../../shared/lib/money';
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
  type BadgeVariant,
  Button,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '../../shared/components/ui';

type OrderTab = 'all' | 'pending' | 'in_progress' | 'completed';

type OrderServerStatus =
  | 'pending'
  | 'confirmed'
  | 'preparing'
  | 'ready'
  | 'served'
  | 'paid'
  | 'completed'
  | 'cancelled'
  | 'refunded'
  | 'voided'
  | 'held';

const PAGE_SIZE = 50;

const IN_PROGRESS_STATUSES = new Set<OrderServerStatus>([
  'confirmed',
  'preparing',
  'ready',
  'served',
  'held',
]);
const COMPLETED_STATUSES = new Set<OrderServerStatus>([
  'paid',
  'completed',
  'cancelled',
  'refunded',
  'voided',
]);

const statusBadgeVariant: Record<string, BadgeVariant> = {
  pending: 'warning',
  confirmed: 'default',
  preparing: 'default',
  ready: 'secondary',
  served: 'secondary',
  held: 'warning',
  paid: 'success',
  completed: 'success',
  cancelled: 'destructive',
  refunded: 'destructive',
  voided: 'destructive',
};

const statusLabel: Record<string, string> = {
  pending: 'Pending',
  confirmed: 'Confirmed',
  preparing: 'Preparing',
  ready: 'Ready',
  served: 'Served',
  held: 'Held',
  paid: 'Paid',
  completed: 'Completed',
  cancelled: 'Cancelled',
  voided: 'Voided',
};

const tabIcons: Record<string, React.ReactNode> = {
  pending: <Clock size={13} />,
  in_progress: <Loader2 size={13} />,
  completed: <CheckCircle2 size={13} />,
};

function orderBucket(status: string): Exclude<OrderTab, 'all'> | null {
  if (status === 'pending') return 'pending';
  if (IN_PROGRESS_STATUSES.has(status as OrderServerStatus)) return 'in_progress';
  if (COMPLETED_STATUSES.has(status as OrderServerStatus)) return 'completed';
  return null;
}

function formatStatus(status: string): string {
  return statusLabel[status] || status.replace(/_/g, ' ');
}

const nextOrderStatus: Partial<
  Record<OrderServerStatus, { status: OrderServerStatus; label: string }>
> = {
  pending: { status: 'confirmed', label: 'Confirm order' },
  confirmed: { status: 'preparing', label: 'Start preparing' },
  preparing: { status: 'ready', label: 'Mark ready' },
  ready: { status: 'served', label: 'Mark served' },
};

const OrdersEmptyRow: React.FC<{ label: string; onRefresh: () => void }> = ({
  label,
  onRefresh,
}) => {
  const { t } = useTranslation();
  return (
    <TableRow>
      <TableCell colSpan={7}>
        <EmptyState
          icon={ShoppingBag}
          title={
            label ? t('orders.emptyTitleFiltered', { status: label }) : t('orders.emptyTitleAll')
          }
          description={t('orders.emptyDescription')}
          action={{ label: t('common.refresh'), onClick: onRefresh }}
        />
      </TableCell>
    </TableRow>
  );
};

function formatTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

type OrderItemWithTotal = HubOrder['items'][number] & { total?: number };

function formatOrderItemTotal(item: OrderItemWithTotal): string {
  const total = Number(item.total);
  return formatMoney(Number.isFinite(total) ? total : item.price * item.quantity);
}

const OrdersTable: React.FC<{
  orders: HubOrder[];
  emptyLabel: string;
  onSelectOrder: (order: HubOrder) => void;
  onRefresh: () => void;
  page: number;
  hasNextPage: boolean;
  onPageChange: (page: number) => void;
}> = ({ orders, emptyLabel, onSelectOrder, onRefresh, page, hasNextPage, onPageChange }) => {
  const { t } = useTranslation();
  const showPagination = hasNextPage || page > 1;

  return (
    <>
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
            <OrdersEmptyRow label={emptyLabel} onRefresh={onRefresh} />
          ) : (
            orders.map((order) => (
              <TableRow key={order.id}>
                <TableCell className="font-mono text-xs">#{order.id}</TableCell>
                <TableCell>{order.customerName || order.customerEmail || '-'}</TableCell>
                <TableCell>{formatMoney(order.subtotal)}</TableCell>
                <TableCell>{formatMoney(order.total)}</TableCell>
                <TableCell>
                  <Badge variant={statusBadgeVariant[order.status] || 'outline'}>
                    {formatStatus(order.status)}
                  </Badge>
                </TableCell>
                <TableCell className="text-gray-400 text-xs">
                  {formatTime(order.createdAt)}
                </TableCell>
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
      {showPagination && (
        <div
          className="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3 text-xs"
          style={{ borderColor: 'var(--dash-border)', color: 'var(--dash-text-muted)' }}
        >
          <span>Page {page}</span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(Math.max(1, page - 1))}
              disabled={page <= 1}
              aria-label={t('common.pagination.previous')}
            >
              {t('common.pagination.previous')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onPageChange(page + 1)}
              disabled={!hasNextPage}
              aria-label={t('common.pagination.next')}
            >
              {t('common.pagination.next')}
            </Button>
          </div>
        </div>
      )}
    </>
  );
};

const Orders: React.FC = () => {
  const [tab, setTab] = useState<OrderTab>('all');
  const [page, setPage] = useState(1);
  const [selectedOrder, setSelectedOrder] = useState<HubOrder | null>(null);
  const {
    data: ordersPage,
    isLoading,
    isError,
    refetch,
  } = useOrders({
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
  });
  // Accept a legacy cached array while the page-shaped adapter is active; the
  // live hook now always supplies the metadata-bearing object.
  const allOrders = useMemo(
    () => (Array.isArray(ordersPage) ? ordersPage : (ordersPage?.orders ?? [])),
    [ordersPage]
  );
  const updateStatusMutation = useUpdateOrderStatus();

  const handleStatusUpdate = (newStatus: OrderServerStatus) => {
    if (!selectedOrder) return;
    updateStatusMutation.mutate(
      { id: selectedOrder.id, status: newStatus },
      {
        onSuccess: () => {
          setSelectedOrder((prev) => (prev ? { ...prev, status: newStatus } : null));
        },
      }
    );
  };

  const pendingOrders = useMemo(
    () => allOrders.filter((o) => orderBucket(o.status) === 'pending'),
    [allOrders]
  );
  const inProgressOrders = useMemo(
    () => allOrders.filter((o) => orderBucket(o.status) === 'in_progress'),
    [allOrders]
  );
  const completedOrders = useMemo(
    () => allOrders.filter((o) => orderBucket(o.status) === 'completed'),
    [allOrders]
  );
  const hasNextPage = Array.isArray(ordersPage)
    ? allOrders.length >= PAGE_SIZE
    : (ordersPage?.hasMore ?? false);

  const ordersForTab: Record<OrderTab, HubOrder[]> = {
    all: allOrders,
    pending: pendingOrders,
    in_progress: inProgressOrders,
    completed: completedOrders,
  };
  const selectedStatus = selectedOrder?.status as OrderServerStatus | undefined;
  const nextAction = selectedStatus ? nextOrderStatus[selectedStatus] : undefined;
  const canCancel =
    selectedStatus === 'pending' ||
    selectedStatus === 'confirmed' ||
    selectedStatus === 'preparing' ||
    selectedStatus === 'ready';

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
          <Button size="sm" variant="outline" onClick={() => void refetch()}>
            Refresh
          </Button>
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
        <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
          Status counts for the current page.
        </p>
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
              <Tabs
                value={tab}
                onValueChange={(v) => {
                  setTab(v as OrderTab);
                  setPage(1);
                  setSelectedOrder(null);
                }}
              >
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
              onRefresh={() => void refetch()}
              page={page}
              hasNextPage={hasNextPage}
              onPageChange={(nextPage) => {
                setPage(Math.max(1, nextPage));
                setSelectedOrder(null);
              }}
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
                  <Badge variant={statusBadgeVariant[selectedOrder.status] || 'outline'}>
                    {formatStatus(selectedOrder.status)}
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
                    {formatMoney(selectedOrder.subtotal)}
                  </span>
                </div>
                <div
                  className="flex justify-between border-b pb-2"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <span>Total</span>
                  <span className="font-bold text-base" style={{ color: 'var(--dash-text)' }}>
                    {formatMoney(selectedOrder.total)}
                  </span>
                </div>
                <div className="flex justify-between pt-1 text-xs">
                  <span>Placed At</span>
                  <span>{new Date(selectedOrder.createdAt).toLocaleString()}</span>
                </div>
              </div>

              {selectedOrder.items && selectedOrder.items.length > 0 && (
                <div
                  className="space-y-2 pt-2 border-t"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <h4 className="text-xs font-semibold text-primary">Line Items</h4>
                  <div className="space-y-1.5 max-h-40 overflow-y-auto pr-1">
                    {selectedOrder.items.map((item, idx) => (
                      <div
                        key={item.id || idx}
                        className="flex justify-between items-center text-xs p-2 rounded-lg"
                        style={{ background: 'var(--dash-hover)' }}
                      >
                        <div>
                          <span className="font-medium text-primary">{item.name}</span>
                          <span className="text-muted ml-2">x{item.quantity}</span>
                        </div>
                        <span className="font-semibold text-primary">
                          {formatOrderItemTotal(item)}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              <div
                className="flex items-center justify-end gap-2 pt-3 border-t"
                style={{ borderColor: 'var(--dash-border)' }}
              >
                {canCancel && (
                  <Button
                    variant="destructive"
                    size="sm"
                    disabled={updateStatusMutation.isPending}
                    onClick={() => handleStatusUpdate('cancelled')}
                  >
                    Cancel Order
                  </Button>
                )}
                {nextAction && (
                  <Button
                    variant="default"
                    size="sm"
                    disabled={updateStatusMutation.isPending}
                    onClick={() => handleStatusUpdate(nextAction.status)}
                  >
                    {nextAction.label}
                  </Button>
                )}
                <Button variant="outline" size="sm" onClick={() => setSelectedOrder(null)}>
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
