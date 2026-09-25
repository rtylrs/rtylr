import React from 'react';

import styles from './DirectAnswer.module.scss';

interface DirectAnswerProps {
  /** Small uppercase label above the paragraph. Optional. */
  label?: string;
  /**
   * 40 to 80 words, plain, stating what the thing is and who it is for.
   * One paragraph. If it needs two, it is not a direct answer.
   */
  children: React.ReactNode;
}

/**
 * The direct-answer opening.
 *
 * Every commercial page here already had a strong hero, and none of them
 * said what the product was. "The register that keeps up" is good copy
 * and useless to an extractor: nothing in it resolves to "rtylr POS is X
 * for Y". Featured snippets and AI answers are built by lifting a
 * self-contained paragraph that makes a plain claim, and a page with no
 * such paragraph gets summarised from whatever fragment is nearest.
 *
 * So this sits under the hero rather than replacing it. The hero keeps
 * the voice, this carries the definition, and no static prose was
 * sacrificed to get it.
 *
 * Keep it a single paragraph of real prose. Do not turn it into a
 * bulleted feature list: the extraction behaviour this exists for
 * depends on one continuous block of text.
 */
const DirectAnswer = ({ label, children }: DirectAnswerProps): React.JSX.Element => (
  <section className={styles.directAnswer}>
    <div className={styles.inner}>
      {label && <span className={styles.label}>{label}</span>}
      <p className={styles.answer}>{children}</p>
    </div>
  </section>
);

export default DirectAnswer;
