/**
 * Centralized query key factory for all CRM TanStack Query queries.
 * Using a factory pattern ensures consistent cache invalidation and prevents typos.
 */

export const crmKeys = {
  // Customers
  customers: {
    all: () => ['crm', 'customers'] as const,
    list: (filters?: Record<string, unknown>) => ['crm', 'customers', 'list', filters] as const,
    detail: (id: number) => ['crm', 'customers', 'detail', id] as const,
    timeline: (id: number) => ['crm', 'customers', 'timeline', id] as const,
  },

  // Campaigns
  campaigns: {
    all: () => ['crm', 'campaigns'] as const,
    list: (filters?: Record<string, unknown>) => ['crm', 'campaigns', 'list', filters] as const,
    detail: (id: number) => ['crm', 'campaigns', 'detail', id] as const,
    stats: (id: number) => ['crm', 'campaigns', 'stats', id] as const,
  },

  // Loyalty
  loyalty: {
    all: () => ['crm', 'loyalty'] as const,
    config: () => ['crm', 'loyalty', 'config'] as const,
    members: () => ['crm', 'loyalty', 'members'] as const,
    rewards: () => ['crm', 'loyalty', 'rewards'] as const,
    tiers: () => ['crm', 'loyalty', 'tiers'] as const,
    transactions: () => ['crm', 'loyalty', 'transactions'] as const,
  },

  // Segments
  segments: {
    all: () => ['crm', 'segments'] as const,
    list: () => ['crm', 'segments', 'list'] as const,
    detail: (id: number) => ['crm', 'segments', 'detail', id] as const,
  },

  // Automations
  automations: {
    all: () => ['crm', 'automations'] as const,
    list: () => ['crm', 'automations', 'list'] as const,
    detail: (id: number) => ['crm', 'automations', 'detail', id] as const,
  },

  // Feedback
  feedback: {
    all: () => ['crm', 'feedback'] as const,
    list: (filters?: Record<string, unknown>) => ['crm', 'feedback', 'list', filters] as const,
  },

  // Templates
  templates: {
    all: () => ['crm', 'templates'] as const,
    list: () => ['crm', 'templates', 'list'] as const,
    detail: (id: number) => ['crm', 'templates', 'detail', id] as const,
  },

  // Dashboard
  dashboard: {
    all: () => ['crm', 'dashboard'] as const,
    stats: () => ['crm', 'dashboard', 'stats'] as const,
  },

  // Settings
  settings: {
    all: () => ['crm', 'settings'] as const,
  },
} as const;
