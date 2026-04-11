import React from 'react';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useKDSOrders, KDSOrder } from './useKDSOrders';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { formatTime } from '../../shared/utils/format';

const STATUS_ORDER: KDSOrder['status'][] = ['new', 'cooking', 'ready', 'served'];

const statusStyles: Record<KDSOrder['status'], string> = {
  new: 'bg-blue-500/15 border-blue-500/40 text-blue-200',
  cooking: 'bg-amber-500/15 border-amber-500/40 text-amber-200',
  ready: 'bg-emerald-500/15 border-emerald-500/40 text-emerald-200',
  served: 'bg-white/5 border-pos-border text-pos-muted',
};

const KDSScreen: React.FC = () => {
  const station = useSettingsStore((s) => s.kdsStation);
  const setStation = useSettingsStore((s) => s.setKdsStation);
  const { orders, loading, setStatus } = useKDSOrders(station === 'all' ? undefined : station);

  const next = (status: KDSOrder['status']): KDSOrder['status'] => {
    const idx = STATUS_ORDER.indexOf(status);
    return STATUS_ORDER[Math.min(STATUS_ORDER.length - 1, idx + 1)];
  };

  return (
    <ScreenLayout
      title="Kitchen Display"
      subtitle={`Station: ${station}`}
      actions={
        <select
          value={station}
          onChange={(e) => setStation(e.target.value)}
          className="h-10 px-3 rounded-lg bg-white/10 text-pos-text text-sm"
        >
          <option value="all">All stations</option>
          <option value="grill">Grill</option>
          <option value="fry">Fry</option>
          <option value="bar">Bar</option>
          <option value="cold">Cold</option>
        </select>
      }
    >
      <div className="p-5">
        {loading ? (
          <Skeleton rows={6} />
        ) : orders.length === 0 ? (
          <EmptyState icon="🍳" title="No active tickets" />
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
            {orders.map((o) => (
              <div
                key={o.id}
                className={`p-4 rounded-xl border-2 ${statusStyles[o.status]} flex flex-col gap-3`}
              >
                <div className="flex items-baseline justify-between">
                  <span className="font-bold text-lg text-pos-text">#{o.number}</span>
                  <span className="text-[10px] uppercase">{o.station}</span>
                </div>
                <div className="text-[10px] opacity-70">{formatTime(o.createdAt)}</div>
                <ul className="space-y-1.5 flex-1">
                  {o.items.map((it, idx) => (
                    <li key={idx} className="text-sm text-pos-text">
                      <span className="font-semibold">{it.qty}×</span> {it.name}
                      {it.notes && (
                        <div className="text-[10px] italic text-pos-muted">{it.notes}</div>
                      )}
                    </li>
                  ))}
                </ul>
                <button
                  onClick={() => setStatus(o.id, next(o.status))}
                  className="h-11 rounded-lg bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20"
                >
                  {o.status === 'new' && 'Start Cooking'}
                  {o.status === 'cooking' && 'Mark Ready'}
                  {o.status === 'ready' && 'Mark Served'}
                  {o.status === 'served' && 'Served'}
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default KDSScreen;
