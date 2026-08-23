import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { HiMenu, HiX, HiChevronDown, HiChevronRight } from 'react-icons/hi';
import { AnimatePresence, motion } from 'motion/react';
import { Button } from '@voxire/components';
import { APP_URLS, ASSETS } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import styles from './HeaderMobile.module.scss';

const HeaderMobile: React.FC = () => {
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
            href="/"
            className={styles['logo-container']}
            aria-label="rtylr home"
            onClick={closeMenu}
          >
            <img src={ASSETS.logos.light} alt="rtylr" />
          </Link>

          <div className={styles['mobile-header-actions']}>
            <Button
              onClick={handleSignupClick}
              variant="dark-outline"
              width="100px"
              height="36px"
              className={styles['mobile-signup-button']}
            >
              Sign Up
            </Button>
            <button
              className={styles['mobile-menu-toggle']}
              onClick={toggleMenu}
              aria-label="Toggle mobile menu"
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
            aria-label="Navigation menu"
            initial={{ opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 10, transition: { duration: 0.18, ease: 'easeIn' } }}
            transition={{ duration: 0.35, ease: [0.32, 0.72, 0, 1] }}
          >
          <div className={styles['mobile-menu-header']}>
            <Link
              href="/"
              className={styles['mobile-menu-logo']}
              aria-label="rtylr home"
              onClick={closeMenu}
            >
              <img src={ASSETS.logos.light} alt="rtylr" />
            </Link>
            <button
              className={styles['mobile-menu-close']}
              onClick={toggleMenu}
              aria-label="Close menu"
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
                Features
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
                  <Link href="/pos" onClick={closeMenu} className={styles['mobile-feature-link']}>
                    <span className={styles['mobile-feature-title']}>POS</span>
                    <span className={styles['mobile-feature-desc']}>
                      Payments, tables, orders, and receipts in one register.
                    </span>
                  </Link>
                  <Link href="/erp" onClick={closeMenu} className={styles['mobile-feature-link']}>
                    <span className={styles['mobile-feature-title']}>ERP</span>
                    <span className={styles['mobile-feature-desc']}>
                      Inventory, suppliers, and purchasing that stay in sync.
                    </span>
                  </Link>
                  <Link href="/hr" onClick={closeMenu} className={styles['mobile-feature-link']}>
                    <span className={styles['mobile-feature-title']}>HR</span>
                    <span className={styles['mobile-feature-desc']}>
                      Staff, schedules, time off, and payroll in one place.
                    </span>
                  </Link>
                  <Link href="/crm" onClick={closeMenu} className={styles['mobile-feature-link']}>
                    <span className={styles['mobile-feature-title']}>CRM</span>
                    <span className={styles['mobile-feature-desc']}>
                      Customer profiles, loyalty, segments, and campaigns.
                    </span>
                  </Link>
                  <Link
                    href="/finance"
                    onClick={closeMenu}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>Finance</span>
                    <span className={styles['mobile-feature-desc']}>
                      Invoices, expenses, tax, and a real P&L in seconds.
                    </span>
                  </Link>
                  <Link href="/flow" onClick={closeMenu} className={styles['mobile-feature-link']}>
                    <span className={styles['mobile-feature-title']}>Flow</span>
                    <span className={styles['mobile-feature-desc']}>
                      Projects, tasks, and team workspaces for every department.
                    </span>
                  </Link>
                </div>
              )}
            </div>

            <div className={styles['mobile-nav-item']}>
              <button
                type="button"
                onClick={toggleSolutionsDropdown}
                aria-expanded={solutionsDropdown}
              >
                Company
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
                  <Link href="/about" onClick={closeMenu}>
                    About
                  </Link>
                  <Link href="/partners" onClick={closeMenu}>
                    Partners
                  </Link>
                  <a
                    href="https://voxire.com/careers"
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={closeMenu}
                  >
                    Careers
                  </a>
                  <Link href="/contact" onClick={closeMenu}>
                    Contact
                  </Link>
                </div>
              )}
            </div>

            <Link href="/pricing" onClick={closeMenu} className={styles['mobile-nav-link']}>
              Pricing <HiChevronRight />
            </Link>

            <div className={styles['mobile-nav-item']}>
              <button
                type="button"
                onClick={toggleResourcesDropdown}
                aria-expanded={resourcesDropdown}
              >
                Resources
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
                  <Link href="/hardware" onClick={closeMenu}>
                    Hardware
                  </Link>
                  <Link href="/integrations" onClick={closeMenu}>
                    Integrations
                  </Link>
                  <Link href="/blog" onClick={closeMenu}>
                    Blog
                  </Link>
                  <a
                    href="https://docs.rtylr.com"
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={closeMenu}
                  >
                    Documentation
                  </a>
                  <Link href="/support" onClick={closeMenu}>
                    Help Center
                  </Link>
                  <Link href="/trust" onClick={closeMenu}>
                    Trust Center
                  </Link>
                  <Link href="/legal/security" onClick={closeMenu}>
                    Security
                  </Link>
                  <Link href="/legal/privacy-policy" onClick={closeMenu}>
                    Privacy Policy
                  </Link>
                  <Link href="/legal/terms" onClick={closeMenu}>
                    Terms of Use
                  </Link>
                  <a
                    href={APP_URLS.status}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={closeMenu}
                  >
                    Status
                  </a>
                </div>
              )}
            </div>

            <Link href="/contact" onClick={closeMenu} className={styles['mobile-nav-link']}>
              Contact <HiChevronRight />
            </Link>
          </div>

          <div className={styles['mobile-menu-footer']}>
            <div className={styles['mobile-menu-buttons']}>
              <Button onClick={handleLoginClick} variant="dark-outline" width="100%" height="44px">
                Sign In
              </Button>
              <Button onClick={handleSignupClick} variant="soft" width="100%" height="44px">
                Sign Up
              </Button>
            </div>
            {/* No language switcher until Phase 2 wires i18n. See the
                note in HeaderDesktop.tsx: display:none hid it from
                browsers and not from the markdown twins. */}
          </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
};

export default HeaderMobile;
