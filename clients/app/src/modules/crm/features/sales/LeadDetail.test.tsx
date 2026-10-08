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
const queryErrors = { lead: false, touchpoints: false, contacts: false, contracts: false };

vi.mock('../../shared/hooks/queries/useSales', () => ({
  useLead: () => ({
    data: queryErrors.lead ? undefined : lead,
    isLoading: false,
    isError: queryErrors.lead,
    refetch: vi.fn(),
  }),
  useUpdateLead: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteLead: () => ({ mutate: vi.fn(), isPending: false }),
  useLeadTouchpoints: () => ({
    data: touchpointsList,
    isError: queryErrors.touchpoints,
    isLoading: false,
    isFetching: false,
    refetch: vi.fn(),
  }),
  useCreateTouchpoint: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteTouchpoint: () => ({ mutate: vi.fn(), isPending: false }),
  useContacts: () => ({
    data: contactsList,
    isError: queryErrors.contacts,
    isLoading: false,
    isFetching: false,
    refetch: vi.fn(),
  }),
  useCreateContact: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteContact: () => ({ mutate: vi.fn(), isPending: false }),
  useContracts: () => ({
    data: contractsList,
    isError: queryErrors.contracts,
    isLoading: false,
    isFetching: false,
    isSuccess: !queryErrors.contracts,
    refetch: vi.fn(),
  }),
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
  queryErrors.lead = false;
  queryErrors.touchpoints = false;
  queryErrors.contacts = false;
  queryErrors.contracts = false;
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

describe('LeadDetail list errors', () => {
  it('does not report an empty timeline when touchpoints fail to load', () => {
    queryErrors.touchpoints = true;
    renderLeadDetail();

    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    expect(screen.queryByText('crm.leadDetail.noTouchpointsTitle')).not.toBeInTheDocument();
  });

  it('does not report no contacts when the contact list fails to load', () => {
    queryErrors.contacts = true;
    renderLeadDetail();

    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    expect(screen.queryByText('crm.leadDetail.noContactsTitle')).not.toBeInTheDocument();
  });

  it('hides contract creation until the contract list is available', () => {
    queryErrors.contracts = true;
    renderLeadDetail();

    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    expect(screen.queryByText('crm.leadDetail.noContract')).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'crm.leadDetail.addContractAria' })
    ).not.toBeInTheDocument();
  });

  it('does not label a failed lead lookup as a missing lead', () => {
    queryErrors.lead = true;
    renderLeadDetail();

    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    expect(screen.queryByText('crm.leadDetail.notFound')).not.toBeInTheDocument();
  });
});
