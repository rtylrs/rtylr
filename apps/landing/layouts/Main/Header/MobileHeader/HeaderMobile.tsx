import React, { useState, useEffect, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import { HiMenu, HiX, HiChevronDown, HiChevronRight } from 'react-icons/hi';
import { Button, APP_URLS, ASSETS } from '@voxire/shared';
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
            <div
              className={styles['mobile-menu-logo']}
              onClick={handleLogoClick}
            >
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
                    href="/features/pos"
                    onClick={() => handleNavigation('/features/pos')}
                  >
                    POS & Payments
                  </Link>
                  <Link
                    href="/features/inventory"
                    onClick={() => handleNavigation('/features/inventory')}
                  >
                    Inventory Management
                  </Link>
                  <Link
                    href="/features/ecommerce"
                    onClick={() => handleNavigation('/features/ecommerce')}
                  >
                    eCommerce & Online Ordering
                  </Link>
                  <Link
                    href="/features/kds"
                    onClick={() => handleNavigation('/features/kds')}
                  >
                    Kitchen Display (KDS)
                  </Link>
                  <Link
                    href="/features/analytics"
                    onClick={() => handleNavigation('/features/analytics')}
                  >
                    Analytics & AI
                  </Link>
                  <Link
                    href="/features/offline"
                    onClick={() => handleNavigation('/features/offline')}
                  >
                    Offline Sync
                  </Link>
                </div>
              )}
            </div>

            <div className={styles['mobile-nav-item']}>
              <p onClick={toggleSolutionsDropdown}>
                Solutions
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
                  <Link
                    href="/solutions/food-beverage"
                    onClick={() => handleNavigation('/solutions/food-beverage')}
                  >
                    Food & Beverage
                  </Link>
                  <Link
                    href="/solutions/retail"
                    onClick={() => handleNavigation('/solutions/retail')}
                  >
                    Retail & Omnichannel
                  </Link>
                  <Link
                    href="/solutions/services"
                    onClick={() => handleNavigation('/solutions/services')}
                  >
                    Services & Appointments
                  </Link>
                  <Link
                    href="/solutions/grocery"
                    onClick={() => handleNavigation('/solutions/grocery')}
                  >
                    Grocery & Convenience
                  </Link>
                  <Link
                    href="/solutions/wholesale"
                    onClick={() => handleNavigation('/solutions/wholesale')}
                  >
                    Wholesale & B2B
                  </Link>
                  <Link
                    href="/solutions/events"
                    onClick={() => handleNavigation('/solutions/events')}
                  >
                    Events & Hospitality
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

            <Link
              href="/hardware"
              onClick={() => handleNavigation('/hardware')}
              className={styles['mobile-nav-link']}
            >
              Hardware <HiChevronRight />
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
                  <Link href="/blog" onClick={() => handleNavigation('/blog')}>
                    Blog
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
                  <Link
                    href="/about"
                    onClick={() => handleNavigation('/about')}
                  >
                    About
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
                  <Link
                    href="/support"
                    onClick={() => handleNavigation('/support')}
                  >
                    Support
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
              <Button
                onClick={handleLoginClick}
                variant="dark-outline"
                width="100%"
                height="44px"
              >
                Sign In
              </Button>
              <Button
                onClick={handleSignupClick}
                variant="soft"
                width="100%"
                height="44px"
              >
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
