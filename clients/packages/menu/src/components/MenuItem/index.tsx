import React from "react";
import classNames from "classnames";
import { cdnImageUrl } from "@voxire/config";
import { ItemDisplayConfig } from "../../types";
import styles from "./MenuItem.module.scss";

export interface MenuItemData {
  id: string;
  name: string;
  description?: string;
  price: number;
  formattedPrice?: string;
  image?: string;
  currency?: string;
  badges?: string[];
  dietaryTags?: string[];
  allergens?: string[];
  calories?: number;
  available?: boolean;
}

interface MenuItemProps {
  item: MenuItemData;
  displayConfig: ItemDisplayConfig;
  onClick?: (item: MenuItemData) => void;
  onAddToCart?: (item: MenuItemData) => void;
}

const RADIUS_MAP: Record<string, string> = {
  none: "0px",
  sm: "8px",
  md: "14px",
  lg: "20px",
};

const IMG_WIDTH_MAP: Record<string, string> = {
  sm: "80px",
  md: "102px",
  lg: "126px",
};

const IMG_ASPECT_MAP: Record<string, string> = {
  sm: "16/10",
  md: "4/3",
  lg: "1/1",
};

export function formatMenuPrice(price: number, currency = "USD"): string {
  const code =
    typeof currency === "string" && currency.trim() ? currency.trim() : "USD";
  try {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: code,
    }).format(price);
  } catch {
    return `${code} ${price.toFixed(2)}`;
  }
}

const MenuItem: React.FC<MenuItemProps> = ({
  item,
  displayConfig,
  onClick,
  onAddToCart,
}) => {
  const {
    imagePosition,
    imageSize,
    showDescription,
    showBadges,
    cardRadius,
    layout,
  } = displayConfig;

  const isUnavailable = item.available === false;
  const showImage = imagePosition !== "none" && layout !== "minimal";
  const isHorizontal = imagePosition === "right" || imagePosition === "left";

  const cardStyle: React.CSSProperties = {
    borderRadius: RADIUS_MAP[cardRadius] ?? "14px",
  };

  const imgStyle: React.CSSProperties = isHorizontal
    ? {
        width: IMG_WIDTH_MAP[imageSize] ?? "102px",
        flexShrink: 0,
        aspectRatio: "1",
      }
    : { aspectRatio: IMG_ASPECT_MAP[imageSize] ?? "4/3" };

  const formattedPrice =
    item.formattedPrice ?? formatMenuPrice(item.price, item.currency);
  const imageUrl = cdnImageUrl(item.image);
  const primaryBadge = item.badges?.[0];
  const metaNotes = [
    ...(item.dietaryTags ?? []).slice(0, 2),
    ...(item.calories && item.calories > 0 ? [`${item.calories} kcal`] : []),
  ];

  return (
    <div
      className={classNames(styles.item, styles[`item--${layout}`], {
        [styles["item--unavailable"]]: isUnavailable,
        [styles["item--horizontal"]]: isHorizontal,
        [styles["item--imgLeft"]]: imagePosition === "left",
      })}
      style={cardStyle}
      onClick={() => !isUnavailable && onClick?.(item)}
    >
      {showImage && (
        <div className={styles.item__image} style={imgStyle}>
          {imageUrl ? (
            <img src={imageUrl} alt={item.name} loading="lazy" />
          ) : (
            <div className={styles.item__imagePlaceholder} />
          )}
          {showBadges && primaryBadge && !isUnavailable && (
            <span className={styles.item__photoBadge}>{primaryBadge}</span>
          )}
          {isUnavailable && (
            <span className={styles.item__unavailableBadge}>Unavailable</span>
          )}
          {onAddToCart && !isUnavailable && !isHorizontal && (
            <button
              type="button"
              className={styles.item__photoAddBtn}
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
      )}

      <div className={styles.item__content}>
        <div className={styles.item__header}>
          <p className={styles.item__name}>{item.name}</p>
          {!showImage && showBadges && primaryBadge && (
            <span className={styles.item__inlineBadge}>{primaryBadge}</span>
          )}
          {!showImage && isUnavailable && (
            <span className={styles.item__inlineBadge}>Unavailable</span>
          )}
        </div>

        {showDescription && item.description && (
          <p className={styles.item__description}>{item.description}</p>
        )}

        {showBadges && metaNotes.length > 0 && (
          <p className={styles.item__metaNotes}>{metaNotes.join(" · ")}</p>
        )}

        <div className={styles.item__footer}>
          <span className={styles.item__price}>{formattedPrice}</span>
          {onAddToCart && !isUnavailable && (isHorizontal || !showImage) && (
            <button
              type="button"
              className={
                layout === "minimal"
                  ? styles.item__addBtnMinimal
                  : styles.item__addBtn
              }
              onClick={(e) => {
                e.stopPropagation();
                onAddToCart(item);
              }}
              aria-label={`Add ${item.name} to cart`}
            >
              {layout === "minimal" ? "Add" : "+"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default MenuItem;
