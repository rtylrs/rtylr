import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using pos.menu_item.v1;

namespace hub.menu.v1;

model ListMenuRequest {}

model ListMenuResponse {
  items: MenuItem[];
}

model UpdateMenuRequest {
  @minItems(1)
  items: MenuItem[];
}

model UpdateMenuResponse {
  success: boolean;
}

model MenuBanner {
  @encode(string)
  id: int64;

  title: string;
  body: string;
  image_url: string;
  link_url: string;
  position: int32;
  is_active?: boolean;
  starts_at: string;
  ends_at: string;
  created_at: string;
  updated_at: string;
}

model ListMenuBannersRequest {}

model ListMenuBannersResponse {
  banners: MenuBanner[];
}

model CreateMenuBannerRequest {
  @Onkt.required
  title: string;

  body: string;
  image_url: string;
  link_url: string;
  position: int32;
  is_active?: boolean;
  starts_at: string;
  ends_at: string;
}

model CreateMenuBannerResponse {
  banner: MenuBanner;
}

model UpdateMenuBannerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  title: string;

  body: string;
  image_url: string;
  link_url: string;
  position: int32;
  is_active: boolean;
  starts_at: string;
  ends_at: string;
}

model UpdateMenuBannerResponse {
  banner: MenuBanner;
}

model DeleteMenuBannerRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteMenuBannerResponse {
  success: boolean;
}

// ── Branded Menu Builder Profiles ─────────────────────────────────────────────

model MenuProfile {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  location_id: int64;

  /** Serialized JSON MenuLayoutConfig for the working draft. */
  draft_config_json: string;

  /** Serialized JSON MenuLayoutConfig currently published to guests. */
  published_config_json: string;

  /** Sanitized, scoped custom CSS in the working draft. */
  custom_css: string;

  /** Sanitized, scoped custom CSS published to guests. */
  published_custom_css: string;

  white_label_hide_badge: boolean;
  default_locale: string;
  enabled_locales: string[];
  default_currency: string;
  is_published: boolean;
  published_at: string;

  @encode(string)
  cloned_from_location_id: int64;

  created_at: string;
  updated_at: string;
}

model GetMenuProfileRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;
}

model GetMenuProfileResponse {
  profile: MenuProfile;
}

model SaveMenuProfileDraftRequest {
  @encode(string)
  location_id?: int64;

  @Onkt.required
  draft_config_json: string;

  custom_css?: string;
  white_label_hide_badge?: boolean;
  default_locale?: string;
  enabled_locales?: string[];
  default_currency?: string;
}

model SaveMenuProfileDraftResponse {
  profile: MenuProfile;
}

model PublishMenuProfileRequest {
  @encode(string)
  location_id?: int64;
}

model PublishMenuProfileResponse {
  profile: MenuProfile;
}

model CloneMenuProfileRequest {
  @encode(string)
  source_location_id: int64;

  @encode(string)
  target_location_id: int64;
}

model CloneMenuProfileResponse {
  profile: MenuProfile;
}

// ── Layout & Theme Presets ────────────────────────────────────────────────────

model MenuPreset {
  @encode(string)
  id: int64;

  name: string;
  config_json: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

model ListMenuPresetsRequest {}

model ListMenuPresetsResponse {
  presets: MenuPreset[];
}

model CreateMenuPresetRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  config_json: string;
}

model CreateMenuPresetResponse {
  preset: MenuPreset;
}

model UpdateMenuPresetRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  config_json: string;
}

model UpdateMenuPresetResponse {
  preset: MenuPreset;
}

model DeleteMenuPresetRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteMenuPresetResponse {
  success: boolean;
}

// ── Per-Branch POS Catalog Overrides & Manual Toggles ─────────────────────────

model BranchMenuItemConfig {
  @encode(string)
  id: int64;

  @encode(string)
  location_id: int64;

  @encode(string)
  menu_item_id: int64;

  is_visible: boolean;
  has_price_override: boolean;
  price_override: float64;
  display_order: int32;
  featured_badge: string;
  dietary_tags: string[];
  allergens: string[];
  calories: int32;
  media_urls: string[];

  @Onkt.encode("string")
  upsell_item_ids: int64[];

  updated_at: string;
}

model ListBranchMenuItemsRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;
}

model ListBranchMenuItemsResponse {
  items: BranchMenuItemConfig[];
}

model UpsertBranchMenuItemRequest {
  @encode(string)
  location_id?: int64;

  @minValueExclusive(0)
  @encode(string)
  menu_item_id: int64;

  is_visible?: boolean;
  has_price_override?: boolean;
  price_override?: float64;
  display_order?: int32;
  featured_badge?: string;
  dietary_tags?: string[];
  allergens?: string[];
  calories?: int32;
  media_urls?: string[];

  @Onkt.encode("string")
  upsell_item_ids: int64[];
}

model UpsertBranchMenuItemResponse {
  item: BranchMenuItemConfig;
}

// ── Multi-Language Translations (AI + Manual Overrides) ───────────────────────

model MenuTranslation {
  @encode(string)
  id: int64;

  @encode(string)
  location_id: int64;

  entity_type: string;
  entity_id: string;
  locale: string;
  field: string;
  value: string;
  source: string;
  updated_at: string;
}

model ListMenuTranslationsRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;

  @query("locale")
  locale?: string;
}

model ListMenuTranslationsResponse {
  translations: MenuTranslation[];
}

model UpsertMenuTranslationRequest {
  @encode(string)
  location_id?: int64;

  @Onkt.required
  entity_type: string;

  @Onkt.required
  entity_id: string;

  @Onkt.required
  locale: string;

  @Onkt.required
  field: string;

  @Onkt.required
  value: string;

  source?: string;
}

model UpsertMenuTranslationResponse {
  translation: MenuTranslation;
}

model MenuTranslationSourceEntry {
  entity_type: string;
  entity_id: string;
  field: string;
  source_text: string;
}

model TranslateMenuRequest {
  @encode(string)
  location_id?: int64;

  @minItems(1)
  target_locales: string[];

  overwrite_manual?: boolean;
  entries?: MenuTranslationSourceEntry[];
}

model TranslateMenuResponse {
  translated_count: int32;
  skipped_manual_count: int32;
  translations: MenuTranslation[];
}

// ── Dynamic Multi-Currency Support ────────────────────────────────────────────

model MenuCurrency {
  @encode(string)
  id: int64;

  @encode(string)
  location_id: int64;

  code: string;
  symbol: string;
  exchange_rate: float64;
  decimal_places: int32;
  is_default: boolean;
  is_active: boolean;
  position: int32;
}

model ListMenuCurrenciesRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;
}

model ListMenuCurrenciesResponse {
  currencies: MenuCurrency[];
}

model UpsertMenuCurrencyRequest {
  @encode(string)
  location_id?: int64;

  @Onkt.required
  code: string;

  @Onkt.required
  symbol: string;

  @minValueExclusive(0)
  exchange_rate: float64;

  decimal_places?: int32;
  is_default?: boolean;
  is_active?: boolean;
  position?: int32;
}

model UpsertMenuCurrencyResponse {
  currencies: MenuCurrency[];
}

model DeleteMenuCurrencyRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;

  @Onkt.required
  code: string;
}

model DeleteMenuCurrencyResponse {
  success: boolean;
}

// ── Loyalty & Digital Stamp Cards ─────────────────────────────────────────────

model MenuStampCard {
  @encode(string)
  id: int64;

  @encode(string)
  location_id: int64;

  title: string;
  description: string;
  stamps_required: int32;
  reward_type: string;
  reward_label: string;

  @encode(string)
  reward_item_id: int64;

  reward_value: float64;
  qualifying_min_amount: float64;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model ListStampCardsRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;
}

model ListStampCardsResponse {
  cards: MenuStampCard[];
}

model UpsertStampCardRequest {
  @encode(string)
  id?: int64;

  @encode(string)
  location_id?: int64;

  @Onkt.required
  title: string;

  description?: string;

  @minValue(1)
  stamps_required: int32;

  reward_type?: string;
  reward_label?: string;

  @encode(string)
  reward_item_id?: int64;

  reward_value?: float64;
  qualifying_min_amount?: float64;
  is_active?: boolean;
}

model UpsertStampCardResponse {
  card: MenuStampCard;
}

model DeleteStampCardRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteStampCardResponse {
  success: boolean;
}

model CustomerStampProgress {
  @encode(string)
  stamp_card_id: int64;

  card_title: string;
  stamps_count: int32;
  stamps_required: int32;
  rewards_unlocked: int32;
  rewards_redeemed: int32;
  reward_label: string;
}

model GetPublicCustomerStampsRequest {
  @minValueExclusive(0)
  @query("business_id")
  @encode(string)
  business_id: int64;

  @Onkt.required
  @query("phone")
  phone: string;
}

model GetPublicCustomerStampsResponse {
  progress: CustomerStampProgress[];
}

// ── Table Actions (Call Waiter / Request Bill / Split Bill) ───────────────────

model MenuTableAction {
  @encode(string)
  id: int64;

  @encode(string)
  location_id: int64;

  @encode(string)
  table_id: int64;

  table_label: string;
  action_type: string;
  status: string;
  note: string;
  created_at: string;
  updated_at: string;
}

model CreatePublicTableActionRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;

  @encode(string)
  location_id?: int64;

  @encode(string)
  table_id?: int64;

  table_label?: string;

  @Onkt.required
  action_type: string;

  note?: string;
}

model CreatePublicTableActionResponse {
  action: MenuTableAction;
}

model ListTableActionsRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;

  @query("status")
  status?: string;
}

model ListTableActionsResponse {
  actions: MenuTableAction[];
}

model UpdateTableActionStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @Onkt.required
  status: string;
}

model UpdateTableActionStatusResponse {
  success: boolean;
}

// ── Funnel Analytics & Telemetry ──────────────────────────────────────────────

model PublicMenuEventInput {
  @encode(string)
  location_id?: int64;

  @encode(string)
  table_id?: int64;

  session_id?: string;

  @Onkt.required
  event_type: string;

  @encode(string)
  item_id?: int64;

  locale?: string;
  currency?: string;
}

model TrackPublicMenuEventsRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;

  @minItems(1)
  events: PublicMenuEventInput[];
}

model TrackPublicMenuEventsResponse {
  accepted: int32;
}

model MenuAnalyticsCount {
  event_type: string;
  count: int64;
}

model GetMenuAnalyticsSummaryRequest {
  @query("location_id")
  @encode(string)
  location_id?: int64;

  @query("days")
  days?: int32;
}

model GetMenuAnalyticsSummaryResponse {
  counts: MenuAnalyticsCount[];
  qr_scans: int64;
  menu_views: int64;
  item_views: int64;
  add_to_carts: int64;
  checkout_starts: int64;
  orders_placed: int64;
}

// ── Public Branded Storefront Bundle & Custom Domain Resolution ───────────────

model ResolvePublicMenuHostRequest {
  @Onkt.required
  @query("host")
  host: string;
}

model ResolvePublicMenuHostResponse {
  found: boolean;

  @encode(string)
  business_id: int64;

  business_slug: string;
  domain: string;
}

model GetPublicMenuBundleRequest {
  @minValueExclusive(0)
  @query("business_id")
  @encode(string)
  business_id: int64;

  @query("location_id")
  @encode(string)
  location_id?: int64;

  @query("locale")
  locale?: string;
}

model GetPublicMenuBundleResponse {
  profile: MenuProfile;
  banners: MenuBanner[];
  branch_items: BranchMenuItemConfig[];
  currencies: MenuCurrency[];
  translations: MenuTranslation[];
  stamp_cards: MenuStampCard[];
}

// ── Public Guest CRM Capture ─────────────────────────────────────────────────

model CapturePublicMenuGuestRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;

  @encode(string)
  location_id?: int64;

  name?: string;
  phone?: string;
  email?: string;
  notes?: string;
  consent: boolean;
}

model CapturePublicMenuGuestResponse {
  captured: boolean;
}
