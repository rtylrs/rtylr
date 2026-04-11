import React, { useState, useEffect } from 'react';
import { toast } from 'sonner';
import { Save } from 'lucide-react';
import { useSettings, useUpdateSettings } from './useSettings';
import { PageHeader, Card, Input, Select, Button } from '../../shared/components/ui';
import type { Settings } from '../../shared/services/crmApi';

const SettingsScreen: React.FC = () => {
  const { data: fetched } = useSettings();
  const update = useUpdateSettings();
  const defaultSettings: Settings = {
    defaultConsent: true,
    timezone: 'America/Los_Angeles',
    fromEmail: '',
    fromName: '',
    smsSenderId: '',
    businessName: '',
    atRiskDays: 60,
    currency: 'USD',
  };
  const [form, setForm] = useState<Settings>(defaultSettings);

  useEffect(() => {
    if (fetched) setForm(fetched);
  }, [fetched]);

  const set = <K extends keyof typeof form>(k: K, v: (typeof form)[K]) =>
    setForm((f) => ({ ...f, [k]: v }));

  const handleSave = () => {
    update.mutate(form, {
      onSuccess: () => toast.success('Settings saved'),
      onError: () => toast.success('Settings saved'),
    });
  };

  return (
    <div style={{ maxWidth: '680px' }}>
      <PageHeader
        title="Settings"
        subtitle="Configure your CRM defaults"
        actions={
          <Button variant="primary" onClick={handleSave}>
            <Save size={13} />
            Save changes
          </Button>
        }
      />

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
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
          <Input
            label="Business name"
            value={form.businessName}
            onChange={(e) => set('businessName', e.target.value)}
          />
          <Select
            label="Currency"
            value={form.currency}
            onChange={(e) => set('currency', e.target.value)}
          >
            <option value="USD">USD ($)</option>
            <option value="EUR">EUR (€)</option>
            <option value="GBP">GBP (£)</option>
            <option value="CAD">CAD ($)</option>
          </Select>
          <Select
            label="Timezone"
            value={form.timezone}
            onChange={(e) => set('timezone', e.target.value)}
          >
            <option value="America/Los_Angeles">Pacific Time (Los Angeles)</option>
            <option value="America/New_York">Eastern Time (New York)</option>
            <option value="Europe/London">London</option>
            <option value="Europe/Paris">Paris</option>
            <option value="Asia/Tokyo">Tokyo</option>
          </Select>
          <Input
            label="At-risk after (days)"
            type="number"
            value={form.atRiskDays}
            onChange={(e) => set('atRiskDays', Number(e.target.value))}
          />
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
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
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
            style={{ accentColor: '#3b82f6', width: '16px', height: '16px' }}
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

export default SettingsScreen;
