import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';
import type { LeadListFilters } from './queries';
import { WORKING_STATUSES } from './model';

export const DEFAULT_LEAD_FILTERS: LeadListFilters = {
  statuses: WORKING_STATUSES,
  rep: 'all',
  source: '',
  importId: undefined,
  followUp: '',
  search: '',
  sort: 'newest',
};

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
    }
  )
);
