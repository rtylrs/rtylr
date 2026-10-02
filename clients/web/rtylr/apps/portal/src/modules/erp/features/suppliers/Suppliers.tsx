import React, { useState } from 'react';
import { motion, type Variants } from 'motion/react';
import { Plus, Mail, Phone, MapPin, BarChart3 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../../../shared/stores/userStore';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { Supplier } from './useSuppliers';
import { useSuppliers, useCreateSupplier } from './useSuppliers';
import SupplierScorecard from '../../shared/components/SupplierScorecard';
import SupplierBillHistory from '../../shared/components/SupplierBillHistory';
import ExportButton from '../../shared/components/ExportButton';
import PaginationControls from '../../shared/components/PaginationControls';
import { canErpPermission } from '../../shared/lib/erpPermissions';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

const AMBER = 'var(--erp-accent)';

// ─── Helpers ──────────────────────────────────────────────────────────────────

const PALETTE = [
  'var(--erp-success)',
  AMBER,
  'var(--accent)',
  'var(--erp-violet)',
  'var(--erp-danger)',
  'var(--erp-blue)',
];

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

function toExportRows(suppliers: Supplier[]): Array<Record<string, unknown>> {
  return suppliers.map((supplier) => ({
    id: supplier.id,
    name: supplier.name,
    email: supplier.email,
    phone: supplier.phone,
    city: supplier.city,
    country: supplier.country,
    status: supplier.isActive ? 'Active' : 'Inactive',
  }));
}

// ─── Animation variants ───────────────────────────────────────────────────────

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.1 } },
};

const itemVariants: Variants = {
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
  const { t } = useTranslation();
  const [page, setPage] = useState(1);
  const pageSize = 50;
  const pageQuery = useSuppliers({ page, pageSize });
  const allSuppliersQuery = useSuppliers();
  const {
    data: supplierResult,
    isLoading: pageLoading,
    isPending: pagePending,
    isError: pageError,
    isPaused: pagePaused,
    failureCount: pageFailureCount,
    refetch: refetchPage,
  } = pageQuery;
  const {
    data: allSupplierResult,
    isLoading: allLoading,
    isPending: allPending,
    isError: allError,
    isPaused: allPaused,
    failureCount: allFailureCount,
    refetch: refetchAll,
  } = allSuppliersQuery;
  const suppliers = supplierResult?.suppliers ?? [];
  const allSuppliers = allSupplierResult?.suppliers ?? [];
  const hasMore = supplierResult?.hasMore ?? false;
  const isLoading = pagePending || pageLoading || allPending || allLoading;
  const loadFailed =
    pageError === true ||
    allError === true ||
    (pagePaused === true && pageFailureCount > 0) ||
    (allPaused === true && allFailureCount > 0);
  const createMut = useCreateSupplier();
  const canManageSuppliers = canErpPermission('erp:suppliers:create');
  const [showScorecards, setShowScorecards] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [newName, setNewName] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newPhone, setNewPhone] = useState('');
  const [newCity, setNewCity] = useState('');
  const [newCountry, setNewCountry] = useState('');
  const [newAddress, setNewAddress] = useState('');

  const activeCount = allSuppliers.filter((s) => s.isActive).length;

  const handleCreate = async (): Promise<void> => {
    if (!newName.trim() || !newEmail.trim() || !newCity.trim() || !newCountry.trim()) {
      toast.error(t('erp.suppliers.requiredFields'));
      return;
    }
    try {
      await createMut.mutateAsync({
        name: newName.trim(),
        email: newEmail.trim(),
        phone: newPhone.trim(),
        address: newAddress.trim(),
        city: newCity.trim(),
        country: newCountry.trim(),
        taxId: '',
        paymentTerms: '',
        notes: '',
        isActive: true,
      });
      toast.success(t('erp.suppliers.created'));
      setShowCreate(false);
      setNewName('');
      setNewEmail('');
      setNewPhone('');
      setNewCity('');
      setNewCountry('');
      setNewAddress('');
    } catch {
      toast.error(t('erp.suppliers.createFailed'));
    }
  };

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="w-full min-w-0 space-y-5"
    >
      <HelmetContainer title={pageTitle(t(routeLabels['/erp/suppliers'].title))} />
      {/* Header */}
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t(routeLabels['/erp/suppliers'].title)}
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {isLoading
              ? '…'
              : loadFailed
                ? t('erpErrors.generic')
                : t('erp.suppliers.activeCount', { count: activeCount })}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setShowScorecards((v) => !v)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-medium"
            style={{
              background: showScorecards
                ? 'color-mix(in srgb, var(--erp-accent) 12%, transparent)'
                : 'var(--hub-hover)',
              color: showScorecards ? AMBER : 'var(--hub-text-muted)',
              border: '1px solid var(--hub-border)',
            }}
          >
            <BarChart3 size={12} /> Scorecards
          </button>
          <ExportButton
            data={toExportRows(allSuppliers)}
            getData={async () => {
              if (allSupplierResult === undefined) {
                throw new Error('Supplier export is still loading');
              }
              return toExportRows(allSuppliers);
            }}
            filename="suppliers.csv"
          />
          {canManageSuppliers && (
            <button
              type="button"
              data-testid="suppliers-add-button"
              onClick={() => setShowCreate(true)}
              className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-opacity"
              style={{ background: AMBER, color: 'var(--dash-text)' }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = '0.9';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.opacity = '1';
              }}
            >
              <Plus size={14} />
              {t('erp.suppliers.add')}
            </button>
          )}
        </div>
      </motion.div>

      {showScorecards && !loadFailed && (
        <motion.div variants={itemVariants} className="grid grid-cols-1 lg:grid-cols-2 gap-3">
          {suppliers.map((s) => (
            <div key={s.id}>
              <SupplierScorecard supplier={s} />
              <SupplierBillHistory supplier={s} />
            </div>
          ))}
        </motion.div>
      )}

      {loadFailed && (
        <div
          role="alert"
          className="flex flex-col items-center gap-3 rounded-xl p-8 text-center"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
            {t('erpErrors.generic')}
          </p>
          <button
            type="button"
            onClick={() => {
              void refetchPage();
              void refetchAll();
            }}
            className="rounded-lg px-3 py-2 text-xs font-medium"
            style={{ background: 'var(--hub-hover)', color: 'var(--hub-text)' }}
          >
            {t('erp.suppliers.retry', { defaultValue: 'Retry' })}
          </button>
        </div>
      )}

      {/* Supplier cards grid */}
      <motion.div
        variants={containerVariants}
        className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3"
      >
        {loadFailed
          ? null
          : isLoading
            ? Array.from({ length: 3 }).map((_, i) => <SkeletonCard key={i} />)
            : suppliers.map((supplier, index) => {
                const color = getSupplierColor(supplier, index);
                const initials = getInitials(supplier.name ?? '');
                return (
                  <motion.div
                    key={supplier.id}
                    variants={itemVariants}
                    className="rounded-xl p-5 flex flex-col gap-4 transition-colors"
                    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                  >
                    {/* Header row */}
                    <div className="flex items-start gap-3">
                      <span
                        className="flex items-center justify-center h-10 w-10 rounded-xl text-sm font-bold shrink-0"
                        style={{
                          background: `color-mix(in srgb, ${color} 9%, transparent)`,
                          color,
                        }}
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
                          {[supplier.city, supplier.country].filter(Boolean).join(', ')}
                        </p>
                      </div>
                      <span
                        className="text-[10px] font-medium px-2 py-0.5 rounded-full shrink-0"
                        style={{
                          background: supplier.isActive
                            ? 'color-mix(in srgb, var(--erp-success) 10%, transparent)'
                            : 'var(--hub-hover)',
                          color: supplier.isActive ? 'var(--erp-success)' : 'var(--hub-text-muted)',
                        }}
                      >
                        {supplier.isActive ? 'Active' : 'Inactive'}
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
                        <MapPin size={12} style={{ color }} />
                        <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                          {supplier.address ||
                            [supplier.city, supplier.country].filter(Boolean).join(', ') ||
                            'No address'}
                        </span>
                      </div>
                      <span
                        className="text-[10px] px-1.5 py-0.5 rounded-md font-medium"
                        style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}
                      >
                        {supplier.paymentTerms || 'No terms'}
                      </span>
                    </div>
                  </motion.div>
                );
              })}
      </motion.div>

      <PaginationControls
        page={page}
        hasMore={hasMore}
        itemCount={suppliers.length}
        onPageChange={setPage}
      />

      {showCreate && (
        <div
          className="fixed inset-0 flex items-center justify-center"
          style={{ background: 'var(--hub-overlay)', zIndex: 9999 }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setShowCreate(false);
          }}
        >
          <div
            className="rounded-xl p-6 space-y-4 w-80"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <h3 className="text-sm font-bold" style={{ color: 'var(--hub-text)' }}>
              {t('erp.suppliers.newSupplier')}
            </h3>
            {[
              { value: newName, setter: setNewName, placeholderKey: 'erp.suppliers.companyName' },
              { value: newEmail, setter: setNewEmail, placeholderKey: 'erp.suppliers.email' },
              { value: newPhone, setter: setNewPhone, placeholderKey: 'erp.suppliers.phone' },
              { value: newCity, setter: setNewCity, placeholderKey: 'erp.suppliers.city' },
              { value: newCountry, setter: setNewCountry, placeholderKey: 'erp.suppliers.country' },
              { value: newAddress, setter: setNewAddress, placeholderKey: 'erp.suppliers.address' },
            ].map(({ value, setter, placeholderKey }) => (
              <input
                key={placeholderKey}
                value={value}
                onChange={(e) => setter(e.target.value)}
                placeholder={t(placeholderKey)}
                className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                style={{
                  background: 'var(--hub-input-bg)',
                  border: '1px solid var(--hub-border)',
                  color: 'var(--hub-text)',
                }}
              />
            ))}
            <div className="flex gap-2 justify-end">
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-3 py-2 rounded-lg text-xs"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                {t('erpCommon.cancel')}
              </button>
              <button
                type="button"
                onClick={handleCreate}
                disabled={
                  createMut.isPending ||
                  !newName.trim() ||
                  !newEmail.trim() ||
                  !newCity.trim() ||
                  !newCountry.trim()
                }
                className="px-3 py-2 rounded-lg text-xs font-semibold disabled:opacity-40"
                style={{ background: AMBER, color: 'var(--dash-text)' }}
              >
                {createMut.isPending ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </motion.div>
  );
};

export default Suppliers;
