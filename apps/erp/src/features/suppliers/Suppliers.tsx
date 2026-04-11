import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Plus, Mail, Phone, Package, BarChart3 } from 'lucide-react';
import { Supplier } from '../../shared/services/erpApi';
import { useSuppliers } from './useSuppliers';
import SupplierScorecard from '../../shared/components/SupplierScorecard';
import ExportButton from '../../shared/components/ExportButton';

const AMBER = '#f59e0b';

// ─── Helpers ──────────────────────────────────────────────────────────────────

const PALETTE = ['#10b981', AMBER, '#4583C4', '#8b5cf6', '#f87171', '#3b82f6'];

function getSupplierColor(supplier: Supplier, index: number): string {
  return PALETTE[index % PALETTE.length];
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

// ─── Animation variants ───────────────────────────────────────────────────────

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.1 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

// ─── Skeleton card ────────────────────────────────────────────────────────────

const SkeletonCard: React.FC = () => (
  <div
    className="rounded-xl p-5 flex flex-col gap-4"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <div className="flex items-start gap-3">
      <div
        className="h-10 w-10 rounded-xl animate-pulse"
        style={{ background: 'var(--hub-border)', opacity: 0.6 }}
      />
      <div className="flex-1 space-y-2">
        <div
          className="h-3 rounded animate-pulse w-3/4"
          style={{ background: 'var(--hub-border)', opacity: 0.6 }}
        />
        <div
          className="h-2.5 rounded animate-pulse w-1/2"
          style={{ background: 'var(--hub-border)', opacity: 0.4 }}
        />
      </div>
    </div>
    <div className="space-y-2">
      <div
        className="h-2.5 rounded animate-pulse"
        style={{ background: 'var(--hub-border)', opacity: 0.5 }}
      />
      <div
        className="h-2.5 rounded animate-pulse w-2/3"
        style={{ background: 'var(--hub-border)', opacity: 0.4 }}
      />
    </div>
  </div>
);

// ─── Component ────────────────────────────────────────────────────────────────

const Suppliers: React.FC = () => {
  const { data: suppliers = [], isLoading } = useSuppliers();
  const [showScorecards, setShowScorecards] = useState(false);

  const activeCount = suppliers.filter((s) => s.status === 'Active').length;

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-5 max-w-5xl"
    >
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Suppliers
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {isLoading ? '…' : `${activeCount} active supplier${activeCount !== 1 ? 's' : ''}`}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowScorecards((v) => !v)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium"
            style={{
              background: showScorecards ? 'rgba(245,158,11,0.12)' : 'var(--hub-hover)',
              color: showScorecards ? AMBER : 'var(--hub-text-muted)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <BarChart3 size={12} /> Scorecards
          </button>
          <ExportButton
            data={suppliers.map((s) => ({
              id: s.id,
              name: s.name,
              contact: s.contactName,
              email: s.email,
              phone: s.phone,
              products: s.productCount,
              status: s.status,
            }))}
            filename="suppliers.csv"
          />
          <button
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-opacity"
            style={{ background: AMBER, color: '#000' }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = '0.9';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.opacity = '1';
            }}
          >
            <Plus size={14} />
            Add Supplier
          </button>
        </div>
      </motion.div>

      {showScorecards && (
        <motion.div variants={itemVariants} className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          {suppliers.map((s) => (
            <SupplierScorecard key={s.id} supplier={s} />
          ))}
        </motion.div>
      )}

      {/* Supplier cards grid */}
      <motion.div
        variants={containerVariants}
        className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3"
      >
        {isLoading
          ? Array.from({ length: 3 }).map((_, i) => <SkeletonCard key={i} />)
          : suppliers.map((supplier, index) => {
              const color = getSupplierColor(supplier, index);
              const initials = getInitials(supplier.name);
              return (
                <motion.div
                  key={supplier.id}
                  variants={itemVariants}
                  className="rounded-xl p-5 flex flex-col gap-4 transition-colors cursor-pointer"
                  style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                  whileHover={{ scale: 1.015, transition: { duration: 0.15 } }}
                >
                  {/* Header row */}
                  <div className="flex items-start gap-3">
                    <span
                      className="flex items-center justify-center h-10 w-10 rounded-xl text-sm font-bold shrink-0"
                      style={{ background: `${color}18`, color }}
                    >
                      {initials}
                    </span>
                    <div className="flex-1 min-w-0">
                      <p
                        className="text-sm font-semibold truncate"
                        style={{ color: 'var(--hub-text)' }}
                      >
                        {supplier.name}
                      </p>
                      <p className="text-xs truncate" style={{ color: 'var(--hub-text-muted)' }}>
                        {supplier.contactName}
                      </p>
                    </div>
                    <span
                      className="text-[10px] font-medium px-2 py-0.5 rounded-full shrink-0"
                      style={{
                        background:
                          supplier.status === 'Active'
                            ? 'rgba(16,185,129,0.1)'
                            : 'var(--hub-hover)',
                        color: supplier.status === 'Active' ? '#10b981' : 'var(--hub-text-muted)',
                      }}
                    >
                      {supplier.status}
                    </span>
                  </div>

                  {/* Contact info */}
                  <div className="space-y-1.5">
                    <a
                      href={`mailto:${supplier.email}`}
                      className="flex items-center gap-2 text-xs no-underline transition-colors"
                      style={{ color: 'var(--hub-text-muted)' }}
                      onMouseEnter={(e) => {
                        (e.currentTarget as HTMLElement).style.color = color;
                      }}
                      onMouseLeave={(e) => {
                        (e.currentTarget as HTMLElement).style.color = 'var(--hub-text-muted)';
                      }}
                    >
                      <Mail size={12} style={{ flexShrink: 0 }} />
                      <span className="truncate">{supplier.email}</span>
                    </a>
                    <div
                      className="flex items-center gap-2 text-xs"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      <Phone size={12} style={{ flexShrink: 0 }} />
                      <span>{supplier.phone}</span>
                    </div>
                  </div>

                  {/* Divider */}
                  <div style={{ height: '1px', background: 'var(--hub-border)' }} />

                  {/* Product count + categories */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-1.5">
                      <Package size={12} style={{ color }} />
                      <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                        {supplier.productCount} products
                      </span>
                    </div>
                    <div className="flex gap-1 flex-wrap justify-end">
                      {(supplier.categories ?? []).map((cat) => (
                        <span
                          key={cat}
                          className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                          style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}
                        >
                          {cat}
                        </span>
                      ))}
                    </div>
                  </div>
                </motion.div>
              );
            })}
      </motion.div>
    </motion.div>
  );
};

export default Suppliers;
