import { describe, expect, it } from 'vitest';
import { getSegmentRuleOperators, SEGMENT_RULE_FIELDS, serializeSegmentRule } from './segmentRules';

describe('segment rules', () => {
  it('serializes rules with canonical field values and normalized whitespace', () => {
    expect(serializeSegmentRule({ field: 'totalSpent', operator: '>', value: ' 500 ' })).toBe(
      'totalSpent > 500'
    );
  });

  it('restricts tier rules to meaningful operators', () => {
    expect(getSegmentRuleOperators('tier')).toEqual(['=', '!=']);
    expect(getSegmentRuleOperators('unknown')).toEqual(['>', '<', '=', '!=']);
  });

  it('keeps the canonical field list stable for every segment builder', () => {
    // lastVisitDays is deliberately absent: onion's segmentrules evaluator has
    // no entry for it, so a saved rule is silently dropped and the segment
    // falls back to every active customer (B42-R07). Add it back only with
    // evaluator support.
    expect([...SEGMENT_RULE_FIELDS]).toEqual(['totalSpent', 'visits', 'loyaltyPoints', 'tier']);
  });
});
