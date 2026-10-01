import React, { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { APP_URLS } from '@voxire/config';
import { useTranslation } from 'react-i18next';
import { Settings as SettingsIcon, Bell, Database, Palette, Users } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settingsStore';
import type { ErpBusinessSettings } from '../../shared/lib/settingsStore';
import { useErpSettings, useUpdateErpSettings } from '../../shared/hooks/queries/useErpSettings';
import { useWarehouses } from '../warehouses/useWarehouses';
import { canErpPermission } from '../../shared/lib/erpPermissions';
import { pageTitle } from '../../shared/lib/pageTitle';
import { routeLabels } from '../../shared/lib/routeLabels';

type TabId = 'general' | 'team';

interface TabDef {
  id: TabId;
  label: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}

const TABS: TabDef[] = [
  { id: 'general', label: 'General', icon: SettingsIcon },
  { id: 'team', label: 'Team', icon: Users },
];

// ── Shared primitives ─────────────────────────────────────────────────────────

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

const AppearanceSection: React.FC = () => {
  const theme = useSettings((state) => state.theme);
  const setTheme = useSettings((state) => state.setTheme);
  return (
    <Section title="Appearance" icon={Palette}>
      <ConfigField label="Theme" description="Light, dark, or follow system">
        <div className="flex gap-1">
          {(['light', 'dark', 'system'] as const).map((option) => (
            <button
              key={option}
              type="button"
              onClick={() => setTheme(option)}
              className="px-2 py-1 rounded-md text-[11px] font-medium"
              style={{
                background: theme === option ? 'var(--hub-active)' : 'var(--hub-hover)',
                color: theme === option ? 'var(--erp-accent)' : 'var(--hub-text-muted)',
              }}
            >
              {option}
            </button>
          ))}
        </div>
      </ConfigField>
    </Section>
  );
};

// ── General Tab ───────────────────────────────────────────────────────────────

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
            style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
          >
            {isFetching ? 'Retrying…' : 'Retry'}
          </button>
        </div>
        <AppearanceSection />
      </>
    );
  }

  if (isLoading || !serverSettings) {
    return (
      <>
        <div className="rounded-xl p-5 mb-5 text-sm" style={{ color: 'var(--hub-text-muted)' }}>
          Loading ERP settings…
        </div>
        <AppearanceSection />
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

      <AppearanceSection />

      {canManageSettings && (
        <div className="flex justify-end">
          <button
            type="button"
            onClick={() => update.mutate(form)}
            disabled={update.isPending}
            className="rounded-lg px-4 py-2 text-xs font-semibold disabled:opacity-40"
            style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
          >
            {update.isPending ? 'Saving…' : 'Save settings'}
          </button>
        </div>
      )}
    </motion.div>
  );
};

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { t } = useTranslation('erp');
  return (
    <div className="w-full min-w-0 space-y-5">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--hub-text)' }}>
          {t('erp.settings.team.managementTitle')}
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--hub-text-muted)' }}>
          {t('erp.settings.team.managementDescription')}
        </p>
      </div>
      <a
        href={`${APP_URLS.dash}/team`}
        className="inline-flex items-center rounded-lg px-3 py-2 text-[12px] font-semibold"
        style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
      >
        {t('erp.settings.team.managementAction')}
      </a>
    </div>
  );
};

// ── Main Settings Component ───────────────────────────────────────────────────

const Settings: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabId>('general');

  return (
    <div className="space-y-0">
      <HelmetContainer title={pageTitle(t(routeLabels['/settings'].title))} />
      <PageHeader title="Settings" description="Configure ERP defaults" />

      {/* Tab bar */}
      <div
        className="flex"
        style={{ borderBottom: '1px solid var(--hub-border)', marginBottom: 20 }}
      >
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className="flex items-center gap-1.5 px-4 py-2.5 text-[13px] transition-colors"
              style={{
                fontWeight: isActive ? 600 : 500,
                color: isActive ? 'var(--erp-accent)' : 'var(--hub-text-muted)',
                background: 'transparent',
                border: 'none',
                borderBottom: isActive ? '2px solid var(--erp-accent)' : '2px solid transparent',
                marginBottom: -1,
                cursor: 'pointer',
              }}
            >
              <Icon size={14} />
              {tab.label}
            </button>
          );
        })}
      </div>

      <motion.div
        key={activeTab}
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.15 }}
      >
        {activeTab === 'general' && <GeneralTab />}
        {activeTab === 'team' && <TeamTab />}
      </motion.div>
    </div>
  );
};

export default Settings;
