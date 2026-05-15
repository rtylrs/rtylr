import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { HiMenu, HiX, HiChevronDown, HiChevronRight } from 'react-icons/hi';
import { Button } from '@voxire/components';
import { APP_URLS, ASSETS } from '@voxire/config';
import { selectLanguage } from '../../../../services/i18next';
import { getCookie } from '../../../../lib/cookies';
import { navigateExternal } from '../../../../lib/navigation';
import styles from './HeaderMobile.module.scss';

const HeaderMobile: React.FC = () => {
  const [isMenuOpen, setIsMenuOpen] = useState<boolean>(false);
  const [scrolled, setScrolled] = useState(false);
  const [featuresDropdown, setFeaturesDropdown] = useState(false);
  const [solutionsDropdown, setSolutionsDropdown] = useState(false);
  const [resourcesDropdown, setResourcesDropdown] = useState(false);
  const [language, setLanguage] = useState<string>('en');

  const router = useRouter();

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

  useEffect(() => {
    const storedLanguage = getCookie('language');
    if (storedLanguage) {
      setLanguage(storedLanguage);
    }
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

  const handleLanguageChange = useCallback((lang: string) => {
    selectLanguage(lang);
    setLanguage(lang);
  }, []);

  const handleLoginClick = useCallback(() => {
    navigateExternal(APP_URLS.auth);
    setIsMenuOpen(false);
  }, []);

  const handleSignupClick = useCallback(() => {
    navigateExternal(`${APP_URLS.auth}/register`);
    setIsMenuOpen(false);
  }, []);

  const handleNavigation = useCallback(
    (path: string) => {
      setFeaturesDropdown(false);
      setSolutionsDropdown(false);
      setResourcesDropdown(false);
      setIsMenuOpen(false);
      router.push(path);
    },
    [router]
  );

  const handleExternalLink = useCallback((url: string) => {
    setResourcesDropdown(false);
    setIsMenuOpen(false);
    window.open(url, '_blank', 'noopener,noreferrer');
  }, []);

  const handleLogoClick = useCallback(() => {
    setIsMenuOpen(false);
    router.push('/');
  }, [router]);

  const headerClasses = `${styles.header} ${styles['mobile-header']} ${
    !scrolled ? styles['initial-position'] : styles.scrolled
  }`;

  return (
    <>
      <header className={headerClasses}>
        <div className={styles['mobile-header-content']}>
          <div className={styles['logo-container']} onClick={handleLogoClick}>
            <img src={ASSETS.logos.light} alt="Logo" />
          </div>

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

      {isMenuOpen && (
        <div className={styles['mobile-menu-overlay']}>
          <div className={styles['mobile-menu-header']}>
            <div className={styles['mobile-menu-logo']} onClick={handleLogoClick}>
              <img src={ASSETS.logos.light} alt="Logo" />
            </div>
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
              <p onClick={toggleFeaturesDropdown}>
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
              </p>
              {featuresDropdown && (
                <div className={styles['mobile-dropdown']}>
                  <Link
                    href="/pos"
                    onClick={() => handleNavigation('/pos')}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>POS</span>
                    <span className={styles['mobile-feature-desc']}>
                      Payments, tables, orders, and receipts in one register.
                    </span>
                  </Link>
                  <Link
                    href="/erp"
                    onClick={() => handleNavigation('/erp')}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>ERP</span>
                    <span className={styles['mobile-feature-desc']}>
                      Inventory, suppliers, and purchasing that stay in sync.
                    </span>
                  </Link>
                  <Link
                    href="/hr"
                    onClick={() => handleNavigation('/hr')}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>HR</span>
                    <span className={styles['mobile-feature-desc']}>
                      Staff, schedules, time off, and payroll in one place.
                    </span>
                  </Link>
                  <Link
                    href="/crm"
                    onClick={() => handleNavigation('/crm')}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>CRM</span>
                    <span className={styles['mobile-feature-desc']}>
                      Customer profiles, loyalty, segments, and campaigns.
                    </span>
                  </Link>
                  <Link
                    href="/finance"
                    onClick={() => handleNavigation('/finance')}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>Finance</span>
                    <span className={styles['mobile-feature-desc']}>
                      Invoices, expenses, tax, and a real P&L in seconds.
                    </span>
                  </Link>
                  <Link
                    href="/flow"
                    onClick={() => handleNavigation('/flow')}
                    className={styles['mobile-feature-link']}
                  >
                    <span className={styles['mobile-feature-title']}>Flow</span>
                    <span className={styles['mobile-feature-desc']}>
                      Projects, tasks, and team workspaces for every department.
                    </span>
                  </Link>
                </div>
              )}
            </div>

            <div className={styles['mobile-nav-item']}>
              <p onClick={toggleSolutionsDropdown}>
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
              </p>
              {solutionsDropdown && (
                <div className={styles['mobile-dropdown']}>
                  <Link href="/about" onClick={() => handleNavigation('/about')}>
                    About
                  </Link>
                  <Link href="/partners" onClick={() => handleNavigation('/partners')}>
                    Partners
                  </Link>
                  <a
                    href="#"
                    onClick={(e) => {
                      e.preventDefault();
                      handleExternalLink('https://voxire.com/careers');
                    }}
                  >
                    Careers
                  </a>
                  <Link href="/contact" onClick={() => handleNavigation('/contact')}>
                    Contact
                  </Link>
                </div>
              )}
            </div>

            <Link
              href="/pricing"
              onClick={() => handleNavigation('/pricing')}
              className={styles['mobile-nav-link']}
            >
              Pricing <HiChevronRight />
            </Link>

            <div className={styles['mobile-nav-item']}>
              <p onClick={toggleResourcesDropdown}>
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
              </p>
              {resourcesDropdown && (
                <div className={styles['mobile-dropdown']}>
                  <Link href="/hardware" onClick={() => handleNavigation('/hardware')}>
                    Hardware
                  </Link>
                  <Link href="/integrations" onClick={() => handleNavigation('/integrations')}>
                    Integrations
                  </Link>
                  <Link href="/blog" onClick={() => handleNavigation('/blog')}>
                    Blog
                  </Link>
                  <Link href="/changelog" onClick={() => handleNavigation('/changelog')}>
                    Changelog
                  </Link>
                  <a
                    href="#"
                    onClick={(e) => {
                      e.preventDefault();
                      handleExternalLink('https://docs.rtylr.com');
                    }}
                  >
                    Documentation
                  </a>
                  <Link href="/support" onClick={() => handleNavigation('/support')}>
                    Help Center
                  </Link>
                  <Link href="/trust" onClick={() => handleNavigation('/trust')}>
                    Trust Center
                  </Link>
                  <a
                    href="#"
                    onClick={(e) => {
                      e.preventDefault();
                      handleExternalLink(APP_URLS.status);
                    }}
                  >
                    Status
                  </a>
                </div>
              )}
            </div>

            <Link
              href="/contact"
              onClick={() => handleNavigation('/contact')}
              className={styles['mobile-nav-link']}
            >
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
            <div className={styles['mobile-language-selector']} style={{ display: 'none' }}>
              <span
                onClick={() => handleLanguageChange('en')}
                className={language === 'en' ? styles.active : ''}
              >
                English
              </span>
              <span
                onClick={() => handleLanguageChange('ar')}
                className={language === 'ar' ? styles.active : ''}
              >
                عربي
              </span>
              <span
                onClick={() => handleLanguageChange('fr')}
                className={language === 'fr' ? styles.active : ''}
              >
                Français
              </span>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default HeaderMobile;
