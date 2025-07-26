import React, { useState, useEffect, useCallback } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faChevronDown } from '@fortawesome/free-solid-svg-icons';

import { selectLanguage } from '../../services/i18next';

const Header: React.FC = () => {
  const [languageDropdown, setLanguageDropdown] = useState(false);
  const [language, setLanguage] = useState<string>('en');

  useEffect(() => {
    const storedLanguage = localStorage.getItem('language');
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

  return (
    <header className="header">
      <div className="header-logo" style={{ cursor: 'pointer' }}>
        <img src={'/logo.png'} alt="Logo" />
      </div>

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
    </header>
  );
};

export default Header;
