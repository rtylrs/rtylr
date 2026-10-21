import React from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { motion, useReducedMotion } from 'motion/react';
import { Check, TrendingUp } from 'lucide-react';
import Logo from './Logo';

// The Caret knot (same geometry as CaretIcon in @voxire/caret).
const RIBBON =
  'M10.55 11.8C8.3 9.8 7.1 7.4 8.7 5.3C10.2 3.3 13.3 3 15.15 4.65C17 6.3 17.35 8.9 16 10.9L14 9.8C14.9 8.5 14.55 6.9 13.5 6.05C12.45 5.2 10.8 5.35 10 6.4C9.15 7.55 10.1 9.15 11.9 10.8Z';

const CaretMark: React.FC = () => (
  <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
    <g transform="translate(12 12) scale(1.12) translate(-12 -12)">
      <path d={RIBBON} />
      <path d={RIBBON} transform="rotate(120 12 12)" />
      <path d={RIBBON} transform="rotate(240 12 12)" />
    </g>
  </svg>
);

const card =
  'rounded-2xl bg-white p-4 text-[13px] text-[#111] shadow-[0_0_0_1px_rgba(0,0,0,0.06),0_12px_32px_-12px_rgba(0,0,0,0.18)]';

/** One floating card: rises in, then bobs gently (still for reduced motion). */
const Float: React.FC<{
  delay: number;
  bob: number;
  className: string;
  children: React.ReactNode;
}> = ({ delay, bob, className, children }) => {
  const reduceMotion = useReducedMotion();
  return (
    <motion.div
      className={`absolute ${className}`}
      initial={{ opacity: 0, y: 16 }}
      animate={reduceMotion ? { opacity: 1, y: 0 } : { opacity: 1, y: [0, -bob, 0] }}
      transition={
        reduceMotion
          ? { duration: 0.4, delay }
          : {
              opacity: { duration: 0.6, delay },
              y: { duration: 5 + bob / 2, delay, repeat: Infinity, ease: 'easeInOut' },
            }
      }
    >
      {children}
    </motion.div>
  );
};

/**
 * Left half of the auth screens: the rtylr pitch in the marketing site's
 * voice (two-tone headline, light surface) and three small cards that show
 * what's on the other side of the form. Decorative; hidden on mobile.
 */
const AuthShowcase: React.FC = () => {
  const { t } = useTranslation();

  return (
    <aside
      className="relative hidden lg:flex lg:w-[52%] flex-col overflow-hidden bg-[#f8f8f7] p-12 xl:p-16"
      aria-hidden="true"
    >
      {/* Soft spectrum glow, the same palette as the Caret ring on the site. */}
      <div
        className="pointer-events-none absolute -bottom-40 -right-24 h-[560px] w-[560px] rounded-full opacity-70 blur-[90px]"
        style={{
          background:
            'radial-gradient(circle at 30% 40%, rgba(255,107,157,0.22), transparent 60%),' +
            'radial-gradient(circle at 60% 30%, rgba(76,201,240,0.22), transparent 60%),' +
            'radial-gradient(circle at 70% 70%, rgba(179,136,235,0.24), transparent 60%)',
        }}
      />

      <Logo variant="dark" className="relative h-9 w-auto self-start select-none" />

      <motion.div
        className="relative mt-auto max-w-[640px]"
        initial={{ opacity: 0, y: 14 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.7, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
      >
        <h1 className="text-[44px] xl:text-[52px] leading-[1.04] font-medium tracking-[-0.04em] text-[#111]">
          <Trans
            i18nKey="auth.showcase.headline"
            components={{ muted: <span className="text-[#111]/40" /> }}
          />
        </h1>
        <p className="mt-5 max-w-md text-[17px] leading-relaxed text-[#111]/60">
          {t('auth.showcase.tagline')}
        </p>
      </motion.div>

      {/* Product cards: what you get on the other side of the form. */}
      <div className="relative mt-12 mb-auto h-[230px]">
        <Float delay={0.35} bob={6} className="left-0 top-2 w-[300px]">
          <div className={card}>
            <div className="mb-2 flex items-center gap-2 font-semibold">
              <span className="grid h-[22px] w-[22px] place-items-center rounded-[7px] bg-[#111214] text-[#f2f2ef]">
                <CaretMark />
              </span>
              Caret
            </div>
            <p className="leading-relaxed text-[#111]/70">{t('auth.showcase.caretBody')}</p>
          </div>
        </Float>

        <Float delay={0.5} bob={8} className="right-0 top-0 w-[240px]">
          <div className={`${card} !py-2.5`}>
            <div className="flex items-center gap-2.5 border-b border-black/[0.06] py-1.5 text-[#111]/40 line-through">
              <span className="grid h-4 w-4 place-items-center rounded-[5px] bg-[#111] text-white">
                <Check size={10} strokeWidth={3.5} />
              </span>
              {t('auth.showcase.taskDone')}
            </div>
            <div className="flex items-center gap-2.5 py-1.5">
              <span className="h-4 w-4 rounded-[5px] shadow-[inset_0_0_0_1.5px_rgba(0,0,0,0.25)]" />
              {t('auth.showcase.taskOpen')}
              <span className="ml-auto rounded-full bg-black/[0.06] px-2 text-[11px] text-[#111]/60">
                {t('auth.showcase.taskDue')}
              </span>
            </div>
          </div>
        </Float>

        <Float delay={0.65} bob={5} className="left-[34%] top-[128px] w-[230px]">
          <div className={card}>
            <div className="text-[#111]/55">{t('auth.showcase.revenueLabel')}</div>
            <div className="mt-1 flex items-baseline gap-2">
              <span className="text-[22px] font-semibold tracking-[-0.03em]">$48,200</span>
              <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-600">
                <TrendingUp size={12} />
                12.4%
              </span>
            </div>
          </div>
        </Float>
      </div>
    </aside>
  );
};

export default AuthShowcase;
