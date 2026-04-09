import React from 'react';
import classNames from 'classnames';
import { motion } from 'framer-motion';
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
  layout?: 'grid' | 'list' | 'minimal';
  onClick?: (item: MenuItemData) => void;
  onAddToCart?: (item: MenuItemData) => void;
  className?: string;
}

const MenuItem: React.FC<MenuItemProps> = ({
  item,
  layout = 'grid',
  onClick,
  onAddToCart,
  className,
}) => {
  const formatPrice = (price: number, currency = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
    }).format(price);
  };

  return (
    <motion.div
      className={classNames(styles.item, styles[`item--${layout}`], className, {
        [styles['item--unavailable']]: item.available === false,
      })}
      onClick={() => onClick?.(item)}
      whileHover={{ scale: layout !== 'minimal' ? 1.01 : 1 }}
      whileTap={{ scale: 0.98 }}
      layout
    >
      {item.image && layout !== 'minimal' && (
        <div className={styles.item__image}>
          <img src={item.image} alt={item.name} loading="lazy" />
          {item.available === false && (
            <div className={styles.item__unavailableBadge}>Unavailable</div>
          )}
        </div>
      )}

      <div className={styles.item__content}>
        <div className={styles.item__header}>
          <h3 className={styles.item__name}>{item.name}</h3>
          {item.badges && item.badges.length > 0 && (
            <div className={styles.item__badges}>
              {item.badges.map((badge) => (
                <span key={badge} className={styles.item__badge}>
                  {badge}
                </span>
              ))}
            </div>
          )}
        </div>

        {item.description && (
          <p className={styles.item__description}>{item.description}</p>
        )}

        <div className={styles.item__footer}>
          <span className={styles.item__price}>
            {formatPrice(item.price, item.currency)}
          </span>
          {onAddToCart && item.available !== false && (
            <button
              className={styles.item__addBtn}
              onClick={(e) => {
                e.stopPropagation();
                onAddToCart(item);
              }}
              aria-label={`Add ${item.name} to cart`}
            >
              +
            </button>
          )}
        </div>
      </div>
    </motion.div>
  );
};

export default MenuItem;
