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
    title: 'Complete Point of Sale & Payments',
    description:
      'Floor and counter sales, table management, and kitchen display, all in one. Accept every payment type and keep the floor moving with a POS built for speed.',
    ctaText: 'EXPLORE POS',
    image: '/images/features/pos.jpg',
  },
  {
    id: 'erp',
    category: 'ERP',
    title: 'Inventory, Suppliers & Purchasing',
    description:
      'Take full control of your supply chain. Manage stock levels, raise purchase orders, track suppliers, and monitor cost of goods, all synced in real time.',
    ctaText: 'EXPLORE ERP',
    image: '/images/features/erp.jpg',
  },
  {
    id: 'hr',
    category: 'HR',
    title: 'Staff, Shifts & Payroll',
    description:
      'Schedule shifts, track timesheets, and run payroll without leaving rtylr. Give your team the tools they need while keeping you in full control of labour costs.',
    ctaText: 'EXPLORE HR',
    image: '/images/features/hr.jpg',
  },
  {
    id: 'crm',
    category: 'CRM',
    title: 'Customers, Loyalty & Campaigns',
    description:
      'Build lasting relationships with rich customer profiles, points-based loyalty programs, and targeted marketing campaigns, all driven by your own sales data.',
    ctaText: 'EXPLORE CRM',
    image: '/images/features/crm.jpg',
  },
  {
    id: 'finance',
    category: 'Finance',
    title: 'Invoices, Expenses & P&L',
    description:
      'Stay on top of your numbers with automated invoicing, expense tracking, profit & loss statements, and tax reporting. No separate accounting software needed.',
    ctaText: 'EXPLORE FINANCE',
    image: '/images/features/finance.jpg',
  },
  {
    id: 'flow',
    category: 'Flow',
    title: 'Projects, Tasks & Team Workspaces',
    description:
      'Plan, assign, and ship work across every department. Flow brings projects, tasks, and team collaboration into one workspace. Think Linear meets Asana, built into your business OS.',
    ctaText: 'EXPLORE FLOW',
    image: '/images/features/flow.jpg',
  },
  {
    id: 'insights',
    category: 'Insights',
    title: 'Analytics & Business Intelligence',
    description:
      'See your entire business in one dashboard. Revenue trends, customer analytics, product performance, and operational metrics, all pulling from your real data across every app.',
    ctaText: 'EXPLORE INSIGHTS',
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
