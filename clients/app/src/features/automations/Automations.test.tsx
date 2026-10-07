import path from 'node:path';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
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

vi.mock('../../modules/crm/shared/hooks/queries/useAutomations', () => ({
  useListAutomations: () => mocks.automations,
  useUpdateAutomation: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteAutomation: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Automations from './Automations';

function renderAutomations(): void {
  render(<Automations />);
}

beforeEach(async () => {
  await initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales'), 'en', [
    path.resolve(import.meta.dirname, '../../modules/crm/locales'),
  ]);
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

    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
    expect(screen.queryByText('No automations yet')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(mocks.automations.refetch).toHaveBeenCalledTimes(1);
  });

  it('shows a "create one" call-to-action that navigates to the automation builder', () => {
    renderAutomations();

    expect(screen.getByText('No automations yet')).toBeInTheDocument();
    expect(
      screen.getByText('Create a workflow to run automatically when an order is placed.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Create one' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/automations/new' });
  });
});
