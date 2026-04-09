import React from 'react';
import classNames from 'classnames';
import styles from './MenuHeader.module.scss';

export interface BusinessInfo {
  name: string;
  logo?: string;
  coverImage?: string;
  description?: string;
  address?: string;
  phone?: string;
  hours?: string;
}

interface MenuHeaderProps {
  business: BusinessInfo;
  categories?: Array<{ id: string; name: string }>;
  activeCategoryId?: string;
  onCategorySelect?: (id: string) => void;
  className?: string;
}

const MenuHeader: React.FC<MenuHeaderProps> = ({
  business,
  categories,
  activeCategoryId,
  onCategorySelect,
  className,
}) => {
  return (
    <header className={classNames(styles.header, className)}>
      {business.coverImage && (
        <div className={styles.header__cover}>
          <img src={business.coverImage} alt={`${business.name} cover`} />
          <div className={styles.header__coverOverlay} />
        </div>
      )}

      <div className={styles.header__info}>
        {business.logo && (
          <div className={styles.header__logo}>
            <img src={business.logo} alt={`${business.name} logo`} />
          </div>
        )}
        <div className={styles.header__text}>
          <h1 className={styles.header__name}>{business.name}</h1>
          {business.description && (
            <p className={styles.header__description}>{business.description}</p>
          )}
          <div className={styles.header__meta}>
            {business.address && (
              <span className={styles.header__metaItem}>{business.address}</span>
            )}
            {business.phone && (
              <span className={styles.header__metaItem}>
                <a href={`tel:${business.phone}`}>{business.phone}</a>
              </span>
            )}
            {business.hours && (
              <span className={styles.header__metaItem}>{business.hours}</span>
            )}
          </div>
        </div>
      </div>

      {categories && categories.length > 0 && (
        <nav className={styles.header__nav}>
          {categories.map((cat) => (
            <button
              key={cat.id}
              className={classNames(styles.header__navItem, {
                [styles['header__navItem--active']]: activeCategoryId === cat.id,
              })}
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
