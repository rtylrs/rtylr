// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children, to }: { children: React.ReactNode; to: string }) =>
    React.createElement('a', { href: to }, children),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

const mockMutate = vi.fn();
vi.mock('../../shared/hooks/useAuth', () => ({
  useLogin: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
  Button: ({ children, ...rest }: { children: React.ReactNode }) => (
    <button type="button" {...rest}>
      {children}
    </button>
  ),
}));

vi.mock('../../shared/components/AuthShell', () => ({
  default: ({ children }: { children: React.ReactNode }) =>
    React.createElement('div', null, children),
}));

// vi.hoisted, because `createOneKitClient` is called at module scope by both
// Login.tsx and social.ts, so the factory dereferences this during the import
// that vitest hoists above a plain `const`.
const { mockGetMe } = vi.hoisted(() => ({ mockGetMe: vi.fn() }));
vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({ getMe: mockGetMe, googleAuthorize: vi.fn() }),
}));

import Login from './Login';

describe('Login Feature Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders email and password inputs', () => {
    render(<Login />);
    expect(screen.getByLabelText(/auth.common.email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth.common.password/i)).toBeInTheDocument();
  });

  it('submits email and password on form submission', async () => {
    render(<Login />);
    const emailInput = screen.getByLabelText(/auth.common.email/i);
    const passwordInput = screen.getByLabelText(/auth.common.password/i);

    fireEvent.change(emailInput, { target: { value: 'user@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'Password123' } });

    const submitBtn = screen.getByRole('button', { name: /auth.login.submit/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        { email: 'user@example.com', password: 'Password123' },
        expect.any(Object)
      );
    });
  });
});

describe('Login redirect target and the MFA gate', () => {
  const effectiveAccessFor = (appId: string) => ({
    businessId: '1',
    dashPermissions: [],
    subscribedApps: [appId],
    licenses: [],
    apps: [
      {
        appId,
        licensed: true,
        hasAccess: true,
        permissions: [],
        permissionScopes: [],
        sources: ['explicit'],
      },
    ],
    isOwner: false,
    isAdmin: false,
    authorizationVersion: 1,
  });

  // window.location.href is assigned directly by the component. Replace the
  // whole object so the assignment is observable and happy-dom does not try to
  // navigate, and set `search` before render because redirectParam is read once
  // in a useMemo.
  const setLocation = (search: string): string[] => {
    const assigned: string[] = [];
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        search,
        get href(): string {
          return `http://localhost:5176/${search}`;
        },
        set href(value: string) {
          assigned.push(value);
        },
      },
    });
    return assigned;
  };

  // The component redirects from the login mutation's onSuccess callback, so
  // driving that callback is the same path a real successful sign-in takes.
  const signIn = async (): Promise<void> => {
    const submit = screen.getByRole('button', { name: /auth.login.submit/i });
    fireEvent.change(screen.getByLabelText(/auth.common.email/i), {
      target: { value: 'user@example.com' },
    });
    fireEvent.change(screen.getByLabelText(/auth.common.password/i), {
      target: { value: 'Password123' },
    });
    fireEvent.click(submit);
    const options = mockMutate.mock.calls[0][1] as { onSuccess: () => void };
    await options.onSuccess();
  };

  const OWES_MFA = {
    provisioningStatus: { mfaRequired: true, mfaEnrolled: false, mfaVerified: false },
    effectiveAccess: effectiveAccessFor('crm'),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('sends a user who owes MFA to Dash even when a redirect names another app', async () => {
    const assigned = setLocation('?redirect=crm');
    mockGetMe.mockResolvedValue(OWES_MFA);

    render(<Login />);
    await signIn();

    await waitFor(() => expect(assigned.length).toBeGreaterThan(0));
    expect(assigned[0]).toBe('http://localhost:5175');
    expect(assigned[0]).not.toContain('5181');
  });

  it('reads MFA state before choosing a destination, not after', async () => {
    const assigned = setLocation('?redirect=crm');
    mockGetMe.mockResolvedValue(OWES_MFA);

    render(<Login />);
    await signIn();

    // The bypass was structural: the redirect param was honoured and the
    // function returned before getMe was ever called.
    await waitFor(() => expect(mockGetMe).toHaveBeenCalledTimes(1));
    expect(assigned).toHaveLength(1);
  });

  it('honours the redirect once MFA is satisfied', async () => {
    const assigned = setLocation('?redirect=crm');
    mockGetMe.mockResolvedValue({
      provisioningStatus: { mfaRequired: true, mfaEnrolled: true, mfaVerified: true },
      effectiveAccess: effectiveAccessFor('crm'),
    });

    render(<Login />);
    await signIn();

    await waitFor(() => expect(assigned.length).toBeGreaterThan(0));
    expect(assigned[0]).toBe('http://localhost:5175/crm');
  });

  it('honours the redirect when MFA is not required at all', async () => {
    const assigned = setLocation('?redirect=flow');
    mockGetMe.mockResolvedValue({
      provisioningStatus: { mfaRequired: false },
      effectiveAccess: effectiveAccessFor('flow'),
    });

    render(<Login />);
    await signIn();

    await waitFor(() => expect(assigned.length).toBeGreaterThan(0));
    expect(assigned[0]).toBe('http://localhost:5175/flow');
  });

  it('fails closed to Dash when the MFA state cannot be read', async () => {
    const assigned = setLocation('?redirect=crm');
    mockGetMe.mockRejectedValue(new Error('network'));

    render(<Login />);
    await signIn();

    await waitFor(() => expect(assigned.length).toBeGreaterThan(0));
    expect(assigned[0]).toBe('http://localhost:5175');
  });

  it('sends a user with an invalid redirect to Dash', async () => {
    const assigned = setLocation('?redirect=https://evil.example.com');
    mockGetMe.mockResolvedValue({
      provisioningStatus: { mfaRequired: false },
      effectiveAccess: effectiveAccessFor('crm'),
    });

    render(<Login />);
    await signIn();

    await waitFor(() => expect(assigned.length).toBeGreaterThan(0));
    expect(assigned[0]).toBe('http://localhost:5175');
    expect(assigned[0]).not.toContain('evil.example.com');
  });
});
