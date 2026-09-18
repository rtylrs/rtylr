import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useCreateLead } from '../../../shared/hooks/queries/useSales';
import type { LeadStage } from '../../../shared/hooks/queries/useSales';
import { Modal, Button, Input, Select, Textarea } from '../../../shared/components/ui';

interface Props {
  open: boolean;
  onClose: () => void;
}

const CreateLeadModal: React.FC<Props> = ({ open, onClose }) => {
  const { t } = useTranslation();
  const createLead = useCreateLead();
  const [form, setForm] = useState({
    company: '',
    industry: '',
    size: '',
    source: '',
    stage: 'cold' as LeadStage,
    mrrPotential: '',
    assignedTo: '',
    notes: '',
  });

  const handleSubmit = (): void => {
    if (!form.company.trim()) {
      toast.error(t('crm.createLead.companyRequired'));
      return;
    }
    createLead.mutate(
      {
        company: form.company,
        industry: form.industry,
        size: form.size,
        source: form.source,
        stage: form.stage,
        mrrPotential: Number(form.mrrPotential) || 0,
        assignedTo: form.assignedTo ? Number(form.assignedTo) : undefined,
        notes: form.notes,
      },
      {
        onSuccess: () => {
          setForm({
            company: '',
            industry: '',
            size: '',
            source: '',
            stage: 'cold',
            mrrPotential: '',
            assignedTo: '',
            notes: '',
          });
          onClose();
        },
        onError: () => toast.error(t('crm.createLead.createFailed')),
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('crm.leads.new')}
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            disabled={createLead.isPending || !form.company.trim()}
          >
            {createLead.isPending ? 'Creating...' : 'Create lead'}
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
            value={form.mrrPotential}
            onChange={(e) => setForm((f) => ({ ...f, mrrPotential: e.target.value }))}
            placeholder="500"
          />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
          <Select
            label={t('crm.leadDetail.stage')}
            value={form.stage}
            onChange={(e) => setForm((f) => ({ ...f, stage: e.target.value as LeadStage }))}
          >
            {(
              [
                'cold',
                'contacted',
                'demo_booked',
                'demo_done',
                'proposal_sent',
                'contracted',
                'lost',
              ] as LeadStage[]
            ).map((stage) => (
              <option key={stage} value={stage}>
                {t(`crm.leadStages.${stage}`)}
              </option>
            ))}
          </Select>
          <Input
            label={t('crm.leadDetail.assignedRep')}
            type="number"
            value={form.assignedTo}
            onChange={(e) => setForm((f) => ({ ...f, assignedTo: e.target.value }))}
            placeholder={t('crm.createLead.userId')}
          />
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

export default CreateLeadModal;
