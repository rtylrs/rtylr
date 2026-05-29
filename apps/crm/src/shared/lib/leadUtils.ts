export type LeadStatus = 'new' | 'contacted' | 'qualified' | 'proposal' | 'won' | 'lost';
export type LeadSource = 'website' | 'referral' | 'social' | 'email' | 'cold_call' | 'event';

export interface Lead {
  id: number;
  name: string;
  email: string;
  status: LeadStatus;
  source: LeadSource;
  value: number;
  createdAt: string;
  updatedAt: string;
}

export function leadStatusLabel(status: LeadStatus): string {
  const labels: Record<LeadStatus, string> = {
    new: 'New',
    contacted: 'Contacted',
    qualified: 'Qualified',
    proposal: 'Proposal',
    won: 'Won',
    lost: 'Lost',
  };
  return labels[status];
}

export function leadSourceLabel(source: LeadSource): string {
  const labels: Record<LeadSource, string> = {
    website: 'Website',
    referral: 'Referral',
    social: 'Social Media',
    email: 'Email',
    cold_call: 'Cold Call',
    event: 'Event',
  };
  return labels[source];
}

export function isActive(lead: Lead): boolean {
  return lead.status !== 'won' && lead.status !== 'lost';
}

export function isClosed(lead: Lead): boolean {
  return lead.status === 'won' || lead.status === 'lost';
}

export function filterByStatus(leads: Lead[], status: LeadStatus): Lead[] {
  return leads.filter((l) => l.status === status);
}

export function filterBySource(leads: Lead[], source: LeadSource): Lead[] {
  return leads.filter((l) => l.source === source);
}

export function totalPipelineValue(leads: Lead[]): number {
  return leads.filter(isActive).reduce((sum, l) => sum + l.value, 0);
}

export function winRate(leads: Lead[]): number {
  const closed = leads.filter(isClosed);
  if (closed.length === 0) return 0;
  const won = closed.filter((l) => l.status === 'won').length;
  return Math.round((won / closed.length) * 100);
}

export function averageLeadValue(leads: Lead[]): number {
  if (leads.length === 0) return 0;
  return leads.reduce((sum, l) => sum + l.value, 0) / leads.length;
}

export function sortByValue(leads: Lead[], dir: 'asc' | 'desc' = 'desc'): Lead[] {
  return [...leads].sort((a, b) => (dir === 'asc' ? a.value - b.value : b.value - a.value));
}

export function sortByDate(leads: Lead[], field: 'createdAt' | 'updatedAt' = 'createdAt'): Lead[] {
  return [...leads].sort((a, b) => b[field].localeCompare(a[field]));
}

export function groupByStatus(leads: Lead[]): Record<LeadStatus, Lead[]> {
  const groups = {} as Record<LeadStatus, Lead[]>;
  for (const lead of leads) {
    (groups[lead.status] ??= []).push(lead);
  }
  return groups;
}
