import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  updateCustomer: vi.fn(),
}));

vi.mock('../../shared/services/crmApi', () => ({
  crmApi: { updateCustomer: mocks.updateCustomer },
}));

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<object>();
  return { ...actual, unwrap: vi.fn(async (value: unknown) => value) };
});

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import EditCustomerModal from './EditCustomerModal';
import { customerSchema, type Customer } from '../../shared/schemas/customers';

const customer: Customer = customerSchema.parse({
  id: '7',
  firstName: 'Rania',
  lastName: 'Haddad',
  email: 'rania@example.test',
  phone: '+96170123456',
  dateOfBirth: '1990-05-14T00:00:00Z',
  address: 'Hamra St, Beirut',
  tags: ['vip'],
  notes: 'Prefers email',
  consent: { email: true, sms: false, marketing: true },
});

function renderModal(onClose = vi.fn()): () => void {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(
    <QueryClientProvider client={queryClient}>
      <EditCustomerModal customer={customer} open onClose={onClose} />
    </QueryClientProvider>
  );
  return onClose;
}

beforeEach(() => {
  vi.clearAllMocks();
  mocks.updateCustomer.mockResolvedValue({ message: 'ok' });
});

describe('EditCustomerModal', () => {
  it('saves the edited details without dropping tags, notes or other consent', async () => {
    const onClose = renderModal();

    fireEvent.change(screen.getByLabelText('crm.customerProfile.lastName'), {
      target: { value: 'Haddad-Khoury' },
    });
    fireEvent.click(screen.getByLabelText('crm.createCustomer.marketing'));
    fireEvent.click(screen.getByRole('button', { name: 'crmCommon.save' }));

    await waitFor(() => expect(mocks.updateCustomer).toHaveBeenCalledTimes(1));
    const [id, payload] = mocks.updateCustomer.mock.calls[0];
    expect(id).toBe(7);
    expect(payload).toMatchObject({
      firstName: 'Rania',
      lastName: 'Haddad-Khoury',
      email: 'rania@example.test',
      dateOfBirth: '1990-05-14',
      address: 'Hamra St, Beirut',
      tags: ['vip'],
      notes: 'Prefers email',
      consent: { email: true, sms: false, marketing: false },
    });
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it('will not save a customer without a first name', () => {
    renderModal();

    fireEvent.change(screen.getByLabelText('crm.customerProfile.firstName'), {
      target: { value: '  ' },
    });

    expect(screen.getByRole('button', { name: 'crmCommon.save' })).toBeDisabled();
  });
});
