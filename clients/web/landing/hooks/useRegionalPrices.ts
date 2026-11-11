import { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import {
  detectCountry,
  fetchRegionalPrices,
  regionOverrideFromSearch,
  type RegionalPrices,
} from '../lib/regionalPricing';

export interface RegionalPricesState {
  /** The region whose prices were found, or null when there are none to show. */
  region: string | null;
  prices: RegionalPrices;
}

const NONE: RegionalPricesState = { region: null, prices: {} };

// One lookup per page load and per override: the pricing page can remount on a
// client-side navigation, and the answer does not change within a visit for
// the same ?region=. The override is part of the key, so opening
// ?region=default and then ?region=CA is two lookups, not one stale answer.
const lookups = new Map<string, Promise<RegionalPricesState>>();

async function resolve(override: string | null): Promise<RegionalPricesState> {
  if (override === 'default') return NONE;
  const region = override ?? (await detectCountry());
  if (!region) return NONE;
  const prices = await fetchRegionalPrices(region);
  return Object.keys(prices).length > 0 ? { region, prices } : NONE;
}

function lookup(override: string | null): Promise<RegionalPricesState> {
  const key = override ?? 'auto';
  let pending = lookups.get(key);
  if (!pending) {
    pending = resolve(override);
    lookups.set(key, pending);
  }
  return pending;
}

/**
 * Regional prices for the visitor, found after the first paint. The static
 * page is the same for everyone, so until this resolves (and for every
 * visitor without regional prices) the page is the no-price page.
 */
export function useRegionalPrices(): RegionalPricesState {
  const [state, setState] = useState<RegionalPricesState>(NONE);
  // The path includes the query string, so a navigation that only changes
  // ?region= re-runs the lookup.
  const { asPath } = useRouter();

  useEffect(() => {
    let active = true;
    const override = regionOverrideFromSearch(window.location.search);
    setState(NONE);
    void lookup(override).then((result) => {
      // An answer for an earlier override must not land on a later one.
      if (active) setState(result);
    });
    return () => {
      active = false;
    };
  }, [asPath]);

  return state;
}
