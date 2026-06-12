import React, { useState } from 'react';
import { motion, type Variants } from 'motion/react';
import {
  Plus,
  FileText,
  Send,
  CheckCircle,
  Clock as ClockIcon,
  ThumbsUp,
  Trash2,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { PurchaseOrder } from '../../shared/services/erpApi';
import {
  usePurchaseOrders,
  useApprovePurchaseOrder,
  useSendPurchaseOrder,
  useCreatePurchaseOrder,
} from './usePurchaseOrders';
import { useSuppliers } from '../suppliers/useSuppliers';
import { useProducts } from '../products/useProducts';
import PrintPOButton from '../../shared/components/PrintPOButton';
import ExportButton from '../../shared/components/ExportButton';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

const statusConfig: Record<
  PurchaseOrder['status'],
  {
    color: string;
    bg: string;
    icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  }
> = {
  Draft: { color: 'var(--hub-text-muted)', bg: 'var(--hub-hover)', icon: FileText },
  'Pending Approval': { color: '#a78bfa', bg: 'rgba(167,139,250,0.12)', icon: ClockIcon },
  Approved: { color: '#60a5fa', bg: 'rgba(96,165,250,0.12)', icon: ThumbsUp },
  Sent: { color: AMBER, bg: 'rgba(245,158,11,0.1)', icon: Send },
  Received: { color: '#10b981', bg: 'rgba(16,185,129,0.1)', icon: CheckCircle },
};

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2, ease: 'easeOut' } },
};

const columns = ['Order', 'Supplier', 'Expected', 'Items', 'Total', 'Status', 'Actions'];

const SkeletonTableRow: React.FC = () => (
  <div
    className="grid px-4 py-3 items-center gap-2"
    style={{ gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr' }}
  >
    {Array.from({ length: 7 }).map((_, i) => (
      <div
        key={i}
        className="h-3 rounded animate-pulse"
        style={{ background: 'var(--hub-border)', opacity: 0.6 }}
      />
    ))}
  </div>
);

const Purchasing: React.FC = () => {
  const [page, setPage] = useState(1);
  const { data: ordersResult, isLoading } = usePurchaseOrders({ page });
  const orders = ordersResult?.orders ?? [];
  const hasMore = ordersResult?.hasMore ?? false;
  const approveMut = useApprovePurchaseOrder();
  const sendMut = useSendPurchaseOrder();
  const createMut = useCreatePurchaseOrder();
  const { data: supplierResult } = useSuppliers();
  const suppliers = supplierResult?.suppliers ?? [];
  const { data: productResult } = useProducts();
  const products = productResult?.products ?? [];
  const [filter, setFilter] = useState<PurchaseOrder['status'] | 'All'>('All');
  const [showCreate, setShowCreate] = useState(false);
  const [newSupplierId, setNewSupplierId] = useState('');
  const [newPoNumber, setNewPoNumber] = useState('');
  const [newExpectedDelivery, setNewExpectedDelivery] = useState(
    new Date(Date.now() + 7 * 86400000).toISOString().slice(0, 10)
  );
  const [newNotes, setNewNotes] = useState('');
  const [newItems, setNewItems] = useState<
    Array<{ productId: string; quantityOrdered: string; unitCost: string }>
  >([{ productId: '', quantityOrdered: '', unitCost: '' }]);

  const filteredOrders = filter === 'All' ? orders : orders.filter((o) => o.status === filter);

  const statusCounts: Record<string, number> = {
    Draft: orders.filter((o) => o.status === 'Draft').length,
    'Pending Approval': orders.filter((o) => o.status === 'Pending Approval').length,
    Sent: orders.filter((o) => o.status === 'Sent').length,
    Received: orders.filter((o) => o.status === 'Received').length,
  };

  const handleApprove = (id: number): void => {
    approveMut.mutate(id, {
      onSuccess: () => toast.success(`PO-${id} approved`),
      onError: () => toast.error(`Failed to approve PO-${id}`),
    });
  };

  const handleSend = (id: number): void => {
    sendMut.mutate(id, {
      onSuccess: () => toast.success(`PO-${id} sent to supplier`),
      onError: () => toast.error(`Failed to send PO-${id}`),
    });
  };

  const handleCreatePO = async (): Promise<void> => {
    if (!newSupplierId || !newPoNumber.trim()) {
      toast.error('Supplier and PO number are required');
      return;
    }
    const validItems = newItems.filter((it) => it.productId && Number(it.quantityOrdered) > 0);
    if (validItems.length === 0) {
      toast.error('At least one item with a product and quantity is required');
      return;
    }
    try {
      await createMut.mutateAsync({
        supplierId: Number(newSupplierId),
        poNumber: newPoNumber.trim(),
        taxAmount: 0,
        expectedDelivery: newExpectedDelivery,
        notes: newNotes.trim(),
        items: validItems.map((it) => ({
          productId: Number(it.productId),
          quantityOrdered: Number(it.quantityOrdered),
          unitCost: it.unitCost ? Number(it.unitCost) : 0,
        })),
      });
      toast.success('Purchase order draft created');
      setShowCreate(false);
      setNewSupplierId('');
      setNewPoNumber('');
      setNewExpectedDelivery(new Date(Date.now() + 7 * 86400000).toISOString().slice(0, 10));
      setNewNotes('');
      setNewItems([{ productId: '', quantityOrdered: '', unitCost: '' }]);
    } catch {
      toast.error('Failed to create purchase order');
    }
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title="Purchase orders · rtylr" />
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Purchase Orders
        </h2>
        <div className="flex items-center gap-2">
          <ExportButton
            data={orders.map((o) => ({
              id: o.id,
              supplier: o.supplier,
              date: o.date,
              expectedDate: o.expectedDate,
              items: o.items,
              total: o.total,
              status: o.status,
            }))}
            filename="purchase-orders.csv"
          />
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
            style={{ background: AMBER, color: '#000' }}
          >
            <Plus size={14} />
            New Order
          </button>
        </div>
      </motion.div>

      <motion.div variants={itemVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        {(Object.entries(statusCounts) as [PurchaseOrder['status'], number][]).map(
          ([status, count]) => {
            const sc = statusConfig[status];
            if (!sc) return null;
            const Icon = sc.icon;
            const active = filter === status;
            return (
              <button
                key={status}
                onClick={() => setFilter(active ? 'All' : status)}
                className="flex items-center gap-3 px-4 py-3 rounded-xl transition-all text-left"
                style={{
                  background: 'var(--hub-card)',
                  border: active ? '1px solid rgba(245,158,11,0.5)' : '1px solid var(--hub-border)',
                }}
              >
                <span
                  className="flex items-center justify-center h-8 w-8 rounded-lg shrink-0"
                  style={{ background: sc.bg }}
                >
                  <Icon size={15} style={{ color: sc.color }} />
                </span>
                <div>
                  <p className="text-lg font-bold" style={{ color: 'var(--hub-text)' }}>
                    {isLoading ? '…' : count}
                  </p>
                  <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                    {status}
                  </p>
                </div>
              </button>
            );
          }
        )}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl overflow-hidden"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div
          className="grid px-4 py-3"
          style={{
            gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr',
            borderBottom: '1px solid var(--hub-border)',
          }}
        >
          {columns.map((col) => (
            <span
              key={col}
              className="text-[10px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--hub-text-dim)' }}
            >
              {col}
            </span>
          ))}
        </div>

        <div className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {isLoading
            ? Array.from({ length: 6 }).map((_, i) => <SkeletonTableRow key={i} />)
            : filteredOrders.map((order) => {
                const sc = statusConfig[order.status] || statusConfig['Draft'];
                return (
                  <div
                    key={order.id}
                    className="grid px-4 py-3 items-center transition-colors"
                    style={{ gridTemplateColumns: '1fr 1.5fr 1fr 0.6fr 0.9fr 1fr 1.2fr' }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.background = 'transparent';
                    }}
                  >
                    <span
                      className="text-xs font-mono font-medium"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      PO-{order.id}
                    </span>
                    <span
                      className="text-sm truncate pr-2"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      {order.supplier}
                    </span>
                    <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                      {order.expectedDate}
                    </span>
                    <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {order.items}
                    </span>
                    <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                      {order.total}
                    </span>
                    <span>
                      <span
                        className="text-[10px] font-medium px-2 py-0.5 rounded-full"
                        style={{ background: sc.bg, color: sc.color }}
                      >
                        {order.status}
                      </span>
                    </span>
                    <div className="flex items-center gap-1">
                      {order.status === 'Pending Approval' && (
                        <button
                          onClick={() => handleApprove(order.id)}
                          className="text-[10px] font-semibold px-2 py-1 rounded-md"
                          style={{ background: 'rgba(96,165,250,0.15)', color: '#60a5fa' }}
                        >
                          Approve
                        </button>
                      )}
                      {order.status === 'Draft' && (
                        <button
                          onClick={() => handleSend(order.id)}
                          className="text-[10px] font-semibold px-2 py-1 rounded-md"
                          style={{ background: 'rgba(245,158,11,0.15)', color: AMBER }}
                        >
                          Send
                        </button>
                      )}
                      <PrintPOButton order={order} />
                    </div>
                  </div>
                );
              })}
        </div>
      </motion.div>

      {showCreate && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ background: 'rgba(0,0,0,0.5)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              New Purchase Order
            </h3>
            <select
              value={newSupplierId}
              onChange={(e) => setNewSupplierId(e.target.value)}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            >
              <option value="">Select supplier*</option>
              {suppliers.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
            <input
              value={newPoNumber}
              onChange={(e) => setNewPoNumber(e.target.value)}
              placeholder="PO number*"
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <input
              type="date"
              value={newExpectedDelivery}
              onChange={(e) => setNewExpectedDelivery(e.target.value)}
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <input
              value={newNotes}
              onChange={(e) => setNewNotes(e.target.value)}
              placeholder="Notes (optional)"
              className="w-full px-3 py-2 rounded-lg text-sm outline-none"
              style={{
                background: 'var(--hub-input-bg)',
                border: '1px solid var(--hub-border)',
                color: 'var(--hub-text)',
              }}
            />
            <p
              className="text-[11px] font-semibold uppercase tracking-wide"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              Items *
            </p>
            {newItems.map((item, idx) => (
              <div key={idx} className="flex gap-1 items-center">
                <select
                  value={item.productId}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) => (i === idx ? { ...it, productId: e.target.value } : it))
                    )
                  }
                  className="flex-1 px-2 py-1.5 rounded-lg text-xs outline-none"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                >
                  <option value="">Product</option>
                  {products.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min="1"
                  value={item.quantityOrdered}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) =>
                        i === idx ? { ...it, quantityOrdered: e.target.value } : it
                      )
                    )
                  }
                  placeholder="Qty"
                  className="w-14 px-2 py-1.5 rounded-lg text-xs outline-none"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                />
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={item.unitCost}
                  onChange={(e) =>
                    setNewItems((prev) =>
                      prev.map((it, i) => (i === idx ? { ...it, unitCost: e.target.value } : it))
                    )
                  }
                  placeholder="Cost"
                  className="w-16 px-2 py-1.5 rounded-lg text-xs outline-none"
                  style={{
                    background: 'var(--hub-input-bg)',
                    border: '1px solid var(--hub-border)',
                    color: 'var(--hub-text)',
                  }}
                />
                <button
                  type="button"
                  onClick={() => setNewItems((prev) => prev.filter((_, i) => i !== idx))}
                  disabled={newItems.length === 1}
                  className="p-1 rounded disabled:opacity-30"
                  style={{ color: 'var(--hub-text-muted)' }}
                >
                  <Trash2 size={12} />
                </button>
              </div>
            ))}
            <button
              type="button"
              onClick={() =>
                setNewItems((prev) => [
                  ...prev,
                  { productId: '', quantityOrdered: '', unitCost: '' },
                ])
              }
              className="text-xs"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              + Add item
            </button>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                Cancel
              </button>
              <button
                onClick={handleCreatePO}
                disabled={createMut.isPending || !newSupplierId || !newPoNumber.trim()}
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: '#000' }}
              >
                {createMut.isPending ? 'Creating…' : 'Create Draft'}
              </button>
            </div>
          </div>
        </div>
      )}

      {(page > 1 || hasMore) && (
        <div className="flex items-center justify-between px-1 py-3">
          <span className="text-xs" style={{ color: 'var(--erp-text-dim)' }}>
            Page {page}
          </span>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1.5 rounded-md text-xs font-medium disabled:opacity-30"
              style={{ background: 'var(--erp-hover)', color: 'var(--erp-text-muted)' }}
            >
              Prev
            </button>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={!hasMore}
              className="px-3 py-1.5 rounded-md text-xs font-medium disabled:opacity-30"
              style={{ background: 'var(--erp-hover)', color: 'var(--erp-text-muted)' }}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Purchasing;
