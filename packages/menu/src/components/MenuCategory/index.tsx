import React from 'react';
import classNames from 'classnames';
import MenuItem, { MenuItemData } from '../MenuItem';
import styles from './MenuCategory.module.scss';

export interface MenuCategoryData {
  id: string;
  name: string;
  description?: string;
  image?: string;
  items: MenuItemData[];
}

interface MenuCategoryProps {
  category: MenuCategoryData;
  layout?: 'grid' | 'list' | 'minimal';
  onItemClick?: (item: MenuItemData) => void;
  onAddToCart?: (item: MenuItemData) => void;
  className?: string;
}

const MenuCategory: React.FC<MenuCategoryProps> = ({
  category,
  layout = 'grid',
  onItemClick,
  onAddToCart,
  className,
}) => {
  if (!category.items || category.items.length === 0) return null;

  return (
    <section className={classNames(styles.category, className)} id={`category-${category.id}`}>
      <div className={styles.category__header}>
        <h2 className={styles.category__name}>{category.name}</h2>
        {category.description && (
          <p className={styles.category__description}>{category.description}</p>
        )}
      </div>

      <div className={classNames(styles.category__items, styles[`category__items--${layout}`])}>
        {category.items.map((item) => (
          <MenuItem
            key={item.id}
            item={item}
            layout={layout}
            onClick={onItemClick}
            onAddToCart={onAddToCart}
          />
        ))}
      </div>
    </section>
  );
};

export default MenuCategory;
