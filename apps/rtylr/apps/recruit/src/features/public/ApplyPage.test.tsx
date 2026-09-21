import React from 'react';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import ApplyPage from './ApplyPage';
import { uploadResume } from '../../shared/hooks/queries/usePublic';
import { initTestI18n } from '../../shared/testing/i18n';
import type { Customization, RecruitCareer } from '../../shared/lib/types';

/**
 * Integration test for the candidate apply flow - the most important
 * conversion surface in the recruit product. Covers:
 *
 *   - render the form using data returned by the public-career query
 *   - fill required fields + submit
 *   - see the inline success state (no native alert)
 *
 * We mock the hooks module rather than raw axios because the hooks
 * hide a lot of request-shape plumbing we don't want to recreate in
 * tests. Mocking at the hook boundary keeps the test coupled to the
 * component's actual contract, not to URL paths.
 */

// Stable fakes so multiple tests don't share state unless they mean to.
const fakeCareer: RecruitCareer = {
  id: 7,
  ownerId: 1,
  slug: 'backend-engineer',
  title: 'Backend Engineer',
  department: 'Engineering',
  location: 'Remote',
  employmentType: 'Full-time',
  workMode: 'Remote',
  seniority: 'Senior',
  details: 'Build great things.',
  requirements: '',
  salaryMin: 0,
  salaryMax: 0,
  salaryCurrency: 'USD',
  isPublished: true,
  isActive: true,
  createdAt: '2026-04-01T00:00:00Z',
  updatedAt: '2026-04-01T00:00:00Z',
};

const fakeCustomization: Customization = {
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
  requireConsent: false,
  notificationEmail: '',
  emailNotifications: true,
  createdAt: '',
  updatedAt: '',
};

// The submit mutate we'll spy on, declared at module scope so the mock
// factory below can reference it and each test can assert calls.
const submitMutate = vi.fn();

vi.mock('../../shared/hooks/queries/usePublic', () => ({
  usePublicCareer: () => ({
    data: { career: fakeCareer, customization: fakeCustomization, fields: [] },
    isLoading: false,
    error: null,
  }),
  useSubmitApplication: () => ({
    mutate: submitMutate,
    isPending: false,
  }),
  uploadResume: vi.fn(),
}));

// Sonner would try to render into a Toaster portal we haven't mounted;
// stub so the test focuses on the form's own output.
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

async function renderAt(path: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const history = createMemoryHistory({ initialEntries: [path] });
  const rootRoute = createRootRoute();
  const applyRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/$slug/$careerSlug/apply',
    component: ApplyPage,
  });
  const router = createRouter({
    routeTree: rootRoute.addChildren([applyRoute]),
    history,
  });
  await router.load();
  let result: ReturnType<typeof render>;
  await act(async () => {
    result = render(
      <QueryClientProvider client={qc}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );
  });
  return result!;
}

describe('ApplyPage integration', () => {
  // Real locale files, so an assertion written in English also proves the key
  // resolves. See the helper for why an identity `t` mock would not.
  beforeAll(async () => {
    await initTestI18n('en');
  });

  it('renders the role title and brand name in the header', async () => {
    await renderAt('/acme/backend-engineer/apply');
    expect(screen.getByText('Apply for Backend Engineer')).toBeInTheDocument();
    // Header shows the business brand, not "rtylr".
    expect(screen.getAllByText(/Acme/).length).toBeGreaterThan(0);
  });

  it('submits the form with the entered name + email and shows inline success', async () => {
    submitMutate.mockImplementation((_body, opts: { onSuccess?: () => void }) => {
      // Simulate the mutation resolving successfully so the component
      // transitions to its `submitted` branch.
      opts.onSuccess?.();
    });

    await renderAt('/acme/backend-engineer/apply');

    const nameInput = screen.getByRole('textbox', { name: /full name/i });
    const emailInput = screen.getByRole('textbox', { name: /^email/i });
    fireEvent.change(nameInput, { target: { value: 'Jamie Doe' } });
    fireEvent.change(emailInput, { target: { value: 'jamie@example.com' } });

    fireEvent.click(screen.getByRole('button', { name: /submit application/i }));

    await waitFor(() => {
      expect(submitMutate).toHaveBeenCalledTimes(1);
    });
    const body = submitMutate.mock.calls[0][0];
    expect(body).toMatchObject({
      slug: 'acme',
      careerSlug: 'backend-engineer',
      candidateName: 'Jamie Doe',
      email: 'jamie@example.com',
    });

    // After onSuccess fires, the component swaps to the success state.
    await waitFor(() => {
      expect(screen.getByText(/you're in\./i)).toBeInTheDocument();
    });
    // Form is gone - no more submit button on screen.
    expect(screen.queryByRole('button', { name: /submit application/i })).toBeNull();
  });

  it('does not submit when required fields are empty (browser validation)', async () => {
    submitMutate.mockReset();
    await renderAt('/acme/backend-engineer/apply');
    fireEvent.click(screen.getByRole('button', { name: /submit application/i }));
    // Required inputs block form submission at the DOM level - the
    // mutate shouldn't fire.
    expect(submitMutate).not.toHaveBeenCalled();
  });

  it('drops a resume upload that resolves after the file was removed', async () => {
    submitMutate.mockReset();
    const uploadResumeMock = vi.mocked(uploadResume);
    let resolveUpload: (key: string) => void = () => {};
    uploadResumeMock.mockReset();
    uploadResumeMock.mockImplementation(
      () =>
        new Promise<string>((resolve) => {
          resolveUpload = resolve;
        })
    );

    await renderAt('/acme/backend-engineer/apply');

    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(fileInput, {
      target: { files: [new File(['%PDF-1.4'], 'resume.pdf', { type: 'application/pdf' })] },
    });

    // The card shows the file while the PUT is still in flight.
    expect(screen.getByText('resume.pdf')).toBeInTheDocument();
    expect(screen.getByText('Uploading…')).toBeInTheDocument();

    // Remove it before the upload resolves.
    fireEvent.click(screen.getByTitle('Remove'));
    expect(screen.queryByText('resume.pdf')).toBeNull();

    await act(async () => {
      resolveUpload('resumes/removed.pdf');
    });

    // The stale key must not restore the card or hide the picker.
    expect(screen.queryByText('resume.pdf')).toBeNull();
    expect(screen.queryByText('Uploaded')).toBeNull();
    expect(screen.getByText('Upload your resume')).toBeInTheDocument();

    // ...and the removed resume must not be submitted with the application.
    fireEvent.change(screen.getByRole('textbox', { name: /full name/i }), {
      target: { value: 'Jamie Doe' },
    });
    fireEvent.change(screen.getByRole('textbox', { name: /^email/i }), {
      target: { value: 'jamie@example.com' },
    });
    fireEvent.click(screen.getByRole('button', { name: /submit application/i }));

    await waitFor(() => {
      expect(submitMutate).toHaveBeenCalledTimes(1);
    });
    expect(submitMutate.mock.calls[0][0].resumeUrl).toBeUndefined();
  });
});
