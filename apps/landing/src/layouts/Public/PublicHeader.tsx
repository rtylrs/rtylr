import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
  faBars,
  faChevronDown,
  faRightFromBracket,
  faTimes,
} from '@fortawesome/free-solid-svg-icons';
import { Link, useLocation, useNavigate } from 'react-router-dom';

import VbitezButton from '../../components/VbitezButton/VbitezButton';
import useColors from '../../hooks/useColors';
import { useWindowDimensions } from '../../hooks/useWindowDimensions';
import { useSelector } from 'react-redux';
import { selectLanguage } from '../../services/i18next';
import { RootState } from '../../redux/store';

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

  const location = useLocation();
  const navigate = useNavigate();
  const { primary } = useColors();
  const { width } = useWindowDimensions();
  const user = useSelector((state: RootState) => state.user);

  const isMobile = useMemo(() => width < 968, [width]);
  const isQuestionnaire = useMemo(
    () => location.pathname.includes('/auth/questionnaire'),
    [location.pathname]
  );
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
    const storedLanguage = localStorage.getItem('language');
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
        navigate('/app/business/');
      } else {
        navigate('/app/user/');
      }
    } else {
      navigate('/auth/register');
    }
  }, [user.role, navigate]);

  const handleLogoClick = useCallback(() => {
    navigate('/');
  }, [navigate]);

  const handleLoginClick = useCallback(() => {
    navigate('/auth/login');
  }, [navigate]);

  const handleNavigation = useCallback(
    (path: string) => {
      setDropdown(false);
      navigate(path);
    },
    [navigate]
  );

  const handleExternalLink = useCallback(() => {
    setDropdown(false);
    window.open('https://voxire.com/careers', '_blank');
  }, []);

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
        <img src={'/logo.png'} alt="Logo" />
      </div>

      <div
        className="header-left"
        style={{
          transform: 'translateX(5%)',
        }}
      >
        <LinkWithColor to="/restaurant/69">Demo</LinkWithColor>
        <LinkWithColor to="/pricing">Pricing</LinkWithColor>
        <LinkWithColor to="/restaurants">Restaurants</LinkWithColor>
        <LinkWithColor to="/uses">Use Cases</LinkWithColor>
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
          <div className="header-dropdown">
            <h6 onClick={() => handleNavigation('/about')}>About Us</h6>
            <h6 onClick={() => handleNavigation('/contact')}>Contact Us</h6>
            <h6 onClick={() => handleNavigation('/faqs')}>FAQs</h6>
            <h6 onClick={handleExternalLink}>Careers (Jobs)</h6>
          </div>
        )}
      </div>

      <div className="header-right">
        <button
          className="mobile-menu-toggle"
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

        {!isQuestionnaire && (
          <button
            className="login-button"
            style={{
              display: isMobile ? 'none' : user.token ? 'none' : 'flex',
            }}
            onClick={handleLoginClick}
          >
            <p>Login</p>
          </button>
        )}

        <VbitezButton onClick={handleClick} hidden={isMobile} variant="primary">
          {isQuestionnaire ? (
            <div
              style={{
                display: 'flex',
                flexDirection: 'row',
                gap: 10,
              }}
            >
              <FontAwesomeIcon icon={faRightFromBracket} />
              <p>Logout</p>
            </div>
          ) : (
            <>{user.token ? 'Dashboard' : 'Register Now'}</>
          )}
        </VbitezButton>

        <div className="lang-dropdown-container">
          <div className="selected-language" onClick={toggleLanguageDropdown}>
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
          <LinkWithColor to="/menu">Menu</LinkWithColor>
          <LinkWithColor to="/deals">Deals</LinkWithColor>
          <LinkWithColor to="/blog">Blog</LinkWithColor>
          <LinkWithColor to="/about">About</LinkWithColor>
          <LinkWithColor to="/contact">Contact Us</LinkWithColor>
          <VbitezButton
            styles={{
              display: user.token ? 'none' : 'flex',
              marginBottom: '0px',
            }}
            onClick={handleLoginClick}
          >
            <p>Login</p>
          </VbitezButton>

          <VbitezButton onClick={handleClick}>
            {user.token ? 'Dashboard' : 'Register Now'}
          </VbitezButton>
        </div>
      )}
    </header>
  );
};

export default Header;
