/**
 * Customization utility helpers for the Recruit app.
 */

import type { Customization, TemplateId } from './types';

export const DEFAULT_PRIMARY_COLOR = '#6366f1';
export const DEFAULT_ACCENT_COLOR = '#f59e0b';

const VALID_TEMPLATES: TemplateId[] = ['minimal', 'editorial', 'bold'];

/** Return the primary brand color, falling back to the default when empty. */
export function primaryColor(c: Customization): string {
  return c.primaryColor || DEFAULT_PRIMARY_COLOR;
}

/** Return the accent brand color, falling back to the default when empty. */
export function accentColor(c: Customization): string {
  return c.accentColor || DEFAULT_ACCENT_COLOR;
}

/** Return true when the customization has a logo URL set. */
export function hasLogo(c: Customization): boolean {
  return Boolean(c.logoUrl);
}

/** Return true when the customization has a hero image URL set. */
export function hasHeroImage(c: Customization): boolean {
  return Boolean(c.heroImageUrl);
}

/** Return the brand name for display, falling back to the slug. */
export function displayBrandName(c: Customization): string {
  return c.brandName || c.slug;
}

/** Return the careers page meta title, constructing one when not set. */
export function careersMetaTitle(c: Customization): string {
  if (c.metaTitle) return c.metaTitle;
  return `${c.brandName} Careers`;
}

/** Return the careers page meta description, falling back to the tagline. */
export function careersMetaDescription(c: Customization): string {
  return c.metaDescription || c.tagline || '';
}

/** Return true when the template is a known valid template ID. */
export function isValidTemplate(template: string): template is TemplateId {
  return (VALID_TEMPLATES as string[]).includes(template);
}

/** Return the effective template, defaulting to "minimal" for unknown values. */
export function effectiveTemplate(c: Customization): TemplateId {
  return isValidTemplate(c.template) ? c.template : 'minimal';
}

/** Return true when the font pairing is a custom (non-default) pairing. */
export function hasCustomFontPairing(c: Customization): boolean {
  return Boolean(c.fontPairing) && c.fontPairing !== 'default';
}

/** Build a CSS variable string from brand colors. */
export function buildCssVars(c: Customization): string {
  const vars = [
    `--brand-primary: ${primaryColor(c)}`,
    `--brand-accent: ${accentColor(c)}`,
    `--brand-bg: ${c.bgColor || '#ffffff'}`,
    `--brand-text: ${c.textColor || '#000000'}`,
    `--brand-muted: ${c.mutedTextColor || '#6b7280'}`,
  ];
  return vars.join(';');
}

/** Return true when both privacy and terms URLs are set. */
export function hasLegalLinks(c: Customization): boolean {
  return Boolean(c.privacyUrl) && Boolean(c.termsUrl);
}
