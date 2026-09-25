import React from 'react';
import { ChevronDown } from 'lucide-react';

import type { FAQEntry } from '../../lib/structured-data';
import styles from './FaqSection.module.scss';

interface FaqSectionProps {
  /** Section heading. Keep it a question or a plain label, never "FAQ". */
  title: string;
  /** One or two sentences of context. Optional. */
  intro?: string;
  faqs: FAQEntry[];
  /** Heading level for the section title. Defaults to h2. */
  headingLevel?: 'h2' | 'h3';
}

/**
 * Visible FAQ. Two deliberate choices, both learned from what this
 * section replaced.
 *
 * **Native `<details>`, not a React accordion.** The /pricing/ accordion
 * used AnimatePresence, which mounts the answer panel only while it is
 * open. The consequence was invisible and total: the answers reached no
 * crawler, no markdown twin, and no reader without JavaScript. The only
 * copy of that text in the served bytes was inside the FAQPage JSON-LD,
 * so the markup Google stopped rewarding was also the only thing
 * delivering the content. `<details>` keeps every answer in the DOM,
 * collapsed by CSS, and costs no JavaScript at all.
 *
 * **A real heading inside `<summary>`.** The HTML spec allows a single
 * heading element as summary content, so each question appears in the
 * document outline. That is what "question-shaped H2s" is supposed to
 * buy: headings a reader scans and an extractor can match against the
 * words someone actually typed.
 *
 * No FAQPage schema accompanies this. Google removed the FAQ rich result
 * in May 2026 and deleted the documentation in June, so the markup earns
 * nothing. The visible content is the point.
 */
const FaqSection = ({
  title,
  intro,
  faqs,
  headingLevel = 'h2',
}: FaqSectionProps): React.JSX.Element => {
  const Heading = headingLevel;
  const QuestionHeading = headingLevel === 'h2' ? 'h3' : 'h4';

  return (
    <section className={styles.faqSection}>
      <Heading className={styles.sectionTitle}>{title}</Heading>
      {intro && <p className={styles.sectionIntro}>{intro}</p>}

      <div className={styles.faqList}>
        {faqs.map((faq) => (
          <details key={faq.question} className={styles.faqItem}>
            <summary className={styles.faqSummary}>
              <QuestionHeading className={styles.faqQuestion}>{faq.question}</QuestionHeading>
              <ChevronDown size={18} className={styles.faqChevron} aria-hidden="true" />
            </summary>
            <p className={styles.faqAnswer}>{faq.answer}</p>
          </details>
        ))}
      </div>
    </section>
  );
};

export default FaqSection;
