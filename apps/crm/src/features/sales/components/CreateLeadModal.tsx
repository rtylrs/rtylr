import React, { useState } from 'react';
import { toast } from 'sonner';
import { useCreateLead } from '../../../shared/hooks/queries/useSales';
import type { LeadStage } from '../../../shared/hooks/queries/useSales';
import { Modal, Button, Input, Select, Textarea } from '../../../shared/components/ui';

interface Props {
  open: boolean;
  onClose: () => void;
}

const CreateLeadModal: React.FC<Props> = ({ open, onClose }) => {
  const createLead = useCreateLead();
  const [form, setForm] = useState({
    company: '',
    industry: '',
    size: '',
    source: '',
    stage: 'cold' as LeadStage,
    mrrPotential: '',
    assignedRep: '',
    notes: '',
  });

  const handleSubmit = (): void => {
    if (!form.company.trim()) {
      toast.error('Company name is required');
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
        assignedRep: form.assignedRep,
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
            assignedRep: '',
            notes: '',
          });
          onClose();
        },
        onError: () => toast.error('Failed to create lead'),
      }
    );
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="New Lead"
      size="md"
      footer={
        <>
          <Button variant="outline" onClick={onClose}>
            Cancel
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
          label="Company *"
          value={form.company}
          onChange={(e) => setForm((f) => ({ ...f, company: e.target.value }))}
          placeholder="Acme Corp"
        />
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
          <Input
            label="Industry"
            value={form.industry}
            onChange={(e) => setForm((f) => ({ ...f, industry: e.target.value }))}
            placeholder="SaaS"
          />
          <Input
            label="Company size"
            value={form.size}
            onChange={(e) => setForm((f) => ({ ...f, size: e.target.value }))}
            placeholder="50-200"
          />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
          <Input
            label="Source"
            value={form.source}
            onChange={(e) => setForm((f) => ({ ...f, source: e.target.value }))}
            placeholder="LinkedIn, Referral..."
          />
          <Input
            label="MRR Potential ($)"
            type="number"
            value={form.mrrPotential}
            onChange={(e) => setForm((f) => ({ ...f, mrrPotential: e.target.value }))}
            placeholder="500"
          />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px' }}>
          <Select
            label="Stage"
            value={form.stage}
            onChange={(e) => setForm((f) => ({ ...f, stage: e.target.value as LeadStage }))}
          >
            <option value="cold">Cold</option>
            <option value="contacted">Contacted</option>
            <option value="demo_booked">Demo Booked</option>
            <option value="demo_done">Demo Done</option>
            <option value="proposal_sent">Proposal Sent</option>
            <option value="contracted">Contracted</option>
            <option value="lost">Lost</option>
          </Select>
          <Input
            label="Assigned Rep"
            value={form.assignedRep}
            onChange={(e) => setForm((f) => ({ ...f, assignedRep: e.target.value }))}
            placeholder="Your name"
          />
        </div>
        <Textarea
          label="Notes"
          value={form.notes}
          onChange={(e) => setForm((f) => ({ ...f, notes: e.target.value }))}
          placeholder="Context about this lead..."
        />
      </div>
    </Modal>
  );
};

export default CreateLeadModal;
