import React from 'react';
import { cn } from '../../../lib/cn';
import { motion } from 'motion/react';
import { SplitWords } from '../../motion/SplitWords';
import styles from './TestimonialsSection.module.scss';

const testimonials = [
  {
    quote:
      'We used to close the books on the 8th of every month. Now we close them on the 2nd - nothing needs reconciling.',
    author: 'Layla Ahmed',
    role: 'Owner, Byblos Kitchen',
    company: 'Restaurant · 3 branches',
    initials: 'LA',
    avatarColor: '#d6b88e',
  },
  {
    quote:
      "The waiter hit 20 visits and rtylr sent the thank-you email on its own. My manager asked if I set it up. I didn't.",
    author: 'Karim Mansour',
    role: 'General Manager, The Terrace',
    company: 'Restaurant',
    initials: 'KM',
    avatarColor: '#c9a48a',
  },
  {
    quote:
      "The internet went down on a Saturday night with 200 tables. The cashier didn't stop. I didn't expect software to handle that.",
    author: 'Tarek Zein',
    role: 'Owner, Mountain Café',
    company: 'Restaurant',
    initials: 'TZ',
    avatarColor: '#b89a8a',
  },
  {
    quote:
      'Schedules used to live on WhatsApp. Now shifts go out in four taps and payroll runs automatically every Sunday night.',
    author: 'Sara Halabi',
    role: 'Operations Manager, Urban Chain',
    company: 'Retail · 5 branches',
    initials: 'SH',
    avatarColor: '#a89680',
  },
  {
    quote:
      "Switching from the old system saved us roughly $600 per branch per month. And that's before I count the add-ons.",
    author: 'Nadine Farhat',
    role: 'Founder, Fresh Market',
    company: 'Grocery · 4 branches',
    initials: 'NF',
    avatarColor: '#d6b88e',
  },
  {
    quote:
      'The task is linked to the invoice, the customer, the shift. I stopped opening four tabs for four apps - I just open Flow.',
    author: 'Jad Maalof',
    role: 'Manager, Tech Store Pro',
    company: 'Retail',
    initials: 'JM',
    avatarColor: '#c9a48a',
  },
];

const TestimonialsSection: React.FC = () => {
  const duplicatedTestimonials = [...testimonials, ...testimonials];

  return (
    <section className={styles.testimonialsSection}>
      <div className={styles.container}>
        <SplitWords
          as="h2"
          className={styles.sectionTitle}
          text="What operators say about it."
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
                  <p className={styles.quote}>&ldquo;{testimonial.quote}&rdquo;</p>
                  <div className={styles.authorInfo}>
                    <div
                      className={styles.authorImage}
                      style={{ background: testimonial.avatarColor }}
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
                  <p className={styles.quote}>&ldquo;{testimonial.quote}&rdquo;</p>
                  <div className={styles.authorInfo}>
                    <div
                      className={styles.authorImage}
                      style={{ background: testimonial.avatarColor }}
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
