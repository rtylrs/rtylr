import React, { useRef, useState, useCallback } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { motion, useMotionValue, useSpring, useTransform } from 'motion/react';
import Logo from './Logo';

const GLOW_SIZE = 560;

interface Ripple {
  id: number;
  x: number;
  y: number;
}

const AuthShowcase: React.FC = () => {
  const { t } = useTranslation();
  const containerRef = useRef<HTMLElement>(null);
  const [ripples, setRipples] = useState<Ripple[]>([]);

  // Cursor glow
  const glowX = useMotionValue(-GLOW_SIZE);
  const glowY = useMotionValue(-GLOW_SIZE);
  const x = useSpring(glowX, { stiffness: 80, damping: 22, mass: 0.6 });
  const y = useSpring(glowY, { stiffness: 80, damping: 22, mass: 0.6 });

  // Normalised cursor offset from center (-0.5 → 0.5) for parallax
  const mouseNX = useMotionValue(0);
  const mouseNY = useMotionValue(0);
  const smoothNX = useSpring(mouseNX, { stiffness: 40, damping: 20, mass: 0.8 });
  const smoothNY = useSpring(mouseNY, { stiffness: 40, damping: 20, mass: 0.8 });

  const handleMouseMove = useCallback(
    (e: React.MouseEvent<HTMLElement>): void => {
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return;
      glowX.set(e.clientX - rect.left - GLOW_SIZE / 2);
      glowY.set(e.clientY - rect.top - GLOW_SIZE / 2);
      mouseNX.set((e.clientX - rect.left) / rect.width - 0.5);
      mouseNY.set((e.clientY - rect.top) / rect.height - 0.5);
    },
    [glowX, glowY, mouseNX, mouseNY]
  );

  const handleClick = useCallback((e: React.MouseEvent<HTMLElement>): void => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const id = Date.now();
    setRipples((prev) => [...prev, { id, x: e.clientX - rect.left, y: e.clientY - rect.top }]);
    setTimeout(() => setRipples((prev) => prev.filter((r) => r.id !== id)), 900);
  }, []);

  // Parallax offsets per orb — different factors create depth layers
  const orb1X = useTransform(smoothNX, [-0.5, 0.5], [-35, 35]);
  const orb1Y = useTransform(smoothNY, [-0.5, 0.5], [-20, 20]);
  const orb2X = useTransform(smoothNX, [-0.5, 0.5], [25, -25]);
  const orb2Y = useTransform(smoothNY, [-0.5, 0.5], [18, -18]);
  const orb3X = useTransform(smoothNX, [-0.5, 0.5], [-18, 18]);
  const orb3Y = useTransform(smoothNY, [-0.5, 0.5], [-28, 28]);
  const orb4X = useTransform(smoothNX, [-0.5, 0.5], [30, -30]);
  const orb4Y = useTransform(smoothNY, [-0.5, 0.5], [22, -22]);

  return (
    <aside
      ref={containerRef}
      onMouseMove={handleMouseMove}
      onClick={handleClick}
      className="relative hidden lg:flex lg:w-[60%] flex-col justify-between overflow-hidden p-12 xl:p-16 cursor-crosshair"
      aria-hidden="true"
    >
      {/* Base background */}
      <div
        className="absolute inset-0"
        style={{
          background:
            'radial-gradient(ellipse 80% 60% at 20% 15%, rgba(69,131,196,0.10) 0%, transparent 60%),' +
            'radial-gradient(ellipse 70% 50% at 85% 85%, rgba(126,68,150,0.12) 0%, transparent 55%),' +
            'radial-gradient(ellipse 60% 40% at 60% 50%, rgba(100,197,185,0.06) 0%, transparent 65%),' +
            'linear-gradient(135deg, #151515 0%, #141414 50%, #151515 100%)',
        }}
      />

      {/* Cursor-following glow */}
      <motion.div
        className="absolute rounded-full pointer-events-none blur-[90px] mix-blend-screen"
        style={{
          width: GLOW_SIZE,
          height: GLOW_SIZE,
          x,
          y,
          background:
            'radial-gradient(circle, rgba(69,131,196,0.28) 0%, rgba(126,68,150,0.16) 40%, rgba(100,197,185,0.07) 60%, transparent 72%)',
        }}
      />

      {/* Dot grid overlay */}
      <div
        className="absolute inset-0 opacity-[0.18] pointer-events-none"
        style={{
          backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.25) 1px, transparent 1px)',
          backgroundSize: '22px 22px',
        }}
      />

      {/* Click ripples */}
      {ripples.map((r) => (
        <motion.div
          key={r.id}
          className="absolute rounded-full pointer-events-none border border-white/20"
          style={{ left: r.x, top: r.y, translateX: '-50%', translateY: '-50%' }}
          initial={{ width: 0, height: 0, opacity: 0.6 }}
          animate={{ width: 320, height: 320, opacity: 0 }}
          transition={{ duration: 0.85, ease: [0.2, 0.8, 0.4, 1] }}
        />
      ))}

      {/* Orb 1 — blue, top-left, deep layer */}
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 420,
          height: 420,
          top: '-80px',
          left: '-60px',
          x: orb1X,
          y: orb1Y,
          background: 'radial-gradient(circle, rgba(69,131,196,0.22) 0%, rgba(69,131,196,0) 70%)',
        }}
        animate={{ x: [0, 60, -30, 0], y: [0, 40, 80, 0], scale: [1, 1.12, 0.95, 1] }}
        transition={{ duration: 22, repeat: Infinity, ease: 'easeInOut' }}
      />

      {/* Orb 2 — purple, bottom-right, mid layer */}
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 380,
          height: 380,
          bottom: '-100px',
          right: '-40px',
          x: orb2X,
          y: orb2Y,
          background: 'radial-gradient(circle, rgba(126,68,150,0.20) 0%, rgba(126,68,150,0) 70%)',
        }}
        animate={{ x: [0, -50, 30, 0], y: [0, -40, -20, 0], scale: [1, 1.08, 1.15, 1] }}
        transition={{ duration: 26, repeat: Infinity, ease: 'easeInOut' }}
      />

      {/* Orb 3 — teal, mid-right, shallow layer */}
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 320,
          height: 320,
          top: '40%',
          right: '15%',
          x: orb3X,
          y: orb3Y,
          background: 'radial-gradient(circle, rgba(100,197,185,0.14) 0%, rgba(100,197,185,0) 70%)',
        }}
        animate={{ x: [0, 30, -40, 0], y: [0, -60, 20, 0], scale: [1, 0.9, 1.1, 1] }}
        transition={{ duration: 30, repeat: Infinity, ease: 'easeInOut' }}
      />

      {/* Orb 4 — cyan, top-right, foreground */}
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 260,
          height: 260,
          top: '15%',
          right: '25%',
          x: orb4X,
          y: orb4Y,
          background: 'radial-gradient(circle, rgba(111,206,235,0.12) 0%, rgba(111,206,235,0) 70%)',
        }}
        animate={{ x: [0, -40, 50, 0], y: [0, 60, -30, 0], scale: [1, 1.15, 0.92, 1] }}
        transition={{ duration: 24, repeat: Infinity, ease: 'easeInOut' }}
      />

      {/* Top-left logo */}
      <motion.div
        className="relative z-10 flex items-center gap-2.5"
        initial={{ opacity: 0, y: -12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
      >
        <Logo variant="light" className="h-20 w-auto select-none" />
      </motion.div>

      {/* Middle: hero copy */}
      <motion.div
        className="relative z-10 max-w-xl"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.8, delay: 0.15, ease: [0.16, 1, 0.3, 1] }}
      >
        <h1 className="text-[44px] xl:text-[56px] leading-[1.05] font-semibold tracking-tight text-white">
          <Trans
            i18nKey="auth.showcase.headline"
            components={{
              gradient: (
                <span
                  className="bg-clip-text text-transparent"
                  style={{
                    backgroundImage:
                      'linear-gradient(120deg, #6FCEEB 0%, #64C5B9 50%, #4583C4 100%)',
                  }}
                />
              ),
            }}
          />
        </h1>
        <p className="mt-6 text-base xl:text-lg text-white/65 max-w-lg leading-relaxed">
          {t('auth.showcase.tagline')}
        </p>
      </motion.div>

      {/* Bottom: testimonial */}
      <motion.figure
        className="relative z-10 max-w-lg"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, delay: 0.45, ease: [0.16, 1, 0.3, 1] }}
      >
        <div className="rounded-2xl border border-white/10 bg-white/[0.03] backdrop-blur-sm p-5 xl:p-6">
          <blockquote className="text-[15px] leading-relaxed text-white/80">
            &ldquo;Rtylr replaced four separate tools for us. Our staff learned it in a day and our
            closing reports take half the time.&rdquo;
          </blockquote>
          <figcaption className="mt-4 flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-full flex items-center justify-center text-sm font-semibold text-white"
              style={{ background: 'linear-gradient(135deg, #4583C4 0%, #7E4496 100%)' }}
            >
              MH
            </div>
            <div className="text-xs">
              <div className="text-white/90 font-medium">Mohamad Homsi</div>
              <div className="text-white/50">{t('auth.showcase.personRole')}</div>
            </div>
          </figcaption>
        </div>
      </motion.figure>
    </aside>
  );
};

export default AuthShowcase;
