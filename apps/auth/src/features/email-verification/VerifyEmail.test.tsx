// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';

const { verifyMutate, resendMutate, search, navigateMock, notifyMock } = vi.hoisted(() => ({
  verifyMutate: vi.fn(),
  resendMutate: vi.fn(),
  search: { value: '' },
  navigateMock: vi.fn(),
  notifyMock: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
  useRouterState: (opts: { select: (s: unknown) => unknown }) =>
    opts.select({ location: { searchStr: search.value } }),
}));

// Returns the key, so an assertion that finds a bare `verify.*` string is
// proof the namespace is wrong rather than proof the copy is missing.
vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('./useEmailVerification', () => ({
  useEmailVerification: () => ({ mutate: verifyMutate, isPending: false }),
  useResendVerificationEmail: () => ({ mutate: resendMutate, isPending: false }),
}));

vi.mock('../../shared/components/Logo', () => ({ default: () => null }));
vi.mock('../../shared/lib/notify/notify', () => ({ default: notifyMock }));
vi.mock('@voxire/components', () => ({
  Button: ({ children, ...rest }: { children: React.ReactNode }) => (
    <button type="button" {...rest}>
      {children}
    </button>
  ),
}));

import VerifyEmail from './VerifyEmail';

describe('VerifyEmail (#1027)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    search.value = '';
  });

  it('asks for the code instead of reporting a failure that has not happened', () => {
    render(<VerifyEmail />);

    // onion emails a six-digit code and never a link, so arriving with no
    // token is the normal path, not an error.
    expect(screen.getByText('auth.verify.enterCodePrompt')).toBeInTheDocument();
    expect(screen.queryByText('auth.verify.codeInvalid')).not.toBeInTheDocument();
    expect(verifyMutate).not.toHaveBeenCalled();
  });

  it('offers the code form on that first view', () => {
    render(<VerifyEmail />);

    expect(screen.getByPlaceholderText('auth.verify.codePlaceholder')).toBeInTheDocument();
    expect(screen.getByText('auth.verify.verifyCode')).toBeInTheDocument();
  });

  it('resolves every string through the auth namespace', () => {
    render(<VerifyEmail />);

    // Every key here lives in auth.json, so it resolves under `auth.`. The
    // component called t('verify.…'), which i18next renders as the key itself,
    // so the whole screen was literal key text.
    for (const bare of [
      'verify.enterCodePrompt',
      'verify.codePlaceholder',
      'verify.verifyCode',
      'verify.resend',
      'common.email',
    ]) {
      expect(screen.queryByText(bare)).not.toBeInTheDocument();
      expect(screen.queryByPlaceholderText(bare)).not.toBeInTheDocument();
    }
  });

  it('sends the typed code as the token the contract carries', () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '123456' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    // onion's VerifyEmail hands req.GetToken() straight to VerifyEmailCode, so
    // the single `token` field is the code. This is the supported path.
    expect(verifyMutate).toHaveBeenCalledWith({ token: '123456' }, expect.any(Object));
  });

  it('does not send a code that is not six digits', () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '123' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    expect(verifyMutate).not.toHaveBeenCalled();
    expect(screen.getByText('auth.verify.codeInvalid')).toBeInTheDocument();
  });

  it('verifies a link token on arrival when one is present', async () => {
    search.value = '?token=opaque-link-token';

    render(<VerifyEmail />);

    await waitFor(() =>
      expect(verifyMutate).toHaveBeenCalledWith({ token: 'opaque-link-token' }, expect.any(Object))
    );
  });

  it('resends to the address typed into the form', () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.common.email'), {
      target: { value: 'someone@example.com' },
    });
    fireEvent.submit(screen.getByPlaceholderText('auth.common.email').closest('form')!);

    expect(resendMutate).toHaveBeenCalledWith({ email: 'someone@example.com' }, expect.any(Object));
  });

  it('asks for an address in translated copy when the form is submitted empty', () => {
    render(<VerifyEmail />);

    // The input carries `required`, so a click is stopped by native validation
    // and never reaches the handler. Submitting the form directly is how the
    // guard is reachable, which is what it is there for.
    fireEvent.submit(screen.getByPlaceholderText('auth.common.email').closest('form')!);

    expect(resendMutate).not.toHaveBeenCalled();
    expect(notifyMock.error).toHaveBeenCalledWith('auth.verify.emailRequired');
  });
});
