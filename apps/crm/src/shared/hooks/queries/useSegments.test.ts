import { describe, expect, it } from 'vitest';
import {
  getNextSegmentCustomerPage,
  parseSegmentCustomersPage,
  type SegmentCustomersPage,
} from './useSegments';

const pageWithCustomers = (count: number, total: number, page: number): SegmentCustomersPage => ({
  customers: Array.from({ length: count }, () => ({}) as SegmentCustomersPage['customers'][number]),
  total,
  page,
  pageSize: 50,
});

describe('segment customer pagination', () => {
  it('preserves the API pagination metadata', () => {
    expect(
      parseSegmentCustomersPage({ customers: [], total: '101', page: '2', pageSize: '50' }, 1)
    ).toEqual({ customers: [], total: 101, page: 2, pageSize: 50 });
  });

  it('requests the next page until the total has been loaded', () => {
    expect(getNextSegmentCustomerPage(pageWithCustomers(50, 101, 1))).toBe(2);
    expect(getNextSegmentCustomerPage(pageWithCustomers(50, 101, 2))).toBe(3);
    expect(getNextSegmentCustomerPage(pageWithCustomers(1, 101, 3))).toBeUndefined();
  });

  it('stops when a page is empty or reaches the total', () => {
    expect(getNextSegmentCustomerPage(pageWithCustomers(50, 100, 2))).toBeUndefined();
    expect(getNextSegmentCustomerPage(pageWithCustomers(0, 101, 1))).toBeUndefined();
  });
});
