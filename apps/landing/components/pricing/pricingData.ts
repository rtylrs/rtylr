export type BillingCycle = 'monthly' | 'annual';

export interface Plan {
  name: string;
  price: { monthly: string; annual: string };
  priceSuffix: string;
  priceNote: { monthly: string; annual: string };
  description: string;
  features: string[];
  ctaText: string;
  ctaAction: 'register' | 'register-pro' | 'contact';
  featured?: boolean;
}

export const PLANS: Plan[] = [
  {
    name: 'Free',
    price: { monthly: '$0', annual: '$0' },
    priceSuffix: 'forever',
    priceNote: {
      monthly: 'No credit card required · cancel anytime',
      annual: 'No credit card required · cancel anytime',
    },
    description: 'Everything you need to run a single location.',
    features: [
      'Up to 10 team members',
      'All 7 apps included',
      '1 location',
      '0% transaction fees under $10k/mo',
      '1,000 AI credits / month',
    ],
    ctaText: 'Start free',
    ctaAction: 'register',
  },
  {
    name: 'Pro',
    price: { monthly: '$6.99', annual: '$4.99' },
    priceSuffix: '/seat/month',
    priceNote: {
      monthly: '$4.99/seat billed annually · up to 30 seats',
      annual: 'Billed annually · $59.88/seat/year',
    },
    description: 'Growing teams scaling across multiple locations.',
    features: [
      'Up to 30 team members',
      'Up to 5 locations',
      '0% transaction fees under $100k/mo',
      '10,000 AI credits / month + 500 per seat',
      'Priority email + chat support',
    ],
    ctaText: 'Start with Pro',
    ctaAction: 'register-pro',
    featured: true,
  },
  {
    name: 'Scale',
    price: { monthly: '$4.99', annual: '$2.99' },
    priceSuffix: '/seat/month',
    priceNote: {
      monthly: '$2.99/seat billed annually · unlocks at 30+ active seats',
      annual: '$2.99/seat billed annually · unlocks at 30+ active seats',
    },
    description: 'Unlimited seats and every feature unlocked.',
    features: [
      'Unlimited team members',
      'Unlimited locations',
      '0% transaction fees (no cap)',
      '50,000 AI credits / month + 1,000 per seat',
      'SSO, audit logs, dedicated success manager',
    ],
    ctaText: 'Talk to sales',
    ctaAction: 'contact',
  },
];
