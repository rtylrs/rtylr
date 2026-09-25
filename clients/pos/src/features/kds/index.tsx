import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { ChefHat } from 'lucide-react';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useKDSOrders, KDSOrder } from './useKDSOrders';
import { useSettingsStore } from '../../shared/stores/settingsStore';
import { useKdsTick, formatElapsed, elapsedSeconds } from './useKdsTimer';
import { sendKdsTicket } from '../../shared/services/hardware/kdsHardware';

const STATUS_ORDER: KDSOrder['status'][] = ['new', 'cooking', 'ready', 'served'];

const statusStyles: Record<KDSOrder['status'], string> = {
  new: 'bg-blue-500/15 border-blue-500/40 text-blue-200',
  cooking: 'bg-amber-500/15 border-amber-500/40 text-amber-200',
  ready: 'bg-emerald-500/15 border-emerald-500/40 text-emerald-200',
  served: 'bg-white/5 border-pos-border text-pos-muted',
};

const urgencyStyles: Record<'ok' | 'warn' | 'crit', string> = {
  ok: 'text-pos-muted',
  warn: 'text-amber-400 font-semibold',
  crit: 'text-red-400 font-bold animate-pulse',
};

const ACTIVE_STATUSES = new Set<KDSOrder['status']>(['new', 'cooking']);

interface KdsTicketCardProps {
  order: KDSOrder;
  shortcutNumber?: number;
  isSelected: boolean;
  onBump: (order: KDSOrder) => void;
}

// Memoized card that subscribes to the shared 1s tick itself, so elapsed
// timers stay live without re-rendering the whole grid every second.
const KdsTicketCard = React.memo<KdsTicketCardProps>(function KdsTicketCard({
  order,
  shortcutNumber,
  isSelected,
  onBump,
}) {
  useKdsTick();
  const elapsed = elapsedSeconds(order.createdAt);
  const { label: elapsedLabel, urgency } = formatElapsed(elapsed);
  const isActive = ACTIVE_STATUSES.has(order.status);

  return (
    <div
      className={`p-4 rounded-xl border-2 ${statusStyles[order.status]} flex flex-col gap-3 transition-all ${
        isSelected ? 'ring-2 ring-white/40' : ''
      }`}
    >
      <div className="flex items-baseline justify-between">
        <div className="flex items-center gap-2">
          <span className="font-bold text-lg text-pos-text">#{order.number}</span>
          {isActive && (
            <span className={`text-[10px] tabular-nums ${urgencyStyles[urgency]}`}>
              {elapsedLabel}
            </span>
          )}
        </div>
        <span className="text-[10px] uppercase opacity-70">{order.station}</span>
      </div>

      {shortcutNumber != null && (
        <span className="text-[9px] text-pos-muted opacity-40">⌨ {shortcutNumber}</span>
      )}

      <ul className="space-y-1.5 flex-1">
        {order.items.map((it, i) => (
          <li key={it.id || `${it.name}-${i}`} className="text-sm text-pos-text">
            <span className="font-semibold">{it.qty}×</span> {it.name}
            {it.notes && <div className="text-[10px] italic text-pos-muted">{it.notes}</div>}
          </li>
        ))}
      </ul>

      {order.status !== 'served' && (
        <button
          onClick={() => onBump(order)}
          className="h-11 rounded-lg bg-white/10 text-pos-text font-semibold text-sm hover:bg-white/20 active:scale-95 transition-all"
        >
          {order.status === 'new' && 'Start Cooking'}
          {order.status === 'cooking' && 'Mark Ready'}
          {order.status === 'ready' && 'Mark Served'}
        </button>
      )}
    </div>
  );
});

const KDSScreen: React.FC = () => {
  const station = useSettingsStore((s) => s.kdsStation);
  const setStation = useSettingsStore((s) => s.setKdsStation);
  const kdsHardware = useSettingsStore((s) => s.kdsHardware);
  const { orders, loading, error, setStatus } = useKDSOrders(station === 'all' ? undefined : station);

  const availableStations = useMemo(() => {
    const seen = new Set<string>();
    for (const o of orders) {
      if (o.station && o.station !== 'Kitchen') seen.add(o.station);
    }
    return Array.from(seen).sort();
  }, [orders]);

  const [selectedIdx, setSelectedIdx] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);
  const sentIds = useRef(new Set<string>());

  const next = (status: KDSOrder['status']): KDSOrder['status'] => {
    const idx = STATUS_ORDER.indexOf(status);
    return STATUS_ORDER[Math.min(STATUS_ORDER.length - 1, idx + 1)];
  };

  // setStatus is a stable useCallback, so bump (and therefore the keyboard
  // listener below) no longer churns on every render.
  const bump = useCallback(
    async (order: KDSOrder) => {
      await setStatus(order.id, next(order.status));
    },
    [setStatus],
  );

  // Bump-bar keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (orders.length === 0) return;
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLSelectElement) return;

      if (e.key >= '1' && e.key <= '9') {
        const idx = Number(e.key) - 1;
        if (idx < orders.length) void bump(orders[idx]);
        return;
      }

      const activeOrders = orders.filter((o) => ACTIVE_STATUSES.has(o.status));

      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        const target = activeOrders[selectedIdx] ?? activeOrders[0];
        if (target) void bump(target);
        return;
      }

      if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIdx((i) => Math.min(i + 1, activeOrders.length - 1));
      } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIdx((i) => Math.max(i - 1, 0));
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [orders, selectedIdx, bump]);

  // Send new orders to hardware KDS display (once per order)
  useEffect(() => {
    if (kdsHardware.transport === 'none') return;
    orders
      .filter((o) => o.status === 'new' && !sentIds.current.has(o.id))
      .forEach((o) => {
        sentIds.current.add(o.id);
        void sendKdsTicket(kdsHardware, {
          orderNumber: o.number,
          station: o.station || undefined,
          items: o.items.map((it) => ({
            name: it.name,
            quantity: it.qty,
            notes: it.notes,
            modifiers: [],
          })),
          createdAt: o.createdAt,
        });
      });
  }, [orders, kdsHardware]);

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
          {availableStations.map((s) => (
            <option key={s} value={s.toLowerCase()}>
              {s}
            </option>
          ))}
        </select>
      }
    >
      <div className="p-5" ref={containerRef}>
        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState icon={ChefHat} title="Couldn't load tickets" message={error} />
        ) : orders.length === 0 ? (
          <EmptyState icon={ChefHat} title="No active tickets" />
        ) : (
          <>
            <p className="text-[10px] text-pos-muted mb-3">
              Keys 1–9 bump ticket · Arrow keys navigate · Enter bump selected
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
              {orders.map((o, idx) => {
                const isActive = ACTIVE_STATUSES.has(o.status);
                return (
                  <KdsTicketCard
                    key={o.id}
                    order={o}
                    shortcutNumber={idx < 9 ? idx + 1 : undefined}
                    isSelected={isActive && idx === selectedIdx}
                    onBump={bump}
                  />
                );
              })}
            </div>
          </>
        )}
      </div>
    </ScreenLayout>
  );
};

export default KDSScreen;
