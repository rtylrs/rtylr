import { describe, it, expect } from 'vitest';
import { customerSchema, timelineEntrySchema } from '../customers';

const baseCustomer = {
  id: 1,
  name: 'Jane Doe',
  email: 'jane@example.com',
  phone: '+1 555-0100',
  totalSpent: 1250.0,
  visits: 18,
  loyaltyPoints: 850,
  lastVisit: '2024-05-15',
  initials: 'JD',
};

describe('customerSchema', () => {
  it('parses a minimal valid customer', () => {
    const result = customerSchema.parse(baseCustomer);
    expect(result.id).toBe(1);
    expect(result.name).toBe('Jane Doe');
    expect(result.loyaltyPoints).toBe(850);
  });

  it('coerces string id to number', () => {
    const result = customerSchema.parse({ ...baseCustomer, id: '42' });
    expect(result.id).toBe(42);
  });

  it('accepts optional tags array', () => {
    const result = customerSchema.parse({ ...baseCustomer, tags: ['vip', 'regular'] });
    expect(result.tags).toEqual(['vip', 'regular']);
  });

  it('accepts optional consent object', () => {
    const result = customerSchema.parse({
      ...baseCustomer,
      consent: { email: true, sms: false, marketing: true },
    });
    expect(result.consent?.email).toBe(true);
    expect(result.consent?.sms).toBe(false);
  });

  it('allows null referredBy', () => {
    const result = customerSchema.parse({ ...baseCustomer, referredBy: null });
    expect(result.referredBy).toBeNull();
  });

  it('accepts optional birthday and notes', () => {
    const result = customerSchema.parse({
      ...baseCustomer,
      birthday: '1990-06-15',
      notes: 'Prefers oat milk',
    });
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

  it('coerces string id to number', () => {
    const entry = { id: '99', type: 'note', title: 'A note', at: '2024-05-10' };
    const result = timelineEntrySchema.parse(entry);
    expect(result.id).toBe(99);
  });
});
