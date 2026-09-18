import { describe, expect, it } from 'vitest';
import { buildCustomerAttachment, mapCustomers } from './customerLookupWorkflow';

describe('customerLookupWorkflow', () => {
  it('maps customers with defaults and nullable fields normalized', () => {
    expect(
      mapCustomers([{ id: 12 as unknown as string }]),
    ).toEqual([
      {
        id: '12',
        name: 'Customer',
        phone: undefined,
        email: undefined,
      },
    ]);
  });

  it('builds a valid cart customer attachment', () => {
    expect(buildCustomerAttachment({ id: '42', name: ' Ada Lovelace ' })).toEqual({
      customerId: 42,
      customerName: 'Ada Lovelace',
    });
  });

  it('rejects customer attachments without a valid customer ID', () => {
    expect(() => buildCustomerAttachment({ id: 'local-1', name: 'Ada' })).toThrow(
      'Select a valid customer before attaching it to the order.',
    );
  });
});
