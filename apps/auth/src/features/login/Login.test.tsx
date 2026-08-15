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

vi.mock('../../shared/hooks/useFeatureGate', () => ({
  useFeatureGate: () => false,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
  LicenseBanner: () => null,
  Button: ({ children, ...props }: any) => React.createElement('button', props, children),
}));

vi.mock('../../shared/components/AuthShell', () => ({
  default: ({ children }: { children: React.ReactNode }) =>
    React.createElement('div', null, children),
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
