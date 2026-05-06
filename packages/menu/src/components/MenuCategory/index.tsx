import React from 'react';
import classNames from 'classnames';
import MenuItem, { MenuItemData } from '../MenuItem';
import { ItemDisplayConfig } from '../../types';
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
  displayConfig: ItemDisplayConfig;
  onItemClick?: (item: MenuItemData) => void;
  onAddToCart?: (item: MenuItemData) => void;
  className?: string;
}

const MenuCategory: React.FC<MenuCategoryProps> = ({
  category,
  displayConfig,
  onItemClick,
  onAddToCart,
  className,
}) => {
  if (!category.items || category.items.length === 0) return null;

  const { layout, columns } = displayConfig;

  const gridStyle: React.CSSProperties =
    layout === 'grid' ? { gridTemplateColumns: `repeat(${columns}, 1fr)` } : {};

  return (
    <section className={classNames(styles.category, className)} id={`category-${category.id}`}>
      <div className={styles.category__header}>
        <h2 className={styles.category__name}>{category.name}</h2>
        {category.description && (
          <p className={styles.category__description}>{category.description}</p>
        )}
      </div>

      <div
        className={classNames(styles.category__items, styles[`category__items--${layout}`])}
        style={gridStyle}
      >
        {category.items.map((item) => (
          <MenuItem
            key={item.id}
            item={item}
            displayConfig={displayConfig}
            onClick={onItemClick}
            onAddToCart={onAddToCart}
          />
        ))}
      </div>
    </section>
  );
};

export default MenuCategory;
