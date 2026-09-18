import type { HRCompanyConfigPayload } from '@voxire/voxie/hr/setup/v1/types';

/**
 * Whether this tenant has finished HR setup.
 *
 * `setupCompletedAt` is the only marker the app gates on, and it is written
 * exactly once - by the wizard's final step. `GetSetupStatusResponse.isComplete`
 * looks interchangeable but is not: it reports that all eight wizard steps are
 * ticked, which is the precondition for finalising rather than the marker
 * itself. A tenant can sit with every step ticked and no timestamp, whenever
 * the closing `PUT /hr/config` never landed.
 *
 * Every guard that routes on completion must call this and nothing else. While
 * `HRConfigRoute` and `SetupWizard` each carried their own predicate they
 * disagreed for precisely that tenant and bounced `/` <-> `/setup` forever.
 */
export function isHRSetupComplete(config: HRCompanyConfigPayload | null | undefined): boolean {
  // The timestamp is an int64, so it crosses the wire as a string and an unset
  // value can arrive as absent, null, "" or "0". Comparing numerically treats
  // all of them as unfinished and mirrors the backend's own `@gt(0)` bounds.
  const completedAt = Number(config?.setupCompletedAt ?? 0);
  return Number.isFinite(completedAt) && completedAt > 0;
}
