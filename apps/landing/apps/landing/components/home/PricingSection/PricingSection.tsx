import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../lib/navigation';
import { blurFadeInSlow, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import PricingCards from '../../pricing/PricingCards';
import type { Plan } from '../../pricing/pricingData';
import styles from './PricingSection.module.scss';

const PricingSection = (): React.JSX.Element => {
  const router = useRouter();

  const handleCta = (action: Plan['ctaAction']): void => {
    if (action === 'contact') {
      void router.push('/contact');
      return;
    }
    const plan = action === 'register-pro' ? 'pro' : undefined;
    const url = plan ? `${APP_URLS.auth}/register?plan=${plan}` : `${APP_URLS.auth}/register`;
    navigateExternal(url);
  };

  return (
    <section className={styles.pricingSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={blurFadeInSlow}
        >
          <span className={styles.badge}>Pricing</span>
          <SplitWords
            as="h2"
            className={styles.sectionTitle}
            text={'Simple pricing.\nAll 7 apps, one price.'}
            trigger="view"
          />
          <p className={styles.sectionDescription}>
            Start free. Upgrade when you need more. No per-location fees, no module upsells.
          </p>
        </motion.div>

        <PricingCards onCta={handleCta} />
      </div>
    </section>
  );
};

export default PricingSection;
