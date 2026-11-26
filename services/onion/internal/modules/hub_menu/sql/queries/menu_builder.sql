-- name: GetMenuProfile :one
SELECT
    id,
    business_id,
    location_id,
    draft_config,
    COALESCE(published_config, JSON_OBJECT()) AS published_config,
    COALESCE(custom_css, '') AS custom_css,
    COALESCE(published_custom_css, '') AS published_custom_css,
    white_label_hide_badge,
    default_locale,
    enabled_locales,
    default_currency,
    is_published,
    published_at,
    cloned_from_location_id,
    created_at,
    updated_at
FROM hub_menu_profiles
WHERE business_id = ? AND location_id = ?
LIMIT 1;

-- name: UpsertMenuProfileDraft :exec
INSERT INTO hub_menu_profiles (
    business_id,
    location_id,
    draft_config,
    custom_css,
    white_label_hide_badge,
    default_locale,
    enabled_locales,
    default_currency,
    cloned_from_location_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON DUPLICATE KEY UPDATE
    draft_config = VALUES(draft_config),
    custom_css = VALUES(custom_css),
    white_label_hide_badge = VALUES(white_label_hide_badge),
    default_locale = VALUES(default_locale),
    enabled_locales = VALUES(enabled_locales),
    default_currency = VALUES(default_currency),
    cloned_from_location_id = COALESCE(VALUES(cloned_from_location_id), cloned_from_location_id),
    updated_at = CURRENT_TIMESTAMP;

-- name: PublishMenuProfile :exec
UPDATE hub_menu_profiles
SET published_config = draft_config,
    published_custom_css = custom_css,
    is_published = 1,
    published_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND location_id = ?;

-- name: ListMenuPresets :many
SELECT
    id,
    business_id,
    name,
    config,
    is_system,
    created_at,
    updated_at
FROM hub_menu_presets
WHERE business_id = ?
ORDER BY id ASC
LIMIT 100;

-- name: GetMenuPreset :one
SELECT
    id,
    business_id,
    name,
    config,
    is_system,
    created_at,
    updated_at
FROM hub_menu_presets
WHERE business_id = ? AND id = ?
LIMIT 1;

-- name: CreateMenuPreset :execresult
INSERT INTO hub_menu_presets (business_id, name, config, is_system)
VALUES (?, ?, ?, ?);

-- name: UpdateMenuPreset :exec
UPDATE hub_menu_presets
SET name = ?,
    config = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND id = ?;

-- name: DeleteMenuPreset :exec
DELETE FROM hub_menu_presets
WHERE business_id = ? AND id = ?;

-- name: ListBranchMenuItems :many
SELECT
    id,
    business_id,
    location_id,
    menu_item_id,
    is_visible,
    price_override,
    display_order,
    COALESCE(featured_badge, '') AS featured_badge,
    dietary_tags,
    allergens,
    calories,
    media_urls,
    upsell_item_ids,
    created_at,
    updated_at
FROM hub_menu_branch_items
WHERE business_id = ? AND location_id = ?
ORDER BY display_order ASC, id ASC
LIMIT 1000;

-- name: UpsertBranchMenuItem :exec
INSERT INTO hub_menu_branch_items (
    business_id,
    location_id,
    menu_item_id,
    is_visible,
    price_override,
    display_order,
    featured_badge,
    dietary_tags,
    allergens,
    calories,
    media_urls,
    upsell_item_ids
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON DUPLICATE KEY UPDATE
    is_visible = VALUES(is_visible),
    price_override = VALUES(price_override),
    display_order = VALUES(display_order),
    featured_badge = VALUES(featured_badge),
    dietary_tags = VALUES(dietary_tags),
    allergens = VALUES(allergens),
    calories = VALUES(calories),
    media_urls = VALUES(media_urls),
    upsell_item_ids = VALUES(upsell_item_ids),
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteBranchMenuItemsForLocation :exec
DELETE FROM hub_menu_branch_items
WHERE business_id = ? AND location_id = ?;

-- name: ListMenuTranslations :many
SELECT
    id,
    business_id,
    location_id,
    entity_type,
    entity_id,
    locale,
    field,
    value,
    source,
    created_at,
    updated_at
FROM hub_menu_translations
WHERE business_id = ? AND location_id = ? AND (? = '' OR locale = ?)
ORDER BY entity_type ASC, entity_id ASC, locale ASC, field ASC
LIMIT 5000;

-- name: UpsertMenuTranslation :exec
INSERT INTO hub_menu_translations (
    business_id,
    location_id,
    entity_type,
    entity_id,
    locale,
    field,
    value,
    source
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) ON DUPLICATE KEY UPDATE
    value = VALUES(value),
    source = VALUES(source),
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteMenuTranslationsForLocation :exec
DELETE FROM hub_menu_translations
WHERE business_id = ? AND location_id = ?;

-- name: ListMenuCurrencies :many
SELECT
    id,
    business_id,
    location_id,
    code,
    symbol,
    exchange_rate,
    decimal_places,
    is_default,
    is_active,
    position,
    created_at,
    updated_at
FROM hub_menu_currencies
WHERE business_id = ? AND location_id = ?
ORDER BY is_default DESC, position ASC, code ASC
LIMIT 50;

-- name: ClearDefaultMenuCurrency :exec
UPDATE hub_menu_currencies
SET is_default = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND location_id = ?;

-- name: UpsertMenuCurrency :exec
INSERT INTO hub_menu_currencies (
    business_id,
    location_id,
    code,
    symbol,
    exchange_rate,
    decimal_places,
    is_default,
    is_active,
    position
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON DUPLICATE KEY UPDATE
    symbol = VALUES(symbol),
    exchange_rate = VALUES(exchange_rate),
    decimal_places = VALUES(decimal_places),
    is_default = VALUES(is_default),
    is_active = VALUES(is_active),
    position = VALUES(position),
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteMenuCurrency :exec
DELETE FROM hub_menu_currencies
WHERE business_id = ? AND location_id = ? AND code = ?;

-- name: DeleteMenuCurrenciesForLocation :exec
DELETE FROM hub_menu_currencies
WHERE business_id = ? AND location_id = ?;

-- name: ListStampCards :many
SELECT
    id,
    business_id,
    location_id,
    title,
    COALESCE(description, '') AS description,
    stamps_required,
    reward_type,
    reward_label,
    reward_item_id,
    reward_value,
    qualifying_min_amount,
    is_active,
    created_at,
    updated_at
FROM hub_menu_stamp_cards
WHERE business_id = ? AND (location_id = ? OR location_id = 0)
ORDER BY id ASC
LIMIT 50;

-- name: GetStampCard :one
SELECT
    id,
    business_id,
    location_id,
    title,
    COALESCE(description, '') AS description,
    stamps_required,
    reward_type,
    reward_label,
    reward_item_id,
    reward_value,
    qualifying_min_amount,
    is_active,
    created_at,
    updated_at
FROM hub_menu_stamp_cards
WHERE business_id = ? AND id = ?
LIMIT 1;

-- name: CreateStampCard :execresult
INSERT INTO hub_menu_stamp_cards (
    business_id,
    location_id,
    title,
    description,
    stamps_required,
    reward_type,
    reward_label,
    reward_item_id,
    reward_value,
    qualifying_min_amount,
    is_active
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateStampCard :exec
UPDATE hub_menu_stamp_cards
SET location_id = ?,
    title = ?,
    description = ?,
    stamps_required = ?,
    reward_type = ?,
    reward_label = ?,
    reward_item_id = ?,
    reward_value = ?,
    qualifying_min_amount = ?,
    is_active = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND id = ?;

-- name: DeleteStampCard :exec
DELETE FROM hub_menu_stamp_cards
WHERE business_id = ? AND id = ?;

-- name: GetCustomerStampProgress :many
SELECT
    cs.id,
    cs.business_id,
    cs.stamp_card_id,
    cs.customer_phone,
    COALESCE(cs.customer_name, '') AS customer_name,
    cs.stamps_count,
    cs.rewards_unlocked,
    cs.rewards_redeemed,
    cs.last_order_id,
    sc.title AS card_title,
    sc.stamps_required,
    sc.reward_label
FROM hub_menu_customer_stamps cs
JOIN hub_menu_stamp_cards sc ON sc.id = cs.stamp_card_id AND sc.business_id = cs.business_id
WHERE cs.business_id = ? AND cs.customer_phone = ? AND sc.is_active = 1
ORDER BY cs.id ASC;

-- name: AddCustomerStamp :exec
INSERT INTO hub_menu_customer_stamps (
    business_id,
    stamp_card_id,
    customer_phone,
    customer_name,
    stamps_count,
    rewards_unlocked,
    last_order_id
) VALUES (
    ?, ?, ?, ?, 1, IF(1 >= ?, 1, 0), ?
) ON DUPLICATE KEY UPDATE
    customer_name = COALESCE(NULLIF(VALUES(customer_name), ''), customer_name),
    stamps_count = stamps_count + 1,
    rewards_unlocked = FLOOR(stamps_count / ?),
    last_order_id = VALUES(last_order_id),
    updated_at = CURRENT_TIMESTAMP;

-- name: CreateTableAction :execresult
INSERT INTO hub_menu_table_actions (
    business_id,
    location_id,
    table_id,
    table_label,
    action_type,
    status,
    note
) VALUES (?, ?, ?, ?, ?, 'open', ?);

-- name: ListTableActions :many
SELECT
    id,
    business_id,
    location_id,
    table_id,
    table_label,
    action_type,
    status,
    COALESCE(note, '') AS note,
    created_at,
    updated_at
FROM hub_menu_table_actions
WHERE business_id = ? AND (? = 0 OR location_id = ?) AND (? = '' OR status = ?)
ORDER BY created_at DESC
LIMIT 200;

-- name: UpdateTableActionStatus :exec
UPDATE hub_menu_table_actions
SET status = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE business_id = ? AND id = ?;

-- name: InsertMenuEvent :exec
INSERT INTO hub_menu_events (
    business_id,
    location_id,
    table_id,
    session_id,
    event_type,
    item_id,
    locale,
    currency
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: SummarizeMenuEvents :many
SELECT
    event_type,
    COUNT(*) AS total_count
FROM hub_menu_events
WHERE business_id = ?
  AND (? = 0 OR location_id = ?)
  AND created_at >= ?
GROUP BY event_type;

-- name: ListActivePublicBanners :many
SELECT
    id,
    business_id,
    title,
    COALESCE(body, '') AS body,
    COALESCE(image_url, '') AS image_url,
    COALESCE(link_url, '') AS link_url,
    position,
    is_active,
    starts_at,
    ends_at,
    created_at,
    updated_at
FROM hub_menu_banners
WHERE business_id = ?
  AND is_active = 1
  AND (starts_at IS NULL OR starts_at <= CURRENT_TIMESTAMP)
  AND (ends_at IS NULL OR ends_at >= CURRENT_TIMESTAMP)
ORDER BY position ASC, id ASC
LIMIT 50;
