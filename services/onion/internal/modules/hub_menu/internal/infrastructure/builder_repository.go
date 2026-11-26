package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/hub_menu/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) GetMenuProfile(ctx context.Context, businessID, locationID int64) (domain.MenuProfile, bool, error) {
	row, err := r.queries.GetMenuProfile(ctx, sqlc.GetMenuProfileParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.MenuProfile{}, false, nil
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to load hub menu profile", logger.Error(err))
		return domain.MenuProfile{}, false, err
	}
	return menuProfileFromRow(row), true, nil
}

func (r *Repository) UpsertMenuProfileDraft(ctx context.Context, profile domain.MenuProfile) error {
	enabledLocalesRaw, err := json.Marshal(nonNilStrings(profile.EnabledLocales))
	if err != nil {
		return err
	}
	clonedFrom := sql.NullInt64{}
	if profile.ClonedFromLocationID > 0 {
		clonedFrom = sql.NullInt64{Int64: profile.ClonedFromLocationID, Valid: true}
	}
	return r.queries.UpsertMenuProfileDraft(ctx, sqlc.UpsertMenuProfileDraftParams{
		BusinessID:           profile.BusinessID,
		LocationID:           profile.LocationID,
		DraftConfig:          json.RawMessage(profile.DraftConfigJSON),
		CustomCss:            sqlutil.NullString(profile.CustomCSS),
		WhiteLabelHideBadge:  profile.WhiteLabelHideBadge,
		DefaultLocale:        profile.DefaultLocale,
		EnabledLocales:       enabledLocalesRaw,
		DefaultCurrency:      profile.DefaultCurrency,
		ClonedFromLocationID: clonedFrom,
	})
}

func (r *Repository) PublishMenuProfile(ctx context.Context, businessID, locationID int64) error {
	return r.queries.PublishMenuProfile(ctx, sqlc.PublishMenuProfileParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
}

func (r *Repository) CloneMenuProfile(ctx context.Context, businessID, sourceLocationID, targetLocationID int64, sourceProfile domain.MenuProfile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	enabledLocalesRaw, err := json.Marshal(nonNilStrings(sourceProfile.EnabledLocales))
	if err != nil {
		return err
	}
	if err := qtx.UpsertMenuProfileDraft(ctx, sqlc.UpsertMenuProfileDraftParams{
		BusinessID:           businessID,
		LocationID:           targetLocationID,
		DraftConfig:          json.RawMessage(sourceProfile.DraftConfigJSON),
		CustomCss:            sqlutil.NullString(sourceProfile.CustomCSS),
		WhiteLabelHideBadge:  sourceProfile.WhiteLabelHideBadge,
		DefaultLocale:        sourceProfile.DefaultLocale,
		EnabledLocales:       enabledLocalesRaw,
		DefaultCurrency:      sourceProfile.DefaultCurrency,
		ClonedFromLocationID: sql.NullInt64{Int64: sourceLocationID, Valid: true},
	}); err != nil {
		return err
	}

	// Copy branch item overrides.
	sourceItems, err := qtx.ListBranchMenuItems(ctx, sqlc.ListBranchMenuItemsParams{
		BusinessID: businessID,
		LocationID: sourceLocationID,
	})
	if err != nil {
		return err
	}
	if err := qtx.DeleteBranchMenuItemsForLocation(ctx, sqlc.DeleteBranchMenuItemsForLocationParams{
		BusinessID: businessID,
		LocationID: targetLocationID,
	}); err != nil {
		return err
	}
	for _, item := range sourceItems {
		if err := qtx.UpsertBranchMenuItem(ctx, sqlc.UpsertBranchMenuItemParams{
			BusinessID:    businessID,
			LocationID:    targetLocationID,
			MenuItemID:    item.MenuItemID,
			IsVisible:     item.IsVisible,
			PriceOverride: item.PriceOverride,
			DisplayOrder:  item.DisplayOrder,
			FeaturedBadge: sqlutil.NullString(item.FeaturedBadge),
			DietaryTags:   item.DietaryTags,
			Allergens:     item.Allergens,
			Calories:      item.Calories,
			MediaUrls:     item.MediaUrls,
			UpsellItemIds: item.UpsellItemIds,
		}); err != nil {
			return err
		}
	}

	// Copy currencies.
	sourceCurrencies, err := qtx.ListMenuCurrencies(ctx, sqlc.ListMenuCurrenciesParams{
		BusinessID: businessID,
		LocationID: sourceLocationID,
	})
	if err != nil {
		return err
	}
	if err := qtx.DeleteMenuCurrenciesForLocation(ctx, sqlc.DeleteMenuCurrenciesForLocationParams{
		BusinessID: businessID,
		LocationID: targetLocationID,
	}); err != nil {
		return err
	}
	for _, cur := range sourceCurrencies {
		if err := qtx.UpsertMenuCurrency(ctx, sqlc.UpsertMenuCurrencyParams{
			BusinessID:    businessID,
			LocationID:    targetLocationID,
			Code:          cur.Code,
			Symbol:        cur.Symbol,
			ExchangeRate:  cur.ExchangeRate,
			DecimalPlaces: cur.DecimalPlaces,
			IsDefault:     cur.IsDefault,
			IsActive:      cur.IsActive,
			Position:      cur.Position,
		}); err != nil {
			return err
		}
	}

	// Copy translations.
	sourceTranslations, err := qtx.ListMenuTranslations(ctx, sqlc.ListMenuTranslationsParams{
		BusinessID: businessID,
		LocationID: sourceLocationID,
		Column3:    "",
		Locale:     "",
	})
	if err != nil {
		return err
	}
	if err := qtx.DeleteMenuTranslationsForLocation(ctx, sqlc.DeleteMenuTranslationsForLocationParams{
		BusinessID: businessID,
		LocationID: targetLocationID,
	}); err != nil {
		return err
	}
	for _, tr := range sourceTranslations {
		if err := qtx.UpsertMenuTranslation(ctx, sqlc.UpsertMenuTranslationParams{
			BusinessID: businessID,
			LocationID: targetLocationID,
			EntityType: tr.EntityType,
			EntityID:   tr.EntityID,
			Locale:     tr.Locale,
			Field:      tr.Field,
			Value:      tr.Value,
			Source:     tr.Source,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) ListMenuPresets(ctx context.Context, businessID int64) ([]domain.MenuPreset, error) {
	rows, err := r.queries.ListMenuPresets(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MenuPreset, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.MenuPreset{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			ConfigJSON: string(row.Config),
			IsSystem:   row.IsSystem,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *Repository) GetMenuPreset(ctx context.Context, businessID, id int64) (domain.MenuPreset, error) {
	row, err := r.queries.GetMenuPreset(ctx, sqlc.GetMenuPresetParams{BusinessID: businessID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.MenuPreset{}, domain.ErrMenuPresetNotFound
	}
	if err != nil {
		return domain.MenuPreset{}, err
	}
	return domain.MenuPreset{
		ID:         row.ID,
		BusinessID: row.BusinessID,
		Name:       row.Name,
		ConfigJSON: string(row.Config),
		IsSystem:   row.IsSystem,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *Repository) CreateMenuPreset(ctx context.Context, businessID int64, name, configJSON string) (int64, error) {
	res, err := r.queries.CreateMenuPreset(ctx, sqlc.CreateMenuPresetParams{
		BusinessID: businessID,
		Name:       name,
		Config:     json.RawMessage(configJSON),
		IsSystem:   false,
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateMenuPreset(ctx context.Context, businessID, id int64, name, configJSON string) error {
	return r.queries.UpdateMenuPreset(ctx, sqlc.UpdateMenuPresetParams{
		Name:       name,
		Config:     json.RawMessage(configJSON),
		BusinessID: businessID,
		ID:         id,
	})
}

func (r *Repository) DeleteMenuPreset(ctx context.Context, businessID, id int64) error {
	return r.queries.DeleteMenuPreset(ctx, sqlc.DeleteMenuPresetParams{BusinessID: businessID, ID: id})
}

func (r *Repository) ListBranchMenuItems(ctx context.Context, businessID, locationID int64) ([]domain.BranchMenuItem, error) {
	rows, err := r.queries.ListBranchMenuItems(ctx, sqlc.ListBranchMenuItemsParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.BranchMenuItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, branchItemFromRow(row))
	}
	return out, nil
}

func (r *Repository) UpsertBranchMenuItem(ctx context.Context, item domain.BranchMenuItem) error {
	dietaryRaw, err := json.Marshal(nonNilStrings(item.DietaryTags))
	if err != nil {
		return err
	}
	allergensRaw, err := json.Marshal(nonNilStrings(item.Allergens))
	if err != nil {
		return err
	}
	mediaRaw, err := json.Marshal(nonNilStrings(item.MediaURLs))
	if err != nil {
		return err
	}
	upsellsRaw, err := json.Marshal(nonNilInt64s(item.UpsellItemIDs))
	if err != nil {
		return err
	}
	priceOverride := sql.NullString{}
	if item.HasPriceOverride {
		priceOverride = sql.NullString{String: fmt.Sprintf("%.2f", item.PriceOverride), Valid: true}
	}
	calories := sql.NullInt32{}
	if item.Calories > 0 {
		calories = sql.NullInt32{Int32: item.Calories, Valid: true}
	}
	return r.queries.UpsertBranchMenuItem(ctx, sqlc.UpsertBranchMenuItemParams{
		BusinessID:    item.BusinessID,
		LocationID:    item.LocationID,
		MenuItemID:    item.MenuItemID,
		IsVisible:     item.IsVisible,
		PriceOverride: priceOverride,
		DisplayOrder:  item.DisplayOrder,
		FeaturedBadge: sqlutil.NullString(item.FeaturedBadge),
		DietaryTags:   dietaryRaw,
		Allergens:     allergensRaw,
		Calories:      calories,
		MediaUrls:     mediaRaw,
		UpsellItemIds: upsellsRaw,
	})
}

func (r *Repository) ListMenuTranslations(ctx context.Context, businessID, locationID int64, locale string) ([]domain.MenuTranslation, error) {
	rows, err := r.queries.ListMenuTranslations(ctx, sqlc.ListMenuTranslationsParams{
		BusinessID: businessID,
		LocationID: locationID,
		Column3:    locale,
		Locale:     locale,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.MenuTranslation, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.MenuTranslation{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			LocationID: row.LocationID,
			EntityType: row.EntityType,
			EntityID:   row.EntityID,
			Locale:     row.Locale,
			Field:      row.Field,
			Value:      row.Value,
			Source:     string(row.Source),
			UpdatedAt:  row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *Repository) UpsertMenuTranslation(ctx context.Context, tr domain.MenuTranslation) error {
	src := sqlc.HubMenuTranslationsSourceManual
	if tr.Source == "ai" {
		src = sqlc.HubMenuTranslationsSourceAi
	}
	return r.queries.UpsertMenuTranslation(ctx, sqlc.UpsertMenuTranslationParams{
		BusinessID: tr.BusinessID,
		LocationID: tr.LocationID,
		EntityType: tr.EntityType,
		EntityID:   tr.EntityID,
		Locale:     tr.Locale,
		Field:      tr.Field,
		Value:      tr.Value,
		Source:     src,
	})
}

func (r *Repository) ListMenuCurrencies(ctx context.Context, businessID, locationID int64) ([]domain.MenuCurrency, error) {
	rows, err := r.queries.ListMenuCurrencies(ctx, sqlc.ListMenuCurrenciesParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.MenuCurrency, 0, len(rows))
	for _, row := range rows {
		rate, _ := strconv.ParseFloat(row.ExchangeRate, 64)
		if rate <= 0 {
			rate = 1
		}
		out = append(out, domain.MenuCurrency{
			ID:            row.ID,
			BusinessID:    row.BusinessID,
			LocationID:    row.LocationID,
			Code:          row.Code,
			Symbol:        row.Symbol,
			ExchangeRate:  rate,
			DecimalPlaces: row.DecimalPlaces,
			IsDefault:     row.IsDefault,
			IsActive:      row.IsActive,
			Position:      row.Position,
		})
	}
	return out, nil
}

func (r *Repository) UpsertMenuCurrency(ctx context.Context, cur domain.MenuCurrency) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	if cur.IsDefault {
		if err := qtx.ClearDefaultMenuCurrency(ctx, sqlc.ClearDefaultMenuCurrencyParams{
			BusinessID: cur.BusinessID,
			LocationID: cur.LocationID,
		}); err != nil {
			return err
		}
	}
	if err := qtx.UpsertMenuCurrency(ctx, sqlc.UpsertMenuCurrencyParams{
		BusinessID:    cur.BusinessID,
		LocationID:    cur.LocationID,
		Code:          cur.Code,
		Symbol:        cur.Symbol,
		ExchangeRate:  fmt.Sprintf("%.6f", cur.ExchangeRate),
		DecimalPlaces: cur.DecimalPlaces,
		IsDefault:     cur.IsDefault,
		IsActive:      cur.IsActive,
		Position:      cur.Position,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteMenuCurrency(ctx context.Context, businessID, locationID int64, code string) error {
	return r.queries.DeleteMenuCurrency(ctx, sqlc.DeleteMenuCurrencyParams{
		BusinessID: businessID,
		LocationID: locationID,
		Code:       code,
	})
}

func (r *Repository) ListStampCards(ctx context.Context, businessID, locationID int64) ([]domain.StampCard, error) {
	rows, err := r.queries.ListStampCards(ctx, sqlc.ListStampCardsParams{
		BusinessID: businessID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.StampCard, 0, len(rows))
	for _, row := range rows {
		out = append(out, stampCardFromListRow(row))
	}
	return out, nil
}

func (r *Repository) GetStampCard(ctx context.Context, businessID, id int64) (domain.StampCard, error) {
	row, err := r.queries.GetStampCard(ctx, sqlc.GetStampCardParams{BusinessID: businessID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StampCard{}, domain.ErrStampCardNotFound
	}
	if err != nil {
		return domain.StampCard{}, err
	}
	return stampCardFromGetRow(row), nil
}

func (r *Repository) CreateStampCard(ctx context.Context, card domain.StampCard) (int64, error) {
	rewardItemID := sql.NullInt64{}
	if card.RewardItemID > 0 {
		rewardItemID = sql.NullInt64{Int64: card.RewardItemID, Valid: true}
	}
	res, err := r.queries.CreateStampCard(ctx, sqlc.CreateStampCardParams{
		BusinessID:          card.BusinessID,
		LocationID:          card.LocationID,
		Title:               card.Title,
		Description:         sqlutil.NullString(card.Description),
		StampsRequired:      card.StampsRequired,
		RewardType:          toStampRewardType(card.RewardType),
		RewardLabel:         card.RewardLabel,
		RewardItemID:        rewardItemID,
		RewardValue:         fmt.Sprintf("%.2f", card.RewardValue),
		QualifyingMinAmount: fmt.Sprintf("%.2f", card.QualifyingMinAmount),
		IsActive:            card.IsActive,
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateStampCard(ctx context.Context, card domain.StampCard) error {
	rewardItemID := sql.NullInt64{}
	if card.RewardItemID > 0 {
		rewardItemID = sql.NullInt64{Int64: card.RewardItemID, Valid: true}
	}
	return r.queries.UpdateStampCard(ctx, sqlc.UpdateStampCardParams{
		LocationID:          card.LocationID,
		Title:               card.Title,
		Description:         sqlutil.NullString(card.Description),
		StampsRequired:      card.StampsRequired,
		RewardType:          toStampRewardType(card.RewardType),
		RewardLabel:         card.RewardLabel,
		RewardItemID:        rewardItemID,
		RewardValue:         fmt.Sprintf("%.2f", card.RewardValue),
		QualifyingMinAmount: fmt.Sprintf("%.2f", card.QualifyingMinAmount),
		IsActive:            card.IsActive,
		BusinessID:          card.BusinessID,
		ID:                  card.ID,
	})
}

func (r *Repository) DeleteStampCard(ctx context.Context, businessID, id int64) error {
	return r.queries.DeleteStampCard(ctx, sqlc.DeleteStampCardParams{BusinessID: businessID, ID: id})
}

func (r *Repository) GetCustomerStampProgress(ctx context.Context, businessID int64, phone string) ([]domain.CustomerStampProgress, error) {
	rows, err := r.queries.GetCustomerStampProgress(ctx, sqlc.GetCustomerStampProgressParams{
		BusinessID:    businessID,
		CustomerPhone: phone,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CustomerStampProgress, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.CustomerStampProgress{
			StampCardID:     row.StampCardID,
			CardTitle:       row.CardTitle,
			StampsCount:     row.StampsCount,
			StampsRequired:  row.StampsRequired,
			RewardsUnlocked: row.RewardsUnlocked,
			RewardsRedeemed: row.RewardsRedeemed,
			RewardLabel:     row.RewardLabel,
		})
	}
	return out, nil
}

func (r *Repository) AddCustomerStamp(ctx context.Context, businessID, stampCardID int64, phone, name string, stampsRequired int32, orderID int64) error {
	lastOrder := sql.NullInt64{}
	if orderID > 0 {
		lastOrder = sql.NullInt64{Int64: orderID, Valid: true}
	}
	return r.queries.AddCustomerStamp(ctx, sqlc.AddCustomerStampParams{
		BusinessID:    businessID,
		StampCardID:   stampCardID,
		CustomerPhone: phone,
		CustomerName:  sqlutil.NullString(name),
		Column5:       stampsRequired,
		LastOrderID:   lastOrder,
		StampsCount:   stampsRequired,
	})
}

func (r *Repository) CreateTableAction(ctx context.Context, action domain.TableAction) (int64, error) {
	res, err := r.queries.CreateTableAction(ctx, sqlc.CreateTableActionParams{
		BusinessID: action.BusinessID,
		LocationID: action.LocationID,
		TableID:    action.TableID,
		TableLabel: action.TableLabel,
		ActionType: toTableActionType(action.ActionType),
		Note:       sqlutil.NullString(action.Note),
	})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) ListTableActions(ctx context.Context, businessID, locationID int64, status string) ([]domain.TableAction, error) {
	rows, err := r.queries.ListTableActions(ctx, sqlc.ListTableActionsParams{
		BusinessID: businessID,
		Column2:    locationID,
		LocationID: locationID,
		Column4:    status,
		Status:     sqlc.HubMenuTableActionsStatus(status),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TableAction, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.TableAction{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			LocationID: row.LocationID,
			TableID:    row.TableID,
			TableLabel: row.TableLabel,
			ActionType: string(row.ActionType),
			Status:     string(row.Status),
			Note:       row.Note,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *Repository) UpdateTableActionStatus(ctx context.Context, businessID, id int64, status string) error {
	return r.queries.UpdateTableActionStatus(ctx, sqlc.UpdateTableActionStatusParams{
		Status:     sqlc.HubMenuTableActionsStatus(status),
		BusinessID: businessID,
		ID:         id,
	})
}

func (r *Repository) InsertMenuEvents(ctx context.Context, businessID int64, events []domain.MenuEventInput) error {
	for _, ev := range events {
		itemID := sql.NullInt64{}
		if ev.ItemID > 0 {
			itemID = sql.NullInt64{Int64: ev.ItemID, Valid: true}
		}
		if err := r.queries.InsertMenuEvent(ctx, sqlc.InsertMenuEventParams{
			BusinessID: businessID,
			LocationID: ev.LocationID,
			TableID:    ev.TableID,
			SessionID:  ev.SessionID,
			EventType:  sqlc.HubMenuEventsEventType(ev.EventType),
			ItemID:     itemID,
			Locale:     sqlutil.NullString(ev.Locale),
			Currency:   sqlutil.NullString(ev.Currency),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) SummarizeMenuEvents(ctx context.Context, businessID, locationID int64, since time.Time) (domain.AnalyticsSummary, error) {
	rows, err := r.queries.SummarizeMenuEvents(ctx, sqlc.SummarizeMenuEventsParams{
		BusinessID: businessID,
		Column2:    locationID,
		LocationID: locationID,
		CreatedAt:  since,
	})
	if err != nil {
		return domain.AnalyticsSummary{}, err
	}
	summary := domain.AnalyticsSummary{
		Counts: make([]domain.AnalyticsCount, 0, len(rows)),
	}
	for _, row := range rows {
		et := string(row.EventType)
		summary.Counts = append(summary.Counts, domain.AnalyticsCount{EventType: et, Count: row.TotalCount})
		switch et {
		case "qr_scan":
			summary.QRScans = row.TotalCount
		case "menu_view":
			summary.MenuViews = row.TotalCount
		case "item_view":
			summary.ItemViews = row.TotalCount
		case "add_to_cart":
			summary.AddToCarts = row.TotalCount
		case "checkout_start":
			summary.CheckoutStarts = row.TotalCount
		case "order_placed":
			summary.OrdersPlaced = row.TotalCount
		}
	}
	return summary, nil
}

func (r *Repository) ListActivePublicBanners(ctx context.Context, businessID int64) ([]domain.Banner, error) {
	rows, err := r.queries.ListActivePublicBanners(ctx, businessID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Banner, 0, len(rows))
	for _, row := range rows {
		out = append(out, bannerFromFields(domain.Banner{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			Title:      row.Title,
			Body:       row.Body,
			ImageURL:   row.ImageUrl,
			LinkURL:    row.LinkUrl,
			Position:   row.Position,
			IsActive:   row.IsActive,
		}, row.StartsAt, row.EndsAt, row.CreatedAt, row.UpdatedAt))
	}
	return out, nil
}

func menuProfileFromRow(row sqlc.GetMenuProfileRow) domain.MenuProfile {
	var locales []string
	_ = json.Unmarshal(row.EnabledLocales, &locales)
	if len(locales) == 0 {
		locales = []string{"en"}
	}
	var publishedConfig string
	if row.IsPublished && len(row.PublishedConfig) > 0 {
		publishedConfig = string(row.PublishedConfig)
	}
	var publishedAt time.Time
	if row.PublishedAt.Valid {
		publishedAt = row.PublishedAt.Time
	}
	var clonedFrom int64
	if row.ClonedFromLocationID.Valid {
		clonedFrom = row.ClonedFromLocationID.Int64
	}
	return domain.MenuProfile{
		ID:                   row.ID,
		BusinessID:           row.BusinessID,
		LocationID:           row.LocationID,
		DraftConfigJSON:      string(row.DraftConfig),
		PublishedConfigJSON:  publishedConfig,
		CustomCSS:            row.CustomCss,
		PublishedCustomCSS:   row.PublishedCustomCss,
		WhiteLabelHideBadge:  row.WhiteLabelHideBadge,
		DefaultLocale:        row.DefaultLocale,
		EnabledLocales:       locales,
		DefaultCurrency:      row.DefaultCurrency,
		IsPublished:          row.IsPublished,
		PublishedAt:          publishedAt,
		ClonedFromLocationID: clonedFrom,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

func branchItemFromRow(row sqlc.ListBranchMenuItemsRow) domain.BranchMenuItem {
	var dietary, allergens, media []string
	var upsells []int64
	_ = json.Unmarshal(row.DietaryTags, &dietary)
	_ = json.Unmarshal(row.Allergens, &allergens)
	_ = json.Unmarshal(row.MediaUrls, &media)
	_ = json.Unmarshal(row.UpsellItemIds, &upsells)
	var priceOverride float64
	if row.PriceOverride.Valid {
		priceOverride, _ = strconv.ParseFloat(row.PriceOverride.String, 64)
	}
	var calories int32
	if row.Calories.Valid {
		calories = row.Calories.Int32
	}
	return domain.BranchMenuItem{
		ID:               row.ID,
		BusinessID:       row.BusinessID,
		LocationID:       row.LocationID,
		MenuItemID:       row.MenuItemID,
		IsVisible:        row.IsVisible,
		HasPriceOverride: row.PriceOverride.Valid,
		PriceOverride:    priceOverride,
		DisplayOrder:     row.DisplayOrder,
		FeaturedBadge:    row.FeaturedBadge,
		DietaryTags:      nonNilStrings(dietary),
		Allergens:        nonNilStrings(allergens),
		Calories:         calories,
		MediaURLs:        nonNilStrings(media),
		UpsellItemIDs:    nonNilInt64s(upsells),
		UpdatedAt:        row.UpdatedAt,
	}
}

func stampCardFromListRow(row sqlc.ListStampCardsRow) domain.StampCard {
	rv, _ := strconv.ParseFloat(row.RewardValue, 64)
	qm, _ := strconv.ParseFloat(row.QualifyingMinAmount, 64)
	var rewardItem int64
	if row.RewardItemID.Valid {
		rewardItem = row.RewardItemID.Int64
	}
	return domain.StampCard{
		ID:                  row.ID,
		BusinessID:          row.BusinessID,
		LocationID:          row.LocationID,
		Title:               row.Title,
		Description:         row.Description,
		StampsRequired:      row.StampsRequired,
		RewardType:          string(row.RewardType),
		RewardLabel:         row.RewardLabel,
		RewardItemID:        rewardItem,
		RewardValue:         rv,
		QualifyingMinAmount: qm,
		IsActive:            row.IsActive,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func stampCardFromGetRow(row sqlc.GetStampCardRow) domain.StampCard {
	return stampCardFromListRow(sqlc.ListStampCardsRow(row))
}

func toStampRewardType(val string) sqlc.HubMenuStampCardsRewardType {
	switch val {
	case "percent_discount":
		return sqlc.HubMenuStampCardsRewardTypePercentDiscount
	case "fixed_discount":
		return sqlc.HubMenuStampCardsRewardTypeFixedDiscount
	default:
		return sqlc.HubMenuStampCardsRewardTypeFreeItem
	}
}

func toTableActionType(val string) sqlc.HubMenuTableActionsActionType {
	switch val {
	case "request_bill":
		return sqlc.HubMenuTableActionsActionTypeRequestBill
	case "split_bill":
		return sqlc.HubMenuTableActionsActionTypeSplitBill
	default:
		return sqlc.HubMenuTableActionsActionTypeCallWaiter
	}
}

func nonNilStrings(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

func nonNilInt64s(in []int64) []int64 {
	if in == nil {
		return []int64{}
	}
	return in
}
