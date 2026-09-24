import { useTranslation } from 'react-i18next';

import type { FAQEntry } from './structured-data';

/**
 * Visible FAQ content, resolved from the locale files.
 *
 * The copy used to live in this file as seven exported arrays. It moved to
 * `locales/*.json` under `faq.<section>` when the site became
 * translatable, because a module-scope constant cannot call `t()` and
 * importing the i18next instance here would resolve once at module load
 * and then never update on a language change.
 *
 * What stays here is the part that is not copy: the editorial rules. They
 * are in a .ts file rather than beside the strings because JSON has no
 * comments, and every one of them was broken at some point by content that
 * looked perfectly reasonable.
 *
 * ## Rules for anything written into faq.<section>.items
 *
 * 1. **It must render on the page.** These answers exist to be read. Every
 *    array here was once imported solely to build `FAQPage` JSON-LD, with
 *    nothing rendering the questions, which is markup describing content
 *    the page does not have. There is no FAQPage markup any more: Google
 *    removed the rich result in May 2026, and by then the schema had become
 *    the only copy of these answers in the served HTML. Render them, via
 *    `components/FaqSection`.
 * 2. **No prices.** rtylr publishes none. Tier names and what a tier
 *    includes are fine; a rate, a discount or a currency amount is not.
 *    These answers once claimed per-seat daily proration and a free tier at
 *    3 seats, when the plan data said 10 and no price was published
 *    anywhere.
 * 3. **No uncited competitor claim.** "Toast is POS-first and everything
 *    else is bolted on" needs a dated citation to Toast's own
 *    documentation. Until a comparison page carries one, answer with what
 *    rtylr does and let the reader draw the comparison. Competitors read
 *    these pages and they complain to Google.
 * 4. **No em dashes, and no hyphen standing in for one.** Restructure, or
 *    use a colon.
 * 5. **Roughly 40 to 80 words.** That is the length that survives snippet
 *    extraction intact.
 * 6. **Ask it the way a buyer types it.** "how much does X cost" beats "our
 *    pricing explained".
 *
 * The keys live under one `faq` namespace rather than on each page's own
 * namespace so that `useFaq` can build them with a literal `faq.` prefix.
 * A key assembled with no static prefix cannot be checked for orphans
 * without exempting every key in the file from the check.
 *
 * Adding an entry means adding it to `en.json` only. `ar.json` and
 * `fr.json` fill in when the copy is written, not translated, and reviewed.
 * `tests/unit/i18n-keys.test.mjs` holds the structural half of this
 * contract; the rules above are the half a test cannot check.
 */

/** Sections that carry their own FAQ. Matches the locale namespaces. */
export type FaqSection =
  | 'home'
  | 'pricing'
  | 'pos'
  | 'erp'
  | 'crm'
  | 'hr'
  | 'finance'
  | 'flow'
  | 'demo';

interface Faq {
  /** Section heading for the block. */
  title: string;
  /** Optional sentence under the heading. Only the home page has one. */
  intro?: string;
  /** Question and answer pairs, in render order. */
  items: FAQEntry[];
}

/**
 * Resolve one section's FAQ. Returns the heading and the entries together,
 * so a caller cannot render a heading with somebody else's questions under
 * it.
 */
export function useFaq(section: FaqSection): Faq {
  const { t } = useTranslation();
  // i18next returns the key itself when a key is missing, so an absent
  // intro comes back as the string "home.faq.intro" rather than undefined.
  // Compare against the key to tell "no intro" from "an intro".
  const introKey = `faq.${section}.intro`;
  const intro = t(introKey);

  return {
    title: t(`faq.${section}.title`),
    intro: intro === introKey ? undefined : intro,
    items: t(`faq.${section}.items`, { returnObjects: true }) as unknown as FAQEntry[],
  };
}
