import React from 'react';
import classNames from 'classnames';
import { cdnImageUrl } from '@voxire/config';
import { BusinessInfo, HeaderConfig, NavConfig, PageConfig } from '../../types';
import styles from './MenuHeader.module.scss';

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
  sm: '96px',
  md: '132px',
  lg: '176px',
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
  topStartSlot?: React.ReactNode;
  topEndSlot?: React.ReactNode;
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
  topStartSlot,
  topEndSlot,
  className,
}) => {
  const enabledPages = PAGE_DEFS.filter(
    (p) =>
      pages[p.id as keyof PageConfig] !== false &&
      (p.id === 'menu' || pages[p.id as keyof PageConfig] === true)
  );
  const showPageNav = enabledPages.length > 1;
  const showCategoryNav = activePage === 'menu' && categories && categories.length > 0;
  const coverH = COVER_HEIGHT[headerConfig.coverHeight] ?? '132px';
  const coverImageUrl = cdnImageUrl(headerConfig.coverImageUrl || business.coverImage);
  const logoUrl = cdnImageUrl(headerConfig.logoUrl || business.logo);
  const hasCover = headerConfig.coverHeight !== 'none' && Boolean(coverImageUrl);
  const hasTopBar = Boolean(topStartSlot) || Boolean(topEndSlot);

  const metaParts = [business.address, business.hours, business.phone].filter(Boolean);

  return (
    <header
      className={classNames(
        styles.header,
        { [styles['header--sticky']]: headerConfig.sticky },
        className
      )}
    >
      {/* Hero cover with floating glass top controls */}
      {hasCover ? (
        <div className={styles.header__cover} style={{ height: coverH }}>
          <img src={coverImageUrl} alt="" />
          <div className={styles.header__coverOverlay} />
          {hasTopBar && (
            <div className={styles.header__topBarOverlay}>
              <div className={styles.header__topSlot}>{topStartSlot}</div>
              <div className={styles.header__topSlot}>{topEndSlot}</div>
            </div>
          )}
        </div>
      ) : (
        hasTopBar && (
          <div className={styles.header__topBarInline}>
            <div className={styles.header__topSlot}>{topStartSlot}</div>
            <div className={styles.header__topSlot}>{topEndSlot}</div>
          </div>
        )
      )}

      {/* Compact brand identity row */}
      <div
        className={classNames(styles.header__info, {
          [styles['header__info--overlapped']]: hasCover && headerConfig.showLogo && Boolean(logoUrl),
        })}
      >
        {headerConfig.showLogo && logoUrl && (
          <div className={styles.header__logo}>
            <img src={logoUrl} alt={`${business.name} logo`} />
          </div>
        )}
        <div className={styles.header__text}>
          <div className={styles.header__titleRow}>
            <h1 className={styles.header__name}>{business.name}</h1>
            {showPageNav && (
              <nav className={styles.header__pagePills} aria-label="Menu pages">
                {enabledPages.map((p) => (
                  <button
                    key={p.id}
                    type="button"
                    className={classNames(styles.header__pagePill, {
                      [styles['header__pagePill--active']]: activePage === p.id,
                    })}
                    onClick={() => onPageChange(p.id)}
                  >
                    {p.label}
                  </button>
                ))}
              </nav>
            )}
          </div>
          {headerConfig.showDescription && business.description && (
            <p className={styles.header__description}>{business.description}</p>
          )}
          {headerConfig.showMeta && metaParts.length > 0 && (
            <p className={styles.header__metaLine}>{metaParts.join(' · ')}</p>
          )}
        </div>
      </div>

      {/* Single sleek horizontal category rail */}
      {showCategoryNav && (
        <nav
          className={classNames(styles.header__nav, styles['header__nav--categories'])}
          aria-label="Menu categories"
        >
          {categories!.map((cat) => (
            <button
              key={cat.id}
              type="button"
              className={classNames(
                styles.header__navItem,
                styles[`header__navItem--${navConfig.style}`],
                styles[`header__navItem--${navConfig.size}`],
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
