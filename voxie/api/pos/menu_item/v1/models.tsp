import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.menu_item.v1;

/** MenuItem is a sellable item on the POS menu. */
model MenuItem {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  @encode(string)
  category_id: int64;

  name: string;
  sku: string;
  sale_price: float64;
  is_active: boolean;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  /** Public storefront image for the menu item. */
  image_url: string;

  /** Display category name for the menu item. */
  category: string;

  /** Scanned barcode value for the menu item. */
  barcode: string;

  /** Customer-facing menu item description. */
  description: string;

  /** Sale tax rate for the menu item. */
  tax_rate: float64;

  /** Current tracked stock quantity for the menu item. */
  stock: float64;

  /** Whether inventory stock tracking is enabled on the linked ERP product. */
  track_inventory?: boolean;

  /** Whether backorders are permitted when tracked stock reaches zero. */
  allow_backorder?: boolean;

  /**
   * product_id is the ERP product this item is sold from; absent for items
   * that are not stocked as a single product. When it is absent, cost and
   * ERP-side stock for this item are unknown rather than zero. The server
   * must verify the referenced product belongs to the caller's business.
   */
  @encode(string)
  product_id?: int64;

  /**
   * unit_cost is the cost of one unit at the time of the read, copied from
   * the linked ERP product. Absent when product_id is absent. Reported for
   * display and margin only: the authoritative value is the ERP product's.
   */
  unit_cost?: float64;
}

model CreateMenuItemRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  sku: string;

  @minValue(0)
  sale_price: float64;

  /** Optional customer-facing description for the menu item. */
  description?: string;

  /** Optional public storefront image URL for the menu item. */
  image_url?: string;
}

model CreateMenuItemResponse {
  @encode(string)
  item_id: int64;

  message: string;
}

model ListMenuItemsRequest {}

model ListMenuItemsResponse {
  items: MenuItem[];
}

model ListPublicMenuItemsRequest {
  @minValueExclusive(0)
  @query("business_id")
  @encode(string)
  business_id: int64;

  /** Optional. Accepted for multi-location menu URLs; ignored until per-location stock scoping ships. */
  @query("location_id")
  @encode(string)
  location_id: int64;
}

model ListPublicMenuItemsResponse {
  items: MenuItem[];
}

model UpdateMenuItemBody {
  /** Display name for the menu item. */
  name: string;

  /** Sell price for the menu item. */
  @minValue(0)
  sale_price: float64;

  /** Controls whether the item can be sold. */
  is_active: boolean;

  /** Current tracked stock quantity for the menu item. */
  @minValue(0)
  stock: float64;

  /** Optional customer-facing description; omitted preserves existing description. */
  description?: string;

  /** Optional primary image URL; empty string clears the primary image, omitted preserves existing image. */
  image_url?: string;
}

model UpdateMenuItemRequest {
  /** id is the unique identifier for the menu item. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable menu item fields bound from the HTTP payload. */
  @body
  body: UpdateMenuItemBody;
}

model UpdateMenuItemResponse {
  message: string;
}
