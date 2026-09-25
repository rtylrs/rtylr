import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import PreviewRouteUnavailable from './PreviewRouteUnavailable';
import { initTestI18n } from '../testing/i18n';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('PreviewRouteUnavailable', () => {
  it('explains why a deep link is unavailable and keeps a useful escape route', () => {
    render(<PreviewRouteUnavailable />);

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Preview route unavailable' })).toBeInTheDocument();
    expect(screen.getByText(/not enabled in the current environment/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Return to My Issues' })).toHaveAttribute(
      'href',
      '/my-issues'
    );
  });
});
