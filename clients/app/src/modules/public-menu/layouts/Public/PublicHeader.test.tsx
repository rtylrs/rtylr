import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const navigateMock = vi.hoisted(() => vi.fn());
const selectLanguageMock = vi.hoisted(() => vi.fn());
const cookieStore = vi.hoisted(() => ({
  token: null as string | null,
  language: 'en' as string | null,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
  useRouterState: () => '/',
  Link: ({ children, to, style }: any) => (
    <a href={to} style={style}>
      {children}
    </a>
  ),
}));

vi.mock('@voxire/utils', () => ({
  useWindowDimensions: () => ({ width: 1200, height: 800 }),
}));

vi.mock('@voxire/components', () => ({
  Button: ({ children, onClick, ...props }: any) => (
    <button onClick={onClick} {...props}>
      {children}
    </button>
  ),
}));

vi.mock('../../../../shared/services/i18next', () => ({
  selectLanguage: selectLanguageMock,
}));

vi.mock('../../shared/lib/cookies', () => ({
  getCookie: (name: string) => {
    if (name === 'token') return cookieStore.token;
    if (name === 'language') return cookieStore.language;
    return null;
  },
}));

import Header from './PublicHeader';

describe('PublicHeader component', () => {
  beforeEach(() => {
    navigateMock.mockClear();
    selectLanguageMock.mockClear();
    cookieStore.token = null;
    cookieStore.language = 'en';
  });

  it('renders logo with click handler navigating to /', () => {
    render(<Header />);

    const logo = screen.getByAltText('rtylr');
    expect(logo).toBeInTheDocument();

    fireEvent.click(logo.parentElement!);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/' });
  });

  it('renders Demo link to POS menu', () => {
    render(<Header />);
    const demoLink = screen.getByRole('link', { name: 'Demo' });
    expect(demoLink).toBeInTheDocument();
    expect(demoLink).toHaveAttribute('href', '/b/69/menu');
  });

  it('renders landing page navigation links', () => {
    render(<Header />);
    expect(screen.getByRole('link', { name: 'Pricing' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Restaurants' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Use Cases' })).toBeInTheDocument();
  });

  it('toggles Company dropdown and handles external links', () => {
    const openSpy = vi.spyOn(window, 'open').mockImplementation(() => null);

    render(<Header />);
    const companyToggle = screen.getByText('Company');
    fireEvent.click(companyToggle);

    expect(screen.getByText('About Us')).toBeInTheDocument();
    expect(screen.getByText('Contact Us')).toBeInTheDocument();
    expect(screen.getByText('FAQs')).toBeInTheDocument();
    expect(screen.getByText('Careers (Jobs)')).toBeInTheDocument();

    fireEvent.click(screen.getByText('About Us'));
    expect(openSpy).toHaveBeenCalledWith(
      expect.stringContaining('/about'),
      '_blank',
      'noopener,noreferrer'
    );

    openSpy.mockRestore();
  });

  it('shows Login button when unauthenticated and hides when token is present', () => {
    const { rerender } = render(<Header />);
    const loginButton = screen.getByText('Login').closest('button')!;
    expect(loginButton).toHaveStyle({ display: 'flex' });
    expect(screen.getByText('Register Now')).toBeInTheDocument();

    cookieStore.token = 'active-jwt-token';
    rerender(<Header />);
    expect(screen.getByText('Login').closest('button')).toHaveStyle({
      display: 'none',
    });
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
  });

  it('toggles language switcher and selects Arabic or French', () => {
    render(<Header />);

    // Click language dropdown trigger (shows current code 'en')
    const langTrigger = screen.getByText('en');
    fireEvent.click(langTrigger);

    const arabicOption = screen.getByText('Arabic');
    expect(arabicOption).toBeInTheDocument();

    fireEvent.click(arabicOption);
    expect(selectLanguageMock).toHaveBeenCalledWith('ar');
  });
});
