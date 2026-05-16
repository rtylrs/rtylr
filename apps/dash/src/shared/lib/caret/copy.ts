/** Reusable user-facing strings for the Caret UI. */
export const CARET_COPY = {
  title: 'Caret',
  tagline: 'Your AI command center for rtylr.',
  empty: {
    title: 'Ask Caret anything',
    subtitle: 'Your AI assistant with access to all your business data.',
  },
  composer: {
    placeholder: 'Ask Caret anything…',
    placeholderAgent: 'Tell Caret what to do…',
    placeholderAnalyst: 'Describe the report you need…',
  },
  approval: {
    title: 'Caret wants to run a tool',
    approve: 'Approve',
    deny: 'Deny',
  },
  errors: {
    network: 'Lost connection. Check your internet and try again.',
    rateLimit: 'Too many requests. Try again in a minute.',
    server: 'Caret had a hiccup. Try again in a moment.',
  },
} as const;
