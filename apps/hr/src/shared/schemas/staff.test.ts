import { describe, it, expect } from 'vitest';
import {
  staffMemberSchema,
  employeeProfileSchema,
  birthdaySchema,
  anniversarySchema,
} from './staff';

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

describe('employeeProfileSchema', () => {
  it('parses a full profile', () => {
    const result = employeeProfileSchema.parse({
      id: '10',
      photoUrl: 'https://cdn.example.com/avatar.jpg',
      phone: '+1 555 0001',
      address: '123 Main St',
      birthday: '1990-03-15',
      hiredAt: '2021-06-01',
      emergency: { name: 'Fran Smith', relation: 'Spouse', phone: '+1 555 0002' },
      bio: 'Experienced team lead.',
      managerId: 7,
      skills: ['Leadership', 'TypeScript'],
    });
    expect(result.id).toBe(10);
    expect(result.emergency?.relation).toBe('Spouse');
    expect(result.skills).toHaveLength(2);
  });

  it('allows all optional fields to be absent', () => {
    const result = employeeProfileSchema.parse({ id: 11 });
    expect(result.id).toBe(11);
    expect(result.photoUrl).toBeUndefined();
    expect(result.emergency).toBeUndefined();
    expect(result.skills).toBeUndefined();
  });

  it('coerces string id to number', () => {
    const result = employeeProfileSchema.parse({ id: '55' });
    expect(result.id).toBe(55);
  });

  it('rejects missing id', () => {
    expect(() => employeeProfileSchema.parse({ phone: '+1 555 0000' })).toThrow();
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
