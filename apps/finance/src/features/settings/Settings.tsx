import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import {
  RotateCcw,
  Save,
  Image as ImageIcon,
  Settings2,
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
import { useSettings } from '../../shared/lib/settings';
import { CURRENCY_OPTIONS, type Currency } from '../../shared/lib/money';
import {
  cardStyle,
  inputStyle,
  labelStyle,
  primaryBtn,
  ghostBtn,
  ACCENT,
} from '../../shared/lib/tokens';
import {
  useFinanceTeam,
  useFinanceInviteMember,
  useFinanceDeactivateMember,
} from '../../shared/hooks/queries/useFinanceTeam';
import {
  useFinanceSubscription,
  useFinanceSubscriptionBilling,
  centsToDisplay,
  trialDaysLeft,
  formatPeriodRange,
  computeSeatProration,
} from '../../shared/hooks/queries/useFinanceSubscription';

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

type TabId = 'general' | 'team' | 'billing';

interface Tab {
  id: TabId;
  label: string;
  icon: React.ComponentType<{ size?: number }>;
}

const TABS: Tab[] = [
  { id: 'general', label: 'General', icon: Settings2 },
  { id: 'team', label: 'Team', icon: Users },
  { id: 'billing', label: 'Billing', icon: CreditCard },
];

// ── General Tab ───────────────────────────────────────────────────────────────

const GeneralTab: React.FC = () => {
  const settings = useSettings();

  const handleReset = (): void => {
    settings.reset();
    toast.success('Settings reset');
  };

  const handleSave = (): void => {
    toast.success('Settings saved');
  };

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginBottom: 20 }}>
        <button style={ghostBtn} onClick={handleReset}>
          <RotateCcw size={13} /> Reset
        </button>
        <button style={primaryBtn} onClick={handleSave}>
          <Save size={13} /> Save
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <Card title="Company">
          <Field label="Company name">
            <input
              value={settings.companyName}
              onChange={(e) => settings.setSetting('companyName', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label="Billing email">
            <input
              value={settings.companyEmail}
              onChange={(e) => settings.setSetting('companyEmail', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label="Brand color">
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <input
                type="color"
                value={settings.brandColor}
                onChange={(e) => settings.setSetting('brandColor', e.target.value)}
                style={{
                  width: 44,
                  height: 32,
                  border: 'none',
                  borderRadius: 8,
                  background: 'transparent',
                }}
              />
              <input
                value={settings.brandColor}
                onChange={(e) => settings.setSetting('brandColor', e.target.value)}
                style={inputStyle}
              />
            </div>
          </Field>
          <Field label="Logo URL">
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                value={settings.logoUrl}
                onChange={(e) => settings.setSetting('logoUrl', e.target.value)}
                placeholder="https://..."
                style={inputStyle}
              />
              <button style={{ ...ghostBtn, padding: '8px 12px' }}>
                <ImageIcon size={13} />
              </button>
            </div>
          </Field>
        </Card>

        <Card title="Invoicing">
          <Field label="Invoice prefix">
            <input
              value={settings.invoicePrefix}
              onChange={(e) => settings.setSetting('invoicePrefix', e.target.value)}
              style={inputStyle}
            />
          </Field>
          <Field label="Default tax rate (%)">
            <input
              type="number"
              step="0.1"
              value={settings.defaultTaxRate}
              onChange={(e) =>
                settings.setSetting('defaultTaxRate', parseFloat(e.target.value) || 0)
              }
              style={inputStyle}
            />
          </Field>
          <Field label="Late fee (%)">
            <input
              type="number"
              step="0.1"
              value={settings.lateFeePercent}
              onChange={(e) =>
                settings.setSetting('lateFeePercent', parseFloat(e.target.value) || 0)
              }
              style={inputStyle}
            />
          </Field>
          <Field label="Grace period (days)">
            <input
              type="number"
              value={settings.lateFeeGraceDays}
              onChange={(e) =>
                settings.setSetting('lateFeeGraceDays', parseInt(e.target.value) || 0)
              }
              style={inputStyle}
            />
          </Field>
          <Field label="Default notes">
            <textarea
              value={settings.invoiceNotes}
              onChange={(e) => settings.setSetting('invoiceNotes', e.target.value)}
              rows={3}
              style={{ ...inputStyle, resize: 'vertical', fontFamily: 'inherit' }}
            />
          </Field>
        </Card>

        <Card title="Accounting">
          <Field label="Currency">
            <select
              value={settings.currency}
              onChange={(e) => settings.setSetting('currency', e.target.value as Currency)}
              style={inputStyle}
            >
              {CURRENCY_OPTIONS.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Fiscal year start">
            <select
              value={settings.fiscalYearStartMonth}
              onChange={(e) =>
                settings.setSetting('fiscalYearStartMonth', parseInt(e.target.value) || 1)
              }
              style={inputStyle}
            >
              {MONTHS.map((m, i) => (
                <option key={m} value={i + 1}>
                  {m}
                </option>
              ))}
            </select>
          </Field>
        </Card>

        <Card title="About">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {[
              { label: 'Application', value: appConfig.name },
              { label: 'Version', value: `v${appConfig.version}` },
            ].map(({ label, value }) => (
              <div
                key={label}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '4px 0',
                  borderBottom: '1px solid var(--hub-border)',
                }}
              >
                <span style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{label}</span>
                <span style={{ fontSize: 12, fontFamily: 'monospace', color: 'var(--hub-text)' }}>
                  {value}
                </span>
              </div>
            ))}
          </div>
        </Card>

        <Card title="Shortcuts">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <ShortcutRow label="Global search" keys={['cmd', 'k']} />
            <ShortcutRow label="New invoice" keys={['n']} />
            <ShortcutRow label="New expense" keys={['e']} />
            <ShortcutRow label="Focus search" keys={['/']} />
          </div>
        </Card>
      </div>
    </>
  );
};

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { data: members = [], isLoading } = useFinanceTeam();
  const invite = useFinanceInviteMember();
  const deactivate = useFinanceDeactivateMember();
  const { data: sub } = useFinanceSubscription();

  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [showInvite, setShowInvite] = useState(false);

  const handleInvite = (): void => {
    if (!email.trim() || !fullName.trim()) return;
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

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 600 }}>
      <div>
        <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--hub-text)', marginBottom: 4 }}>
          Team
        </div>
        <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
          People who have access to Finance.
        </div>
      </div>

      {showInvite ? (
        <div
          style={{ ...cardStyle, padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
            Invite someone to Finance
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Field label="Full name">
              <input
                type="text"
                placeholder="Jane Smith"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                style={inputStyle}
              />
            </Field>
            <Field label="Email">
              <input
                type="email"
                placeholder="jane@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleInvite();
                }}
                style={inputStyle}
              />
            </Field>
          </div>
          {sub &&
            sub.pricePerSeatCents > 0 &&
            (() => {
              const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                computeSeatProration(sub.pricePerSeatCents);
              return (
                <p style={{ fontSize: 11, color: 'var(--hub-text-muted)' }}>
                  Adding this member costs{' '}
                  <strong style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(proratedCents)}
                  </strong>{' '}
                  this month ({daysRemaining}/{daysInMonth} days), then{' '}
                  <strong style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(sub.pricePerSeatCents)}/mo
                  </strong>{' '}
                  from {nextMonthName} 1st.
                </p>
              );
            })()}
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              type="button"
              disabled={!email.trim() || !fullName.trim() || invite.isPending}
              onClick={handleInvite}
              style={{
                ...primaryBtn,
                opacity: !email.trim() || !fullName.trim() || invite.isPending ? 0.4 : 1,
              }}
            >
              <UserPlus size={13} />
              {invite.isPending ? 'Sending…' : 'Send invite'}
            </button>
            <button type="button" onClick={() => setShowInvite(false)} style={ghostBtn}>
              Cancel
            </button>
          </div>
        </div>
      ) : (
        <div>
          <button type="button" onClick={() => setShowInvite(true)} style={primaryBtn}>
            <UserPlus size={13} />
            Invite member
          </button>
        </div>
      )}

      {isLoading ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              style={{
                height: 56,
                borderRadius: 12,
                background: 'var(--hub-hover)',
                animation: 'pulse 1.5s ease-in-out infinite',
              }}
            />
          ))}
        </div>
      ) : members.length === 0 ? (
        <div style={{ padding: '32px 0', textAlign: 'center', color: 'var(--hub-text-muted)' }}>
          <Users size={24} style={{ margin: '0 auto 8px', opacity: 0.4, display: 'block' }} />
          <p style={{ fontSize: 13 }}>No team members yet.</p>
        </div>
      ) : (
        <div style={{ ...cardStyle, overflow: 'hidden' }}>
          {members.map((m, idx) => (
            <div
              key={m.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                padding: '12px 16px',
                borderBottom: idx < members.length - 1 ? '1px solid var(--hub-border)' : 'none',
              }}
            >
              <div
                style={{
                  height: 32,
                  width: 32,
                  borderRadius: '50%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 12,
                  fontWeight: 600,
                  flexShrink: 0,
                  background: `color-mix(in srgb, ${ACCENT} 20%, transparent)`,
                  color: ACCENT,
                }}
              >
                {m.name.charAt(0).toUpperCase()}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <p
                  style={{
                    fontSize: 13,
                    fontWeight: 500,
                    color: 'var(--hub-text)',
                    margin: 0,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {m.name}
                </p>
                <p
                  style={{
                    fontSize: 11,
                    color: 'var(--hub-text-muted)',
                    margin: 0,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                >
                  {m.email}
                </p>
              </div>
              <span
                style={{
                  fontSize: 10,
                  fontWeight: 600,
                  padding: '2px 8px',
                  borderRadius: 999,
                  background: 'var(--hub-hover)',
                  color: 'var(--hub-text-muted)',
                  flexShrink: 0,
                }}
              >
                {m.role}
              </span>
              <button
                type="button"
                aria-label={`Remove ${m.name}`}
                onClick={() => {
                  deactivate.mutate(m.id);
                }}
                style={{
                  height: 28,
                  width: 28,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  borderRadius: 6,
                  border: 'none',
                  background: 'transparent',
                  color: 'var(--hub-text-muted)',
                  cursor: 'pointer',
                  flexShrink: 0,
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
  const { data: sub, isLoading: subLoading } = useFinanceSubscription();
  const { data: billing, isLoading: billingLoading } = useFinanceSubscriptionBilling();

  const isLoading = subLoading || billingLoading;

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 600 }}>
        {[32, 112, 80].map((h, i) => (
          <div
            key={i}
            style={{
              height: h,
              borderRadius: 12,
              background: 'var(--hub-hover)',
            }}
          />
        ))}
      </div>
    );
  }

  const StatusIcon = sub ? (STATUS_ICON[sub.status] ?? AlertCircle) : null;
  const statusColor = sub ? (STATUS_COLOR[sub.status] ?? STATUS_COLOR['expired']) : '';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 600 }}>
      <div>
        <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--hub-text)', marginBottom: 4 }}>
          Billing
        </div>
        <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
          Your Finance subscription and usage this month.
        </div>
      </div>

      {!sub ? (
        <div
          style={{
            ...cardStyle,
            padding: '32px 24px',
            textAlign: 'center',
            display: 'flex',
            flexDirection: 'column',
            gap: 12,
            alignItems: 'center',
          }}
        >
          <CreditCard size={24} style={{ opacity: 0.3, color: 'var(--hub-text)' }} />
          <p style={{ fontSize: 13, fontWeight: 500, color: 'var(--hub-text)', margin: 0 }}>
            No active subscription
          </p>
          <p style={{ fontSize: 12, color: 'var(--hub-text-muted)', margin: 0 }}>
            Contact your account owner or visit billing to activate Finance.
          </p>
        </div>
      ) : (
        <>
          <div
            style={{ ...cardStyle, padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <p style={{ fontSize: 14, fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}>
                {sub.planName}
              </p>
              {StatusIcon && (
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                    fontSize: 11,
                    fontWeight: 600,
                    color: statusColor,
                  }}
                >
                  <StatusIcon size={12} />
                  <span style={{ textTransform: 'capitalize' }}>{sub.status}</span>
                </div>
              )}
            </div>

            <div
              style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between' }}
            >
              <div>
                <p style={{ fontSize: 22, fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
                  {centsToDisplay(billing?.totalCents ?? sub.seatCount * sub.pricePerSeatCents)}
                </p>
                <p style={{ fontSize: 11, color: 'var(--hub-text-muted)', margin: '4px 0 0' }}>
                  {sub.seatCount} seat{sub.seatCount !== 1 ? 's' : ''} ×{' '}
                  {centsToDisplay(sub.pricePerSeatCents)}/seat · billed 1st of each month
                </p>
              </div>
              {sub.status === 'trial' && (
                <div
                  style={{
                    fontSize: 11,
                    fontWeight: 600,
                    padding: '4px 10px',
                    borderRadius: 8,
                    flexShrink: 0,
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
            <div style={{ ...cardStyle, overflow: 'hidden' }}>
              <div
                style={{
                  padding: '10px 16px',
                  fontSize: 11,
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.05em',
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
                  style={{
                    padding: '12px 16px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    gap: 16,
                    borderBottom:
                      i < billing.lineItems.length - 1 ? '1px solid var(--hub-border)' : 'none',
                  }}
                >
                  <div>
                    <p style={{ fontSize: 12, color: 'var(--hub-text)', margin: 0 }}>
                      {item.seatCount} seat{item.seatCount !== 1 ? 's' : ''} ·{' '}
                      {centsToDisplay(item.pricePerSeatCents)}/seat/mo
                    </p>
                    <p
                      style={{
                        fontSize: 11,
                        fontFamily: 'monospace',
                        color: 'var(--hub-text-muted)',
                        margin: '2px 0 0',
                      }}
                    >
                      {formatPeriodRange(item.periodStart, item.periodEnd)}
                      {!item.periodEnd && (
                        <span
                          style={{
                            marginLeft: 6,
                            fontFamily: 'inherit',
                            fontSize: 10,
                            padding: '2px 6px',
                            borderRadius: 999,
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
                  <p style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}>
                    {centsToDisplay(item.amountCents)}
                  </p>
                </div>
              ))}
              <div
                style={{
                  padding: '12px 16px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  background: 'var(--hub-hover)',
                  borderTop: '1px solid var(--hub-border)',
                }}
              >
                <p style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text)', margin: 0 }}>
                  {sub.status === 'trial'
                    ? 'Estimated total (trial, not charged)'
                    : 'Estimated total'}
                </p>
                <p style={{ fontSize: 15, fontWeight: 700, color: 'var(--hub-text)', margin: 0 }}>
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
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <PageHeader title="Settings" />

      {/* Tab bar */}
      <div
        style={{
          display: 'flex',
          gap: 0,
          borderBottom: '1px solid var(--hub-border)',
        }}
      >
        {TABS.map((tab) => {
          const Icon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '10px 16px',
                fontSize: 13,
                fontWeight: isActive ? 600 : 500,
                color: isActive ? ACCENT : 'var(--hub-text-muted)',
                background: 'transparent',
                border: 'none',
                borderBottom: isActive ? `2px solid ${ACCENT}` : '2px solid transparent',
                marginBottom: -1,
                cursor: 'pointer',
                transition: 'color 0.15s, border-color 0.15s',
              }}
            >
              <Icon size={14} />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Tab content */}
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
    </motion.div>
  );
};

// ── Shared sub-components ─────────────────────────────────────────────────────

const Card: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <div style={{ ...cardStyle, padding: 20 }}>
    <div style={{ fontSize: 14, fontWeight: 700, color: 'var(--hub-text)', marginBottom: 14 }}>
      {title}
    </div>
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>{children}</div>
  </div>
);

const Field: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
    <label style={labelStyle}>{label}</label>
    {children}
  </div>
);

const ShortcutRow: React.FC<{ label: string; keys: string[] }> = ({ label, keys }) => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '6px 0',
      borderBottom: '1px solid var(--hub-border)',
    }}
  >
    <span style={{ fontSize: 13, color: 'var(--hub-text-muted)' }}>{label}</span>
    <div style={{ display: 'flex', gap: 4 }}>
      {keys.map((k) => (
        <kbd
          key={k}
          style={{
            padding: '3px 8px',
            borderRadius: 6,
            background: 'var(--hub-hover)',
            border: '1px solid var(--hub-border)',
            fontSize: 11,
            color: 'var(--hub-text)',
            textTransform: 'uppercase',
          }}
        >
          {k}
        </kbd>
      ))}
    </div>
  </div>
);

export default Settings;
