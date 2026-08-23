export interface Plan {
  name: string;
  description: string;
  features: string[];
  ctaText: string;
  ctaAction: 'register' | 'register-pro' | 'contact';
  featured?: boolean;
}

/**
 * Tier names and what each tier includes. No prices.
 *
 * There were prices here, and they were placeholders: $6.99 and $4.99
 * per seat per month, a 20% annual discount, and transaction-fee caps
 * in dollars. None of it was final, all of it rendered on the home page
 * and on /pricing/, and the JSON-LD asserted it to Google as fact. A
 * price a customer can read is a commitment. These were not.
 *
 * When real pricing exists it goes back in this file and nowhere else,
 * and the /pricing/ Offer schema is generated from these same objects
 * so the markup cannot disagree with the page.
 */
export const PLANS: Plan[] = [
  {
    name: 'Free',
    description: 'Everything you need to run a single location.',
    features: [
      'Up to 10 team members',
      'All 7 apps included',
      '1 location',
      '1,000 AI credits / month',
      'No credit card required',
    ],
    ctaText: 'Start free',
    ctaAction: 'register',
  },
  {
    name: 'Pro',
    description: 'Growing teams scaling across multiple locations.',
    features: [
      'Up to 30 team members',
      'Up to 5 locations',
      '10,000 AI credits / month, plus 500 per seat',
      'AI assistant',
      'Priority email and chat support',
    ],
    ctaText: 'Start with Pro',
    ctaAction: 'register-pro',
    featured: true,
  },
  {
    name: 'Scale',
    description: 'Unlimited seats and every feature unlocked.',
    features: [
      'Unlimited team members',
      'Unlimited locations',
      '50,000 AI credits / month, plus 1,000 per seat',
      'SSO and audit logs',
      'Dedicated success manager and SLA',
    ],
    ctaText: 'Talk to sales',
    ctaAction: 'contact',
  },
];
