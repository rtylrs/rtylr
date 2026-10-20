import type {
  Activity,
  Contract,
  Deal,
  Lead,
  LeadFields,
  LeadImport,
  Pipeline,
  SalesTask,
  Stage,
} from '@voxire/voxie/crm/sales/v1/types';

export type { Activity, Contract, Deal, Lead, LeadFields, LeadImport, Pipeline, SalesTask, Stage };

// Ids only: labels come from crmSales.<group>.<id>.
export const LEAD_STATUSES = [
  'new',
  'attempting',
  'connected',
  'nurture',
  'qualified',
  'disqualified',
] as const;
export type LeadStatus = (typeof LEAD_STATUSES)[number];
/** Statuses of leads still being prospected. */
export const WORKING_STATUSES: LeadStatus[] = ['new', 'attempting', 'connected', 'nurture'];

export const CALL_OUTCOMES = [
  'connected',
  'no_answer',
  'voicemail',
  'busy',
  'wrong_number',
  'callback',
  'meeting_booked',
  'not_interested',
] as const;
export type CallOutcome = (typeof CALL_OUTCOMES)[number];
/** Outcomes where the person picked up, which usually deserve a note. */
export const REACHED_OUTCOMES: CallOutcome[] = [
  'connected',
  'callback',
  'meeting_booked',
  'not_interested',
];

export const ACTIVITY_TYPES = ['call', 'email', 'meeting', 'sms', 'note'] as const;
export type ActivityType = (typeof ACTIVITY_TYPES)[number];

export const TASK_TYPES = ['call', 'email', 'meeting', 'todo'] as const;
export type TaskType = (typeof TASK_TYPES)[number];

export const BILLINGS = ['one_time', 'monthly', 'yearly'] as const;
export type Billing = (typeof BILLINGS)[number];

export const LEAD_SORTS = ['newest', 'follow_up', 'recent_activity', 'company'] as const;
export type LeadSort = (typeof LEAD_SORTS)[number];

export const DEAL_SORTS = ['newest', 'close_date', 'value'] as const;
export type DealSort = (typeof DEAL_SORTS)[number];

export const FOLLOW_UP_WINDOWS = [
  'overdue',
  'today',
  'due',
  'upcoming',
  'scheduled',
  'none',
] as const;
export type FollowUpWindow = (typeof FOLLOW_UP_WINDOWS)[number];

/** Whether a field is filled in: any, yes or no. */
export type Presence = '' | 'yes' | 'no';

export const DEAL_STATUSES = ['open', 'won', 'lost'] as const;
export type DealStatus = (typeof DEAL_STATUSES)[number];

/** A rep filter: everyone, the caller, nobody, or one teammate's id. */
export type RepFilter = 'all' | 'unassigned' | string;

/** The assigned_to query value for a rep filter; "0" means unassigned. */
export function repFilterParam(filter: RepFilter): string | undefined {
  if (filter === 'all' || filter === '') return undefined;
  if (filter === 'unassigned') return '0';
  return filter;
}

export function personName(person: { firstName?: string; lastName?: string }): string {
  return [person.firstName, person.lastName]
    .map((part) => part?.trim())
    .filter(Boolean)
    .join(' ');
}

/** What a lead is called in lists: the company, else the person. */
export function leadTitle(lead: Lead): string {
  return lead.companyName?.trim() || personName(lead) || '—';
}

export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  return (
    parts
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? '')
      .join('') || '?'
  );
}

/** A decimal string from the API as a number of major units. */
export function amountOf(value?: string): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

/** Unix seconds from the API as a Date. */
export function dateOf(seconds?: string): Date | undefined {
  const parsed = Number(seconds);
  return Number.isFinite(parsed) && parsed > 0 ? new Date(parsed * 1000) : undefined;
}

function pad(value: number): string {
  return String(value).padStart(2, '0');
}

/** A Date as the caller's local calendar date, YYYY-MM-DD. */
export function isoDate(date: Date): string {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

export function todayISO(): string {
  return isoDate(new Date());
}

export function addDaysISO(days: number, from = new Date()): string {
  const next = new Date(from);
  next.setDate(next.getDate() + days);
  return isoDate(next);
}

/** The caller's offset from UTC in minutes, east positive. */
export function utcOffsetMinutes(): number {
  return -new Date().getTimezoneOffset();
}

export type DueState = 'overdue' | 'today' | 'upcoming';

export function dueState(date: string | undefined, today = todayISO()): DueState | undefined {
  if (!date) return undefined;
  if (date < today) return 'overdue';
  if (date === today) return 'today';
  return 'upcoming';
}

/** Formats a YYYY-MM-DD date for the reader's locale without shifting it. */
export function formatDay(date: string | undefined, locale: string): string {
  if (!date) return '—';
  const [year, month, day] = date.split('-').map(Number);
  if (!year || !month || !day) return date;
  return new Intl.DateTimeFormat(locale, { day: 'numeric', month: 'short' }).format(
    new Date(year, month - 1, day)
  );
}

const RELATIVE_UNITS: Array<[Intl.RelativeTimeFormatUnit, number]> = [
  ['year', 365 * 24 * 3600],
  ['month', 30 * 24 * 3600],
  ['week', 7 * 24 * 3600],
  ['day', 24 * 3600],
  ['hour', 3600],
  ['minute', 60],
];

/** "3 days ago", in the reader's locale. */
export function formatRelative(date: Date | undefined, locale: string, now = Date.now()): string {
  if (!date) return '—';
  const seconds = Math.round((date.getTime() - now) / 1000);
  const format = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });
  for (const [unit, size] of RELATIVE_UNITS) {
    if (Math.abs(seconds) >= size) return format.format(Math.round(seconds / size), unit);
  }
  return format.format(0, 'minute');
}

export function telHref(phone?: string): string | undefined {
  const cleaned = phone?.replace(/[^\d+]/g, '');
  return cleaned ? `tel:${cleaned}` : undefined;
}

export function stageById(
  pipeline: Pipeline | undefined,
  id: string | undefined
): Stage | undefined {
  return pipeline?.stages?.find((stage) => stage.id === id);
}

export function openStages(pipeline: Pipeline | undefined): Stage[] {
  return (pipeline?.stages ?? []).filter((stage) => stage.kind === 'open');
}

export function isClosedDeal(deal: Deal): boolean {
  return deal.status === 'won' || deal.status === 'lost';
}

/** Empty strings dropped, so optional API fields stay absent. */
export function compact<T extends Record<string, unknown>>(value: T): Partial<T> {
  const out: Partial<T> = {};
  for (const [key, field] of Object.entries(value)) {
    if (field !== '' && field !== undefined && field !== null) {
      (out as Record<string, unknown>)[key] = field;
    }
  }
  return out;
}
