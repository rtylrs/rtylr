import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { cn } from '../../../../lib/cn';
import { HiChevronDown } from 'react-icons/hi';
import { AnimatePresence, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import Link from 'next/link';
import { useRouter } from 'next/router';

import { ASSETS, resolveAuthUrl } from '@voxire/config';
import { navigateExternal } from '../../../../lib/navigation';
import { localeFromPath, navPath, navPathWithHash } from '../../../../lib/locales';
import { FEATURE_LINKS, RESOURCE_LINKS, COMPANY_LINKS } from '../../../../lib/navLinks';
import { LanguageSwitcher } from '../../../../components/LanguageSwitcher/LanguageSwitcher';
import styles from './HeaderDesktop.module.scss';

interface LinkWithColorProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  color?: string;
  to: string;
}

// Sheet-style dropdown: scales out of the trigger with a spring,
// settles fast, exits quicker than it enters.
const DropdownPanel: React.FC<React.PropsWithChildren<{ open: boolean; className: string }>> = ({
  open,
  className,
  children,
}) => (
  <AnimatePresence>
    {open && (
      <motion.div
        className={className}
        initial={{ opacity: 0, scale: 0.96, y: -6 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.97, y: -4, transition: { duration: 0.15, ease: 'easeIn' } }}
        transition={{ type: 'spring', stiffness: 480, damping: 32, mass: 0.6 }}
        style={{ transformOrigin: 'top center' }}
      >
        {children}
      </motion.div>
    )}
  </AnimatePresence>
);

const HeaderDesktop: React.FC = () => {
  const { t } = useTranslation();
  const router = useRouter();
  const locale = localeFromPath(router.asPath).code;
  const [scrolled, setScrolled] = useState(false);
  const [featuresDropdown, setFeaturesDropdown] = useState(false);
  const [solutionsDropdown, setSolutionsDropdown] = useState(false);
  const [resourcesDropdown, setResourcesDropdown] = useState(false);

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
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setFeaturesDropdown(false);
        setSolutionsDropdown(false);
        setResourcesDropdown(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleEscape);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleEscape);
    };
  }, [featuresDropdown, solutionsDropdown, resourcesDropdown]);

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

  const handleSignupClick = useCallback(() => {
    navigateExternal(`${resolveAuthUrl()}/register`);
  }, []);

  const handleLoginClick = useCallback(() => {
    navigateExternal(resolveAuthUrl());
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
        <Link
          href={navPath('/', locale)}
          className={styles['logo-container']}
          aria-label={t('common.a11y.home')}
        >
          <img src={ASSETS.logos.light} alt={t('common.a11y.logoAlt')} />
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
              {t('common.nav.features')}
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: featuresDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                  transition: 'transform 0.3s cubic-bezier(0.32, 0.72, 0, 1)',
                }}
              />
            </button>
            <DropdownPanel open={featuresDropdown} className={styles['header-dropdown-wide']}>
              {FEATURE_LINKS.map((item) => (
                <Link
                  key={item.href}
                  href={navPathWithHash(item.href, locale)}
                  className={styles['dropdown-item']}
                  onClick={closeDropdowns}
                >
                  <div className={styles['dropdown-item-title']}>
                    {t(`common.nav.featureLinks.${item.key}.title`)}
                  </div>
                  <div className={styles['dropdown-item-desc']}>
                    {t(`common.nav.featureLinks.${item.key}.desc`)}
                  </div>
                </Link>
              ))}
            </DropdownPanel>
          </div>

          <LinkWithColor to={navPath('/pricing', locale)} className={styles['nav-link']}>
            {t('common.nav.pricing')}
          </LinkWithColor>

          <div className={cn(styles['nav-item'], 'resources-dropdown-container')}>
            <button
              type="button"
              className={styles['nav-link']}
              onClick={toggleResourcesDropdown}
              aria-expanded={resourcesDropdown}
              aria-haspopup="true"
            >
              {t('common.nav.resources')}
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: resourcesDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                  transition: 'transform 0.3s cubic-bezier(0.32, 0.72, 0, 1)',
                }}
              />
            </button>
            <DropdownPanel open={resourcesDropdown} className={styles['header-dropdown']}>
              {RESOURCE_LINKS.map((item) =>
                item.external ? (
                  <a
                    key={item.key}
                    href={item.href}
                    className={styles['dropdown-link']}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={closeDropdowns}
                  >
                    {t(`common.nav.resourceLinks.${item.key}`)}
                  </a>
                ) : (
                  <Link
                    key={item.key}
                    href={navPath(item.href, locale)}
                    className={styles['dropdown-link']}
                    onClick={closeDropdowns}
                  >
                    {t(`common.nav.resourceLinks.${item.key}`)}
                  </Link>
                )
              )}
            </DropdownPanel>
          </div>

          <div className={cn(styles['nav-item'], 'solutions-dropdown-container')}>
            <button
              type="button"
              className={styles['nav-link']}
              onClick={toggleSolutionsDropdown}
              aria-expanded={solutionsDropdown}
              aria-haspopup="true"
            >
              {t('common.nav.company')}
              <HiChevronDown
                className={styles['dropdown-icon']}
                style={{
                  transform: solutionsDropdown ? 'rotate(180deg)' : 'rotate(0deg)',
                  transition: 'transform 0.3s cubic-bezier(0.32, 0.72, 0, 1)',
                }}
              />
            </button>
            <DropdownPanel open={solutionsDropdown} className={styles['header-dropdown']}>
              {COMPANY_LINKS.map((item) =>
                item.external ? (
                  <a
                    key={item.key}
                    href={item.href}
                    className={styles['dropdown-link']}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={closeDropdowns}
                  >
                    {t(`common.nav.companyLinks.${item.key}`)}
                  </a>
                ) : (
                  <Link
                    key={item.key}
                    href={navPath(item.href, locale)}
                    className={styles['dropdown-link']}
                    onClick={closeDropdowns}
                  >
                    {t(`common.nav.companyLinks.${item.key}`)}
                  </Link>
                )
              )}
            </DropdownPanel>
          </div>
        </nav>
      </div>

      <div className={styles['header-right']}>
        <LanguageSwitcher />

        <div className={styles['header-buttons']}>
          {/* nav-link opts the plain sign-in out of the global button fill */}
          <button
            type="button"
            onClick={handleLoginClick}
            className={cn(styles['nav-link'], styles['header-signin'])}
          >
            {t('common.nav.signIn')}
          </button>
          <button type="button" onClick={handleSignupClick} className={styles['header-cta']}>
            {t('common.nav.signUp')}
          </button>
        </div>
      </div>
    </header>
  );
};

export default HeaderDesktop;
