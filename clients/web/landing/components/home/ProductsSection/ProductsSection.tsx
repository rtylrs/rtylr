import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import Link from 'next/link';
import { ArrowRight, BarChart2, Briefcase, Package, Users, Wallet, Heart } from 'lucide-react';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import styles from './ProductsSection.module.scss';

// Structure only. Each area's label, title and body live in the locale files
// under home.products.areas, matched by position.
interface Area {
  href: string;
  Icon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
}

const AREAS: Area[] = [
  { href: '/flow', Icon: Briefcase },
  { href: '/finance', Icon: Wallet },
  { href: '/crm', Icon: Heart },
  { href: '/hr', Icon: Users },
  { href: '/erp', Icon: Package },
  { href: 'https://insights.rtylr.com', Icon: BarChart2 },
];

interface AreaCopy {
  label: string;
  title: string;
  body: string;
}

/**
 * The workspace at a glance: one ruled grid, one cell per area. Each cell
 * reads as a two-tone sentence (bold claim, muted follow-up), the same
 * voice as the section heading.
 */
const ProductsSection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const copy = t('home.products.areas', { returnObjects: true }) as unknown as AreaCopy[];

  return (
    <section className={styles.productsSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={blurFadeInSlow}
        >
          <motion.span className={styles.badge} variants={blurFadeIn}>
            {t('home.products.badge')}
          </motion.span>
          <SplitWords
            as="h2"
            className={styles.sectionTitle}
            text={t('home.products.title')}
            restLinesClassName={styles.sectionTitleMuted}
            trigger="view"
          />
          <motion.p className={styles.sectionSubtitle} variants={blurFadeIn}>
            {t('home.products.subtitle')}
          </motion.p>
        </motion.div>

        <motion.div
          className={styles.grid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          {AREAS.map(({ href, Icon }, i) => (
            <motion.div key={href} className={styles.cell} variants={blurFadeIn} custom={i}>
              <span className={styles.cellLabel}>
                <Icon size={15} strokeWidth={1.8} />
                {copy[i].label}
              </span>
              <p className={styles.cellText}>
                <strong>{copy[i].title}</strong> {copy[i].body}
              </p>
              <Link
                href={href}
                className={styles.cellLink}
                aria-label={t('home.products.exploreLabel', { capability: copy[i].label })}
              >
                {t('home.products.explore')}
                <ArrowRight size={14} aria-hidden="true" />
              </Link>
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default ProductsSection;
