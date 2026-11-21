import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.menu_item.v1;

/** MenuItemService defines POS menu item endpoints. */
@route("/pos/menu-item/v1")
interface MenuItemService {
  /** Returns the shared product catalog for POS operators who can create orders. */
  @get
  @route("/catalog")
  listPOSCatalog(...ListMenuItemsRequest): ListMenuItemsResponse;

  /** Returns menu items for the active business. */
  @get
  @route("/menu-items")
  listMenuItems(...ListMenuItemsRequest): ListMenuItemsResponse;

  /** Returns menu items for anonymous public menu surfaces. */
  @get
  @route("/menu-items/public")
  listPublicMenuItems(
    ...ListPublicMenuItemsRequest,
  ): ListPublicMenuItemsResponse;

  /** Creates a new menu item. */
  @post
  @route("/menu-items")
  createMenuItem(...CreateMenuItemRequest): CreateMenuItemResponse;

  /** Updates an existing menu item. */
  @put
  @route("/menu-items/{id}")
  updateMenuItem(...UpdateMenuItemRequest): UpdateMenuItemResponse;
}
