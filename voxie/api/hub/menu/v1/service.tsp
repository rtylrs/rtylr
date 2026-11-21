import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.menu.v1;

/** MenuService defines hub navigation menu endpoints. */
@route("/hub/menu/v1")
interface MenuService {
  /** Returns the navigation menu for the current user. */
  @get
  @route("/menu")
  listMenu(...ListMenuRequest): ListMenuResponse;

  /** Replaces the navigation menu items. */
  @put
  @route("/menu")
  updateMenu(...UpdateMenuRequest): UpdateMenuResponse;

  /** Returns menu banners for the current business. */
  @get
  @route("/menu/banners")
  listMenuBanners(...ListMenuBannersRequest): ListMenuBannersResponse;

  /** Creates a menu banner for the current business. */
  @post
  @route("/menu/banners")
  createMenuBanner(...CreateMenuBannerRequest): CreateMenuBannerResponse;

  /** Updates a menu banner for the current business. */
  @put
  @route("/menu/banners/{id}")
  updateMenuBanner(...UpdateMenuBannerRequest): UpdateMenuBannerResponse;

  /** Deletes a menu banner for the current business. */
  @delete
  @route("/menu/banners/{id}")
  deleteMenuBanner(...DeleteMenuBannerRequest): DeleteMenuBannerResponse;
}
