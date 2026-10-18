import { describe, expect, it } from 'vitest';
import {
  EMPTY_CONTRACT,
  contractDraftIssue,
  contractStatus,
  toContractRequest,
} from './ContractForm';
import { parseContract } from '../../../shared/hooks/queries/useSales';

describe('contract drafts', () => {
  it('needs what was sold and a value, and a term that does not end before it starts', () => {
    expect(contractDraftIssue(EMPTY_CONTRACT)).toBe('crm.contractForm.planRequired');
    expect(contractDraftIssue({ ...EMPTY_CONTRACT, plan: 'Setup' })).toBe(
      'crm.contractForm.valueRequired'
    );
    expect(
      contractDraftIssue({
        ...EMPTY_CONTRACT,
        plan: 'Setup',
        value: '300',
        startDate: '2026-11-01',
        endDate: '2026-10-01',
      })
    ).toBe('crm.contractForm.endBeforeStart');
    expect(contractDraftIssue({ ...EMPTY_CONTRACT, plan: 'Setup', value: '0' })).toBeNull();
  });

  it('sends an unsigned draft without a signature date and with no rep', () => {
    expect(
      toContractRequest({ ...EMPTY_CONTRACT, plan: ' Setup ', value: '300', billing: 'one_time' })
    ).toMatchObject({
      plan: 'Setup',
      value: 300,
      billing: 'one_time',
      signedAt: undefined,
      closedBy: null,
    });
  });
});

describe('contractStatus', () => {
  it('is a draft until signed, and ended once the end date has passed', () => {
    expect(contractStatus({ signedAt: '', endDate: undefined })).toBe('draft');
    expect(contractStatus({ signedAt: '2026-01-01', endDate: undefined })).toBe('signed');
    expect(contractStatus({ signedAt: '2025-01-01', endDate: '2025-12-31' })).toBe('ended');
  });
});

describe('parseContract', () => {
  it('reads a contract saved before value existed as monthly at its MRR', () => {
    const contract = parseContract({ id: '3', lead_id: '9', plan: 'pro', mrr: '99.00', seats: 5 });
    expect(contract).toMatchObject({ id: 3, leadId: 9, value: 99, billing: 'monthly' });
  });

  it('keeps the value, billing, term and linked customer', () => {
    const contract = parseContract({
      id: 4,
      plan: 'Annual maintenance',
      mrr: '100.00',
      value: '1200.00',
      billing: 'yearly',
      start_date: '2026-11-01',
      customer_id: 42,
      closed_by: 1,
    });
    expect(contract).toMatchObject({
      value: 1200,
      billing: 'yearly',
      startDate: '2026-11-01',
      endDate: undefined,
      customerId: 42,
      closedBy: 1,
    });
  });
});
