import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { Check, ArrowLeft, ArrowRight } from 'lucide-react';
import { useCreateCustomer } from '../../shared/hooks/queries/useCustomers';
import { Breadcrumbs, Button, Card, Input, PageHeader } from '../../shared/components/ui';

interface FormData {
  name: string;
  email: string;
  phone: string;
  birthday: string;
  address: string;
  consentEmail: boolean;
  consentSms: boolean;
  consentMarketing: boolean;
}

// Ids only: this is module scope, evaluated before i18next has a language.
const STEP_KEYS = ['info', 'contact', 'consent'] as const;

const CreateCustomer: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const create = useCreateCustomer();
  const [step, setStep] = useState(0);
  const [data, setData] = useState<FormData>({
    name: '',
    email: '',
    phone: '',
    birthday: '',
    address: '',
    consentEmail: false,
    consentSms: false,
    consentMarketing: false,
  });

  const set = <K extends keyof FormData>(k: K, v: FormData[K]) =>
    setData((d) => ({ ...d, [k]: v }));

  const canNext = () => {
    if (step === 0) return data.name.trim().length > 0;
    if (step === 1) return data.email.trim().length > 0 || data.phone.trim().length > 0;
    return true;
  };

  const handleSubmit = () => {
    create.mutate(
      {
        name: data.name,
        firstName: data.name.trim().split(/\s+/)[0] ?? '',
        lastName: data.name.trim().split(/\s+/).slice(1).join(' '),
        email: data.email,
        phone: data.phone,
        dateOfBirth: data.birthday,
        address: data.address,
        consent: {
          email: data.consentEmail,
          sms: data.consentSms,
          marketing: data.consentMarketing,
        },
      },
      {
        onSuccess: () => {
          navigate({ to: '/crm/customers' });
        },
      }
    );
  };

  return (
    <>
      <HelmetContainer title={`${t('crm.createCustomer.pageTitle')} · rtylr`} />
      <div style={{ maxWidth: '680px' }}>
        <Breadcrumbs
          items={[
            { label: t('crmNav.customers'), to: '/crm/customers' },
            { label: t('crm.customers.breadcrumbNew') },
          ]}
        />
        <PageHeader
          title={t('crm.createCustomer.pageTitle')}
          subtitle={t('crm.createCustomer.subtitle')}
        />

        {/* Stepper */}
        <div style={{ display: 'flex', gap: '8px', marginBottom: '20px' }}>
          {STEP_KEYS.map((stepKey, i) => (
            <div
              key={stepKey}
              style={{
                flex: 1,
                padding: '10px 14px',
                borderRadius: '10px',
                background: i <= step ? 'rgba(59,130,246,0.1)' : 'var(--hub-card)',
                border:
                  i === step ? '1px solid rgba(59,130,246,0.4)' : '1px solid var(--hub-border)',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
              }}
            >
              <div
                style={{
                  width: '22px',
                  height: '22px',
                  borderRadius: '50%',
                  background:
                    i < step ? '#3b82f6' : i === step ? 'rgba(59,130,246,0.2)' : 'var(--hub-hover)',
                  color: i < step ? '#fff' : i === step ? '#3b82f6' : 'var(--hub-text-muted)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '11px',
                  fontWeight: 700,
                }}
              >
                {i < step ? <Check size={12} /> : i + 1}
              </div>
              <span
                style={{
                  fontSize: '12px',
                  fontWeight: 600,
                  color: i <= step ? 'var(--hub-text)' : 'var(--hub-text-muted)',
                }}
              >
                {t(`crm.createCustomer.step${stepKey[0].toUpperCase()}${stepKey.slice(1)}`)}
              </span>
            </div>
          ))}
        </div>

        <Card>
          {step === 0 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
              <Input
                label={t('crm.createCustomer.fullName')}
                value={data.name}
                onChange={(e) => set('name', e.target.value)}
                placeholder={t('crm.leadDetail.namePlaceholder')}
              />
              <Input
                label={t('crm.createCustomer.birthday')}
                type="date"
                value={data.birthday}
                onChange={(e) => set('birthday', e.target.value)}
              />
              <Input
                label={t('crm.createCustomer.address')}
                value={data.address}
                onChange={(e) => set('address', e.target.value)}
                placeholder={t('crm.createCustomer.addressPlaceholder')}
              />
            </div>
          )}
          {step === 1 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
              <Input
                label={t('crmCommon.email')}
                type="email"
                value={data.email}
                onChange={(e) => set('email', e.target.value)}
                placeholder="jane@example.com"
              />
              <Input
                label={t('crmCommon.phone')}
                type="tel"
                value={data.phone}
                onChange={(e) => set('phone', e.target.value)}
                placeholder="+1 555-000-0000"
              />
            </div>
          )}
          {step === 2 && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              {[
                {
                  key: 'consentEmail' as const,
                  label: t('crm.createCustomer.emailUpdates'),
                  desc: t('crm.createCustomer.consentEmailDesc'),
                },
                {
                  key: 'consentSms' as const,
                  label: t('crm.createCustomer.smsUpdates'),
                  desc: t('crm.createCustomer.consentSmsDesc'),
                },
                {
                  key: 'consentMarketing' as const,
                  label: t('crm.createCustomer.marketing'),
                  desc: t('crm.createCustomer.consentMarketingDesc'),
                },
              ].map((c) => (
                <label
                  key={c.key}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '12px',
                    padding: '12px 14px',
                    borderRadius: '10px',
                    background: 'var(--hub-hover)',
                    cursor: 'pointer',
                    border: '1px solid var(--hub-border)',
                  }}
                >
                  <input
                    type="checkbox"
                    checked={data[c.key]}
                    onChange={(e) => set(c.key, e.target.checked)}
                    style={{ accentColor: '#3b82f6', width: '16px', height: '16px' }}
                  />
                  <div>
                    <div style={{ fontSize: '13px', fontWeight: 600, color: 'var(--hub-text)' }}>
                      {c.label}
                    </div>
                    <div style={{ fontSize: '11px', color: 'var(--hub-text-muted)' }}>{c.desc}</div>
                  </div>
                </label>
              ))}
            </div>
          )}
        </Card>

        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '16px' }}>
          <Button
            variant="outline"
            onClick={() => (step > 0 ? setStep(step - 1) : navigate({ to: '/crm/customers' }))}
          >
            <ArrowLeft size={13} />
            {step === 0 ? t('crmCommon.cancel') : t('crm.createCustomer.back')}
          </Button>
          {step < STEP_KEYS.length - 1 ? (
            <Button variant="primary" disabled={!canNext()} onClick={() => setStep(step + 1)}>
              {t('crm.createCustomer.next')} <ArrowRight size={13} />
            </Button>
          ) : (
            <Button
              variant="primary"
              onClick={handleSubmit}
              disabled={create.isPending || !data.name.trim()}
            >
              {create.isPending ? t('crm.createCustomer.creating') : t('crm.createCustomer.submit')}
            </Button>
          )}
        </div>
      </div>
    </>
  );
};

export default CreateCustomer;
