import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { HiMenu, HiX, HiChevronDown, HiChevronRight } from 'react-icons/hi';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Button } from '@voxire/components';
import { APP_URLS, ASSETS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import { localeFromPath, navPath, navPathWithHash } from '../../../../lib/locales';
import { FEATURE_LINKS, RESOURCE_LINKS, COMPANY_LINKS } from '../../../../lib/navLinks';
import { LanguageSwitcher } from '../../../../components/LanguageSwitcher/LanguageSwitcher';
import styles from './HeaderMobile.module.scss';

const HeaderMobile: React.FC = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;
  const [isMenuOpen, setIsMenuOpen] = useState<boolean>(false);
  const [scrolled, setScrolled] = useState(false);
  const [featuresDropdown, setFeaturesDropdown] = useState(false);
  const [solutionsDropdown, setSolutionsDropdown] = useState(false);
  const [resourcesDropdown, setResourcesDropdown] = useState(false);

  useEffect(() => {
    let ticking = false;

    const handleScroll = () => {
      if (!ticking) {
        window.requestAnimationFrame(() => {
          const shouldBeScrolled = window.scrollY > 0;
          setScrolled(shouldBeScrolled);
          ticking = false;
        });
        ticking = true;
      }
    };

    const shouldBeScrolled = window.scrollY > 0;
    setScrolled(shouldBeScrolled);

    window.addEventListener('scroll', handleScroll, { passive: true });
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  // Lock body scroll when menu is open
  useEffect(() => {
    if (isMenuOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isMenuOpen]);

  useEffect(() => {
    if (!isMenuOpen) return;
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsMenuOpen(false);
    };
    document.addEventListener('keydown', handleEscape);
    return () => document.removeEventListener('keydown', handleEscape);
  }, [isMenuOpen]);

  const toggleMenu = useCallback(() => {
    setIsMenuOpen((prev) => !prev);
  }, []);

  const toggleFeaturesDropdown = useCallback(() => {
    setFeaturesDropdown((prev) => !prev);
    setSolutionsDropdown(false);
    setResourcesDropdown(false);
  }, []);

  const toggleSolutionsDropdown = useCallback(() => {
    setSolutionsDropdown((prev) => !prev);
    setFeaturesDropdown(false);
    setResourcesDropdown(false);
  }, []);

  const toggleResourcesDropdown = useCallback(() => {
    setResourcesDropdown((prev) => !prev);
    setFeaturesDropdown(false);
    setSolutionsDropdown(false);
  }, []);

  const handleLoginClick = useCallback(() => {
    navigateExternal(APP_URLS.auth);
    setIsMenuOpen(false);
  }, []);

  const handleSignupClick = useCallback(() => {
    navigateExternal(`${APP_URLS.auth}/register`);
    setIsMenuOpen(false);
  }, []);

  const closeMenu = useCallback(() => {
    setFeaturesDropdown(false);
    setSolutionsDropdown(false);
    setResourcesDropdown(false);
    setIsMenuOpen(false);
  }, []);

  const headerClasses = `${styles.header} ${styles['mobile-header']} ${
    !scrolled ? styles['initial-position'] : styles.scrolled
  }`;

  return (
    <>
      <header className={headerClasses}>
        <div className={styles['mobile-header-content']}>
          <Link
            href={navPath('/', locale)}
            className={styles['logo-container']}
            aria-label={t('common.a11y.home')}
            onClick={closeMenu}
          >
            <img src={ASSETS.logos.light} alt={t('common.a11y.logoAlt')} />
          </Link>

          <div className={styles['mobile-header-actions']}>
            <Button
              onClick={handleSignupClick}
              variant="dark-outline"
              width="100px"
              height="36px"
              className={styles['mobile-signup-button']}
            >
              {t('common.nav.signUp')}
            </Button>
            <button
              className={styles['mobile-menu-toggle']}
              onClick={toggleMenu}
              aria-label={t('common.a11y.toggleMenu')}
            >
              {isMenuOpen ? <HiX /> : <HiMenu />}
            </button>
          </div>
        </div>
      </header>

      <AnimatePresence>
        {isMenuOpen && (
          <motion.div
            className={styles['mobile-menu-overlay']}
            role="dialog"
            aria-modal="true"
            aria-label={t('common.a11y.navMenu')}
            initial={{ opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 10, transition: { duration: 0.18, ease: 'easeIn' } }}
            transition={{ duration: 0.35, ease: [0.32, 0.72, 0, 1] }}
          >
          <div className={styles['mobile-menu-header']}>
            <Link
              href={navPath('/', locale)}
              className={styles['mobile-menu-logo']}
              aria-label={t('common.a11y.home')}
              onClick={closeMenu}
            >
              <img src={ASSETS.logos.light} alt={t('common.a11y.logoAlt')} />
            </Link>
            <button
              className={styles['mobile-menu-close']}
              onClick={toggleMenu}
              aria-label={t('common.a11y.closeMenu')}
            >
              <HiX />
            </button>
          </div>

          <div className={styles['mobile-menu-content']}>
            <div className={styles['mobile-nav-item']}>
              <button
                type="button"
                onClick={toggleFeaturesDropdown}
                aria-expanded={featuresDropdown}
              >
                {t('common.nav.features')}
                {featuresDropdown ? (
                  <HiChevronDown
                    style={{
                      transform: 'rotate(180deg)',
                    }}
                  />
                ) : (
                  <HiChevronRight />
                )}
              </button>
              {featuresDropdown && (
                <div className={styles['mobile-dropdown']}>
                  {FEATURE_LINKS.map((item) => (
                    <Link
                      key={item.href}
                      href={navPathWithHash(item.href, locale)}
                      onClick={closeMenu}
                      className={styles['mobile-feature-link']}
                    >
                      <span className={styles['mobile-feature-title']}>
                        {t(`common.nav.featureLinks.${item.key}.title`)}
                      </span>
                      <span className={styles['mobile-feature-desc']}>
                        {t(`common.nav.featureLinks.${item.key}.desc`)}
                      </span>
                    </Link>
                  ))}
                </div>
              )}
            </div>

            <div className={styles['mobile-nav-item']}>
              <button
                type="button"
                onClick={toggleSolutionsDropdown}
                aria-expanded={solutionsDropdown}
              >
                {t('common.nav.company')}
                {solutionsDropdown ? (
                  <HiChevronDown
                    style={{
                      transform: 'rotate(180deg)',
                    }}
                  />
                ) : (
                  <HiChevronRight />
                )}
              </button>
              {solutionsDropdown && (
                <div className={styles['mobile-dropdown']}>
                  {COMPANY_LINKS.map((item) =>
                    item.external ? (
                      <a
                        key={item.key}
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={closeMenu}
                      >
                        {t(`common.nav.companyLinks.${item.key}`)}
                      </a>
                    ) : (
                      <Link key={item.key} href={navPath(item.href, locale)} onClick={closeMenu}>
                        {t(`common.nav.companyLinks.${item.key}`)}
                      </Link>
                    )
                  )}
                </div>
              )}
            </div>

            <Link
              href={navPath('/pricing', locale)}
              onClick={closeMenu}
              className={styles['mobile-nav-link']}
            >
              {t('common.nav.pricing')} <HiChevronRight />
            </Link>

            <div className={styles['mobile-nav-item']}>
              <button
                type="button"
                onClick={toggleResourcesDropdown}
                aria-expanded={resourcesDropdown}
              >
                {t('common.nav.resources')}
                {resourcesDropdown ? (
                  <HiChevronDown
                    style={{
                      transform: 'rotate(180deg)',
                    }}
                  />
                ) : (
                  <HiChevronRight />
                )}
              </button>
              {resourcesDropdown && (
                <div className={styles['mobile-dropdown']}>
                  {RESOURCE_LINKS.map((item) =>
                    item.external ? (
                      <a
                        key={item.key}
                        href={item.href}
                        target="_blank"
                        rel="noopener noreferrer"
                        onClick={closeMenu}
                      >
                        {t(`common.nav.resourceLinks.${item.key}`)}
                      </a>
                    ) : (
                      <Link key={item.key} href={navPath(item.href, locale)} onClick={closeMenu}>
                        {t(`common.nav.resourceLinks.${item.key}`)}
                      </Link>
                    )
                  )}
                </div>
              )}
            </div>

            <Link
              href={navPath('/contact', locale)}
              onClick={closeMenu}
              className={styles['mobile-nav-link']}
            >
              {t('common.nav.companyLinks.contact')} <HiChevronRight />
            </Link>
          </div>

          <div className={styles['mobile-menu-footer']}>
            <LanguageSwitcher openUpward />
            <div className={styles['mobile-menu-buttons']}>
              <Button onClick={handleLoginClick} variant="dark-outline" width="100%" height="44px">
                {t('common.nav.signIn')}
              </Button>
              <Button onClick={handleSignupClick} variant="soft" width="100%" height="44px">
                {t('common.nav.signUp')}
              </Button>
            </div>
          </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
};

export default HeaderMobile;
