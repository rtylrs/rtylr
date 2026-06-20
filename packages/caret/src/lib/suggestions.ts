import type { CaretPageContext, CaretPageContextValue } from '../types/context';
import type { CaretMode } from '../types/mode';

export interface CaretStarter {
  prompt: string;
  label?: string;
  description?: string;
  mode?: CaretMode;
}

export interface CaretContextSignal {
  label: string;
  value: string;
  tone?: 'neutral' | 'accent' | 'warn';
}

export interface CaretEmptyStateModel {
  title: string;
  subtitle: string;
  contextSignals: CaretContextSignal[];
  primaryAction: CaretStarter;
  suggestions: CaretStarter[];
}

interface BuildCaretEmptyStateInput {
  appId: string;
  pageContext?: CaretPageContext;
}

type StarterInput = string | CaretStarter;

const starter = (
  prompt: string,
  mode: CaretMode = 'ask',
  label?: string,
  description?: string
): CaretStarter => ({
  prompt,
  mode,
  ...(label ? { label } : {}),
  ...(description ? { description } : {}),
});

const GENERIC_STARTERS: CaretStarter[] = [
  starter(
    'Give me a 5-minute owner brief',
    'analyst',
    'Owner brief',
    'The numbers, risks, and next moves.'
  ),
  starter('What should I handle first?', 'ask', 'Handle first', 'A plain priority list.'),
  starter('Find anything that looks off', 'analyst', 'Check risks', 'Spot unusual changes.'),
  starter('Turn this into an action plan', 'agent', 'Action plan', 'Draft the next steps.'),
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

const APP_STARTERS: Record<string, CaretStarter[]> = {
  crm: [
    starter('Which customers need follow-up?', 'ask'),
    starter('Show customers at risk', 'analyst'),
    starter('Who are my best customers this month?', 'analyst'),
    starter('What should sales do next?', 'agent'),
  ],
  dash: [
    starter('What should I handle first today?', 'ask'),
    starter('Which numbers need attention?', 'analyst'),
    starter('Summarize the business right now', 'analyst'),
    starter('What changed since yesterday?', 'analyst'),
  ],
  erp: [
    starter('What needs attention in operations?', 'ask'),
    starter('Which inventory items need action?', 'analyst'),
    starter('Summarize open operational issues', 'analyst'),
    starter('What should I reorder first?', 'agent'),
  ],
  finance: [
    starter('How much revenue today?', 'analyst'),
    starter('What changed in revenue this week?', 'analyst'),
    starter('Which expenses need review?', 'analyst'),
    starter('Summarize cash flow right now', 'analyst'),
  ],
  flow: [
    starter('What work is blocked right now?', 'ask'),
    starter('Summarize this workspace', 'analyst'),
    starter('What should the team handle next?', 'agent'),
    starter('Find overdue tasks', 'analyst'),
  ],
  hr: [
    starter('Who is working today?', 'ask'),
    starter('Which staff issues need attention?', 'analyst'),
    starter('Find missing clock-outs', 'analyst'),
    starter('Summarize team activity', 'analyst'),
  ],
  insights: [
    starter('What should I know from these numbers?', 'analyst'),
    starter('Find the biggest change', 'analyst'),
    starter('Turn this into a quick report', 'analyst'),
    starter('What should I watch next?', 'analyst'),
  ],
  pos: [
    starter('What needs attention right now?', 'ask'),
    starter("Show today's sales and open orders", 'analyst'),
    starter('Which products are low on stock?', 'analyst'),
    starter('Summarize the current shift', 'analyst'),
  ],
};

const MODULE_STARTERS: Record<string, CaretStarter[]> = {
  apps: [
    starter('Which apps should I enable next?', 'agent'),
    starter('What is included in my current setup?', 'ask'),
    starter('Find anything missing from this workspace', 'analyst'),
    starter('What would help this business most?', 'agent'),
  ],
  caret: [
    starter('What can you help me do here?', 'support'),
    starter('Show me what needs attention today', 'ask'),
    starter('Create a short business brief', 'analyst'),
    starter('Remember an important business preference', 'agent'),
  ],
  dashboard: APP_STARTERS.dash,
  settings: [
    starter('Check this setup for anything risky', 'ask'),
    starter('What settings should I review?', 'support'),
    starter('Explain what matters on this page', 'support'),
    starter('Find anything incomplete', 'analyst'),
  ],
  storefront: [
    starter('What should I fix in the storefront?', 'agent'),
    starter('Which menu items need attention?', 'analyst'),
    starter('How can I improve online ordering?', 'agent'),
    starter('Find anything incomplete for customers', 'analyst'),
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
  return entryCount(value) > 0;
}

function entryCount(value: Record<string, CaretPageContextValue> | undefined): number {
  return value ? Object.keys(value).filter((key) => value[key] !== undefined).length : 0;
}

function pluralize(value: number, singular: string, plural = `${singular}s`): string {
  return `${value} ${value === 1 ? singular : plural}`;
}

function truncate(value: string, limit = 42): string {
  return value.length > limit ? `${value.slice(0, limit - 1).trim()}...` : value;
}

function formatContextValue(value: CaretPageContextValue | undefined): string | null {
  if (typeof value === 'string') return truncate(value.trim());
  if (typeof value === 'number') return Number.isFinite(value) ? String(value) : null;
  if (typeof value === 'boolean') return value ? 'yes' : 'no';
  if (Array.isArray(value)) return pluralize(value.length, 'item');
  if (value && typeof value === 'object') {
    const keys = Object.keys(value).filter((key) => value[key] !== undefined);
    return keys.length > 0 ? pluralize(keys.length, 'field') : null;
  }
  return null;
}

function firstSelectedLabel(selected: CaretPageContext['selected']): string | null {
  if (!selected) return null;
  const key = Object.keys(selected)[0];
  if (!key) return null;
  return SELECTED_LABELS[key] ?? humanize(key);
}

function hintStarter(hints: string[] | undefined): CaretStarter | null {
  const hint = cleanText(hints?.[0]);
  return hint ? starter(`What should I check for ${hint}?`) : null;
}

function normalizeStarter(value: StarterInput | null | undefined): CaretStarter | null {
  if (!value) return null;
  if (typeof value === 'string') return starter(value);
  const prompt = cleanText(value.prompt);
  if (!prompt) return null;
  return { ...value, prompt };
}

function unique(values: Array<StarterInput | null | undefined>, limit: number): CaretStarter[] {
  const seen = new Set<string>();
  const result: CaretStarter[] = [];

  for (const value of values) {
    const item = normalizeStarter(value);
    if (!item) continue;
    const key = item.prompt.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    result.push(item);
    if (result.length >= limit) break;
  }

  return result;
}

function buildContextSignals({
  pageContext,
  page,
  selectedLabel,
  visibleRows,
  hasFilters,
}: {
  pageContext?: CaretPageContext;
  page: string | null;
  selectedLabel: string | null;
  visibleRows: number;
  hasFilters: boolean;
}): CaretContextSignal[] {
  const signals: CaretContextSignal[] = [];
  const role = cleanText(pageContext?.role);
  const freshness = formatContextValue(pageContext?.dataFreshness);
  const firstHint = cleanText(pageContext?.hints?.[0]);
  const filterCount = entryCount(pageContext?.filters);
  const memoryCount = entryCount(pageContext?.businessMemory);

  if (isUsefulLabel(page)) signals.push({ label: 'Screen', value: page });
  if (selectedLabel) signals.push({ label: 'Focus', value: selectedLabel, tone: 'accent' });
  if (visibleRows > 0) signals.push({ label: 'Visible', value: pluralize(visibleRows, 'row') });
  if (hasFilters) signals.push({ label: 'Filter', value: `${filterCount || 1} active` });
  if (freshness) signals.push({ label: 'Data', value: freshness });
  if (role) signals.push({ label: 'Role', value: truncate(role, 24) });
  if (memoryCount > 0)
    signals.push({
      label: 'Notes',
      value: pluralize(memoryCount, 'saved item'),
    });
  if (firstHint) signals.push({ label: 'Watch', value: truncate(firstHint), tone: 'warn' });

  return signals.slice(0, 6);
}

function buildPrimaryAction({
  appKey,
  moduleKey,
  selectedLabel,
  visibleRows,
  pageContext,
}: {
  appKey: string;
  moduleKey: string;
  selectedLabel: string | null;
  visibleRows: number;
  pageContext?: CaretPageContext;
}): CaretStarter {
  const firstHint = cleanText(pageContext?.hints?.[0]);

  if (selectedLabel) {
    return starter(
      `Review this ${selectedLabel} and tell me the next best action`,
      'ask',
      `Review this ${selectedLabel}`,
      'Summary, risk check, and next move.'
    );
  }

  if (firstHint) {
    return starter(
      `Check ${firstHint} and tell me what needs action`,
      'ask',
      `Check ${truncate(firstHint, 28)}`,
      'Focus on the thing most likely to matter.'
    );
  }

  if (visibleRows > 0) {
    return starter(
      `Prioritize these ${visibleRows} visible rows and tell me what I should do first`,
      'analyst',
      'Prioritize this list',
      'Rank what matters and explain why.'
    );
  }

  if (moduleKey === 'storefront') {
    return starter(
      'Audit my storefront and tell me what to fix first',
      'agent',
      'Audit storefront',
      'Customer-facing issues and next actions.'
    );
  }

  if (appKey === 'pos') {
    return starter(
      'Summarize the current shift and tell me what needs attention',
      'analyst',
      'Current shift brief',
      'Sales, orders, stock, and watchouts.'
    );
  }

  return starter(
    'Give me a 5-minute owner brief and tell me what to handle first',
    'analyst',
    'Start with a brief',
    'Important numbers, risks, and next moves.'
  );
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
  const primaryAction = buildPrimaryAction({
    appKey,
    moduleKey,
    selectedLabel,
    visibleRows,
    pageContext,
  });

  const baseStarters =
    MODULE_STARTERS[moduleKey] ??
    APP_STARTERS[appKey] ??
    APP_STARTERS[fallbackAppKey] ??
    GENERIC_STARTERS;

  const contextStarters = [
    selectedLabel
      ? starter(
          `Summarize this ${selectedLabel}`,
          'ask',
          `Summarize ${selectedLabel}`,
          'What matters and why.'
        )
      : null,
    selectedLabel
      ? starter(
          `What should I check before changing this ${selectedLabel}?`,
          'agent',
          'Before changing it',
          'Risks and approval context.'
        )
      : null,
    visibleRows > 0
      ? starter(
          `Prioritize these ${visibleRows} visible rows`,
          'analyst',
          'Prioritize visible rows',
          'Rank what needs attention.'
        )
      : null,
    visibleRows > 0
      ? starter(
          'Find anything unusual in this list',
          'analyst',
          'Find unusual items',
          'Look for gaps or surprises.'
        )
      : null,
    hasFilters
      ? starter(
          'What stands out in these filtered results?',
          'analyst',
          'Filtered results',
          'Explain what changed in this view.'
        )
      : null,
  ];

  const suggestions = unique(
    [...contextStarters, ...baseStarters, hintStarter(pageContext?.hints), ...GENERIC_STARTERS],
    6
  )
    .filter((item) => item.prompt.toLowerCase() !== primaryAction.prompt.toLowerCase())
    .slice(0, 4);

  return {
    title: isUsefulLabel(page) ? `Caret is ready on ${page}` : `Caret is ready in ${appName}`,
    subtitle: pageContext?.businessName
      ? `Start with the most useful next move for ${pageContext.businessName}.`
      : 'Start with what matters right now.',
    contextSignals: buildContextSignals({
      pageContext,
      page,
      selectedLabel,
      visibleRows,
      hasFilters,
    }),
    primaryAction,
    suggestions,
  };
}
