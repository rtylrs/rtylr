import React, { useCallback, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../../lib/navigation';
import { track, EV_MASCOT_CATCH, EV_SIGNUP_CLICK } from '../../../lib/analytics';
import styles from './CaretMascot.module.scss';

/**
 * Caret, hiding on the homepage. It peeks out from behind the top edge of
 * whichever marked spot ([data-mascot-spot]) is on screen, watches the
 * cursor, ducks when you get close, and counts how often you catch it.
 * Three catches open a small reveal. Purely decorative: hidden from
 * assistive tech, never covers content, and never rendered for reduced
 * motion, touch or narrow screens (see CaretMascotLoader).
 */

const BODY_W = 48;
const BODY_H = 52;
const CLIP_H = 64; // the window Caret peeks through, above the spot's edge
const PEEK_MS = 3000;
const FLEE_RADIUS = 130;
const FLEE_DELAY_MS = 500;
const GOAL = 3;
const STORAGE_KEY = 'rtylr-caret-caught';

type Phase = 'hidden' | 'peeking' | 'caught';

interface Spot {
  left: number; // document coordinates of the clip window
  top: number;
}

const readCount = (): number => {
  try {
    return Math.min(GOAL, Number(window.localStorage.getItem(STORAGE_KEY)) || 0);
  } catch {
    return 0;
  }
};

const writeCount = (n: number): void => {
  try {
    window.localStorage.setItem(STORAGE_KEY, String(n));
  } catch {
    // Private mode or blocked storage: the game still works for the visit.
  }
};

const randomBetween = (a: number, b: number): number => a + Math.random() * (b - a);

const Face: React.FC<{ happy: boolean; look: { x: number; y: number } }> = ({ happy, look }) => (
  <svg width={BODY_W} height={BODY_H} viewBox="0 0 48 52" className={styles.face}>
    <rect x="2" y="2" width="44" height="48" rx="15" fill="#111214" />
    {/* The Caret knot, small, on the forehead. */}
    <g transform="translate(24 15) scale(0.42) translate(-12 -12)" fill="#f2f2ef">
      {[0, 120, 240].map((r) => (
        <path
          key={r}
          transform={`rotate(${r} 12 12)`}
          d="M10.55 11.8C8.3 9.8 7.1 7.4 8.7 5.3C10.2 3.3 13.3 3 15.15 4.65C17 6.3 17.35 8.9 16 10.9L14 9.8C14.9 8.5 14.55 6.9 13.5 6.05C12.45 5.2 10.8 5.35 10 6.4C9.15 7.55 10.1 9.15 11.9 10.8Z"
        />
      ))}
    </g>
    {happy ? (
      <g stroke="#fff" strokeWidth="2.4" strokeLinecap="round" fill="none">
        <path d="M12 33 q4 -5 8 0" />
        <path d="M28 33 q4 -5 8 0" />
      </g>
    ) : (
      <g className={styles.eyes}>
        <ellipse cx="16" cy="33" rx="5.2" ry="5.8" fill="#fff" />
        <ellipse cx="32" cy="33" rx="5.2" ry="5.8" fill="#fff" />
        <circle cx={16 + look.x} cy={33 + look.y} r="2.6" fill="#111214" />
        <circle cx={32 + look.x} cy={33 + look.y} r="2.6" fill="#111214" />
      </g>
    )}
  </svg>
);

const CaretMascot: React.FC = () => {
  const { t } = useTranslation();
  const [phase, setPhase] = useState<Phase>('hidden');
  const [spot, setSpot] = useState<Spot | null>(null);
  const [look, setLook] = useState({ x: 0, y: 0 });
  const [bubble, setBubble] = useState<string | null>(null);
  const [revealOpen, setRevealOpen] = useState(false);

  const phaseRef = useRef<Phase>('hidden');
  const spotRef = useRef<Spot | null>(null);
  const lastSpotEl = useRef<Element | null>(null);
  const nextPeekAt = useRef(0);
  const mouse = useRef({ x: -9999, y: -9999 });
  const timers = useRef<number[]>([]);
  const fleeTimer = useRef<number | null>(null);

  const later = (fn: () => void, ms: number): void => {
    timers.current.push(window.setTimeout(fn, ms));
  };

  const setPhaseBoth = (p: Phase): void => {
    phaseRef.current = p;
    setPhase(p);
  };

  const retreat = useCallback((delay: [number, number]) => {
    if (fleeTimer.current !== null) {
      window.clearTimeout(fleeTimer.current);
      fleeTimer.current = null;
    }
    setPhaseBoth('hidden');
    setBubble(null);
    nextPeekAt.current = Date.now() + randomBetween(delay[0], delay[1]);
  }, []);

  useEffect(() => {
    nextPeekAt.current = Date.now() + 2500;
    const pending = timers.current;
    return () => pending.forEach((id) => window.clearTimeout(id));
  }, []);

  // Pick a marked spot whose top edge is comfortably on screen and peek.
  useEffect(() => {
    const tick = window.setInterval(() => {
      if (phaseRef.current !== 'hidden' || revealOpen || Date.now() < nextPeekAt.current) return;
      const candidates = Array.from(document.querySelectorAll<HTMLElement>('[data-mascot-spot]')).filter(
        (el) => {
          const r = el.getBoundingClientRect();
          return r.top > 140 && r.top < window.innerHeight - 80 && r.width > BODY_W;
        },
      );
      if (candidates.length === 0) return;
      const fresh = candidates.filter((el) => el !== lastSpotEl.current);
      const el = (fresh.length ? fresh : candidates)[Math.floor(Math.random() * (fresh.length || candidates.length))];
      lastSpotEl.current = el;

      const r = el.getBoundingClientRect();
      const at = Number(el.dataset.mascotSpot) || 0.5;
      const next = {
        left: r.left + window.scrollX + r.width * at - BODY_W / 2,
        top: r.top + window.scrollY - CLIP_H,
      };
      spotRef.current = next;
      setSpot(next);
      setPhaseBoth('peeking');

      const caught = readCount();
      if (caught === 0) setBubble(t('home.mascot.hint'));
      else if (caught >= GOAL) setBubble(t('home.mascot.hiAgain'));

      later(() => {
        if (phaseRef.current === 'peeking') retreat([5000, 9000]);
      }, PEEK_MS);
    }, 400);
    return () => window.clearInterval(tick);
  }, [revealOpen, retreat, t]);

  // Eyes follow the cursor; getting close makes Caret duck shortly after.
  useEffect(() => {
    let raf = 0;
    const onMove = (e: PointerEvent): void => {
      mouse.current = { x: e.clientX, y: e.clientY };
      if (raf) return;
      raf = requestAnimationFrame(() => {
        raf = 0;
        const s = spotRef.current;
        if (!s || phaseRef.current !== 'peeking') return;
        const cx = s.left - window.scrollX + BODY_W / 2;
        const cy = s.top - window.scrollY + CLIP_H - BODY_H / 2;
        const dx = mouse.current.x - cx;
        const dy = mouse.current.y - cy;
        const dist = Math.hypot(dx, dy) || 1;
        setLook({ x: (dx / dist) * 2.2, y: (dy / dist) * 2.2 });
        if (dist < FLEE_RADIUS && fleeTimer.current === null) {
          fleeTimer.current = window.setTimeout(() => {
            fleeTimer.current = null;
            if (phaseRef.current === 'peeking') retreat([4000, 7000]);
          }, FLEE_DELAY_MS);
        }
      });
    };
    window.addEventListener('pointermove', onMove);
    return () => {
      window.removeEventListener('pointermove', onMove);
      cancelAnimationFrame(raf);
    };
  }, [retreat]);

  // Layout moved under Caret: hide rather than float in the wrong place.
  useEffect(() => {
    const onResize = (): void => {
      if (phaseRef.current !== 'hidden') retreat([3000, 5000]);
    };
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, [retreat]);

  const onCatch = (): void => {
    if (phaseRef.current !== 'peeking') return;
    if (fleeTimer.current !== null) {
      window.clearTimeout(fleeTimer.current);
      fleeTimer.current = null;
    }
    const next = Math.min(GOAL, readCount() + 1);
    const firstTimeFound = next === GOAL && readCount() < GOAL;
    writeCount(next);
    setPhaseBoth('caught');
    setBubble(t('home.mascot.caught', { count: next, goal: GOAL }));
    track(EV_MASCOT_CATCH, { count: next });
    later(() => {
      retreat([3500, 6000]);
      if (firstTimeFound) setRevealOpen(true);
    }, 1400);
  };

  const meetCaret = (): void => {
    track(EV_SIGNUP_CLICK, { source: 'mascot' });
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const shown = phase !== 'hidden';

  return createPortal(
    <>
      {spot && (
        <div className={styles.anchor} style={{ left: spot.left, top: spot.top }} aria-hidden="true">
          <AnimatePresence>
            {bubble && shown && (
              <motion.div
                key={bubble}
                className={styles.bubble}
                initial={{ opacity: 0, y: 6, scale: 0.9 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9 }}
                transition={{ duration: 0.25, delay: phase === 'peeking' ? 0.45 : 0 }}
              >
                {bubble}
              </motion.div>
            )}
          </AnimatePresence>

          {phase === 'caught' && (
            <div className={styles.sparkles}>
              {Array.from({ length: 8 }, (_, i) => {
                const a = (i / 8) * Math.PI * 2;
                return (
                  <motion.span
                    key={i}
                    className={styles.sparkle}
                    initial={{ x: 0, y: 0, opacity: 1, scale: 1 }}
                    animate={{ x: Math.cos(a) * 46, y: Math.sin(a) * 46 - 10, opacity: 0, scale: 0.4 }}
                    transition={{ duration: 0.7, ease: 'easeOut' }}
                  />
                );
              })}
            </div>
          )}

          <div className={styles.clip}>
            <motion.div
              className={styles.body}
              initial={{ y: BODY_H + 4 }}
              animate={
                phase === 'caught'
                  ? { y: [10, -6, 10], rotate: [0, -8, 8, 0] }
                  : { y: shown ? 10 : BODY_H + 4, rotate: 0 }
              }
              transition={
                phase === 'caught'
                  ? { duration: 0.6 }
                  : { type: 'spring', stiffness: shown ? 260 : 420, damping: shown ? 18 : 32 }
              }
              onPointerDown={onCatch}
            >
              <Face happy={phase === 'caught'} look={look} />
            </motion.div>
          </div>
        </div>
      )}

      <AnimatePresence>
        {revealOpen && (
          <motion.div
            className={styles.reveal}
            role="dialog"
            aria-labelledby="caret-found-title"
            initial={{ opacity: 0, y: 24, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 16 }}
            transition={{ type: 'spring', stiffness: 300, damping: 26 }}
          >
            <div className={styles.revealFace}>
              <Face happy look={{ x: 0, y: 0 }} />
            </div>
            <div className={styles.revealCopy}>
              <h2 id="caret-found-title" className={styles.revealTitle}>
                {t('home.mascot.foundTitle')}
              </h2>
              <p className={styles.revealBody}>{t('home.mascot.foundBody')}</p>
              <div className={styles.revealActions}>
                <button type="button" className={styles.revealPrimary} onClick={meetCaret}>
                  {t('home.mascot.foundCta')}
                </button>
                <button
                  type="button"
                  className={styles.revealSecondary}
                  onClick={() => setRevealOpen(false)}
                >
                  {t('home.mascot.close')}
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>,
    document.body,
  );
};

export default CaretMascot;
