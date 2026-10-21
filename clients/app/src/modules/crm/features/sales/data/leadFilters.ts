import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import type { LeadListFilters } from './queries';
import { WORKING_STATUSES } from './model';

/** Everything the Filter popover sets, cleared. */
export const NO_LEAD_FILTERS = {
  reps: [],
  sources: [],
  industries: [],
  sizes: [],
  countries: [],
  cities: [],
  billing: [],
  importId: undefined,
  followUp: '',
  hasPhone: '',
  hasEmail: '',
  hasWebsite: '',
  hasLinkedin: '',
  hasCompany: '',
  hasValue: '',
  contacted: '',
  attemptsMin: '',
  attemptsMax: '',
  valueMin: '',
  valueMax: '',
  createdFrom: '',
  createdTo: '',
  activeFrom: '',
  activeTo: '',
} satisfies Omit<LeadListFilters, 'statuses' | 'search' | 'sort'>;

export const DEFAULT_LEAD_FILTERS: LeadListFilters = {
  ...NO_LEAD_FILTERS,
  statuses: WORKING_STATUSES,
  search: '',
  sort: 'newest',
};

/**
 * The popover's filter groups, each with the fields it owns. A range or a
 * pair of dates counts once.
 */
export const LEAD_FILTER_GROUPS = {
  reps: ['reps'],
  sources: ['sources'],
  importId: ['importId'],
  industries: ['industries'],
  sizes: ['sizes'],
  countries: ['countries'],
  cities: ['cities'],
  billing: ['billing'],
  followUp: ['followUp'],
  hasPhone: ['hasPhone'],
  hasEmail: ['hasEmail'],
  hasWebsite: ['hasWebsite'],
  hasLinkedin: ['hasLinkedin'],
  hasCompany: ['hasCompany'],
  hasValue: ['hasValue'],
  contacted: ['contacted'],
  attempts: ['attemptsMin', 'attemptsMax'],
  value: ['valueMin', 'valueMax'],
  created: ['createdFrom', 'createdTo'],
  active: ['activeFrom', 'activeTo'],
} as const satisfies Record<string, readonly (keyof typeof NO_LEAD_FILTERS)[]>;

export type LeadFilterGroup = keyof typeof LEAD_FILTER_GROUPS;

const isSet = (value: unknown): boolean =>
  Array.isArray(value) ? value.length > 0 : value !== undefined && value !== '';

/** The popover groups that currently narrow the list. */
export function activeLeadFilterGroups(filters: LeadListFilters): LeadFilterGroup[] {
  return (Object.keys(LEAD_FILTER_GROUPS) as LeadFilterGroup[]).filter((group) =>
    LEAD_FILTER_GROUPS[group].some((field) => isSet(filters[field]))
  );
}

/** The change that clears one group. */
export function clearLeadFilterGroup(group: LeadFilterGroup): Partial<LeadListFilters> {
  return Object.fromEntries(
    LEAD_FILTER_GROUPS[group].map((field) => [field, NO_LEAD_FILTERS[field]])
  );
}

interface LeadFilterState {
  filters: LeadListFilters;
  setFilters: (change: Partial<LeadListFilters>) => void;
  reset: () => void;
}

/**
 * The leads list's filters. Calling mode works through the same list, so the
 * filters outlive the page for the browser session.
 */
export const useLeadFilters = create<LeadFilterState>()(
  persist(
    (set) => ({
      filters: DEFAULT_LEAD_FILTERS,
      setFilters: (change) => set((state) => ({ filters: { ...state.filters, ...change } })),
      reset: () => set({ filters: DEFAULT_LEAD_FILTERS }),
    }),
    {
      name: 'crm-sales-lead-filters',
      storage: createJSONStorage(() => sessionStorage),
      // v1 added the popover's filters; older sessions start over.
      version: 1,
      migrate: () => ({ filters: DEFAULT_LEAD_FILTERS }),
    }
  )
);
