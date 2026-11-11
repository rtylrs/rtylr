import { useEffect, useState } from 'react';
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

// One lookup per page load: the pricing page can remount on a client-side
// navigation, and the answer does not change within a visit.
let lookup: Promise<RegionalPricesState> | null = null;

async function resolve(): Promise<RegionalPricesState> {
  const override = regionOverrideFromSearch(window.location.search);
  if (override === 'default') return NONE;
  const region = override ?? (await detectCountry());
  if (!region) return NONE;
  const prices = await fetchRegionalPrices(region);
  return Object.keys(prices).length > 0 ? { region, prices } : NONE;
}

/**
 * Regional prices for the visitor, found after the first paint. The static
 * page is the same for everyone, so until this resolves (and for every
 * visitor without regional prices) the page is the no-price page.
 */
export function useRegionalPrices(): RegionalPricesState {
  const [state, setState] = useState<RegionalPricesState>(NONE);

  useEffect(() => {
    let active = true;
    lookup ??= resolve();
    void lookup.then((result) => {
      if (active) setState(result);
    });
    return () => {
      active = false;
    };
  }, []);

  return state;
}
