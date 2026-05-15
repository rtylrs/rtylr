/**
 * Per-page suggestion lists shown in the Caret empty state. Keeps
 * dashboard prompts focused on what the user is most likely to ask
 * given the page they're on.
 */

const SUGGESTIONS: Record<string, string[]> = {
  '/': [
    'How much revenue today?',
    'Which products are low on stock?',
    'Show my top 5 customers',
    'How many staff are clocked in?',
  ],
  '/team': [
    'Who joined the team this week?',
    'List inactive members',
    'How many seats am I using?',
    "Show me Sara's permissions",
  ],
  '/apps': [
    'Which apps do I have enabled?',
    'What does the CRM app do?',
    'How do I add the Finance app?',
  ],
  '/menus': [
    'What is my best-selling item this week?',
    'Add a new pizza for $14.99',
    'Mark all sushi items as out of stock',
  ],
  '/orders': ['How many orders today?', 'Show pending refunds', 'Average order value this month'],
};

const FALLBACK = SUGGESTIONS['/'];

export function suggestionsFor(pathname: string): string[] {
  if (SUGGESTIONS[pathname]) return SUGGESTIONS[pathname];
  const match = Object.keys(SUGGESTIONS).find((p) => pathname.startsWith(p) && p !== '/');
  return match ? SUGGESTIONS[match] : FALLBACK;
}
