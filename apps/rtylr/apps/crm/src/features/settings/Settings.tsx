import React, { useState, useEffect, useRef } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useUserStore } from '@voxire/auth';
import { Save, Users, Settings2, UserPlus, UserMinus } from 'lucide-react';
import { motion } from 'motion/react';
import { EmptyState } from '@voxire/ui';
import { useCrmSettings, useUpdateCrmSettings } from '../../shared/hooks/queries/useCrmSettings';
import {
  PageHeader,
  Card,
  Input,
  Select,
  Button,
  SkeletonList,
  ACCENT,
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
  type CrmSubscription,
} from '../../shared/hooks/queries/useCrmSubscription';
import { canCrmPermission } from '../../shared/lib/crmPermissions';

type TabId = 'general' | 'team';

interface TabDef {
  id: TabId;
  label: string;
  icon: React.ComponentType<{ size?: number }>;
}

// Ids only: module scope runs before i18next has a language. The labels come
// from crm.settings.tab<Id> where the tab bar renders.
const TABS: Omit<TabDef, 'label'>[] = [
  { id: 'general', icon: Settings2 },
  { id: 'team', icon: Users },
];

// ── General Tab ───────────────────────────────────────────────────────────────

const GeneralTab: React.FC = () => {
  const { t } = useTranslation();
  const { data: fetched, isLoading, isError, refetch } = useCrmSettings();
  const update = useUpdateCrmSettings();
  const defaultSettings: Settings = {
    defaultConsent: true,
    timezone: 'America/Los_Angeles',
    fromEmail: '',
    fromName: '',
    smsSenderId: '',
  };
  const [form, setForm] = useState<Settings>(defaultSettings);
  const isDirty = useRef(false);
  const saveInFlight = useRef(false);

  useEffect(() => {
    if (fetched && !isDirty.current) setForm(fetched);
  }, [fetched]);

  const set = <K extends keyof typeof form>(k: K, v: (typeof form)[K]): void => {
    isDirty.current = true;
    setForm((f) => ({ ...f, [k]: v }));
  };

  const handleSave = (): void => {
    if (!fetched || update.isPending || saveInFlight.current) return;
    saveInFlight.current = true;
    update.mutate(form, {
      onSuccess: () => {
        isDirty.current = false;
        toast.success(t('crm.settings.saved'));
      },
      onError: () => toast.error(t('crm.settings.saveFailed')),
      onSettled: () => {
        saveInFlight.current = false;
      },
    });
  };

  if (isError) {
    return (
      <Card>
        <div role="alert" style={{ display: 'grid', justifyItems: 'start', gap: 12 }}>
          <p style={{ margin: 0, color: 'var(--hub-text-muted)' }}>{t('errors.generic')}</p>
          <Button variant="outline" onClick={() => void refetch()}>
            {t('crm.actions.retry')}
          </Button>
        </div>
      </Card>
    );
  }

  if (isLoading || !fetched) {
    return (
      <div aria-busy="true">
        <Card>
          <SkeletonList count={3} />
        </Card>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: '680px' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
        <Button variant="primary" disabled={update.isPending} onClick={handleSave}>
          <Save size={13} />
          {t('crm.settings.saveChanges')}
        </Button>
      </div>

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
          {t('crm.settings.business')}
        </h4>
        <div style={{ display: 'grid', gap: '12px' }}>
          <Select
            label={t('crm.settings.timezone')}
            value={form.timezone}
            disabled={update.isPending}
            onChange={(e) => set('timezone', e.target.value)}
          >
            <option value="America/Los_Angeles">{t('crm.settings.tzLosAngeles')}</option>
            <option value="America/New_York">{t('crm.settings.tzNewYork')}</option>
            <option value="Europe/London">{t('crm.settings.tzLondon')}</option>
            <option value="Europe/Paris">{t('crm.settings.tzParis')}</option>
            <option value="Asia/Tokyo">{t('crm.settings.tzTokyo')}</option>
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
          {t('crm.settings.messaging')}
        </h4>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
          <Input
            label={t('crm.settings.fromEmail')}
            type="email"
            value={form.fromEmail}
            disabled={update.isPending}
            onChange={(e) => set('fromEmail', e.target.value)}
          />
          <Input
            label={t('crm.settings.fromName')}
            value={form.fromName}
            disabled={update.isPending}
            onChange={(e) => set('fromName', e.target.value)}
          />
          <Input
            label={t('crm.settings.smsSenderId')}
            value={form.smsSenderId}
            disabled={update.isPending}
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
          {t('crm.settings.consentDefaults')}
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
            disabled={update.isPending}
            onChange={(e) => set('defaultConsent', e.target.checked)}
            style={{ accentColor: ACCENT, width: '16px', height: '16px' }}
          />
          <div>
            <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('crm.settings.optInByDefault')}
            </div>
            <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>
              {t('crm.settings.optInByDefaultHint')}
            </div>
          </div>
        </label>
      </Card>
    </div>
  );
};

function getInviteSeatQuote(subscription: CrmSubscription | undefined) {
  const pricePerSeatCents = subscription?.pricePerSeatCents;
  const seatCount = subscription?.seatCount;
  const includedSeatLimit = subscription?.includedSeatLimit;
  if (
    typeof pricePerSeatCents !== 'number' ||
    pricePerSeatCents <= 0 ||
    typeof seatCount !== 'number' ||
    seatCount < 0 ||
    typeof includedSeatLimit !== 'number' ||
    includedSeatLimit < 0
  ) {
    return null;
  }

  const incrementalBillableSeats = subscription?.isGranted
    ? 0
    : Math.max(0, Math.min(1, seatCount + 1 - includedSeatLimit));
  const incrementalPriceCents = pricePerSeatCents * incrementalBillableSeats;
  return {
    ...computeSeatProration(incrementalPriceCents),
    monthlyCents: incrementalPriceCents,
  };
}

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { t } = useTranslation();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canInvite = canCrmPermission(role, 'hub:team:invite');
  const canRemove = canCrmPermission(role, 'hub:team:remove');
  const { data: fetchedMembers = [], isLoading, isError, refetch } = useCrmTeam();
  const members = fetchedMembers.filter((member) => member.isActive !== false);
  const invite = useCrmInviteMember();
  const deactivate = useCrmDeactivateMember();
  const { data: sub } = useCrmSubscription();
  const seatQuote = getInviteSeatQuote(sub);

  const [email, setEmail] = useState('');
  const [fullName, setFullName] = useState('');
  const [showInvite, setShowInvite] = useState(false);

  const handleInvite = (): void => {
    if (!email.trim() || !fullName.trim()) {
      toast.error(t('crm.settings.emailAndNameRequired'));
      return;
    }
    invite.mutate(
      { email: email.trim(), fullName: fullName.trim(), role: 'staff' },
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
          {t('crm.settings.teamTitle')}
        </h2>
        <p style={{ fontSize: 12, color: 'var(--hub-text-muted)', margin: 0 }}>
          {t('crm.settings.teamSubtitle')}
        </p>
      </div>

      {canInvite && showInvite ? (
        <Card>
          <h3
            style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 12px' }}
          >
            {t('crm.settings.inviteSomeone')}
          </h3>
          <div
            style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 12 }}
          >
            <div>
              <label
                style={{
                  display: 'block',
                  fontSize: 11,
                  fontWeight: 500,
                  marginBottom: 4,
                  color: 'var(--hub-text-muted)',
                }}
              >
                {t('crm.createCustomer.fullName')}
              </label>
              <input
                type="text"
                placeholder={t('crm.settings.invitePlaceholder')}
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                style={inputStyle}
              />
            </div>
            <div>
              <label
                style={{
                  display: 'block',
                  fontSize: 11,
                  fontWeight: 500,
                  marginBottom: 4,
                  color: 'var(--hub-text-muted)',
                }}
              >
                {t('common.email')}
              </label>
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
            </div>
          </div>
          {seatQuote ? (
            <p style={{ fontSize: 11, color: 'var(--hub-text-muted)' }}>
              Adding this member costs{' '}
              <strong style={{ color: 'var(--hub-text)' }}>
                {centsToDisplay(seatQuote.proratedCents)}
              </strong>{' '}
              this month ({seatQuote.daysRemaining}/{seatQuote.daysInMonth} days), then{' '}
              <strong style={{ color: 'var(--hub-text)' }}>
                {centsToDisplay(seatQuote.monthlyCents)}/mo
              </strong>{' '}
              from {seatQuote.nextMonthName} 1st.
            </p>
          ) : null}
          <div style={{ display: 'flex', gap: 8 }}>
            <Button
              variant="primary"
              disabled={!email.trim() || !fullName.trim() || invite.isPending}
              onClick={handleInvite}
            >
              <UserPlus size={13} />
              {invite.isPending ? 'Sending…' : 'Send invite'}
            </Button>
            <Button variant="ghost" onClick={() => setShowInvite(false)}>
              {t('common.cancel')}
            </Button>
          </div>
        </Card>
      ) : canInvite ? (
        <div>
          <Button variant="primary" onClick={() => setShowInvite(true)}>
            <UserPlus size={13} />
            {t('crm.team.inviteMember')}
          </Button>
        </div>
      ) : null}

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
        <Card>
          <div role="alert" style={{ display: 'grid', justifyItems: 'start', gap: 12 }}>
            <p style={{ margin: 0, color: 'var(--hub-text-muted)' }}>{t('errors.generic')}</p>
            <Button variant="outline" onClick={() => void refetch()}>
              {t('crm.actions.retry')}
            </Button>
          </div>
        </Card>
      ) : members.length === 0 ? (
        <EmptyState
          icon={Users}
          title={t('crm.team.emptyTitle')}
          action={
            canInvite
              ? { label: t('crm.team.inviteMember'), onClick: () => setShowInvite(true) }
              : { label: t('crm.actions.refresh'), onClick: () => refetch() }
          }
        />
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
              {canRemove && (
                <button
                  type="button"
                  aria-label={`Remove ${m.fullName || m.email}`}
                  onClick={() => {
                    const name = m.fullName || m.email || 'this member';
                    // Destructive + irreversible: require confirmation.
                    if (window.confirm(`Remove ${name} from the CRM team?`)) {
                      deactivate.mutate(m.id ?? '');
                    }
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
              )}
            </div>
          ))}
        </Card>
      )}
    </div>
  );
};

// ── Main Settings Component ───────────────────────────────────────────────────

const SettingsScreen: React.FC = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<TabId>('general');

  return (
    <>
      <HelmetContainer title={`${t('crm.settings.pageTitle')} · rtylr`} />
      <div>
        <PageHeader title={t('crm.settings.pageTitle')} subtitle={t('crm.settings.subtitle')} />

        {/* Tab bar */}
        <div
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
                {t(`crm.settings.tab${tab.id === 'general' ? 'General' : 'Team'}`)}
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
    </>
  );
};

export default SettingsScreen;
