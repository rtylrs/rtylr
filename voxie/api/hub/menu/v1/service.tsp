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

  // ── Branded Menu Profiles & Presets ─────────────────────────────────────────

  /** Returns the branch-scoped menu design profile (draft & published). */
  @get
  @route("/branding/profile")
  getMenuProfile(...GetMenuProfileRequest): GetMenuProfileResponse;

  /** Saves the working draft for a branch's branded menu profile. */
  @put
  @route("/branding/profile")
  saveMenuProfileDraft(...SaveMenuProfileDraftRequest): SaveMenuProfileDraftResponse;

  /** Publishes the current draft config & custom CSS to guests. */
  @post
  @route("/branding/profile/publish")
  publishMenuProfile(...PublishMenuProfileRequest): PublishMenuProfileResponse;

  /** Copies another branch's menu profile, overrides, currencies, and translations. */
  @post
  @route("/branding/profile/clone")
  cloneMenuProfile(...CloneMenuProfileRequest): CloneMenuProfileResponse;

  /** Lists saved theme/layout presets for the current business. */
  @get
  @route("/branding/presets")
  listMenuPresets(...ListMenuPresetsRequest): ListMenuPresetsResponse;

  /** Creates a reusable theme/layout preset. */
  @post
  @route("/branding/presets")
  createMenuPreset(...CreateMenuPresetRequest): CreateMenuPresetResponse;

  /** Updates an existing theme/layout preset. */
  @put
  @route("/branding/presets/{id}")
  updateMenuPreset(...UpdateMenuPresetRequest): UpdateMenuPresetResponse;

  /** Deletes a theme/layout preset. */
  @delete
  @route("/branding/presets/{id}")
  deleteMenuPreset(...DeleteMenuPresetRequest): DeleteMenuPresetResponse;

  // ── Per-Branch POS Catalog Overrides ────────────────────────────────────────

  /** Lists branch-specific visibility, pricing, dietary, and upsell overrides. */
  @get
  @route("/menu/branch-items")
  listBranchMenuItems(...ListBranchMenuItemsRequest): ListBranchMenuItemsResponse;

  /** Creates or updates a branch-specific item override. */
  @put
  @route("/menu/branch-items")
  upsertBranchMenuItem(...UpsertBranchMenuItemRequest): UpsertBranchMenuItemResponse;

  // ── Translations (AI + Manual) ──────────────────────────────────────────────

  /** Lists menu translations for a branch and optional locale. */
  @get
  @route("/menu/translations")
  listMenuTranslations(...ListMenuTranslationsRequest): ListMenuTranslationsResponse;

  /** Saves a manual or AI translation entry. */
  @put
  @route("/menu/translations")
  upsertMenuTranslation(...UpsertMenuTranslationRequest): UpsertMenuTranslationResponse;

  /** Auto-translates menu entries into target locales while respecting manual edits. */
  @post
  @route("/menu/translations/ai")
  translateMenu(...TranslateMenuRequest): TranslateMenuResponse;

  // ── Currencies ──────────────────────────────────────────────────────────────

  /** Lists active/configured display currencies for a branch. */
  @get
  @route("/branding/currencies")
  listMenuCurrencies(...ListMenuCurrenciesRequest): ListMenuCurrenciesResponse;

  /** Adds or updates a branch currency (default USD + extra currencies). */
  @put
  @route("/branding/currencies")
  upsertMenuCurrency(...UpsertMenuCurrencyRequest): UpsertMenuCurrencyResponse;

  /** Deletes a non-default currency from a branch. */
  @delete
  @route("/branding/currencies/{code}")
  deleteMenuCurrency(...DeleteMenuCurrencyRequest): DeleteMenuCurrencyResponse;

  // ── Loyalty & Digital Stamp Cards ───────────────────────────────────────────

  /** Lists digital stamp card programs for the business/branch. */
  @get
  @route("/menu/stamp-cards")
  listStampCards(...ListStampCardsRequest): ListStampCardsResponse;

  /** Creates or updates a digital stamp card program. */
  @put
  @route("/menu/stamp-cards")
  upsertStampCard(...UpsertStampCardRequest): UpsertStampCardResponse;

  /** Deletes a digital stamp card program. */
  @delete
  @route("/menu/stamp-cards/{id}")
  deleteStampCard(...DeleteStampCardRequest): DeleteStampCardResponse;

  // ── Table Actions & Analytics (Authenticated) ───────────────────────────────

  /** Lists dine-in table action requests (call waiter, request bill, split bill). */
  @get
  @route("/orders/table-actions")
  listTableActions(...ListTableActionsRequest): ListTableActionsResponse;

  /** Marks a table action acknowledged or resolved. */
  @put
  @route("/orders/table-actions/{id}")
  updateTableActionStatus(...UpdateTableActionStatusRequest): UpdateTableActionStatusResponse;

  /** Returns funnel analytics counts for QR scans, views, carts, and orders. */
  @get
  @route("/analytics/summary")
  getMenuAnalyticsSummary(...GetMenuAnalyticsSummaryRequest): GetMenuAnalyticsSummaryResponse;

  // ── Public Anonymous Endpoints ──────────────────────────────────────────────

  /** Resolves a verified custom domain hostname to a tenant business slug. */
  @get
  @route("/public/resolve-host")
  resolvePublicMenuHost(...ResolvePublicMenuHostRequest): ResolvePublicMenuHostResponse;

  /** Returns the published branded menu bundle for anonymous guests. */
  @get
  @route("/public/bundle")
  getPublicMenuBundle(...GetPublicMenuBundleRequest): GetPublicMenuBundleResponse;

  /** Records a guest dine-in table request (call waiter, request bill, split bill). */
  @post
  @route("/public/table-actions")
  createPublicTableAction(...CreatePublicTableActionRequest): CreatePublicTableActionResponse;

  /** Records anonymous funnel telemetry events (QR scan, view, add to cart). */
  @post
  @route("/public/events")
  trackPublicMenuEvents(...TrackPublicMenuEventsRequest): TrackPublicMenuEventsResponse;

  /** Looks up a guest's active digital stamp card progress by phone number. */
  @get
  @route("/public/stamps")
  getPublicCustomerStamps(...GetPublicCustomerStampsRequest): GetPublicCustomerStampsResponse;

  /** Captures a consenting guest's contact info from the published menu into CRM. */
  @post
  @route("/public/guest")
  capturePublicMenuGuest(...CapturePublicMenuGuestRequest): CapturePublicMenuGuestResponse;

  /** Submits a customer review for merchant moderation. */
  @post
  @route("/public/reviews")
  createPublicMenuReview(...CreatePublicMenuReviewRequest): CreatePublicMenuReviewResponse;

  // ── Reviews Moderation (Authenticated) ──────────────────────────────────────

  /** Lists submitted customer reviews for moderation in Menu Studio. */
  @get
  @route("/menu/reviews")
  listMenuReviews(...ListMenuReviewsRequest): ListMenuReviewsResponse;

  /** Approves, hides, or marks a customer review as pending. */
  @put
  @route("/menu/reviews/{id}")
  updateMenuReviewStatus(...UpdateMenuReviewStatusRequest): UpdateMenuReviewStatusResponse;

  /** Deletes a customer review permanently. */
  @delete
  @route("/menu/reviews/{id}")
  deleteMenuReview(...DeleteMenuReviewRequest): DeleteMenuReviewResponse;
}
