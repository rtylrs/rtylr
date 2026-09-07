import React, { useEffect, useId, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Search, Package, Users, ShoppingCart, LayoutDashboard, ArrowRight } from 'lucide-react';
import { useProducts } from '../../features/products/useProducts';
import { useSuppliers } from '../../features/suppliers/useSuppliers';
import { usePurchaseOrders } from '../../features/purchasing/usePurchaseOrders';
import { useDialogFocus } from '../hooks/useDialogFocus';

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
}

interface Result {
  type: 'product' | 'supplier' | 'order' | 'nav';
  label: string;
  sub?: string;
  to: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}

const NAV_ITEMS: Result[] = [
  { type: 'nav', label: 'Dashboard', to: '/', icon: LayoutDashboard },
  { type: 'nav', label: 'Inventory', to: '/inventory', icon: Package },
  { type: 'nav', label: 'Categories', to: '/categories', icon: Package },
  { type: 'nav', label: 'Brands', to: '/brands', icon: Package },
  { type: 'nav', label: 'Warehouses', to: '/warehouses', icon: Package },
  { type: 'nav', label: 'Stock Transfers', to: '/transfers', icon: Package },
  { type: 'nav', label: 'Stock Takes', to: '/stock-takes', icon: Package },
  { type: 'nav', label: 'Low Stock Alerts', to: '/alerts', icon: Package },
  { type: 'nav', label: 'Goods Receiving', to: '/receiving', icon: Package },
  { type: 'nav', label: 'Reorder Suggestions', to: '/reorder', icon: Package },
  { type: 'nav', label: 'Price Lists', to: '/price-lists', icon: Package },
  { type: 'nav', label: 'Product Variants', to: '/variants', icon: Package },
  { type: 'nav', label: 'Barcode Generator', to: '/barcodes', icon: Package },
  { type: 'nav', label: 'Expiry Tracking', to: '/expiry', icon: Package },
  { type: 'nav', label: 'Inventory Valuation', to: '/valuation', icon: Package },
  { type: 'nav', label: 'Purchasing', to: '/purchasing', icon: ShoppingCart },
  { type: 'nav', label: 'Suppliers', to: '/suppliers', icon: Users },
  { type: 'nav', label: 'Settings', to: '/settings', icon: Package },
];

const CommandPalette: React.FC<CommandPaletteProps> = ({ open, onClose }) => {
  const nav = useNavigate();
  const titleId = useId();
  const dialogRef = useDialogFocus<HTMLDivElement>({ open, onClose });
  const [q, setQ] = useState('');
  const [activeIdx, setActiveIdx] = useState(0);
  const { data: productResult } = useProducts();
  const { data: supplierResult } = useSuppliers();
  const products = useMemo(() => productResult?.products ?? [], [productResult]);
  const suppliers = useMemo(() => supplierResult?.suppliers ?? [], [supplierResult]);
  const { data: ordersResult } = usePurchaseOrders();
  const orders = useMemo(() => ordersResult?.purchaseOrders ?? [], [ordersResult]);
  const supplierNameById = useMemo(
    () => new Map(suppliers.map((supplier) => [supplier.id, supplier.name])),
    [suppliers]
  );

  useEffect(() => {
    if (!open) setQ('');
  }, [open]);

  const results = useMemo<Result[]>(() => {
    const qq = q.toLowerCase().trim();
    if (!qq) return NAV_ITEMS.slice(0, 8);

    const productResults: Result[] = products
      .filter(
        (p) => (p.name ?? '').toLowerCase().includes(qq) || (p.sku ?? '').toLowerCase().includes(qq)
      )
      .slice(0, 5)
      .map((p) => ({
        type: 'product',
        label: p.name ?? '',
        sub: p.sku,
        to: `/products/${p.id}`,
        icon: Package,
      }));

    const supplierResults: Result[] = (suppliers || [])
      .filter((s) => (s.name ?? '').toLowerCase().includes(qq))
      .slice(0, 3)
      .map((s) => ({
        type: 'supplier',
        label: s.name ?? '',
        sub: s.email,
        to: `/suppliers`,
        icon: Users,
      }));

    const orderResults: Result[] = (orders || [])
      .filter((o) => {
        const supplierName = supplierNameById.get(o.supplierId) ?? '';
        return String(o.id).includes(qq) || supplierName.toLowerCase().includes(qq);
      })
      .slice(0, 3)
      .map((o) => ({
        type: 'order',
        label: `PO-${o.id}`,
        sub: supplierNameById.get(o.supplierId) ?? `Supplier #${o.supplierId}`,
        to: `/purchasing`,
        icon: ShoppingCart,
      }));

    const navResults: Result[] = NAV_ITEMS.filter((n) => n.label.toLowerCase().includes(qq)).slice(
      0,
      4
    );

    return [...productResults, ...supplierResults, ...orderResults, ...navResults];
  }, [q, products, suppliers, orders, supplierNameById]);

  useEffect(() => {
    setActiveIdx(0);
  }, [q]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setActiveIdx((i) => Math.min(i + 1, results.length - 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setActiveIdx((i) => Math.max(i - 1, 0));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        const r = results[activeIdx];
        if (r) {
          nav({ to: r.to });
          onClose();
        }
      } else if (e.key === 'Escape') {
        onClose();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [open, results, activeIdx, nav, onClose]);

  if (!open) return null;

  return (
    <div
      ref={dialogRef}
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] px-4"
      style={{ background: 'var(--hub-overlay-medium)' }}
      tabIndex={-1}
      onClick={onClose}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="w-full max-w-xl rounded-xl overflow-hidden"
        style={{
          background: 'var(--hub-card)',
          border: '1px solid var(--hub-border)',
          boxShadow: 'var(--hub-shadow)',
        }}
      >
        <h2 id={titleId} className="sr-only">
          Command palette
        </h2>
        <div
          className="flex items-center gap-2 px-4 py-3"
          style={{ borderBottom: '1px solid var(--hub-border)' }}
        >
          <Search size={14} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            autoFocus
            aria-label="Search products, suppliers, orders, or pages"
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder="Search products, suppliers, orders, pages…"
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--hub-text)' }}
          />
          <kbd
            className="text-[10px] font-mono px-1.5 py-0.5 rounded"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}
          >
            esc
          </kbd>
        </div>
        <div className="max-h-96 overflow-y-auto p-2">
          {results.length === 0 ? (
            <p className="text-center py-6 text-xs" style={{ color: 'var(--hub-text-dim)' }}>
              No results for "{q}"
            </p>
          ) : (
            results.map((r, i) => {
              const Icon = r.icon;
              const active = i === activeIdx;
              return (
                <button
                  type="button"
                  key={`${r.type}-${r.label}-${i}`}
                  onClick={() => {
                    nav({ to: r.to });
                    onClose();
                  }}
                  onMouseEnter={() => setActiveIdx(i)}
                  className="w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left"
                  style={{ background: active ? 'var(--hub-hover)' : 'transparent' }}
                >
                  <Icon size={14} style={{ color: 'var(--erp-accent)' }} />
                  <div className="flex-1 min-w-0">
                    <p
                      className="text-sm font-medium truncate"
                      style={{ color: 'var(--hub-text)' }}
                    >
                      {r.label}
                    </p>
                    {r.sub && (
                      <p className="text-[11px] truncate" style={{ color: 'var(--hub-text-dim)' }}>
                        {r.sub}
                      </p>
                    )}
                  </div>
                  <span
                    className="text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded"
                    style={{ background: 'var(--hub-active)', color: 'var(--hub-text-dim)' }}
                  >
                    {r.type}
                  </span>
                  {active && <ArrowRight size={12} style={{ color: 'var(--hub-text-dim)' }} />}
                </button>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};

export default CommandPalette;
