import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Modal } from '../../../shared/components/ui';
import { type Contract, type Deal, todayISO } from '../data/model';
import { useCreateContract, useDealList, useUpdateContract } from '../data/queries';
import { BillingSelect, Field, FieldGrid, parseAmountInput, TextInput } from './fields';

interface Terms {
  dealId: string;
  plan: string;
  value: string;
  billing: string;
  seats: string;
  startDate: string;
  endDate: string;
  signedAt: string;
  signedByName: string;
  notes: string;
}

function termsOf(contract?: Contract, deal?: Deal): Terms {
  const value = contract?.value ?? deal?.value;
  return {
    dealId: contract?.dealId ?? deal?.id ?? '',
    plan: contract?.plan ?? deal?.title ?? '',
    value: value ? String(Number(value)) : '',
    billing: contract?.billing ?? deal?.billing ?? 'monthly',
    seats: String(contract?.seats ?? 1),
    startDate: contract?.startDate ?? (contract ? '' : todayISO()),
    endDate: contract?.endDate ?? '',
    signedAt: contract?.signedAt ?? '',
    signedByName: contract?.signedByName ?? deal?.contactName ?? '',
    notes: contract?.notes ?? '',
  };
}

/** Drafts a contract (for `deal`, or a won deal picked here) or edits one. */
const ContractModal: React.FC<{
  open: boolean;
  onClose: () => void;
  contract?: Contract;
  deal?: Deal;
}> = ({ open, onClose, contract, deal }) => {
  const { t } = useTranslation();
  const create = useCreateContract();
  const update = useUpdateContract();
  const wonDeals = useDealList({ statuses: ['won'], rep: 'all', search: '', sort: 'newest' }, 100);
  const [terms, setTerms] = useState<Terms>(() => termsOf(contract, deal));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const saving = create.isPending || update.isPending;
  const choosable = !contract && !deal;
  const deals = wonDeals.data?.pages.flatMap((page) => page.deals ?? []) ?? [];

  useEffect(() => {
    if (open) {
      setTerms(termsOf(contract, deal));
      setErrors({});
    }
  }, [open, contract, deal]);

  const set = (key: keyof Terms) => (value: string) =>
    setTerms((current) => ({ ...current, [key]: value }));

  const submit = (event: React.FormEvent): void => {
    event.preventDefault();
    const next: Record<string, string> = {};
    if (!terms.plan.trim()) next.plan = t('crmSales.contracts.needsPlan');
    const value = parseAmountInput(terms.value);
    if (value === null || value === '') next.value = t('crmSales.leadForm.badAmount');
    if (terms.endDate && terms.startDate && terms.endDate < terms.startDate)
      next.endDate = t('crmSales.contracts.endBeforeStart');
    setErrors(next);
    if (Object.keys(next).length > 0 || !value) return;
    const body = {
      plan: terms.plan.trim(),
      value,
      billing: terms.billing,
      seats: Math.max(1, Number(terms.seats) || 1),
      startDate: terms.startDate || undefined,
      endDate: terms.endDate || undefined,
      signedAt: terms.signedAt || undefined,
      signedByName: terms.signedByName.trim(),
      notes: terms.notes,
    };
    if (contract?.id) {
      update.mutate({ id: contract.id, body }, { onSuccess: onClose });
    } else {
      create.mutate({ ...body, dealId: terms.dealId || undefined }, { onSuccess: onClose });
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      size="lg"
      title={contract ? t('crmSales.contracts.editTitle') : t('crmSales.contracts.createTitle')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmSales.common.cancel')}
          </Button>
          <Button type="submit" form="sales-contract-form" disabled={saving}>
            {saving ? t('crmSales.common.saving') : t('crmSales.common.save')}
          </Button>
        </>
      }
    >
      <form id="sales-contract-form" onSubmit={submit}>
        <FieldGrid>
          {choosable && (
            <Field label={t('crmSales.fields.deal')} hint={t('crmSales.contracts.dealHint')} wide>
              {(id) => (
                <select
                  id={id}
                  className="sales-select"
                  style={{ maxWidth: 'none' }}
                  value={terms.dealId}
                  onChange={(event) => {
                    const picked = deals.find((item) => item.id === event.target.value);
                    setTerms(picked ? { ...termsOf(undefined, picked) } : { ...terms, dealId: '' });
                  }}
                >
                  <option value="">{t('crmSales.contracts.noDeal')}</option>
                  {deals.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.title}
                    </option>
                  ))}
                </select>
              )}
            </Field>
          )}
          <Field label={t('crmSales.fields.plan')} error={errors.plan} wide>
            {(id) => <TextInput id={id} value={terms.plan} onChange={set('plan')} />}
          </Field>
          <Field label={t('crmSales.fields.value')} error={errors.value}>
            {(id) => (
              <TextInput id={id} inputMode="decimal" value={terms.value} onChange={set('value')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.billing')}>
            {(id) => <BillingSelect id={id} value={terms.billing} onChange={set('billing')} />}
          </Field>
          <Field label={t('crmSales.fields.seats')}>
            {(id) => (
              <TextInput id={id} inputMode="numeric" value={terms.seats} onChange={set('seats')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.startDate')}>
            {(id) => (
              <TextInput id={id} type="date" value={terms.startDate} onChange={set('startDate')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.endDate')} error={errors.endDate}>
            {(id) => (
              <TextInput id={id} type="date" value={terms.endDate} onChange={set('endDate')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.signedAt')} hint={t('crmSales.win.signedHint')}>
            {(id) => (
              <TextInput id={id} type="date" value={terms.signedAt} onChange={set('signedAt')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.signedBy')}>
            {(id) => (
              <TextInput id={id} value={terms.signedByName} onChange={set('signedByName')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.notes')} wide>
            {(id) => (
              <textarea
                id={id}
                className="sales-textarea"
                value={terms.notes}
                onChange={(event) => set('notes')(event.target.value)}
              />
            )}
          </Field>
        </FieldGrid>
      </form>
    </Modal>
  );
};

export default ContractModal;
