import { test, expect } from './fixtures';

/**
 * End-to-end: a candidate applying to a role on a hosted careers page.
 *
 * This is the recruit product's primary conversion surface. If this
 * flow breaks silently, the business loses applications without
 * knowing - so we lock it down in an integration test that drives a
 * real browser through the real app, stubbing only the API layer.
 *
 * API calls are intercepted via `page.route()` so the test doesn't
 * depend on `localhost:9999` - CI stays deterministic, no backend
 * required, no prod data pollution.
 */

const RECRUIT_URL = process.env.RECRUIT_URL ?? 'http://localhost:5184';

// Fake data matching the Go payload shapes in `careers` module.
const FAKE_CUSTOMIZATION = {
  id: 1,
  owner_id: 1,
  slug: 'acme',
  brand_name: 'Acme',
  tagline: 'We are hiring.',
  about_blurb: 'Small team, big surface.',
  logo_url: '',
  primary_color: '#111111',
  accent_color: '',
  bg_color: '',
  text_color: '',
  muted_text_color: '',
  font_pairing: 'default',
  hero_image_url: '',
  favicon_url: '',
  template: 'minimal',
  header_html: '',
  meta_title: '',
  meta_description: '',
  privacy_url: '',
  terms_url: '',
  ga_measurement_id: '',
  meta_pixel_id: '',
  require_consent: false,
  notification_email: '',
  email_notifications: true,
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
};

const FAKE_CAREER = {
  id: 42,
  owner_id: 1,
  slug: 'backend-engineer',
  title: 'Backend Engineer',
  department: 'Engineering',
  location: 'Remote',
  employment_type: 'Full-time',
  work_mode: 'Remote',
  seniority: 'Senior',
  details: 'Build great things with us.',
  requirements: '',
  salary_min: 0,
  salary_max: 0,
  salary_currency: 'USD',
  is_published: true,
  is_active: true,
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
};

/**
 * stubRecruitAPI installs route handlers for every backend endpoint
 * the recruit public pages call. Keep this in one place so specs don't
 * drift - every new endpoint used by the UI should get a stub here.
 */
async function stubRecruitAPI(page: import('@playwright/test').Page) {
  // Shape-match the GET /recruit/public/career?slug=...&career_slug=... response.
  await page.route('**/recruit/public/career*', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        career: FAKE_CAREER,
        customization: FAKE_CUSTOMIZATION,
        fields: [],
      }),
    });
  });

  // Shape-match the GET /recruit/public/careers?slug=... response.
  await page.route('**/recruit/public/careers*', async (route) => {
    const requestUrl = new URL(route.request().url());
    const requestedSlug =
      requestUrl.searchParams.get('slug') ?? requestUrl.searchParams.get('business_slug');
    if (requestedSlug === 'does-not-exist') {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Careers page not found' }),
      });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        careers: [FAKE_CAREER],
        customization: FAKE_CUSTOMIZATION,
        sections: [],
      }),
    });
  });

  // Capture the submission so we can assert on its body.
  await page.route('**/recruit/public/apply', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ id: 1, message: 'Application received' }),
    });
  });

  // Suppress analytics / OG image calls - not the subject of this spec.
  await page.route('**/recruit/public/og*', (route) => route.fulfill({ status: 204 }));
}

test.describe('Recruit apply flow', () => {
  test.beforeEach(async ({ page }) => {
    await stubRecruitAPI(page);
  });

  test('careers index page renders the role', async ({ page }) => {
    await page.goto(`${RECRUIT_URL}/acme`);
    await expect(page.getByRole('heading', { name: /Open roles at Acme/i })).toBeVisible();
    await expect(page.getByText('Backend Engineer')).toBeVisible();
  });

  test('candidate can fill the apply form and see the success state', async ({ page }) => {
    // Capture the request body so we can assert what the backend
    // would receive. page.route() runs before fulfill, so we can
    // observe the submit payload without a separate listener.
    let submittedBody: Record<string, unknown> | null = null;
    await page.route('**/recruit/public/apply', async (route) => {
      const body = route.request().postDataJSON();
      submittedBody = body;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ id: 1, message: 'Application received' }),
      });
    });

    await page.goto(`${RECRUIT_URL}/acme/backend-engineer/apply`);

    await expect(page.getByRole('heading', { name: /Apply for Backend Engineer/i })).toBeVisible();

    await page.getByRole('textbox', { name: /full name/i }).fill('Jamie Doe');
    await page.getByRole('textbox', { name: /^email/i }).fill('jamie@example.com');

    // The public app shows the consent dialog on a fresh browser context. This
    // flow does not require optional analytics, so decline it before submitting
    // and keep the test focused on the application operation.
    const cookieConsent = page.getByRole('dialog', { name: 'Cookie consent' });
    if (await cookieConsent.isVisible()) {
      await cookieConsent.getByRole('button', { name: 'Decline' }).click();
    }

    await page.getByRole('button', { name: /submit application/i }).click();

    // Inline success block replaces the form.
    await expect(page.getByRole('heading', { name: /You're in\./i })).toBeVisible({
      timeout: 5_000,
    });
    await expect(page.getByRole('button', { name: /submit application/i })).toHaveCount(0);

    // Backend received the canonicalised payload.
    expect(submittedBody).toMatchObject({
      slug: 'acme',
      career_slug: 'backend-engineer',
      candidate_name: 'Jamie Doe',
      email: 'jamie@example.com',
    });
  });

  test('404 renders when the slug is unknown', async ({ page }) => {
    await page.goto(`${RECRUIT_URL}/does-not-exist`);
    await expect(page.getByText(/404 · Not found/i)).toBeVisible();
  });
});
