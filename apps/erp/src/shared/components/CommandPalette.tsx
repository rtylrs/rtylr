import React, { useEffect, useId, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Search, Package, Users, ShoppingCart, LayoutDashboard, ArrowRight } from 'lucide-react';
import { useProducts } from '../../features/products/useProducts';
import { useSuppliers } from '../../features/suppliers/useSuppliers';
import { usePurchaseOrders } from '../../features/purchasing/usePurchaseOrders';
import { useDialogFocus } from '../hooks/useDialogFocus';
import { useTranslation } from 'react-i18next';
import { routeLabels } from '../lib/routeLabels';

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

/**
 * The palette lists every route by its full name, the same vocabulary the page
 * title uses. Labels are keys here and are translated where the list is built,
 * because the query filters on what the reader sees: matching the key would
 * search English no matter the language.
 */
interface NavEntry {
  labelKey: string;
  to: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}

const NAV_ITEMS: readonly NavEntry[] = [
  { labelKey: routeLabels['/'].title, to: '/', icon: LayoutDashboard },
  { labelKey: routeLabels['/inventory'].title, to: '/inventory', icon: Package },
  { labelKey: routeLabels['/categories'].title, to: '/categories', icon: Package },
  { labelKey: routeLabels['/brands'].title, to: '/brands', icon: Package },
  { labelKey: routeLabels['/warehouses'].title, to: '/warehouses', icon: Package },
  { labelKey: routeLabels['/transfers'].title, to: '/transfers', icon: Package },
  { labelKey: routeLabels['/stock-takes'].title, to: '/stock-takes', icon: Package },
  { labelKey: routeLabels['/alerts'].title, to: '/alerts', icon: Package },
  { labelKey: routeLabels['/receiving'].title, to: '/receiving', icon: Package },
  { labelKey: routeLabels['/reorder'].title, to: '/reorder', icon: Package },
  { labelKey: routeLabels['/price-lists'].title, to: '/price-lists', icon: Package },
  { labelKey: routeLabels['/variants'].title, to: '/variants', icon: Package },
  { labelKey: routeLabels['/barcodes'].title, to: '/barcodes', icon: Package },
  { labelKey: routeLabels['/expiry'].title, to: '/expiry', icon: Package },
  { labelKey: routeLabels['/valuation'].title, to: '/valuation', icon: Package },
  { labelKey: routeLabels['/purchasing'].title, to: '/purchasing', icon: ShoppingCart },
  { labelKey: routeLabels['/suppliers'].title, to: '/suppliers', icon: Users },
  { labelKey: routeLabels['/settings'].title, to: '/settings', icon: Package },
];

const CommandPalette: React.FC<CommandPaletteProps> = ({ open, onClose }) => {
  const { t } = useTranslation();
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

  // Translated once, because both the empty query and the filtered query
  // need the reader's language: filtering on the key would search English.
  const navEntries = useMemo<Result[]>(
    () =>
      NAV_ITEMS.map((n) => ({
        type: 'nav' as const,
        label: t(n.labelKey),
        to: n.to,
        icon: n.icon,
      })),
    [t]
  );

  const results = useMemo<Result[]>(() => {
    const qq = q.toLowerCase().trim();
    if (!qq) return navEntries.slice(0, 8);

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

    const navResults: Result[] = navEntries
      .filter((n) => n.label.toLowerCase().includes(qq))
      .slice(0, 4);

    return [...productResults, ...supplierResults, ...orderResults, ...navResults];
  }, [q, products, suppliers, orders, supplierNameById, navEntries]);

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
