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
  useRegisterRestaurant: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
  checkEmailExists: vi.fn().mockResolvedValue({ exists: false }),
}));

vi.mock('../../shared/hooks/useFeatureGate', () => ({
  useFeatureGate: () => false,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
  LicenseBanner: () => null,
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

vi.mock('../../shared/components/PasswordStrengthMeter', () => ({
  default: () => null,
  scorePassword: () => 4,
}));

import Register from './Register';

describe('Register Feature Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders register input fields', () => {
    render(<Register />);
    expect(screen.getByPlaceholderText(/jane doe/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/my coffee shop/i)).toBeInTheDocument();
  });

  it('submits registration form with business name', async () => {
    render(<Register />);
    fireEvent.change(screen.getByPlaceholderText(/jane doe/i), { target: { value: 'Alex Smith' } });
    fireEvent.change(screen.getByPlaceholderText(/my coffee shop/i), {
      target: { value: 'Acme Coffee' },
    });
    fireEvent.change(screen.getByPlaceholderText(/you@company.com/i), {
      target: { value: 'alex@acme.com' },
    });

    const passwordInputs = screen.getAllByPlaceholderText(/characters/i);
    fireEvent.change(passwordInputs[0], { target: { value: 'SecurePass123' } });
    fireEvent.change(screen.getByPlaceholderText(/repeat password/i), {
      target: { value: 'SecurePass123' },
    });

    const submitBtn = screen.getByRole('button', { name: /create account/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          email: 'alex@acme.com',
          name: 'Alex Smith',
          businessName: 'Acme Coffee',
        }),
        expect.any(Object)
      );
    });
  });
});
