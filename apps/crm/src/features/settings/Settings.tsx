import React, { useState, useEffect } from 'react';
import { HelmetContainer } from '@voxire/components';
import { TIMEZONES } from '@voxire/config';
import { toast } from 'sonner';
import { Save, Users, Settings2, UserPlus, UserMinus } from 'lucide-react';
import { motion } from 'motion/react';
import { useCrmSettings, useUpdateCrmSettings } from '../../shared/hooks/queries/useCrmSettings';
import {
  PageHeader,
  Card,
  Input,
  Select,
  Button,
  ACCENT,
  ConfirmModal,
} from '../../shared/components/ui';
import type { Settings } from '../../shared/schemas/settings';
import {
  useCrmTeam,
  useCrmInviteMember,
  useCrmDeactivateMember,
} from '../../shared/hooks/queries/useCrmTeam';
import {
  useCrmSubscription,
  centsToDisplay,
  computeSeatProration,
} from '../../shared/hooks/queries/useCrmSubscription';

type TabId = 'general' | 'team';

interface TabDef {
  id: TabId;
  label: string;
  icon: React.ComponentType<{ size?: number }>;
}

const TABS: TabDef[] = [
  { id: 'general', label: 'General', icon: Settings2 },
  { id: 'team', label: 'Team', icon: Users },
];

// ── General Tab ───────────────────────────────────────────────────────────────

const GeneralTab: React.FC = () => {
  const { data: fetched, isLoading, isError } = useCrmSettings();
  const update = useUpdateCrmSettings();
  const defaultSettings: Settings = {
    defaultConsent: false,
    timezone: 'America/Los_Angeles',
    fromEmail: '',
    fromName: '',
    smsSenderId: '',
  };
  const [form, setForm] = useState<Settings>(defaultSettings);

  useEffect(() => {
    if (fetched) setForm(fetched);
  }, [fetched]);

  const set = <K extends keyof typeof form>(k: K, v: (typeof form)[K]): void =>
    setForm((f) => ({ ...f, [k]: v }));

  const settingsUnavailable = isLoading || isError;
  const timezoneOptions =
    form.timezone && !TIMEZONES.some((timezone) => timezone.value === form.timezone)
      ? [{ value: form.timezone, label: `Current (${form.timezone})` }, ...TIMEZONES]
      : TIMEZONES;

  const handleSave = (): void => {
    if (update.isPending || settingsUnavailable) return;
    update.mutate(form);
  };

  return (
    <div style={{ maxWidth: '680px' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
        <Button
          variant="primary"
          onClick={handleSave}
          disabled={update.isPending || settingsUnavailable}
        >
          <Save size={13} />
          {update.isPending ? 'Saving…' : 'Save changes'}
        </Button>
      </div>

      {isError && (
        <p style={{ color: '#ef4444', fontSize: '12px', margin: '0 0 16px' }}>
          Unable to load your current settings. Refresh before saving changes.
        </p>
      )}

      <Card style={{ marginBottom: '16px' }}>
        <h4
          style={{
            fontSize: '12px',
            fontWeight: 600,
            color: 'var(--hub-text)',
            margin: '0 0 14px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
          }}
        >
          Business
        </h4>
        <div style={{ display: 'grid', gap: '12px' }}>
          <Select
            label="Timezone"
            value={form.timezone}
            onChange={(e) => set('timezone', e.target.value)}
          >
            {!form.timezone && <option value="">Select a timezone...</option>}
            {timezoneOptions.map((timezone) => (
              <option key={timezone.value} value={timezone.value}>
                {timezone.label}
              </option>
            ))}
          </Select>
          <p style={{ margin: 0, fontSize: '11px', color: 'var(--hub-text-muted)' }}>
            Business profile, currency, and at-risk policy are not available from the CRM settings
            service yet.
          </p>
        </div>
      </Card>

      <Card style={{ marginBottom: '16px' }}>
        <h4
          style={{
            fontSize: '12px',
            fontWeight: 600,
            color: 'var(--hub-text)',
            margin: '0 0 14px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
          }}
        >
          Messaging
        </h4>
        <div className="crm-form-grid" style={{ display: 'grid', gap: '12px' }}>
          <Input
            label="From email"
            type="email"
            value={form.fromEmail}
            onChange={(e) => set('fromEmail', e.target.value)}
          />
          <Input
            label="From name"
            value={form.fromName}
            onChange={(e) => set('fromName', e.target.value)}
          />
          <Input
            label="SMS sender ID"
            value={form.smsSenderId}
            onChange={(e) => set('smsSenderId', e.target.value)}
          />
        </div>
      </Card>

      <Card>
        <h4
          style={{
            fontSize: '12px',
            fontWeight: 600,
            color: 'var(--hub-text)',
            margin: '0 0 14px',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
          }}
        >
          Consent defaults
        </h4>
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '10px',
            cursor: 'pointer',
            padding: '10px 0',
          }}
        >
          <input
            type="checkbox"
            checked={form.defaultConsent}
            onChange={(e) => set('defaultConsent', e.target.checked)}
            style={{ accentColor: ACCENT, width: '16px', height: '16px' }}
          />
          <div>
            <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
              Opt in new customers by default
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              New customers will be opted into marketing unless they say otherwise
            </div>
          </div>
        </label>
      </Card>
    </div>
  );
};

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { data: members = [], isLoading, isError } = useCrmTeam();
  const invite = useCrmInviteMember();
  const deactivate = useCrmDeactivateMember();
  const { data: sub } = useCrmSubscription();

  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [showInvite, setShowInvite] = useState(false);
  const [removeTarget, setRemoveTarget] = useState<{ id: string; label: string } | null>(null);
  const emailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());

  const handleInvite = (): void => {
    if (invite.isPending) return;
    if (!email.trim() || !fullName.trim()) {
      toast.error('Email and full name are required');
      return;
    }
    if (!emailValid) {
      toast.error('Enter a valid email address');
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

  const inputStyle: React.CSSProperties = {
    padding: '8px 12px',
    borderRadius: 8,
    border: '1px solid var(--hub-border)',
    background: 'var(--hub-input-bg)',
    color: 'var(--hub-text)',
    fontSize: 13,
    outline: 'none',
    width: '100%',
    boxSizing: 'border-box',
  };

  return (
    <div style={{ maxWidth: '680px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h2 style={{ fontSize: 15, fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 4px' }}>
          Team
        </h2>
        <p style={{ fontSize: 12, color: 'var(--hub-text-muted)', margin: 0 }}>
          People who have access to CRM.
        </p>
      </div>

      {showInvite ? (
        <Card>
          <h3
            style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 12px' }}
          >
            Invite someone to CRM
          </h3>
          <div className="crm-form-grid" style={{ display: 'grid', gap: 12, marginBottom: 12 }}>
            <div>
              <label
                htmlFor="crm-team-full-name"
                style={{
                  display: 'block',
                  fontSize: 11,
                  fontWeight: 500,
                  marginBottom: 4,
                  color: 'var(--hub-text-muted)',
                }}
              >
                Full name
              </label>
              <input
                id="crm-team-full-name"
                type="text"
                autoComplete="name"
                placeholder="Jane Smith"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleInvite();
                }}
                style={inputStyle}
              />
            </div>
            <div>
              <label
                htmlFor="crm-team-email"
                style={{
                  display: 'block',
                  fontSize: 11,
                  fontWeight: 500,
                  marginBottom: 4,
                  color: 'var(--hub-text-muted)',
                }}
              >
                Email
              </label>
              <input
                id="crm-team-email"
                type="email"
                autoComplete="email"
                placeholder="jane@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleInvite();
                }}
                style={inputStyle}
              />
              {email.trim() && !emailValid && (
                <p role="alert" style={{ fontSize: 11, color: '#ef4444', margin: '4px 0 0' }}>
                  Enter a valid email address.
                </p>
              )}
            </div>
          </div>
          {sub &&
            (sub.pricePerSeatCents ?? 0) > 0 &&
            (() => {
              const { daysRemaining, daysInMonth, proratedCents, nextMonthName } =
                computeSeatProration(sub.pricePerSeatCents ?? 0);
              return (
                <p style={{ fontSize: 11, color: 'var(--hub-text-muted)' }}>
                  Adding this member costs{' '}
                  <strong style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(proratedCents)}
                  </strong>{' '}
                  this month ({daysRemaining}/{daysInMonth} days), then{' '}
                  <strong style={{ color: 'var(--hub-text)' }}>
                    {centsToDisplay(sub.pricePerSeatCents ?? 0)}/mo
                  </strong>{' '}
                  from {nextMonthName} 1st.
                </p>
              );
            })()}
          <div style={{ display: 'flex', gap: 8 }}>
            <Button
              variant="primary"
              disabled={!email.trim() || !emailValid || !fullName.trim() || invite.isPending}
              onClick={handleInvite}
            >
              <UserPlus size={13} />
              {invite.isPending ? 'Sending…' : 'Send invite'}
            </Button>
            <Button variant="ghost" onClick={() => setShowInvite(false)}>
              Cancel
            </Button>
          </div>
        </Card>
      ) : (
        <div>
          <Button variant="primary" onClick={() => setShowInvite(true)}>
            <UserPlus size={13} />
            Invite member
          </Button>
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
              }}
            />
          ))}
        </div>
      ) : isError ? (
        <div style={{ padding: '32px 0', textAlign: 'center', color: '#ef4444' }}>
          <p style={{ fontSize: 13, margin: 0 }}>
            Unable to load team members. Refresh and try again.
          </p>
        </div>
      ) : members.length === 0 ? (
        <div style={{ padding: '32px 0', textAlign: 'center', color: 'var(--hub-text-muted)' }}>
          <Users size={24} style={{ margin: '0 auto 8px', opacity: 0.4, display: 'block' }} />
          <p style={{ fontSize: 13, margin: 0 }}>No team members yet.</p>
        </div>
      ) : (
        <Card padded={false}>
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
                  background: 'color-mix(in srgb, var(--crm-accent, #3b82f6) 20%, transparent)',
                  color: 'var(--crm-accent, #3b82f6)',
                }}
              >
                {(m.fullName || m.email || '').charAt(0).toUpperCase()}
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
                  {m.fullName || m.email}
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
                aria-label={`Remove ${m.fullName || m.email}`}
                onClick={() => {
                  const memberId = m.id;
                  if (memberId) {
                    setRemoveTarget({
                      id: memberId,
                      label: m.fullName || m.email || 'this member',
                    });
                  }
                }}
                disabled={!m.id || deactivate.isPending}
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
                  cursor: !m.id || deactivate.isPending ? 'not-allowed' : 'pointer',
                  flexShrink: 0,
                  opacity: !m.id || deactivate.isPending ? 0.5 : 1,
                }}
              >
                <UserMinus size={13} />
              </button>
            </div>
          ))}
        </Card>
      )}

      <ConfirmModal
        open={removeTarget !== null}
        onClose={() => setRemoveTarget(null)}
        onConfirm={() => {
          if (removeTarget && !deactivate.isPending) {
            deactivate.mutate(removeTarget.id, {
              onSuccess: () => setRemoveTarget(null),
            });
          }
        }}
        closeOnConfirm={false}
        title="Remove team member?"
        description={
          removeTarget
            ? `Remove ${removeTarget.label} from CRM access? They will no longer be able to use this workspace.`
            : undefined
        }
        confirmLabel={deactivate.isPending ? 'Removing…' : 'Remove'}
        confirmDisabled={deactivate.isPending}
        variant="danger"
      />
    </div>
  );
};

// ── Main Settings Component ───────────────────────────────────────────────────

const SettingsScreen: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabId>('general');

  return (
    <>
      <HelmetContainer title="Settings · rtylr" />
      <div>
        <PageHeader title="Settings" subtitle="Configure your CRM defaults" />

        {/* Tab bar */}
        <div
          role="tablist"
          style={{
            display: 'flex',
            borderBottom: '1px solid var(--hub-border)',
            marginBottom: 24,
          }}
        >
          {TABS.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                role="tab"
                aria-selected={isActive}
                onClick={() => setActiveTab(tab.id)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 6,
                  padding: '10px 16px',
                  fontSize: 13,
                  fontWeight: isActive ? 600 : 500,
                  color: isActive ? 'var(--crm-accent, #3b82f6)' : 'var(--hub-text-muted)',
                  background: 'transparent',
                  border: 'none',
                  borderBottom: isActive
                    ? '2px solid var(--crm-accent, #3b82f6)'
                    : '2px solid transparent',
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

        <motion.div
          key={activeTab}
          role="tabpanel"
          aria-label={TABS.find((tab) => tab.id === activeTab)?.label}
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.15 }}
        >
          {activeTab === 'general' && <GeneralTab />}
          {activeTab === 'team' && <TeamTab />}
        </motion.div>
      </div>
    </>
  );
};

export default SettingsScreen;
