import { describe, it, expect } from 'vitest';
import { staffMemberSchema, staffDetailSchema, birthdaySchema, anniversarySchema } from './staff';

// ─── staff Zod schemas - pure parse/validation tests ─────────────────────────

describe('staffMemberSchema', () => {
  it('parses a valid staff member with numeric id', () => {
    const result = staffMemberSchema.parse({
      id: 1,
      name: 'Alice Smith',
      role: 'Manager',
      department: 'Sales',
      email: 'alice@example.com',
      employmentType: 'Full-time',
      status: 'Active',
    });
    expect(result.id).toBe(1);
    expect(result.name).toBe('Alice Smith');
    expect(result.employmentType).toBe('Full-time');
    expect(result.status).toBe('Active');
  });

  it('coerces a string id to a number', () => {
    const result = staffMemberSchema.parse({
      id: '99',
      name: 'Bob',
      role: 'Barista',
      department: 'F&B',
      email: 'bob@example.com',
      employmentType: 'Part-time',
      status: 'Inactive',
    });
    expect(result.id).toBe(99);
  });

  it('falls back to "Full-time" for unrecognised employmentType', () => {
    const result = staffMemberSchema.parse({
      id: 2,
      name: 'Carol',
      role: 'Cashier',
      department: 'Retail',
      email: 'carol@example.com',
      employmentType: 'Freelance',
      status: 'Active',
    });
    expect(result.employmentType).toBe('Full-time');
  });

  it('falls back to "Active" for unrecognised status', () => {
    const result = staffMemberSchema.parse({
      id: 3,
      name: 'Dave',
      role: 'Cook',
      department: 'Kitchen',
      email: 'dave@example.com',
      employmentType: 'Contractor',
      status: 'On Leave',
    });
    expect(result.status).toBe('Active');
  });

  it('rejects missing required fields', () => {
    expect(() => staffMemberSchema.parse({ id: 1, name: 'Eve' })).toThrow();
  });
});

// staffDetailSchema replaced employeeProfileSchema, which described the response
// of /hr/staff/detail. No onion module serves that path, so the schema described
// a shape that never arrived. These five fields do arrive, from getStaff.
describe('staffDetailSchema', () => {
  const BASE = {
    id: 10,
    name: 'Alex Chen',
    role: 'Server',
    department: 'FOH',
    email: 'alex@rtylr.com',
    employmentType: 'Full-time',
    status: 'Active',
  };

  it('carries the contact fields getStaff returns', () => {
    const result = staffDetailSchema.parse({
      ...BASE,
      phone: '+1 555 0001',
      address: '123 Main St',
      startDate: '2021-06-01',
      emergencyContactName: 'Fran Smith',
      emergencyContactPhone: '+1 555 0002',
    });
    expect(result.phone).toBe('+1 555 0001');
    expect(result.address).toBe('123 Main St');
    expect(result.startDate).toBe('2021-06-01');
    expect(result.emergencyContactName).toBe('Fran Smith');
    expect(result.emergencyContactPhone).toBe('+1 555 0002');
  });

  it('leaves every contact field optional', () => {
    const result = staffDetailSchema.parse(BASE);
    expect(result.phone).toBeUndefined();
    expect(result.address).toBeUndefined();
    expect(result.startDate).toBeUndefined();
    expect(result.emergencyContactName).toBeUndefined();
    expect(result.emergencyContactPhone).toBeUndefined();
  });

  it('still enforces everything the list shape requires', () => {
    expect(() => staffDetailSchema.parse({ phone: '+1 555 0000' })).toThrow();
  });
});

describe('birthdaySchema', () => {
  it('parses a valid birthday entry', () => {
    const result = birthdaySchema.parse({
      id: '3',
      staffId: 20,
      name: 'Grace',
      date: '1995-08-22',
      department: 'HR',
    });
    expect(result.id).toBe(3);
    expect(result.name).toBe('Grace');
  });

  it('allows department to be absent', () => {
    const result = birthdaySchema.parse({
      id: 4,
      staffId: 21,
      name: 'Hank',
      date: '1988-12-01',
    });
    expect(result.department).toBeUndefined();
  });

  it('rejects missing date', () => {
    expect(() => birthdaySchema.parse({ id: 1, staffId: 1, name: 'Ivy' })).toThrow();
  });
});

describe('anniversarySchema', () => {
  it('parses a valid anniversary entry', () => {
    const result = anniversarySchema.parse({
      id: '5',
      staffId: 30,
      name: 'Jack',
      date: '2021-03-01',
      years: 5,
    });
    expect(result.id).toBe(5);
    expect(result.years).toBe(5);
  });

  it('handles 0 years (first year partial)', () => {
    const result = anniversarySchema.parse({
      id: 6,
      staffId: 31,
      name: 'Karen',
      date: '2026-01-10',
      years: 0,
    });
    expect(result.years).toBe(0);
  });

  it('rejects missing years', () => {
    expect(() =>
      anniversarySchema.parse({ id: 7, staffId: 32, name: 'Leo', date: '2022-05-05' })
    ).toThrow();
  });
});
