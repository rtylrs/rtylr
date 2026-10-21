// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { ApiError } from '@voxire/voxie/common/auth/v1/client';

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

/**
 * Reject the mutation the way onion's generated client does: a 400 whose body
 * is the flat `{ message }` object `onekithttp.NewError` serializes. The
 * component reads the wording out of that body, so a bare `new Error()` would
 * not exercise the path that carries the "request a new one" distinction.
 */
const rejectVerificationWith = async (statusCode: number, message: string) => {
  const { calls } = verifyMutate.mock;
  const call = calls[calls.length - 1];
  if (!call) throw new Error('no verification attempt was made');
  await act(async () => {
    call[1].onError(new ApiError(statusCode, JSON.stringify({ message })));
  });
};

describe('VerifyEmail (#1027)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    search.value = '';
  });

  it('asks for the code instead of reporting a failure that has not happened', () => {
    render(<VerifyEmail />);

    // Arriving with no token is the normal path - the user has the email in
    // front of them and has not used the code yet - so it is not an error.
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

  it('sends the typed token whole, id half included', () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '1042.0123456789' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    // onion's VerifyEmail hands req.GetToken() straight to VerifyEmailCode, so
    // the single `token` field carries the whole token. Dropping the id half
    // made every real code unverifiable, because the server needs it to
    // recompute the account-bound digest.
    expect(verifyMutate).toHaveBeenCalledWith({ token: '1042.0123456789' }, expect.any(Object));
  });

  it('keeps a pasted token whole when the mail client wrapped it in whitespace', () => {
    render(<VerifyEmail />);

    // U+00A0 is what a rendered email wraps a copied token in, and it is not
    // the space the old `.trim()` was guaranteed to be handed.
    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '  1042.0123456789\n' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    expect(verifyMutate).toHaveBeenCalledWith({ token: '1042.0123456789' }, expect.any(Object));
  });

  it('reads the token out of a pasted verification link', () => {
    render(<VerifyEmail />);

    // The email leads with the link, so the link is what gets copied.
    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: 'https://auth.rtylr.com/verify?token=1042.0123456789' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    expect(verifyMutate).toHaveBeenCalledWith({ token: '1042.0123456789' }, expect.any(Object));
  });

  it('does not cut a real token off at six characters', () => {
    render(<VerifyEmail />);

    const input = screen.getByPlaceholderText('auth.verify.codePlaceholder');
    // A 6-digit field truncated "1042.0123456789" to "1042.01", which the
    // server rejects as malformed. The limit has to fit the longest token.
    expect(input).toHaveAttribute('maxLength', String(30));

    fireEvent.change(input, { target: { value: '1042.0123456789' } });
    expect(input).toHaveValue('1042.0123456789');
  });

  it('keeps the separator reachable from the keyboard', () => {
    render(<VerifyEmail />);

    const input = screen.getByPlaceholderText('auth.verify.codePlaceholder');
    // "numeric" offers a keypad with no "." key on iOS, and a number input
    // normalises the value away, so neither can produce the token's separator.
    expect(input).toHaveAttribute('type', 'text');
    expect(input).toHaveAttribute('inputMode', 'text');
  });

  it('forwards a legacy six-digit code so the server can call it too old', () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '123456' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    // onion refuses this before it reaches the database or charges the attempt
    // budget, and it owns the "request a new one" wording. Refusing it here
    // would answer with the generic "invalid" copy and send the user round in
    // circles, which is the one outcome the server change was made to stop.
    expect(verifyMutate).toHaveBeenCalledWith({ token: '123456' }, expect.any(Object));
  });

  it('shows the server\'s own "too old" wording rather than the generic one', async () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '123456' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    await rejectVerificationWith(
      400,
      'This verification code is too old. Please request a new one.'
    );

    expect(
      screen.getByText('This verification code is too old. Please request a new one.')
    ).toBeInTheDocument();
    expect(screen.queryByText('auth.verify.codeInvalid')).not.toBeInTheDocument();
    // The recovery the message asks for has to be on the page, not a step away.
    expect(screen.getByText('auth.verify.resend')).toBeInTheDocument();
  });

  it('shows the server message for a code that is merely wrong', async () => {
    render(<VerifyEmail />);

    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '1042.0123456789' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    await rejectVerificationWith(400, 'Invalid or expired email verification code');

    expect(screen.getByText('Invalid or expired email verification code')).toBeInTheDocument();
  });

  it('does not send a value the server would reject as malformed', () => {
    render(<VerifyEmail />);

    // No separator: the id half is not optional, and onion charges no attempt
    // for a shape it cannot parse, so this never needed a round trip.
    fireEvent.change(screen.getByPlaceholderText('auth.verify.codePlaceholder'), {
      target: { value: '123' },
    });
    fireEvent.click(screen.getByText('auth.verify.verifyCode'));

    expect(verifyMutate).not.toHaveBeenCalled();
    expect(screen.getByText('auth.verify.codeInvalid')).toBeInTheDocument();
  });

  it('offers the code form after a link token failed, so the user is not stranded', async () => {
    search.value = '?token=1042.0123456789';

    render(<VerifyEmail />);

    await waitFor(() => expect(verifyMutate).toHaveBeenCalledTimes(1));
    await rejectVerificationWith(400, 'Invalid or expired email verification code');

    // A link that fails used to hide the only way forward, on a page whose own
    // copy says to ask for a new code.
    expect(screen.getByPlaceholderText('auth.verify.codePlaceholder')).toBeInTheDocument();
  });

  it('verifies a link token on arrival when one is present', async () => {
    search.value = '?token=opaque-link-token';

    render(<VerifyEmail />);

    await waitFor(() =>
      expect(verifyMutate).toHaveBeenCalledWith({ token: 'opaque-link-token' }, expect.any(Object))
    );
  });

  it('sends the link token once, not once per render', async () => {
    search.value = '?token=1042.0123456789';

    const { rerender } = render(<VerifyEmail />);
    rerender(<VerifyEmail />);

    await waitFor(() => expect(verifyMutate).toHaveBeenCalledTimes(1));
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

  it('names the address the signup link went to', () => {
    search.value = '?email=new%40example.com&sent=1';
    render(<VerifyEmail />);
    expect(screen.getByText('auth.verify.sentTo')).toBeInTheDocument();
  });
});
