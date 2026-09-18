import React, { useState, useCallback, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { AnimatePresence, motion } from 'motion/react';
import { HiChevronDown } from 'react-icons/hi';
import { useTranslation } from 'react-i18next';
import { cn } from '../../lib/cn';
import {
  shippedLocales,
  localeFromPath,
  basePathOf,
  localisePath,
  isLocalisedRoute,
} from '../../lib/locales';
import styles from './LanguageSwitcher.module.scss';

/**
 * Endonyms are shown as-is, never translated (that is the whole point of
 * an endonym: "العربية" reads as itself in every locale). Hidden entirely
 * while fewer than two locales are shipped - see the `shipped` flag's
 * contract in lib/locales.ts. A route with no translation in the target
 * locale falls back to that locale's home page rather than linking to a
 * page that does not exist there.
 */
interface LanguageSwitcherProps {
  /** Opens the panel above the trigger instead of below. Use this wherever
   *  the trigger sits near the bottom of its container - the mobile menu
   *  footer, for instance - so the panel doesn't render off-screen. */
  openUpward?: boolean;
}

export const LanguageSwitcher: React.FC<LanguageSwitcherProps> = ({ openUpward = false }) => {
  const { t } = useTranslation();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const locales = shippedLocales();
  const current = localeFromPath(router.asPath);
  const base = basePathOf(router.asPath);

  useEffect(() => {
    if (!open) return undefined;

    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest('.language-switcher-container')) setOpen(false);
    };
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [open]);

  const toggle = useCallback(() => setOpen((prev) => !prev), []);

  if (locales.length < 2) return null;

  const targetHref = (code: string): string =>
    isLocalisedRoute(base) ? localisePath(base, code) : localisePath('/', code);

  return (
    <div className={cn(styles.container, 'language-switcher-container')}>
      <button
        type="button"
        className={styles.trigger}
        onClick={toggle}
        aria-expanded={open}
        aria-haspopup="true"
        aria-label={t('common.a11y.languageSwitcher')}
      >
        {current.endonym}
        <HiChevronDown
          className={styles.icon}
          style={{
            transform: open ? 'rotate(180deg)' : 'rotate(0deg)',
            transition: 'transform 0.3s cubic-bezier(0.32, 0.72, 0, 1)',
          }}
        />
      </button>
      <AnimatePresence>
        {open && (
          <motion.div
            className={cn(styles.panel, openUpward && styles.panelUp)}
            initial={{ opacity: 0, scale: 0.96, y: openUpward ? 6 : -6 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{
              opacity: 0,
              scale: 0.97,
              y: openUpward ? 4 : -4,
              transition: { duration: 0.15, ease: 'easeIn' },
            }}
            transition={{ type: 'spring', stiffness: 480, damping: 32, mass: 0.6 }}
          >
            {locales.map((l) => (
              <Link
                key={l.code}
                href={targetHref(l.code)}
                lang={l.htmlLang}
                className={styles.option}
                aria-current={l.code === current.code ? 'true' : undefined}
                onClick={() => setOpen(false)}
              >
                {l.endonym}
              </Link>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default LanguageSwitcher;
