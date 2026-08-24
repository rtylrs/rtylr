import { describe, it, expect } from 'vitest';
import { toUpdateLoyaltyConfigRequest } from '../useLoyalty';
import { loyaltyConfigSchema } from '../../../schemas/loyalty';

// rtylr#942 regression: useUpdateLoyaltyConfig's mutationFn merges the query
// cache's current config with the caller's partial patch before building the
// wire request (`{ ...current, ...config }`). `current` is populated by
// loyaltyConfigSchema.parse(...) on the real wire response. Before the #942
// fix, bronzeMin/silverMin/goldMin/platinumMin/pointsExpiryDays were plain
// z.number() with no string coercion, so a real wire response (all int64
// fields as strings) parsed those fields to 0 via .catch(0). A later partial
// save that didn't touch those fields would silently re-send zeroed minimums.
describe('toUpdateLoyaltyConfigRequest (rtylr#942)', () => {
  it('preserves a non-zero bronzeMin parsed from a wire-string response through a partial patch', () => {
    const wireResponse = {
      isEnabled: true,
      pointsPerCurrencyUnit: 10,
      currencyPerPoint: 0.01,
      bronzeMin: '250',
      silverMin: '750',
      goldMin: '2000',
      platinumMin: '6000',
      bronzeDiscountPct: 0,
      silverDiscountPct: 5,
      goldDiscountPct: 10,
      platinumDiscountPct: 15,
      pointsExpiryDays: '365',
    };
    const current = loyaltyConfigSchema.parse(wireResponse);

    // Only programEnabled changes; bronzeMin is untouched, exactly like the
    // Loyalty.tsx toggle handler's `updateConfig.mutate({ programEnabled: next })`.
    const request = toUpdateLoyaltyConfigRequest({ ...current, programEnabled: false });

    expect(request.bronzeMin).toBe('250');
    expect(request.silverMin).toBe('750');
    expect(request.goldMin).toBe('2000');
    expect(request.platinumMin).toBe('6000');
    expect(request.pointsExpiryDays).toBe('365');
  });
});
