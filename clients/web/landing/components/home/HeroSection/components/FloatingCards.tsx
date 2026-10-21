import React from 'react';
import { motion, useReducedMotion, useTransform, type MotionValue } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { CalendarDays, Check, Wallet } from 'lucide-react';
import { cn } from '../../../../lib/cn';
import { EASE_OUT_EXPO } from '../../../../lib/animations';
import { CaretMark } from '../../../CaretMark/CaretMark';
import styles from '../HeroSection.module.scss';

interface FloatProps {
  progress: MotionValue<number>;
  // Pixels the card travels across the whole scroll range. Different values
  // per card are what make the layer read as depth rather than a sticker sheet.
  drift: number;
  delay: number;
  className: string;
  children: React.ReactNode;
}

const Float: React.FC<FloatProps> = ({ progress, drift, delay, className, children }) => {
  const reduceMotion = useReducedMotion();
  const y = useTransform(progress, [0, 1], [drift, -drift]);

  return (
    <motion.div
      className={cn(styles.floatCard, className)}
      style={reduceMotion ? undefined : { y }}
      initial={{ opacity: 0, scale: 0.92 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ delay, duration: 0.9, ease: EASE_OUT_EXPO }}
    >
      {children}
    </motion.div>
  );
};

/**
 * Small illustrative cards around the product window, one per part of the
 * workspace. Decorative: the window's own alt text describes the product, so
 * this layer is hidden from assistive tech.
 */
const FloatingCards: React.FC<{ progress: MotionValue<number> }> = ({ progress }) => {
  const { t } = useTranslation();

  return (
    <div className={styles.floatLayer} aria-hidden="true">
      <Float progress={progress} drift={14} delay={1.1} className={styles.floatCaret}>
        <div className={styles.floatHead}>
          <span className={styles.floatCaretMark}>
            <CaretMark size={14} />
          </span>
          {t('home.hero.floating.caretName')}
          <span className={styles.floatMeta}>{t('home.hero.floating.caretTime')}</span>
        </div>
        <p className={styles.floatBody}>{t('home.hero.floating.caretBody')}</p>
      </Float>

      <Float progress={progress} drift={22} delay={1.25} className={styles.floatTask}>
        <div className={styles.floatTaskRow}>
          <span className={styles.floatCheck}>
            <Check size={11} strokeWidth={3} />
          </span>
          <span className={styles.floatDone}>{t('home.hero.floating.taskDone')}</span>
        </div>
        <div className={styles.floatTaskRow}>
          <span className={styles.floatBox} />
          {t('home.hero.floating.taskOpen')}
          <span className={styles.floatTag}>{t('home.hero.floating.taskDue')}</span>
        </div>
      </Float>

      <Float progress={progress} drift={18} delay={1.4} className={styles.floatMoney}>
        <div className={styles.floatHead}>
          <Wallet size={14} />
          {t('home.hero.floating.moneyLabel')}
          <span className={styles.floatMeta}>{t('home.hero.floating.moneyLeft')}</span>
        </div>
        <div className={styles.floatAmount}>{t('home.hero.floating.moneyAmount')}</div>
        <div className={styles.floatBar}>
          <span style={{ width: '72%' }} />
        </div>
      </Float>

      <Float progress={progress} drift={26} delay={1.55} className={styles.floatEvent}>
        <div className={styles.floatHead}>
          <CalendarDays size={14} />
          {t('home.hero.floating.eventDay')}
        </div>
        <div className={styles.floatEventItem}>
          <span className={styles.floatEventBar} />
          <div>
            <div className={styles.floatEventTitle}>{t('home.hero.floating.eventTitle')}</div>
            <div className={styles.floatMeta}>{t('home.hero.floating.eventTime')}</div>
          </div>
        </div>
      </Float>
    </div>
  );
};

export default FloatingCards;
