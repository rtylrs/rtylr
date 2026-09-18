import React from 'react';
import classNames from 'classnames';
import { cdnImageUrl } from '@voxire/config';
import { ItemDisplayConfig } from '../../types';
import styles from './MenuItem.module.scss';

export interface MenuItemData {
  id: string;
  name: string;
  description?: string;
  price: number;
  image?: string;
  currency?: string;
  badges?: string[];
  available?: boolean;
}

interface MenuItemProps {
  item: MenuItemData;
  displayConfig: ItemDisplayConfig;
  onClick?: (item: MenuItemData) => void;
  onAddToCart?: (item: MenuItemData) => void;
}

const RADIUS_MAP: Record<string, string> = {
  none: '0px',
  sm: '4px',
  md: '8px',
  lg: '16px',
};

const IMG_WIDTH_MAP: Record<string, string> = {
  sm: '72px',
  md: '104px',
  lg: '140px',
};

const IMG_ASPECT_MAP: Record<string, string> = {
  sm: '3/2',
  md: '4/3',
  lg: '1/1',
};

const MenuItem: React.FC<MenuItemProps> = ({ item, displayConfig, onClick, onAddToCart }) => {
  const { imagePosition, imageSize, showDescription, showBadges, cardRadius, layout } =
    displayConfig;

  const isUnavailable = item.available === false;
  const showImage = imagePosition !== 'none' && layout !== 'minimal';
  const isHorizontal = imagePosition === 'right' || imagePosition === 'left';

  const cardStyle: React.CSSProperties = {
    borderRadius: RADIUS_MAP[cardRadius] ?? '8px',
  };

  const imgStyle: React.CSSProperties = isHorizontal
    ? {
        width: IMG_WIDTH_MAP[imageSize] ?? '104px',
        flexShrink: 0,
        aspectRatio: '1',
      }
    : { aspectRatio: IMG_ASPECT_MAP[imageSize] ?? '4/3' };

  const formattedPrice = new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: item.currency ?? 'USD',
  }).format(item.price);
  const imageUrl = cdnImageUrl(item.image);

  return (
    <div
      className={classNames(styles.item, styles[`item--${layout}`], {
        [styles['item--unavailable']]: isUnavailable,
        [styles['item--horizontal']]: isHorizontal,
        [styles['item--imgLeft']]: imagePosition === 'left',
      })}
      style={cardStyle}
      onClick={() => !isUnavailable && onClick?.(item)}
    >
      {/* Image */}
      {showImage && (
        <div className={styles.item__image} style={imgStyle}>
          {imageUrl ? (
            <img src={imageUrl} alt={item.name} loading="lazy" />
          ) : (
            <div className={styles.item__imagePlaceholder} />
          )}
          {isUnavailable && <span className={styles.item__unavailableBadge}>Unavailable</span>}
        </div>
      )}

      {/* Content */}
      <div className={styles.item__content}>
        <div className={styles.item__header}>
          <p className={styles.item__name}>{item.name}</p>
          {showBadges && item.badges && item.badges.length > 0 && (
            <div className={styles.item__badges}>
              {item.badges.map((b) => (
                <span key={b} className={styles.item__badge}>
                  {b}
                </span>
              ))}
            </div>
          )}
        </div>

        {showDescription && item.description && (
          <p className={styles.item__description}>{item.description}</p>
        )}

        <div className={styles.item__footer}>
          <span className={styles.item__price}>{formattedPrice}</span>
          {onAddToCart && !isUnavailable && (
            <button
              className={layout === 'minimal' ? styles.item__addBtnMinimal : styles.item__addBtn}
              onClick={(e) => {
                e.stopPropagation();
                onAddToCart(item);
              }}
              aria-label={`Add ${item.name} to cart`}
            >
              {layout === 'minimal' ? 'Add' : '+'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default MenuItem;
