import { isReservedBusinessSlug } from '@voxire/config';
import { errorStatus } from './apiError';

const SLUG_MAX_LENGTH = 60;
const SLUG_ATTEMPTS = 5;
const FALLBACK_SLUG = 'venue';

function slugBase(venueName: string): string {
  return venueName
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .slice(0, SLUG_MAX_LENGTH)
    .replace(/^-+|-+$/g, '');
}

const randomSuffix = (): string => Math.random().toString(36).slice(2, 6).padEnd(4, '0');

/**
 * Slugs to try for a new venue, best first: its name, then its name with a
 * short random ending for when the name is taken. A name with no Latin letters
 * or digits, such as an Arabic one, starts from "venue"; the owner can choose
 * another slug in the general settings.
 */
export function venueSlugCandidates(venueName: string, suffix = randomSuffix): string[] {
  const base = slugBase(venueName);
  const usable = base !== '' && !isReservedBusinessSlug(base);
  const stem = base || FALLBACK_SLUG;
  const candidates = usable ? [base] : [];
  while (candidates.length < SLUG_ATTEMPTS) candidates.push(`${stem}-${suffix()}`);
  return candidates;
}

const SLUG_PATTERN = /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/;

export const normaliseSlug = (raw: string): string => raw.trim().toLowerCase();

/** A slug the business module accepts: lowercase letters, digits, inner dashes, not reserved. */
export function isSlugUsable(slug: string): boolean {
  return slug.length <= SLUG_MAX_LENGTH && SLUG_PATTERN.test(slug) && !isReservedBusinessSlug(slug);
}

export interface ClaimPagerSteps {
  /** Saves the venue's phone number on the business. */
  savePhone: (phone: string) => Promise<unknown>;
  /** Answers 409 when the slug is taken or reserved. */
  setSlug: (slug: string) => Promise<unknown>;
  enablePager: () => Promise<unknown>;
}

/**
 * Turns the caller's business into a pager venue: its phone number, a slug
 * for the poster's link when it has none, and the pager switched on. Each
 * step is safe to repeat, so a failed claim can simply be tried again.
 *
 * A slug the owner confirmed (`chosenSlug`) is the only one tried: posters may
 * already be printed with it, so a taken one fails instead of becoming another.
 */
export async function claimPager(
  steps: ClaimPagerSteps,
  {
    phone,
    venueName,
    slug,
    chosenSlug,
  }: { phone: string; venueName: string; slug: string; chosenSlug?: string }
): Promise<void> {
  await steps.savePhone(phone);
  if (!slug) {
    if (chosenSlug) await steps.setSlug(chosenSlug);
    else await claimSlug(steps.setSlug, venueName);
  }
  await steps.enablePager();
}

async function claimSlug(setSlug: ClaimPagerSteps['setSlug'], venueName: string): Promise<void> {
  let lastError: unknown;
  for (const candidate of venueSlugCandidates(venueName)) {
    try {
      await setSlug(candidate);
      return;
    } catch (error) {
      if (errorStatus(error) !== 409) throw error;
      lastError = error;
    }
  }
  throw lastError;
}

const ARABIC_INDIC_ZERO = 0x0660;
const EASTERN_ARABIC_INDIC_ZERO = 0x06f0;
const PHONE_MIN_DIGITS = 6;
const PHONE_MAX_DIGITS = 15;

/** Arabic-Indic digits become ASCII; everything but digits, a leading +, spaces and dashes goes. */
export function normalisePhone(raw: string): string {
  const ascii = raw.replace(/[٠-٩۰-۹]/g, (char) => {
    const code = char.charCodeAt(0);
    const zero = code >= EASTERN_ARABIC_INDIC_ZERO ? EASTERN_ARABIC_INDIC_ZERO : ARABIC_INDIC_ZERO;
    return String(code - zero);
  });
  const trimmed = ascii.trim();
  const plus = trimmed.startsWith('+') ? '+' : '';
  return (
    plus +
    trimmed
      .replace(/[^\d\s-]/g, '')
      .replace(/\s+/g, ' ')
      .trim()
  );
}

export function isPhoneValid(phone: string): boolean {
  const digits = phone.replace(/\D/g, '').length;
  return digits >= PHONE_MIN_DIGITS && digits <= PHONE_MAX_DIGITS;
}
