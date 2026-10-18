import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  isError: false,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
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
  assignedTo: null,
  mrrPotential: 500,
  dealValue: 500,
};

// Stable references: Leads.tsx's effect depends on this array's identity,
// and a mock that returns a new array on every render would loop forever.
const leadsList: (typeof lead)[] = [];
const refetchLeads = vi.fn();
const deleteLeadMutate = vi.fn();

vi.mock('../../shared/hooks/queries/useSales', () => ({
  useSalesReps: () => ({ data: [] }),
  useUpdateLead: () => ({ mutate: vi.fn(), isPending: false }),
  useLeads: () => ({
    data: leadsList,
    isLoading: false,
    isFetching: false,
    isError: mocks.isError,
    refetch: refetchLeads,
  }),
  useDeleteLead: () => ({ mutate: deleteLeadMutate, isPending: false }),
  useCreateLead: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Leads from './Leads';

function renderLeads(): void {
  render(<Leads />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  mocks.isError = false;
  refetchLeads.mockClear();
  leadsList.length = 0;
});

describe('Leads load-failed state', () => {
  it('shows a "retry" call-to-action when leads fail to load', () => {
    mocks.isError = true;
    renderLeads();

    expect(screen.getByText('crm.leads.loadFailedTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.retry' }));

    expect(refetchLeads).toHaveBeenCalledTimes(1);
  });
});

describe('Leads search-empty state', () => {
  it('shows a "clear search" call-to-action when a search term matches nothing', () => {
    leadsList.push(lead);
    renderLeads();

    fireEvent.change(screen.getByPlaceholderText('crm.leads.searchPlaceholder'), {
      target: { value: 'nonexistent' },
    });

    expect(screen.getByText('crm.leads.emptySearchDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.clearSearch' }));

    expect(screen.getByText('Acme Corp')).toBeInTheDocument();
  });
});

describe('Leads true-empty state', () => {
  it('shows a "new lead" call-to-action that opens the create-lead modal', () => {
    renderLeads();

    expect(screen.getByText('crm.leads.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.leads.createFirstDescription')).toBeInTheDocument();

    // Header and empty-state actions share crm.leads.new, so both carry the
    // same accessible name; assert both exist rather than matching one by luck.
    const buttons = screen.getAllByRole('button', { name: 'crm.leads.new' });
    expect(buttons).toHaveLength(2);
    fireEvent.click(buttons[1]);

    expect(screen.getByRole('heading', { name: 'crm.leads.new' })).toBeInTheDocument();
  });
});
