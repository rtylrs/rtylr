import { CARET_FEATURES } from './featureFlags';

/** Lookup helper for feature flags. Treats missing keys as enabled-false. */
export function isFeatureEnabled(key: keyof typeof CARET_FEATURES): boolean {
  return Boolean(CARET_FEATURES[key]);
}
