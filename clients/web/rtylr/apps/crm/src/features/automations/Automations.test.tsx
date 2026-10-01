import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  effectiveAccess: null as unknown,
  automations: {
    data: [] as unknown[],
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => {
  const state = () => ({
    user: { effectiveAccess: mocks.effectiveAccess },
    effectiveAccess: mocks.effectiveAccess,
  });
  return {
    useUserStore: Object.assign(
      (selector: (value: ReturnType<typeof state>) => unknown) => selector(state()),
      { getState: state }
    ),
  };
});

vi.mock('../../shared/hooks/queries/useAutomations', () => ({
  useListAutomations: () => mocks.automations,
  useUpdateAutomation: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteAutomation: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Automations from './Automations';

function renderAutomations(): void {
  render(<Automations />);
}

beforeEach(() => {
  mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
  mocks.navigate.mockClear();
  mocks.automations.data = [];
  mocks.automations.isLoading = false;
  mocks.automations.isError = false;
  mocks.automations.refetch.mockClear();
});

describe('Automations list states', () => {
  it('shows a retryable error instead of a false empty state when loading fails', () => {
    mocks.automations.isError = true;

    renderAutomations();

    expect(screen.getByRole('alert')).toHaveTextContent('errors.generic');
    expect(screen.queryByText('crm.automations.emptyTitle')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));
    expect(mocks.automations.refetch).toHaveBeenCalledTimes(1);
  });

  it('shows a "create one" call-to-action that navigates to the automation builder', () => {
    renderAutomations();

    expect(screen.getByText('crm.automations.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.automations.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.automations.createOne' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/automations/new' });
  });
});
