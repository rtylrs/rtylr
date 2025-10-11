import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { HiMenu, HiX, HiChevronDown } from 'react-icons/hi';
import Link from 'next/link';
import { useRouter } from 'next/router';

import { Button, APP_URLS, ASSETS } from '@voxire/shared';
import useColors from '../../../hooks/useColors';
import { useWindowDimensions } from '@voxire/shared/utils';
import { selectLanguage } from '../../../services/i18next';
import { getCookie } from '../../../lib/cookies';
import { useUserStore } from '../../../stores/userStore';
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
  const { primary } = useColors();
  const { width } = useWindowDimensions();
  const user = useUserStore();

  const isMobile = useMemo(() => width < 968, [width]);
  const isQuestionnaire = useMemo(
    () =>
      typeof window !== 'undefined' &&
      router.pathname.includes('/auth/questionnaire'),
    [router.pathname]
  );
  const isInnerPage = useMemo(() => {
    if (typeof window === 'undefined') return false;
    return document.location.pathname !== '/';
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    let ticking = false;

    const handleScroll = () => {
      if (!ticking) {
        requestAnimationFrame(() => {
          const shouldBeScrolled = window.scrollY > 150;
          setScrolled(shouldBeScrolled);
          ticking = false;
        });
        ticking = true;
      }
    };

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
    if (typeof window !== 'undefined') {
      window.location.href = `${APP_URLS.auth}/register`;
    }
  }, []);

  const handleLogoClick = useCallback(() => {
    router.push('/');
  }, [router.push]);

  const handleLoginClick = useCallback(() => {
    if (typeof window !== 'undefined') {
      window.location.href = APP_URLS.auth;
    }
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
    if (typeof window !== 'undefined') {
      window.open('https://voxire.com/careers', '_blank');
    }
  }, []);

  const LinkWithColor: React.FC<LinkWithColorProps> = useMemo(() => {
    return (props) => {
      const { color, to, className, ...rest } = props;
      // If no 'to' prop, return a clickable element but not a Link
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
    if (scrolled) classes += ` ${styles.scrolled}`;
    if (isInnerPage) classes += ` ${styles.inner}`;
    return classes;
  }, [isMobileMenuOpen, scrolled, isInnerPage]);

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
        <button
          className={styles['mobile-menu-toggle']}
          onClick={toggleMobileMenu}
          aria-label="Toggle mobile menu"
        >
          {isMobileMenuOpen ? <HiX /> : <HiMenu />}
        </button>

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

      {isMobileMenuOpen && (
        <div className={styles['mobile-menu']}>
          <LinkWithColor to="/menu">Menu</LinkWithColor>
          <LinkWithColor to="/deals">Deals</LinkWithColor>
          <LinkWithColor to="/blog">Blog</LinkWithColor>
          <LinkWithColor to="/about">About</LinkWithColor>
          <LinkWithColor to="/contact">Contact Us</LinkWithColor>
          <Button
            styles={{
              display: 'flex',
              marginBottom: '0px',
            }}
            onClick={handleLoginClick}
          >
            <p>Login</p>
          </Button>

          <Button onClick={handleSignupClick}>Register Now</Button>
        </div>
      )}
    </header>
  );
};

export default Header;
