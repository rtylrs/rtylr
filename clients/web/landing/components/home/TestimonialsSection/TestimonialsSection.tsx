import React from 'react';
import { cn } from '../../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { SplitWords } from '../../motion/SplitWords';
import styles from './TestimonialsSection.module.scss';

// Initials and avatar colour are presentational, not language-dependent,
// so they stay here rather than in the locale files - indexed to line up
// with home.testimonials.items.
const AVATARS = [
  { initials: 'LA', color: '#d6b88e' },
  { initials: 'KM', color: '#c9a48a' },
  { initials: 'TZ', color: '#b89a8a' },
  { initials: 'SH', color: '#a89680' },
  { initials: 'NF', color: '#d6b88e' },
  { initials: 'JM', color: '#c9a48a' },
];

// The quote mark is typography, not copy, so it does not belong in the
// locale files - but a hardcoded &ldquo;/&rdquo; is still an English quote
// mark left on the Arabic and French pages. French and Arabic content in
// this codebase already uses guillemets elsewhere (fr.json's non-breaking-
// space convention, ar.json's «» around quoted questions); this matches
// that rather than inventing a third convention here.
const QUOTE_MARKS: Record<string, [string, string]> = {
  // Non-breaking space inside the guillemets, matching fr.json's own
  // convention, so a reflow can never strand « or » on its own line.
  fr: ['« ', ' »'],
  ar: ['«', '»'],
};
const DEFAULT_QUOTE_MARKS: [string, string] = ['“', '”'];

const TestimonialsSection: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [openQuote, closeQuote] = QUOTE_MARKS[i18n.language] ?? DEFAULT_QUOTE_MARKS;
  const testimonials = (
    t('home.testimonials.items', { returnObjects: true }) as unknown as {
      quote: string;
      author: string;
      role: string;
    }[]
  ).map((item, i) => ({ ...item, ...AVATARS[i] }));
  const duplicatedTestimonials = [...testimonials, ...testimonials];

  return (
    <section className={styles.testimonialsSection}>
      <div className={styles.container}>
        <SplitWords
          as="h2"
          className={styles.sectionTitle}
          text={t('home.testimonials.title')}
          trigger="view"
        />
        <motion.div
          className={styles.tickersWrapper}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, margin: '-50px' }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className={styles.tickerContainer}>
            <div className={cn(styles.tickerTrack, styles.tickerLeft)}>
              {duplicatedTestimonials.map((testimonial, index) => (
                <div key={index} className={styles.testimonialCard}>
                  <p className={styles.quote}>{openQuote}{testimonial.quote}{closeQuote}</p>
                  <div className={styles.authorInfo}>
                    <div
                      className={styles.authorImage}
                      style={{ background: testimonial.color }}
                      aria-hidden="true"
                    >
                      {testimonial.initials}
                    </div>
                    <div className={styles.authorDetails}>
                      <div className={styles.authorName}>{testimonial.author}</div>
                      <div className={styles.authorRole}>{testimonial.role}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className={styles.tickerContainer}>
            <div className={cn(styles.tickerTrack, styles.tickerRight)}>
              {duplicatedTestimonials.map((testimonial, index) => (
                <div key={index} className={styles.testimonialCard}>
                  <p className={styles.quote}>{openQuote}{testimonial.quote}{closeQuote}</p>
                  <div className={styles.authorInfo}>
                    <div
                      className={styles.authorImage}
                      style={{ background: testimonial.color }}
                      aria-hidden="true"
                    >
                      {testimonial.initials}
                    </div>
                    <div className={styles.authorDetails}>
                      <div className={styles.authorName}>{testimonial.author}</div>
                      <div className={styles.authorRole}>{testimonial.role}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default TestimonialsSection;
