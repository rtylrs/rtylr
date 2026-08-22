export const SEGMENT_RULE_FIELDS = [
  { value: 'totalSpent', label: 'Total Spent' },
  { value: 'visits', label: 'Visits' },
  { value: 'lastVisitDays', label: 'Days Since Visit' },
  { value: 'loyaltyPoints', label: 'Loyalty Points' },
  { value: 'tier', label: 'Tier' },
] as const;

export type SegmentRuleField = (typeof SEGMENT_RULE_FIELDS)[number]['value'];

export const SEGMENT_RULE_OPERATORS = ['>', '<', '=', '!='] as const;
export type SegmentRuleOperator = (typeof SEGMENT_RULE_OPERATORS)[number];

export const SEGMENT_RULE_OPERATORS_BY_FIELD: Record<
  SegmentRuleField,
  readonly SegmentRuleOperator[]
> = {
  totalSpent: ['>', '<', '=', '!='],
  visits: ['>', '<', '=', '!='],
  lastVisitDays: ['>', '<', '=', '!='],
  loyaltyPoints: ['>', '<', '=', '!='],
  tier: ['=', '!='],
};

export const SEGMENT_RULE_VALUES: Record<SegmentRuleField, readonly string[]> = {
  totalSpent: ['100', '250', '500', '1000', '2500'],
  visits: ['1', '3', '5', '10', '20'],
  lastVisitDays: ['7', '14', '30', '60', '90'],
  loyaltyPoints: ['100', '250', '500', '1000', '2500'],
  tier: ['Bronze', 'Silver', 'Gold', 'Platinum'],
};

export interface SegmentRule {
  id: number;
  field: SegmentRuleField;
  operator: SegmentRuleOperator;
  value: string;
}

export function getSegmentRuleOperators(field: string): readonly SegmentRuleOperator[] {
  return SEGMENT_RULE_OPERATORS_BY_FIELD[field as SegmentRuleField] ?? SEGMENT_RULE_OPERATORS;
}

export function serializeSegmentRule(rule: {
  field: string;
  operator: string;
  value: string;
}): string {
  return `${rule.field} ${rule.operator} ${rule.value.trim()}`;
}
