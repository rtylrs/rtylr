import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  role: 'admin' as string,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: { user: { role: string } }) => unknown) =>
    selector({ user: { role: mocks.role } }),
}));

vi.mock('../../shared/hooks/queries/useAutomations', () => ({
  useListAutomations: () => ({ data: [], isLoading: false }),
  useUpdateAutomation: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteAutomation: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Automations from './Automations';

function renderAutomations(): void {
  render(<Automations />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
});

describe('Automations empty state', () => {
  it('shows a "create one" call-to-action that navigates to the automation builder', () => {
    renderAutomations();

    expect(screen.getByText('crm.automations.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.automations.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.automations.createOne' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/automations/new' });
  });
});
