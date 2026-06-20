import React from 'react';
import classNames from 'classnames';
import { cdnImageUrl } from '@voxire/config';
import { BusinessInfo, HeaderConfig, NavConfig, PageConfig } from '../../types';
import styles from './MenuHeader.module.scss';

// Re-export for backward compat
export type { BusinessInfo } from '../../types';

const PAGE_DEFS = [
  { id: 'menu', label: 'Menu' },
  { id: 'about', label: 'About' },
  { id: 'gallery', label: 'Gallery' },
  { id: 'reviews', label: 'Reviews' },
  { id: 'offers', label: 'Offers' },
] as const;

const COVER_HEIGHT: Record<string, string> = {
  none: '0px',
  sm: '110px',
  md: '160px',
  lg: '220px',
};

interface MenuHeaderProps {
  business: BusinessInfo;
  headerConfig: HeaderConfig;
  navConfig: NavConfig;
  pages: PageConfig;
  activePage: string;
  onPageChange: (page: string) => void;
  categories?: Array<{ id: string; name: string }>;
  activeCategoryId?: string;
  onCategorySelect?: (id: string) => void;
  className?: string;
}

const MenuHeader: React.FC<MenuHeaderProps> = ({
  business,
  headerConfig,
  navConfig,
  pages,
  activePage,
  onPageChange,
  categories,
  activeCategoryId,
  onCategorySelect,
  className,
}) => {
  const enabledPages = PAGE_DEFS.filter(
    (p) =>
      pages[p.id as keyof PageConfig] !== false &&
      (p.id === 'menu' || pages[p.id as keyof PageConfig] === true)
  );
  const showPageNav = enabledPages.length > 1;
  const showCategoryNav = activePage === 'menu' && categories && categories.length > 0;
  const coverH = COVER_HEIGHT[headerConfig.coverHeight] ?? '160px';
  const coverImageUrl = cdnImageUrl(business.coverImage);
  const logoUrl = cdnImageUrl(business.logo);

  const navItemClass = (active: boolean) =>
    classNames(
      styles.header__navItem,
      styles[`header__navItem--${navConfig.style}`],
      styles[`header__navItem--${navConfig.size}`],
      { [styles['header__navItem--active']]: active }
    );

  return (
    <header
      className={classNames(
        styles.header,
        { [styles['header--sticky']]: headerConfig.sticky },
        className
      )}
    >
      {/* Cover image */}
      {headerConfig.coverHeight !== 'none' && coverImageUrl && (
        <div className={styles.header__cover} style={{ height: coverH }}>
          <img src={coverImageUrl} alt="" />
          <div className={styles.header__coverOverlay} />
        </div>
      )}

      {/* Info row */}
      <div className={styles.header__info}>
        {headerConfig.showLogo && logoUrl && (
          <div className={styles.header__logo}>
            <img src={logoUrl} alt={`${business.name} logo`} />
          </div>
        )}
        <div className={styles.header__text}>
          <h1 className={styles.header__name}>{business.name}</h1>
          {headerConfig.showDescription && business.description && (
            <p className={styles.header__description}>{business.description}</p>
          )}
          {headerConfig.showMeta && (
            <div className={styles.header__meta}>
              {business.address && (
                <span className={styles.header__metaItem}>{business.address}</span>
              )}
              {business.phone && (
                <span className={styles.header__metaItem}>
                  <a href={`tel:${business.phone}`}>{business.phone}</a>
                </span>
              )}
              {business.hours && <span className={styles.header__metaItem}>{business.hours}</span>}
            </div>
          )}
        </div>
      </div>

      {/* Page nav */}
      {showPageNav && (
        <nav className={classNames(styles.header__nav, styles['header__nav--pages'])}>
          {enabledPages.map((p) => (
            <button
              key={p.id}
              className={navItemClass(activePage === p.id)}
              onClick={() => onPageChange(p.id)}
            >
              {p.label}
            </button>
          ))}
        </nav>
      )}

      {/* Category nav (menu page only) */}
      {showCategoryNav && (
        <nav
          className={classNames(styles.header__nav, styles['header__nav--categories'], {
            [styles['header__nav--sub']]: showPageNav,
          })}
        >
          {categories!.map((cat) => (
            <button
              key={cat.id}
              className={classNames(
                styles.header__navItem,
                styles['header__navItem--underline'],
                styles[`header__navItem--${navConfig.size}`],
                styles['header__navItem--category'],
                {
                  [styles['header__navItem--active']]: activeCategoryId === cat.id,
                }
              )}
              onClick={() => onCategorySelect?.(cat.id)}
            >
              {cat.name}
            </button>
          ))}
        </nav>
      )}
    </header>
  );
};

export default MenuHeader;
