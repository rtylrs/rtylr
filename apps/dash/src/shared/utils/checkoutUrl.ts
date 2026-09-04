/**
 * Paddle hosted-checkout allowlist. Exact hosts only: suffix matching would
 * accept lookalikes such as `buy.paddle.com.evil.com`, and bare `paddle.com`
 * serves non-checkout pages, so neither is accepted here.
 */
export const CHECKOUT_HOSTS = ['buy.paddle.com', 'checkout.paddle.com'] as const;

const MAX_URL_LENGTH = 2048;

/** True only for https URLs on exactly one of the Paddle checkout hosts. */
export function isAllowedCheckoutUrl(raw: unknown): raw is string {
  if (typeof raw !== 'string' || raw.length === 0 || raw.length > MAX_URL_LENGTH) return false;
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    return false;
  }
  if (url.protocol !== 'https:') return false;
  const host = url.hostname.toLowerCase();
  return (CHECKOUT_HOSTS as readonly string[]).includes(host);
}
