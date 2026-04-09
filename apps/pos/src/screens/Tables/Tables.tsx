import React, { useEffect } from 'react';
import { motion } from 'framer-motion';
import { Users, DollarSign, CheckCircle2, Clock, XCircle } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { usePOSStore } from '../../stores/posStore';
import type { TableStatus } from '../../stores/posStore';
import { toast } from 'sonner';
import { posApi } from '../../services/posApi';

const STATUS_META: Record<
  TableStatus,
  { label: string; color: string; bg: string; border: string; icon: React.ComponentType<{ size?: number }> }
> = {
  available: {
    label: 'Available',
    color: '#10b981',
    bg: 'rgba(16,185,129,0.10)',
    border: 'rgba(16,185,129,0.25)',
    icon: CheckCircle2,
  },
  occupied: {
    label: 'Occupied',
    color: '#ef4444',
    bg: 'rgba(239,68,68,0.10)',
    border: 'rgba(239,68,68,0.25)',
    icon: Clock,
  },
  reserved: {
    label: 'Reserved',
    color: '#f59e0b',
    bg: 'rgba(245,158,11,0.10)',
    border: 'rgba(245,158,11,0.25)',
    icon: XCircle,
  },
};

const Tables: React.FC = () => {
  const navigate = useNavigate();
  const { tables, setActiveTableId, setTableStatus, setTables } = usePOSStore();

  useEffect(() => {
    posApi.getTables()
      .then((apiTables) => {
        setTables(apiTables.map((t) => ({
          id: t.id,
          label: t.label,
          status: t.status,
          seats: t.seats,
          orderTotal: t.orderTotal,
        })));
      })
      .catch(() => {});
  }, []);

  const available = tables.filter((t) => t.status === 'available').length;
  const occupied  = tables.filter((t) => t.status === 'occupied').length;
  const reserved  = tables.filter((t) => t.status === 'reserved').length;

  const handleSelectTable = (tableId: string, status: TableStatus) => {
    if (status === 'reserved') {
      toast.warning('This table is reserved. Confirm to seat anyway?', {
        action: {
          label: 'Seat Now',
          onClick: () => {
            setTableStatus(tableId, 'occupied');
            setActiveTableId(tableId);
            navigate('/');
          },
        },
      });
      return;
    }
    setActiveTableId(tableId);
    if (status === 'available') {
      setTableStatus(tableId, 'occupied');
    }
    navigate('/');
    toast.success(`Switched to ${tables.find((t) => t.id === tableId)?.label}`);
  };

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Header stats */}
      <div
        className="shrink-0 px-6 py-4 flex items-center gap-4"
        style={{ borderBottom: '1px solid var(--hub-border)', background: 'var(--hub-surface)' }}
      >
        <h2 className="text-sm font-bold mr-4" style={{ color: 'var(--hub-text)' }}>
          Floor Plan
        </h2>
        {[
          { label: 'Available', count: available, color: '#10b981' },
          { label: 'Occupied',  count: occupied,  color: '#ef4444' },
          { label: 'Reserved',  count: reserved,  color: '#f59e0b' },
        ].map(({ label, count, color }) => (
          <div
            key={label}
            className="flex items-center gap-2 px-3 py-1.5 rounded-lg"
            style={{ background: color + '14', border: `1px solid ${color}30` }}
          >
            <span className="h-2 w-2 rounded-full" style={{ background: color }} />
            <span className="text-xs font-semibold" style={{ color }}>
              {count} {label}
            </span>
          </div>
        ))}
      </div>

      {/* Floor plan grid */}
      <div className="flex-1 overflow-y-auto p-6">
        <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))' }}>
          {tables.map((table, idx) => {
            const meta = STATUS_META[table.status];
            const StatusIcon = meta.icon;

            return (
              <motion.button
                key={table.id}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: idx * 0.05 }}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                onClick={() => handleSelectTable(table.id, table.status)}
                className="flex flex-col rounded-2xl p-4 text-left transition-all relative overflow-hidden"
                style={{
                  background: meta.bg,
                  border: `1.5px solid ${meta.border}`,
                }}
              >
                {/* Table number */}
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>
                      {table.label}
                    </p>
                    <div className="flex items-center gap-1 mt-0.5">
                      <Users size={11} style={{ color: 'var(--hub-text-dim)' }} />
                      <span className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                        {table.seats} seats
                      </span>
                    </div>
                  </div>
                  <span
                    className="flex items-center gap-1 px-2 py-1 rounded-full text-[10px] font-semibold"
                    style={{ background: meta.bg, color: meta.color, border: `1px solid ${meta.border}` }}
                  >
                    <StatusIcon size={10} />
                    {meta.label}
                  </span>
                </div>

                {/* Order total if occupied */}
                {table.status === 'occupied' && table.orderTotal != null && (
                  <div
                    className="flex items-center gap-1.5 mt-auto pt-3"
                    style={{ borderTop: `1px solid ${meta.border}` }}
                  >
                    <DollarSign size={12} style={{ color: meta.color }} />
                    <span className="text-sm font-bold tabular-nums" style={{ color: meta.color }}>
                      ${table.orderTotal.toFixed(2)}
                    </span>
                    <span className="text-[10px] ml-auto" style={{ color: 'var(--hub-text-dim)' }}>
                      open tab
                    </span>
                  </div>
                )}

                {table.status === 'available' && (
                  <div className="mt-auto pt-3" style={{ borderTop: `1px solid ${meta.border}` }}>
                    <span className="text-[10px]" style={{ color: '#10b981' }}>
                      Tap to seat
                    </span>
                  </div>
                )}

                {table.status === 'reserved' && (
                  <div className="mt-auto pt-3" style={{ borderTop: `1px solid ${meta.border}` }}>
                    <span className="text-[10px]" style={{ color: '#f59e0b' }}>
                      Reserved · tap to confirm
                    </span>
                  </div>
                )}
              </motion.button>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default Tables;
