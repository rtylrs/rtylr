import React from 'react';
import { motion } from 'motion/react';
import Logo from './Logo';

/**
 * Left-side branded showcase for the auth split-screen.
 * Dark gradient background, animated floating orbs, brand message,
 * and a subtle testimonial.
 *
 * Hidden on mobile (< lg).
 */
const AuthShowcase: React.FC = () => {
  return (
    <aside
      className="relative hidden lg:flex lg:w-[60%] flex-col justify-between overflow-hidden p-12 xl:p-16"
      aria-hidden="true"
    >
      {/* Base background: deep navy with radial brand accents */}
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

      {/* Dot grid overlay */}
      <div
        className="absolute inset-0 opacity-[0.18] pointer-events-none"
        style={{
          backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.25) 1px, transparent 1px)',
          backgroundSize: '22px 22px',
        }}
      />

      {/* Floating gradient orbs */}
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 420,
          height: 420,
          top: '-80px',
          left: '-60px',
          background: 'radial-gradient(circle, rgba(69,131,196,0.18) 0%, rgba(69,131,196,0) 70%)',
        }}
        animate={{
          x: [0, 60, -30, 0],
          y: [0, 40, 80, 0],
          scale: [1, 1.12, 0.95, 1],
        }}
        transition={{ duration: 22, repeat: Infinity, ease: 'easeInOut' }}
      />
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 380,
          height: 380,
          bottom: '-100px',
          right: '-40px',
          background: 'radial-gradient(circle, rgba(126,68,150,0.15) 0%, rgba(126,68,150,0) 70%)',
        }}
        animate={{
          x: [0, -50, 30, 0],
          y: [0, -40, -20, 0],
          scale: [1, 1.08, 1.15, 1],
        }}
        transition={{ duration: 26, repeat: Infinity, ease: 'easeInOut' }}
      />
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 320,
          height: 320,
          top: '40%',
          right: '15%',
          background: 'radial-gradient(circle, rgba(100,197,185,0.10) 0%, rgba(100,197,185,0) 70%)',
        }}
        animate={{
          x: [0, 30, -40, 0],
          y: [0, -60, 20, 0],
          scale: [1, 0.9, 1.1, 1],
        }}
        transition={{ duration: 30, repeat: Infinity, ease: 'easeInOut' }}
      />
      <motion.div
        className="absolute rounded-full blur-3xl pointer-events-none"
        style={{
          width: 260,
          height: 260,
          top: '15%',
          right: '25%',
          background: 'radial-gradient(circle, rgba(111,206,235,0.08) 0%, rgba(111,206,235,0) 70%)',
        }}
        animate={{
          x: [0, -40, 50, 0],
          y: [0, 60, -30, 0],
          scale: [1, 1.15, 0.92, 1],
        }}
        transition={{ duration: 24, repeat: Infinity, ease: 'easeInOut' }}
      />

      {/* Top-left logo */}
      <motion.div
        className="relative z-10 flex items-center gap-2.5"
        initial={{ opacity: 0, y: -12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: [0.16, 1, 0.3, 1] }}
      >
        <Logo variant="light" className="h-8 w-auto select-none" />
      </motion.div>

      {/* Middle: hero copy */}
      <motion.div
        className="relative z-10 max-w-xl"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.8, delay: 0.15, ease: [0.16, 1, 0.3, 1] }}
      >
        <h1 className="text-[44px] xl:text-[56px] leading-[1.05] font-semibold tracking-tight text-white">
          Your entire business,
          <br />
          <span
            className="bg-clip-text text-transparent"
            style={{
              backgroundImage: 'linear-gradient(120deg, #6FCEEB 0%, #64C5B9 50%, #4583C4 100%)',
            }}
          >
            one platform.
          </span>
        </h1>
        <p className="mt-6 text-base xl:text-lg text-white/65 max-w-lg leading-relaxed">
          POS, ERP, HR, CRM, Finance, Flow, and Insights. Everything you need to run your operation,
          unified in a single beautiful workspace.
        </p>

        {/* Product chips */}
        <div className="mt-8 flex flex-wrap gap-2">
          {['POS', 'ERP', 'HR', 'CRM', 'Finance', 'Flow', 'Insights'].map((p, i) => (
            <motion.span
              key={p}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.4, delay: 0.35 + i * 0.05 }}
              className="px-3 py-1.5 rounded-full text-xs font-medium text-white/85 border border-white/10 bg-white/[0.04] backdrop-blur-sm"
            >
              {p}
            </motion.span>
          ))}
        </div>
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
              style={{
                background: 'linear-gradient(135deg, #4583C4 0%, #7E4496 100%)',
              }}
            >
              SK
            </div>
            <div className="text-xs">
              <div className="text-white/90 font-medium">Sara Khalil</div>
              <div className="text-white/50">COO, Marina Group</div>
            </div>
          </figcaption>
        </div>
      </motion.figure>
    </aside>
  );
};

export default AuthShowcase;
