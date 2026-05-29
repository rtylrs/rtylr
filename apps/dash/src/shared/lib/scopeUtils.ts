/**
 * Feature scope and access control utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type FeatureFlag =
  | 'advanced_analytics'
  | 'bulk_export'
  | 'api_keys'
  | 'webhooks'
  | 'sso'
  | 'audit_log'
  | 'custom_roles'
  | 'white_label';

export type PlanTier = 'free' | 'starter' | 'pro' | 'enterprise';

const PLAN_FEATURES: Record<PlanTier, FeatureFlag[]> = {
  free: [],
  starter: ['advanced_analytics', 'bulk_export'],
  pro: ['advanced_analytics', 'bulk_export', 'api_keys', 'webhooks', 'audit_log'],
  enterprise: [
    'advanced_analytics',
    'bulk_export',
    'api_keys',
    'webhooks',
    'audit_log',
    'sso',
    'custom_roles',
    'white_label',
  ],
};

/** Return all features available for the given plan tier. */
export function featuresForPlan(plan: PlanTier): FeatureFlag[] {
  return PLAN_FEATURES[plan] ?? [];
}

/** Return true when the plan includes a specific feature flag. */
export function hasFeature(plan: PlanTier, feature: FeatureFlag): boolean {
  return featuresForPlan(plan).includes(feature);
}

/** Return the minimum plan tier required for a given feature. */
export function minimumPlanFor(feature: FeatureFlag): PlanTier {
  const tiers: PlanTier[] = ['free', 'starter', 'pro', 'enterprise'];
  for (const tier of tiers) {
    if (hasFeature(tier, feature)) return tier;
  }
  return 'enterprise';
}

/** Return a human-readable label for a plan tier. */
export function planLabel(plan: PlanTier): string {
  const labels: Record<PlanTier, string> = {
    free: 'Free',
    starter: 'Starter',
    pro: 'Pro',
    enterprise: 'Enterprise',
  };
  return labels[plan];
}

/** Return true when plan `a` is at least as capable as plan `b`. */
export function planAtLeast(a: PlanTier, b: PlanTier): boolean {
  const order: PlanTier[] = ['free', 'starter', 'pro', 'enterprise'];
  return order.indexOf(a) >= order.indexOf(b);
}
