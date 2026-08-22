import { describe, it, expect } from 'vitest';
import {
  autoMapHeaders,
  importColumnValue,
  parseCsv,
  toCustomerImportRow,
} from '../../../../features/customers/ImportCustomers';
import {
  normalizeCustomerPageSize,
  toCustomerCreateWireInput,
  toCustomerWireInput,
} from '../useCustomers';

describe('useCustomers & customer helpers', () => {
  it('parses CSV with quoted fields, escaped quotes, and multiline rows properly (RFC4180)', () => {
    const csv = `name,email,notes\n"Smith, John",john@example.com,"Line 1\nLine 2"\n"Doe, ""Jane""",jane@example.com,654321`;
    const rows = parseCsv(csv);
    expect(rows).toHaveLength(3);
    expect(rows[0]).toEqual(['name', 'email', 'notes']);
    expect(rows[1]).toEqual(['Smith, John', 'john@example.com', 'Line 1\nLine 2']);
    expect(rows[2]).toEqual(['Doe, "Jane"', 'jane@example.com', '654321']);
  });

  it('maps common customer headers and preserves optional import fields', () => {
    const headers = [
      'First Name',
      'Last Name',
      'Email Address',
      'Phone Number',
      'Date of Birth',
      'Tags',
      'Notes',
    ];
    const mapping = autoMapHeaders(headers);
    expect(mapping).toMatchObject({
      first_name: 'First Name',
      last_name: 'Last Name',
      email: 'Email Address',
      phone: 'Phone Number',
    });
    const row = toCustomerImportRow(
      [
        'Jane',
        'Doe',
        'jane@example.com',
        '+1 555 0100',
        '1990-04-12',
        'vip;newsletter',
        'Important',
      ],
      headers,
      mapping
    );

    expect(row).toEqual({
      first_name: 'Jane',
      last_name: 'Doe',
      email: 'jane@example.com',
      phone: '+1 555 0100',
      date_of_birth: '1990-04-12',
      tags: ['vip', 'newsletter'],
      notes: 'Important',
    });
  });

  it('splits a full name when separate name columns are unavailable', () => {
    const headers = ['Full Name', 'Email'];
    const mapping = autoMapHeaders(headers);
    expect(toCustomerImportRow(['Ada Lovelace', 'ada@example.com'], headers, mapping)).toEqual({
      first_name: 'Ada',
      last_name: 'Lovelace',
      email: 'ada@example.com',
      phone: '',
    });
  });

  it('imports the selected duplicate header column instead of the first match', () => {
    const headers = ['Name', 'Email', 'Email'];
    const mapping = { name: 'Name', email: importColumnValue(headers, 2) };

    expect(
      toCustomerImportRow(['Jane Doe', 'old@example.com', 'new@example.com'], headers, mapping)
    ).toMatchObject({
      first_name: 'Jane',
      last_name: 'Doe',
      email: 'new@example.com',
    });
  });

  it('lets an explicit combined-name edit override cached first and last names', () => {
    expect(
      toCustomerWireInput({ name: 'Ada Byron Lovelace', firstName: 'Old', lastName: 'Name' }, true)
    ).toMatchObject({ firstName: 'Ada', lastName: 'Byron Lovelace' });
  });

  it('omits update-only fields from create requests', () => {
    expect(
      toCustomerCreateWireInput({
        firstName: 'Ada',
        lastName: 'Lovelace',
        loyaltyTier: 'Gold',
        isActive: false,
      })
    ).toEqual({
      firstName: 'Ada',
      lastName: 'Lovelace',
      email: '',
      phone: '',
      dateOfBirth: undefined,
      gender: undefined,
      address: undefined,
      city: undefined,
      country: undefined,
      tags: [],
      notes: '',
      consent: undefined,
    });
  });

  it('caps one customer page at the backend limit while allowing all-page loading to continue', () => {
    expect(normalizeCustomerPageSize(200)).toBe(100);
    expect(normalizeCustomerPageSize(25.9)).toBe(25);
    expect(normalizeCustomerPageSize(Number.NaN)).toBe(100);
  });
});
