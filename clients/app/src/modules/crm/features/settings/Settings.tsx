import React, { useState, useEffect, useRef } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Save } from 'lucide-react';
import { useCrmSettings, useUpdateCrmSettings } from '../../shared/hooks/queries/useCrmSettings';
import { Card, Input, Select, Button, SkeletonList, ACCENT } from '../../shared/components/ui';
import { useUserStore } from '../../../../shared/stores/userStore';
import type { Settings } from '../../shared/schemas/settings';

const GeneralTab: React.FC = () => {
  const { t } = useTranslation();
  const { data: fetched, isLoading, isError, refetch } = useCrmSettings();
  const update = useUpdateCrmSettings();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const canEdit = !!(access?.isOwner || access?.isAdmin);
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
    if (!canEdit || !fetched || update.isPending || saveInFlight.current) return;
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
          <p style={{ margin: 0, color: 'var(--hub-text-muted)' }}>{t('crmErrors.generic')}</p>
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
      {canEdit && (
        <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 16 }}>
          <Button variant="primary" disabled={!canEdit || update.isPending} onClick={handleSave}>
            <Save size={13} />
            {t('crm.settings.saveChanges')}
          </Button>
        </div>
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
          {t('crm.settings.business')}
        </h4>
        <div style={{ display: 'grid', gap: '12px' }}>
          <Select
            label={t('crm.settings.timezone')}
            value={form.timezone}
            disabled={!canEdit || update.isPending}
            onChange={(e) => set('timezone', e.target.value)}
          >
            <option value="America/Los_Angeles">{t('crm.settings.tzLosAngeles')}</option>
            <option value="America/New_York">{t('crm.settings.tzNewYork')}</option>
            <option value="Europe/London">{t('crm.settings.tzLondon')}</option>
            <option value="Europe/Paris">{t('crm.settings.tzParis')}</option>
            <option value="Asia/Tokyo">{t('crm.settings.tzTokyo')}</option>
          </Select>
          <p style={{ margin: 0, fontSize: '11px', color: 'var(--hub-text-muted)' }}>
            {t('crm.settings.businessInPortal')}
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
            disabled={!canEdit || update.isPending}
            onChange={(e) => set('fromEmail', e.target.value)}
          />
          <Input
            label={t('crm.settings.fromName')}
            value={form.fromName}
            disabled={!canEdit || update.isPending}
            onChange={(e) => set('fromName', e.target.value)}
          />
          <Input
            label={t('crm.settings.smsSenderId')}
            value={form.smsSenderId}
            disabled={!canEdit || update.isPending}
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
            disabled={!canEdit || update.isPending}
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

export default function SettingsScreen() {
  const { t } = useTranslation();
  return (
    <section className="portal-crm-settings">
      <HelmetContainer title={`${t('crm.settings.pageTitle')} · rtylr`} />
      <GeneralTab />
    </section>
  );
}
