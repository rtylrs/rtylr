import React from 'react';
import { motion } from 'motion/react';
import { ShoppingCart, Package, Users, Heart, DollarSign, Workflow, BarChart2 } from 'lucide-react';
import {
  blurFadeIn,
  blurFadeInSlow,
  staggerBlur,
  viewportOnce,
  SPRING_SMOOTH,
} from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import styles from './ProductsSection.module.scss';

interface Product {
  label: string;
  title: string;
  description: string;
  Icon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
}

const PRODUCTS_TOP: Product[] = [
  {
    label: 'POS',
    title: 'Point of sale',
    description:
      'Tickets, tables, splits, tips, and every payment type. The register that keeps up with real service.',
    Icon: ShoppingCart,
  },
  {
    label: 'ERP',
    title: 'Inventory',
    description:
      'Recipe-linked depletion, supplier management, auto-reorder. Your shelves tell the database.',
    Icon: Package,
  },
  {
    label: 'HR',
    title: 'People & payroll',
    description:
      'Scheduling, clock-in, payroll, time off, compliance. The people side, without the paperwork.',
    Icon: Users,
  },
  {
    label: 'CRM',
    title: 'Customers & loyalty',
    description:
      'Profiles, segments, campaigns, and loyalty, fed live by every receipt your POS writes.',
    Icon: Heart,
  },
];

const PRODUCTS_BOTTOM: Product[] = [
  {
    label: 'Finance',
    title: 'Books that close themselves',
    description:
      'Invoices, expenses, sales tax, and a live P&L. Every sale reconciles in real time.',
    Icon: DollarSign,
  },
  {
    label: 'Flow',
    title: 'Projects & tasks',
    description:
      'Kanban, automations, and tasks linked to the actual invoice, customer, or shift they belong to.',
    Icon: Workflow,
  },
  {
    label: 'Insights',
    title: 'Analytics across everything',
    description:
      'Revenue trends, customer cohorts, labor margin. Every number pulled live from every other app.',
    Icon: BarChart2,
  },
];

interface ProductCardProps {
  product: Product;
  index: number;
}

const ProductCard: React.FC<ProductCardProps> = ({
  product,
  index,
}: ProductCardProps): React.JSX.Element => {
  const { Icon } = product;
  return (
    <motion.div
      className={styles.card}
      variants={blurFadeIn}
      custom={index}
      whileHover={{ y: -6, transition: SPRING_SMOOTH }}
    >
      <motion.div className={styles.iconWrap} whileHover={{ scale: 1.08, rotate: -3 }}>
        <Icon size={20} strokeWidth={1.6} />
      </motion.div>
      <p className={styles.cardLabel}>{product.label}</p>
      <h3 className={styles.cardTitle}>{product.title}</h3>
      <p className={styles.cardDescription}>{product.description}</p>
      <span className={styles.cardArrow}>Explore →</span>
    </motion.div>
  );
};

const ProductsSection: React.FC = (): React.JSX.Element => {
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
            THE PLATFORM
          </motion.span>
          <SplitWords
            as="h2"
            className={styles.sectionTitle}
            text={'Seven apps.\nOne database.'}
            trigger="view"
          />
          <motion.p className={styles.sectionSubtitle} variants={blurFadeIn}>
            Not bundled. Not integrated. Built into the same platform from the first line of code,
            so a sale in POS updates inventory, payroll, and the P&amp;L in the same breath.
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
            <ProductCard key={product.label} product={product} index={i} />
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
            <ProductCard key={product.label} product={product} index={i} />
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default ProductsSection;
