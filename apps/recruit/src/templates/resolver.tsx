import React from 'react';

import MinimalIndex from './minimal/Index';
import MinimalJob from './minimal/Job';
import EditorialIndex from './editorial/Index';
import EditorialJob from './editorial/Job';
import BoldIndex from './bold/Index';
import BoldJob from './bold/Job';
import { loadFontPairing } from './fonts';
import type { Customization, RecruitCareer, Section } from '../shared/lib/types';

// The template a business picks in their Customization record resolves
// to a component tree here. Each template implements the same two
// entry points: an Index (careers list) and a Job (single role).
// Adding a new template means dropping a folder under templates/ with
// those two components and adding a case here.

export interface IndexProps {
  careers: RecruitCareer[];
  customization: Customization;
  sections: Section[];
}

export interface JobProps {
  career: RecruitCareer;
  customization: Customization;
  slug: string;
}

const INDEX: Record<string, React.FC<IndexProps>> = {
  minimal: MinimalIndex,
  editorial: EditorialIndex,
  bold: BoldIndex,
};

const JOB: Record<string, React.FC<JobProps>> = {
  minimal: MinimalJob,
  editorial: EditorialJob,
  bold: BoldJob,
};

export function resolveIndex(template: string): React.FC<IndexProps> {
  return INDEX[template] ?? MinimalIndex;
}

export function resolveJob(template: string): React.FC<JobProps> {
  return JOB[template] ?? MinimalJob;
}

// TEMPLATE_DEFAULTS is the fallback palette each template expects when
// a Customization field is missing. Keeping these here (rather than in
// each template's CSS) makes it trivial for the admin's "Reset to
// template" button to know what to revert to.
const TEMPLATE_DEFAULTS: Record<string, Record<string, string>> = {
  minimal: {
    accent: '#111111',
    bg: '#fafaf9',
    fg: '#18181b',
    muted: '#6b7280',
    surface: '#ffffff',
    border: '#e7e5e4',
  },
  editorial: {
    accent: '#d94f2f',
    bg: '#ffffff',
    fg: '#111827',
    muted: '#6b7280',
    surface: '#fafaf9',
    border: '#e5e7eb',
  },
  bold: {
    accent: '#1d4ed8',
    bg: '#ffffff',
    fg: '#0f172a',
    muted: '#64748b',
    surface: '#f8fafc',
    border: '#e2e8f0',
  },
};

/**
 * applyBrand pushes the Customization's palette + typography into CSS
 * custom properties read across the template tree. Fields that are
 * empty fall back to the template's default so a half-configured row
 * still renders as a coherent design, not a mash-up.
 *
 * Also mutates the <link rel="icon"> when a favicon_url is set, so
 * candidates see the business's favicon in their browser tab on the
 * public careers site.
 */
export function applyBrand(customization: Customization | undefined): void {
  if (typeof document === 'undefined' || !customization) return;
  const defaults = TEMPLATE_DEFAULTS[customization.template] ?? TEMPLATE_DEFAULTS.minimal;
  const root = document.documentElement;
  const pick = (a: string, b: string): string => (a && a.trim() !== '' ? a : b);

  root.style.setProperty('--recruit-accent', pick(customization.primaryColor, defaults.accent));
  root.style.setProperty(
    '--recruit-accent-2',
    pick(customization.accentColor, pick(customization.primaryColor, defaults.accent))
  );
  root.style.setProperty('--recruit-bg', pick(customization.bgColor, defaults.bg));
  root.style.setProperty('--recruit-fg', pick(customization.textColor, defaults.fg));
  root.style.setProperty('--recruit-muted', pick(customization.mutedTextColor, defaults.muted));
  root.style.setProperty('--recruit-surface', defaults.surface);
  root.style.setProperty('--recruit-border', defaults.border);

  // Typography - loadFontPairing injects the Google Fonts <link> if
  // needed and returns the family names to slot into our CSS vars.
  const pairing = loadFontPairing(customization.fontPairing);
  root.style.setProperty('--recruit-font-body', pairing.body);
  root.style.setProperty('--recruit-font-display', pairing.heading);

  // Favicon swap - this is the small nicety that makes a tenant's
  // careers page feel properly theirs in the browser tab.
  if (customization.faviconUrl) {
    let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    if (!link) {
      link = document.createElement('link');
      link.rel = 'icon';
      document.head.appendChild(link);
    }
    link.href = customization.faviconUrl;
  }
}
