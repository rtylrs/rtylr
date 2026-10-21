import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Modal } from '../../../shared/components/ui';
import { type Deal, openStages, todayISO } from '../data/model';
import {
  useCreateDeal,
  useLoseDeal,
  usePipelines,
  useUpdateDeal,
  useWinDeal,
} from '../data/queries';
import { RepPicker } from './common';
import { BillingSelect, Field, FieldGrid, parseAmountInput, TextInput } from './fields';
import { ReasonPicker } from './ReasonPicker';

/** Opens a deal, or edits one when `deal` is given. */
export const DealFormModal: React.FC<{
  open: boolean;
  onClose: () => void;
  deal?: Deal;
  pipelineId?: string;
  stageId?: string;
  onCreated?: (deal: Deal) => void;
}> = ({ open, onClose, deal, pipelineId, stageId, onCreated }) => {
  const { t } = useTranslation();
  const { data: pipelines = [] } = usePipelines();
  const create = useCreateDeal();
  const update = useUpdateDeal();
  const [form, setForm] = useState({
    title: '',
    contactName: '',
    companyName: '',
    value: '',
    billing: 'one_time',
    expectedCloseDate: '',
    source: '',
    notes: '',
    assignedTo: undefined as string | undefined,
    pipelineId: '',
    stageId: '',
  });
  const [error, setError] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!open) return;
    setForm({
      title: deal?.title ?? '',
      contactName: deal?.contactName ?? '',
      companyName: deal?.companyName ?? '',
      value: deal?.value ? String(Number(deal.value)) : '',
      billing: deal?.billing || 'one_time',
      expectedCloseDate: deal?.expectedCloseDate ?? '',
      source: deal?.source ?? '',
      notes: deal?.notes ?? '',
      assignedTo: deal?.assignedTo,
      pipelineId: pipelineId ?? '',
      stageId: stageId ?? '',
    });
    setError({});
  }, [open, deal, pipelineId, stageId]);

  const pipeline = useMemo(
    () =>
      pipelines.find((item) => item.id === form.pipelineId) ??
      pipelines.find((item) => item.isDefault) ??
      pipelines[0],
    [pipelines, form.pipelineId]
  );
  const stages = openStages(pipeline);
  const set = (key: keyof typeof form) => (value: string | undefined) =>
    setForm((current) => ({ ...current, [key]: value }));
  const saving = create.isPending || update.isPending;

  const submit = (event: React.FormEvent): void => {
    event.preventDefault();
    const next: Record<string, string> = {};
    if (!form.title.trim() && !form.companyName.trim() && !form.contactName.trim()) {
      next.title = t('crmSales.dealForm.needsTitle');
    }
    const value = parseAmountInput(form.value);
    if (value === null) next.value = t('crmSales.leadForm.badAmount');
    setError(next);
    if (Object.keys(next).length > 0) return;
    const fields = {
      title: form.title.trim(),
      contactName: form.contactName.trim(),
      companyName: form.companyName.trim(),
      value: value || '0',
      billing: form.billing,
      expectedCloseDate: form.expectedCloseDate,
      source: form.source.trim(),
      notes: form.notes,
      assignedTo: form.assignedTo,
    };
    if (deal?.id) {
      update.mutate({ id: deal.id, deal: fields }, { onSuccess: onClose });
    } else {
      create.mutate(
        { ...fields, pipelineId: pipeline?.id, stageId: form.stageId || stages[0]?.id },
        {
          onSuccess: (created) => {
            onClose();
            onCreated?.(created);
          },
        }
      );
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      size="lg"
      title={deal ? t('crmSales.dealForm.editTitle') : t('crmSales.dealForm.createTitle')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmSales.common.cancel')}
          </Button>
          <Button type="submit" form="sales-deal-form" disabled={saving}>
            {saving ? t('crmSales.common.saving') : t('crmSales.common.save')}
          </Button>
        </>
      }
    >
      <form id="sales-deal-form" onSubmit={submit}>
        <FieldGrid>
          <Field label={t('crmSales.fields.dealTitle')} error={error.title} wide>
            {(id) => <TextInput id={id} value={form.title} onChange={set('title')} />}
          </Field>
          <Field label={t('crmSales.fields.contact')}>
            {(id) => <TextInput id={id} value={form.contactName} onChange={set('contactName')} />}
          </Field>
          <Field label={t('crmSales.fields.company')}>
            {(id) => <TextInput id={id} value={form.companyName} onChange={set('companyName')} />}
          </Field>
          {!deal && pipelines.length > 1 && (
            <Field label={t('crmSales.fields.pipeline')}>
              {(id) => (
                <select
                  id={id}
                  className="sales-select"
                  style={{ maxWidth: 'none' }}
                  value={pipeline?.id ?? ''}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      pipelineId: event.target.value,
                      stageId: '',
                    }))
                  }
                >
                  {pipelines.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </select>
              )}
            </Field>
          )}
          {!deal && (
            <Field label={t('crmSales.fields.stage')}>
              {(id) => (
                <select
                  id={id}
                  className="sales-select"
                  style={{ maxWidth: 'none' }}
                  value={form.stageId || stages[0]?.id || ''}
                  onChange={(event) => set('stageId')(event.target.value)}
                >
                  {stages.map((stage) => (
                    <option key={stage.id} value={stage.id}>
                      {stage.name}
                    </option>
                  ))}
                </select>
              )}
            </Field>
          )}
          <Field label={t('crmSales.fields.value')} error={error.value}>
            {(id) => (
              <TextInput id={id} inputMode="decimal" value={form.value} onChange={set('value')} />
            )}
          </Field>
          <Field label={t('crmSales.fields.billing')}>
            {(id) => <BillingSelect id={id} value={form.billing} onChange={set('billing')} />}
          </Field>
          <Field label={t('crmSales.fields.expectedClose')}>
            {(id) => (
              <TextInput
                id={id}
                type="date"
                value={form.expectedCloseDate}
                onChange={set('expectedCloseDate')}
              />
            )}
          </Field>
          <Field label={t('crmSales.fields.owner')}>
            {(id) => (
              <RepPicker
                id={id}
                label={t('crmSales.fields.owner')}
                value={form.assignedTo}
                onChange={set('assignedTo')}
              />
            )}
          </Field>
          <Field label={t('crmSales.fields.source')}>
            {(id) => <TextInput id={id} value={form.source} onChange={set('source')} />}
          </Field>
          <Field label={t('crmSales.fields.notes')} wide>
            {(id) => (
              <textarea
                id={id}
                className="sales-textarea"
                value={form.notes}
                onChange={(event) => set('notes')(event.target.value)}
              />
            )}
          </Field>
        </FieldGrid>
      </form>
    </Modal>
  );
};

/** Closes a deal as won and drafts its contract from the deal's terms. */
export const WinDealModal: React.FC<{
  deal: Deal;
  open: boolean;
  onClose: () => void;
}> = ({ deal, open, onClose }) => {
  const { t } = useTranslation();
  const win = useWinDeal();
  const [draft, setDraft] = useState(true);
  const [terms, setTerms] = useState({
    plan: '',
    value: '',
    billing: 'one_time',
    startDate: '',
    endDate: '',
    signedAt: '',
    signedByName: '',
  });
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open) return;
    setDraft(true);
    setTerms({
      plan: deal.title ?? '',
      value: deal.value ? String(Number(deal.value)) : '',
      billing: deal.billing || 'one_time',
      startDate: todayISO(),
      endDate: '',
      signedAt: '',
      signedByName: deal.contactName ?? '',
    });
    setError('');
  }, [open, deal]);

  const set = (key: keyof typeof terms) => (value: string) =>
    setTerms((current) => ({ ...current, [key]: value }));

  const submit = (): void => {
    const value = parseAmountInput(terms.value);
    if (draft && value === null) {
      setError(t('crmSales.leadForm.badAmount'));
      return;
    }
    win.mutate(
      {
        id: deal.id ?? '',
        contract: draft
          ? {
              plan: terms.plan.trim(),
              value: value || undefined,
              billing: terms.billing,
              startDate: terms.startDate || undefined,
              endDate: terms.endDate || undefined,
              signedAt: terms.signedAt || undefined,
              signedByName: terms.signedByName.trim(),
            }
          : undefined,
      },
      { onSuccess: onClose }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crmSales.win.title')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmSales.common.cancel')}
          </Button>
          <Button onClick={submit} disabled={win.isPending}>
            {win.isPending ? t('crmSales.common.saving') : t('crmSales.win.action')}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
          <input
            type="checkbox"
            checked={draft}
            onChange={(event) => setDraft(event.target.checked)}
          />
          {t('crmSales.win.draftContract')}
        </label>
        {draft && (
          <FieldGrid>
            <Field label={t('crmSales.fields.plan')} wide>
              {(id) => <TextInput id={id} value={terms.plan} onChange={set('plan')} />}
            </Field>
            <Field label={t('crmSales.fields.value')} error={error}>
              {(id) => (
                <TextInput
                  id={id}
                  inputMode="decimal"
                  value={terms.value}
                  onChange={set('value')}
                />
              )}
            </Field>
            <Field label={t('crmSales.fields.billing')}>
              {(id) => <BillingSelect id={id} value={terms.billing} onChange={set('billing')} />}
            </Field>
            <Field label={t('crmSales.fields.startDate')}>
              {(id) => (
                <TextInput
                  id={id}
                  type="date"
                  value={terms.startDate}
                  onChange={set('startDate')}
                />
              )}
            </Field>
            <Field label={t('crmSales.fields.endDate')}>
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
          </FieldGrid>
        )}
      </div>
    </Modal>
  );
};

const LOSS_REASONS = [
  'price',
  'competitor',
  'noDecision',
  'timing',
  'noResponse',
  'notAFit',
  'features',
] as const;

export const LoseDealModal: React.FC<{
  deal: Deal;
  open: boolean;
  onClose: () => void;
}> = ({ deal, open, onClose }) => {
  const { t } = useTranslation();
  const lose = useLoseDeal();
  const [reason, setReason] = useState('');
  useEffect(() => {
    if (open) setReason('');
  }, [open]);
  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crmSales.lose.title')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmSales.common.cancel')}
          </Button>
          <Button
            variant="danger"
            disabled={lose.isPending || !reason.trim()}
            onClick={() =>
              lose.mutate({ id: deal.id ?? '', reason: reason.trim() }, { onSuccess: onClose })
            }
          >
            {t('crmSales.lose.action')}
          </Button>
        </>
      }
    >
      <ReasonPicker
        reasons={LOSS_REASONS}
        labelPrefix="crmSales.lose.reasons"
        value={reason}
        onChange={setReason}
        label={t('crmSales.lose.reason')}
      />
    </Modal>
  );
};
