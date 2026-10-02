import React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import ChartOfAccounts from './ChartOfAccounts';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const { createAccountMock } = vi.hoisted(() => ({
  createAccountMock: vi.fn(),
}));
const authState = vi.hoisted(() => ({ effectiveAccess: undefined as unknown }));

vi.mock('../../shared/hooks/useChartOfAccounts', () => ({
  useChartOfAccounts: () => ({ data: [], isLoading: false }),
  useCreateAccount: () => ({ mutateAsync: createAccountMock, isPending: false }),
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (
      selector: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      selector({
        user: { effectiveAccess: authState.effectiveAccess },
        effectiveAccess: authState.effectiveAccess,
      }),
    {
      subscribe: vi.fn(),
      getState: () => ({
        user: { effectiveAccess: authState.effectiveAccess },
        effectiveAccess: authState.effectiveAccess,
      }),
    }
  ),
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
  authState.effectiveAccess = createEffectiveAccess('finance', { isOwner: true });
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
