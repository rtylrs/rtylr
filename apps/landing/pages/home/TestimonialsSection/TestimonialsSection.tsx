import React from 'react';
import { motion } from 'motion/react';
import { fadeInUp, viewportOnce } from '../../../lib/animations';
import styles from './TestimonialsSection.module.scss';

const testimonials = [
  {
    quote:
      'Rtylr transformed how we manage our restaurant. The inventory tracking alone saved us hours every week.',
    author: 'Sarah Johnson',
    role: 'Owner, The Green Table',
    company: 'Restaurant',
    image: 'https://placehold.co/80x80/60a5fa/ffffff?text=SJ',
  },
  {
    quote:
      'Switching to Rtylr was the best decision we made. Our sales increased by 30% in the first quarter.',
    author: 'Michael Chen',
    role: 'Manager, Urban Retail Co.',
    company: 'Retail',
    image: 'https://placehold.co/80x80/a78bfa/ffffff?text=MC',
  },
  {
    quote:
      "The analytics features give us insights we never had before. It's like having a business consultant built-in.",
    author: 'Emily Rodriguez',
    role: 'CEO, Fresh Market',
    company: 'Grocery',
    image: 'https://placehold.co/80x80/f0f9ff/101010?text=ER',
  },
  {
    quote:
      'The offline mode is a game-changer. We never lose a sale, even when the internet goes down.',
    author: 'David Park',
    role: 'Founder, Mountain Cafe',
    company: 'Restaurant',
    image: 'https://placehold.co/80x80/60a5fa/ffffff?text=DP',
  },
  {
    quote:
      'Customer support is incredible. They helped us set up everything in one day and we were ready to go.',
    author: 'Lisa Thompson',
    role: 'Operations Manager, City Bakery',
    company: 'Food Service',
    image: 'https://placehold.co/80x80/a78bfa/ffffff?text=LT',
  },
  {
    quote:
      'The eCommerce integration is seamless. Our online orders sync perfectly with our in-store system.',
    author: 'James Wilson',
    role: 'Owner, Tech Store Pro',
    company: 'Retail',
    image: 'https://placehold.co/80x80/f0f9ff/101010?text=JW',
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
          What Our Customers Say
        </motion.h2>
        <motion.div
          className={styles.tickersWrapper}
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, margin: '-50px' }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          <div className={styles.tickerContainer}>
            <div className={`${styles.tickerTrack} ${styles.tickerLeft}`}>
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
                      <div className={styles.authorName}>
                        {testimonial.author}
                      </div>
                      <div className={styles.authorRole}>
                        {testimonial.role}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div className={styles.tickerContainer}>
            <div className={`${styles.tickerTrack} ${styles.tickerRight}`}>
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
                      <div className={styles.authorName}>
                        {testimonial.author}
                      </div>
                      <div className={styles.authorRole}>
                        {testimonial.role}
                      </div>
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
