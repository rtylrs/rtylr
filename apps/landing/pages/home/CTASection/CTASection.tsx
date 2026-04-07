import React from 'react';
import { Button } from '@voxire/components';
import { useRouter } from 'next/router';
import styles from './CTASection.module.scss';

const CTASection: React.FC = () => {
  const router = useRouter();

  const handleBookDemo = () => {
    router.push('/demo');
  };

  const handleGetStarted = () => {
    router.push('/demo');
  };

  return (
    <section className={styles.ctaSection}>
      <div className={styles.container}>
        <h2 className={styles.ctaTitle}>Ready to Transform Your Business?</h2>
        <p className={styles.ctaDescription}>
          Join thousands of businesses using Rtylr to streamline operations and
          grow revenue.
        </p>
        <div className={styles.ctaButtons}>
          <Button
            variant="primary"
            onClick={handleGetStarted}
            width="200px"
            height="50px"
          >
            Get Started Free
          </Button>
          <Button
            variant="dark-outline"
            onClick={handleBookDemo}
            width="200px"
            height="50px"
          >
            Book a Demo
          </Button>
        </div>
      </div>
    </section>
  );
};

export default CTASection;

