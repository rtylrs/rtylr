import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { cn } from '../../../../lib/cn';
import { HiChevronDown } from 'react-icons/hi';
import Link from 'next/link';

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

const FEATURE_LINKS = [
  { href: '/pos', title: 'POS', desc: 'Payments, tables, orders, and receipts in one register.' },
  { href: '/erp', title: 'ERP', desc: 'Inventory, suppliers, and purchasing that stay in sync.' },
  { href: '/hr', title: 'HR', desc: 'Staff, schedules, time off, and payroll in one place.' },
  { href: '/crm', title: 'CRM', desc: 'Customer profiles, loyalty, segments, and campaigns.' },
  {
    href: '/finance',
    title: 'Finance',
    desc: 'Invoices, expenses, tax, and a real P&L in seconds.',
  },
  {
    href: '/flow',
    title: 'Flow',
    desc: 'Projects, tasks, and team workspaces for every department.',
  },
];

const RESOURCE_LINKS = [
  { href: '/hardware', label: 'Hardware' },
  { href: '/integrations', label: 'Integrations' },
  { href: '/blog', label: 'Blog' },
  { href: 'https://docs.rtylr.com/changelog', label: 'Changelog', external: true },
  { href: 'https://docs.rtylr.com', label: 'Documentation', external: true },
  { href: '/support', label: 'Help Center' },
  { href: '/trust', label: 'Trust Center' },
  { href: '/legal/security', label: 'Security' },
  { href: '/legal/privacy-policy', label: 'Privacy Policy' },
  { href: '/legal/terms', label: 'Terms of Use' },
  { href: APP_URLS.status, label: 'Status', external: true },
];

const COMPANY_LINKS = [
  { href: '/about', label: 'About' },
  { href: '/partners', label: 'Partners' },
  { href: 'https://voxire.com/careers', label: 'Careers', external: true },
  { href: '/contact', label: 'Contact' },
];

const HeaderDesktop: React.FC = () => {
  const [scrolled, setScrolled] = useState(false);
  const [featuresDropdown, setFeaturesDropdown] = useState(false);
  const [solutionsDropdown, setSolutionsDropdown] = useState(false);
  const [resourcesDropdown, setResourcesDropdown] = useState(false);
  const [languageDropdown, setLanguageDropdown] = useState(false);
  const [language, setLanguage] = useState<string>('en');

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

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setFeaturesDropdown(false);
        setSolutionsDropdown(false);
        setResourcesDropdown(false);
        setLanguageDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
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

  const handleLoginClick = useCallback(() => {
    navigateExternal(APP_URLS.auth);
  }, []);

  const closeDropdowns = useCallback(() => {
    setFeaturesDropdown(false);
    setSolutionsDropdown(false);
    setResourcesDropdown(false);
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
        <Link href="/" className={styles['logo-container']} aria-label="rtylr home">
          <img src={ASSETS.logos.light} alt="rtylr" />
        </Link>

        <nav className={styles['nav-links']}>
          <div className={cn(styles['nav-item'], 'features-dropdown-container')}>
            <button
              type="button"
              className={styles['nav-link']}
              onClick={toggleFeaturesDropdown}
              aria-expanded={featuresDropdown}
              aria-haspopup="true"
            >
              Features
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: featuresDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </button>
            {featuresDropdown && (
              <div className={styles['header-dropdown-wide']}>
                {FEATURE_LINKS.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={styles['dropdown-item']}
                    onClick={closeDropdowns}
                  >
                    <div className={styles['dropdown-item-title']}>{item.title}</div>
                    <div className={styles['dropdown-item-desc']}>{item.desc}</div>
                  </Link>
                ))}
              </div>
            )}
          </div>

          <LinkWithColor to="/pricing" className={styles['nav-link']}>
            Pricing
          </LinkWithColor>

          <div className={cn(styles['nav-item'], 'resources-dropdown-container')}>
            <button
              type="button"
              className={styles['nav-link']}
              onClick={toggleResourcesDropdown}
              aria-expanded={resourcesDropdown}
              aria-haspopup="true"
            >
              Resources
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: resourcesDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </button>
            {resourcesDropdown && (
              <div className={styles['header-dropdown']}>
                {RESOURCE_LINKS.map((item) =>
                  item.external ? (
                    <a
                      key={item.label}
                      href={item.href}
                      className={styles['dropdown-link']}
                      target="_blank"
                      rel="noopener noreferrer"
                      onClick={closeDropdowns}
                    >
                      {item.label}
                    </a>
                  ) : (
                    <Link
                      key={item.label}
                      href={item.href}
                      className={styles['dropdown-link']}
                      onClick={closeDropdowns}
                    >
                      {item.label}
                    </Link>
                  )
                )}
              </div>
            )}
          </div>

          <div className={cn(styles['nav-item'], 'solutions-dropdown-container')}>
            <button
              type="button"
              className={styles['nav-link']}
              onClick={toggleSolutionsDropdown}
              aria-expanded={solutionsDropdown}
              aria-haspopup="true"
            >
              Company
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: solutionsDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </button>
            {solutionsDropdown && (
              <div className={styles['header-dropdown']}>
                {COMPANY_LINKS.map((item) =>
                  item.external ? (
                    <a
                      key={item.label}
                      href={item.href}
                      className={styles['dropdown-link']}
                      target="_blank"
                      rel="noopener noreferrer"
                      onClick={closeDropdowns}
                    >
                      {item.label}
                    </a>
                  ) : (
                    <Link
                      key={item.label}
                      href={item.href}
                      className={styles['dropdown-link']}
                      onClick={closeDropdowns}
                    >
                      {item.label}
                    </Link>
                  )
                )}
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
