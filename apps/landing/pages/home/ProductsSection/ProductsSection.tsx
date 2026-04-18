import React from 'react';
import { motion } from 'motion/react';
import { ShoppingCart, Package, Users, Heart, DollarSign, Workflow, BarChart2 } from 'lucide-react';
import { blurFadeIn, blurFadeInSlow, staggerBlur, viewportOnce } from '../../../lib/animations';
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
    title: 'Point of Sale',
    description:
      'Floor and counter sales, table management, and kitchen display — all in one flow.',
    Icon: ShoppingCart,
  },
  {
    label: 'ERP',
    title: 'Inventory & Purchasing',
    description: 'Products, suppliers, and purchase orders. Know your stock in real time.',
    Icon: Package,
  },
  {
    label: 'HR',
    title: 'People & Payroll',
    description: 'Staff profiles, shift scheduling, payroll runs, and leave management.',
    Icon: Users,
  },
  {
    label: 'CRM',
    title: 'Customer Loyalty',
    description: 'Loyalty programs, campaigns, and customer segments that keep people coming back.',
    Icon: Heart,
  },
];

const PRODUCTS_BOTTOM: Product[] = [
  {
    label: 'Finance',
    title: 'Invoices & Expenses',
    description: 'Invoicing, expense tracking, and a live P&L — no accountant required.',
    Icon: DollarSign,
  },
  {
    label: 'Flow',
    title: 'Projects & Tasks',
    description: 'Workspaces, tasks, and team collaboration for every project your business runs.',
    Icon: Workflow,
  },
  {
    label: 'Insights',
    title: 'Analytics',
    description: 'Business-wide reporting and dashboards so you always know what matters.',
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
    <motion.div className={styles.card} variants={blurFadeIn} custom={index} whileHover={{ y: 0 }}>
      <div className={styles.iconWrap}>
        <Icon size={20} strokeWidth={1.6} />
      </div>
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
            The Platform
          </motion.span>
          <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
            One platform.
            <br />
            Seven products.
          </motion.h2>
          <motion.p className={styles.sectionSubtitle} variants={blurFadeIn}>
            Every app ships built-in and connected. No integrations to configure, no data silos to
            bridge.
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
