import React, { useState, useEffect, useRef } from 'react';
import { Button } from '@voxire/shared';
import { useRouter } from 'next/router';
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
    category: 'Point of Sale',
    title: 'Complete POS & Payment Processing',
    description:
      'Accept all major credit cards, mobile payments, and more. Process transactions seamlessly with integrated payment processing that keeps your business running smoothly.',
    ctaText: 'EXPLORE POS',
    image: 'https://placehold.co/600x400',
  },
  {
    id: 'inventory',
    category: 'Operations',
    title: 'Real-time Inventory Management',
    description:
      'Track inventory in real-time with automated alerts and smart reordering. Never run out of stock again with intelligent inventory management that adapts to your business needs.',
    ctaText: 'MANAGE INVENTORY',
    image: 'https://placehold.co/600x400',
  },
  {
    id: 'analytics',
    category: 'Business Intelligence',
    title: 'AI-Powered Analytics & Insights',
    description:
      'Make data-driven decisions with AI-powered insights. Understand your business better with predictive analytics, custom reports, and actionable intelligence.',
    ctaText: 'VIEW ANALYTICS',
    image: 'https://placehold.co/600x400',
  },
  {
    id: 'ecommerce',
    category: 'Online Sales',
    title: 'eCommerce & Online Ordering',
    description:
      'Seamless online ordering integration. Manage both in-store and online orders from one unified system that grows with your business.',
    ctaText: 'GO ONLINE',
    image: 'https://placehold.co/600x400',
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

    // Observe all feature sections
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

  const activeFeatureData =
    features.find((f) => f.id === activeFeature) || features[0];

  return (
    <section className={styles.featuresSection}>
      <div className={styles.container}>
        <div className={styles.content}>
          <div className={styles.featuresList}>
            {features.map((feature, index) => (
              <React.Fragment key={feature.id}>
                {index > 0 && <div className={styles.featureDivider} />}
                <div
                  ref={(el) => {
                    sectionRefs.current[feature.id] = el;
                  }}
                  data-feature-id={feature.id}
                  className={`${styles.featureItem} ${
                    activeFeature === feature.id ? styles.active : ''
                  }`}
                >
                  <span className={styles.featureCategory}>
                    {feature.category}
                  </span>
                  <h3 className={styles.featureTitle}>{feature.title}</h3>
                  <p className={styles.featureDescription}>
                    {feature.description}
                  </p>
                  <div className={styles.featureActions}>
                    <Button
                      variant="dark-outline"
                      onClick={() => router.push('/demo')}
                      className={styles.featureButton}
                    >
                      {feature.ctaText}
                    </Button>
                  </div>
                </div>
              </React.Fragment>
            ))}
          </div>
        </div>

        <div className={styles.showcase}>
          <div className={styles.showcaseImageContainer}>
            {features.map((feature) => (
              <div
                key={feature.id}
                className={`${styles.showcaseImage} ${
                  activeFeature === feature.id ? styles.active : ''
                }`}
              >
                <img
                  src={feature.image}
                  alt={feature.title}
                  className={styles.showcaseImg}
                />
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default FeaturesSection;
