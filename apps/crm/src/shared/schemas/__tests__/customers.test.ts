import { describe, it, expect } from 'vitest';
import { customerSchema, timelineEntrySchema } from '../customers';

const baseCustomer = {
  id: 1,
  firstName: 'Jane',
  lastName: 'Doe',
  email: 'jane@example.com',
  phone: '+1 555-0100',
  totalSpent: 1250.0,
  visitCount: 18,
  loyaltyPoints: 850,
  lastVisitAt: '2024-05-15',
  loyaltyTier: 'Gold',
};

describe('customerSchema', () => {
  it('parses a customer from the backend wire shape', () => {
    const result = customerSchema.parse(baseCustomer);
    expect(result.id).toBe(1);
    expect(result.firstName).toBe('Jane');
    expect(result.lastName).toBe('Doe');
    expect(result.loyaltyPoints).toBe(850);
    expect(result.visitCount).toBe(18);
    expect(result.lastVisitAt).toBe('2024-05-15');
    expect(result.loyaltyTier).toBe('Gold');
  });

  it('derives display name, initials, visits and lastVisit', () => {
    const result = customerSchema.parse(baseCustomer);
    expect(result.name).toBe('Jane Doe');
    expect(result.initials).toBe('JD');
    expect(result.visits).toBe(18);
    expect(result.lastVisit).toBe('2024-05-15');
  });

  it('coerces string id to number', () => {
    const result = customerSchema.parse({ ...baseCustomer, id: '42' });
    expect(result.id).toBe(42);
  });

  it('accepts optional tags array', () => {
    const result = customerSchema.parse({ ...baseCustomer, tags: ['vip', 'regular'] });
    expect(result.tags).toEqual(['vip', 'regular']);
  });

  it('allows null referredBy', () => {
    const result = customerSchema.parse({ ...baseCustomer, referredBy: null });
    expect(result.referredBy).toBeNull();
  });

  it('accepts optional dateOfBirth and notes', () => {
    const result = customerSchema.parse({
      ...baseCustomer,
      dateOfBirth: '1990-06-15',
      notes: 'Prefers oat milk',
    });
    expect(result.dateOfBirth).toBe('1990-06-15');
    expect(result.birthday).toBe('1990-06-15');
    expect(result.notes).toBe('Prefers oat milk');
  });
});

describe('timelineEntrySchema', () => {
  it('parses a valid visit entry', () => {
    const entry = {
      id: 1,
      type: 'visit',
      title: 'Store Visit',
      at: '2024-05-01T14:00:00Z',
    };
    const result = timelineEntrySchema.parse(entry);
    expect(result.type).toBe('visit');
    expect(result.title).toBe('Store Visit');
  });

  it('falls back to visit for unknown type', () => {
    const entry = { id: 2, type: 'unknown', title: 'Mystery', at: '2024-05-01' };
    const result = timelineEntrySchema.parse(entry);
    expect(result.type).toBe('visit');
  });

  it('accepts optional description and meta', () => {
    const entry = {
      id: 3,
      type: 'purchase',
      title: 'Purchased latte',
      description: 'Double espresso',
      at: '2024-05-10',
      meta: { amount: 5.5 },
    };
    const result = timelineEntrySchema.parse(entry);
    expect(result.description).toBe('Double espresso');
    expect(result.meta?.amount).toBe(5.5);
  });

  it('preserves backend loyalty activity types', () => {
    expect(
      timelineEntrySchema.parse({ id: 4, type: 'award', title: 'Awarded points', at: '2024-05-10' })
        .type
    ).toBe('award');
    expect(
      timelineEntrySchema.parse({
        id: 5,
        type: 'expire',
        title: 'Expired points',
        at: '2024-05-11',
      }).type
    ).toBe('expire');
  });

  it('coerces string id to number', () => {
    const entry = { id: '99', type: 'note', title: 'A note', at: '2024-05-10' };
    const result = timelineEntrySchema.parse(entry);
    expect(result.id).toBe(99);
  });
});
