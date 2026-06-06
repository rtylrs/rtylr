import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { cn } from '../../../../lib/cn';
import { HiChevronDown } from 'react-icons/hi';
import Link from 'next/link';
import { useRouter } from 'next/router';

import { Button } from '@voxire/components';
import { APP_URLS, ASSETS } from '@voxire/config';
import { selectLanguage } from '../../../../services/i18next';
import { getCookie } from '../../../../lib/cookies';
import { navigateExternal } from '../../../../lib/navigation';
import styles from './HeaderDesktop.module.scss';

interface LinkWithColorProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  color?: string;
  to: string;
}

const HeaderDesktop: React.FC = () => {
  const [scrolled, setScrolled] = useState(false);
  const [featuresDropdown, setFeaturesDropdown] = useState(false);
  const [solutionsDropdown, setSolutionsDropdown] = useState(false);
  const [resourcesDropdown, setResourcesDropdown] = useState(false);
  const [languageDropdown, setLanguageDropdown] = useState(false);
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

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;

      if (featuresDropdown && !target.closest('.features-dropdown-container')) {
        setFeaturesDropdown(false);
      }

      if (solutionsDropdown && !target.closest('.solutions-dropdown-container')) {
        setSolutionsDropdown(false);
      }

      if (resourcesDropdown && !target.closest('.resources-dropdown-container')) {
        setResourcesDropdown(false);
      }

      if (languageDropdown && !target.closest('.lang-dropdown-container')) {
        setLanguageDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [featuresDropdown, solutionsDropdown, resourcesDropdown, languageDropdown]);

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

  const toggleLanguageDropdown = useCallback(() => {
    setLanguageDropdown((prev) => !prev);
  }, []);

  const handleLanguageChange = useCallback((lang: string) => {
    selectLanguage(lang);
    setLanguage(lang);
    setLanguageDropdown(false);
  }, []);

  const handleSignupClick = useCallback(() => {
    navigateExternal(`${APP_URLS.auth}/register`);
  }, []);

  const handleLogoClick = useCallback(() => {
    router.push('/');
  }, [router]);

  const handleLoginClick = useCallback(() => {
    navigateExternal(APP_URLS.auth);
  }, []);

  const handleNavigation = useCallback(
    (path: string) => {
      setFeaturesDropdown(false);
      setSolutionsDropdown(false);
      setResourcesDropdown(false);
      router.push(path);
    },
    [router]
  );

  const prefetchRoute = useCallback(
    (path: string) => {
      router.prefetch(path).catch(() => {});
    },
    [router]
  );

  const handleExternalLink = useCallback((url: string) => {
    setResourcesDropdown(false);
    window.open(url, '_blank', 'noopener,noreferrer');
  }, []);

  const LinkWithColor: React.FC<LinkWithColorProps> = useMemo(() => {
    return (props) => {
      const { color, to, className, ...rest } = props;
      if (!to) {
        return (
          <a
            href="#"
            className={className}
            style={{
              color: color || 'white',
              textDecoration: 'none',
            }}
            onClick={(e) => e.preventDefault()}
            {...rest}
          />
        );
      }
      return (
        <Link
          href={to}
          className={className}
          style={{
            color: color || 'white',
            textDecoration: 'none',
          }}
          {...rest}
        />
      );
    };
  }, []);

  const headerClasses = useMemo(() => {
    let classes = `${styles.header} ${styles['desktop-header']}`;
    if (!scrolled) {
      classes += ` ${styles['initial-position']}`;
    } else {
      classes += ` ${styles.scrolled}`;
    }
    return classes;
  }, [scrolled]);

  return (
    <header className={headerClasses}>
      <div className={styles['header-left']}>
        <div className={styles['logo-container']} onClick={handleLogoClick}>
          <img src={ASSETS.logos.light} alt="Logo" />
        </div>

        <nav className={styles['nav-links']}>
          <div className={cn(styles['nav-item'], 'features-dropdown-container')}>
            <p className={styles['nav-link']} onClick={toggleFeaturesDropdown}>
              Features
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: featuresDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </p>
            {featuresDropdown && (
              <div className={styles['header-dropdown-wide']}>
                <div
                  className={styles['dropdown-item']}
                  onClick={() => handleNavigation('/pos')}
                  onMouseEnter={() => prefetchRoute('/pos')}
                >
                  <div className={styles['dropdown-item-title']}>POS</div>
                  <div className={styles['dropdown-item-desc']}>
                    Payments, tables, orders, and receipts in one register.
                  </div>
                </div>
                <div
                  className={styles['dropdown-item']}
                  onClick={() => handleNavigation('/erp')}
                  onMouseEnter={() => prefetchRoute('/erp')}
                >
                  <div className={styles['dropdown-item-title']}>ERP</div>
                  <div className={styles['dropdown-item-desc']}>
                    Inventory, suppliers, and purchasing that stay in sync.
                  </div>
                </div>
                <div
                  className={styles['dropdown-item']}
                  onClick={() => handleNavigation('/hr')}
                  onMouseEnter={() => prefetchRoute('/hr')}
                >
                  <div className={styles['dropdown-item-title']}>HR</div>
                  <div className={styles['dropdown-item-desc']}>
                    Staff, schedules, time off, and payroll in one place.
                  </div>
                </div>
                <div
                  className={styles['dropdown-item']}
                  onClick={() => handleNavigation('/crm')}
                  onMouseEnter={() => prefetchRoute('/crm')}
                >
                  <div className={styles['dropdown-item-title']}>CRM</div>
                  <div className={styles['dropdown-item-desc']}>
                    Customer profiles, loyalty, segments, and campaigns.
                  </div>
                </div>
                <div
                  className={styles['dropdown-item']}
                  onClick={() => handleNavigation('/finance')}
                  onMouseEnter={() => prefetchRoute('/finance')}
                >
                  <div className={styles['dropdown-item-title']}>Finance</div>
                  <div className={styles['dropdown-item-desc']}>
                    Invoices, expenses, tax, and a real P&L in seconds.
                  </div>
                </div>
                <div
                  className={styles['dropdown-item']}
                  onClick={() => handleNavigation('/flow')}
                  onMouseEnter={() => prefetchRoute('/flow')}
                >
                  <div className={styles['dropdown-item-title']}>Flow</div>
                  <div className={styles['dropdown-item-desc']}>
                    Projects, tasks, and team workspaces for every department.
                  </div>
                </div>
              </div>
            )}
          </div>

          <LinkWithColor to="/pricing" className={styles['nav-link']}>
            Pricing
          </LinkWithColor>

          <div className={cn(styles['nav-item'], 'resources-dropdown-container')}>
            <p className={styles['nav-link']} onClick={toggleResourcesDropdown}>
              Resources
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: resourcesDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </p>
            {resourcesDropdown && (
              <div className={styles['header-dropdown']}>
                <h6
                  onClick={() => handleNavigation('/hardware')}
                  onMouseEnter={() => prefetchRoute('/hardware')}
                >
                  Hardware
                </h6>
                <h6
                  onClick={() => handleNavigation('/integrations')}
                  onMouseEnter={() => prefetchRoute('/integrations')}
                >
                  Integrations
                </h6>
                <h6
                  onClick={() => handleNavigation('/blog')}
                  onMouseEnter={() => prefetchRoute('/blog')}
                >
                  Blog
                </h6>
                <h6 onClick={() => handleExternalLink('https://docs.rtylr.com/changelog')}>
                  Changelog
                </h6>
                <h6 onClick={() => handleExternalLink('https://docs.rtylr.com')}>Documentation</h6>
                <h6
                  onClick={() => handleNavigation('/support')}
                  onMouseEnter={() => prefetchRoute('/support')}
                >
                  Help Center
                </h6>
                <h6
                  onClick={() => handleNavigation('/trust')}
                  onMouseEnter={() => prefetchRoute('/trust')}
                >
                  Trust Center
                </h6>
                <h6
                  onClick={() => handleNavigation('/legal/security')}
                  onMouseEnter={() => prefetchRoute('/legal/security')}
                >
                  Security
                </h6>
                <h6
                  onClick={() => handleNavigation('/legal/privacy-policy')}
                  onMouseEnter={() => prefetchRoute('/legal/privacy-policy')}
                >
                  Privacy Policy
                </h6>
                <h6
                  onClick={() => handleNavigation('/legal/terms')}
                  onMouseEnter={() => prefetchRoute('/legal/terms')}
                >
                  Terms of Use
                </h6>
                <h6 onClick={() => handleExternalLink(APP_URLS.status)}>Status</h6>
              </div>
            )}
          </div>

          <div className={cn(styles['nav-item'], 'solutions-dropdown-container')}>
            <p className={styles['nav-link']} onClick={toggleSolutionsDropdown}>
              Company
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: solutionsDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </p>
            {solutionsDropdown && (
              <div className={styles['header-dropdown']}>
                <h6
                  onClick={() => handleNavigation('/about')}
                  onMouseEnter={() => prefetchRoute('/about')}
                >
                  About
                </h6>
                <h6
                  onClick={() => handleNavigation('/partners')}
                  onMouseEnter={() => prefetchRoute('/partners')}
                >
                  Partners
                </h6>
                <h6 onClick={() => handleExternalLink('https://voxire.com/careers')}>Careers</h6>
                <h6
                  onClick={() => handleNavigation('/contact')}
                  onMouseEnter={() => prefetchRoute('/contact')}
                >
                  Contact
                </h6>
              </div>
            )}
          </div>
        </nav>
      </div>

      <div className={styles['header-right']}>
        <div
          className={cn(styles['language-selector-container'], 'lang-dropdown-container')}
          style={{ display: 'none' }}
        >
          <span className={styles['language-selector']} onClick={toggleLanguageDropdown}>
            {language === 'en' ? 'English' : language === 'ar' ? 'عربي' : 'Français'}
          </span>
          {languageDropdown && (
            <div className={styles['languages-dropdown']}>
              <span onClick={() => handleLanguageChange('en')}>English</span>
              <span onClick={() => handleLanguageChange('ar')}>عربي</span>
              <span onClick={() => handleLanguageChange('fr')}>Français</span>
            </div>
          )}
        </div>

        <div className={styles['header-buttons']}>
          <Button onClick={handleLoginClick} variant="soft" width="140px" height="40px">
            Sign In
          </Button>
          <Button onClick={handleSignupClick} variant="dark-outline" width="140px" height="40px">
            Sign Up
          </Button>
        </div>
      </div>
    </header>
  );
};

export default HeaderDesktop;
