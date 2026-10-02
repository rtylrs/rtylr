import { describe, expect, it } from 'vitest';
import { centsToDisplay } from './useFlowSubscription';

// FlowSettings.test.tsx mocks centsToDisplay, so nothing asserted its real
// output. These cases pin the product rule the function documents: RTYLR
// subscription amounts are USD, and the function accepts no currency to label
// them with anything else.
describe('centsToDisplay', () => {
  it('formats cents as a USD amount', () => {
    expect(centsToDisplay(1200)).toBe('$12.00');
  });

  it('keeps sub-unit precision rather than rounding to whole dollars', () => {
    expect(centsToDisplay(1234)).toBe('$12.34');
  });

  it('formats zero as a currency amount rather than an empty string', () => {
    expect(centsToDisplay(0)).toBe('$0.00');
  });

  it('groups thousands', () => {
    expect(centsToDisplay(123456789)).toBe('$1,234,567.89');
  });
});
