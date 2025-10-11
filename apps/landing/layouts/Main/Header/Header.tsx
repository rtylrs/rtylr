import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { HiChevronDown } from 'react-icons/hi';
import Link from 'next/link';
import { useRouter } from 'next/router';

import { Button, APP_URLS, ASSETS } from '@voxire/shared';
import { selectLanguage } from '../../../services/i18next';
import { getCookie } from '../../../lib/cookies';
import { navigateExternal } from '../../../lib/navigation';
import HeaderMobile from './HeaderMobile';
import styles from './Header.module.scss';

interface LinkWithColorProps
  extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  color?: string;
  to: string;
}

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState<boolean>(false);
  const [scrolled, setScrolled] = useState(false);
  const [dropdown, setDropdown] = useState(false);
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

      if (dropdown && !target.closest('.header-left')) {
        setDropdown(false);
      }

      if (languageDropdown && !target.closest('.lang-dropdown-container')) {
        setLanguageDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [dropdown, languageDropdown]);

  const toggleMobileMenu = useCallback(() => {
    setIsMobileMenuOpen((prev) => !prev);
  }, []);

  const toggleDropdown = useCallback(() => {
    setDropdown((prev) => !prev);
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
  }, [router.push]);

  const handleLoginClick = useCallback(() => {
    navigateExternal(APP_URLS.auth);
  }, []);

  const handleNavigation = useCallback(
    (path: string) => {
      setDropdown(false);
      router.push(path);
    },
    [router.push]
  );

  const handleExternalLink = useCallback(() => {
    setDropdown(false);
    window.open('https://voxire.com/careers', '_blank', 'noopener,noreferrer');
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
    let classes = styles.header;
    if (isMobileMenuOpen) classes += ` ${styles['mobile-menu-open']}`;
    if (!scrolled) {
      classes += ` ${styles['initial-position']}`;
    } else {
      classes += ` ${styles.scrolled}`;
    }
    return classes;
  }, [isMobileMenuOpen, scrolled]);

  return (
    <header className={headerClasses}>
      <div className={styles['header-left']}>
        <div className={styles['logo-container']} onClick={handleLogoClick}>
          <img src={ASSETS.logos.light} alt="Logo" />
        </div>

        <nav className={styles['nav-links']}>
          <div className={styles['nav-item']}>
            <p className={styles['nav-link']} onClick={toggleDropdown}>
              Features
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: dropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                }}
              />
            </p>
            {dropdown && (
              <div className={styles['header-dropdown']}>
                <h6 onClick={() => handleNavigation('/about')}>About Us</h6>
                <h6 onClick={() => handleNavigation('/contact')}>Contact Us</h6>
                <h6 onClick={handleExternalLink}>Careers (Jobs)</h6>
              </div>
            )}
          </div>

          <LinkWithColor to="/pricing" className={styles['nav-link']}>
            Pricing
          </LinkWithColor>
          <LinkWithColor to="/usecases" className={styles['nav-link']}>
            Use Cases
          </LinkWithColor>
          <LinkWithColor to="/contact" className={styles['nav-link']}>
            Contact
          </LinkWithColor>
          <LinkWithColor to="/faq" className={styles['nav-link']}>
            FAQ
          </LinkWithColor>
        </nav>
      </div>

      <div className={styles['header-right']}>
        <HeaderMobile
          isOpen={isMobileMenuOpen}
          onToggle={toggleMobileMenu}
          onLoginClick={handleLoginClick}
          onSignupClick={handleSignupClick}
        />

        <div className={styles['language-selector-container']}>
          <span
            className={styles['language-selector']}
            onClick={toggleLanguageDropdown}
          >
            {language === 'en'
              ? 'English'
              : language === 'ar'
              ? 'عربي'
              : 'Français'}
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
          <Button
            onClick={handleLoginClick}
            variant="soft"
            width="140px"
            height="40px"
          >
            Sign In
          </Button>
          <Button
            onClick={handleSignupClick}
            variant="dark-outline"
            width="140px"
            height="40px"
          >
            Sign Up
          </Button>
        </div>
      </div>
    </header>
  );
};

export default Header;
