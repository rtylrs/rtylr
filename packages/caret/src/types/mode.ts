export type CaretMode = 'ask' | 'agent' | 'analyst' | 'support';

export interface CaretModeDef {
  id: CaretMode;
  label: string;
  description: string;
  systemPromptAddendum: string;
  allowsTools: boolean;
  defaultRequiresApproval: boolean;
}

export const CARET_MODES: Record<CaretMode, CaretModeDef> = {
  ask: {
    id: 'ask',
    label: 'Ask',
    description: 'Quick answers from live data. Read-only.',
    systemPromptAddendum:
      'You are in Ask mode. Answer questions concisely and use read-only tools when live business data is needed. Do not invoke or propose any tool that modifies data.',
    allowsTools: true,
    defaultRequiresApproval: true,
  },
  agent: {
    id: 'agent',
    label: 'Agent',
    description: 'Plans actions across apps. Changes need approval.',
    systemPromptAddendum:
      'You are in Agent mode. You may use read tools and propose write tools. Summarize the plain-language plan before any mutating tool and request approval before claiming changes.',
    allowsTools: true,
    defaultRequiresApproval: true,
  },
  analyst: {
    id: 'analyst',
    label: 'Analyst',
    description: 'Reports, tables, and KPI breakdowns. Read-only.',
    systemPromptAddendum:
      'You are in Analyst mode. Use read-only business and analytics tools for numbers, trends, and comparisons. Prefer concise tables and bullets, do not describe internal queries, and do not propose write tools.',
    allowsTools: true,
    defaultRequiresApproval: true,
  },
  support: {
    id: 'support',
    label: 'Support',
    description: 'Product navigation and troubleshooting. No tools.',
    systemPromptAddendum:
      'You are in Support mode. Help the user navigate the product and explain features from the current app context. Do not call business tools or mutate data.',
    allowsTools: false,
    defaultRequiresApproval: true,
  },
};

export const DEFAULT_MODE: CaretMode = 'ask';
