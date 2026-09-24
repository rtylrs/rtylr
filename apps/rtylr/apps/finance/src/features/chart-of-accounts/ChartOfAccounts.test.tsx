import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import ChartOfAccounts from './ChartOfAccounts';
import { initTestI18n } from '../../shared/testing/i18n';

const { createAccountMock } = vi.hoisted(() => ({
  createAccountMock: vi.fn(),
}));

vi.mock('../../shared/hooks/useChartOfAccounts', () => ({
  useChartOfAccounts: () => ({ data: [], isLoading: false }),
  useCreateAccount: () => ({ mutateAsync: createAccountMock, isPending: false }),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: { user: { role: string } }) => unknown) =>
    selector({ user: { role: 'owner' } }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
  createAccountMock.mockResolvedValue({});
});

describe('ChartOfAccounts account creation', () => {
  it('sends the trimmed optional account code', async () => {
    render(<ChartOfAccounts />);
    fireEvent.click(screen.getByRole('button', { name: 'New Account' }));

    fireEvent.change(screen.getByPlaceholderText('Name'), { target: { value: 'Operating cash' } });
    fireEvent.change(screen.getByPlaceholderText('Account code (optional)'), {
      target: { value: ' 4000 ' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Create' }));

    await waitFor(() =>
      expect(createAccountMock).toHaveBeenCalledWith({
        name: 'Operating cash',
        code: '4000',
        type: 'asset',
        currency: 'USD',
        balance: 0,
        isActive: true,
      })
    );
  });
});
