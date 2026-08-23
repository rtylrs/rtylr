import React from 'react';

import styles from './ProductTruthSection.module.scss';

/**
 * Product truth: claims about rtylr that a reader can check without
 * talking to us, plus the things rtylr does not do yet.
 *
 * This sits next to the testimonials on purpose. A quote is a claim the
 * reader has to take on trust; everything below is a claim they can
 * verify from a terminal or a browser tab, which is the only kind of
 * proof this site is allowed to publish while there is no customer data
 * and no benchmark of our own.
 *
 * Every item was read off the build output before it was written here.
 * Two candidates were cut for failing that check:
 *
 *   - "25 integrations." /integrations/ lists 24 and every one of them
 *     renders a "Coming soon" badge. Zero are live. The page is honest;
 *     a claim built on its length would not have been.
 *   - "Public REST API." public/openapi.json describes the marketing
 *     site's own API and has two paths. It is not the product API, so it
 *     is described below as exactly what it is.
 *
 * The limits list is not modesty. A page that wins on every line reads
 * as marketing and gets cited by nobody, and every item in it is a
 * question a buyer will otherwise ask on a call.
 *
 * Rule for editing: if you cannot tell a reader how to check it, it does
 * not belong in the grid. Move it to the copy above, or leave it out.
 */

interface Claim {
  title: string;
  body: string;
  /** How a reader verifies it. Rendered as copyable code. */
  check?: string;
  checkLabel?: string;
}

const CLAIMS: Claim[] = [
  {
    title: 'One database, not seven integrations',
    body: 'POS, ERP, CRM, HR, Finance and Flow read and write the same records. A sale depletes stock and lands in the ledger as it happens, because there is no second system to sync to.',
  },
  {
    title: 'Every page is readable by a machine',
    body: 'Ask any URL on this site for Markdown instead of HTML and you get Markdown, with the correct Vary header. No scraping, no rendering, no JavaScript.',
    checkLabel: 'Check it:',
    check: 'curl -H "Accept: text/markdown" https://rtylr.com/pos/',
  },
  {
    title: 'The whole site is indexed as data',
    body: 'A machine-readable index of all 23 pages, with titles and descriptions, generated at build time. It cannot drift from the site, because the site is what generates it.',
    checkLabel: 'Check it:',
    check: 'https://rtylr.com/pages.json',
  },
  {
    title: 'All 7 apps on every plan',
    body: 'No module upsells and no per-app licence. The plans differ by how many people and locations you run, not by which parts of the product you are allowed to open.',
    checkLabel: 'Check it:',
    check: 'https://rtylr.com/pricing/',
  },
  {
    title: 'The Free plan needs no card',
    body: 'Ten team members at one location, all 7 apps, no payment method and no trial clock. It is a plan, not a countdown.',
    checkLabel: 'Check it:',
    check: 'https://rtylr.com/pricing/',
  },
  {
    title: 'An OpenAPI document, published',
    body: 'This site publishes a spec for its own endpoints as OpenAPI 3.1. It covers the marketing site, not the product API, and it is linked here so nobody has to guess which.',
    checkLabel: 'Check it:',
    check: 'https://rtylr.com/openapi.json',
  },
];

const LIMITS: string[] = [
  'Prices are not published. The Free plan is free; Pro and Scale are quoted.',
  'No integration is live yet. /integrations/ lists what is planned and marks every entry Coming soon.',
  'Payroll filing is not ours. rtylr HR calculates pay and exports files for Gusto, ADP or Rippling to file.',
  'Complex accounting still wants a ledger. Multi-entity consolidation and heavy audit trails are better served by QuickBooks or Xero, and rtylr Finance exports to both.',
  'The apps are in English. Arabic and French are being written, not translated, and are not shipped yet.',
];

const ProductTruthSection = (): React.JSX.Element => (
  <section className={styles.truthSection}>
    <div className={styles.container}>
      <div className={styles.header}>
        <span className={styles.badge}>Checkable</span>
        <h2 className={styles.sectionTitle}>What we can prove.</h2>
        <p className={styles.sectionDescription}>
          Six claims you can verify from a browser tab or a terminal, and five things rtylr does not
          do yet. Nothing here needs a sales call to confirm.
        </p>
      </div>

      <div className={styles.grid}>
        {CLAIMS.map((claim) => (
          <div key={claim.title} className={styles.claim}>
            <h3 className={styles.claimTitle}>{claim.title}</h3>
            <p className={styles.claimBody}>{claim.body}</p>
            {claim.check && (
              <span className={styles.check}>
                {claim.checkLabel} <code className={styles.checkCode}>{claim.check}</code>
              </span>
            )}
          </div>
        ))}
      </div>

      <div className={styles.limits}>
        <h3 className={styles.limitsTitle}>What rtylr does not do yet</h3>
        <ul className={styles.limitsList}>
          {LIMITS.map((limit) => (
            <li key={limit} className={styles.limitsItem}>
              {limit}
            </li>
          ))}
        </ul>
      </div>
    </div>
  </section>
);

export default ProductTruthSection;
