import { render, screen, fireEvent } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import ClientSelect from './ClientSelect';
import { initTestI18n } from '../testing/i18n';

const { listCrmCustomersMock } = vi.hoisted(() => ({ listCrmCustomersMock: vi.fn() }));

vi.mock('../lib/clients', () => ({
  listCrmCustomers: listCrmCustomersMock,
}));

// The transport boundary, not the behavior under test: unwrap turns an
// ApiError into a rejection, which the mocked list call already represents.
vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => Promise.resolve(value),
}));

// i18next is not mocked here: the real locale files mean an assertion
// written in English also proves the key resolves (rtylr#1269).
beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  listCrmCustomersMock.mockReset();
});

const openPicker = () => fireEvent.click(screen.getByRole('button', { name: 'Select client' }));

describe('ClientSelect CRM load states', () => {
  it('shows the generated CRM customer name and retains its ID when selected', async () => {
    listCrmCustomersMock.mockResolvedValue({
      customers: [
        { id: '42', firstName: 'Acme', lastName: 'Customer', email: 'acme@example.test' },
      ],
    });
    const onChange = vi.fn();
    render(<ClientSelect value={null} onChange={onChange} />);
    openPicker();
    fireEvent.click(await screen.findByRole('button', { name: /Acme Customer/ }));
    expect(onChange).toHaveBeenCalledWith(
      expect.objectContaining({ id: '42', crmCustomerId: '42', name: 'Acme Customer' })
    );
  });

  it('shows a retryable error instead of "No clients found" when the CRM load fails', async () => {
    listCrmCustomersMock.mockRejectedValue(new Error('network down'));
    render(<ClientSelect value={null} onChange={vi.fn()} />);

    openPicker();

    expect(await screen.findByRole('alert')).toHaveTextContent("Couldn't load clients");
    expect(screen.queryByText('No clients found')).not.toBeInTheDocument();
  });

  it('reloads the directory when the retry action is used', async () => {
    listCrmCustomersMock
      .mockRejectedValueOnce(new Error('network down'))
      .mockResolvedValue({ customers: [{ id: 42, name: 'Acme Corp' }] });
    render(<ClientSelect value={null} onChange={vi.fn()} />);

    openPicker();
    fireEvent.click(await screen.findByRole('button', { name: 'Try again' }));

    expect(await screen.findByText('Acme Corp')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('does not report an empty directory while the request is in flight', async () => {
    let resolveLoad: (value: unknown) => void = () => {};
    listCrmCustomersMock.mockReturnValue(
      new Promise((resolve) => {
        resolveLoad = resolve;
      })
    );
    render(<ClientSelect value={null} onChange={vi.fn()} />);

    openPicker();

    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.queryByText('No clients found')).not.toBeInTheDocument();

    resolveLoad({ customers: [] });

    expect(await screen.findByText('No clients found')).toBeInTheDocument();
  });
});
