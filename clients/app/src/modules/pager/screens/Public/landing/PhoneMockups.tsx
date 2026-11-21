import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AnimatePresence, motion, useMotionValue, useSpring } from 'motion/react';
import { BellRing, Check, QrCode, RotateCcw, VolumeX } from 'lucide-react';
import MatchBadge from '../../../shared/components/MatchBadge';
import { RenniliLogo, RenniliMark } from '../../../shared/components/RenniliLogo';
import { matchOf } from '../../../shared/lib/matchWords';
import { useInView, useReducedMotion } from './useLandingObservers';

const SAMPLE_MATCH = matchOf({ word: 'strawberry' });

/** Sleek bezel phone frame with fixed dimensions (zero CLS) and optional morphing Dynamic Island. */
export const PhoneFrame = ({
  children,
  variant = 'default',
  ready = false,
  islandText,
  className = '',
  decorative = true,
}: {
  children: React.ReactNode;
  variant?: 'default' | 'compact' | 'hero';
  ready?: boolean;
  islandText?: string;
  className?: string;
  decorative?: boolean;
}): React.JSX.Element => (
  <div
    className={[
      'rl-phone',
      variant !== 'default' ? `rl-phone--${variant}` : '',
      ready ? 'rl-phone--ready' : '',
      className,
    ]
      .filter(Boolean)
      .join(' ')}
    aria-hidden={decorative ? 'true' : undefined}
  >
    {islandText ? (
      <motion.div
        className={`rl-phone-island ${ready ? 'rl-phone-island--ready' : ''}`}
        animate={{ width: ready ? 196 : 148 }}
        transition={{ type: 'spring', stiffness: 380, damping: 28 }}
        aria-hidden="true"
      >
        {ready ? (
          <span className="rl-island-eq" aria-hidden="true">
            <span />
            <span />
            <span />
            <span />
          </span>
        ) : (
          <span className="rl-phone-island-dot" />
        )}
        <span className="rl-phone-island-text">{islandText}</span>
      </motion.div>
    ) : (
      <div className="rl-phone-notch" aria-hidden="true" />
    )}
    <div className="rl-phone-screen">{children}</div>
  </div>
);

/** Miniature of step 1: camera pointing at the single counter QR poster. */
export const ScreenScan = ({ active = true }: { active?: boolean }): React.JSX.Element => {
  const { t } = useTranslation('pager');
  return (
    <div className="rl-screen rl-screen--scan">
      <div className="rl-screen-top">
        <span className="rl-screen-venue">{t('pager.landing.hero.demoVenue')}</span>
        <span className="rl-screen-lang">EN · ع</span>
      </div>
      <div className="rl-scan-finder">
        <div className="rl-scan-poster">
          <span className="rl-scan-poster-title">{t('pager.landing.how.qrCaption')}</span>
          <div className="rl-scan-qr-box">
            <QrCode size={76} strokeWidth={1.6} />
          </div>
          <RenniliLogo size={14} />
        </div>
        <span className="rl-scan-corner rl-scan-corner--tl" />
        <span className="rl-scan-corner rl-scan-corner--tr" />
        <span className="rl-scan-corner rl-scan-corner--bl" />
        <span className="rl-scan-corner rl-scan-corner--br" />
        {active && <span className="rl-scan-beam" />}
      </div>
      <p className="rl-screen-caption">{t('pager.scan.scanning')}</p>
    </div>
  );
};

/** Miniature of step 2: real emoji-word code shown to the cashier (`Ticket.tsx`). */
export const ScreenCode = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  return (
    <div className="rl-screen rl-screen--code">
      <div className="rl-screen-top">
        <span className="rl-screen-venue">{t('pager.landing.hero.demoVenue')}</span>
        <span className="rl-screen-lang">EN · ع</span>
      </div>
      <div className="rl-screen-body">
        <span className="rl-screen-kicker">{t('pager.ticket.codeLabel')}</span>
        {SAMPLE_MATCH && (
          <MatchBadge
            match={SAMPLE_MATCH}
            color={1}
            className="rl-mock-match"
            otherClassName="rl-mock-match-other"
          />
        )}
        <p className="rl-screen-status">{t('pager.ticket.showCode')}</p>
        <span className="rl-mock-notify-btn">{t('pager.ticket.notify')}</span>
      </div>
      <div className="rl-screen-foot">
        <RenniliLogo size={14} />
      </div>
    </div>
  );
};

/** Miniature of step 3: order number + queue count ahead (no countdown timer). */
export const ScreenWaiting = ({ ahead = 2 }: { ahead?: number }): React.JSX.Element => {
  const { t } = useTranslation('pager');
  return (
    <div className="rl-screen rl-screen--waiting">
      <div className="rl-screen-top">
        <span className="rl-screen-venue">{t('pager.landing.hero.demoVenue')}</span>
        <span className="rl-screen-lang">EN · ع</span>
      </div>
      <div className="rl-screen-body">
        <span className="rl-screen-kicker">{t('pager.ticket.orderLabel')}</span>
        <span className="rl-mock-order-num">42</span>
        <p className="rl-screen-status">{t('pager.ticket.waiting')}</p>
        <div className="rl-ahead-pill">
          <span className="rl-ahead-dot" />
          <span>{t('pager.ticket.ahead', { count: ahead })}</span>
        </div>
      </div>
      <div className="rl-screen-foot">
        <RenniliLogo size={14} />
      </div>
    </div>
  );
};

/** Miniature of step 4: full-screen green ready alert matching `Ticket.tsx`. */
export const ScreenReady = ({ ringKey = 0 }: { ringKey?: number }): React.JSX.Element => {
  const { t } = useTranslation('pager');
  return (
    <div className="rl-screen rl-screen--ready">
      <div className="rl-screen-top rl-screen-top--light">
        <span className="rl-screen-venue">{t('pager.landing.hero.demoVenue')}</span>
      </div>
      <div className="rl-screen-body">
        <span className="rl-screen-kicker rl-screen-kicker--light">
          {t('pager.ticket.orderLabel')} 42
        </span>
        <div className="rl-ready-mark-wrap">
          <RenniliMark key={ringKey} size={44} wiggle className="rl-ready-mark" />
        </div>
        <p className="rl-screen-status rl-screen-status--ready">{t('pager.ticket.ready')}</p>
        <p className="rl-screen-sub rl-screen-sub--light">{t('pager.ticket.readyHint')}</p>
        <div className="rl-ready-actions">
          <span className="rl-mock-btn rl-mock-btn--white">
            <Check size={15} strokeWidth={2.6} />
            {t('pager.ticket.gotIt')}
          </span>
          <span className="rl-mock-btn rl-mock-btn--ghost">
            <VolumeX size={14} strokeWidth={2.2} />
            {t('pager.ticket.stopSound')}
          </span>
        </div>
      </div>
      <div className="rl-screen-foot rl-screen-foot--light">
        <RenniliLogo size={14} />
      </div>
    </div>
  );
};

/**
 * Interactive Hero Simulator: a live cashier "Tap Ready" control paired with
 * the customer phone. Clicking Ready triggers a spring buzz, radial acoustic
 * waves, and morphs the phone into the full-screen emerald Ready state.
 */
export const HeroInteractiveStage = (): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const wrapRef = useRef<HTMLDivElement>(null);
  const inView = useInView(wrapRef, { once: false, threshold: 0.1 });
  const reducedMotion = useReducedMotion();
  const [phase, setPhase] = useState<'waiting' | 'ready'>('waiting');
  const [ringCount, setRingCount] = useState(0);
  const [userControlled, setUserControlled] = useState(false);

  const tiltX = useMotionValue(0);
  const tiltY = useMotionValue(0);
  const smoothRotateX = useSpring(tiltX, { stiffness: 220, damping: 24 });
  const smoothRotateY = useSpring(tiltY, { stiffness: 220, damping: 24 });

  // Gentle auto-cycle only until the visitor takes control themselves.
  useEffect(() => {
    if (reducedMotion || !inView || userControlled) return undefined;
    const timer = window.setInterval(() => {
      setPhase((prev) => {
        const next = prev === 'waiting' ? 'ready' : 'waiting';
        if (next === 'ready') setRingCount((n) => n + 1);
        return next;
      });
    }, 4200);
    return () => window.clearInterval(timer);
  }, [reducedMotion, inView, userControlled]);

  const isReady = phase === 'ready';

  const handleToggle = (): void => {
    setUserControlled(true);
    setPhase((prev) => {
      const next = prev === 'waiting' ? 'ready' : 'waiting';
      if (next === 'ready') setRingCount((n) => n + 1);
      return next;
    });
  };

  const handlePointerMove = (event: React.PointerEvent<HTMLDivElement>): void => {
    if (reducedMotion || event.pointerType !== 'mouse') return;
    const rect = event.currentTarget.getBoundingClientRect();
    const px = (event.clientX - rect.left) / rect.width - 0.5;
    const py = (event.clientY - rect.top) / rect.height - 0.5;
    tiltY.set(px * 12);
    tiltX.set(-py * 10);
  };

  const handlePointerLeave = (): void => {
    tiltX.set(0);
    tiltY.set(0);
  };

  return (
    <div
      ref={wrapRef}
      className="rl-hero-stage"
      onPointerMove={handlePointerMove}
      onPointerLeave={handlePointerLeave}
    >
      {/* Expanding acoustic rings behind the phone */}
      <div className={`rl-hero-waves ${isReady ? 'rl-hero-waves--ready' : ''}`} aria-hidden="true">
        <span className="rl-hero-wave rl-hero-wave--1" />
        <span className="rl-hero-wave rl-hero-wave--2" />
        <span className="rl-hero-wave rl-hero-wave--3" />
      </div>

      {/* Live interactive Cashier Tap-Ready console */}
      <div className="rl-sim-dock">
        <div className="rl-sim-meta">
          <span className="rl-sim-label">{t('pager.landing.hero.simLabel')}</span>
          <strong className="rl-sim-order">{t('pager.landing.hero.simOrder')}</strong>
        </div>

        <motion.button
          type="button"
          whileTap={{ scale: 0.96 }}
          onClick={handleToggle}
          aria-pressed={isReady}
          className={`rl-sim-btn ${isReady ? 'rl-sim-btn--ready' : ''}`}
        >
          {isReady ? (
            <>
              <RotateCcw size={15} strokeWidth={2.3} aria-hidden="true" />
              <span>{t('pager.landing.hero.simReset')}</span>
            </>
          ) : (
            <>
              <BellRing size={15} strokeWidth={2.3} aria-hidden="true" />
              <span>{t('pager.landing.hero.simTapReady')}</span>
            </>
          )}
        </motion.button>
      </div>

      {/* Customer phone + minimalist counter QR stand so the product reads visually in 2 seconds */}
      <div className="rl-hero-studio-scene">
        <motion.div
          className="rl-hero-qr-stand"
          aria-hidden="true"
          initial={{ opacity: 0, y: 18 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.16, ease: [0.22, 1, 0.36, 1] }}
        >
          <span className="rl-hero-qr-pill">{t('pager.landing.hero.qrStandLabel')}</span>
          <div className="rl-hero-qr-glyph">
            <QrCode size={52} strokeWidth={1.55} />
          </div>
          <span className="rl-hero-qr-sub">{t('pager.landing.hero.qrStandCaption')}</span>
          <span className="rl-hero-qr-foot">
            <RenniliLogo size={13} />
          </span>
        </motion.div>

        <div className="rl-hero-tilt-stage">
          <motion.div
            className="rl-hero-phone-shell"
            style={reducedMotion ? undefined : { rotateX: smoothRotateX, rotateY: smoothRotateY }}
            animate={
              isReady && !reducedMotion
                ? { x: [0, -5, 5, -4, 4, -2, 0], rotate: [0, -1, 1, -0.7, 0.7, 0] }
                : { x: 0, rotate: 0 }
            }
            transition={{ duration: 0.55, ease: 'easeInOut' }}
          >
            <AnimatePresence>
              {isReady && (
                <motion.div
                  key="hero-toast"
                  className="rl-hero-toast"
                  initial={{ opacity: 0, y: 14, scale: 0.94 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: 8, scale: 0.96 }}
                  transition={{ type: 'spring', stiffness: 420, damping: 28 }}
                  aria-hidden="true"
                >
                  <span className="rl-hero-toast-icon">
                    <BellRing size={15} strokeWidth={2.4} />
                  </span>
                  <span className="rl-hero-toast-copy">
                    <strong>{t('pager.landing.hero.toastTitle')}</strong>
                    <span>{t('pager.landing.hero.toastBody')}</span>
                  </span>
                </motion.div>
              )}
            </AnimatePresence>

            <PhoneFrame
              variant="hero"
              ready={isReady}
              islandText={
                isReady
                  ? t('pager.landing.hero.islandReady')
                  : t('pager.landing.hero.islandWaiting')
              }
            >
              <AnimatePresence mode="wait" initial={false}>
                {isReady ? (
                  <motion.div
                    key="ready"
                    className="rl-phone-motion-fill"
                    initial={{ opacity: 0, scale: 0.96 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.98 }}
                    transition={{ duration: 0.28, ease: [0.22, 1, 0.36, 1] }}
                  >
                    <ScreenReady ringKey={ringCount} />
                  </motion.div>
                ) : (
                  <motion.div
                    key="waiting"
                    className="rl-phone-motion-fill"
                    initial={{ opacity: 0, scale: 0.98 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.96 }}
                    transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
                  >
                    <ScreenWaiting ahead={2} />
                  </motion.div>
                )}
              </AnimatePresence>
            </PhoneFrame>
          </motion.div>
        </div>
      </div>

      <p className="rl-sim-hint" role="status" aria-live="polite">
        <span className={`rl-sim-hint-dot ${isReady ? 'rl-sim-hint-dot--ready' : ''}`} />
        {isReady ? t('pager.landing.hero.simRung') : t('pager.landing.hero.simHint')}
      </p>
    </div>
  );
};
