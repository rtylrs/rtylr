import React, { useState, useEffect, useRef } from 'react';
import { motion } from 'motion/react';
import { Button } from '@voxire/components';
import { useRouter } from 'next/router';
import { fadeInUp, viewportOnce } from '../../../lib/animations';
import styles from './FeaturesSection.module.scss';

interface Feature {
  id: string;
  category: string;
  title: string;
  description: string;
  ctaText: string;
  image: string;
}

const features: Feature[] = [
  {
    id: 'pos',
    category: 'POS',
    title: 'A register that keeps up with real service.',
    description:
      'Tickets, tables, splits, tips, and every payment type, card, Apple Pay, cash, gift, comp. Works offline when the WiFi doesn\u2019t. < 200ms per checkout.',
    ctaText: 'OPEN POS',
    image: '/images/features/pos.jpg',
  },
  {
    id: 'erp',
    category: 'ERP',
    title: 'Inventory that counts itself.',
    description:
      'Recipe-linked depletion as sales ring up. Auto-reorder at threshold. Multi-location transfers with a full audit trail. Your shelves tell the database. Not the other way around.',
    ctaText: 'OPEN ERP',
    image: '/images/features/erp.jpg',
  },
  {
    id: 'hr',
    category: 'HR',
    title: 'The people side, without the paperwork.',
    description:
      'Drag-to-publish schedules. Clock-in from the POS. Payroll that rolls up tips, hours, and deductions automatically. Compliance flags before they become violations.',
    ctaText: 'OPEN HR',
    image: '/images/features/hr.jpg',
  },
  {
    id: 'crm',
    category: 'CRM',
    title: 'Every customer, remembered.',
    description:
      'Every POS sale enriches the profile. Segment by spend, visit cadence, birthday, tags. Send the campaign from the same app. No Mailchimp detour required.',
    ctaText: 'OPEN CRM',
    image: '/images/features/crm.jpg',
  },
  {
    id: 'finance',
    category: 'Finance',
    title: 'A real P&L, in real time.',
    description:
      'Recurring invoices that collect themselves. Expenses captured at source. Sales tax by jurisdiction. A P&L that updates with every sale, not at month-end.',
    ctaText: 'OPEN FINANCE',
    image: '/images/features/finance.jpg',
  },
  {
    id: 'flow',
    category: 'Flow',
    title: 'Work that moves between apps.',
    description:
      'Kanban boards, automations, and tasks linked to the invoice, customer, or shift they\u2019re actually about. Operations that don\u2019t slip between tools.',
    ctaText: 'OPEN FLOW',
    image: '/images/features/flow.jpg',
  },
  {
    id: 'insights',
    category: 'Insights',
    title: 'The dashboard that sees everything.',
    description:
      'Revenue trends, customer cohorts, product performance, labor margin. Every number pulled live from every other app. No SQL, no nightly ETL.',
    ctaText: 'OPEN INSIGHTS',
    image: '/images/features/insights.jpg',
  },
];

const FeaturesSection: React.FC = () => {
  const [activeFeature, setActiveFeature] = useState<string>(features[0].id);
  const sectionRefs = useRef<{ [key: string]: HTMLDivElement | null }>({});
  const router = useRouter();

  useEffect(() => {
    const observerOptions = {
      root: null,
      rootMargin: '-20% 0px -60% 0px',
      threshold: 0,
    };

    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          const featureId = entry.target.getAttribute('data-feature-id');
          if (featureId) {
            setActiveFeature(featureId);
          }
        }
      });
    }, observerOptions);

    features.forEach((feature) => {
      const element = sectionRefs.current[feature.id];
      if (element) {
        observer.observe(element);
      }
    });

    return () => {
      features.forEach((feature) => {
        const element = sectionRefs.current[feature.id];
        if (element) {
          observer.unobserve(element);
        }
      });
    };
  }, []);

  const activeFeatureData = features.find((f) => f.id === activeFeature) || features[0];

  return (
    <motion.section
      className={styles.featuresSection}
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true, margin: '-100px' }}
      transition={{ duration: 0.6 }}
    >
      <div className={styles.container}>
        <div className={styles.content}>
          <div className={styles.featuresList}>
            {features.map((feature, index) => (
              <React.Fragment key={feature.id}>
                {index > 0 && <div className={styles.featureDivider} />}
                <motion.div
                  ref={(el) => {
                    sectionRefs.current[feature.id] = el;
                  }}
                  data-feature-id={feature.id}
                  className={`${styles.featureItem} ${
                    activeFeature === feature.id ? styles.active : ''
                  }`}
                  initial={{ opacity: 0, x: -30 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true, margin: '-50px' }}
                  transition={{
                    delay: index * 0.1,
                    duration: 0.5,
                    ease: [0.25, 0.46, 0.45, 0.94],
                  }}
                >
                  <span className={styles.featureCategory}>{feature.category}</span>
                  <h3 className={styles.featureTitle}>{feature.title}</h3>
                  <p className={styles.featureDescription}>{feature.description}</p>
                  <div className={styles.featureActions}>
                    <Button
                      variant="dark-outline"
                      onClick={() => router.push('/demo')}
                      className={styles.featureButton}
                    >
                      {feature.ctaText}
                    </Button>
                  </div>
                </motion.div>
              </React.Fragment>
            ))}
          </div>
        </div>

        <motion.div
          className={styles.showcase}
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true, margin: '-100px' }}
          transition={{ duration: 0.6, ease: 'easeOut' }}
        >
          <div className={styles.showcaseImageContainer}>
            {features.map((feature) => (
              <div
                key={feature.id}
                className={`${styles.showcaseImage} ${
                  activeFeature === feature.id ? styles.active : ''
                }`}
              >
                <img src={feature.image} alt={feature.title} className={styles.showcaseImg} />
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </motion.section>
  );
};

export default FeaturesSection;
