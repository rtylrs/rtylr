/**
 * Hand-curated quick-action templates surfaced as suggestion pills above
 * the composer. Each one expands to a fully formed user prompt.
 */

export interface QuickAction {
  id: string;
  label: string;
  prompt: string;
  scope: 'global' | 'orders' | 'inventory' | 'team';
}

export const QUICK_ACTIONS: QuickAction[] = [
  {
    id: 'today-revenue',
    label: "Today's revenue",
    prompt: 'How much revenue did we make today vs yesterday?',
    scope: 'global',
  },
  {
    id: 'top-customers',
    label: 'Top customers',
    prompt: 'Show my top 10 customers by lifetime value.',
    scope: 'global',
  },
  {
    id: 'low-stock',
    label: 'Low stock items',
    prompt: 'List products at or below their reorder threshold.',
    scope: 'inventory',
  },
  {
    id: 'pending-refunds',
    label: 'Pending refunds',
    prompt: 'Are there any pending refunds I need to approve?',
    scope: 'orders',
  },
  {
    id: 'team-active',
    label: 'Who is online',
    prompt: 'Which team members are currently active?',
    scope: 'team',
  },
];

export function quickActionsByScope(scope: QuickAction['scope']): QuickAction[] {
  return QUICK_ACTIONS.filter((q) => q.scope === scope);
}
