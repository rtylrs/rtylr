import React from 'react';
import fs from 'node:fs';
import path from 'node:path';
import i18next from 'i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children, ...props }: { to: string; children: React.ReactNode }) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));

const qrValues = vi.hoisted(() => [] as string[]);
vi.mock('react-qrcode-logo', () => ({
  QRCode: ({ value }: { value: string }) => {
    qrValues.push(value);
    return <canvas data-testid="qr" />;
  },
}));

const getConfig = vi.hoisted(() => vi.fn());
vi.mock('../../../../shared/lib/clients', () => ({ configClient: { getConfig } }));

import Poster from './Poster';
import { useUserStore } from '../../../../shared/stores/userStore';
import { initTestI18n } from '../../shared/testing/i18n';
import { RENNILI_URL } from '@voxire/config';

function renderPoster() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <Poster />
    </QueryClientProvider>
  );
}

beforeAll(async () => {
  await initTestI18n('en');
  // The test backend serves one language; the poster always prints Arabic too.
  const arabic = JSON.parse(
    fs.readFileSync(path.resolve(__dirname, '../../locales/ar/pager.json'), 'utf8')
  );
  i18next.addResourceBundle('ar', 'pager', { pager: arabic }, true, true);
});

beforeEach(() => {
  qrValues.length = 0;
  getConfig.mockReset().mockResolvedValue({ business: { slug: 'acme', name: 'Acme Coffee' } });
  useUserStore.setState({ user: { businessSlug: 'acme', businessName: 'Acme Coffee' } });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('pager counter poster', () => {
  it('points the QR at the venue scan page and names the venue', async () => {
    renderPoster();

    const joinUrl = `${RENNILI_URL}/acme`;
    expect(qrValues).toContain(joinUrl);
    const poster = screen.getAllByRole('article')[0];
    expect(within(poster).getByText('Acme Coffee')).toBeInTheDocument();
    expect(within(poster).getByText(joinUrl)).toBeInTheDocument();
  });

  it('carries the three steps in English and Arabic', () => {
    renderPoster();

    const poster = screen.getAllByRole('article')[0];
    const english = poster.querySelector('[lang="en"]') as HTMLElement;
    const arabic = poster.querySelector('[lang="ar"]') as HTMLElement;
    expect(within(english).getAllByRole('listitem')).toHaveLength(3);
    expect(within(english).getByText(/scan/i)).toBeInTheDocument();
    // The customer shows a code; nobody types a receipt number any more.
    expect(
      within(english).getByText('Show the code on your phone to the cashier.')
    ).toBeInTheDocument();
    expect(poster.textContent).not.toMatch(/receipt/i);
    expect(arabic).toHaveAttribute('dir', 'rtl');
    expect(within(arabic).getAllByRole('listitem')).toHaveLength(3);
    expect(arabic.textContent).toMatch(/[؀-ۿ]/);
  });

  it('prints from the button', () => {
    // The test DOM has no window.print of its own.
    const print = vi.fn();
    vi.stubGlobal('print', print);
    renderPoster();

    fireEvent.click(screen.getByRole('button', { name: 'Print poster' }));

    expect(print).toHaveBeenCalledOnce();
  });

  it('links back to the board', () => {
    renderPoster();

    expect(screen.getByRole('link', { name: 'Back to the pager' })).toHaveAttribute(
      'href',
      '/board'
    );
  });

  it('says it is preparing while the business link is still loading', () => {
    useUserStore.setState({ user: {} });
    getConfig.mockReturnValue(new Promise(() => undefined));
    renderPoster();

    expect(screen.getByRole('status')).toHaveTextContent('Preparing the poster...');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Print poster' })).toBeDisabled();
  });

  it('explains instead of printing a QR to nowhere when the business has no slug', async () => {
    useUserStore.setState({ user: { businessName: 'Acme Coffee' } });
    getConfig.mockResolvedValue({ business: { name: 'Acme Coffee' } });
    renderPoster();

    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(screen.queryByRole('article')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Print poster' })).toBeDisabled();
  });
});
