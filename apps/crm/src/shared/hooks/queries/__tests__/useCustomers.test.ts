import { describe, it, expect } from 'vitest';
import { parseCsv } from '../../../../features/customers/ImportCustomers';

describe('useCustomers & customer helpers', () => {
  it('parses CSV with quoted fields and multiline rows properly (RFC4180)', () => {
    const csv = `name,email,phone\n"Smith, John",john@example.com,123456\n"Doe, ""Jane""",jane@example.com,654321`;
    const rows = parseCsv(csv);
    expect(rows).toHaveLength(3);
    expect(rows[0]).toEqual(['name', 'email', 'phone']);
    expect(rows[1]).toEqual(['Smith, John', 'john@example.com', '123456']);
    expect(rows[2]).toEqual(['Doe, "Jane"', 'jane@example.com', '654321']);
  });
});
