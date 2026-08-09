/** Feature flag helpers for the auth app. */

export type FeatureFlags = Record<string, boolean>;

/**
 * Returns true if the given feature flag is enabled.
 */
export function isEnabled(flags: FeatureFlags, key: string): boolean {
  return Boolean(flags[key]);
}

/**
 * Returns all enabled flag keys from the given flags object.
 */
export function enabledFlags(flags: FeatureFlags): string[] {
  return Object.entries(flags)
    .filter(([, value]) => value)
    .map(([key]) => key);
}

/**
 * Returns all disabled flag keys.
 */
export function disabledFlags(flags: FeatureFlags): string[] {
  return Object.entries(flags)
    .filter(([, value]) => !value)
    .map(([key]) => key);
}

/**
 * Returns a new flags object with the given overrides applied.
 */
export function withOverrides(flags: FeatureFlags, overrides: Partial<FeatureFlags>): FeatureFlags {
  return { ...flags, ...overrides };
}

/**
 * Returns true if all the given keys are enabled.
 */
export function allEnabled(flags: FeatureFlags, keys: string[]): boolean {
  return keys.every((k) => isEnabled(flags, k));
}

/**
 * Returns true if at least one of the given keys is enabled.
 */
export function anyEnabled(flags: FeatureFlags, keys: string[]): boolean {
  return keys.some((k) => isEnabled(flags, k));
}
