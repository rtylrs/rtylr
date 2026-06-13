import { unwrap } from '@voxire/api';
import React, { useEffect, useState } from 'react';
import { AlertTriangle, ReceiptText } from 'lucide-react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../shared/components/SearchBar';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import OrderStatusBadge from '../../shared/components/OrderStatusBadge';
import { posService } from '../../shared/services/api/posService';
import type { Order, OrderStatus } from '../../shared/types/pos';
import { formatCurrency, formatDate } from '../../shared/utils/format';
import { isVoidable, orderStatusLabel } from '../../shared/utils/orderStatus';
import { normalizeOrdersResponse } from './ordersWorkflow';

const FILTERS = ['all', 'pending', 'preparing', 'ready', 'paid', 'voided', 'refunded'] as const;
type OrderFilter = (typeof FILTERS)[number];

const OrdersScreen: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<OrderFilter>('all');
  const [reopening, setReopening] = useState<string | null>(null);
  const [voiding, setVoiding] = useState<string | null>(null);
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const reopenMode = params.get('reopen') === '1';

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const res = await unwrap(posService.getOrders());
        setOrders(normalizeOrdersResponse(res));
      } catch (e) {
        const err = e as { response?: { data?: { message?: string } }; message?: string };
        setError(err?.response?.data?.message ?? 'Failed to load orders');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleReopen = async (e: React.MouseEvent, order: Order) => {
    e.stopPropagation();
    setReopening(String(order.id));
    try {
      await unwrap(posService.reopenOrder(Number(order.id)));
      setOrders((prev) => prev.map((o) => (o.id === order.id ? { ...o, status: 'pending' } : o)));
      toast.success(`Order #${order.number ?? order.id} reopened`);
    } catch (err) {
      const e = err as { response?: { data?: { message?: string } }; message?: string };
      toast.error(e?.response?.data?.message ?? 'Failed to reopen order');
    } finally {
      setReopening(null);
    }
  };

  const handleVoid = async (e: React.MouseEvent, order: Order) => {
    e.stopPropagation();
    if (!window.confirm(`Void order #${order.number ?? order.id}? This cannot be undone.`)) return;
    setVoiding(String(order.id));
    try {
      await unwrap(posService.voidOrder(Number(order.id), { reason: 'Manager void' }));
      setOrders((prev) =>
        prev.map((o) => (o.id === order.id ? { ...o, status: 'voided' } : o)),
      );
      toast.success(`Order #${order.number ?? order.id} voided`);
    } catch (err) {
      const e = err as { response?: { data?: { message?: string } }; message?: string };
      toast.error(e?.response?.data?.message ?? 'Failed to void order');
    } finally {
      setVoiding(null);
    }
  };

  const filtered = orders.filter((o) => {
    if (filter !== 'all' && o.status !== filter) return false;
    if (search && !(o.number ?? o.id).toString().includes(search)) return false;
    return true;
  });

  return (
    <ScreenLayout title="Orders">
      <div className="p-5 max-w-5xl space-y-4">
        {reopenMode && (
          <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs">
            Manager override approved - select a paid order below to reopen it.
          </div>
        )}
        <SearchBar value={search} onChange={setSearch} placeholder="Search orders..." />
        <div className="flex gap-2">
          {FILTERS.map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
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
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Failed to load" message={error} />
        ) : filtered.length === 0 ? (
          <EmptyState icon={ReceiptText} title="No orders" />
        ) : (
          <ul className="space-y-2">
            {filtered.map((o) => (
              <li
                key={o.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border hover:border-white/20 cursor-pointer transition"
                onClick={() => !reopenMode && navigate(`/payment/${o.id}`)}
              >
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <span className="font-bold text-pos-text">#{o.number ?? o.id}</span>
                    <OrderStatusBadge status={o.status} />
                    {o.tableNumber && (
                      <span className="text-xs text-pos-muted">Table {o.tableNumber}</span>
                    )}
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-bold text-pos-text">{formatCurrency(o.total)}</span>
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
                    {!reopenMode && isVoidable(o.status) && (
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
                <div className="text-[11px] text-pos-muted mt-1">{formatDate(o.createdAt)}</div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </ScreenLayout>
  );
};

export default OrdersScreen;
