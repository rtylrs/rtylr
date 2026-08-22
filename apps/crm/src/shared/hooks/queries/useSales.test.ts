import { describe, expect, it } from 'vitest';
import {
  normalizeContract,
  normalizeLead,
  normalizeTouchpoint,
  mergeContactUpdate,
  mergeContractUpdate,
  toContractMutationPayload,
  normalizeSalesPageSize,
} from './useSales';

describe('sales pagination helpers', () => {
  it('clamps list pages to the backend limit', () => {
    expect(normalizeSalesPageSize(250)).toBe(100);
    expect(normalizeSalesPageSize(0)).toBe(1);
    expect(normalizeSalesPageSize(Number.NaN)).toBe(100);
  });
});

describe('sales contract wire mappings', () => {
  it('normalizes the backend contract response into the CRM model', () => {
    expect(
      normalizeContract({
        id: '4',
        lead_id: '12',
        company_name: 'Acme Inc',
        plan: 'scale',
        mrr: 2500,
        seats: 30,
        signed_at: '2026-08-21',
        signed_by_name: 'Jane Smith',
        closed_by: 17,
        notes: 'Annual plan',
      })
    ).toMatchObject({
      id: 4,
      leadId: 12,
      company: 'Acme Inc',
      plan: 'scale',
      signedBy: 'Jane Smith',
      closedBy: '17',
    });
  });

  it('uses backend field names and numeric closed_by for mutations', () => {
    expect(
      toContractMutationPayload({
        leadId: 12,
        plan: 'pro',
        mrr: 1200,
        seats: 10,
        signedAt: '2026-08-21',
        signedBy: 'Jane Smith',
        closedBy: '17',
        notes: 'Monthly plan',
      })
    ).toEqual({
      leadId: 12,
      plan: 'pro',
      mrr: 1200,
      seats: 10,
      signedAt: '2026-08-21',
      signedByName: 'Jane Smith',
      closedBy: 17,
      notes: 'Monthly plan',
    });
  });
});

describe('sales response normalization', () => {
  it('keeps both company aliases and coerces lead numeric fields', () => {
    expect(
      normalizeLead({
        id: '3',
        company_name: 'Acme Inc',
        stage: 'contacted',
        mrr_potential: '1250',
        assigned_to: '9',
      })
    ).toMatchObject({
      id: 3,
      company: 'Acme Inc',
      companyName: 'Acme Inc',
      stage: 'contacted',
      mrrPotential: 1250,
      assignedTo: 9,
    });
  });

  it('normalizes touchpoint field aliases and defaults invalid enums safely', () => {
    expect(
      normalizeTouchpoint({
        id: '5',
        lead_id: '3',
        type: 'unknown',
        outcome: 'unknown',
        rep_name: 'Alex',
        next_follow_up: '2026-09-01',
      })
    ).toMatchObject({
      id: 5,
      leadId: 3,
      type: 'phone',
      outcome: 'neutral',
      repName: 'Alex',
      nextFollowUp: '2026-09-01',
    });
  });

  it('uses the backend-created author when no rep_name is returned', () => {
    expect(normalizeTouchpoint({ id: 5, lead_id: 3, created_by: '42' })).toMatchObject({
      repName: '42',
    });
  });

  it('preserves required contact fields for partial updates', () => {
    expect(
      mergeContactUpdate(
        { id: 4, leadId: 3, name: 'Jane', role: 'CTO', phone: '1', email: 'a', createdAt: '' },
        { email: 'b' }
      )
    ).toEqual({ name: 'Jane', role: 'CTO', phone: '1', email: 'b' });
  });

  it('preserves required contract fields for partial updates', () => {
    expect(
      mergeContractUpdate(
        {
          id: 4,
          leadId: 3,
          company: 'Acme',
          plan: 'pro',
          mrr: 1200,
          seats: 10,
          signedAt: '2026-08-21',
          signedBy: 'Jane',
          closedBy: '17',
          notes: 'Existing',
          createdAt: '2026-08-21',
        },
        { notes: 'Updated' }
      )
    ).toMatchObject({ plan: 'pro', mrr: 1200, seats: 10, signedBy: 'Jane', notes: 'Updated' });
  });
});
