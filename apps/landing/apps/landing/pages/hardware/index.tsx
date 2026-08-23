import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  Scale,
  Archive,
  Printer,
  ScanLine,
  CreditCard,
  Monitor,
  Tablet,
  MonitorSmartphone,
  Tag,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { productPageSchema } from '../../lib/structured-data';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './hardware.module.scss';

// Model number and icon only. The name and description resolve from
// hardware.items, matched by position. A SKU is not copy.
const HARDWARE_SPECS: { model: string; icon: LucideIcon }[] = [
  { model: 'HS-117', icon: Scale },
  { model: 'CD-460', icon: Archive },
  { model: 'HS-8250', icon: Printer },
  { model: 'BS-200', icon: ScanLine },
  { model: 'CT-500', icon: CreditCard },
  { model: 'KDS-15', icon: Monitor },
  { model: 'TS-360', icon: Tablet },
  { model: 'CD-Dual', icon: MonitorSmartphone },
  { model: 'LP-420', icon: Tag },
];

const HardwarePage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const items = t('hardware.items', { returnObjects: true }) as unknown as {
    name: string;
    description: string;
  }[];
  const router = useRouter();

  const handleBookDemo = (): void => {
    void router.push('/demo');
  };

  const handleContactSales = (): void => {
    void router.push('/contact');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('hardware.meta.title')}
          description={t('hardware.meta.description')}
          canonical="https://rtylr.com/hardware/"
          structuredData={productPageSchema({
            name: 'Hardware',
            path: '/hardware/',
            description: t('hardware.meta.description'),
          })}
        />

        <div className={styles.hardwarePage}>
          {/* Hero */}
          <motion.section
            className={styles.hero}
            initial="hidden"
            animate="visible"
            variants={staggerBlur}
          >
            <div className={styles.heroInner}>
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <span className={styles.eyebrowDot} />
                {t('hardware.hero.eyebrow')}
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                {t('hardware.hero.title')}
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('hardware.hero.subtitle')}
              </motion.p>
            </div>
          </motion.section>

          {/* Hardware grid */}
          <motion.section
            className={styles.gridSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              {/* The device names below are h3. Without an h2 above them the
                  page jumped h1 to h3, which is the one heading rule this
                  repo gates on. */}
              <motion.h2 className={styles.gridTitle} variants={blurFadeIn}>
                {t('hardware.gridTitle')}
              </motion.h2>
              <div className={styles.hardwareGrid}>
                {HARDWARE_SPECS.map((spec, i) => {
                  const Icon = spec.icon;
                  const item = items[i];
                  return (
                    <motion.div
                      key={spec.model}
                      className={styles.hardwareCard}
                      variants={blurFadeIn}
                    >
                      <div className={styles.cardAccentBar} aria-hidden="true" />
                      <div className={styles.cardContent}>
                        <div className={styles.cardIconWrap}>
                          <Icon size={24} strokeWidth={1.5} />
                        </div>
                        <div className={styles.cardMeta}>
                          <span className={styles.cardModel}>{spec.model}</span>
                          <h3 className={styles.cardName}>{item.name}</h3>
                        </div>
                        <p className={styles.cardDescription}>{item.description}</p>
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            </div>
          </motion.section>

          {/* Integration strip */}
          <motion.section
            className={styles.integrationStrip}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={blurFadeIn}
          >
            <div className={styles.stripInner}>
              <span className={styles.stripBadge}>{t('hardware.strip.badge')}</span>
              <p className={styles.stripText}>
                {t('hardware.strip.text')}
              </p>
            </div>
          </motion.section>

          {/* CTA */}
          <motion.section
            className={styles.ctaSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.sectionContainer}>
              <motion.div className={styles.ctaInner} variants={blurFadeIn}>
                <h2 className={styles.ctaTitle}>{t('hardware.cta.title')}</h2>
                <p className={styles.ctaSubtitle}>
                  {t('hardware.cta.subtitle')}
                </p>
                <div className={styles.ctaButtons}>
                  <button className={styles.ctaPrimary} onClick={handleBookDemo}>
                    {t('common.cta.bookDemo')}
                  </button>
                  <button className={styles.ctaSecondary} onClick={handleContactSales}>
                    {t('hardware.cta.contactSales')}
                  </button>
                </div>
              </motion.div>
            </div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default HardwarePage;
