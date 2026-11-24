import { isGone } from './apiError';
import type { PublicPagerClient } from './pagerClients';
import { isOpenStatus } from './ticketPoller';

const KEY_PREFIX = 'pager:scan:';
const TOKEN_PREFIX = 'pager:token:';

/**
 * The ticket this browser opened at a venue, so a reload, or a second scan of
 * the QR that the camera opens in a new tab, keeps the same ticket and code
 * while the order is still active instead of putting a second phone on the
 * board. Storage can be missing or refuse writes (private mode, blocked
 * storage); then every load is a new scan, which still works.
 */
export function rememberedScan(slug: string): string | null {
  try {
    return window.localStorage.getItem(KEY_PREFIX + slug);
  } catch {
    return null;
  }
}

export function rememberScan(slug: string, publicToken: string): void {
  try {
    const previous = window.localStorage.getItem(KEY_PREFIX + slug);
    if (previous && previous !== publicToken) {
      window.localStorage.removeItem(TOKEN_PREFIX + previous);
    }
    window.localStorage.setItem(KEY_PREFIX + slug, publicToken);
    window.localStorage.setItem(TOKEN_PREFIX + publicToken, slug);
  } catch {
    // Without storage a reload scans again.
  }
}

export function forgetScan(slug: string): void {
  try {
    const token = window.localStorage.getItem(KEY_PREFIX + slug);
    window.localStorage.removeItem(KEY_PREFIX + slug);
    if (token) window.localStorage.removeItem(TOKEN_PREFIX + token);
  } catch {
    // Nothing stored to forget.
  }
}

export function rememberedSlugForToken(publicToken: string): string | null {
  if (!publicToken) return null;
  try {
    return window.localStorage.getItem(TOKEN_PREFIX + publicToken);
  } catch {
    return null;
  }
}

/**
 * Clears the remembered venue scan when this ticket is finished (collected,
 * expired, or acknowledged), so the next scan immediately starts a fresh order.
 */
export function forgetScanByToken(publicToken: string): void {
  if (!publicToken) return;
  try {
    const slug = window.localStorage.getItem(TOKEN_PREFIX + publicToken);
    if (slug && window.localStorage.getItem(KEY_PREFIX + slug) === publicToken) {
      window.localStorage.removeItem(KEY_PREFIX + slug);
    }
  } catch {
    // Nothing stored to forget.
  }
}

/**
 * The ticket for this phone at the venue: the one this browser already opened
 * while it is still open and unacknowledged, otherwise a new scan. Pass
 * `forceNew` to always open a fresh ticket (e.g. a second order).
 */
export async function openScanTicket(
  client: PublicPagerClient,
  slug: string,
  { remember, forceNew = false }: { remember: boolean; forceNew?: boolean }
): Promise<string> {
  if (forceNew && remember) {
    forgetScan(slug);
  }
  const stored = remember && !forceNew ? rememberedScan(slug) : null;
  if (stored) {
    try {
      const ticket = await client.getPublicTicket({ publicToken: stored });
      // Once the customer taps "I have it" (acknowledgedAt) or staff collect it,
      // the order is done and a new scan should open a fresh ticket.
      if (isOpenStatus(ticket.status) && !ticket.acknowledgedAt) return stored;
    } catch (error) {
      if (!isGone(error)) return stored;
    }
    forgetScan(slug);
  }
  const { publicToken } = await client.scanTicket({ slug });
  if (!publicToken) throw new Error('The pager answered without a ticket token');
  if (remember) rememberScan(slug, publicToken);
  return publicToken;
}
