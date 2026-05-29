import { describe, it, expect } from 'vitest';
import { featuresForPlan, hasFeature, minimumPlanFor, planLabel, planAtLeast } from './scopeUtils';

describe('featuresForPlan', () => {
  it('returns empty array for free plan', () => {
    expect(featuresForPlan('free')).toEqual([]);
  });

  it('returns features for starter plan', () => {
    const features = featuresForPlan('starter');
    expect(features).toContain('advanced_analytics');
    expect(features).toContain('bulk_export');
  });

  it('enterprise has all features', () => {
    const features = featuresForPlan('enterprise');
    expect(features).toContain('sso');
    expect(features).toContain('white_label');
    expect(features).toContain('custom_roles');
  });
});

describe('hasFeature', () => {
  it('returns true when plan has the feature', () => {
    expect(hasFeature('pro', 'api_keys')).toBe(true);
  });

  it('returns false when plan does not have the feature', () => {
    expect(hasFeature('starter', 'sso')).toBe(false);
  });

  it('returns false for free plan', () => {
    expect(hasFeature('free', 'bulk_export')).toBe(false);
  });
});

describe('minimumPlanFor', () => {
  it('returns starter for bulk_export', () => {
    expect(minimumPlanFor('bulk_export')).toBe('starter');
  });

  it('returns pro for api_keys', () => {
    expect(minimumPlanFor('api_keys')).toBe('pro');
  });

  it('returns enterprise for sso', () => {
    expect(minimumPlanFor('sso')).toBe('enterprise');
  });
});

describe('planLabel', () => {
  it('returns "Free" for free', () => {
    expect(planLabel('free')).toBe('Free');
  });

  it('returns "Enterprise" for enterprise', () => {
    expect(planLabel('enterprise')).toBe('Enterprise');
  });
});

describe('planAtLeast', () => {
  it('returns true when plans are equal', () => {
    expect(planAtLeast('pro', 'pro')).toBe(true);
  });

  it('returns true when plan is higher tier', () => {
    expect(planAtLeast('enterprise', 'starter')).toBe(true);
  });

  it('returns false when plan is lower tier', () => {
    expect(planAtLeast('free', 'pro')).toBe(false);
  });
});
