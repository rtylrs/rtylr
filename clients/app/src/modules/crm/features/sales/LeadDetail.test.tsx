import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '1' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const lead = {
  id: 1,
  company: 'Acme Corp',
  companyName: 'Acme Corp',
  industry: 'SaaS',
  source: 'Referral',
  stage: 'cold' as const,
  size: '50-200',
  mrrPotential: 500,
  dealValue: 500,
  assignedTo: null,
};

// Stable empty-array references: LeadDetail reads these straight from the
// query hooks, and a mock returning a fresh array on every render would
// re-trigger any effect that depends on it.
const touchpointsList: unknown[] = [];
const contactsList: unknown[] = [];
const contractsList: unknown[] = [];

vi.mock('../../shared/hooks/queries/useSales', () => ({
  useLead: () => ({ data: lead, isLoading: false }),
  useUpdateLead: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteLead: () => ({ mutate: vi.fn(), isPending: false }),
  useLeadTouchpoints: () => ({ data: touchpointsList }),
  useCreateTouchpoint: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteTouchpoint: () => ({ mutate: vi.fn(), isPending: false }),
  useContacts: () => ({ data: contactsList }),
  useCreateContact: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteContact: () => ({ mutate: vi.fn(), isPending: false }),
  useContracts: () => ({ data: contractsList }),
  useCreateContract: () => ({ mutate: vi.fn(), isPending: false }),
}));

import LeadDetail from './LeadDetail';

function renderLeadDetail(): void {
  render(<LeadDetail />);
}

beforeEach(() => {
  touchpointsList.length = 0;
  contactsList.length = 0;
  contractsList.length = 0;
});

describe('LeadDetail touchpoints empty state', () => {
  it('shows a "log touchpoint" call-to-action that opens the touchpoint form', () => {
    renderLeadDetail();

    expect(screen.getByText('crm.leadDetail.noTouchpointsTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.leadDetail.noTouchpointsDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.leadDetail.logTouchpoint' }));

    // Opens the inline touchpoint form rather than mutating anything itself.
    expect(screen.getByPlaceholderText('crm.leadDetail.notesPlaceholder')).toBeInTheDocument();
  });
});

describe('LeadDetail contacts empty state', () => {
  it('shows an "add contact" call-to-action that opens the contact form', () => {
    renderLeadDetail();

    expect(screen.getByText('crm.leadDetail.noContactsTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.leadDetail.noContactsDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.leadDetail.addContact' }));

    // Opens the inline contact form (and the empty state itself disappears).
    expect(screen.getByPlaceholderText('crm.leadDetail.namePlaceholder')).toBeInTheDocument();
    expect(screen.queryByText('crm.leadDetail.noContactsTitle')).not.toBeInTheDocument();
  });
});
