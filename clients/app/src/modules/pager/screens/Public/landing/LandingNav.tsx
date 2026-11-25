import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import { RTYLR_CONTACT_CHANNELS } from '@voxire/config';
import { motion } from 'motion/react';
import { ArrowRight, Menu, MessageCircle, X } from 'lucide-react';
import LanguageSwitch from '../../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo, RtylrBrandText } from '../../../shared/components/RenniliLogo';
import {
  PAGER_FAQ_PATH,
  PAGER_FEATURES_PATH,
  PAGER_HOW_PATH,
  PAGER_LANDING_PATH,
  PAGER_RTYLR_PATH,
  PAGER_TRY_PATH,
  pagerClaimUrl,
} from '../../../pagerPaths';

const FOCUSABLE_SELECTOR = 'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])';

interface LandingNavProps {
  currentPath: string;
  scrolled: boolean;
  showMobileBar: boolean;
}

export const LandingNav = ({
  currentPath,
  scrolled,
  showMobileBar,
}: LandingNavProps): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const [sheetOpen, setSheetOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const sheetRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);

  const whatsapp = RTYLR_CONTACT_CHANNELS.find((channel) => channel.id === 'whatsapp');

  // Keyboard focus trap + Escape handling for the mobile bottom sheet.
  useEffect(() => {
    if (!sheetOpen) return undefined;
    closeRef.current?.focus();

    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape') {
        event.preventDefault();
        setSheetOpen(false);
        triggerRef.current?.focus();
        return;
      }
      if (event.key !== 'Tab' || !sheetRef.current) return;
      const focusable = Array.from(
        sheetRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)
      );
      if (focusable.length === 0) return;
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [sheetOpen]);

  const closeSheet = (restoreFocus = false): void => {
    setSheetOpen(false);
    if (restoreFocus) triggerRef.current?.focus();
  };

  const links: Array<{ to: string; label: string }> = [
    { to: PAGER_HOW_PATH, label: t('pager.landing.nav.how') },
    { to: PAGER_FEATURES_PATH, label: t('pager.landing.nav.features') },
    { to: PAGER_RTYLR_PATH, label: t('pager.landing.nav.rtylr') },
    { to: PAGER_FAQ_PATH, label: t('pager.landing.nav.faq') },
  ];

  return (
    <>
      {/* The strip scrolls away; the bar under it stays pinned. */}
      <div className={`rl-nav-shell ${scrolled ? 'rl-nav-shell--scrolled' : ''}`}>
        <div className="rl-strip">
          <div className="rl-strip-inner">
            <p className="rl-strip-text">
              <span className="rl-strip-tagline">{t('pager.landing.nav.strip')}</span>
              <span className="rl-strip-dot" aria-hidden="true" />
              <ByRtylr />
            </p>
            <LanguageSwitch />
          </div>
        </div>

        <header className="rl-nav-bar">
          <div className="rl-nav-bar-inner">
            <Link
              to={PAGER_LANDING_PATH}
              className="rl-nav-logo-link"
              aria-label={t('pager.landing.nav.backHome')}
            >
              <RenniliLogo size={26} />
            </Link>

            <nav className="rl-nav-links" aria-label={t('pager.landing.nav.label')}>
              {links.map((item) => {
                const active = currentPath === item.to;
                return (
                  <Link
                    key={item.to}
                    to={item.to}
                    className={`rl-nav-link ${active ? 'rl-nav-link--active' : ''}`}
                    aria-current={active ? 'page' : undefined}
                  >
                    {active && (
                      <motion.span
                        layoutId="rl-nav-active-bg"
                        className="rl-nav-active-bg"
                        transition={{ type: 'spring', stiffness: 420, damping: 34 }}
                        aria-hidden="true"
                      />
                    )}
                    <span className="rl-nav-link-text">
                      <RtylrBrandText text={item.label} />
                    </span>
                  </Link>
                );
              })}
            </nav>

            <div className="rl-nav-end">
              <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--primary rl-btn--nav">
                {t('pager.landing.nav.tryCta')}
              </Link>
              <button
                ref={triggerRef}
                type="button"
                className="rl-menu-btn"
                aria-label={t('pager.landing.nav.menuOpen')}
                aria-expanded={sheetOpen}
                aria-controls="rl-mobile-sheet"
                onClick={() => setSheetOpen(true)}
              >
                <Menu size={20} strokeWidth={2.2} aria-hidden="true" />
              </button>
            </div>
          </div>
        </header>
      </div>

      {sheetOpen && (
        <div className="rl-sheet-backdrop" onClick={() => closeSheet(true)}>
          <div
            ref={sheetRef}
            id="rl-mobile-sheet"
            role="dialog"
            aria-modal="true"
            aria-label={t('pager.landing.nav.label')}
            className="rl-sheet"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="rl-sheet-head">
              <RenniliLogo size={24} />
              <button
                ref={closeRef}
                type="button"
                className="rl-sheet-close"
                aria-label={t('pager.landing.nav.menuClose')}
                onClick={() => closeSheet(true)}
              >
                <X size={20} strokeWidth={2.2} aria-hidden="true" />
              </button>
            </div>

            <nav className="rl-sheet-links" aria-label={t('pager.landing.nav.label')}>
              {links.map((item) => (
                <Link
                  key={item.to}
                  to={item.to}
                  className={`rl-sheet-link ${currentPath === item.to ? 'rl-sheet-link--active' : ''}`}
                  onClick={() => closeSheet(false)}
                >
                  <RtylrBrandText text={item.label} />
                </Link>
              ))}
            </nav>

            <div className="rl-sheet-actions">
              <Link
                to={PAGER_TRY_PATH}
                className="rl-btn rl-btn--primary rl-btn--full"
                onClick={() => closeSheet(false)}
              >
                {t('pager.landing.nav.tryCta')}
                <ArrowRight
                  size={17}
                  strokeWidth={2.3}
                  aria-hidden="true"
                  className="rtl:rotate-180"
                />
              </Link>
              <a href={pagerClaimUrl('register')} className="rl-btn rl-btn--secondary rl-btn--full">
                {t('pager.landing.free.primaryCta')}
              </a>
            </div>

            <div className="rl-sheet-foot">
              <ByRtylr />
            </div>
          </div>
        </div>
      )}

      <div className={`rl-mobile-dock ${showMobileBar ? 'rl-mobile-dock--visible' : ''}`}>
        <Link to={PAGER_TRY_PATH} className="rl-btn rl-btn--primary rl-mobile-dock-cta">
          {t('pager.landing.mobileBar.tryCta')}
          <ArrowRight size={16} strokeWidth={2.3} aria-hidden="true" className="rtl:rotate-180" />
        </Link>
        {whatsapp && (
          <a
            href={whatsapp.href}
            target="_blank"
            rel="noreferrer"
            className="rl-mobile-dock-wa"
            aria-label={t('pager.landing.mobileBar.whatsappLabel')}
          >
            <MessageCircle size={20} strokeWidth={2.2} aria-hidden="true" />
          </a>
        )}
      </div>
    </>
  );
};
