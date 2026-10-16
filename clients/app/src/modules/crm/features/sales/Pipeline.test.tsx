import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Lead, PipelineColumn } from '../../shared/hooks/queries/useSales';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  updateLeadMutate: vi.fn(),
  refetchPipeline: vi.fn(),
  refetchLeads: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const mockPipeline: PipelineColumn[] = [
  { stage: 'cold', count: 1, totalMrr: 1500 },
  { stage: 'contacted', count: 0, totalMrr: 0 },
  { stage: 'demo_booked', count: 0, totalMrr: 0 },
  { stage: 'demo_done', count: 0, totalMrr: 0 },
  { stage: 'proposal_sent', count: 0, totalMrr: 0 },
  { stage: 'contracted', count: 1, totalMrr: 5000 },
  { stage: 'lost', count: 0, totalMrr: 0 },
];

const mockLeads: Lead[] = [
  {
    id: 1,
    company: 'Acme Corp',
    companyName: 'Acme Corp',
    industry: 'SaaS',
    source: 'Referral',
    stage: 'cold',
    assignedTo: null,
    mrrPotential: 1500,
    dealValue: 1500,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: 2,
    company: 'Globex Inc',
    companyName: 'Globex Inc',
    industry: 'Fintech',
    source: 'Website',
    stage: 'contracted',
    assignedTo: 42,
    mrrPotential: 5000,
    dealValue: 5000,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

vi.mock('../../shared/hooks/queries/useSales', () => ({
  usePipeline: () => ({
    data: mockPipeline,
    isPending: false,
    isError: false,
    refetch: mocks.refetchPipeline,
  }),
  useLeads: () => ({
    data: mockLeads,
    isPending: false,
    isFetching: false,
    isError: false,
    refetch: mocks.refetchLeads,
  }),
  useUpdateLead: () => ({
    mutate: mocks.updateLeadMutate,
    isPending: false,
  }),
  useCreateLead: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

import Pipeline from './Pipeline';

function renderPipeline(): void {
  render(<Pipeline />);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('Pipeline Board & Stage Moves', () => {
  it('renders stages with counts and deal totals', () => {
    renderPipeline();

    expect(screen.getByText('crm.leadStages.cold')).toBeInTheDocument();
    expect(screen.getAllByText('$1,500/mo').length).toBe(2);
    expect(screen.getByText('Acme Corp')).toBeInTheDocument();
    expect(screen.getAllByText('$5,000/mo').length).toBe(2);
    expect(screen.getByText('Globex Inc')).toBeInTheDocument();
  });

  it('allows advancing an active lead to the next pipeline stage', () => {
    renderPipeline();

    // Acme Corp is in cold stage; next stage button is Contacted
    const nextButtons = screen.getAllByTitle(/crm.pipeline.nextStage/);
    expect(nextButtons.length).toBeGreaterThan(0);

    fireEvent.click(nextButtons[0]);

    expect(mocks.updateLeadMutate).toHaveBeenCalledWith({
      id: 1,
      data: {
        stage: 'contacted',
        companyName: 'Acme Corp',
      },
    });
  });

  it('allows marking an active lead as lost', () => {
    renderPipeline();

    const markLostButtons = screen.getAllByTitle('crm.pipeline.markLost');
    expect(markLostButtons.length).toBeGreaterThan(0);

    fireEvent.click(markLostButtons[0]);

    expect(mocks.updateLeadMutate).toHaveBeenCalledWith({
      id: 1,
      data: {
        stage: 'lost',
        companyName: 'Acme Corp',
      },
    });
  });

  it('does not display stage transition buttons on terminal stages (contracted, lost)', () => {
    renderPipeline();

    // Globex Inc is contracted (terminal); only 1 advance and 1 lost button (for Acme Corp) should exist
    const nextButtons = screen.getAllByTitle(/crm.pipeline.nextStage/);
    expect(nextButtons.length).toBe(1);

    const markLostButtons = screen.getAllByTitle('crm.pipeline.markLost');
    expect(markLostButtons.length).toBe(1);
  });
});
