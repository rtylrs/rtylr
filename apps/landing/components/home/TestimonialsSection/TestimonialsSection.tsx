import React from 'react';
import { cn } from '../../../lib/cn';
import { motion } from 'motion/react';
import { fadeInUp, viewportOnce } from '../../../lib/animations';
import styles from './TestimonialsSection.module.scss';

const testimonials = [
  {
    quote:
      'We used to close the books on the 8th of every month. Now we close them on the 2nd, because there\u2019s nothing left to reconcile.',
    author: 'Laila Ahmed',
    role: 'Owner, Byblos Kitchen',
    company: 'Restaurant · 3 locations',
    image: 'https://placehold.co/80x80/d6b88e/1a1612?text=LA',
  },
  {
    quote:
      'A server hit 20 lifetime visits and rtylr drafted the thank-you email on its own. My manager asked if I\u2019d set that up. I hadn\u2019t.',
    author: 'Marcus Hale',
    role: 'GM, Terrace',
    company: 'Restaurant',
    image: 'https://placehold.co/80x80/c9a48a/1a1612?text=MH',
  },
  {
    quote:
      'Lost WiFi on a Saturday night. 200 covers. The POS kept going. I didn\u2019t know software could do that.',
    author: 'David Park',
    role: 'Owner, Mountain Cafe',
    company: 'Restaurant',
    image: 'https://placehold.co/80x80/b89a8a/1a1612?text=DP',
  },
  {
    quote:
      'Scheduling used to live in a WhatsApp group. Now shifts publish in four clicks and payroll reconciles on its own Sunday night.',
    author: 'Juno Park',
    role: 'Ops Lead, Urban Retail Co.',
    company: 'Retail · 5 stores',
    image: 'https://placehold.co/80x80/a89680/1a1612?text=JP',
  },
  {
    quote:
      'Moving off Toast saved us roughly $340 per location every month. And that\u2019s before I counted the module add-ons we were paying for.',
    author: 'Sarah Johnson',
    role: 'Founder, Fresh Market',
    company: 'Grocery · 4 stores',
    image: 'https://placehold.co/80x80/d6b88e/1a1612?text=SJ',
  },
  {
    quote:
      'Tasks link to the invoice, the customer, the shift. I stopped opening four tabs for four tools, I just open Flow.',
    author: 'James Wilson',
    role: 'Manager, Tech Store Pro',
    company: 'Retail',
    image: 'https://placehold.co/80x80/c9a48a/1a1612?text=JW',
  },
];

const TestimonialsSection: React.FC = () => {
  const duplicatedTestimonials = [...testimonials, ...testimonials];

  return (
    <section className={styles.testimonialsSection}>
      <div className={styles.container}>
        <motion.h2
          className={styles.sectionTitle}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          What operators say about it.
        </motion.h2>
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
                  <p className={styles.quote}>"{testimonial.quote}"</p>
                  <div className={styles.authorInfo}>
                    <img
                      src={testimonial.image}
                      alt={testimonial.author}
                      className={styles.authorImage}
                    />
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
                  <p className={styles.quote}>"{testimonial.quote}"</p>
                  <div className={styles.authorInfo}>
                    <img
                      src={testimonial.image}
                      alt={testimonial.author}
                      className={styles.authorImage}
                    />
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
