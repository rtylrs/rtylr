import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const searchCustomersMock = vi.hoisted(() => vi.fn());
const createCustomerMock = vi.hoisted(() => vi.fn());

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    searchCustomers: searchCustomersMock,
    createCustomer: createCustomerMock,
  }),
}));

vi.mock('../../layouts/ScreenLayout', () => ({
  default: ({
    title,
    actions,
    children,
  }: {
    title: string;
    actions?: React.ReactNode;
    children: React.ReactNode;
  }) => (
    <div>
      <h1>{title}</h1>
      <div>{actions}</div>
      {children}
    </div>
  ),
}));

import CustomerLookupScreen from './index';
import { useCartStore } from '../../shared/stores/cartStore';

describe('CustomerLookupScreen', () => {
  beforeEach(() => {
    searchCustomersMock.mockReset();
    createCustomerMock.mockReset();
    useCartStore.getState().clearCart();
  });

  afterEach(() => {
    cleanup();
  });

  it('displays loyalty points on search results and creates + attaches a new customer', async () => {
    searchCustomersMock.mockResolvedValue({
      customers: [
        {
          id: '14',
          firstName: 'Nadia',
          lastName: 'Khoury',
          phone: '+9613111222',
          email: 'nadia@example.com',
          loyaltyPoints: 120,
        },
      ],
    });
    createCustomerMock.mockResolvedValue({
      customer: {
        id: '25',
        firstName: 'Tariq',
        lastName: 'Mansour',
        phone: '+96170999888',
        email: 'tariq@example.com',
        loyaltyPoints: 0,
      },
    });

    render(
      <MemoryRouter>
        <CustomerLookupScreen />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByPlaceholderText('Search by name, phone, or email...'), {
      target: { value: 'Nadia' },
    });

    await waitFor(() => {
      expect(screen.getByText('Nadia Khoury')).toBeTruthy();
      expect(screen.getByText('120 loyalty points')).toBeTruthy();
    });

    fireEvent.click(screen.getByRole('button', { name: 'New Customer' }));
    fireEvent.change(screen.getByLabelText('First name'), { target: { value: 'Tariq' } });
    fireEvent.change(screen.getByLabelText('Last name'), { target: { value: 'Mansour' } });
    fireEvent.change(screen.getByLabelText('Phone number'), { target: { value: '+96170999888' } });
    fireEvent.change(screen.getByLabelText('Email address'), { target: { value: 'tariq@example.com' } });
    fireEvent.click(screen.getByRole('button', { name: 'Create & Attach' }));

    await waitFor(() => {
      expect(createCustomerMock).toHaveBeenCalledWith({
        firstName: 'Tariq',
        lastName: 'Mansour',
        phone: '+96170999888',
        email: 'tariq@example.com',
      });
    });

    expect(useCartStore.getState().customerId).toBe(25);
    expect(useCartStore.getState().customerName).toBe('Tariq Mansour');
  });
});
