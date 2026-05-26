import React, { useState } from 'react';
import { toast } from 'sonner';
import { motion } from 'motion/react';
import {
  Settings as SettingsIcon,
  Bell,
  Database,
  Palette,
  Users,
  CreditCard,
  UserPlus,
  UserMinus,
  Clock,
  CheckCircle2,
  XCircle,
  AlertCircle,
} from 'lucide-react';
import appConfig from '../../app.json';
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
  useErpSubscriptionBilling,
  centsToDisplay,
  trialDaysLeft,
  formatPeriodRange,
  computeSeatProration,
} from '../../shared/hooks/queries/useErpSubscription';

type TabId = 'general' | 'team' | 'billing';

interface TabDef {
  id: TabId;
  label: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}

const TABS: TabDef[] = [
  { id: 'general', label: 'General', icon: SettingsIcon },
  { id: 'team', label: 'Team', icon: Users },
  { id: 'billing', label: 'Billing', icon: CreditCard },
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
      <Icon size={14} style={{ color: 'var(--erp-accent, #f59e0b)' }} />
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
    style={{ background: value ? 'var(--erp-accent, #f59e0b)' : 'var(--hub-hover)' }}
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
      className="space-y-5 max-w-3xl"
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
            onChange={(e) => settings.setLowStockThreshold(parseInt(e.target.value) || 0)}
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
                  color:
                    settings.theme === t ? 'var(--erp-accent, #f59e0b)' : 'var(--hub-text-muted)',
                }}
              >
                {t}
              </button>
            ))}
          </div>
        </ConfigField>
      </Section>

      <Section title="About" icon={SettingsIcon}>
        {[
          { label: 'Application', value: appConfig.name },
          { label: 'Version', value: `v${appConfig.version}` },
        ].map(({ label, value }) => (
          <ConfigField key={label} label={label}>
            <span className="text-xs font-mono" style={{ color: 'var(--hub-text)' }}>
              {value}
            </span>
          </ConfigField>
        ))}
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
    <div className="space-y-5 max-w-2xl">
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
            sub.pricePerSeatCents > 0 &&
            (() => {
              const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                computeSeatProration(sub.pricePerSeatCents);
              return (
                <p className="text-[11px]" style={{ color: 'var(--hub-text-muted)' }}>
                  Adding this member costs{' '}
                  <span className="font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(proratedCents)}
                  </span>{' '}
                  this month ({daysRemaining}/{daysInMonth} days), then{' '}
                  <span className="font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(sub.pricePerSeatCents)}/mo
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
              style={{ background: 'var(--erp-accent, #f59e0b)', color: '#000' }}
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
          style={{ background: 'var(--erp-accent, #f59e0b)', color: '#000' }}
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
                  background: 'color-mix(in srgb, var(--erp-accent, #f59e0b) 20%, transparent)',
                  color: 'var(--erp-accent, #f59e0b)',
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

// ── Billing Tab ───────────────────────────────────────────────────────────────

const STATUS_ICON: Record<string, React.ComponentType<{ size?: number }>> = {
  active: CheckCircle2,
  trial: Clock,
  cancelled: XCircle,
  expired: AlertCircle,
};

const STATUS_COLOR: Record<string, string> = {
  active: 'var(--status-active, #16a34a)',
  trial: 'var(--status-trial, #d97706)',
  cancelled: 'var(--status-cancelled, #dc2626)',
  expired: 'var(--status-expired, #6b7280)',
};

const BillingTab: React.FC = () => {
  const { data: sub, isLoading: subLoading } = useErpSubscription();
  const { data: billing, isLoading: billingLoading } = useErpSubscriptionBilling();

  const isLoading = subLoading || billingLoading;

  if (isLoading) {
    return (
      <div className="space-y-4 max-w-2xl">
        <div
          className="h-6 w-32 rounded animate-pulse"
          style={{ background: 'var(--hub-hover)' }}
        />
        <div className="h-28 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
        <div className="h-20 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
      </div>
    );
  }

  const StatusIcon = sub ? (STATUS_ICON[sub.status] ?? AlertCircle) : null;
  const statusColor = sub ? (STATUS_COLOR[sub.status] ?? STATUS_COLOR['expired']) : '';

  return (
    <div className="space-y-5 max-w-2xl">
      <div>
        <h2 className="text-[15px] font-semibold mb-1" style={{ color: 'var(--hub-text)' }}>
          Billing
        </h2>
        <p className="text-[12px]" style={{ color: 'var(--hub-text-muted)' }}>
          Your ERP subscription and usage this month.
        </p>
      </div>

      {!sub ? (
        <div
          className="rounded-xl p-6 text-center space-y-3"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <CreditCard
            size={24}
            className="mx-auto opacity-30"
            style={{ color: 'var(--hub-text)' }}
          />
          <p className="text-[13px] font-medium" style={{ color: 'var(--hub-text)' }}>
            No active subscription
          </p>
          <p className="text-[12px]" style={{ color: 'var(--hub-text-muted)' }}>
            Contact your account owner or visit billing to activate ERP.
          </p>
        </div>
      ) : (
        <>
          <div
            className="rounded-xl p-4 space-y-3"
            style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
          >
            <div className="flex items-center justify-between">
              <p className="text-[14px] font-semibold" style={{ color: 'var(--hub-text)' }}>
                {sub.planName}
              </p>
              {StatusIcon && (
                <div
                  className="flex items-center gap-1.5 text-[11px] font-semibold"
                  style={{ color: statusColor }}
                >
                  <StatusIcon size={12} />
                  <span className="capitalize">{sub.status}</span>
                </div>
              )}
            </div>

            <div className="flex items-end justify-between">
              <div>
                <p className="text-[22px] font-bold" style={{ color: 'var(--hub-text)' }}>
                  {centsToDisplay(billing?.totalCents ?? sub.seatCount * sub.pricePerSeatCents)}
                </p>
                <p className="text-[11px]" style={{ color: 'var(--hub-text-muted)' }}>
                  {sub.seatCount} seat{sub.seatCount !== 1 ? 's' : ''} ×{' '}
                  {centsToDisplay(sub.pricePerSeatCents)}/seat · billed 1st of each month
                </p>
              </div>
              {sub.status === 'trial' && (
                <div
                  className="text-[11px] font-semibold px-2.5 py-1 rounded-lg shrink-0"
                  style={{
                    background: 'color-mix(in srgb, var(--status-trial, #d97706) 12%, transparent)',
                    color: 'var(--status-trial, #d97706)',
                  }}
                >
                  {trialDaysLeft(sub.trialEndsAt)}d left in trial
                </div>
              )}
            </div>
          </div>

          {billing && billing.lineItems.length > 0 && (
            <div
              className="rounded-xl overflow-hidden"
              style={{ border: '1px solid var(--hub-border)' }}
            >
              <div
                className="px-4 py-2.5 text-[11px] font-semibold uppercase tracking-wider"
                style={{
                  background: 'var(--hub-hover)',
                  color: 'var(--hub-text-muted)',
                  borderBottom: '1px solid var(--hub-border)',
                }}
              >
                This month
              </div>
              {billing.lineItems.map((item, i) => (
                <div
                  key={i}
                  className="px-4 py-3 flex items-center justify-between gap-4"
                  style={{
                    borderBottom:
                      i < billing.lineItems.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                >
                  <div>
                    <p className="text-[12px]" style={{ color: 'var(--hub-text)' }}>
                      {item.seatCount} seat{item.seatCount !== 1 ? 's' : ''} ·{' '}
                      {centsToDisplay(item.pricePerSeatCents)}/seat/mo
                    </p>
                    <p
                      className="text-[11px] font-mono mt-0.5"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      {formatPeriodRange(item.periodStart, item.periodEnd)}
                      {!item.periodEnd && (
                        <span
                          className="ml-1.5 font-sans text-[10px] px-1.5 py-0.5 rounded-full"
                          style={{
                            background:
                              'color-mix(in srgb, var(--status-trial, #d97706) 12%, transparent)',
                            color: 'var(--status-trial, #d97706)',
                          }}
                        >
                          in progress
                        </span>
                      )}
                    </p>
                  </div>
                  <p className="text-[13px] font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(item.amountCents)}
                  </p>
                </div>
              ))}
              <div
                className="px-4 py-3 flex items-center justify-between"
                style={{ background: 'var(--hub-hover)', borderTop: '1px solid var(--hub-border)' }}
              >
                <p className="text-[12px] font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {sub.status === 'trial'
                    ? 'Estimated total (trial, not charged)'
                    : 'Estimated total'}
                </p>
                <p className="text-[15px] font-bold" style={{ color: 'var(--hub-text)' }}>
                  {centsToDisplay(billing.totalCents)}
                </p>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
};

// ── Main Settings Component ───────────────────────────────────────────────────

const Settings: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabId>('general');

  return (
    <div className="space-y-0">
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
                color: isActive ? 'var(--erp-accent, #f59e0b)' : 'var(--hub-text-muted)',
                background: 'transparent',
                border: 'none',
                borderBottom: isActive
                  ? '2px solid var(--erp-accent, #f59e0b)'
                  : '2px solid transparent',
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
        {activeTab === 'billing' && <BillingTab />}
      </motion.div>
    </div>
  );
};

export default Settings;
