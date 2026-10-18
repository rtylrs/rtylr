import { describe, expect, it } from 'vitest';
import { parseLead } from './useSales';

describe('parseLead wire normalization', () => {
  it('normalizes raw snake_case backend payload into camelCase Lead', () => {
    const rawBackendLead = {
      id: 42,
      owner_id: 1,
      business_id: 10,
      company_name: 'Stark Industries',
      industry: 'Defense',
      size: '100-500',
      source: 'Referral',
      stage: 'cold',
      assigned_to: 7,
      mrr_potential: '2500.00',
      notes: 'High priority lead',
      created_at: '2026-05-01T12:00:00Z',
      updated_at: '2026-05-02T12:00:00Z',
    };

    const lead = parseLead(rawBackendLead as any);

    expect(lead.id).toBe(42);
    expect(lead.companyName).toBe('Stark Industries');
    expect(lead.company).toBe('Stark Industries');
    expect(lead.stage).toBe('cold');
    expect(lead.industry).toBe('Defense');
    expect(lead.mrrPotential).toBe(2500);
    expect(lead.dealValue).toBe(2500);
    expect(lead.assignedTo).toBe(7);
    expect(lead.createdAt).toBe('2026-05-01T12:00:00Z');
    expect(lead.updatedAt).toBe('2026-05-02T12:00:00Z');
  });

  it('handles null assignedTo and missing optional fields gracefully', () => {
    const raw = {
      id: '99',
      stage: 'contacted',
      assigned_to: null,
      mrr_potential: null,
    };

    const lead = parseLead(raw as any);

    expect(lead.id).toBe(99);
    expect(lead.stage).toBe('contacted');
    expect(lead.assignedTo).toBeNull();
    expect(lead.mrrPotential).toBeUndefined();
    expect(lead.dealValue).toBeUndefined();
    expect(lead.companyName).toBe('');
  });

  it('carries the next follow-up date, and leaves it out when there is none', () => {
    expect(parseLead({ id: 1, stage: 'contacted', next_follow_up: '2026-10-25' })).toMatchObject({
      nextFollowUp: '2026-10-25',
    });
    expect(parseLead({ id: 2, stage: 'contacted' }).nextFollowUp).toBeUndefined();
  });
});
