import React, { useEffect, useState } from 'react';
import { motion } from 'motion/react';
import { Settings as SettingsIcon, Bell, Database } from 'lucide-react';
import { useSettings } from '../../shared/lib/settingsStore';
import type { ErpBusinessSettings } from '../../shared/lib/settingsStore';
import { useErpSettings, useUpdateErpSettings } from '../../shared/hooks/queries/useErpSettings';
import { useWarehouses } from '../warehouses/useWarehouses';
import { canErpPermission } from '../../shared/lib/erpPermissions';

const Section: React.FC<{
  title: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  children: React.ReactNode;
}> = ({ title, icon: Icon, children }) => (
  <div
    className="rounded-xl p-5 space-y-4"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <div className="flex items-center gap-2">
      <Icon size={14} style={{ color: 'var(--erp-accent)' }} />
      <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
        {title}
      </h3>
    </div>
    <div className="space-y-3">{children}</div>
  </div>
);

const ConfigField: React.FC<{ label: string; description?: string; children: React.ReactNode }> = ({
  label,
  description,
  children,
}) => (
  <div className="flex items-center justify-between gap-4">
    <div className="flex-1">
      <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
        {label}
      </p>
      {description && (
        <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
          {description}
        </p>
      )}
    </div>
    <div>{children}</div>
  </div>
);

const Toggle: React.FC<{
  value: boolean;
  onChange: (v: boolean) => void;
  disabled?: boolean;
}> = ({ value, onChange, disabled = false }) => (
  <button
    type="button"
    disabled={disabled}
    onClick={() => onChange(!value)}
    className="relative h-5 w-9 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50"
    style={{ background: value ? 'var(--erp-accent)' : 'var(--hub-hover)' }}
  >
    <span
      className="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-all"
      style={{ left: value ? '18px' : '2px' }}
    />
  </button>
);

const GeneralTab: React.FC = () => {
  const settings = useSettings();
  const canManageSettings = canErpPermission('erp:settings:manage');
  const { data: serverSettings, isError, isFetching, isLoading, refetch } = useErpSettings();
  const update = useUpdateErpSettings();
  const { data: warehouses = [] } = useWarehouses();
  const [form, setForm] = useState<ErpBusinessSettings>(() => ({
    currency: settings.currency,
    taxRate: settings.taxRate,
    lowStockThreshold: settings.lowStockThreshold,
    defaultWarehouseId: settings.defaultWarehouseId,
    notifyOnLowStock: settings.notifyOnLowStock,
    notifyOnExpiry: settings.notifyOnExpiry,
  }));

  useEffect(() => {
    if (serverSettings) setForm(serverSettings);
  }, [serverSettings]);

  if (isError && !serverSettings) {
    return (
      <>
        <div
          role="alert"
          className="rounded-xl p-5 mb-5 space-y-3"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <p className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
            Unable to load ERP settings
          </p>
          <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
            Business settings could not be loaded from the tenant API.
          </p>
          <button
            type="button"
            onClick={() => void refetch()}
            disabled={isFetching}
            className="rounded-lg px-3 py-1.5 text-xs font-semibold disabled:opacity-40"
            style={{ background: 'var(--dash-active)', color: 'var(--dash-text)' }}
          >
            {isFetching ? 'Retrying…' : 'Retry'}
          </button>
        </div>
      </>
    );
  }

  if (isLoading || !serverSettings) {
    return (
      <>
        <div className="rounded-xl p-5 mb-5 text-sm" style={{ color: 'var(--hub-text-muted)' }}>
          Loading ERP settings…
        </div>
      </>
    );
  }

  const updateField = <K extends keyof ErpBusinessSettings>(
    field: K,
    value: ErpBusinessSettings[K]
  ): void => {
    if (!canManageSettings) return;
    setForm((current) => ({ ...current, [field]: value }));
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <Section title="General" icon={SettingsIcon}>
        <ConfigField label="Currency" description="Default currency for pricing and reports">
          <select
            value={form.currency}
            disabled={!canManageSettings}
            onChange={(e) => updateField('currency', e.target.value)}
            className="px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            <option value="USD">USD ($)</option>
            <option value="EUR">EUR (€)</option>
            <option value="GBP">GBP (£)</option>
            <option value="JPY">JPY (¥)</option>
            <option value="CAD">CAD</option>
          </select>
        </ConfigField>
        <ConfigField label="Default Tax Rate" description="Sales tax percentage applied by default">
          <input
            type="number"
            step="0.01"
            value={form.taxRate}
            disabled={!canManageSettings}
            onChange={(e) => {
              const v = parseFloat(e.target.value);
              if (!isNaN(v) && v >= 0) updateField('taxRate', v);
            }}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </ConfigField>
        <ConfigField label="Default Warehouse" description="Used for new products and transfers">
          <select
            value={form.defaultWarehouseId}
            disabled={!canManageSettings}
            onChange={(e) => updateField('defaultWarehouseId', parseInt(e.target.value, 10))}
            className="px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            {warehouses.map((w) => (
              <option key={w.id} value={w.id}>
                {w.name}
              </option>
            ))}
          </select>
        </ConfigField>
      </Section>

      <Section title="Inventory" icon={Database}>
        <ConfigField
          label="Default Low Stock Threshold"
          description="Used when a product does not specify one"
        >
          <input
            type="number"
            value={form.lowStockThreshold}
            disabled={!canManageSettings}
            onChange={(e) => {
              const v = parseInt(e.target.value, 10);
              if (!isNaN(v) && v >= 0) updateField('lowStockThreshold', v);
            }}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </ConfigField>
      </Section>

      <Section title="Notifications" icon={Bell}>
        <ConfigField
          label="Low Stock Alerts"
          description="Get notified when products fall below threshold"
        >
          <Toggle
            value={form.notifyOnLowStock}
            onChange={(value) => updateField('notifyOnLowStock', value)}
            disabled={!canManageSettings}
          />
        </ConfigField>
        <ConfigField label="Expiry Alerts" description="Get notified when items approach expiry">
          <Toggle
            value={form.notifyOnExpiry}
            onChange={(value) => updateField('notifyOnExpiry', value)}
            disabled={!canManageSettings}
          />
        </ConfigField>
      </Section>

      {canManageSettings && (
        <div className="flex justify-end">
          <button
            type="button"
            onClick={() => update.mutate(form)}
            disabled={update.isPending}
            className="rounded-lg px-4 py-2 text-xs font-semibold disabled:opacity-40"
            style={{ background: 'var(--dash-active)', color: 'var(--dash-text)' }}
          >
            {update.isPending ? 'Saving…' : 'Save settings'}
          </button>
        </div>
      )}
    </motion.div>
  );
};

const Settings: React.FC = () => (
  <div className="portal-erp-settings">
    <GeneralTab />
  </div>
);
export default Settings;
