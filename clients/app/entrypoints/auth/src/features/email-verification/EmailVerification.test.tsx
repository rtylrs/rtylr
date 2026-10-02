// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { ApiError } from '@voxire/voxie/common/auth/v1/client';

const { mutateAsync, navigateMock, params } = vi.hoisted(() => ({
  mutateAsync: vi.fn(),
  navigateMock: vi.fn(),
  params: { value: {} as Record<string, string | undefined> },
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
  useParams: () => params.value,
}));

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('./useEmailVerification', () => ({
  useEmailVerification: () => ({ mutateAsync }),
}));

vi.mock('../../shared/components/Logo', () => ({ default: () => null }));
vi.mock('@voxire/components', () => ({
  Button: ({ children, ...rest }: { children: React.ReactNode }) => (
    <button type="button" {...rest}>
      {children}
    </button>
  ),
  HelmetContainer: () => null,
}));

import EmailVerification from './EmailVerification';

/**
 * `/verify/$token` is the route a verification link opens. It is a different
 * component from the `/verify` form, so a change to the token format has to
 * reach both: when only the form was widened, a link that arrived here still
 * answered with copy that never mentioned the code had to be replaced.
 */
describe('EmailVerification', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    params.value = { token: '1042.0123456789' };
    mutateAsync.mockResolvedValue({ verified: true, message: 'Email verified' });
  });

  it('spends the id-bearing token the email link carries', async () => {
    render(<EmailVerification />);

    await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith({ token: '1042.0123456789' }));
  });

  it('confirms when the server verified the token', async () => {
    render(<EmailVerification />);

    expect(await screen.findByText('Email verified')).toBeInTheDocument();
  });

  it("shows the server's own wording when the code is too old to retry", async () => {
    mutateAsync.mockRejectedValue(
      new ApiError(
        400,
        JSON.stringify({ message: 'This verification code is too old. Please request a new one.' })
      )
    );

    render(<EmailVerification />);

    // The generic failure copy used to replace this, telling a user to try
    // again when trying again could never succeed.
    expect(
      await screen.findByText('This verification code is too old. Please request a new one.')
    ).toBeInTheDocument();
    expect(screen.queryByText('auth.verify.defaultFailure')).not.toBeInTheDocument();
  });

  it('falls back to translated copy when the server sends no message', async () => {
    mutateAsync.mockRejectedValue(new Error(''));

    render(<EmailVerification />);

    expect(await screen.findByText('auth.verify.defaultFailure')).toBeInTheDocument();
  });

  it('sends the token once, not once per render', async () => {
    const { rerender } = render(<EmailVerification />);
    rerender(<EmailVerification />);

    await waitFor(() => expect(mutateAsync).toHaveBeenCalledTimes(1));
  });

  it('does not send anything when the link carries no token', async () => {
    params.value = {};

    render(<EmailVerification />);

    expect(await screen.findByText('auth.verify.tokenMissing')).toBeInTheDocument();
    expect(mutateAsync).not.toHaveBeenCalled();
  });
});
