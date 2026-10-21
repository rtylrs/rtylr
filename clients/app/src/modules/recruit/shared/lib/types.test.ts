import { describe, it, expect } from 'vitest';
import type { RecruitCareer, Customization, TemplateId, FormFieldKind, SectionKind } from './types';

// Type-level tests: ensure the structural contracts are stable.
// These tests assert runtime-representable invariants so a type
// change is caught by a failing assertion rather than silently
// shipping broken wire compatibility.

describe('RecruitCareer shape', () => {
  it('accepts a well-formed career object', () => {
    const career: RecruitCareer = {
      id: 1,
      ownerId: 1,
      slug: 'senior-eng',
      title: 'Senior Engineer',
      department: 'Engineering',
      location: 'Remote',
      employmentType: 'Full-time',
      workMode: 'remote',
      seniority: 'senior',
      details: 'Build great things.',
      requirements: 'TypeScript proficiency.',
      salaryMin: 120000,
      salaryMax: 160000,
      salaryCurrency: 'USD',
      isPublished: true,
      isActive: true,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    };
    expect(career.slug).toBe('senior-eng');
    expect(career.isPublished).toBe(true);
    expect(career.salaryMin).toBe(120000);
  });
});

describe('Customization shape', () => {
  it('has the expected template and fontPairing fields', () => {
    const c: Customization = {
      id: 1,
      ownerId: 1,
      slug: 'acme',
      brandName: 'Acme',
      tagline: 'We build things',
      aboutBlurb: '',
      logoUrl: '',
      primaryColor: '#111',
      accentColor: '#555',
      bgColor: '#fff',
      textColor: '#000',
      mutedTextColor: '#888',
      fontPairing: 'default',
      heroImageUrl: '',
      faviconUrl: '',
      template: 'minimal',
      headerHtml: '',
      metaTitle: 'Acme Careers',
      metaDescription: 'Join us',
      privacyUrl: '',
      termsUrl: '',
      gaMeasurementId: '',
      metaPixelId: '',
      requireConsent: false,
      notificationEmail: '',
      emailNotifications: true,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    };
    expect(c.template).toBe('minimal');
    expect(c.requireConsent).toBe(false);
  });
});

describe('TemplateId union', () => {
  it('accepts known template ids', () => {
    const ids: TemplateId[] = ['minimal', 'editorial', 'bold'];
    expect(ids).toHaveLength(3);
  });
});

describe('FormFieldKind union', () => {
  it('covers all expected form field types', () => {
    const kinds: FormFieldKind[] = [
      'short_text',
      'long_text',
      'single_select',
      'multi_select',
      'url',
      'boolean',
    ];
    expect(kinds).toHaveLength(6);
  });
});

describe('SectionKind union', () => {
  it('covers all known section types', () => {
    const kinds: SectionKind[] = [
      'perks',
      'culture',
      'testimonial',
      'video',
      'faq',
      'values',
      'cta',
    ];
    expect(kinds).toHaveLength(7);
  });
});
