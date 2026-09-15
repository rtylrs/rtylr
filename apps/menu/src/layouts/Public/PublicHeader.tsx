import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { FaBars, FaChevronDown, FaSignOutAlt, FaTimes } from 'react-icons/fa';
import { Link, useNavigate, useRouterState } from '@tanstack/react-router';

import { APP_URLS } from '@voxire/config';
import { Button } from '@voxire/components';
import useColors from '../../shared/hooks/useColors';
import { useWindowDimensions } from '@voxire/utils';
import { selectLanguage } from '../../shared/services/i18next';
import { getCookie } from '../../shared/lib/cookies';

interface LinkWithColorProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  color?: string;
  to: string;
}

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState<boolean>(false);
  const [scrolled, setScrolled] = useState(false);
  const [dropdown, setDropdown] = useState(false);
  const [languageDropdown, setLanguageDropdown] = useState(false);
  const [language, setLanguage] = useState<string>('en');

  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const navigate = useNavigate();
  const { primary } = useColors();
  const { width } = useWindowDimensions();
  const token = getCookie('token');

  const isMobile = useMemo(() => width < 968, [width]);
  const isQuestionnaire = useMemo(() => pathname.includes('/auth/questionnaire'), [pathname]);
  const isInnerPage = useMemo(() => document.location.pathname !== '/', []);

  useEffect(() => {
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
    navigate({ to: '/' });
  }, [navigate]);

  const handleLogoClick = useCallback(() => {
    navigate({ to: '/' });
  }, [navigate]);

  const handleLoginClick = useCallback(() => {
    window.location.assign(`${APP_URLS.auth}/login`);
  }, []);

  const landingUrl = (APP_URLS.landing || 'https://rtylr.com').replace(/\/$/, '');

  const handleNavigation = useCallback(
    (path: string) => {
      setDropdown(false);
      window.open(`${landingUrl}${path}`, '_blank', 'noopener,noreferrer');
    },
    [landingUrl]
  );

  const handleExternalLink = useCallback(() => {
    setDropdown(false);
    window.open('https://hire.voxire.com/jobs/', '_blank', 'noopener,noreferrer');
  }, []);

  const externalLinkStyle = (color?: string): React.CSSProperties => ({
    color: color || primary,
  });

  const LinkWithColor: React.FC<LinkWithColorProps> = useMemo(() => {
    return (props) => {
      const { color, to, ...rest } = props;
      return (
        <Link
          style={{
            color: color || primary,
          }}
          to={to}
          {...rest}
        />
      );
    };
  }, [primary]);

  const headerClasses = useMemo(() => {
    const classes = ['header'];
    if (isMobileMenuOpen) classes.push('mobile-menu-open');
    if (scrolled) classes.push('scrolled');
    if (isInnerPage) classes.push('inner');
    return classes.join(' ');
  }, [isMobileMenuOpen, scrolled, isInnerPage]);

  return (
    <header
      className={headerClasses}
      style={{
        position: 'fixed',
      }}
    >
      <div
        className="header-center"
        style={{
          cursor: 'pointer',
        }}
        onClick={handleLogoClick}
      >
        <img src="/logo-dark.png" alt="rtylr" />
      </div>

      <div
        className="header-left"
        style={{
          transform: 'translateX(5%)',
        }}
      >
        <LinkWithColor to="/b/69">Demo</LinkWithColor>
        <a
          href={`${landingUrl}/pricing`}
          target="_blank"
          rel="noopener noreferrer"
          style={externalLinkStyle()}
        >
          Pricing
        </a>
        <a
          href={`${landingUrl}/pos`}
          target="_blank"
          rel="noopener noreferrer"
          style={externalLinkStyle()}
        >
          Restaurants
        </a>
        <a
          href={`${landingUrl}/pos`}
          target="_blank"
          rel="noopener noreferrer"
          style={externalLinkStyle()}
        >
          Use Cases
        </a>
        <p
          style={{
            color: primary,
          }}
          onClick={toggleDropdown}
        >
          Company
          <FaChevronDown
            style={{
              marginLeft: '5px',
              fontSize: '14px',
              transform: dropdown ? 'rotate(180deg)' : 'rotate(0deg)',
              transition: 'transform 0.2s ease',
            }}
          />
        </p>

        {dropdown && (
          <div className="header-dropdown">
            <h6 onClick={() => handleNavigation('/about')}>About Us</h6>
            <h6 onClick={() => handleNavigation('/contact')}>Contact Us</h6>
            <h6 onClick={() => handleNavigation('/support')}>FAQs</h6>
            <h6 onClick={handleExternalLink}>Careers (Jobs)</h6>
          </div>
        )}
      </div>

      <div className="header-right">
        <button
          type="button"
          className="mobile-menu-toggle"
          onClick={toggleMobileMenu}
          style={{
            color: primary,
          }}
          aria-label="Toggle mobile menu"
        >
          {isMobileMenuOpen ? (
            <FaTimes
              style={{
                transition: 'transform 0.2s ease',
              }}
            />
          ) : (
            <FaBars
              style={{
                transition: 'transform 0.2s ease',
              }}
            />
          )}
        </button>

        {!isQuestionnaire && (
          <button
            type="button"
            className="login-button"
            style={{
              display: isMobile ? 'none' : token ? 'none' : 'flex',
            }}
            onClick={handleLoginClick}
          >
            <p>Login</p>
          </button>
        )}

        <Button onClick={handleClick} hidden={isMobile} variant="primary">
          {isQuestionnaire ? (
            <div
              style={{
                display: 'flex',
                flexDirection: 'row',
                gap: 10,
              }}
            >
              <FaSignOutAlt />
              <p>Logout</p>
            </div>
          ) : (
            <>{token ? 'Dashboard' : 'Register Now'}</>
          )}
        </Button>

        <div className="lang-dropdown-container">
          <div className="selected-language" onClick={toggleLanguageDropdown}>
            {language}
            <FaChevronDown
              style={{
                transform: languageDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                transition: 'transform 0.2s ease',
              }}
            />
          </div>
          {languageDropdown && (
            <div className="languages-dropdown">
              <span onClick={() => handleLanguageChange('en')}>English</span>
              <span onClick={() => handleLanguageChange('ar')}>Arabic</span>
              <span onClick={() => handleLanguageChange('fr')}>French</span>
            </div>
          )}
        </div>
      </div>

      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <a
            href={`${landingUrl}/pos`}
            target="_blank"
            rel="noopener noreferrer"
            style={externalLinkStyle()}
          >
            Menu
          </a>
          <a
            href={`${landingUrl}/pricing`}
            target="_blank"
            rel="noopener noreferrer"
            style={externalLinkStyle()}
          >
            Deals
          </a>
          <a
            href={`${landingUrl}/blog`}
            target="_blank"
            rel="noopener noreferrer"
            style={externalLinkStyle()}
          >
            Blog
          </a>
          <a
            href={`${landingUrl}/about`}
            target="_blank"
            rel="noopener noreferrer"
            style={externalLinkStyle()}
          >
            About
          </a>
          <a
            href={`${landingUrl}/contact`}
            target="_blank"
            rel="noopener noreferrer"
            style={externalLinkStyle()}
          >
            Contact Us
          </a>
          <Button
            styles={{
              display: token ? 'none' : 'flex',
              marginBottom: '0px',
            }}
            onClick={handleLoginClick}
          >
            <p>Login</p>
          </Button>

          <Button onClick={handleClick}>{token ? 'Dashboard' : 'Register Now'}</Button>
        </div>
      )}
    </header>
  );
};

export default Header;
