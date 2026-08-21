import { createApiClient, safeFetch } from '@voxire/api';
import type { LicenseBannerProps, LicenseState } from '@voxire/components';

/**
 * The app's REST transport, for onion routes that have no oneKit contract.
 *
 * Most of this app goes through the generated clients. A few call sites cannot:
 * they hit real onion routes with no `.onk` definition, so no client is
 * generated for them (part of #873's contract drift). Those are reached here and
 * parsed at the boundary with Zod, per CLAUDE.md rule 5.
 */
export const api = createApiClient('insights');
export const safe = safeFetch(api);

/**
 * Satisfies @voxire/components' LicenseBanner without a generated client.
 *
 * The prop's type is a minimal structural interface with one method, so an
 * adapter over the REST route it documents is a complete implementation rather
 * than a stand-in. `GET /license/state` is the endpoint that type's own doc
 * comment names. Derived from LicenseBannerProps because the interface itself is
 * not exported from the package index.
 */
export const licenseClient: LicenseBannerProps['client'] = {
  getLicenseState: async (): Promise<LicenseState> => {
    const result = await safe.get<LicenseState>('/license/state');
    if (result.isErr()) throw result.error;

    return result.value;
  },
};
