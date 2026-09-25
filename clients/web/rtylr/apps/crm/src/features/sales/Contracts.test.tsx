import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const refetchContracts = vi.fn();

// Stable references: Contracts.tsx's effect depends on this array's identity,
// and a mock that returns a new array on every render would loop forever.
const contractsList: unknown[] = [];
const leadsList: unknown[] = [];

const mocks = vi.hoisted(() => ({
  isError: false,
}));

vi.mock('../../shared/hooks/queries/useSales', () => ({
  useContracts: () => ({
    data: contractsList,
    isLoading: false,
    isFetching: false,
    isError: mocks.isError,
    refetch: refetchContracts,
  }),
  useLeads: () => ({ data: leadsList }),
  useCreateContract: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Contracts from './Contracts';

function renderContracts(): void {
  render(<Contracts />);
}

beforeEach(() => {
  refetchContracts.mockClear();
  mocks.isError = false;
  contractsList.length = 0;
});

describe('Contracts load-failed state', () => {
  it('shows a "retry" call-to-action when contracts fail to load', () => {
    mocks.isError = true;
    renderContracts();

    expect(screen.getByText('crm.contracts.loadFailedTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.retry' }));

    expect(refetchContracts).toHaveBeenCalledTimes(1);
  });
});

describe('Contracts true-empty state', () => {
  // The header action and the empty-state action are the same action offered
  // twice, and since both went through crm.contracts.new they now carry the
  // same accessible name. getAllBy is deliberate: it asserts that both are
  // present rather than silently matching whichever comes first.
  const newContractButtons = (): HTMLElement[] =>
    screen.getAllByRole('button', { name: 'crm.contracts.new' });

  it('shows a "new contract" call-to-action that opens the create-contract modal', () => {
    renderContracts();

    expect(screen.getByText('crm.contracts.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.contracts.emptyDescription')).toBeInTheDocument();

    const buttons = newContractButtons();
    expect(buttons).toHaveLength(2);
    fireEvent.click(buttons[1]);

    expect(screen.getByRole('heading', { name: 'crm.contracts.new' })).toBeInTheDocument();
  });

  it('the pre-existing header action opens the same modal', () => {
    renderContracts();

    fireEvent.click(newContractButtons()[0]);

    expect(screen.getByRole('heading', { name: 'crm.contracts.new' })).toBeInTheDocument();
  });
});
