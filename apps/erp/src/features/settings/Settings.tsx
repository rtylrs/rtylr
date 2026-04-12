import React from 'react';
import { motion } from 'motion/react';
import { Settings as SettingsIcon, Bell, Database, Palette } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settingsStore';
import { useWarehouses } from '../warehouses/useWarehouses';
import { toast } from 'sonner';

const AMBER = '#f59e0b';

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
      <Icon size={14} style={{ color: AMBER }} />
      <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
        {title}
      </h3>
    </div>
    <div className="space-y-3">{children}</div>
  </div>
);

const Field: React.FC<{ label: string; description?: string; children: React.ReactNode }> = ({
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

const Toggle: React.FC<{ value: boolean; onChange: (v: boolean) => void }> = ({
  value,
  onChange,
}) => (
  <button
    onClick={() => onChange(!value)}
    className="relative h-5 w-9 rounded-full transition-colors"
    style={{ background: value ? AMBER : 'var(--hub-hover)' }}
  >
    <span
      className="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-all"
      style={{ left: value ? '18px' : '2px' }}
    />
  </button>
);

const Settings: React.FC = () => {
  const settings = useSettings();
  const { data: warehouses = [] } = useWarehouses();

  const save = () => toast.success('Settings saved');

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-5 max-w-3xl"
    >
      <PageHeader title="Settings" description="Configure ERP defaults" />

      <Section title="General" icon={SettingsIcon}>
        <Field label="Currency" description="Default currency for pricing and reports">
          <select
            value={settings.currency}
            onChange={(e) => settings.setCurrency(e.target.value)}
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
        </Field>
        <Field label="Default Tax Rate" description="Sales tax percentage applied by default">
          <input
            type="number"
            step="0.01"
            value={settings.taxRate}
            onChange={(e) => settings.setTaxRate(parseFloat(e.target.value) || 0)}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </Field>
        <Field label="Default Warehouse" description="Used for new products and transfers">
          <select
            value={settings.defaultWarehouseId}
            onChange={(e) => settings.setDefaultWarehouseId(parseInt(e.target.value))}
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
        </Field>
      </Section>

      <Section title="Inventory" icon={Database}>
        <Field
          label="Default Low Stock Threshold"
          description="Used when a product does not specify one"
        >
          <input
            type="number"
            value={settings.lowStockThreshold}
            onChange={(e) => settings.setLowStockThreshold(parseInt(e.target.value) || 0)}
            className="w-20 px-2 py-1.5 rounded-lg text-xs outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </Field>
      </Section>

      <Section title="Notifications" icon={Bell}>
        <Field
          label="Low Stock Alerts"
          description="Get notified when products fall below threshold"
        >
          <Toggle value={settings.notifyOnLowStock} onChange={settings.setNotifyOnLowStock} />
        </Field>
        <Field label="Expiry Alerts" description="Get notified when items approach expiry">
          <Toggle value={settings.notifyOnExpiry} onChange={settings.setNotifyOnExpiry} />
        </Field>
      </Section>

      <Section title="Appearance" icon={Palette}>
        <Field label="Theme" description="Light, dark, or follow system">
          <div className="flex gap-1">
            {(['light', 'dark', 'system'] as const).map((t) => (
              <button
                key={t}
                onClick={() => settings.setTheme(t)}
                className="px-2 py-1 rounded-md text-[11px] font-medium"
                style={{
                  background: settings.theme === t ? 'var(--hub-active)' : 'var(--hub-hover)',
                  color: settings.theme === t ? AMBER : 'var(--hub-text-muted)',
                }}
              >
                {t}
              </button>
            ))}
          </div>
        </Field>
      </Section>

      <div className="flex justify-end">
        <button
          onClick={save}
          className="px-4 py-2 rounded-lg text-sm font-semibold"
          style={{ background: AMBER, color: '#000' }}
        >
          Save Settings
        </button>
      </div>
    </motion.div>
  );
};

export default Settings;
