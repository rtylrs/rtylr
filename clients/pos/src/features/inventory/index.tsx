import { createOneKitClient } from '@voxire/api';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Package, Store } from 'lucide-react';
import { toast } from 'sonner';
import { AccessServiceClient } from '@voxire/voxie/pos/access/v1/client';
import { LocationServiceClient } from '@voxire/voxie/pos/location/v1/client';
import ScreenLayout from '../../layouts/ScreenLayout';
import EmptyState from '../../shared/components/EmptyState';
import Skeleton from '../../shared/components/Skeleton';
import { useEmployeeSessionStore } from '../../shared/stores/employeeSessionStore';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import {
  buildStockAction,
  mapInventoryItems,
  mapInventoryLocations,
  summarizeLowStockItems,
} from './inventoryWorkflow';
import { listBranchInventory, listLowStock, recordStockAction } from './stockActionClient';
import type { InventoryItem, InventoryLocation, StockAction } from './inventoryWorkflow';

const accessClient = createOneKitClient(AccessServiceClient, 'pos');
const locationClient = createOneKitClient(LocationServiceClient, 'pos');

const ACTION_LABELS: Record<StockAction, string> = {
  receive: 'Receive',
  adjust: 'Adjust',
  waste: 'Waste',
  transfer: 'Transfer',
};

const ACTION_COLORS: Record<StockAction, string> = {
  receive: 'bg-emerald-600 hover:bg-emerald-500',
  adjust: 'bg-blue-600 hover:bg-blue-500',
  waste: 'bg-red-600 hover:bg-red-500',
  transfer: 'bg-amber-600 hover:bg-amber-500',
};

interface ActionModalProps {
  item: InventoryItem;
  action: StockAction;
  locations: InventoryLocation[];
  currentBranchId: number;
  onClose: () => void;
  onSubmit: (
    itemId: number,
    action: StockAction,
    qty: string,
    reason: string,
    toBranchId?: number | null,
  ) => Promise<void>;
}

function ActionModal({
  item,
  action,
  locations,
  currentBranchId,
  onClose,
  onSubmit,
}: ActionModalProps) {
  const [qty, setQty] = useState('');
  const [reason, setReason] = useState('');
  const [toBranchId, setToBranchId] = useState('');
  const [busy, setBusy] = useState(false);
  const destinationLocations = locations.filter((location) => location.id !== currentBranchId);
  const needsDestination = action === 'transfer';

  const submit = async () => {
    setBusy(true);
    try {
      await onSubmit(item.itemId, action, qty, reason, Number(toBranchId) || null);
      onClose();
    } catch {
      // Failures are toasted by the parent; the modal stays open so the
      // entered quantity and reason are not lost.
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-full max-w-sm rounded-2xl bg-pos-card border border-pos-border p-5 space-y-4">
        <h2 className="text-base font-bold text-pos-text">
          {ACTION_LABELS[action]}: {item.name}
        </h2>
        <p className="text-xs text-pos-muted">
          Current stock: {item.currentQty} {item.unit}
        </p>
        {needsDestination && (
          <div className="space-y-1">
            <label className="text-xs text-pos-muted">Destination location</label>
            <select
              value={toBranchId}
              onChange={(e) => setToBranchId(e.target.value)}
              className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
            >
              <option value="">Select location</option>
              {destinationLocations.map((location) => (
                <option key={location.id} value={location.id}>
                  {location.name}
                </option>
              ))}
            </select>
          </div>
        )}
        <div className="space-y-1">
          <label className="text-xs text-pos-muted">Quantity ({item.unit})</label>
          <input
            autoFocus
            type="number"
            inputMode="decimal"
            value={qty}
            onChange={(e) => setQty(e.target.value)}
            placeholder="0"
            className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
          />
        </div>
        <div className="space-y-1">
          <label className="text-xs text-pos-muted">Reason / note</label>
          <input
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="e.g. Weekly delivery"
            className="w-full h-12 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
          />
        </div>
        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 h-11 rounded-lg bg-white/8 text-pos-text text-sm font-semibold"
          >
            Cancel
          </button>
          <button
            onClick={submit}
            disabled={busy || !qty || !reason.trim() || (needsDestination && !toBranchId)}
            className={`flex-1 h-11 rounded-lg text-white text-sm font-semibold transition disabled:opacity-50 ${ACTION_COLORS[action]}`}
          >
            {busy ? 'Saving…' : ACTION_LABELS[action]}
          </button>
        </div>
      </div>
    </div>
  );
}

const InventoryScreen: React.FC = () => {
  const locationId = useTerminalContextStore((s) => s.locationId);
  const employee = useEmployeeSessionStore((s) => s.employee);
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [lowStock, setLowStock] = useState<InventoryItem[]>([]);
  const [locations, setLocations] = useState<InventoryLocation[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [modal, setModal] = useState<{ item: InventoryItem; action: StockAction } | null>(null);

  const branchId = locationId ? Number(locationId) : 0;

  const load = useCallback(async () => {
    if (!branchId) return;
    setLoading(true);
    try {
      const [inv, low] = await Promise.all([
        listBranchInventory(String(branchId)),
        listLowStock(String(branchId)),
      ]);
      setItems(mapInventoryItems(inv.inventory ?? []));
      setLowStock(mapInventoryItems(low.items ?? []));
    } catch {
      toast.error('Failed to load inventory');
    } finally {
      setLoading(false);
    }
  }, [branchId]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    void Promise.allSettled([accessClient.getMyLocations({}), locationClient.listLocations({})]).then(
      (results) => {
        const merged = new Map<number, InventoryLocation>();
        for (const result of results) {
          if (result.status !== 'fulfilled') continue;
          for (const location of mapInventoryLocations(result.value.locations ?? [])) {
            if (!merged.has(location.id)) merged.set(location.id, location);
          }
        }
        setLocations([...merged.values()]);
      },
    );
  }, []);

  const handleAction = async (
    itemId: number,
    action: StockAction,
    quantityInput: string,
    reason: string,
    toBranchId?: number | null,
  ) => {
    try {
      const built = buildStockAction({
        itemId,
        action,
        quantityInput,
        reason,
        branchId,
        toBranchId,
        employeeId: Number(employee?.id) || null,
      });
      await recordStockAction(built);
      toast.success(`${ACTION_LABELS[action]} recorded`);
      void load();
    } catch (err) {
      // Toast the failure and rethrow so the modal stays open (its submit
      // swallows the rejection) instead of silently discarding the input.
      toast.error(err instanceof Error ? err.message : 'Action failed');
      throw err;
    }
  };

  const term = search.trim().toLowerCase();
  const filtered = useMemo(
    () =>
      items.filter(
        (it) =>
          !term ||
          it.name.toLowerCase().includes(term) ||
          (it.sku ? it.sku.toLowerCase().includes(term) : false)
      ),
    [items, term]
  );

  const lowStockIds = useMemo(() => new Set(lowStock.map((l) => l.itemId)), [lowStock]);
  const lowStockSummary = useMemo(() => summarizeLowStockItems(lowStock), [lowStock]);
  const hasTransferDestinations = locations.some((location) => location.id !== branchId);

  return (
    <ScreenLayout
      title="Inventory"
      subtitle={branchId ? `Branch ${branchId}` : 'No branch selected'}
      actions={
        <button
          onClick={() => void load()}
          className="h-9 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold"
        >
          Refresh
        </button>
      }
    >
      <div className="w-full min-w-0 p-5 space-y-4">
        {lowStock.length > 0 && (
          <div className="rounded-lg bg-amber-500/10 border border-amber-500/30 px-4 py-3">
            <p className="text-xs font-semibold text-amber-300">
              {lowStockSummary.count} item{lowStockSummary.count !== 1 ? 's' : ''} low on stock
            </p>
            {lowStockSummary.names.length > 0 && (
              <p className="mt-1 truncate text-xs text-amber-100/80">
                {lowStockSummary.names.join(', ')}
                {lowStockSummary.hiddenCount > 0 && `, +${lowStockSummary.hiddenCount} more`}
              </p>
            )}
          </div>
        )}

        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search items…"
          className="w-full h-11 px-4 rounded-lg bg-white/8 border border-pos-border text-pos-text outline-none focus:border-white/25"
        />

        {loading ? (
          <Skeleton rows={8} />
        ) : !branchId ? (
          <EmptyState icon={Store} title="No branch selected" message="Assign a terminal first." />
        ) : filtered.length === 0 ? (
          <EmptyState icon={Package} title="No items found" />
        ) : (
          <div className="space-y-2">
            {filtered.map((item) => (
              <div
                key={item.itemId}
                className="rounded-xl bg-pos-card border border-pos-border px-4 py-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold text-pos-text truncate">
                      {item.name}
                    </span>
                    {lowStockIds.has(item.itemId) && (
                      <span className="shrink-0 rounded-full bg-amber-500/15 px-2 py-0.5 text-[10px] font-semibold text-amber-400">
                        LOW
                      </span>
                    )}
                  </div>
                  {item.sku && (
                    <p className="text-[10px] text-pos-muted font-mono">{item.sku}</p>
                  )}
                </div>
                <div className="shrink-0 text-right">
                  <span
                    className={`text-base font-bold tabular-nums ${
                      lowStockIds.has(item.itemId) ? 'text-amber-400' : 'text-pos-text'
                    }`}
                  >
                    {item.currentQty}
                  </span>
                  <span className="ml-1 text-xs text-pos-muted">{item.unit}</span>
                </div>
                <div className="shrink-0 flex flex-wrap justify-end gap-1">
                  {(['receive', 'adjust', 'waste', 'transfer'] as StockAction[]).map((action) => {
                    const transferDisabled = action === 'transfer' && !hasTransferDestinations;
                    return (
                      <button
                        key={action}
                        onClick={() => setModal({ item, action })}
                        disabled={transferDisabled}
                        className={`h-8 px-2 rounded-lg text-white text-xs font-medium transition disabled:opacity-40 ${ACTION_COLORS[action]}`}
                      >
                        {ACTION_LABELS[action]}
                      </button>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {modal && (
        <ActionModal
          item={modal.item}
          action={modal.action}
          locations={locations}
          currentBranchId={branchId}
          onClose={() => setModal(null)}
          onSubmit={handleAction}
        />
      )}
    </ScreenLayout>
  );
};

export default InventoryScreen;
