import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useCreateLead, useSalesReps, useUpdateLead } from '../../../shared/hooks/queries/useSales';
import type { Lead, LeadStage } from '../../../shared/hooks/queries/useSales';
import { Modal, Button, Input, Select, Textarea } from '../../../shared/components/ui';

const STAGES: LeadStage[] = [
  'cold',
  'contacted',
  'demo_booked',
  'demo_done',
  'proposal_sent',
  'contracted',
  'lost',
];

interface Props {
  open: boolean;
  onClose: () => void;
  /** Edits this lead when given; creates a new one otherwise. */
  lead?: Lead;
}

interface Form {
  company: string;
  industry: string;
  size: string;
  source: string;
  stage: LeadStage;
  mrrPotential: string;
  assignedTo: string;
  notes: string;
}

const EMPTY_FORM: Form = {
  company: '',
  industry: '',
  size: '',
  source: '',
  stage: 'cold',
  mrrPotential: '',
  assignedTo: '',
  notes: '',
};

function formFrom(lead?: Lead): Form {
  if (!lead) return { ...EMPTY_FORM };
  return {
    company: lead.companyName ?? lead.company ?? '',
    industry: lead.industry ?? '',
    size: lead.size ?? '',
    source: lead.source ?? '',
    stage: lead.stage,
    mrrPotential:
      lead.mrrPotential !== undefined && lead.mrrPotential !== 0 ? String(lead.mrrPotential) : '',
    assignedTo: lead.assignedTo != null ? String(lead.assignedTo) : '',
    notes: lead.notes ?? '',
  };
}

function parseMrrPotential(value: string): number | null {
  const normalized = value.trim();
  if (normalized === '') return 0;
  const parsed = Number(normalized);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : null;
}

const LeadFormModal: React.FC<Props> = ({ open, onClose, lead }) => {
  const { t } = useTranslation();
  const createLead = useCreateLead();
  const updateLead = useUpdateLead();
  const { data: reps = [] } = useSalesReps();
  const submissionId = useRef(0);
  const [form, setForm] = useState<Form>(() => formFrom(lead));
  const mrrPotential = parseMrrPotential(form.mrrPotential);
  const editing = lead !== undefined;
  const pending = createLead.isPending || updateLead.isPending;

  // Reopening starts from the saved lead, not from an abandoned edit.
  useEffect(() => {
    if (open) setForm(formFrom(lead));
  }, [open, lead]);

  const closeModal = (): void => {
    submissionId.current += 1;
    onClose();
  };

  const handleSubmit = (): void => {
    if (!form.company.trim()) {
      toast.error(t('crm.createLead.companyRequired'));
      return;
    }
    if (mrrPotential === null) {
      toast.error(t('crmErrors.generic'));
      return;
    }

    const requestId = ++submissionId.current;
    const fields = {
      company: form.company.trim(),
      companyName: form.company.trim(),
      industry: form.industry,
      size: form.size,
      source: form.source,
      mrrPotential,
      notes: form.notes,
    };
    const onSuccess = (): void => {
      if (requestId !== submissionId.current) return;
      closeModal();
    };

    if (lead) {
      // Stage changes go through the stage control, which enforces the
      // pipeline's forward-only rule, so editing keeps the current stage.
      updateLead.mutate(
        {
          id: lead.id,
          data: { ...fields, assignedTo: form.assignedTo ? Number(form.assignedTo) : null },
        },
        { onSuccess }
      );
      return;
    }
    createLead.mutate(
      {
        ...fields,
        stage: form.stage,
        assignedTo: form.assignedTo ? Number(form.assignedTo) : undefined,
      },
      { onSuccess }
    );
  };

  const submitLabel = editing
    ? pending
      ? t('crm.states.saving')
      : t('crmCommon.save')
    : pending
      ? t('crm.createLead.creating')
      : t('crm.createLead.submit');

  return (
    <Modal
      open={open}
      onClose={closeModal}
      title={editing ? t('crm.createLead.editTitle') : t('crm.leads.new')}
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={closeModal}>
            {t('crmCommon.cancel')}
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            disabled={pending || !form.company.trim() || mrrPotential === null}
          >
            {submitLabel}
          </Button>
        </>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        <Input
          label={t('crm.createLead.companyLabel')}
          value={form.company}
          onChange={(e) => setForm((f) => ({ ...f, company: e.target.value }))}
          placeholder={t('crm.createLead.companyPlaceholder')}
        />
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
          <Input
            label={t('crm.leadDetail.industry')}
            value={form.industry}
            onChange={(e) => setForm((f) => ({ ...f, industry: e.target.value }))}
            placeholder={t('crm.createLead.industryPlaceholder')}
          />
          <Input
            label={t('crm.createLead.companySize')}
            value={form.size}
            onChange={(e) => setForm((f) => ({ ...f, size: e.target.value }))}
            placeholder="50-200"
          />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
          <Input
            label={t('crm.leadDetail.source')}
            value={form.source}
            onChange={(e) => setForm((f) => ({ ...f, source: e.target.value }))}
            placeholder={t('crm.createLead.sourcePlaceholder')}
          />
          <Input
            label={t('crm.createLead.mrrPotentialUsd')}
            type="number"
            min={0}
            step="any"
            aria-invalid={mrrPotential === null}
            value={form.mrrPotential}
            onChange={(e) => setForm((f) => ({ ...f, mrrPotential: e.target.value }))}
            placeholder="500"
          />
        </div>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: editing ? '1fr' : '1fr 1fr',
            gap: '10px',
          }}
        >
          {!editing && (
            <Select
              label={t('crm.leadDetail.stage')}
              value={form.stage}
              onChange={(e) => setForm((f) => ({ ...f, stage: e.target.value as LeadStage }))}
            >
              {STAGES.map((stage) => (
                <option key={stage} value={stage}>
                  {t(`crm.leadStages.${stage}`)}
                </option>
              ))}
            </Select>
          )}
          <Select
            label={t('crm.leadDetail.assignedRep')}
            value={form.assignedTo}
            onChange={(e) => setForm((f) => ({ ...f, assignedTo: e.target.value }))}
          >
            <option value="">{t('crm.createLead.unassigned')}</option>
            {reps.map((rep) => (
              <option key={rep.id} value={String(rep.id)}>
                {rep.name}
              </option>
            ))}
          </Select>
        </div>
        <Textarea
          label={t('crm.leadDetail.notesLabel')}
          value={form.notes}
          onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
          placeholder={t('crm.createLead.notesPlaceholder')}
        />
      </div>
    </Modal>
  );
};

export default LeadFormModal;
