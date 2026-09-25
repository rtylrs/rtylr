// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children, to }: { children: React.ReactNode; to: string }) =>
    React.createElement('a', { href: to }, children),
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

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
import { initTestI18n } from '../../shared/testing/i18n';

const captureRedirectLocation = (search: string): unknown[] => {
  const assigned: unknown[] = [];
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: {
      search,
      set href(value: unknown) {
        assigned.push(value);
      },
    },
  });
  return assigned;
};

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Register Feature Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders register input fields', () => {
    render(<Register />);
    expect(screen.getByPlaceholderText(/jane doe/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/my coffee shop/i)).toBeInTheDocument();
  });

  it('shows terms and privacy links next to signup without marketing consent', () => {
    render(<Register />);
    expect(screen.getByRole('link', { name: 'Terms of Service' })).toHaveAttribute(
      'href',
      'https://rtylr.com/legal/terms/'
    );
    expect(screen.getByRole('link', { name: 'Privacy Policy' })).toHaveAttribute(
      'href',
      'https://rtylr.com/legal/privacy-policy/'
    );
    expect(screen.getByText(/This does not subscribe you to marketing/)).toBeInTheDocument();
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

  it.each(['constructor', 'api'])(
    'falls back to Dash for an invalid redirect app id: %s',
    async (redirect) => {
      const assigned = captureRedirectLocation(`?redirect=${redirect}`);
      render(<Register />);
      fireEvent.change(screen.getByPlaceholderText(/jane doe/i), {
        target: { value: 'Alex Smith' },
      });
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

      fireEvent.click(screen.getByRole('button', { name: /create account/i }));
      await waitFor(() => expect(mockMutate).toHaveBeenCalledOnce());
      const options = mockMutate.mock.calls[0][1] as { onSuccess: () => void };
      options.onSuccess();

      expect(assigned).toEqual(['http://localhost:5175']);
    }
  );
});
