package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubmenuv1 "github.com/voxire/voxie/gen/go/hub/menu/v1"
)

// ── Branded Menu Profiles & Presets ───────────────────────────────────────────

func (a *MenuAPI) GetMenuProfile(ctx context.Context, req *hubmenuv1.GetMenuProfileRequest) (*hubmenuv1.GetMenuProfileResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	profile, err := a.Service.GetMenuProfile(ctx, businessID, req.GetLocationId())
	if err != nil {
		return nil, builderError(err, "Failed to load menu profile")
	}
	return &hubmenuv1.GetMenuProfileResponse{Profile: toProtoProfile(profile)}, nil
}

func (a *MenuAPI) SaveMenuProfileDraft(ctx context.Context, req *hubmenuv1.SaveMenuProfileDraftRequest) (*hubmenuv1.SaveMenuProfileDraftResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	profile, err := a.Service.SaveMenuProfileDraft(ctx, businessID, domain.SaveMenuProfileDraftInput{
		LocationID:          req.GetLocationId(),
		DraftConfigJSON:     req.GetDraftConfigJson(),
		CustomCSS:           req.CustomCss,
		WhiteLabelHideBadge: req.WhiteLabelHideBadge,
		DefaultLocale:       req.DefaultLocale,
		EnabledLocales:      req.GetEnabledLocales(),
		DefaultCurrency:     req.DefaultCurrency,
	})
	if err != nil {
		return nil, builderError(err, "Failed to save menu profile draft")
	}
	return &hubmenuv1.SaveMenuProfileDraftResponse{Profile: toProtoProfile(profile)}, nil
}

func (a *MenuAPI) PublishMenuProfile(ctx context.Context, req *hubmenuv1.PublishMenuProfileRequest) (*hubmenuv1.PublishMenuProfileResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	profile, err := a.Service.PublishMenuProfile(ctx, businessID, req.GetLocationId())
	if err != nil {
		return nil, builderError(err, "Failed to publish menu profile")
	}
	return &hubmenuv1.PublishMenuProfileResponse{Profile: toProtoProfile(profile)}, nil
}

func (a *MenuAPI) CloneMenuProfile(ctx context.Context, req *hubmenuv1.CloneMenuProfileRequest) (*hubmenuv1.CloneMenuProfileResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	profile, err := a.Service.CloneMenuProfile(ctx, businessID, req.GetSourceLocationId(), req.GetTargetLocationId())
	if err != nil {
		return nil, builderError(err, "Failed to clone menu profile")
	}
	return &hubmenuv1.CloneMenuProfileResponse{Profile: toProtoProfile(profile)}, nil
}

func (a *MenuAPI) ListMenuPresets(ctx context.Context, _ *hubmenuv1.ListMenuPresetsRequest) (*hubmenuv1.ListMenuPresetsResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	presets, err := a.Service.ListMenuPresets(ctx, businessID)
	if err != nil {
		return nil, builderError(err, "Failed to list menu presets")
	}
	out := make([]*hubmenuv1.MenuPreset, 0, len(presets))
	for _, p := range presets {
		out = append(out, toProtoPreset(p))
	}
	return &hubmenuv1.ListMenuPresetsResponse{Presets: out}, nil
}

func (a *MenuAPI) CreateMenuPreset(ctx context.Context, req *hubmenuv1.CreateMenuPresetRequest) (*hubmenuv1.CreateMenuPresetResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	preset, err := a.Service.CreateMenuPreset(ctx, businessID, req.GetName(), req.GetConfigJson())
	if err != nil {
		return nil, builderError(err, "Failed to create menu preset")
	}
	return &hubmenuv1.CreateMenuPresetResponse{Preset: toProtoPreset(preset)}, nil
}

func (a *MenuAPI) UpdateMenuPreset(ctx context.Context, req *hubmenuv1.UpdateMenuPresetRequest) (*hubmenuv1.UpdateMenuPresetResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	preset, err := a.Service.UpdateMenuPreset(ctx, businessID, req.GetId(), req.GetName(), req.GetConfigJson())
	if err != nil {
		return nil, builderError(err, "Failed to update menu preset")
	}
	return &hubmenuv1.UpdateMenuPresetResponse{Preset: toProtoPreset(preset)}, nil
}

func (a *MenuAPI) DeleteMenuPreset(ctx context.Context, req *hubmenuv1.DeleteMenuPresetRequest) (*hubmenuv1.DeleteMenuPresetResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteMenuPreset(ctx, businessID, req.GetId()); err != nil {
		return nil, builderError(err, "Failed to delete menu preset")
	}
	return &hubmenuv1.DeleteMenuPresetResponse{Success: true}, nil
}

// ── Per-Branch POS Catalog Overrides ──────────────────────────────────────────

func (a *MenuAPI) ListBranchMenuItems(ctx context.Context, req *hubmenuv1.ListBranchMenuItemsRequest) (*hubmenuv1.ListBranchMenuItemsResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:view")
	if err != nil {
		return nil, err
	}
	items, err := a.Service.ListBranchMenuItems(ctx, businessID, req.GetLocationId())
	if err != nil {
		return nil, builderError(err, "Failed to list branch menu items")
	}
	out := make([]*hubmenuv1.BranchMenuItemConfig, 0, len(items))
	for _, it := range items {
		out = append(out, toProtoBranchItem(it))
	}
	return &hubmenuv1.ListBranchMenuItemsResponse{Items: out}, nil
}

func (a *MenuAPI) UpsertBranchMenuItem(ctx context.Context, req *hubmenuv1.UpsertBranchMenuItemRequest) (*hubmenuv1.UpsertBranchMenuItemResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = req.GetIsVisible()
	}
	item, err := a.Service.UpsertBranchMenuItem(ctx, businessID, domain.BranchMenuItem{
		LocationID:       req.GetLocationId(),
		MenuItemID:       req.GetMenuItemId(),
		IsVisible:        isVisible,
		HasPriceOverride: req.GetHasPriceOverride(),
		PriceOverride:    req.GetPriceOverride(),
		DisplayOrder:     req.GetDisplayOrder(),
		FeaturedBadge:    req.GetFeaturedBadge(),
		DietaryTags:      req.GetDietaryTags(),
		Allergens:        req.GetAllergens(),
		Calories:         req.GetCalories(),
		MediaURLs:        req.GetMediaUrls(),
		UpsellItemIDs:    req.GetUpsellItemIds(),
	})
	if err != nil {
		return nil, builderError(err, "Failed to save branch menu item")
	}
	return &hubmenuv1.UpsertBranchMenuItemResponse{Item: toProtoBranchItem(item)}, nil
}

// ── Multi-Language Translations ───────────────────────────────────────────────

func (a *MenuAPI) ListMenuTranslations(ctx context.Context, req *hubmenuv1.ListMenuTranslationsRequest) (*hubmenuv1.ListMenuTranslationsResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:view")
	if err != nil {
		return nil, err
	}
	list, err := a.Service.ListMenuTranslations(ctx, businessID, req.GetLocationId(), req.GetLocale())
	if err != nil {
		return nil, builderError(err, "Failed to list menu translations")
	}
	out := make([]*hubmenuv1.MenuTranslation, 0, len(list))
	for _, tr := range list {
		out = append(out, toProtoTranslation(tr))
	}
	return &hubmenuv1.ListMenuTranslationsResponse{Translations: out}, nil
}

func (a *MenuAPI) UpsertMenuTranslation(ctx context.Context, req *hubmenuv1.UpsertMenuTranslationRequest) (*hubmenuv1.UpsertMenuTranslationResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	tr, err := a.Service.UpsertMenuTranslation(ctx, businessID, domain.MenuTranslation{
		LocationID: req.GetLocationId(),
		EntityType: req.GetEntityType(),
		EntityID:   req.GetEntityId(),
		Locale:     req.GetLocale(),
		Field:      req.GetField(),
		Value:      req.GetValue(),
		Source:     req.GetSource(),
	})
	if err != nil {
		return nil, builderError(err, "Failed to save menu translation")
	}
	return &hubmenuv1.UpsertMenuTranslationResponse{Translation: toProtoTranslation(tr)}, nil
}

func (a *MenuAPI) TranslateMenu(ctx context.Context, req *hubmenuv1.TranslateMenuRequest) (*hubmenuv1.TranslateMenuResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	entries := make([]domain.TranslationSourceEntry, 0, len(req.GetEntries()))
	for _, entry := range req.GetEntries() {
		entries = append(entries, domain.TranslationSourceEntry{
			EntityType: entry.GetEntityType(),
			EntityID:   entry.GetEntityId(),
			Field:      entry.GetField(),
			SourceText: entry.GetSourceText(),
		})
	}
	translated, skipped, list, err := a.Service.TranslateMenu(
		ctx,
		businessID,
		req.GetLocationId(),
		req.GetTargetLocales(),
		req.GetOverwriteManual(),
		entries,
	)
	if err != nil {
		return nil, builderError(err, "Failed to translate menu")
	}
	out := make([]*hubmenuv1.MenuTranslation, 0, len(list))
	for _, tr := range list {
		out = append(out, toProtoTranslation(tr))
	}
	return &hubmenuv1.TranslateMenuResponse{
		TranslatedCount:    translated,
		SkippedManualCount: skipped,
		Translations:       out,
	}, nil
}

// ── Currencies ────────────────────────────────────────────────────────────────

func (a *MenuAPI) ListMenuCurrencies(ctx context.Context, req *hubmenuv1.ListMenuCurrenciesRequest) (*hubmenuv1.ListMenuCurrenciesResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	list, err := a.Service.ListMenuCurrencies(ctx, businessID, req.GetLocationId())
	if err != nil {
		return nil, builderError(err, "Failed to list menu currencies")
	}
	return &hubmenuv1.ListMenuCurrenciesResponse{Currencies: toProtoCurrencies(list)}, nil
}

func (a *MenuAPI) UpsertMenuCurrency(ctx context.Context, req *hubmenuv1.UpsertMenuCurrencyRequest) (*hubmenuv1.UpsertMenuCurrencyResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = req.GetIsActive()
	}
	decPlaces := int32(2)
	if req.DecimalPlaces != nil {
		decPlaces = req.GetDecimalPlaces()
	}
	list, err := a.Service.UpsertMenuCurrency(ctx, businessID, domain.MenuCurrency{
		LocationID:    req.GetLocationId(),
		Code:          req.GetCode(),
		Symbol:        req.GetSymbol(),
		ExchangeRate:  req.GetExchangeRate(),
		DecimalPlaces: decPlaces,
		IsDefault:     req.GetIsDefault(),
		IsActive:      isActive,
		Position:      req.GetPosition(),
	})
	if err != nil {
		return nil, builderError(err, "Failed to save menu currency")
	}
	return &hubmenuv1.UpsertMenuCurrencyResponse{Currencies: toProtoCurrencies(list)}, nil
}

func (a *MenuAPI) DeleteMenuCurrency(ctx context.Context, req *hubmenuv1.DeleteMenuCurrencyRequest) (*hubmenuv1.DeleteMenuCurrencyResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:branding:edit")
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteMenuCurrency(ctx, businessID, req.GetLocationId(), req.GetCode()); err != nil {
		return nil, builderError(err, "Failed to delete menu currency")
	}
	return &hubmenuv1.DeleteMenuCurrencyResponse{Success: true}, nil
}

// ── Digital Stamp Cards ───────────────────────────────────────────────────────

func (a *MenuAPI) ListStampCards(ctx context.Context, req *hubmenuv1.ListStampCardsRequest) (*hubmenuv1.ListStampCardsResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:view")
	if err != nil {
		return nil, err
	}
	cards, err := a.Service.ListStampCards(ctx, businessID, req.GetLocationId())
	if err != nil {
		return nil, builderError(err, "Failed to list stamp cards")
	}
	return &hubmenuv1.ListStampCardsResponse{Cards: toProtoStampCards(cards)}, nil
}

func (a *MenuAPI) UpsertStampCard(ctx context.Context, req *hubmenuv1.UpsertStampCardRequest) (*hubmenuv1.UpsertStampCardResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = req.GetIsActive()
	}
	card, err := a.Service.UpsertStampCard(ctx, businessID, domain.StampCard{
		ID:                  req.GetId(),
		LocationID:          req.GetLocationId(),
		Title:               req.GetTitle(),
		Description:         req.GetDescription(),
		StampsRequired:      req.GetStampsRequired(),
		RewardType:          req.GetRewardType(),
		RewardLabel:         req.GetRewardLabel(),
		RewardItemID:        req.GetRewardItemId(),
		RewardValue:         req.GetRewardValue(),
		QualifyingMinAmount: req.GetQualifyingMinAmount(),
		IsActive:            isActive,
	})
	if err != nil {
		return nil, builderError(err, "Failed to save stamp card")
	}
	return &hubmenuv1.UpsertStampCardResponse{Card: toProtoStampCard(card)}, nil
}

func (a *MenuAPI) DeleteStampCard(ctx context.Context, req *hubmenuv1.DeleteStampCardRequest) (*hubmenuv1.DeleteStampCardResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteStampCard(ctx, businessID, req.GetId()); err != nil {
		return nil, builderError(err, "Failed to delete stamp card")
	}
	return &hubmenuv1.DeleteStampCardResponse{Success: true}, nil
}

// ── Table Actions & Analytics (Authenticated) ─────────────────────────────────

func (a *MenuAPI) ListTableActions(ctx context.Context, req *hubmenuv1.ListTableActionsRequest) (*hubmenuv1.ListTableActionsResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:orders:view")
	if err != nil {
		return nil, err
	}
	actions, err := a.Service.ListTableActions(ctx, businessID, req.GetLocationId(), req.GetStatus())
	if err != nil {
		return nil, builderError(err, "Failed to list table actions")
	}
	out := make([]*hubmenuv1.MenuTableAction, 0, len(actions))
	for _, act := range actions {
		out = append(out, toProtoTableAction(act))
	}
	return &hubmenuv1.ListTableActionsResponse{Actions: out}, nil
}

func (a *MenuAPI) UpdateTableActionStatus(ctx context.Context, req *hubmenuv1.UpdateTableActionStatusRequest) (*hubmenuv1.UpdateTableActionStatusResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	if err := a.Service.UpdateTableActionStatus(ctx, businessID, req.GetId(), req.GetStatus()); err != nil {
		return nil, builderError(err, "Failed to update table action")
	}
	return &hubmenuv1.UpdateTableActionStatusResponse{Success: true}, nil
}

func (a *MenuAPI) GetMenuAnalyticsSummary(ctx context.Context, req *hubmenuv1.GetMenuAnalyticsSummaryRequest) (*hubmenuv1.GetMenuAnalyticsSummaryResponse, error) {
	businessID, err := requireMenuPermission(ctx, "menu:analytics:view")
	if err != nil {
		return nil, err
	}
	summary, err := a.Service.GetMenuAnalyticsSummary(ctx, businessID, req.GetLocationId(), req.GetDays())
	if err != nil {
		return nil, builderError(err, "Failed to summarize menu analytics")
	}
	counts := make([]*hubmenuv1.MenuAnalyticsCount, 0, len(summary.Counts))
	for _, c := range summary.Counts {
		counts = append(counts, &hubmenuv1.MenuAnalyticsCount{EventType: c.EventType, Count: c.Count})
	}
	return &hubmenuv1.GetMenuAnalyticsSummaryResponse{
		Counts:         counts,
		QrScans:        summary.QRScans,
		MenuViews:      summary.MenuViews,
		ItemViews:      summary.ItemViews,
		AddToCarts:     summary.AddToCarts,
		CheckoutStarts: summary.CheckoutStarts,
		OrdersPlaced:   summary.OrdersPlaced,
	}, nil
}

// ── Public Anonymous Endpoints ────────────────────────────────────────────────

func (a *MenuAPI) ResolvePublicMenuHost(ctx context.Context, req *hubmenuv1.ResolvePublicMenuHostRequest) (*hubmenuv1.ResolvePublicMenuHostResponse, error) {
	res, err := a.Service.ResolvePublicMenuHost(ctx, req.GetHost())
	if err != nil {
		return nil, builderError(err, "Failed to resolve host")
	}
	return &hubmenuv1.ResolvePublicMenuHostResponse{
		Found:        res.Found,
		BusinessId:   res.BusinessID,
		BusinessSlug: res.BusinessSlug,
		Domain:       res.Domain,
	}, nil
}

func (a *MenuAPI) GetPublicMenuBundle(ctx context.Context, req *hubmenuv1.GetPublicMenuBundleRequest) (*hubmenuv1.GetPublicMenuBundleResponse, error) {
	bundle, err := a.Service.GetPublicMenuBundle(ctx, req.GetBusinessId(), req.GetLocationId(), req.GetLocale())
	if err != nil {
		return nil, builderError(err, "Failed to load public menu bundle")
	}
	banners := make([]*hubmenuv1.MenuBanner, 0, len(bundle.Banners))
	for _, b := range bundle.Banners {
		banners = append(banners, toProtoBanner(b))
	}
	branchItems := make([]*hubmenuv1.BranchMenuItemConfig, 0, len(bundle.BranchItems))
	for _, bi := range bundle.BranchItems {
		branchItems = append(branchItems, toProtoBranchItem(bi))
	}
	translations := make([]*hubmenuv1.MenuTranslation, 0, len(bundle.Translations))
	for _, tr := range bundle.Translations {
		translations = append(translations, toProtoTranslation(tr))
	}
	// Always strip draft fields from public response so unpublished edits/CSS never leak.
	publicProfile := bundle.Profile
	publicProfile.DraftConfigJSON = ""
	publicProfile.CustomCSS = ""
	if !publicProfile.IsPublished {
		publicProfile.PublishedConfigJSON = ""
		publicProfile.PublishedCustomCSS = ""
	}
	return &hubmenuv1.GetPublicMenuBundleResponse{
		Profile:      toProtoProfile(publicProfile),
		Banners:      banners,
		BranchItems:  branchItems,
		Currencies:   toProtoCurrencies(bundle.Currencies),
		Translations: translations,
		StampCards:   toProtoStampCards(bundle.StampCards),
	}, nil
}

func (a *MenuAPI) CreatePublicTableAction(ctx context.Context, req *hubmenuv1.CreatePublicTableActionRequest) (*hubmenuv1.CreatePublicTableActionResponse, error) {
	action, err := a.Service.CreatePublicTableAction(ctx, req.GetBusinessId(), domain.TableAction{
		LocationID: req.GetLocationId(),
		TableID:    req.GetTableId(),
		TableLabel: req.GetTableLabel(),
		ActionType: req.GetActionType(),
		Note:       req.GetNote(),
	})
	if err != nil {
		return nil, builderError(err, "Failed to submit table request")
	}
	return &hubmenuv1.CreatePublicTableActionResponse{Action: toProtoTableAction(action)}, nil
}

func (a *MenuAPI) TrackPublicMenuEvents(ctx context.Context, req *hubmenuv1.TrackPublicMenuEventsRequest) (*hubmenuv1.TrackPublicMenuEventsResponse, error) {
	events := make([]domain.MenuEventInput, 0, len(req.GetEvents()))
	for _, ev := range req.GetEvents() {
		events = append(events, domain.MenuEventInput{
			LocationID: ev.GetLocationId(),
			TableID:    ev.GetTableId(),
			SessionID:  ev.GetSessionId(),
			EventType:  ev.GetEventType(),
			ItemID:     ev.GetItemId(),
			Locale:     ev.GetLocale(),
			Currency:   ev.GetCurrency(),
		})
	}
	accepted, err := a.Service.TrackPublicMenuEvents(ctx, req.GetBusinessId(), events)
	if err != nil {
		return nil, builderError(err, "Failed to record menu events")
	}
	return &hubmenuv1.TrackPublicMenuEventsResponse{Accepted: accepted}, nil
}

func (a *MenuAPI) GetPublicCustomerStamps(ctx context.Context, req *hubmenuv1.GetPublicCustomerStampsRequest) (*hubmenuv1.GetPublicCustomerStampsResponse, error) {
	progress, err := a.Service.GetPublicCustomerStamps(ctx, req.GetBusinessId(), req.GetPhone())
	if err != nil {
		return nil, builderError(err, "Failed to load stamp progress")
	}
	out := make([]*hubmenuv1.CustomerStampProgress, 0, len(progress))
	for _, p := range progress {
		out = append(out, &hubmenuv1.CustomerStampProgress{
			StampCardId:     p.StampCardID,
			CardTitle:       p.CardTitle,
			StampsCount:     p.StampsCount,
			StampsRequired:  p.StampsRequired,
			RewardsUnlocked: p.RewardsUnlocked,
			RewardsRedeemed: p.RewardsRedeemed,
			RewardLabel:     p.RewardLabel,
		})
	}
	return &hubmenuv1.GetPublicCustomerStampsResponse{Progress: out}, nil
}

// ── Converters & Error Mapping ────────────────────────────────────────────────

func builderError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrMenuPresetNotFound), errors.Is(err, domain.ErrStampCardNotFound):
		return onekiterrors.New(http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrMenuProfileInvalidInput),
		errors.Is(err, domain.ErrMenuPresetInvalidInput),
		errors.Is(err, domain.ErrBranchItemInvalidInput),
		errors.Is(err, domain.ErrTranslationInvalidInput),
		errors.Is(err, domain.ErrCurrencyInvalidInput),
		errors.Is(err, domain.ErrStampCardInvalidInput),
		errors.Is(err, domain.ErrTableActionInvalidInput),
		errors.Is(err, domain.ErrEventInvalidInput),
		errors.Is(err, domain.ErrUnsafeCustomCSS):
		return onekiterrors.New(http.StatusBadRequest, err.Error())
	default:
		return onekiterrors.New(http.StatusInternalServerError, fallback)
	}
}

func formatTimeUTC(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func toProtoProfile(p domain.MenuProfile) *hubmenuv1.MenuProfile {
	return &hubmenuv1.MenuProfile{
		Id:                   p.ID,
		BusinessId:           p.BusinessID,
		LocationId:           p.LocationID,
		DraftConfigJson:      p.DraftConfigJSON,
		PublishedConfigJson:  p.PublishedConfigJSON,
		CustomCss:            p.CustomCSS,
		PublishedCustomCss:   p.PublishedCustomCSS,
		WhiteLabelHideBadge:  p.WhiteLabelHideBadge,
		DefaultLocale:        p.DefaultLocale,
		EnabledLocales:       p.EnabledLocales,
		DefaultCurrency:      p.DefaultCurrency,
		IsPublished:          p.IsPublished,
		PublishedAt:          formatTimeUTC(p.PublishedAt),
		ClonedFromLocationId: p.ClonedFromLocationID,
		CreatedAt:            formatTimeUTC(p.CreatedAt),
		UpdatedAt:            formatTimeUTC(p.UpdatedAt),
	}
}

func toProtoPreset(p domain.MenuPreset) *hubmenuv1.MenuPreset {
	return &hubmenuv1.MenuPreset{
		Id:         p.ID,
		Name:       p.Name,
		ConfigJson: p.ConfigJSON,
		IsSystem:   p.IsSystem,
		CreatedAt:  formatTimeUTC(p.CreatedAt),
		UpdatedAt:  formatTimeUTC(p.UpdatedAt),
	}
}

func toProtoBranchItem(it domain.BranchMenuItem) *hubmenuv1.BranchMenuItemConfig {
	return &hubmenuv1.BranchMenuItemConfig{
		Id:               it.ID,
		LocationId:       it.LocationID,
		MenuItemId:       it.MenuItemID,
		IsVisible:        it.IsVisible,
		HasPriceOverride: it.HasPriceOverride,
		PriceOverride:    it.PriceOverride,
		DisplayOrder:     it.DisplayOrder,
		FeaturedBadge:    it.FeaturedBadge,
		DietaryTags:      it.DietaryTags,
		Allergens:        it.Allergens,
		Calories:         it.Calories,
		MediaUrls:        it.MediaURLs,
		UpsellItemIds:    it.UpsellItemIDs,
		UpdatedAt:        formatTimeUTC(it.UpdatedAt),
	}
}

func toProtoTranslation(tr domain.MenuTranslation) *hubmenuv1.MenuTranslation {
	return &hubmenuv1.MenuTranslation{
		Id:         tr.ID,
		LocationId: tr.LocationID,
		EntityType: tr.EntityType,
		EntityId:   tr.EntityID,
		Locale:     tr.Locale,
		Field:      tr.Field,
		Value:      tr.Value,
		Source:     tr.Source,
		UpdatedAt:  formatTimeUTC(tr.UpdatedAt),
	}
}

func toProtoCurrencies(list []domain.MenuCurrency) []*hubmenuv1.MenuCurrency {
	out := make([]*hubmenuv1.MenuCurrency, 0, len(list))
	for _, c := range list {
		out = append(out, &hubmenuv1.MenuCurrency{
			Id:            c.ID,
			LocationId:    c.LocationID,
			Code:          c.Code,
			Symbol:        c.Symbol,
			ExchangeRate:  c.ExchangeRate,
			DecimalPlaces: c.DecimalPlaces,
			IsDefault:     c.IsDefault,
			IsActive:      c.IsActive,
			Position:      c.Position,
		})
	}
	return out
}

func toProtoStampCards(list []domain.StampCard) []*hubmenuv1.MenuStampCard {
	out := make([]*hubmenuv1.MenuStampCard, 0, len(list))
	for _, c := range list {
		out = append(out, toProtoStampCard(c))
	}
	return out
}

func toProtoStampCard(c domain.StampCard) *hubmenuv1.MenuStampCard {
	return &hubmenuv1.MenuStampCard{
		Id:                  c.ID,
		LocationId:          c.LocationID,
		Title:               c.Title,
		Description:         c.Description,
		StampsRequired:      c.StampsRequired,
		RewardType:          c.RewardType,
		RewardLabel:         c.RewardLabel,
		RewardItemId:        c.RewardItemID,
		RewardValue:         c.RewardValue,
		QualifyingMinAmount: c.QualifyingMinAmount,
		IsActive:            c.IsActive,
		CreatedAt:           formatTimeUTC(c.CreatedAt),
		UpdatedAt:           formatTimeUTC(c.UpdatedAt),
	}
}

func toProtoTableAction(act domain.TableAction) *hubmenuv1.MenuTableAction {
	return &hubmenuv1.MenuTableAction{
		Id:         act.ID,
		LocationId: act.LocationID,
		TableId:    act.TableID,
		TableLabel: act.TableLabel,
		ActionType: act.ActionType,
		Status:     act.Status,
		Note:       act.Note,
		CreatedAt:  formatTimeUTC(act.CreatedAt),
		UpdatedAt:  formatTimeUTC(act.UpdatedAt),
	}
}
