import React, { useEffect, useRef, useState } from 'react';
import { AnimatePresence, motion, useInView, useReducedMotion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import {
  BarChart2,
  Briefcase,
  Heart,
  Package,
  ShoppingCart,
  Users,
  Wallet,
} from 'lucide-react';
import { cn } from '../../../lib/cn';
import { blurFadeIn, blurFadeInSlow, viewportOnce } from '../../../lib/animations';
import { SplitWords } from '../../motion/SplitWords';
import styles from './ConnectedSection.module.scss';

type Area = 'flow' | 'finance' | 'crm' | 'hr' | 'erp' | 'pos' | 'insights';

// Positions are percentages of the diagram box; colours match the portal.
const NODES: Record<Area, { x: number; y: number; color: string; Icon: React.ComponentType<{ size?: number }> }> = {
  pos: { x: 12, y: 22, color: '#3b82f6', Icon: ShoppingCart },
  erp: { x: 10, y: 74, color: '#64c5b9', Icon: Package },
  hr: { x: 34, y: 92, color: '#10b981', Icon: Users },
  finance: { x: 66, y: 92, color: '#f59e0b', Icon: Wallet },
  crm: { x: 90, y: 74, color: '#8b5cf6', Icon: Heart },
  insights: { x: 88, y: 22, color: '#6fa3db', Icon: BarChart2 },
  flow: { x: 50, y: 8, color: '#ef4444', Icon: Briefcase },
};
const CENTER = { x: 50, y: 50 };

// Each event starts in one area and lands in others. Captions are matched by
// index to home.connected.events.
const EVENTS: { from: Area; to: Area[] }[] = [
  { from: 'pos', to: ['erp', 'finance', 'crm'] },
  { from: 'hr', to: ['finance', 'insights'] },
  { from: 'finance', to: ['crm', 'flow', 'insights'] },
  { from: 'erp', to: ['flow', 'finance'] },
];
const EVENT_MS = 3200;
const LEG_S = 0.6; // one hop: area → centre, or centre → area

/**
 * "One update, everywhere": the areas sit around rtylr's shared records, and
 * a real event travels in from one area and fans out to the ones it touches.
 */
const ConnectedSection: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const reduceMotion = useReducedMotion();
  const boxRef = useRef<HTMLDivElement>(null);
  const inView = useInView(boxRef, { margin: '-80px' });
  const [index, setIndex] = useState(0);

  const events = t('home.connected.events', { returnObjects: true }) as {
    title: string;
    effect: string;
  }[];

  useEffect(() => {
    if (!inView || reduceMotion) return;
    const timer = setInterval(() => setIndex((i) => (i + 1) % EVENTS.length), EVENT_MS);
    return () => clearInterval(timer);
  }, [inView, reduceMotion]);

  const event = EVENTS[index];
  const animate = inView && !reduceMotion;

  return (
    <section className={styles.connectedSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={blurFadeInSlow}
        >
          <span className={styles.eyebrow}>{t('home.connected.eyebrow')}</span>
          <SplitWords
            as="h2"
            className={styles.sectionTitle}
            text={t('home.connected.title')}
            restLinesClassName={styles.sectionTitleMuted}
            trigger="view"
          />
          <motion.p className={styles.subtitle} variants={blurFadeIn}>
            {t('home.connected.subtitle')}
          </motion.p>
        </motion.div>

        <div className={styles.diagram} ref={boxRef} aria-hidden="true">
          {/* Peek spot sits on the core tile, not the diagram box. */}
          <span
            className={styles.mascotSpot}
            style={{ left: `${CENTER.x}%`, top: `${CENTER.y}%` }}
            data-mascot-spot="0.5"
          />
          <svg className={styles.lines} viewBox="0 0 100 100" preserveAspectRatio="none">
            {(Object.keys(NODES) as Area[]).map((id) => {
              const n = NODES[id];
              const lit = id === event.from || event.to.includes(id);
              return (
                <line
                  key={id}
                  x1={CENTER.x}
                  y1={CENTER.y}
                  x2={n.x}
                  y2={n.y}
                  vectorEffect="non-scaling-stroke"
                  className={cn(styles.line, lit && styles.lineLit)}
                  style={{ '--node': n.color } as React.CSSProperties}
                />
              );
            })}
          </svg>

          {/* Packets: one in from the source, then one out per target. */}
          {animate && (
            <React.Fragment key={index}>
              <motion.span
                className={styles.packet}
                style={{ background: NODES[event.from].color }}
                initial={{ left: `${NODES[event.from].x}%`, top: `${NODES[event.from].y}%`, opacity: 0 }}
                animate={{
                  left: `${CENTER.x}%`,
                  top: `${CENTER.y}%`,
                  opacity: [0, 1, 1, 0],
                }}
                transition={{ duration: LEG_S, ease: 'easeInOut', delay: 0.2 }}
              />
              {event.to.map((id) => (
                <motion.span
                  key={id}
                  className={styles.packet}
                  style={{ background: NODES[id].color }}
                  initial={{ left: `${CENTER.x}%`, top: `${CENTER.y}%`, opacity: 0 }}
                  animate={{ left: `${NODES[id].x}%`, top: `${NODES[id].y}%`, opacity: [0, 1, 1, 0] }}
                  transition={{ duration: LEG_S, ease: 'easeInOut', delay: 0.2 + LEG_S }}
                />
              ))}
            </React.Fragment>
          )}

          <motion.div
            className={styles.core}
            style={{ left: `${CENTER.x}%`, top: `${CENTER.y}%` }}
            animate={animate ? { scale: [1, 1.06, 1] } : undefined}
            transition={{ duration: 0.5, delay: 0.2 + LEG_S * 0.8 }}
            key={`core-${index}`}
          >
            <img className={styles.coreMark} src="/logo-mark-light.png" alt="" />
          </motion.div>

          {(Object.keys(NODES) as Area[]).map((id) => {
            const n = NODES[id];
            const isSource = id === event.from;
            const isTarget = event.to.includes(id);
            return (
              <div
                key={id}
                className={cn(styles.node, (isSource || isTarget) && styles.nodeLit)}
                style={
                  {
                    left: `${n.x}%`,
                    top: `${n.y}%`,
                    '--node': n.color,
                    // Targets light up as their packet lands.
                    '--lit-delay': isTarget && animate ? `${0.2 + LEG_S * 2 - 0.1}s` : '0s',
                  } as React.CSSProperties
                }
              >
                <span className={styles.nodeIcon}>
                  <n.Icon size={14} />
                </span>
                {t(`home.hero.portal.apps.${id}`)}
              </div>
            );
          })}
        </div>

        <div className={styles.caption} aria-live="polite">
          <AnimatePresence mode="wait" initial={false}>
            <motion.div
              key={index}
              className={styles.captionInner}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -6 }}
              transition={{ duration: 0.3 }}
            >
              <span className={styles.captionDot} style={{ background: NODES[event.from].color }} />
              <strong>{events[index].title}</strong>
              <span className={styles.captionEffect}>{events[index].effect}</span>
            </motion.div>
          </AnimatePresence>
        </div>
      </div>
    </section>
  );
};

export default ConnectedSection;
