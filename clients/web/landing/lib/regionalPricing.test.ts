import { describe, expect, it } from 'vitest';

import {
  formatRegionalPrice,
  parseTraceCountry,
  regionalPricesFromPlans,
  regionName,
  regionOverrideFromSearch,
} from './regionalPricing';

describe('parseTraceCountry', () => {
  it('reads loc from a Cloudflare trace body', () => {
    const body = 'fl=1\nip=1.2.3.4\ncolo=YUL\nloc=CA\ntls=TLSv1.3\n';
    expect(parseTraceCountry(body)).toBe('CA');
  });

  it('upper-cases the code', () => {
    expect(parseTraceCountry('loc=ca\n')).toBe('CA');
  });

  it('returns null for unknown (XX) and Tor (T1), and for no loc line', () => {
    expect(parseTraceCountry('loc=XX\n')).toBeNull();
    expect(parseTraceCountry('loc=T1\n')).toBeNull();
    expect(parseTraceCountry('colo=YUL\n')).toBeNull();
    expect(parseTraceCountry('')).toBeNull();
  });

  it('ignores a loc that is not the whole line value', () => {
    expect(parseTraceCountry('xloc=CA\n')).toBeNull();
    expect(parseTraceCountry('loc=CAN\n')).toBeNull();
  });
});

describe('regionOverrideFromSearch', () => {
  it('forces a region from ?region=', () => {
    expect(regionOverrideFromSearch('?region=ca')).toBe('CA');
    expect(regionOverrideFromSearch('?utm_source=x&region=CA')).toBe('CA');
  });

  it('forces the default with ?region=default', () => {
    expect(regionOverrideFromSearch('?region=default')).toBe('default');
    expect(regionOverrideFromSearch('?region=DEFAULT')).toBe('default');
  });

  it('is no override for missing or malformed values', () => {
    expect(regionOverrideFromSearch('')).toBeNull();
    expect(regionOverrideFromSearch('?region=')).toBeNull();
    expect(regionOverrideFromSearch('?region=CAN')).toBeNull();
    expect(regionOverrideFromSearch('?region=1A')).toBeNull();
  });
});

describe('regionalPricesFromPlans', () => {
  it('keeps only plans priced for a region', () => {
    const prices = regionalPricesFromPlans([
      { slug: 'free', currency: 'USD' },
      {
        slug: 'pro',
        price_per_seat_cents: 899,
        annual_price_per_seat_cents: 649,
        currency: 'CAD',
        region: 'CA',
      },
      { slug: 'scale', price_per_seat_cents: 499, currency: 'USD' },
    ]);
    expect(prices).toEqual({
      pro: { currency: 'CAD', region: 'CA', monthlyCents: 899, annualCents: 649 },
    });
  });

  it('never surfaces the default USD prices', () => {
    expect(
      regionalPricesFromPlans([{ slug: 'pro', price_per_seat_cents: 699, currency: 'USD' }])
    ).toEqual({});
  });

  it('treats omitted zero prices as free, and ignores unknown slugs and codes', () => {
    const prices = regionalPricesFromPlans([
      { slug: 'free', currency: 'CAD', region: 'CA' },
      { slug: 'flow_solo', currency: 'CAD', region: 'CA', price_per_seat_cents: 100 },
      { slug: 'pro', currency: 'CAD', region: 'CAN', price_per_seat_cents: 100 },
      { slug: 'scale', region: 'CA', price_per_seat_cents: 100 },
    ]);
    expect(prices).toEqual({
      free: { currency: 'CAD', region: 'CA', monthlyCents: 0, annualCents: 0 },
    });
  });
});

describe('formatRegionalPrice', () => {
  const cad = { currency: 'CAD', region: 'CA', monthlyCents: 899, annualCents: 649 };

  it('formats for the price region in the visitor language', () => {
    expect(formatRegionalPrice(899, cad, 'en')).toBe('$8.99');
    expect(formatRegionalPrice(899, cad, 'fr')).toMatch(/^8,99\s\$$/u);
  });

  it('names the currency when the visitor locale would not', () => {
    expect(formatRegionalPrice(899, { ...cad, region: 'US' }, 'en')).toBe('CA$8.99');
  });

  it('falls back to a plain amount for an invalid currency', () => {
    expect(formatRegionalPrice(899, { ...cad, currency: 'nope!' }, 'en')).toBe('8.99 nope!');
  });
});

describe('regionName', () => {
  it('names the region in the visitor language', () => {
    expect(regionName('CA', 'en')).toBe('Canada');
    expect(regionName('CA', 'fr')).toBe('Canada');
  });
});
