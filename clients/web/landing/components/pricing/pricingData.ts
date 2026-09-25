export interface Plan {
  /** Stable id, and the key segment under pricing.plans in the locale files. */
  id: 'free' | 'pro' | 'scale';
  ctaAction: 'register' | 'register-pro' | 'contact';
  featured?: boolean;
}

/**
 * The tiers, as structure only. Every string moved to the locale files
 * under `pricing.plans.<id>`, because a module-scope constant cannot call
 * t(): the hook exists only inside a component. Importing the i18next
 * instance here instead would resolve once at module load and then never
 * update when the language changes, which is the kind of bug that only
 * shows up after someone clicks the switcher.
 *
 * What is left here is the part that is not copy: the order, which tier is
 * featured, and which action its button takes.
 *
 * There were prices here, and they were placeholders: $6.99 and $4.99 per
 * seat per month, a 20% annual discount, and transaction-fee caps in
 * dollars. None of it was final, all of it rendered on the home page and on
 * /pricing/, and the JSON-LD asserted it to Google as fact. A price a
 * customer can read is a commitment. These were not.
 *
 * When real pricing exists it goes back in this file and nowhere else, and
 * the /pricing/ Offer schema is generated from these same objects so the
 * markup cannot disagree with the page.
 */
export const PLANS: Plan[] = [
  { id: 'free', ctaAction: 'register' },
  { id: 'pro', ctaAction: 'register-pro', featured: true },
  { id: 'scale', ctaAction: 'contact' },
];
