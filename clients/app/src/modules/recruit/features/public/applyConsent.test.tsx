import React from 'react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen, act, within } from '@testing-library/react';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import ApplyPage from './ApplyPage';
import { initTestI18n } from '../../shared/testing/i18n';
import type { Customization, RecruitCareer } from '../../shared/lib/types';

/**
 * The consent row is the first use of react-i18next's `<Trans>` in this
 * monorepo, so it gets its own suite.
 *
 * It reads differently for each combination of legal links the business has
 * set, and each combination is a separate whole sentence rather than fragments
 * glued together in JSX, so that ar and fr can put the links where their own
 * grammar wants them. That is four variants, and a wrong `i18nKey` in any of
 * them would render a sentence missing its links with nothing failing, so each
 * is asserted directly.
 *
 * `<link>` is deliberately not the tag name: HTML's `<link>` is a void
 * element, and Trans renders the anchor empty with the text outside it.
 * Measured, before choosing `<privacy>` and `<terms>`:
 *
 *   <link>   paragraph="voir la politique ici"  anchorText=""
 *   <policy> paragraph="voir la politique ici"  anchorText="politique"
 */
const career: RecruitCareer = {
  id: 7,
  ownerId: 1,
  slug: 'backend-engineer',
  title: 'Backend Engineer',
  department: '',
  location: 'Remote',
  employmentType: '',
  workMode: '',
  seniority: '',
  details: '',
  requirements: '',
  salaryMin: 0,
  salaryMax: 0,
  salaryCurrency: 'USD',
  isPublished: true,
  isActive: true,
  createdAt: '2026-04-01T00:00:00Z',
  updatedAt: '2026-04-01T00:00:00Z',
};

const baseCustomization: Customization = {
  id: 1,
  ownerId: 1,
  slug: 'acme',
  brandName: 'Acme',
  tagline: '',
  aboutBlurb: '',
  logoUrl: '',
  primaryColor: '#111111',
  accentColor: '',
  bgColor: '',
  textColor: '',
  mutedTextColor: '',
  fontPairing: 'default',
  heroImageUrl: '',
  faviconUrl: '',
  template: 'minimal',
  headerHtml: '',
  metaTitle: '',
  metaDescription: '',
  privacyUrl: '',
  termsUrl: '',
  gaMeasurementId: '',
  metaPixelId: '',
  requireConsent: true,
  notificationEmail: '',
  emailNotifications: true,
  createdAt: '',
  updatedAt: '',
};

// Each test sets this before rendering. Rebuilt in beforeEach so one test's
// URLs cannot leak into the next.
let customization: Customization = { ...baseCustomization };

vi.mock('../../shared/hooks/queries/usePublic', () => ({
  usePublicCareer: () => ({
    data: { career, customization, fields: [] },
    isLoading: false,
    error: null,
  }),
  useSubmitApplication: () => ({ mutate: vi.fn(), isPending: false }),
  uploadResume: vi.fn(),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

async function renderApply() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const rootRoute = createRootRoute();
  const applyRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/$slug/$careerSlug/apply',
    component: ApplyPage,
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([applyRoute]),
    history: createMemoryHistory({ initialEntries: ['/acme/backend-engineer/apply'] }),
  });
  await router.load();
  await act(async () => {
    render(
      <QueryClientProvider client={qc}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );
  });
}

/**
 * The page has other anchors, the back link among them, so every assertion
 * here is scoped to the consent label rather than the document.
 */
function consentRow(): HTMLElement {
  return screen
    .getByRole('checkbox', { name: /processing of my application data/i })
    .closest('label') as HTMLElement;
}

describe('apply consent row', () => {
  beforeAll(async () => {
    await initTestI18n('en');
  });

  beforeEach(() => {
    customization = { ...baseCustomization };
  });

  it('names the business and offers no links when neither is set', async () => {
    await renderApply();

    expect(consentRow().textContent).toContain(
      "I agree to Acme's processing of my application data."
    );
    expect(within(consentRow()).queryByRole('link')).not.toBeInTheDocument();
  });

  it('links the privacy policy when only that is set', async () => {
    customization.privacyUrl = 'https://acme.test/privacy';
    await renderApply();

    const row = within(consentRow());
    expect(row.getByRole('link', { name: 'privacy policy' })).toHaveAttribute(
      'href',
      'https://acme.test/privacy'
    );
    expect(row.getAllByRole('link')).toHaveLength(1);
  });

  it('links the terms when only those are set', async () => {
    customization.termsUrl = 'https://acme.test/terms';
    await renderApply();

    const row = within(consentRow());
    expect(row.getByRole('link', { name: 'terms' })).toHaveAttribute(
      'href',
      'https://acme.test/terms'
    );
    expect(row.getAllByRole('link')).toHaveLength(1);
  });

  it('links both, each to its own url, when both are set', async () => {
    customization.privacyUrl = 'https://acme.test/privacy';
    customization.termsUrl = 'https://acme.test/terms';
    await renderApply();

    const row = within(consentRow());
    expect(row.getByRole('link', { name: 'privacy policy' })).toHaveAttribute(
      'href',
      'https://acme.test/privacy'
    );
    expect(row.getByRole('link', { name: 'terms' })).toHaveAttribute(
      'href',
      'https://acme.test/terms'
    );
    expect(row.getAllByRole('link')).toHaveLength(2);
    expect(consentRow().textContent).toContain('See the privacy policy and terms.');
  });
});
