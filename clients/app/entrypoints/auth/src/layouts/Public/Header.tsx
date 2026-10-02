import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { FaChevronDown } from 'react-icons/fa';
import { useNavigate } from '@tanstack/react-router';

import Logo from '../../shared/components/Logo';
import { selectLanguage } from '../../shared/services/i18next';
import { getCookie } from '@voxire/auth';

const Header: React.FC = () => {
  const { t } = useTranslation();
  const [languageDropdown, setLanguageDropdown] = useState(false);
  const [language, setLanguage] = useState<string>('en');
  const navigate = useNavigate();

  useEffect(() => {
    const storedLanguage = getCookie('language');
    if (storedLanguage) {
      setLanguage(storedLanguage);
    }
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;

      if (languageDropdown && !target.closest('.lang-dropdown-container')) {
        setLanguageDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [languageDropdown]);

  const toggleLanguageDropdown = useCallback(() => {
    setLanguageDropdown((prev) => !prev);
  }, []);

  const handleLanguageChange = useCallback((lang: string) => {
    selectLanguage(lang);
    setLanguage(lang);
    setLanguageDropdown(false);
  }, []);

  const getLanguageDisplay = (lang: string) => {
    switch (lang) {
      case 'en':
        return 'english';
      case 'ar':
        return 'arabic';
      default:
        return 'english';
    }
  };

  return (
    <header className="header">
      <div
        className="header-logo"
        style={{ cursor: 'pointer' }}
        onClick={() => navigate({ to: '/' })}
        role="button"
        aria-label={t('auth.common.goToHome')}
        tabIndex={0}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') navigate({ to: '/' });
        }}
      >
        <Logo variant="dark" />
      </div>

      <div className="lang-dropdown-container">
        <div
          className="selected-language"
          onClick={toggleLanguageDropdown}
          role="button"
          aria-haspopup="listbox"
          aria-expanded={languageDropdown}
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') toggleLanguageDropdown();
          }}
        >
          {getLanguageDisplay(language)}
          <FaChevronDown
            style={{
              transform: languageDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
              transition: 'transform 0.2s ease',
            }}
          />
        </div>
        {languageDropdown && (
          <div
            className="languages-dropdown"
            role="listbox"
            aria-label={t('auth.common.selectLanguage')}
          >
            <span
              role="option"
              aria-selected={language === 'en'}
              onClick={() => handleLanguageChange('en')}
            >
              english
            </span>
            <span
              role="option"
              aria-selected={language === 'ar'}
              onClick={() => handleLanguageChange('ar')}
            >
              arabic
            </span>
          </div>
        )}
      </div>
    </header>
  );
};

export default Header;
