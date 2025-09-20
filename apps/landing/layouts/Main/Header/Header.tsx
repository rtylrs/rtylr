import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faBars,
  faChevronDown,
  faRightFromBracket,
  faTimes,
} from '@fortawesome/free-solid-svg-icons';
import Link from 'next/link';
import { useRouter } from 'next/router';

import { Button } from '@voxire/shared';
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
    () => typeof window !== 'undefined' && router.pathname.includes('/auth/questionnaire'),
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

  const handleClick = useCallback(() => {
    const token = localStorage.getItem('token');

    if (token) {
      if (user.role === 'business' || user.role === 'admin') {
        router.push('/app/business/');
      } else {
        router.push('/app/user/');
      }
    } else {
      if (typeof window !== 'undefined') {
        window.location.href = 'https://auth.rtylr.com/register';
      }
    }
  }, [user.role, router.push]);

  const handleLogoClick = useCallback(() => {
    router.push('/');
  }, [router.push]);

  const handleLoginClick = useCallback(() => {
    if (typeof window !== 'undefined') {
      window.location.href = 'https://auth.rtylr.com/';
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
      const { color, to, ...rest } = props;
      // If no 'to' prop, return a clickable element but not a Link
      if (!to) {
        return (
          <a
            href="#"
            style={{
              color: color || primary,
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
          style={{
            color: color || primary,
            textDecoration: 'none',
          }}
          {...rest}
        />
      );
    };
  }, [primary]);

  const headerClasses = useMemo(() => {
    let classes = styles.header;
    if (isMobileMenuOpen) classes += ` ${styles['mobile-menu-open']}`;
    if (scrolled) classes += ` ${styles.scrolled}`;
    if (isInnerPage) classes += ` ${styles.inner}`;
    return classes;
  }, [isMobileMenuOpen, scrolled, isInnerPage]);

  return (
    <header
      className={headerClasses}
      style={{
        position: 'fixed',
      }}
    >
      <div
        className={styles['header-center']}
        style={{
          cursor: 'pointer',
        }}
        onClick={handleLogoClick}
      >
        <img src={'/logo.png'} alt="Logo" />
      </div>

      <div
        className={styles['header-left']}
        style={{
          transform: 'translateX(5%)',
        }}
      >
        <LinkWithColor to="/restaurant/69">Demo</LinkWithColor>
        <LinkWithColor to="/pricing">Pricing</LinkWithColor>
        <p
          style={{
            color: primary,
          }}
          onClick={toggleDropdown}
        >
          Company
          <FontAwesomeIcon
            icon={faChevronDown}
            style={{
              marginLeft: '5px',
              fontSize: '14px',
              transform: dropdown ? 'rotate(180deg)' : 'rotate(0deg)',
              transition: 'transform 0.2s ease',
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

      <div className={styles['header-right']}>
        <button
          className={styles['mobile-menu-toggle']}
          onClick={toggleMobileMenu}
          style={{
            color: primary,
          }}
          aria-label="Toggle mobile menu"
        >
          <FontAwesomeIcon
            icon={isMobileMenuOpen ? faTimes : faBars}
            style={{
              transition: 'transform 0.2s ease',
            }}
          />
        </button>

        <a
          href="https://auth.rtylr.com/"
          style={{
            color: primary,
            textDecoration: 'none',
            fontWeight: '600',
            fontSize: '16px',
            marginRight: '20px',
            cursor: 'pointer',
          }}
        >
          Login
        </a>

        <Button
          onClick={handleClick}
          variant="primary"
        >
          Register Now
        </Button>

        <div className={styles['lang-dropdown-container']}>
          <div className={styles['selected-language']} onClick={toggleLanguageDropdown}>
            {language}
            <FontAwesomeIcon
              icon={faChevronDown}
              style={{
                transform: languageDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                transition: 'transform 0.2s ease',
              }}
            />
          </div>
          {languageDropdown && (
            <div className={styles['languages-dropdown']}>
              <span onClick={() => handleLanguageChange('en')}>English</span>
              <span onClick={() => handleLanguageChange('ar')}>Arabic</span>
              <span onClick={() => handleLanguageChange('fr')}>French</span>
            </div>
          )}
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

          <Button onClick={handleClick}>
            Register Now
          </Button>
        </div>
      )}
    </header>
  );
};

export default Header;
