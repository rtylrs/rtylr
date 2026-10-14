import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, beforeAll } from 'vitest';

import Footer from './PublicFooter';
import { initTestI18n } from '../../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('PublicFooter component', () => {
  it('renders footer sections and external links', () => {
    render(<Footer />);

    expect(screen.getByText('Company')).toBeInTheDocument();
    expect(screen.getByText('Quick Links')).toBeInTheDocument();
    expect(screen.getByText('Legal')).toBeInTheDocument();

    expect(screen.getByRole('link', { name: 'About Us' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Contact Us' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Careers' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'FAQs' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Pricing' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Privacy Policy' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Terms and Conditions' })).toBeInTheDocument();
  });

  it('renders contact email and Instagram links', () => {
    render(<Footer />);

    const emailLink = screen.getByRole('link', { name: 'info@rtylr.com' });
    expect(emailLink).toHaveAttribute('href', 'mailto:info@rtylr.com');

    const instaLink = screen.getByRole('link', { name: /instagram/i });
    expect(instaLink).toHaveAttribute('href', 'https://www.instagram.com/voxire.tech');
  });

  it('renders copyright and current year', () => {
    render(<Footer />);
    const year = new Date().getFullYear();
    expect(screen.getByText(new RegExp(String(year)))).toBeInTheDocument();
  });
});
