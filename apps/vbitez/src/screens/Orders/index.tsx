import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ScreenLayout from '../../layouts/ScreenLayout';
import SearchBar from '../../components/SearchBar';
import EmptyState from '../../components/EmptyState';
import Skeleton from '../../components/Skeleton';
import OrderStatusBadge from '../../components/OrderStatusBadge';
import { posService } from '../../services/posService';
import type { Order } from '../../types/pos';
import { formatCurrency, formatDate } from '../../utils/format';

const OrdersScreen: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<'all' | 'open' | 'paid' | 'voided'>('all');
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const res = await posService.getOrders();
        const raw = res?.data?.orders ?? res?.data ?? [];
        const mapped: Order[] = (Array.isArray(raw) ? raw : []).map((o: any) => ({
          id: String(o.id),
          number: o.number,
          tableNumber: o.tableNumber,
          status: (o.status ?? 'open') as Order['status'],
          items: o.items ?? [],
          subtotal: Number(o.subtotal ?? 0),
          tax: Number(o.tax ?? 0),
          tip: Number(o.tip ?? 0),
          discount: Number(o.discount ?? 0),
          total: Number(o.total ?? 0),
          createdAt: o.createdAt ?? new Date().toISOString(),
        }));
        setOrders(mapped);
      } catch (e: any) {
        setError(e?.response?.data?.message ?? 'Failed to load orders');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const filtered = orders.filter((o) => {
    if (filter !== 'all' && o.status !== filter) return false;
    if (search && !(o.number ?? o.id).toString().includes(search)) return false;
    return true;
  });

  return (
    <ScreenLayout title="Orders">
      <div className="p-5 max-w-5xl space-y-4">
        <SearchBar value={search} onChange={setSearch} placeholder="Search orders..." />
        <div className="flex gap-2">
          {(['all', 'open', 'paid', 'voided'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`h-10 px-4 rounded-lg text-xs font-semibold uppercase ${
                filter === f ? 'bg-pos-text text-pos-bg' : 'bg-white/8 text-pos-text'
              }`}
            >
              {f}
            </button>
          ))}
        </div>
        {loading ? (
          <Skeleton rows={8} />
        ) : error ? (
          <EmptyState icon="⚠" title="Failed to load" message={error} />
        ) : filtered.length === 0 ? (
          <EmptyState icon="🧾" title="No orders" />
        ) : (
          <ul className="space-y-2">
            {filtered.map((o) => (
              <li
                key={o.id}
                className="p-4 rounded-xl bg-pos-card border border-pos-border hover:border-white/20 cursor-pointer transition"
                onClick={() => navigate(`/payment/${o.id}`)}
              >
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <span className="font-bold text-pos-text">#{o.number ?? o.id}</span>
                    <OrderStatusBadge status={o.status} />
                    {o.tableNumber && (
                      <span className="text-xs text-pos-muted">Table {o.tableNumber}</span>
                    )}
                  </div>
                  <span className="text-sm font-bold text-pos-text">
                    {formatCurrency(o.total)}
                  </span>
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
