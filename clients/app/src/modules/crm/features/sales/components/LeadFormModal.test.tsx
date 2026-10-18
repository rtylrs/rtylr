import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  createLead: vi.fn(),
  updateLead: vi.fn(),
}));

vi.mock('../../../shared/hooks/queries/useSales', () => ({
  useCreateLead: () => ({ mutate: mocks.createLead, isPending: false }),
  useUpdateLead: () => ({ mutate: mocks.updateLead, isPending: false }),
  useSalesReps: () => ({
    data: [
      { id: 1, name: 'Dana Owner' },
      { id: 4, name: 'Sami Rep' },
    ],
  }),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import LeadFormModal from './LeadFormModal';
import type { Lead } from '../../../shared/hooks/queries/useSales';

const lead: Lead = {
  id: 9,
  companyName: 'Phoenicia Hotels',
  company: 'Phoenicia Hotels',
  industry: 'Hospitality',
  size: '200-500',
  source: 'Referral',
  stage: 'demo_booked',
  mrrPotential: 1200,
  assignedTo: 4,
  notes: 'Three branches',
  createdAt: '2026-10-18T00:00:00Z',
  updatedAt: '2026-10-18T00:00:00Z',
};

beforeEach(() => vi.clearAllMocks());

describe('LeadFormModal', () => {
  it('assigns a new lead to a teammate picked by name', () => {
    render(<LeadFormModal open onClose={vi.fn()} />);

    fireEvent.change(screen.getByLabelText('crm.createLead.companyLabel'), {
      target: { value: 'Cedar Bakery' },
    });
    fireEvent.change(screen.getByLabelText('crm.leadDetail.assignedRep'), {
      target: { value: '4' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'crm.createLead.submit' }));

    expect(mocks.createLead).toHaveBeenCalledWith(
      expect.objectContaining({ company: 'Cedar Bakery', stage: 'cold', assignedTo: 4 }),
      expect.anything()
    );
  });

  it('unassigns an existing lead and keeps its stage out of the edit', () => {
    render(<LeadFormModal open onClose={vi.fn()} lead={lead} />);

    expect(screen.queryByLabelText('crm.leadDetail.stage')).toBeNull();
    fireEvent.change(screen.getByLabelText('crm.leadDetail.assignedRep'), {
      target: { value: '' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'crmCommon.save' }));

    const [payload] = mocks.updateLead.mock.calls[0];
    expect(payload.id).toBe(9);
    expect(payload.data).toMatchObject({ company: 'Phoenicia Hotels', assignedTo: null });
    expect(payload.data).not.toHaveProperty('stage');
  });
});
