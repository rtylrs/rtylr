import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { ShoppingCart, Package, Users, Heart, DollarSign, Workflow, BarChart2 } from 'lucide-react';
import Link from 'next/link';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import styles from './ProductsSection.module.scss';

// Structure only. The card title and description live in the locale files
// under home.products.cardsTop / cardsBottom, matched by position. The
// label stays here: POS and ERP are product names, not copy, and
// translating one would break the link to the page it points at.
interface Product {
  label: string;
  href: string;
  Icon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
}

const PRODUCTS_TOP: Product[] = [
  {
    label: 'POS',
    href: '/pos',
    Icon: ShoppingCart,
  },
  {
    label: 'ERP',
    href: '/erp',
    Icon: Package,
  },
  {
    label: 'HR',
    href: '/hr',
    Icon: Users,
  },
  {
    label: 'CRM',
    href: '/crm',
    Icon: Heart,
  },
];

const PRODUCTS_BOTTOM: Product[] = [
  {
    label: 'Finance',
    href: '/finance',
    Icon: DollarSign,
  },
  {
    label: 'Flow',
    href: '/flow',
    Icon: Workflow,
  },
  {
    label: 'Insights',
    href: 'https://insights.rtylr.com',
    Icon: BarChart2,
  },
];

interface ProductCardProps {
  product: Product;
  index: number;
  /** Resolved copy, passed in so the card does not need to know its index
      into the locale array. */
  title: string;
  description: string;
}

const ProductCard: React.FC<ProductCardProps> = ({
  product,
  index,
  title,
  description,
}: ProductCardProps): React.JSX.Element => {
  const { Icon } = product;
  const { t } = useTranslation();
  return (
    <motion.div className={styles.card} variants={blurFadeIn} custom={index}>
      <motion.div className={styles.iconWrap} whileHover={{ scale: 1.08, rotate: -3 }}>
        <Icon size={20} strokeWidth={1.6} />
      </motion.div>
      <p className={styles.cardLabel}>{product.label}</p>
      <h3 className={styles.cardTitle}>{title}</h3>
      <p className={styles.cardDescription}>{description}</p>
      <Link
        href={product.href}
        className={styles.cardArrow}
        aria-label={t('home.products.exploreLabel', { capability: product.label })}
      >
        {t('home.products.explore')}
      </Link>
    </motion.div>
  );
};

interface Card {
  title: string;
  description: string;
}

const ProductsSection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const cardsTop = t('home.products.cardsTop', { returnObjects: true }) as unknown as Card[];
  const cardsBottom = t('home.products.cardsBottom', {
    returnObjects: true,
  }) as unknown as Card[];
  return (
    <section className={styles.productsSection}>
      <div className={styles.container}>
        {/* Header */}
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
            trigger="view"
          />
          <motion.p className={styles.sectionSubtitle} variants={blurFadeIn}>
            {t('home.products.subtitle')}
          </motion.p>
        </motion.div>

        {/* Top row: 4 cards */}
        <motion.div
          className={styles.grid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          {PRODUCTS_TOP.map((product, i) => (
            <ProductCard
              key={product.label}
              product={product}
              index={i}
              title={cardsTop[i].title}
              description={cardsTop[i].description}
            />
          ))}
        </motion.div>

        {/* Bottom row: 3 cards */}
        <motion.div
          className={styles.gridBottom}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerBlur}
        >
          {PRODUCTS_BOTTOM.map((product, i) => (
            <ProductCard
              key={product.label}
              product={product}
              index={i}
              title={cardsBottom[i].title}
              description={cardsBottom[i].description}
            />
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default ProductsSection;
