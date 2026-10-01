import React, { useState, useEffect, useRef } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { APP_URLS } from '@voxire/config';
import { Save, Users, Settings2 } from 'lucide-react';
import { motion } from 'motion/react';
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

// ── Team Tab ──────────────────────────────────────────────────────────────────

const TeamTab: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div style={{ maxWidth: '680px', display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div>
        <h2 style={{ fontSize: 15, fontWeight: 600, color: 'var(--hub-text)', margin: '0 0 4px' }}>
          {t('crm.settings.teamManagementTitle')}
        </h2>
        <p style={{ fontSize: 12, color: 'var(--hub-text-muted)', margin: 0 }}>
          {t('crm.settings.teamManagementDescription')}
        </p>
      </div>
      <a
        href={`${APP_URLS.dash}/team`}
        style={{
          display: 'inline-flex',
          alignSelf: 'flex-start',
          padding: '8px 12px',
          borderRadius: 8,
          background: 'var(--crm-accent, #3b82f6)',
          color: '#fff',
          fontSize: 12,
          fontWeight: 600,
          textDecoration: 'none',
        }}
      >
        {t('crm.settings.teamManagementAction')}
      </a>
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
