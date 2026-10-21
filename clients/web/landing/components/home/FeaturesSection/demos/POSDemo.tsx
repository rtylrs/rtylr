import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Check, Database, Globe, Headphones, Rocket, UserPlus, Video } from 'lucide-react';
import { cn } from '../../../../lib/cn';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { money, useDemoStep } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './POSDemo.module.scss';

// Prices by index into home.features.demos.pos.items.
const PRICES = [49, 120, 15, 39, 12, 80];
const ICONS = [Rocket, Video, UserPlus, Headphones, Globe, Database];
// Products tapped in order, one per step; then charge, then paid.
const TAPS = [0, 2, 3];
const STEPS = TAPS.length + 2;

export const POSDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const step = useDemoStep(isActive, STEPS, 1400, STEPS - 1);
  const items = t('home.features.demos.pos.items', { returnObjects: true }) as string[];

  const inOrder = TAPS.slice(0, Math.min(step + 1, TAPS.length));
  const total = inOrder.reduce((sum, i) => sum + PRICES[i], 0);
  const charging = step === TAPS.length;
  const paid = step === TAPS.length + 1;
  const lastTap = step < TAPS.length ? TAPS[step] : -1;

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="pos"
      crumb={t('home.features.demos.pos.crumb', { returnObjects: true }) as string[]}
    >
      <div className={local.layout}>
        <div className={local.products}>
          {items.map((name, i) => (
            <div key={name} className={cn(local.product, i === lastTap && local.productTapped)}>
              {React.createElement(ICONS[i], { size: 16, className: local.productIcon })}
              <span className={local.productName}>{name}</span>
              <span className={styles.muted}>{money(PRICES[i])}</span>
            </div>
          ))}
        </div>

        <div className={cn(styles.card, local.order)}>
          <span className={local.orderTitle}>{t('home.features.demos.pos.order')}</span>
          <div className={local.lines}>
            <AnimatePresence initial={false}>
              {inOrder.map((i) => (
                <motion.div
                  key={i}
                  className={styles.spread}
                  initial={{ opacity: 0, y: -6 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  <span>1 × {items[i]}</span>
                  <span className={local.price}>{money(PRICES[i])}</span>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>

          <div className={cn(styles.spread, local.total)}>
            <span>{t('home.features.demos.pos.total')}</span>
            <span className={local.price}>{money(total)}</span>
          </div>

          <div className={cn(styles.primaryBtn, local.charge, charging && local.charging)}>
            <AnimatePresence mode="wait" initial={false}>
              {paid ? (
                <motion.span
                  key="paid"
                  className={styles.row}
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                >
                  <Check size={14} strokeWidth={3} />
                  {t('home.features.demos.pos.paid')}
                </motion.span>
              ) : (
                <motion.span key="charge" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
                  {t('home.features.demos.pos.charge')} {money(total)}
                </motion.span>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>
    </PortalFrame>
  );
};
