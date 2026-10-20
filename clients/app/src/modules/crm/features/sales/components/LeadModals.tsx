import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button, Modal } from '../../../shared/components/ui';
import { type Deal, type Lead, leadTitle, openStages } from '../data/model';
import { useConvertLead, usePipelines, useUpdateLead } from '../data/queries';
import { RepPicker } from './common';
import { BillingSelect, Field, FieldGrid, parseAmountInput, TextInput } from './fields';
import { leadFieldsOf } from './LeadFormModal';
import { ReasonPicker } from './ReasonPicker';

/**
 * Qualifies a lead: the person becomes a customer, the company is matched or
 * created, and, by default, a deal opens with the lead's open tasks.
 */
export const ConvertLeadModal: React.FC<{
  lead: Lead;
  open: boolean;
  onClose: () => void;
  onConverted: (deal?: Deal) => void;
}> = ({ lead, open, onClose, onConverted }) => {
  const { t } = useTranslation();
  const { data: pipelines = [] } = usePipelines();
  const convert = useConvertLead();
  const [createDeal, setCreateDeal] = useState(true);
  const [title, setTitle] = useState('');
  const [pipelineId, setPipelineId] = useState<string>('');
  const [stageId, setStageId] = useState<string>('');
  const [value, setValue] = useState('');
  const [billing, setBilling] = useState('one_time');
  const [closeDate, setCloseDate] = useState('');
  const [owner, setOwner] = useState<string | undefined>();
  const [error, setError] = useState('');

  const pipeline = useMemo(
    () =>
      pipelines.find((item) => item.id === pipelineId) ??
      pipelines.find((item) => item.isDefault) ??
      pipelines[0],
    [pipelines, pipelineId]
  );
  const stages = openStages(pipeline);

  useEffect(() => {
    if (!open) return;
    setCreateDeal(true);
    setTitle(leadTitle(lead));
    setPipelineId('');
    setStageId('');
    setValue(lead.value ? String(Number(lead.value)) : '');
    setBilling(lead.billing || 'one_time');
    setCloseDate('');
    setOwner(lead.assignedTo);
    setError('');
  }, [open, lead]);

  const submit = (): void => {
    const amount = parseAmountInput(value);
    if (createDeal && amount === null) {
      setError(t('crmSales.leadForm.badAmount'));
      return;
    }
    convert.mutate(
      {
        id: lead.id ?? '',
        createDeal,
        deal: createDeal
          ? {
              title: title.trim(),
              pipelineId: pipeline?.id,
              stageId: stageId || stages[0]?.id,
              value: amount || undefined,
              billing,
              expectedCloseDate: closeDate || undefined,
              assignedTo: owner,
            }
          : undefined,
      },
      {
        onSuccess: (result) => {
          onClose();
          onConverted(result.deal);
        },
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crmSales.convert.title')}
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('crmSales.common.cancel')}
          </Button>
          <Button onClick={submit} disabled={convert.isPending}>
            {convert.isPending ? t('crmSales.common.saving') : t('crmSales.convert.action')}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text-muted)', lineHeight: 1.5 }}>
          {t('crmSales.convert.explain')}
        </p>
        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
          <input
            type="checkbox"
            checked={createDeal}
            onChange={(event) => setCreateDeal(event.target.checked)}
          />
          {t('crmSales.convert.openDeal')}
        </label>
        {createDeal && (
          <FieldGrid>
            <Field label={t('crmSales.fields.dealTitle')} wide>
              {(id) => <TextInput id={id} value={title} onChange={setTitle} />}
            </Field>
            {pipelines.length > 1 && (
              <Field label={t('crmSales.fields.pipeline')}>
                {(id) => (
                  <select
                    id={id}
                    className="sales-select"
                    style={{ maxWidth: 'none' }}
                    value={pipeline?.id ?? ''}
                    onChange={(event) => {
                      setPipelineId(event.target.value);
                      setStageId('');
                    }}
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
            <Field label={t('crmSales.fields.stage')}>
              {(id) => (
                <select
                  id={id}
                  className="sales-select"
                  style={{ maxWidth: 'none' }}
                  value={stageId || stages[0]?.id || ''}
                  onChange={(event) => setStageId(event.target.value)}
                >
                  {stages.map((stage) => (
                    <option key={stage.id} value={stage.id}>
                      {stage.name}
                    </option>
                  ))}
                </select>
              )}
            </Field>
            <Field label={t('crmSales.fields.value')} error={error}>
              {(id) => <TextInput id={id} inputMode="decimal" value={value} onChange={setValue} />}
            </Field>
            <Field label={t('crmSales.fields.billing')}>
              {(id) => <BillingSelect id={id} value={billing} onChange={setBilling} />}
            </Field>
            <Field label={t('crmSales.fields.expectedClose')}>
              {(id) => <TextInput id={id} type="date" value={closeDate} onChange={setCloseDate} />}
            </Field>
            <Field label={t('crmSales.fields.owner')}>
              {(id) => (
                <RepPicker
                  id={id}
                  label={t('crmSales.fields.owner')}
                  value={owner}
                  onChange={setOwner}
                />
              )}
            </Field>
          </FieldGrid>
        )}
      </div>
    </Modal>
  );
};

const DISQUALIFY_REASONS = [
  'notAFit',
  'noBudget',
  'noNeed',
  'unreachable',
  'wrongContact',
  'competitor',
  'doNotContact',
] as const;

/** Drops a lead with a reason, or parks it as nurture for later. */
export const DisqualifyModal: React.FC<{
  lead: Lead;
  open: boolean;
  onClose: () => void;
}> = ({ lead, open, onClose }) => {
  const { t } = useTranslation();
  const update = useUpdateLead();
  const [reason, setReason] = useState('');

  useEffect(() => {
    if (open) setReason('');
  }, [open]);

  const apply = (status: 'disqualified' | 'nurture'): void => {
    update.mutate(
      {
        id: lead.id ?? '',
        lead: { ...leadFieldsOf(lead), value: lead.value },
        status,
        disqualifyReason: status === 'disqualified' ? reason.trim() : undefined,
      },
      { onSuccess: onClose }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crmSales.disqualify.title')}
      footer={
        <>
          <Button variant="outline" onClick={() => apply('nurture')} disabled={update.isPending}>
            {t('crmSales.disqualify.nurtureInstead')}
          </Button>
          <Button
            variant="danger"
            onClick={() => apply('disqualified')}
            disabled={update.isPending || !reason.trim()}
          >
            {t('crmSales.disqualify.action')}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text-muted)', lineHeight: 1.5 }}>
          {t('crmSales.disqualify.explain')}
        </p>
        <ReasonPicker
          reasons={DISQUALIFY_REASONS}
          labelPrefix="crmSales.disqualify.reasons"
          value={reason}
          onChange={setReason}
          label={t('crmSales.disqualify.reason')}
        />
      </div>
    </Modal>
  );
};
