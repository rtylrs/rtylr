import React from 'react';
import fs from 'node:fs';
import path from 'node:path';
import i18next from 'i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

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

const alarm = vi.hoisted(() => ({ sync: vi.fn(), release: vi.fn() }));
vi.mock('../../../../shared/alarm/readyAlarm', () => ({ readyAlarm: alarm }));

const getConfig = vi.hoisted(() => vi.fn());
vi.mock('../../../../shared/lib/clients', () => ({ configClient: { getConfig } }));

import Display from './Display';
import { useUserStore } from '../../../../shared/stores/userStore';
import { initTestI18n } from '../../shared/testing/i18n';
import { RENNILI_URL } from '@voxire/config';

function renderDisplay() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <Display />
    </QueryClientProvider>
  );
}

beforeAll(async () => {
  await initTestI18n('en');
  const arabic = JSON.parse(
    fs.readFileSync(path.resolve(__dirname, '../../locales/ar/pager.json'), 'utf8')
  );
  i18next.addResourceBundle('ar', 'pager', { pager: arabic }, true, true);
});

beforeEach(() => {
  qrValues.length = 0;
  alarm.sync.mockReset();
  alarm.release.mockReset();
  getConfig.mockReset().mockResolvedValue({ business: { slug: 'acme', name: 'Acme Coffee' } });
  useUserStore.setState({ user: { businessSlug: 'acme', businessName: 'Acme Coffee' } });
});

describe('pager counter display', () => {
  it('shows the venue QR, its name and the line in English and Arabic', () => {
    renderDisplay();

    expect(qrValues).toContain(`${RENNILI_URL}/acme`);
    expect(screen.getByText('Acme Coffee')).toBeInTheDocument();
    expect(screen.getByText('Scan to get your order alert')).toHaveAttribute('lang', 'en');
    const arabic = screen.getByText('امسح الرمز ليصلك تنبيه طلبك');
    expect(arabic).toHaveAttribute('dir', 'rtl');
    expect(screen.getByRole('link', { name: 'Close the counter display' })).toHaveAttribute(
      'href',
      '/board'
    );
  });

  it('keeps the screen awake while open and lets go when closed', () => {
    const { unmount } = renderDisplay();
    expect(alarm.sync).toHaveBeenCalledWith({ keepAwake: true, ringKey: null });

    unmount();

    expect(alarm.release).toHaveBeenCalledTimes(1);
  });
});
