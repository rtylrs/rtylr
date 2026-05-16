/**
 * Curated example prompts shown in the Caret help/landing screen. Each one
 * is meant to be both representative and immediately useful.
 */
export interface ExamplePrompt {
  category: 'kpi' | 'crm' | 'inventory' | 'team' | 'orders' | 'finance';
  prompt: string;
}

export const EXAMPLE_PROMPTS: ExamplePrompt[] = [
  { category: 'kpi', prompt: 'What is my revenue this week vs last week?' },
  { category: 'kpi', prompt: 'Which day of the week is my busiest?' },
  { category: 'crm', prompt: 'Find customers who spent over $500 last month.' },
  { category: 'crm', prompt: 'Show repeat customers from the last 30 days.' },
  { category: 'inventory', prompt: 'Which products are below their reorder threshold?' },
  { category: 'inventory', prompt: 'Set the price of all medium pizzas to $14.99.' },
  { category: 'team', prompt: 'List staff hours worked this week.' },
  { category: 'team', prompt: 'Invite a new manager called Sara.' },
  { category: 'orders', prompt: 'Show me todays pending refunds.' },
  { category: 'orders', prompt: 'Cancel order #4821 and send the customer an email.' },
  { category: 'finance', prompt: 'Summarize unpaid invoices over 30 days old.' },
];
