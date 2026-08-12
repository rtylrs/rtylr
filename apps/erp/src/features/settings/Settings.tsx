import React, { useState } from 'react';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import {
  Settings as SettingsIcon,
  Bell,
  Database,
  Palette,
  Users,
  UserPlus,
  UserMinus,
} from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { useSettings } from '../../shared/lib/settingsStore';
import { useWarehouses } from '../warehouses/useWarehouses';
import {
  useErpTeam,
  useErpInviteMember,
  useErpDeactivateMember,
} from '../../shared/hooks/queries/useErpTeam';
import {
  useErpSubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useErpSubscription';

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

const Toggle: React.FC<{ value: boolean; onChange: (v: boolean) => void }> = ({
  value,
  onChange,
}) => (
  <button
    type="button"
    onClick={() => onChange(!value)}
    className="relative h-5 w-9 rounded-full transition-colors"
    style={{ background: value ? 'var(--erp-accent)' : 'var(--hub-hover)' }}
  >
    <span
      className="absolute top-0.5 h-4 w-4 rounded-full bg-white transition-all"
      style={{ left: value ? '18px' : '2px' }}
    />
  </button>
);

// ── General Tab ───────────────────────────────────────────────────────────────

const GeneralTab: React.FC = () => {
  const settings = useSettings();
  const { data: warehouses = [] } = useWarehouses();

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      className="w-full min-w-0 space-y-5"
    >
      <Section title="General" icon={SettingsIcon}>
        <ConfigField label="Currency" description="Default currency for pricing and reports">
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
        </ConfigField>
        <ConfigField label="Default Tax Rate" description="Sales tax percentage applied by default">
          <input
            type="number"
            step="0.01"
            value={settings.taxRate}
            onChange={(e) => {
              const v = parseFloat(e.target.value);
              if (!isNaN(v) && v >= 0) settings.setTaxRate(v);
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
        </ConfigField>
      </Section>

      <Section title="Inventory" icon={Database}>
        <ConfigField
          label="Default Low Stock Threshold"
          description="Used when a product does not specify one"
        >
          <input
            type="number"
            value={settings.lowStockThreshold}
            onChange={(e) => {
              const v = parseInt(e.target.value, 10);
              if (!isNaN(v) && v >= 0) settings.setLowStockThreshold(v);
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
          <Toggle value={settings.notifyOnLowStock} onChange={settings.setNotifyOnLowStock} />
        </ConfigField>
        <ConfigField label="Expiry Alerts" description="Get notified when items approach expiry">
          <Toggle value={settings.notifyOnExpiry} onChange={settings.setNotifyOnExpiry} />
        </ConfigField>
      </Section>

      <Section title="Appearance" icon={Palette}>
        <ConfigField label="Theme" description="Light, dark, or follow system">
          <div className="flex gap-1">
            {(['light', 'dark', 'system'] as const).map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => settings.setTheme(t)}
                className="px-2 py-1 rounded-md text-[11px] font-medium"
                style={{
                  background: settings.theme === t ? 'var(--hub-active)' : 'var(--hub-hover)',
                  color: settings.theme === t ? 'var(--erp-accent)' : 'var(--hub-text-muted)',
                }}
              >
                {t}
              </button>
            ))}
          </div>
        </ConfigField>
      </Section>
    </motion.div>
  );
};

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { data: members = [], isLoading } = useErpTeam();
  const invite = useErpInviteMember();
  const deactivate = useErpDeactivateMember();
  const { data: sub } = useErpSubscription();

  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [showInvite, setShowInvite] = useState(false);

  const handleInvite = (): void => {
    if (!email.trim() || !fullName.trim()) {
      toast.error('Email and full name are required');
      return;
    }
    invite.mutate(
      { email: email.trim(), fullName: fullName.trim(), role: 'member' },
      {
        onSuccess: () => {
          setEmail('');
          setFullName('');
          setShowInvite(false);
        },
      }
    );
  };

  const inputCls = 'w-full rounded-lg px-3 py-2 text-[13px] outline-none';
  const inputSt = {
    background: 'var(--hub-input-bg)',
    border: '1px solid var(--hub-border)',
    color: 'var(--hub-text)',
  };

  return (
    <div className="w-full min-w-0 space-y-5">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--hub-text)' }}>
          Team
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--hub-text-muted)' }}>
          People who have access to ERP.
        </p>
      </div>

      {showInvite ? (
        <div
          className="rounded-xl p-4 space-y-3"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <h3 className="text-[13px] font-semibold" style={{ color: 'var(--hub-text)' }}>
            Invite someone to ERP
          </h3>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                className="block text-[11px] font-medium mb-1"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                Full name
              </label>
              <input
                type="text"
                placeholder="Jane Smith"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                className={inputCls}
                style={inputSt}
              />
            </div>
            <div>
              <label
                className="block text-[11px] font-medium mb-1"
                style={{ color: 'var(--hub-text-muted)' }}
              >
                Email
              </label>
              <input
                type="email"
                placeholder="jane@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleInvite();
                }}
                className={inputCls}
                style={inputSt}
              />
            </div>
          </div>
          {sub &&
            (sub.pricePerSeatCents ?? 0) > 0 &&
            (() => {
              const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                computeSeatProration(sub.pricePerSeatCents ?? 0);
              return (
                <p className="text-[11px]" style={{ color: 'var(--hub-text-muted)' }}>
                  Adding this member costs{' '}
                  <span className="font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(proratedCents)}
                  </span>{' '}
                  this month ({daysRemaining}/{daysInMonth} days), then{' '}
                  <span className="font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(sub.pricePerSeatCents ?? 0)}/mo
                  </span>{' '}
                  from {nextMonthName} 1st.
                </p>
              );
            })()}
          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={!email.trim() || !fullName.trim() || invite.isPending}
              onClick={handleInvite}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold disabled:opacity-40"
              style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
            >
              <UserPlus size={12} />
              {invite.isPending ? 'Sending…' : 'Send invite'}
            </button>
            <button
              type="button"
              onClick={() => setShowInvite(false)}
              className="px-3 py-1.5 rounded-lg text-[12px] font-medium"
              style={{ color: 'var(--hub-text-muted)' }}
            >
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => setShowInvite(true)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-semibold"
          style={{ background: 'var(--erp-accent)', color: 'var(--erp-accent-foreground)' }}
        >
          <UserPlus size={12} />
          Invite member
        </button>
      )}

      {isLoading ? (
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-14 rounded-xl animate-pulse"
              style={{ background: 'var(--hub-hover)' }}
            />
          ))}
        </div>
      ) : members.length === 0 ? (
        <div className="py-8 text-center" style={{ color: 'var(--hub-text-muted)' }}>
          <Users size={24} className="mx-auto mb-2 opacity-40" />
          <p className="text-[13px]">No team members yet.</p>
        </div>
      ) : (
        <div className="rounded-xl divide-y" style={{ border: '1px solid var(--hub-border)' }}>
          {members.map((m) => (
            <div key={m.id} className="flex items-center gap-3 px-4 py-3">
              <div
                className="h-8 w-8 rounded-full flex items-center justify-center text-[12px] font-semibold shrink-0"
                style={{
                  background: 'color-mix(in srgb, var(--erp-accent) 20%, transparent)',
                  color: 'var(--erp-accent)',
                }}
              >
                {m.name.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1 min-w-0">
                <p
                  className="text-[13px] font-medium truncate"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {m.name}
                </p>
                <p className="text-[11px] truncate" style={{ color: 'var(--hub-text-muted)' }}>
                  {m.email}
                </p>
              </div>
              <span
                className="text-[10px] font-semibold px-2 py-0.5 rounded-full shrink-0"
                style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
              >
                {m.role}
              </span>
              <button
                type="button"
                aria-label={`Remove ${m.name}`}
                onClick={() => {
                  deactivate.mutate(m.id);
                }}
                className="h-7 w-7 flex items-center justify-center rounded hover:bg-[var(--hub-hover)] transition-colors shrink-0"
                style={{
                  border: 'none',
                  background: 'transparent',
                  cursor: 'pointer',
                  color: 'var(--hub-text-muted)',
                }}
              >
                <UserMinus size={13} />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

// ── Main Settings Component ───────────────────────────────────────────────────

const Settings: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabId>('general');

  return (
    <div className="space-y-0">
      <HelmetContainer title="Settings · rtylr" />
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
