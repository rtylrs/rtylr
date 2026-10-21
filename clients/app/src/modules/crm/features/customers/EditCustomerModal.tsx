import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Input, Modal } from '../../shared/components/ui';
import { useUpdateCustomer } from '../../shared/hooks/queries/useCustomers';
import type { Customer } from '../../shared/schemas/customers';

type ConsentKey = 'email' | 'sms' | 'marketing';

const CONSENT_KEYS: ReadonlyArray<{ key: ConsentKey; label: string }> = [
  { key: 'email', label: 'crm.createCustomer.emailUpdates' },
  { key: 'sms', label: 'crm.createCustomer.smsUpdates' },
  { key: 'marketing', label: 'crm.createCustomer.marketing' },
];

interface Draft {
  firstName: string;
  lastName: string;
  email: string;
  phone: string;
  dateOfBirth: string;
  address: string;
  consent: Record<ConsentKey, boolean>;
}

function draftFrom(customer: Customer): Draft {
  return {
    firstName: customer.firstName ?? '',
    lastName: customer.lastName ?? '',
    email: customer.email ?? '',
    phone: customer.phone ?? '',
    // The wire may carry a full timestamp; the date input takes YYYY-MM-DD.
    dateOfBirth: (customer.dateOfBirth ?? '').slice(0, 10),
    address: customer.address ?? '',
    consent: {
      email: customer.consent?.email ?? false,
      sms: customer.consent?.sms ?? false,
      marketing: customer.consent?.marketing ?? false,
    },
  };
}

const EditCustomerModal: React.FC<{
  customer: Customer;
  open: boolean;
  onClose: () => void;
}> = ({ customer, open, onClose }) => {
  const { t } = useTranslation();
  const update = useUpdateCustomer();
  const [draft, setDraft] = useState<Draft>(() => draftFrom(customer));

  // Reopening starts from the saved record, not from an abandoned edit.
  useEffect(() => {
    if (open) setDraft(draftFrom(customer));
  }, [open, customer]);

  const set = <K extends keyof Draft>(key: K, value: Draft[K]): void =>
    setDraft((prev) => ({ ...prev, [key]: value }));

  const canSave = draft.firstName.trim().length > 0 && !update.isPending;

  const save = (): void => {
    if (!canSave) return;
    update.mutate(
      {
        id: customer.id,
        data: {
          // The whole record is the base so tags, notes and the rest survive
          // even when the customer is not in the query cache the update
          // merges over. First and last name go separately because a cached
          // firstName would otherwise win over a new combined `name`.
          ...customer,
          firstName: draft.firstName.trim(),
          lastName: draft.lastName.trim(),
          email: draft.email.trim(),
          phone: draft.phone.trim(),
          dateOfBirth: draft.dateOfBirth || undefined,
          address: draft.address.trim(),
          consent: { ...customer.consent, ...draft.consent },
        },
      },
      { onSuccess: onClose }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crm.customerProfile.editTitle')}
      footer={
        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
          <Button variant="outline" onClick={onClose}>
            {t('crmCommon.cancel')}
          </Button>
          <Button variant="primary" onClick={save} disabled={!canSave}>
            {update.isPending ? t('crm.states.saving') : t('crmCommon.save')}
          </Button>
        </div>
      }
    >
      <form
        onSubmit={(event) => {
          event.preventDefault();
          save();
        }}
        style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}
      >
        <div style={{ display: 'flex', gap: '12px' }}>
          <Input
            label={t('crm.customerProfile.firstName')}
            value={draft.firstName}
            onChange={(e) => set('firstName', e.target.value)}
            required
            autoFocus
          />
          <Input
            label={t('crm.customerProfile.lastName')}
            value={draft.lastName}
            onChange={(e) => set('lastName', e.target.value)}
          />
        </div>
        <Input
          label={t('crmCommon.email')}
          type="email"
          value={draft.email}
          onChange={(e) => set('email', e.target.value)}
        />
        <Input
          label={t('crmCommon.phone')}
          type="tel"
          value={draft.phone}
          onChange={(e) => set('phone', e.target.value)}
        />
        <Input
          label={t('crm.createCustomer.birthday')}
          type="date"
          value={draft.dateOfBirth}
          onChange={(e) => set('dateOfBirth', e.target.value)}
        />
        <Input
          label={t('crm.createCustomer.address')}
          placeholder={t('crm.createCustomer.addressPlaceholder')}
          value={draft.address}
          onChange={(e) => set('address', e.target.value)}
        />
        <fieldset style={{ border: 'none', padding: 0, margin: 0 }}>
          <legend
            style={{
              fontSize: '12px',
              fontWeight: 600,
              color: 'var(--hub-text-muted)',
              marginBottom: '6px',
            }}
          >
            {t('crm.createCustomer.stepConsent')}
          </legend>
          {CONSENT_KEYS.map(({ key, label }) => (
            <label
              key={key}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                fontSize: '13px',
                color: 'var(--hub-text)',
                padding: '4px 0',
              }}
            >
              <input
                type="checkbox"
                checked={draft.consent[key]}
                onChange={(e) => set('consent', { ...draft.consent, [key]: e.target.checked })}
              />
              {t(label)}
            </label>
          ))}
        </fieldset>
        {/* Lets Enter submit from any field. */}
        <button type="submit" hidden />
      </form>
    </Modal>
  );
};

export default EditCustomerModal;
