import { isGone } from './apiError';
import type { PublicPagerClient } from './pagerClients';
import { isOpenStatus } from './ticketPoller';

const KEY_PREFIX = 'pager:scan:';

/**
 * The ticket this browser opened at a venue, so a reload, or a second scan of
 * the QR that the camera opens in a new tab, keeps the same ticket and code
 * instead of putting a second phone on the board. Storage can be missing or
 * refuse writes (private mode, blocked storage); then every load is a new
 * scan, which still works.
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
    window.localStorage.setItem(KEY_PREFIX + slug, publicToken);
  } catch {
    // Without storage a reload scans again.
  }
}

export function forgetScan(slug: string): void {
  try {
    window.localStorage.removeItem(KEY_PREFIX + slug);
  } catch {
    // Nothing stored to forget.
  }
}

/**
 * The ticket for this phone at the venue: the one this browser already opened
 * while it is still open, otherwise a new scan. When the remembered ticket
 * cannot be checked (a weak signal), it is kept: the ticket page retries, and
 * a second ticket would put the same phone on the board twice.
 */
export async function openScanTicket(
  client: PublicPagerClient,
  slug: string,
  { remember }: { remember: boolean }
): Promise<string> {
  const stored = remember ? rememberedScan(slug) : null;
  if (stored) {
    try {
      const ticket = await client.getPublicTicket({ publicToken: stored });
      if (isOpenStatus(ticket.status)) return stored;
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
