import type { CaretPageContext, CaretPageContextValue } from '../types/context';

export interface CaretEmptyStateModel {
  title: string;
  subtitle: string;
  suggestions: string[];
}

interface BuildCaretEmptyStateInput {
  appId: string;
  pageContext?: CaretPageContext;
}

const GENERIC_STARTERS = [
  'What should I handle first?',
  'Summarize what matters right now',
  'Find anything that looks off',
  'What changed since yesterday?',
];

const APP_LABELS: Record<string, string> = {
  crm: 'CRM',
  dash: 'Dash',
  erp: 'ERP',
  finance: 'Finance',
  flow: 'Flow',
  hr: 'HR',
  insights: 'Insights',
  pos: 'POS',
};

const APP_STARTERS: Record<string, string[]> = {
  crm: [
    'Which customers need follow-up?',
    'Show customers at risk',
    'Who are my best customers this month?',
    'What should sales do next?',
  ],
  dash: [
    'What should I handle first today?',
    'Which numbers need attention?',
    'Summarize the business right now',
    'What changed since yesterday?',
  ],
  erp: [
    'What needs attention in operations?',
    'Which inventory items need action?',
    'Summarize open operational issues',
    'What should I reorder first?',
  ],
  finance: [
    'How much revenue today?',
    'What changed in revenue this week?',
    'Which expenses need review?',
    'Summarize cash flow right now',
  ],
  flow: [
    'What work is blocked right now?',
    'Summarize this workspace',
    'What should the team handle next?',
    'Find overdue tasks',
  ],
  hr: [
    'Who is working today?',
    'Which staff issues need attention?',
    'Find missing clock-outs',
    'Summarize team activity',
  ],
  insights: [
    'What should I know from these numbers?',
    'Find the biggest change',
    'Turn this into a quick report',
    'What should I watch next?',
  ],
  pos: [
    'What needs attention right now?',
    "Show today's sales and open orders",
    'Which products are low on stock?',
    'Summarize the current shift',
  ],
};

const MODULE_STARTERS: Record<string, string[]> = {
  apps: [
    'Which apps should I enable next?',
    'What is included in my current setup?',
    'Find anything missing from this workspace',
    'What would help this business most?',
  ],
  caret: [
    'What can you help me do here?',
    'Show me what needs attention today',
    'Create a short business brief',
    'Remember an important business preference',
  ],
  dashboard: APP_STARTERS.dash,
  settings: [
    'Check this setup for anything risky',
    'What settings should I review?',
    'Explain what matters on this page',
    'Find anything incomplete',
  ],
  storefront: [
    'What should I fix in the storefront?',
    'Which menu items need attention?',
    'How can I improve online ordering?',
    'Find anything incomplete for customers',
  ],
  team: APP_STARTERS.hr,
};

const SELECTED_LABELS: Record<string, string> = {
  customer_id: 'customer',
  employee_id: 'team member',
  location_id: 'location',
  menu_id: 'menu',
  order_id: 'order',
  product_id: 'product',
};

function cleanText(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}

function isUsefulLabel(value: string | null): value is string {
  return Boolean(value && !/^unknown\b/i.test(value));
}

function normalizeKey(value: unknown): string {
  return (
    cleanText(value)
      ?.toLowerCase()
      .replace(/[^a-z0-9]+/g, '-') ?? ''
  );
}

function humanize(value: string): string {
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\bid\b/gi, '')
    .replace(/\s+/g, ' ')
    .trim()
    .toLowerCase();
}

function hasEntries(value: Record<string, CaretPageContextValue> | undefined): boolean {
  return Boolean(value && Object.keys(value).length > 0);
}

function firstSelectedLabel(selected: CaretPageContext['selected']): string | null {
  if (!selected) return null;
  const key = Object.keys(selected)[0];
  if (!key) return null;
  return SELECTED_LABELS[key] ?? humanize(key);
}

function hintStarter(hints: string[] | undefined): string | null {
  const hint = cleanText(hints?.[0]);
  return hint ? `What should I check for ${hint}?` : null;
}

function unique(values: Array<string | null | undefined>, limit: number): string[] {
  const seen = new Set<string>();
  const result: string[] = [];

  for (const value of values) {
    const text = cleanText(value);
    if (!text) continue;
    const key = text.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(text);
    if (result.length >= limit) break;
  }

  return result;
}

export function buildCaretEmptyState({
  appId,
  pageContext,
}: BuildCaretEmptyStateInput): CaretEmptyStateModel {
  const appKey = normalizeKey(pageContext?.app ?? appId);
  const fallbackAppKey = normalizeKey(appId);
  const moduleKey = normalizeKey(pageContext?.module);
  const page = cleanText(pageContext?.page);
  const appName = cleanText(pageContext?.appName) ?? APP_LABELS[appKey] ?? 'rtylr';
  const selectedLabel = firstSelectedLabel(pageContext?.selected);
  const visibleRows = Array.isArray(pageContext?.visibleRows) ? pageContext.visibleRows.length : 0;
  const hasFilters = hasEntries(pageContext?.filters);

  const baseStarters =
    MODULE_STARTERS[moduleKey] ??
    APP_STARTERS[appKey] ??
    APP_STARTERS[fallbackAppKey] ??
    GENERIC_STARTERS;

  const contextStarters = [
    selectedLabel ? `Summarize this ${selectedLabel}` : null,
    selectedLabel ? `What should I check before changing this ${selectedLabel}?` : null,
    visibleRows > 0 ? 'Summarize the visible rows' : null,
    visibleRows > 0 ? 'Find anything unusual in this list' : null,
    hasFilters ? 'What stands out in these filtered results?' : null,
  ];

  return {
    title: isUsefulLabel(page) ? `Ask Caret about ${page}` : `Ask Caret in ${appName}`,
    subtitle: pageContext?.businessName
      ? `Start with what matters for ${pageContext.businessName}.`
      : 'Start with what matters right now.',
    suggestions: unique(
      [...contextStarters, ...baseStarters, hintStarter(pageContext?.hints), ...GENERIC_STARTERS],
      4
    ),
  };
}
