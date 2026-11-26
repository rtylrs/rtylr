package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// HostResolver resolves a verified custom domain hostname to (businessID, slug, found).
type HostResolver func(ctx context.Context, host string) (businessID int64, businessSlug string, found bool, err error)

// EventDispatcher emits domain events to RTYLR's webhook / eventbus fanout.
type EventDispatcher func(ctx context.Context, event eventbus.Event) error

// AITranslator translates a batch of source strings into a target locale.
type AITranslator func(ctx context.Context, targetLocale string, inputs []string) ([]string, error)

// FeatureFlagChecker checks a global platform super-admin feature flag.
type FeatureFlagChecker func(ctx context.Context, key string, defaultVal bool) bool

// GuestCapturer saves a consenting guest from the published menu into CRM.
type GuestCapturer func(ctx context.Context, businessID, locationID int64, name, phone, email, notes string) error

// TenantValidator verifies that referenced branch/table/item IDs belong to businessID.
type TenantValidator struct {
	ValidateLocation func(ctx context.Context, businessID, locationID int64) error
	ValidateTable    func(ctx context.Context, businessID, tableID int64) error
	ValidateMenuItem func(ctx context.Context, businessID, menuItemID int64) error
}

func (s *Service) WithHostResolver(fn HostResolver) *Service {
	s.hostResolver = fn
	return s
}

func (s *Service) WithEventDispatcher(fn EventDispatcher) *Service {
	s.eventDispatcher = fn
	return s
}

func (s *Service) WithTranslator(fn AITranslator) *Service {
	s.translator = fn
	return s
}

func (s *Service) WithTenantValidator(v TenantValidator) *Service {
	s.tenantValidator = v
	return s
}

func (s *Service) WithFeatureFlagChecker(fn FeatureFlagChecker) *Service {
	s.featureFlagChecker = fn
	return s
}

func (s *Service) WithGuestCapturer(fn GuestCapturer) *Service {
	s.guestCapturer = fn
	return s
}

func (s *Service) isPlatformFeatureEnabled(ctx context.Context, key string) bool {
	if s.featureFlagChecker == nil {
		return true
	}
	return s.featureFlagChecker(ctx, key, true)
}

func (s *Service) validateLocation(ctx context.Context, businessID, locationID int64) error {
	if locationID <= 0 || s.tenantValidator.ValidateLocation == nil {
		return nil
	}
	if err := s.tenantValidator.ValidateLocation(ctx, businessID, locationID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrMenuProfileInvalidInput, err)
	}
	return nil
}

func (s *Service) validateTable(ctx context.Context, businessID, tableID int64) error {
	if tableID <= 0 || s.tenantValidator.ValidateTable == nil {
		return nil
	}
	if err := s.tenantValidator.ValidateTable(ctx, businessID, tableID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrTableActionInvalidInput, err)
	}
	return nil
}

func (s *Service) validateMenuItem(ctx context.Context, businessID, menuItemID int64) error {
	if menuItemID <= 0 || s.tenantValidator.ValidateMenuItem == nil {
		return nil
	}
	if err := s.tenantValidator.ValidateMenuItem(ctx, businessID, menuItemID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrBranchItemInvalidInput, err)
	}
	return nil
}

func (s *Service) GetMenuProfile(ctx context.Context, businessID, locationID int64) (domain.MenuProfile, error) {
	if locationID < 0 {
		locationID = 0
	}
	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return domain.MenuProfile{}, err
	}
	if !found {
		return defaultMenuProfile(businessID, locationID), nil
	}
	return profile, nil
}

func (s *Service) SaveMenuProfileDraft(ctx context.Context, businessID int64, input domain.SaveMenuProfileDraftInput) (domain.MenuProfile, error) {
	if input.LocationID < 0 {
		return domain.MenuProfile{}, fmt.Errorf("%w: location_id must be non-negative", domain.ErrMenuProfileInvalidInput)
	}
	if err := s.validateLocation(ctx, businessID, input.LocationID); err != nil {
		return domain.MenuProfile{}, err
	}
	trimmedJSON := strings.TrimSpace(input.DraftConfigJSON)
	if trimmedJSON == "" || !json.Valid([]byte(trimmedJSON)) {
		return domain.MenuProfile{}, fmt.Errorf("%w: draft_config_json must be valid JSON", domain.ErrMenuProfileInvalidInput)
	}

	current, found, err := s.repo.GetMenuProfile(ctx, businessID, input.LocationID)
	if err != nil {
		return domain.MenuProfile{}, err
	}
	if !found {
		current = defaultMenuProfile(businessID, input.LocationID)
	}

	current.DraftConfigJSON = trimmedJSON
	if input.CustomCSS != nil {
		sanitized, cssErr := domain.SanitizeCustomCSS(*input.CustomCSS)
		if cssErr != nil {
			return domain.MenuProfile{}, cssErr
		}
		current.CustomCSS = sanitized
	}
	current.WhiteLabelHideBadge = false
	if input.DefaultLocale != nil {
		loc := normalizeLocale(*input.DefaultLocale)
		if loc == "" {
			return domain.MenuProfile{}, fmt.Errorf("%w: invalid default_locale", domain.ErrMenuProfileInvalidInput)
		}
		current.DefaultLocale = loc
	}
	if len(input.EnabledLocales) > 0 {
		current.EnabledLocales = normalizeLocales(input.EnabledLocales, current.DefaultLocale)
	}
	if input.DefaultCurrency != nil {
		code := strings.ToUpper(strings.TrimSpace(*input.DefaultCurrency))
		if len(code) < 2 || len(code) > 12 {
			return domain.MenuProfile{}, fmt.Errorf("%w: invalid default_currency", domain.ErrMenuProfileInvalidInput)
		}
		current.DefaultCurrency = code
	}

	if err := s.repo.UpsertMenuProfileDraft(ctx, current); err != nil {
		return domain.MenuProfile{}, err
	}
	return s.GetMenuProfile(ctx, businessID, input.LocationID)
}

func (s *Service) PublishMenuProfile(ctx context.Context, businessID, locationID int64) (domain.MenuProfile, error) {
	if locationID < 0 {
		locationID = 0
	}
	if err := s.validateLocation(ctx, businessID, locationID); err != nil {
		return domain.MenuProfile{}, err
	}
	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return domain.MenuProfile{}, err
	}
	if !found {
		profile = defaultMenuProfile(businessID, locationID)
		if err := s.repo.UpsertMenuProfileDraft(ctx, profile); err != nil {
			return domain.MenuProfile{}, err
		}
	}
	if err := s.repo.PublishMenuProfile(ctx, businessID, locationID); err != nil {
		return domain.MenuProfile{}, err
	}
	published, err := s.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return domain.MenuProfile{}, err
	}
	s.emitEvent(ctx, businessID, "menu.profile.published", map[string]any{
		"business_id":  businessID,
		"location_id":  locationID,
		"published_at": published.PublishedAt.UTC().Format(time.RFC3339),
	})
	return published, nil
}

func (s *Service) CloneMenuProfile(ctx context.Context, businessID, sourceLocationID, targetLocationID int64) (domain.MenuProfile, error) {
	if sourceLocationID < 0 || targetLocationID < 0 {
		return domain.MenuProfile{}, fmt.Errorf("%w: location ids must be non-negative", domain.ErrMenuProfileInvalidInput)
	}
	if sourceLocationID == targetLocationID {
		return domain.MenuProfile{}, fmt.Errorf("%w: source and target locations must differ", domain.ErrMenuProfileInvalidInput)
	}
	if err := s.validateLocation(ctx, businessID, sourceLocationID); err != nil {
		return domain.MenuProfile{}, err
	}
	if err := s.validateLocation(ctx, businessID, targetLocationID); err != nil {
		return domain.MenuProfile{}, err
	}
	sourceProfile, err := s.GetMenuProfile(ctx, businessID, sourceLocationID)
	if err != nil {
		return domain.MenuProfile{}, err
	}
	if err := s.repo.CloneMenuProfile(ctx, businessID, sourceLocationID, targetLocationID, sourceProfile); err != nil {
		return domain.MenuProfile{}, err
	}
	return s.GetMenuProfile(ctx, businessID, targetLocationID)
}

func (s *Service) ListMenuPresets(ctx context.Context, businessID int64) ([]domain.MenuPreset, error) {
	return s.repo.ListMenuPresets(ctx, businessID)
}

func (s *Service) CreateMenuPreset(ctx context.Context, businessID int64, name, configJSON string) (domain.MenuPreset, error) {
	name = strings.TrimSpace(name)
	configJSON = strings.TrimSpace(configJSON)
	if name == "" || len(name) > 120 {
		return domain.MenuPreset{}, fmt.Errorf("%w: name must be 1-120 characters", domain.ErrMenuPresetInvalidInput)
	}
	if configJSON == "" || !json.Valid([]byte(configJSON)) {
		return domain.MenuPreset{}, fmt.Errorf("%w: config_json must be valid JSON", domain.ErrMenuPresetInvalidInput)
	}
	id, err := s.repo.CreateMenuPreset(ctx, businessID, name, configJSON)
	if err != nil {
		return domain.MenuPreset{}, err
	}
	return s.repo.GetMenuPreset(ctx, businessID, id)
}

func (s *Service) UpdateMenuPreset(ctx context.Context, businessID, id int64, name, configJSON string) (domain.MenuPreset, error) {
	if id <= 0 {
		return domain.MenuPreset{}, domain.ErrMenuPresetNotFound
	}
	name = strings.TrimSpace(name)
	configJSON = strings.TrimSpace(configJSON)
	if name == "" || len(name) > 120 {
		return domain.MenuPreset{}, fmt.Errorf("%w: name must be 1-120 characters", domain.ErrMenuPresetInvalidInput)
	}
	if configJSON == "" || !json.Valid([]byte(configJSON)) {
		return domain.MenuPreset{}, fmt.Errorf("%w: config_json must be valid JSON", domain.ErrMenuPresetInvalidInput)
	}
	if err := s.repo.UpdateMenuPreset(ctx, businessID, id, name, configJSON); err != nil {
		return domain.MenuPreset{}, err
	}
	return s.repo.GetMenuPreset(ctx, businessID, id)
}

func (s *Service) DeleteMenuPreset(ctx context.Context, businessID, id int64) error {
	if id <= 0 {
		return domain.ErrMenuPresetNotFound
	}
	return s.repo.DeleteMenuPreset(ctx, businessID, id)
}

func (s *Service) ListBranchMenuItems(ctx context.Context, businessID, locationID int64) ([]domain.BranchMenuItem, error) {
	if locationID < 0 {
		locationID = 0
	}
	return s.repo.ListBranchMenuItems(ctx, businessID, locationID)
}

func (s *Service) UpsertBranchMenuItem(ctx context.Context, businessID int64, item domain.BranchMenuItem) (domain.BranchMenuItem, error) {
	if item.MenuItemID <= 0 {
		return domain.BranchMenuItem{}, fmt.Errorf("%w: menu_item_id must be positive", domain.ErrBranchItemInvalidInput)
	}
	if item.LocationID < 0 {
		item.LocationID = 0
	}
	if err := s.validateLocation(ctx, businessID, item.LocationID); err != nil {
		return domain.BranchMenuItem{}, err
	}
	if err := s.validateMenuItem(ctx, businessID, item.MenuItemID); err != nil {
		return domain.BranchMenuItem{}, err
	}
	if item.HasPriceOverride && item.PriceOverride < 0 {
		return domain.BranchMenuItem{}, fmt.Errorf("%w: price_override must be non-negative", domain.ErrBranchItemInvalidInput)
	}
	item.BusinessID = businessID
	item.FeaturedBadge = strings.TrimSpace(item.FeaturedBadge)
	if len(item.FeaturedBadge) > 64 {
		item.FeaturedBadge = item.FeaturedBadge[:64]
	}
	if err := s.repo.UpsertBranchMenuItem(ctx, item); err != nil {
		return domain.BranchMenuItem{}, err
	}
	items, err := s.repo.ListBranchMenuItems(ctx, businessID, item.LocationID)
	if err != nil {
		return domain.BranchMenuItem{}, err
	}
	for _, stored := range items {
		if stored.MenuItemID == item.MenuItemID {
			return stored, nil
		}
	}
	return item, nil
}

func (s *Service) ListMenuTranslations(ctx context.Context, businessID, locationID int64, locale string) ([]domain.MenuTranslation, error) {
	if locationID < 0 {
		locationID = 0
	}
	return s.repo.ListMenuTranslations(ctx, businessID, locationID, normalizeLocale(locale))
}

func (s *Service) UpsertMenuTranslation(ctx context.Context, businessID int64, tr domain.MenuTranslation) (domain.MenuTranslation, error) {
	tr.EntityType = strings.ToLower(strings.TrimSpace(tr.EntityType))
	tr.EntityID = strings.TrimSpace(tr.EntityID)
	tr.Locale = normalizeLocale(tr.Locale)
	tr.Field = strings.ToLower(strings.TrimSpace(tr.Field))
	tr.Value = strings.TrimSpace(tr.Value)
	if tr.LocationID < 0 {
		tr.LocationID = 0
	}
	if err := s.validateLocation(ctx, businessID, tr.LocationID); err != nil {
		return domain.MenuTranslation{}, err
	}
	if tr.EntityType == "" || tr.EntityID == "" || tr.Locale == "" || tr.Field == "" || tr.Value == "" {
		return domain.MenuTranslation{}, fmt.Errorf("%w: entity_type, entity_id, locale, field, and value are required", domain.ErrTranslationInvalidInput)
	}
	if tr.Source != "ai" {
		tr.Source = "manual"
	}
	tr.BusinessID = businessID
	if err := s.repo.UpsertMenuTranslation(ctx, tr); err != nil {
		return domain.MenuTranslation{}, err
	}
	tr.UpdatedAt = time.Now().UTC()
	return tr, nil
}

func (s *Service) TranslateMenu(
	ctx context.Context,
	businessID, locationID int64,
	targetLocales []string,
	overwriteManual bool,
	entries []domain.TranslationSourceEntry,
) (translatedCount int32, skippedManual int32, results []domain.MenuTranslation, err error) {
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.aiTranslate") {
		return 0, 0, nil, fmt.Errorf("%w: AI menu translation is disabled", domain.ErrFeatureDisabled)
	}
	if locationID < 0 {
		locationID = 0
	}
	if err := s.validateLocation(ctx, businessID, locationID); err != nil {
		return 0, 0, nil, err
	}
	if len(targetLocales) == 0 {
		return 0, 0, nil, fmt.Errorf("%w: at least one target locale is required", domain.ErrTranslationInvalidInput)
	}

	existing, err := s.repo.ListMenuTranslations(ctx, businessID, locationID, "")
	if err != nil {
		return 0, 0, nil, err
	}
	manualSet := make(map[string]struct{}, len(existing))
	for _, ex := range existing {
		if ex.Source == "manual" {
			manualSet[translationKey(ex.EntityType, ex.EntityID, ex.Locale, ex.Field)] = struct{}{}
		}
	}

	for _, rawLocale := range targetLocales {
		loc := normalizeLocale(rawLocale)
		if loc == "" {
			continue
		}
		var pending []domain.TranslationSourceEntry
		var texts []string
		for _, entry := range entries {
			srcText := strings.TrimSpace(entry.SourceText)
			if srcText == "" {
				continue
			}
			key := translationKey(entry.EntityType, entry.EntityID, loc, entry.Field)
			if !overwriteManual {
				if _, isManual := manualSet[key]; isManual {
					skippedManual++
					continue
				}
			}
			pending = append(pending, entry)
			texts = append(texts, srcText)
		}
		if len(pending) == 0 {
			continue
		}

		var translatedTexts []string
		if s.translator != nil {
			translatedTexts, err = s.translator(ctx, loc, texts)
			if err != nil {
				return translatedCount, skippedManual, results, err
			}
		} else {
			translatedTexts = texts
		}

		for i, entry := range pending {
			val := entry.SourceText
			if i < len(translatedTexts) && strings.TrimSpace(translatedTexts[i]) != "" {
				val = strings.TrimSpace(translatedTexts[i])
			}
			stored, upsertErr := s.UpsertMenuTranslation(ctx, businessID, domain.MenuTranslation{
				LocationID: locationID,
				EntityType: entry.EntityType,
				EntityID:   entry.EntityID,
				Locale:     loc,
				Field:      entry.Field,
				Value:      val,
				Source:     "ai",
			})
			if upsertErr != nil {
				return translatedCount, skippedManual, results, upsertErr
			}
			results = append(results, stored)
			translatedCount++
		}
	}

	return translatedCount, skippedManual, results, nil
}

func (s *Service) ListMenuCurrencies(ctx context.Context, businessID, locationID int64) ([]domain.MenuCurrency, error) {
	if locationID < 0 {
		locationID = 0
	}
	list, err := s.repo.ListMenuCurrencies(ctx, businessID, locationID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 && locationID > 0 {
		// Inherit default brand currencies (location_id = 0) when a branch has
		// not configured branch-specific currency overrides.
		inherited, inheritErr := s.repo.ListMenuCurrencies(ctx, businessID, 0)
		if inheritErr != nil {
			return nil, inheritErr
		}
		if len(inherited) > 0 {
			return inherited, nil
		}
	}
	if len(list) == 0 {
		return []domain.MenuCurrency{{
			BusinessID:    businessID,
			LocationID:    locationID,
			Code:          "USD",
			Symbol:        "$",
			ExchangeRate:  1,
			DecimalPlaces: 2,
			IsDefault:     true,
			IsActive:      true,
			Position:      0,
		}}, nil
	}
	return list, nil
}

func (s *Service) UpsertMenuCurrency(ctx context.Context, businessID int64, cur domain.MenuCurrency) ([]domain.MenuCurrency, error) {
	cur.Code = strings.ToUpper(strings.TrimSpace(cur.Code))
	cur.Symbol = strings.TrimSpace(cur.Symbol)
	if cur.LocationID < 0 {
		cur.LocationID = 0
	}
	if err := s.validateLocation(ctx, businessID, cur.LocationID); err != nil {
		return nil, err
	}
	if len(cur.Code) < 2 || len(cur.Code) > 12 {
		return nil, fmt.Errorf("%w: code must be 2-12 uppercase letters", domain.ErrCurrencyInvalidInput)
	}
	if cur.Symbol == "" || len(cur.Symbol) > 16 {
		return nil, fmt.Errorf("%w: symbol is required", domain.ErrCurrencyInvalidInput)
	}
	if cur.ExchangeRate <= 0 {
		return nil, fmt.Errorf("%w: exchange_rate must be positive", domain.ErrCurrencyInvalidInput)
	}
	if cur.DecimalPlaces < 0 || cur.DecimalPlaces > 4 {
		cur.DecimalPlaces = 2
	}
	cur.BusinessID = businessID
	existing, listErr := s.repo.ListMenuCurrencies(ctx, businessID, cur.LocationID)
	if listErr != nil {
		return nil, listErr
	}
	if len(existing) == 0 && !cur.IsDefault && cur.Code != "USD" {
		// Seed the base USD currency so adding a first secondary currency appends
		// alongside USD instead of replacing the synthetic default.
		if seedErr := s.repo.UpsertMenuCurrency(ctx, domain.MenuCurrency{
			BusinessID:    businessID,
			LocationID:    cur.LocationID,
			Code:          "USD",
			Symbol:        "$",
			ExchangeRate:  1,
			DecimalPlaces: 2,
			IsDefault:     true,
			IsActive:      true,
			Position:      0,
		}); seedErr != nil {
			return nil, seedErr
		}
		if cur.Position == 0 {
			cur.Position = 1
		}
	}
	if err := s.repo.UpsertMenuCurrency(ctx, cur); err != nil {
		return nil, err
	}
	return s.ListMenuCurrencies(ctx, businessID, cur.LocationID)
}

func (s *Service) DeleteMenuCurrency(ctx context.Context, businessID, locationID int64, code string) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("%w: code is required", domain.ErrCurrencyInvalidInput)
	}
	if locationID < 0 {
		locationID = 0
	}
	if err := s.validateLocation(ctx, businessID, locationID); err != nil {
		return err
	}
	return s.repo.DeleteMenuCurrency(ctx, businessID, locationID, code)
}

func (s *Service) ListStampCards(ctx context.Context, businessID, locationID int64) ([]domain.StampCard, error) {
	if locationID < 0 {
		locationID = 0
	}
	return s.repo.ListStampCards(ctx, businessID, locationID)
}

func (s *Service) UpsertStampCard(ctx context.Context, businessID int64, card domain.StampCard) (domain.StampCard, error) {
	card.Title = strings.TrimSpace(card.Title)
	card.Description = strings.TrimSpace(card.Description)
	card.RewardLabel = strings.TrimSpace(card.RewardLabel)
	if card.LocationID < 0 {
		card.LocationID = 0
	}
	if err := s.validateLocation(ctx, businessID, card.LocationID); err != nil {
		return domain.StampCard{}, err
	}
	if card.Title == "" || len(card.Title) > 160 {
		return domain.StampCard{}, fmt.Errorf("%w: title is required", domain.ErrStampCardInvalidInput)
	}
	if card.StampsRequired <= 0 {
		return domain.StampCard{}, fmt.Errorf("%w: stamps_required must be at least 1", domain.ErrStampCardInvalidInput)
	}
	card.BusinessID = businessID
	if card.ID > 0 {
		if err := s.repo.UpdateStampCard(ctx, card); err != nil {
			return domain.StampCard{}, err
		}
		return s.repo.GetStampCard(ctx, businessID, card.ID)
	}
	id, err := s.repo.CreateStampCard(ctx, card)
	if err != nil {
		return domain.StampCard{}, err
	}
	return s.repo.GetStampCard(ctx, businessID, id)
}

func (s *Service) DeleteStampCard(ctx context.Context, businessID, id int64) error {
	if id <= 0 {
		return domain.ErrStampCardNotFound
	}
	return s.repo.DeleteStampCard(ctx, businessID, id)
}

func (s *Service) GetPublicCustomerStamps(ctx context.Context, businessID int64, phone string) ([]domain.CustomerStampProgress, error) {
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.loyalty") {
		return []domain.CustomerStampProgress{}, nil
	}
	phone = strings.TrimSpace(phone)
	digits := 0
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	if businessID <= 0 || digits < 7 {
		return []domain.CustomerStampProgress{}, nil
	}
	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, 0)
	if err != nil {
		return nil, err
	}
	if found && profile.IsPublished && !domain.IsLoyaltyStampsEnabled(profile.PublishedConfigJSON) {
		return []domain.CustomerStampProgress{}, nil
	}
	return s.repo.GetCustomerStampProgress(ctx, businessID, phone)
}

func (s *Service) AwardStampsForOrder(ctx context.Context, businessID, locationID, orderID int64, phone, name string, orderTotal float64) error {
	phone = strings.TrimSpace(phone)
	if businessID <= 0 || phone == "" {
		return nil
	}
	cards, err := s.repo.ListStampCards(ctx, businessID, locationID)
	if err != nil {
		return err
	}
	for _, card := range cards {
		if !card.IsActive || orderTotal < card.QualifyingMinAmount {
			continue
		}
		if err := s.repo.AddCustomerStamp(ctx, businessID, card.ID, phone, name, card.StampsRequired, orderID); err != nil {
			return err
		}
		s.emitEvent(ctx, businessID, "menu.loyalty.stamp_earned", map[string]any{
			"business_id":    businessID,
			"location_id":    locationID,
			"order_id":       orderID,
			"stamp_card_id":  card.ID,
			"customer_phone": phone,
		})
	}
	return nil
}

func (s *Service) CreatePublicTableAction(ctx context.Context, businessID int64, action domain.TableAction) (domain.TableAction, error) {
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.tableActions") {
		return domain.TableAction{}, fmt.Errorf("%w: table actions are disabled", domain.ErrFeatureDisabled)
	}
	if businessID <= 0 {
		return domain.TableAction{}, fmt.Errorf("%w: business_id is required", domain.ErrTableActionInvalidInput)
	}
	if err := s.validateLocation(ctx, businessID, action.LocationID); err != nil {
		return domain.TableAction{}, err
	}
	if err := s.validateTable(ctx, businessID, action.TableID); err != nil {
		return domain.TableAction{}, err
	}
	action.ActionType = strings.ToLower(strings.TrimSpace(action.ActionType))
	switch action.ActionType {
	case "call_waiter", "request_bill", "split_bill":
	default:
		return domain.TableAction{}, fmt.Errorf("%w: unsupported action_type", domain.ErrTableActionInvalidInput)
	}
	action.BusinessID = businessID
	action.TableLabel = strings.TrimSpace(action.TableLabel)
	action.Note = strings.TrimSpace(action.Note)
	if len(action.Note) > 500 {
		action.Note = action.Note[:500]
	}
	id, err := s.repo.CreateTableAction(ctx, action)
	if err != nil {
		return domain.TableAction{}, err
	}
	action.ID = id
	action.Status = "open"
	action.CreatedAt = time.Now().UTC()
	action.UpdatedAt = action.CreatedAt

	s.emitEvent(ctx, businessID, "menu.table.action_requested", map[string]any{
		"action_id":   action.ID,
		"business_id": businessID,
		"location_id": action.LocationID,
		"table_id":    action.TableID,
		"table_label": action.TableLabel,
		"action_type": action.ActionType,
		"note":        action.Note,
	})
	return action, nil
}

func (s *Service) ListTableActions(ctx context.Context, businessID, locationID int64, status string) ([]domain.TableAction, error) {
	if locationID < 0 {
		locationID = 0
	}
	if err := s.validateLocation(ctx, businessID, locationID); err != nil {
		return nil, err
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" && status != "open" && status != "acknowledged" && status != "resolved" {
		return nil, fmt.Errorf("%w: invalid status", domain.ErrTableActionInvalidInput)
	}
	return s.repo.ListTableActions(ctx, businessID, locationID, status)
}

func (s *Service) UpdateTableActionStatus(ctx context.Context, businessID, id int64, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if id <= 0 || (status != "open" && status != "acknowledged" && status != "resolved") {
		return fmt.Errorf("%w: invalid table action status", domain.ErrTableActionInvalidInput)
	}
	return s.repo.UpdateTableActionStatus(ctx, businessID, id, status)
}

func (s *Service) TrackPublicMenuEvents(ctx context.Context, businessID int64, events []domain.MenuEventInput) (int32, error) {
	if businessID <= 0 || len(events) == 0 {
		return 0, fmt.Errorf("%w: business_id and at least one event are required", domain.ErrEventInvalidInput)
	}
	if len(events) > 50 {
		events = events[:50]
	}
	valid := make([]domain.MenuEventInput, 0, len(events))
	for _, ev := range events {
		if err := s.validateLocation(ctx, businessID, ev.LocationID); err != nil {
			return 0, fmt.Errorf("%w: invalid location_id", domain.ErrEventInvalidInput)
		}
		ev.EventType = strings.ToLower(strings.TrimSpace(ev.EventType))
		switch ev.EventType {
		case "qr_scan", "menu_view", "item_view", "add_to_cart", "checkout_start", "order_placed":
			ev.SessionID = strings.TrimSpace(ev.SessionID)
			if len(ev.SessionID) > 64 {
				ev.SessionID = ev.SessionID[:64]
			}
			valid = append(valid, ev)
		}
	}
	if len(valid) == 0 {
		return 0, fmt.Errorf("%w: no valid event types provided", domain.ErrEventInvalidInput)
	}
	if err := s.repo.InsertMenuEvents(ctx, businessID, valid); err != nil {
		return 0, err
	}
	return int32(len(valid)), nil
}

func (s *Service) GetMenuAnalyticsSummary(ctx context.Context, businessID, locationID int64, days int32) (domain.AnalyticsSummary, error) {
	if days <= 0 || days > 365 {
		days = 30
	}
	since := time.Now().UTC().AddDate(0, 0, -int(days))
	return s.repo.SummarizeMenuEvents(ctx, businessID, locationID, since)
}

func (s *Service) ResolvePublicMenuHost(ctx context.Context, host string) (domain.PublicHostResolution, error) {
	normalized := strings.ToLower(strings.TrimSpace(host))
	if idx := strings.Index(normalized, ":"); idx > 0 {
		normalized = normalized[:idx]
	}
	if normalized == "" || s.hostResolver == nil {
		return domain.PublicHostResolution{Found: false, Domain: normalized}, nil
	}
	businessID, slug, found, err := s.hostResolver(ctx, normalized)
	if err != nil {
		return domain.PublicHostResolution{}, err
	}
	return domain.PublicHostResolution{
		Found:        found,
		BusinessID:   businessID,
		BusinessSlug: slug,
		Domain:       normalized,
	}, nil
}

func (s *Service) GetPublicMenuBundle(ctx context.Context, businessID, locationID int64, locale string) (domain.PublicMenuBundle, error) {
	if businessID <= 0 {
		return domain.PublicMenuBundle{}, fmt.Errorf("%w: business_id is required", domain.ErrMenuProfileInvalidInput)
	}
	if locationID < 0 {
		locationID = 0
	}

	// Prefer a published branch profile; fall back to the published business
	// default (location_id = 0) when the branch has no profile or only an
	// unpublished draft so a draft never wipes the live brand storefront.
	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return domain.PublicMenuBundle{}, err
	}
	if (!found || !profile.IsPublished) && locationID > 0 {
		fallback, fallbackFound, fallbackErr := s.repo.GetMenuProfile(ctx, businessID, 0)
		if fallbackErr != nil {
			return domain.PublicMenuBundle{}, fallbackErr
		}
		if fallbackFound && fallback.IsPublished {
			profile = fallback
			found = true
		}
	}
	if !found || !profile.IsPublished {
		// Never expose unpublished draft config or draft CSS to public customers.
		profile = defaultMenuProfile(businessID, locationID)
		profile.DraftConfigJSON = ""
		profile.CustomCSS = ""
		profile.PublishedConfigJSON = ""
		profile.PublishedCustomCSS = ""
	} else {
		profile.DraftConfigJSON = ""
		profile.CustomCSS = ""
	}

	banners, err := s.repo.ListActivePublicBanners(ctx, businessID)
	if err != nil {
		return domain.PublicMenuBundle{}, err
	}
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.banners") {
		banners = nil
	}
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.customCss") {
		profile.PublishedCustomCSS = ""
	}
	// Always honor branch-specific item overrides (POS 86'ing and branch pricing)
	// even when the branch inherits the default brand profile.
	branchItems, err := s.repo.ListBranchMenuItems(ctx, businessID, locationID)
	if err != nil {
		return domain.PublicMenuBundle{}, err
	}
	if len(branchItems) == 0 && locationID > 0 {
		branchItems, err = s.repo.ListBranchMenuItems(ctx, businessID, 0)
		if err != nil {
			return domain.PublicMenuBundle{}, err
		}
	}
	currencies, err := s.ListMenuCurrencies(ctx, businessID, locationID)
	if err != nil {
		return domain.PublicMenuBundle{}, err
	}
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.currencies") && len(currencies) > 1 {
		// Keep only the default currency when multi-currency switching is globally disabled.
		defaultOnly := make([]domain.MenuCurrency, 0, 1)
		for _, cur := range currencies {
			if cur.IsDefault {
				defaultOnly = append(defaultOnly, cur)
				break
			}
		}
		if len(defaultOnly) == 0 {
			defaultOnly = currencies[:1]
		}
		currencies = defaultOnly
	}
	normLocale := normalizeLocale(locale)
	translations, err := s.repo.ListMenuTranslations(ctx, businessID, locationID, normLocale)
	if err != nil {
		return domain.PublicMenuBundle{}, err
	}
	if len(translations) == 0 && locationID > 0 {
		translations, err = s.repo.ListMenuTranslations(ctx, businessID, 0, normLocale)
		if err != nil {
			return domain.PublicMenuBundle{}, err
		}
	}
	activeCards := []domain.StampCard{}
	if s.isPlatformFeatureEnabled(ctx, "app.menu.loyalty") {
		stampCards, err := s.repo.ListStampCards(ctx, businessID, locationID)
		if err != nil {
			return domain.PublicMenuBundle{}, err
		}
		for _, card := range stampCards {
			if card.IsActive {
				activeCards = append(activeCards, card)
			}
		}
	}

	return domain.PublicMenuBundle{
		Profile:      profile,
		Banners:      banners,
		BranchItems:  branchItems,
		Currencies:   currencies,
		Translations: translations,
		StampCards:   activeCards,
	}, nil
}

// CapturePublicMenuGuest validates and stores an optional consenting guest from
// the published digital menu into the business's CRM.
func (s *Service) CapturePublicMenuGuest(ctx context.Context, businessID, locationID int64, name, phone, email, notes string, consent bool) error {
	if businessID <= 0 {
		return fmt.Errorf("%w: business_id is required", domain.ErrGuestCaptureInvalidInput)
	}
	if locationID < 0 {
		locationID = 0
	}
	if err := s.validateLocation(ctx, businessID, locationID); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrGuestCaptureInvalidInput, err)
	}
	if !consent {
		return fmt.Errorf("%w: consent is required", domain.ErrGuestCaptureInvalidInput)
	}
	if !s.isPlatformFeatureEnabled(ctx, "app.menu.crmCapture") {
		return fmt.Errorf("%w: guest CRM capture is disabled", domain.ErrFeatureDisabled)
	}

	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	email = strings.ToLower(strings.TrimSpace(email))
	notes = strings.TrimSpace(notes)
	if phone == "" && email == "" {
		return fmt.Errorf("%w: phone or email is required", domain.ErrGuestCaptureInvalidInput)
	}
	if email != "" && (!strings.Contains(email, "@") || !strings.Contains(email, ".") || len(email) > 254) {
		return fmt.Errorf("%w: invalid email address", domain.ErrGuestCaptureInvalidInput)
	}
	if len(name) > 160 {
		name = name[:160]
	}
	if len(phone) > 40 {
		phone = phone[:40]
	}
	if len(notes) > 500 {
		notes = notes[:500]
	}

	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, locationID)
	if err != nil {
		return err
	}
	if !found && locationID > 0 {
		profile, found, err = s.repo.GetMenuProfile(ctx, businessID, 0)
		if err != nil {
			return err
		}
	}
	if found && profile.IsPublished && !domain.IsCRMCaptureEnabled(profile.PublishedConfigJSON) {
		return fmt.Errorf("%w: guest CRM capture is turned off for this menu", domain.ErrFeatureDisabled)
	}

	if s.guestCapturer != nil {
		if err := s.guestCapturer(ctx, businessID, locationID, name, phone, email, notes); err != nil {
			return err
		}
	}

	s.emitEvent(ctx, businessID, "menu.guest.captured", map[string]any{
		"business_id": businessID,
		"location_id": locationID,
		"has_phone":   phone != "",
		"has_email":   email != "",
	})
	return nil
}

// IsPublishedOrderingEnabled reports whether the business's published menu profile
// allows customer ordering (used to enable standalone Menu-only ordering without POS).
func (s *Service) IsPublishedOrderingEnabled(ctx context.Context, businessID int64) bool {
	if businessID <= 0 || !s.isPlatformFeatureEnabled(ctx, "app.menu.ordering") {
		return false
	}
	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, 0)
	if err != nil || !found || !profile.IsPublished {
		// Default menu profile allows ordering when the business enables Digital Menu.
		return true
	}
	return domain.IsOrderingEnabled(profile.PublishedConfigJSON)
}

// ResolvePaymentRedirectURL returns the configured third-party payment redirect URL
// strictly from a published profile when payment.mode == "redirect".
func (s *Service) ResolvePaymentRedirectURL(ctx context.Context, businessID, locationID int64) string {
	profile, found, err := s.repo.GetMenuProfile(ctx, businessID, locationID)
	if err != nil || !found || !profile.IsPublished {
		if locationID > 0 {
			profile, found, err = s.repo.GetMenuProfile(ctx, businessID, 0)
			if err != nil || !found || !profile.IsPublished {
				return ""
			}
		} else {
			return ""
		}
	}
	return domain.ExtractPaymentRedirectURL(profile.PublishedConfigJSON)
}

// ResolveBranchItemOverride returns visibility and optional branch price override for POS checkout.
func (s *Service) ResolveBranchItemOverride(ctx context.Context, businessID, locationID, menuItemID int64) (visible bool, hasPriceOverride bool, priceOverride float64, err error) {
	if businessID <= 0 || menuItemID <= 0 {
		return true, false, 0, nil
	}
	if locationID < 0 {
		locationID = 0
	}
	items, err := s.repo.ListBranchMenuItems(ctx, businessID, locationID)
	if err != nil {
		return true, false, 0, err
	}
	for _, item := range items {
		if item.MenuItemID == menuItemID {
			return item.IsVisible, item.HasPriceOverride, item.PriceOverride, nil
		}
	}
	if locationID > 0 {
		defaultItems, err := s.repo.ListBranchMenuItems(ctx, businessID, 0)
		if err != nil {
			return true, false, 0, err
		}
		for _, item := range defaultItems {
			if item.MenuItemID == menuItemID {
				return item.IsVisible, item.HasPriceOverride, item.PriceOverride, nil
			}
		}
	}
	return true, false, 0, nil
}

func (s *Service) emitEvent(ctx context.Context, businessID int64, eventType string, payload map[string]any) {
	if s.eventDispatcher == nil {
		return
	}
	_ = s.eventDispatcher(ctx, eventbus.Event{
		Type:        eventType,
		BusinessID:  businessID,
		SubjectType: "menu",
		Payload:     payload,
		OccurredAt:  time.Now().UTC(),
	})
}

func defaultMenuProfile(businessID, locationID int64) domain.MenuProfile {
	return domain.MenuProfile{
		BusinessID:      businessID,
		LocationID:      locationID,
		DraftConfigJSON: domain.DefaultDraftConfigJSON(),
		DefaultLocale:   "en",
		EnabledLocales:  []string{"en", "fr", "ar"},
		DefaultCurrency: "USD",
	}
}

func normalizeLocale(locale string) string {
	loc := strings.ToLower(strings.TrimSpace(locale))
	if len(loc) > 16 {
		return ""
	}
	return loc
}

func normalizeLocales(locales []string, defaultLocale string) []string {
	seen := make(map[string]struct{}, len(locales)+1)
	out := make([]string, 0, len(locales)+1)
	add := func(raw string) {
		loc := normalizeLocale(raw)
		if loc == "" {
			return
		}
		if _, exists := seen[loc]; exists {
			return
		}
		seen[loc] = struct{}{}
		out = append(out, loc)
	}
	if defaultLocale != "" {
		add(defaultLocale)
	}
	for _, loc := range locales {
		add(loc)
	}
	if len(out) == 0 {
		return []string{"en"}
	}
	return out
}

func translationKey(entityType, entityID, locale, field string) string {
	return strings.ToLower(entityType) + "|" + entityID + "|" + strings.ToLower(locale) + "|" + strings.ToLower(field)
}
