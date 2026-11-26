import { createOneKitClient } from '@voxire/api';
import { MenuServiceClient } from '@voxire/voxie/hub/menu/v1/client';
import type { MenuTableAction } from '@voxire/voxie/hub/menu/v1/types';
import { TableServiceClient } from '@voxire/voxie/pos/table/v1/client';
import React, { useCallback, useEffect, useState } from 'react';
import { AlertTriangle, Armchair, Bell, Check, CheckCheck, RefreshCw } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import ScreenLayout from '../../layouts/ScreenLayout';
import TableCard from '../../shared/components/TableCard';
import Skeleton from '../../shared/components/Skeleton';
import EmptyState from '../../shared/components/EmptyState';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import { mapRemoteTables, type TableRow } from './tablesWorkflow';

const tableClient = createOneKitClient(TableServiceClient, 'pos');
const menuClient = createOneKitClient(MenuServiceClient, 'pos');

const ACTION_LABELS: Record<string, string> = {
  call_waiter: 'Call Staff',
  request_bill: 'Request Bill',
  split_bill: 'Split Bill',
};

const TablesScreen: React.FC = () => {
  const [tables, setTables] = useState<TableRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tableActions, setTableActions] = useState<MenuTableAction[]>([]);
  const [updatingActionId, setUpdatingActionId] = useState<string | null>(null);
  const locationId = useTerminalContextStore((s) => s.locationId);
  const navigate = useNavigate();

  const loadTableActions = useCallback(async () => {
    try {
      const res = await menuClient.listTableActions({
        ...(locationId && locationId > 0 ? { locationId: String(locationId) } : {}),
      });
      setTableActions((res.actions ?? []).filter((a) => a.status !== 'resolved'));
    } catch {
      // Non-fatal: keep floor view functional even if table actions cannot load.
    }
  }, [locationId]);

  useEffect(() => {
    (async () => {
      setLoading(true);
      try {
        const res = await tableClient.listTables({});
        setTables(mapRemoteTables(res.tables ?? []));
      } catch (e) {
        const err = e as { response?: { data?: { message?: string } }; message?: string };
        setError(err?.response?.data?.message ?? 'Failed to load tables');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  useEffect(() => {
    void loadTableActions();
    const timer = window.setInterval(() => {
      void loadTableActions();
    }, 15_000);
    return () => window.clearInterval(timer);
  }, [loadTableActions]);

  const handleUpdateAction = useCallback(
    async (id: string, status: 'acknowledged' | 'resolved') => {
      setUpdatingActionId(id);
      try {
        await menuClient.updateTableActionStatus({ id, status });
        setTableActions((prev) =>
          status === 'resolved'
            ? prev.filter((a) => a.id !== id)
            : prev.map((a) => (a.id === id ? { ...a, status } : a))
        );
        toast.success(status === 'resolved' ? 'Table request resolved' : 'Table request acknowledged');
      } catch {
        toast.error('Failed to update table request');
      } finally {
        setUpdatingActionId(null);
      }
    },
    []
  );

  const openTable = useCallback(
    (tableNumber: string) => navigate(`/?table=${tableNumber}`),
    [navigate]
  );

  return (
    <ScreenLayout
      title="Tables"
      actions={
        <button
          type="button"
          onClick={() => void loadTableActions()}
          className="h-9 px-3 rounded-lg bg-white/10 text-pos-text text-xs font-semibold hover:bg-white/20 inline-flex items-center gap-1.5"
        >
          <RefreshCw size={14} />
          Refresh Requests
        </button>
      }
    >
      <div className="p-5 space-y-5">
        {/* Live Table Service Queue (Call Staff / Request Bill / Split Bill) */}
        {tableActions.length > 0 && (
          <section
            aria-label="Table service requests"
            className="rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 space-y-3"
          >
            <div className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 text-sm font-semibold text-amber-200">
                <Bell size={16} />
                <span>Table Service Queue ({tableActions.length})</span>
              </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2.5">
              {tableActions.map((act) => {
                const actionId = String(act.id ?? '');
                const busy = updatingActionId === actionId;
                return (
                  <div
                    key={actionId}
                    className="rounded-lg border border-white/10 bg-black/25 p-3 flex items-center justify-between gap-3"
                  >
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-pos-text truncate">
                        {act.tableLabel || `Table ${act.tableId ?? ''}`}
                      </p>
                      <p className="text-xs text-amber-200/90">
                        {ACTION_LABELS[act.actionType ?? ''] ?? act.actionType}
                        {act.status === 'acknowledged' ? ' · On the way' : ''}
                      </p>
                      {act.note && (
                        <p className="text-xs text-white/60 truncate mt-0.5">{act.note}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      {act.status === 'open' && actionId && (
                        <button
                          type="button"
                          disabled={busy}
                          onClick={() => void handleUpdateAction(actionId, 'acknowledged')}
                          className="h-8 px-2.5 rounded-md bg-white/15 hover:bg-white/25 text-xs font-medium text-pos-text inline-flex items-center gap-1 disabled:opacity-50"
                        >
                          <Check size={13} />
                          Ack
                        </button>
                      )}
                      {actionId && (
                        <button
                          type="button"
                          disabled={busy}
                          onClick={() => void handleUpdateAction(actionId, 'resolved')}
                          className="h-8 px-2.5 rounded-md bg-emerald-600 hover:bg-emerald-500 text-xs font-semibold text-white inline-flex items-center gap-1 disabled:opacity-50"
                        >
                          <CheckCheck size={13} />
                          Done
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </section>
        )}

        {loading ? (
          <Skeleton rows={6} />
        ) : error ? (
          <EmptyState icon={AlertTriangle} title="Failed to load" message={error} />
        ) : tables.length === 0 ? (
          <EmptyState icon={Armchair} title="No tables configured" />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {tables.map((t) => (
              <TableCard key={t.id} {...t} onClick={() => openTable(t.number)} />
            ))}
          </div>
        )}
      </div>
    </ScreenLayout>
  );
};

export default TablesScreen;
