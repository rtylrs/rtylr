import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Modal } from '../../../shared/components/ui';
import { useCreateLead, useUpdateLead } from '../data/queries';
import type { Lead, LeadFields } from '../data/model';
import {
  BillingSelect,
  Field,
  FieldGrid,
  looksLikeEmail,
  parseAmountInput,
  TextInput,
} from './fields';
import { RepPicker } from './common';

/** A lead's editable fields, for forms and full updates. */
export function leadFieldsOf(lead?: Lead): LeadFields {
  return {
    firstName: lead?.firstName ?? '',
    lastName: lead?.lastName ?? '',
    jobTitle: lead?.jobTitle ?? '',
    email: lead?.email ?? '',
    phone: lead?.phone ?? '',
    companyName: lead?.companyName ?? '',
    website: lead?.website ?? '',
    linkedin: lead?.linkedin ?? '',
    industry: lead?.industry ?? '',
    size: lead?.size ?? '',
    city: lead?.city ?? '',
    country: lead?.country ?? '',
    source: lead?.source ?? '',
    assignedTo: lead?.assignedTo,
    value: lead?.value ? String(Number(lead.value)) : '',
    billing: lead?.billing || 'one_time',
    notes: lead?.notes ?? '',
  };
}

/** Creates a lead, or edits one when `lead` is given. */
const LeadFormModal: React.FC<{
  open: boolean;
  onClose: () => void;
  lead?: Lead;
  onCreated?: (lead: Lead) => void;
}> = ({ open, onClose, lead, onCreated }) => {
  const { t } = useTranslation();
  const [form, setForm] = useState<LeadFields>(() => leadFieldsOf(lead));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const create = useCreateLead();
  const update = useUpdateLead();
  const saving = create.isPending || update.isPending;

  useEffect(() => {
    if (open) {
      setForm(leadFieldsOf(lead));
      setErrors({});
    }
  }, [open, lead]);

  const set = (key: keyof LeadFields) => (value: string | undefined) =>
    setForm((current) => ({ ...current, [key]: value }));

  const submit = (event: React.FormEvent): void => {
    event.preventDefault();
    const next: Record<string, string> = {};
    const named = form.companyName?.trim() || form.firstName?.trim() || form.lastName?.trim();
    if (!named) next.name = t('crmSales.leadForm.needsName');
    if (form.email?.trim() && !looksLikeEmail(form.email))
      next.email = t('crmSales.leadForm.badEmail');
    const value = parseAmountInput(form.value ?? '');
    if (value === null) next.value = t('crmSales.leadForm.badAmount');
    setErrors(next);
    if (Object.keys(next).length > 0) return;

    const payload: LeadFields = { ...form, value: value || undefined };
    if (lead?.id) {
      update.mutate(
        {
          id: lead.id,
          lead: payload,
          status: lead.status,
          disqualifyReason: lead.disqualifyReason,
        },
        { onSuccess: onClose }
      );
    } else {
      create.mutate(payload, {
        onSuccess: (created) => {
          onClose();
          onCreated?.(created);
        },
      });
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      size="lg"
      title={lead ? t('crmSales.leadForm.editTitle') : t('crmSales.leadForm.createTitle')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmSales.common.cancel')}
          </Button>
          <Button type="submit" form="sales-lead-form" disabled={saving}>
            {saving
              ? t('crmSales.common.saving')
              : lead
                ? t('crmSales.common.save')
                : t('crmSales.leadForm.create')}
          </Button>
        </>
      }
    >
      <form
        id="sales-lead-form"
        onSubmit={submit}
        style={{ display: 'flex', flexDirection: 'column', gap: 16 }}
      >
        <section>
          <h3 style={{ margin: '0 0 8px', fontSize: 13 }}>{t('crmSales.leadForm.person')}</h3>
          <FieldGrid>
            <Field label={t('crmSales.fields.firstName')} error={errors.name}>
              {(id) => (
                <TextInput
                  id={id}
                  value={form.firstName}
                  onChange={set('firstName')}
                  autoComplete="off"
                />
              )}
            </Field>
            <Field label={t('crmSales.fields.lastName')}>
              {(id) => (
                <TextInput
                  id={id}
                  value={form.lastName}
                  onChange={set('lastName')}
                  autoComplete="off"
                />
              )}
            </Field>
            <Field label={t('crmSales.fields.jobTitle')}>
              {(id) => <TextInput id={id} value={form.jobTitle} onChange={set('jobTitle')} />}
            </Field>
            <Field label={t('crmSales.fields.phone')}>
              {(id) => <TextInput id={id} type="tel" value={form.phone} onChange={set('phone')} />}
            </Field>
            <Field label={t('crmSales.fields.email')} error={errors.email}>
              {(id) => (
                <TextInput id={id} type="email" value={form.email} onChange={set('email')} />
              )}
            </Field>
            <Field label={t('crmSales.fields.linkedin')}>
              {(id) => <TextInput id={id} value={form.linkedin} onChange={set('linkedin')} />}
            </Field>
          </FieldGrid>
        </section>
        <section>
          <h3 style={{ margin: '0 0 8px', fontSize: 13 }}>{t('crmSales.leadForm.company')}</h3>
          <FieldGrid>
            <Field label={t('crmSales.fields.company')}>
              {(id) => <TextInput id={id} value={form.companyName} onChange={set('companyName')} />}
            </Field>
            <Field label={t('crmSales.fields.website')}>
              {(id) => <TextInput id={id} value={form.website} onChange={set('website')} />}
            </Field>
            <Field label={t('crmSales.fields.industry')}>
              {(id) => <TextInput id={id} value={form.industry} onChange={set('industry')} />}
            </Field>
            <Field label={t('crmSales.fields.size')}>
              {(id) => <TextInput id={id} value={form.size} onChange={set('size')} />}
            </Field>
            <Field label={t('crmSales.fields.city')}>
              {(id) => <TextInput id={id} value={form.city} onChange={set('city')} />}
            </Field>
            <Field label={t('crmSales.fields.country')}>
              {(id) => <TextInput id={id} value={form.country} onChange={set('country')} />}
            </Field>
          </FieldGrid>
        </section>
        <section>
          <h3 style={{ margin: '0 0 8px', fontSize: 13 }}>{t('crmSales.leadForm.prospecting')}</h3>
          <FieldGrid>
            <Field label={t('crmSales.fields.source')} hint={t('crmSales.leadForm.sourceHint')}>
              {(id) => <TextInput id={id} value={form.source} onChange={set('source')} />}
            </Field>
            <Field label={t('crmSales.fields.rep')}>
              {(id) => (
                <RepPicker
                  id={id}
                  label={t('crmSales.fields.rep')}
                  value={form.assignedTo}
                  onChange={set('assignedTo')}
                />
              )}
            </Field>
            <Field label={t('crmSales.fields.estimatedValue')} error={errors.value}>
              {(id) => (
                <TextInput id={id} inputMode="decimal" value={form.value} onChange={set('value')} />
              )}
            </Field>
            <Field label={t('crmSales.fields.billing')}>
              {(id) => <BillingSelect id={id} value={form.billing} onChange={set('billing')} />}
            </Field>
            <Field label={t('crmSales.fields.notes')} wide>
              {(id) => (
                <textarea
                  id={id}
                  className="sales-textarea"
                  value={form.notes ?? ''}
                  onChange={(event) => set('notes')(event.target.value)}
                />
              )}
            </Field>
          </FieldGrid>
        </section>
      </form>
    </Modal>
  );
};

export default LeadFormModal;
