import type { Plan } from '../components/pricing/pricingData';
import api from './api';

/**
 * Regional prices for /pricing/.
 *
 * rtylr publishes no price on the default page (see pricingData.ts for why),
 * so this only ever fills a price in where an admin has set a regional price
 * for the visitor's region. A plan with no regional price keeps the
 * no-price card, and so does every visitor whose region has none. The default
 * USD prices in the plans table are never shown here: they are what checkout
 * falls back to, not something rtylr has chosen to say out loud.
 *
 * The country comes from Cloudflare, not from the backend or a lookup
 * service: /cdn-cgi/trace is served by the edge in front of rtylr.com and
 * reports the visitor's country as `loc=`. The prices come from onion's
 * anonymous plans endpoint, which does the region lookup.
 */

export interface RegionalPrice {
  /** ISO 4217 code the amounts are quoted in. */
  currency: string;
  /** ISO 3166-1 alpha-2 region the price belongs to. */
  region: string;
  /** Per seat, per month, in the currency's minor unit. */
  monthlyCents: number;
  /** Per seat, per month when billed annually, in the minor unit. 0 = none. */
  annualCents: number;
}

export type RegionalPrices = Partial<Record<Plan['id'], RegionalPrice>>;

const TRACE_URL = '/cdn-cgi/trace';
const TRACE_TIMEOUT_MS = 4000;
const PLANS_PATH = '/hub/subscription/v1/plans';

/** What the plans endpoint returns for one plan. Zero values are omitted. */
interface WirePlan {
  slug?: string;
  price_per_seat_cents?: number;
  annual_price_per_seat_cents?: number;
  currency?: string;
  region?: string;
}

const PLAN_IDS: readonly Plan['id'][] = ['free', 'pro', 'scale'];

function isRegionCode(value: string | null | undefined): value is string {
  return typeof value === 'string' && /^[A-Z]{2}$/.test(value);
}

/**
 * The country in a Cloudflare trace body (`loc=CA`). Cloudflare reports XX
 * when it does not know the country and T1 for Tor, neither of which is a
 * place a price list can be for.
 */
export function parseTraceCountry(body: string): string | null {
  const match = /^loc=([A-Za-z]{2})$/m.exec(body);
  const country = match?.[1]?.toUpperCase();
  if (!isRegionCode(country) || country === 'XX' || country === 'T1') return null;
  return country;
}

/**
 * `?region=CA` forces that region's prices (a Canadian on a VPN, a link in a
 * partner's email). `?region=default` forces no regional prices. Anything
 * else is no override.
 */
export function regionOverrideFromSearch(search: string): string | 'default' | null {
  const raw = new URLSearchParams(search).get('region')?.trim();
  if (!raw) return null;
  if (raw.toLowerCase() === 'default') return 'default';
  const code = raw.toUpperCase();
  return isRegionCode(code) ? code : null;
}

/**
 * Plans that carry a regional price. A plan the backend returned at its
 * default price has no `region`, and is left out. The free plan has no price
 * of its own to quote: with a regional row it shows as Free.
 */
export function regionalPricesFromPlans(plans: readonly WirePlan[]): RegionalPrices {
  const out: RegionalPrices = {};
  for (const plan of plans) {
    const id = PLAN_IDS.find((candidate) => candidate === plan.slug);
    const region = plan.region?.toUpperCase();
    const currency = plan.currency?.toUpperCase();
    if (!id || !isRegionCode(region) || !currency) continue;
    out[id] = {
      currency,
      region,
      monthlyCents: plan.price_per_seat_cents ?? 0,
      annualCents: plan.annual_price_per_seat_cents ?? 0,
    };
  }
  return out;
}

/** The visitor's country per Cloudflare, or null when it cannot be told. */
export async function detectCountry(): Promise<string | null> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TRACE_TIMEOUT_MS);
  try {
    const res = await fetch(TRACE_URL, { signal: controller.signal, cache: 'no-store' });
    if (!res.ok) return null;
    return parseTraceCountry(await res.text());
  } catch {
    return null;
  } finally {
    clearTimeout(timer);
  }
}

/** The regional prices for a region, or none when it has none or the call fails. */
export async function fetchRegionalPrices(region: string): Promise<RegionalPrices> {
  try {
    const base = String(api.defaults.baseURL ?? '').replace(/\/+$/, '');
    // Plain fetch rather than the axios instance: it would attach a signed-in
    // visitor's bearer token to an endpoint that is anonymous.
    const res = await fetch(`${base}${PLANS_PATH}?region=${encodeURIComponent(region)}`, {
      headers: { accept: 'application/json' },
      credentials: 'omit',
    });
    if (!res.ok) return {};
    const data = (await res.json()) as { plans?: WirePlan[] };
    return regionalPricesFromPlans(Array.isArray(data.plans) ? data.plans : []);
  } catch {
    return {};
  }
}

/**
 * An amount in the price's currency, formatted for the visitor's language
 * and the price's region ("$8.99" in en-CA, "8,99 $" in fr-CA, "CA$8.99" for
 * an en-US visitor looking at Canadian prices).
 */
export function formatRegionalPrice(cents: number, price: RegionalPrice, language: string): string {
  const lang = language.split('-')[0] || 'en';
  for (const locale of [`${lang}-${price.region}`, lang]) {
    try {
      return new Intl.NumberFormat(locale, { style: 'currency', currency: price.currency }).format(
        cents / 100
      );
    } catch {
      // Unknown locale or currency: try the next, then fall through.
    }
  }
  return `${(cents / 100).toFixed(2)} ${price.currency}`;
}

/** "Canada" for CA, in the visitor's language. Falls back to the code. */
export function regionName(region: string, language: string): string {
  try {
    return new Intl.DisplayNames([language], { type: 'region' }).of(region) ?? region;
  } catch {
    return region;
  }
}
