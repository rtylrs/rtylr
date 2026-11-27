import { createOneKitClient } from '@voxire/api';
import React, { useCallback, useEffect, useState } from 'react';
import { AlertTriangle, ChevronLeft, ChevronRight, Globe, ReceiptText } from 'lucide-react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { useDebounce } from '@voxire/utils';
import ScreenLayout from '../../layouts/ScreenLayout';
import { useCapabilities } from '../../shared/hooks/useCapabilities';
import SearchBar from '../../shared/components/SearchBar';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import OrderStatusBadge from '../../shared/components/OrderStatusBadge';
import type { OrderStatus } from '../../shared/types/pos';
import { formatCurrency, formatDate } from '../../shared/utils/format';
import { isOrderFinalized, isVoidable, orderStatusLabel } from '../../shared/utils/orderStatus';
import { OrderServiceClient } from '@voxire/voxie/pos/orders/v1/client';
import type { Order } from '@voxire/voxie/pos/orders/v1/types';
import { resolveOpenSessionId } from '../close-session/closeSessionWorkflow';

/** Orders cannot be charged again once settled ('completed' is a legacy server value). */
function isSettled(status: string | undefined): boolean {
  return isOrderFinalized((status ?? '') as OrderStatus) || status === 'completed';
}

const SCOPES = [
  { id: 'all', label: 'All Orders' },
  { id: 'online', label: 'Online Menu' },
  { id: 'shift', label: 'This Shift' },
] as const;
type OrderScope = (typeof SCOPES)[number]['id'];

const FILTERS = ['all', 'pending', 'confirmed', 'preparing', 'ready', 'paid', 'voided', 'refunded'] as const;
type OrderFilter = (typeof FILTERS)[number];
const PAGE_SIZE = 25;
const orderClient = createOneKitClient(OrderServiceClient, 'pos');

const NEXT_KITCHEN_ACTION: Record<string, { status: string; label: string }> = {
  pending: { status: 'confirmed', label: 'Confirm & Send to Kitchen' },
  confirmed: { status: 'preparing', label: 'Start Preparing' },
  preparing: { status: 'ready', label: 'Mark Ready' },
  ready: { status: 'served', label: 'Mark Served' },
};

function isOnlineMenuOrder(order: Order): boolean {
  const sid = String(order.sessionId ?? '').trim();
  return sid === '' || sid === '0';
}

function orderItemsPreview(order: Order): string | null {
  const items = order.items ?? [];
  if (items.length === 0) return null;

  const preview = items
    .slice(0, 3)
    .map((item) => `${item.quantity}x ${item.name}`)
    .join(', ');
  const remaining = items.length - 3;

  return remaining > 0 ? `${preview}, +${remaining} more` : preview;
}

const OrdersScreen: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 250);
  const [scope, setScope] = useState<OrderScope>('all');
  const [filter, setFilter] = useState<OrderFilter>('all');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const [reopening, setReopening] = useState<string | null>(null);
  const [updatingStatus, setUpdatingStatus] = useState<string | null>(null);
  const [voiding, setVoiding] = useState<string | null>(null);
  const [noSession, setNoSession] = useState(false);
  const navigate = useNavigate();
  const [params] = useSearchParams();
  // Reopening PAID orders is a manager capability, not a URL toggle: without
  // this gate any cashier could append ?reopen=1 and un-pay an order.
  const { canVoid } = useCapabilities();
  const reopenMode = params.get('reopen') === '1' && canVoid;

  const loadOrders = useCallback(async () => {
    setLoading(true);
    setError(null);
    setNoSession(false);
    try {
      // Only scope by open register session on "This Shift"; "All Orders" and
      // "Online Menu" omit sessionId so public digital menu orders (session_id IS NULL)
      // appear directly in POS alongside register orders.
      const sessionId = scope === 'shift' ? String(await resolveOpenSessionId()) : undefined;
      const res = await orderClient.listOrders({
        ...(sessionId ? { sessionId } : {}),
        status: filter !== 'all' ? filter : '',
        dateFrom: '',
        dateTo: '',
        search: debouncedSearch.trim(),
        page,
        pageSize: PAGE_SIZE,
      });
      const rawOrders = res.orders ?? [];
      const nextOrders = scope === 'online' ? rawOrders.filter(isOnlineMenuOrder) : rawOrders;
      setOrders(nextOrders);
      setTotal(scope === 'online' ? nextOrders.length : (res.total ?? 0));
      setHasMore(scope === 'online' ? false : (res.hasMore ?? false));
    } catch (e) {
      if (scope === 'shift' && e instanceof Error && e.message.includes('No open register session')) {
        setNoSession(true);
        setOrders([]);
        setTotal(0);
        setHasMore(false);
        return;
      }
      const err = e as {
        response?: { data?: { message?: string } };
        message?: string;
      };
      setError(err?.response?.data?.message ?? 'Failed to load orders');
    } finally {
      setLoading(false);
    }
  }, [debouncedSearch, filter, page, scope]);

  useEffect(() => {
    void loadOrders();
  }, [loadOrders]);

  const handleAdvanceStatus = async (e: React.MouseEvent, order: Order, nextStatus: string) => {
    e.stopPropagation();
    const id = String(order.id);
    setUpdatingStatus(id);
    try {
      await orderClient.updateOrderStatus({
        orderId: id,
        body: { status: nextStatus },
      });
      setOrders((prev) => prev.map((o) => (o.id === order.id ? { ...o, status: nextStatus } : o)));
      toast.success(
        nextStatus === 'confirmed'
          ? `Order #${order.orderNumber || order.id} confirmed and sent to kitchen`
          : `Order #${order.orderNumber || order.id} marked ${nextStatus}`
      );
    } catch (err) {
      const ex = err as {
        response?: { data?: { message?: string } };
        message?: string;
      };
      toast.error(ex?.response?.data?.message ?? 'Failed to update order status');
    } finally {
      setUpdatingStatus(null);
    }
  };

  const handleReopen = async (e: React.MouseEvent, order: Order) => {
    e.stopPropagation();
    setReopening(String(order.id));
    try {
      await orderClient.updateOrderStatus({
        orderId: String(order.id),
        body: { status: 'pending' },
      });
      setOrders((prev) => prev.map((o) => (o.id === order.id ? { ...o, status: 'pending' } : o)));
      toast.success(`Order #${order.orderNumber || order.id} reopened`);
    } catch (err) {
      const e = err as {
        response?: { data?: { message?: string } };
        message?: string;
      };
      toast.error(e?.response?.data?.message ?? 'Failed to reopen order');
    } finally {
      setReopening(null);
    }
  };

  const handleVoid = async (e: React.MouseEvent, order: Order) => {
    e.stopPropagation();
    if (!window.confirm(`Void order #${order.orderNumber || order.id}? This cannot be undone.`)) return;
    setVoiding(String(order.id));
    try {
      await orderClient.voidOrder({ orderId: String(order.id), reason: 'Manager void' });
      setOrders((prev) => prev.map((o) => (o.id === order.id ? { ...o, status: 'voided' } : o)));
      toast.success(`Order #${order.orderNumber || order.id} voided`);
    } catch (err) {
      const e = err as {
        response?: { data?: { message?: string } };
        message?: string;
      };
      toast.error(e?.response?.data?.message ?? 'Failed to void order');
    } finally {
      setVoiding(null);
    }
  };

  const pageStart = orders.length > 0 ? (page - 1) * PAGE_SIZE + 1 : 0;
  const pageEnd = orders.length > 0 ? pageStart + orders.length - 1 : 0;

  return (
    <ScreenLayout title="Orders">
      <div className="w-full p-5 space-y-4">
        {reopenMode && (
          <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs">
            Manager override approved - select a paid order below to reopen it.
          </div>
        )}
        <SearchBar
          value={search}
          onChange={(value) => {
            setPage(1);
            setSearch(value);
          }}
          placeholder="Search orders..."
        />
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="flex gap-2">
            {SCOPES.map((s) => (
              <button
                key={s.id}
                type="button"
                onClick={() => {
                  setPage(1);
                  setScope(s.id);
                }}
                className={`h-9 px-3.5 rounded-lg text-xs font-semibold transition-colors ${
                  scope === s.id
                    ? 'bg-emerald-500/20 border border-emerald-500/40 text-emerald-300'
                    : 'bg-white/8 text-pos-muted hover:text-pos-text'
                }`}
              >
                {s.label}
              </button>
            ))}
          </div>
          <button
            type="button"
            onClick={() => void loadOrders()}
            className="h-9 px-3 rounded-lg bg-white/8 text-xs font-semibold text-pos-text hover:bg-white/15"
          >
            Refresh
          </button>
        </div>
        <div className="flex flex-wrap gap-2">
          {FILTERS.map((f) => (
            <button
              key={f}
              onClick={() => {
                setPage(1);
                setFilter(f);
              }}
              className={`h-10 px-4 rounded-lg text-xs font-semibold uppercase ${
                filter === f ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
              }`}
            >
              {f === 'all' ? 'All' : orderStatusLabel(f as OrderStatus)}
            </button>
          ))}
        </div>
        {loading ? (
          <Skeleton rows={8} />
        ) : noSession ? (
          <EmptyState
            icon={AlertTriangle}
            title="No active session"
            message="Open a register session to view orders."
          />
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Failed to load" message={error} />
        ) : orders.length === 0 ? (
          <EmptyState icon={ReceiptText} title="No orders" />
        ) : (
          <>
            <ul className="space-y-2">
              {orders.map((o) => {
                const preview = orderItemsPreview(o);
                const onlineOrder = isOnlineMenuOrder(o);
                const nextAction = NEXT_KITCHEN_ACTION[String(o.status ?? '').toLowerCase()];
                // Settled orders must not re-enter the payment flow; reopening
                // a paid order is the explicit manager exception.
                const settled = isSettled(o.status as string);

                return (
                  <li
                    key={o.id}
                    className="p-4 rounded-xl bg-pos-card border border-pos-border hover:border-white/20 cursor-pointer transition"
                    title={
                      settled && !reopenMode
                        ? 'This order is already settled and cannot be charged again'
                        : undefined
                    }
                    onClick={() => {
                      if (reopenMode || settled) return;
                      navigate(`/payment/${o.id}`);
                    }}
                  >
                    <div className="flex flex-wrap justify-between items-center gap-3">
                      <div className="flex flex-wrap items-center gap-2.5 min-w-0">
                        <span className="font-bold text-pos-text truncate">
                          #{o.orderNumber || o.id}
                        </span>
                        <OrderStatusBadge status={o.status as OrderStatus} />
                        {onlineOrder && (
                          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-sky-500/15 border border-sky-500/30 text-sky-300">
                            <Globe className="w-3 h-3" />
                            Online Menu
                          </span>
                        )}
                        {o.customerName && (
                          <span className="text-xs text-pos-text/80 truncate">
                            {o.customerName}
                          </span>
                        )}
                        {o.tableId && o.tableId !== '0' && (
                          <span className="text-xs text-pos-muted">Table {o.tableId}</span>
                        )}
                      </div>
                      <div className="flex items-center gap-2 shrink-0">
                        <span className="text-sm font-bold text-pos-text">
                          {formatCurrency(o.total ?? 0)}
                        </span>
                        {!reopenMode && nextAction && (
                          <button
                            type="button"
                            disabled={updatingStatus === String(o.id)}
                            onClick={(e) => void handleAdvanceStatus(e, o, nextAction.status)}
                            className="h-8 px-3 rounded-lg bg-emerald-500/20 border border-emerald-500/40 text-emerald-300 text-xs font-semibold hover:bg-emerald-500/30 disabled:opacity-50 transition-colors"
                          >
                            {updatingStatus === String(o.id) ? '...' : nextAction.label}
                          </button>
                        )}
                        {reopenMode && o.status === 'paid' && (
                          <button
                            type="button"
                            disabled={reopening === String(o.id)}
                            onClick={(e) => void handleReopen(e, o)}
                            className="h-8 px-3 rounded-lg bg-amber-500/20 border border-amber-500/40 text-amber-300 text-xs font-semibold hover:bg-amber-500/30 disabled:opacity-50 transition-colors"
                          >
                            {reopening === String(o.id) ? '...' : 'Reopen'}
                          </button>
                        )}
                        {!reopenMode && isVoidable(o.status as OrderStatus) && (
                          <button
                            type="button"
                            disabled={voiding === String(o.id)}
                            onClick={(e) => void handleVoid(e, o)}
                            className="h-8 px-3 rounded-lg bg-red-500/15 border border-red-500/30 text-red-400 text-xs font-semibold hover:bg-red-500/25 disabled:opacity-50 transition-colors"
                          >
                            {voiding === String(o.id) ? '...' : 'Void'}
                          </button>
                        )}
                      </div>
                    </div>
                    <div className="mt-1 space-y-1">
                      <div className="text-[11px] text-pos-muted">{formatDate(o.createdAt ?? '')}</div>
                      {preview && <div className="text-xs text-pos-muted truncate">{preview}</div>}
                    </div>
                  </li>
                );
              })}
            </ul>
            <div className="flex items-center justify-between gap-3 pt-2 text-xs text-pos-muted">
              <span>
                Page {page}
                {total > 0 ? ` · ${pageStart}-${pageEnd} of ${total}` : ''}
              </span>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  disabled={page <= 1}
                  onClick={() => setPage((value) => Math.max(1, value - 1))}
                  className="h-9 px-3 rounded-lg bg-white/8 text-pos-text disabled:opacity-40 inline-flex items-center gap-1.5"
                >
                  <ChevronLeft className="h-4 w-4" />
                  Prev
                </button>
                <button
                  type="button"
                  disabled={!hasMore}
                  onClick={() => setPage((value) => value + 1)}
                  className="h-9 px-3 rounded-lg bg-white/8 text-pos-text disabled:opacity-40 inline-flex items-center gap-1.5"
                >
                  Next
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </ScreenLayout>
  );
};

export default OrdersScreen;
