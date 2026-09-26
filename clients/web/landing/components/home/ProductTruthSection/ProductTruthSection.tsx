import React from 'react';
import { useTranslation } from 'react-i18next';

import styles from './ProductTruthSection.module.scss';

/**
 * Product truth: claims about rtylr that a reader can check without
 * talking to us.
 *
 * Everything below is a claim a reader can verify from a terminal or a
 * browser tab, which is the only kind of proof this site publishes while
 * there is no customer data and no benchmark of our own. The invented
 * testimonials that used to sit beside it were removed for the same reason
 * (rtylr#1357).
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
 * Rule for editing: if you cannot tell a reader how to check it, it does
 * not belong in the grid. Move it to the copy above, or leave it out.
 */

interface Claim {
  title: string;
  body: string;
  /** How a reader verifies it. Rendered as copyable code. */
  check?: string;
}

const ProductTruthSection = (): React.JSX.Element => {
  const { t } = useTranslation();
  const claims = t('home.truth.claims', { returnObjects: true }) as unknown as Claim[];

  return (
    <section className={styles.truthSection}>
      <div className={styles.container}>
        <div className={styles.header}>
          <span className={styles.badge}>{t('home.truth.badge')}</span>
          <h2 className={styles.sectionTitle}>{t('home.truth.title')}</h2>
          <p className={styles.sectionDescription}>{t('home.truth.description')}</p>
        </div>

        <div className={styles.grid}>
          {claims.map((claim) => (
            <div key={claim.title} className={styles.claim}>
              <h3 className={styles.claimTitle}>{claim.title}</h3>
              <p className={styles.claimBody}>{claim.body}</p>
              {claim.check && (
                <span className={styles.check}>
                  {t('home.truth.checkLabel')}{' '}
                  {/* dir="ltr" is load-bearing under dir="rtl": a curl command or URL is
                      not text with an inherent reading direction, and letting it inherit
                      rtl scrambles the visual order of / and : while the DOM order stays
                      correct - copying it still works, reading it does not. */}
                  <code className={styles.checkCode} dir="ltr">
                    {claim.check}
                  </code>
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default ProductTruthSection;
