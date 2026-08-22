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
    expect(SEGMENT_RULE_FIELDS.map((field) => field.value)).toEqual([
      'totalSpent',
      'visits',
      'lastVisitDays',
      'loyaltyPoints',
      'tier',
    ]);
  });
});
