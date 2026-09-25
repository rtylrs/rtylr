import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  createCustomer: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
  Link: ({ children }: { children?: React.ReactNode }) => <span>{children}</span>,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/services/crmApi', () => ({
  crmApi: { createCustomer: mocks.createCustomer },
}));

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<object>();
  return { ...actual, unwrap: vi.fn(async (value: unknown) => value) };
});

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { toast } from 'sonner';
import CreateCustomer from './CreateCustomer';

let queryClient: QueryClient;

function renderWizard(): void {
  render(
    <QueryClientProvider client={queryClient}>
      <CreateCustomer />
    </QueryClientProvider>
  );
}

function fillWizardAndSubmit(): void {
  fireEvent.change(screen.getByLabelText('crm.createCustomer.fullName'), {
    target: { value: 'Jane Doe' },
  });
  fireEvent.click(screen.getByRole('button', { name: /crm\.createCustomer\.next/ }));
  fireEvent.change(screen.getByLabelText('common.email'), {
    target: { value: 'jane@example.com' },
  });
  fireEvent.click(screen.getByRole('button', { name: /crm\.createCustomer\.next/ }));
  fireEvent.click(screen.getByRole('button', { name: /crm\.createCustomer\.submit/ }));
}

beforeEach(() => {
  vi.clearAllMocks();
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
});

describe('CreateCustomer step-0 cancel (B57-R01)', () => {
  it('leaves the wizard when Cancel is clicked on the first step', () => {
    renderWizard();

    const cancel = screen.getByRole('button', { name: 'common.cancel' });
    expect(cancel).toBeEnabled();
    fireEvent.click(cancel);

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/customers' });
  });
});

describe('CreateCustomer submit notifications (B57-R01)', () => {
  it('shows the created notification once', async () => {
    mocks.createCustomer.mockResolvedValue({ customerId: 5 });
    renderWizard();

    fillWizardAndSubmit();

    await waitFor(() => expect(mocks.navigate).toHaveBeenCalledWith({ to: '/customers' }));
    expect(mocks.createCustomer).toHaveBeenCalledTimes(1);
    expect(toast.success).toHaveBeenCalledTimes(1);
    expect(toast.success).toHaveBeenCalledWith('crm.customers.created');
  });

  it('shows the create-failed notification once', async () => {
    mocks.createCustomer.mockRejectedValue(new Error('boom'));
    renderWizard();

    fillWizardAndSubmit();

    await waitFor(() => expect(toast.error).toHaveBeenCalledTimes(1));
    expect(mocks.createCustomer).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith('crm.customers.createFailed');
  });
});
