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
    description: 'Conversational answers from your business data. Read-only.',
    systemPromptAddendum:
      'You are in Ask mode. Answer questions concisely. Do not invoke any tools that modify data.',
    allowsTools: true,
    defaultRequiresApproval: true,
  },
  agent: {
    id: 'agent',
    label: 'Agent',
    description: 'Plans and executes multi-step actions across your apps.',
    systemPromptAddendum:
      'You are in Agent mode. You may call tools to perform actions. Always summarize your plan before executing mutating tools and request approval for irreversible actions.',
    allowsTools: true,
    defaultRequiresApproval: true,
  },
  analyst: {
    id: 'analyst',
    label: 'Analyst',
    description: 'Generates reports, charts, and breakdowns of your KPIs.',
    systemPromptAddendum:
      'You are in Analyst mode. Prefer concise tables, bullets, and short prose. Use read-only tools to gather data before answering.',
    allowsTools: true,
    defaultRequiresApproval: true,
  },
  support: {
    id: 'support',
    label: 'Support',
    description: 'Helps the user navigate the product and troubleshoot issues.',
    systemPromptAddendum:
      'You are in Support mode. Help the user find features, explain the product, and link to relevant pages. Do not mutate data without explicit approval.',
    allowsTools: false,
    defaultRequiresApproval: true,
  },
};

export const DEFAULT_MODE: CaretMode = 'ask';
