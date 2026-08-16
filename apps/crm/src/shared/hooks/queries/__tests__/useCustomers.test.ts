import { describe, it, expect } from 'vitest';
import { parseCsv } from '../../../../features/customers/ImportCustomers';

describe('useCustomers & customer helpers', () => {
  it('parses CSV with quoted fields, escaped quotes, and multiline rows properly (RFC4180)', () => {
    const csv = `name,email,notes\n"Smith, John",john@example.com,"Line 1\nLine 2"\n"Doe, ""Jane""",jane@example.com,654321`;
    const rows = parseCsv(csv);
    expect(rows).toHaveLength(3);
    expect(rows[0]).toEqual(['name', 'email', 'notes']);
    expect(rows[1]).toEqual(['Smith, John', 'john@example.com', 'Line 1\nLine 2']);
    expect(rows[2]).toEqual(['Doe, "Jane"', 'jane@example.com', '654321']);
  });
});
