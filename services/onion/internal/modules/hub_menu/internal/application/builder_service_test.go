package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hub_menu/internal/domain"
	"github.com/voxire/onion/internal/platform/eventbus"
)

type fakeBuilderRepo struct {
	repository
	profiles     map[int64]domain.MenuProfile
	translations []domain.MenuTranslation
	currencies   []domain.MenuCurrency
	stampCards   []domain.StampCard
	tableActions []domain.TableAction
	events       []domain.MenuEventInput
	clonedSource int64
	clonedTarget int64
	stampsAdded  int
}

func newFakeBuilderRepo() *fakeBuilderRepo {
	return &fakeBuilderRepo{profiles: make(map[int64]domain.MenuProfile)}
}

func (f *fakeBuilderRepo) GetMenuProfile(_ context.Context, _ int64, locationID int64) (domain.MenuProfile, bool, error) {
	p, ok := f.profiles[locationID]
	return p, ok, nil
}

func (f *fakeBuilderRepo) UpsertMenuProfileDraft(_ context.Context, profile domain.MenuProfile) error {
	f.profiles[profile.LocationID] = profile
	return nil
}

func (f *fakeBuilderRepo) PublishMenuProfile(_ context.Context, _ int64, locationID int64) error {
	p := f.profiles[locationID]
	p.PublishedConfigJSON = p.DraftConfigJSON
	p.PublishedCustomCSS = p.CustomCSS
	p.IsPublished = true
	p.PublishedAt = time.Now().UTC()
	f.profiles[locationID] = p
	return nil
}

func (f *fakeBuilderRepo) CloneMenuProfile(_ context.Context, businessID, sourceLocationID, targetLocationID int64, source domain.MenuProfile) error {
	f.clonedSource = sourceLocationID
	f.clonedTarget = targetLocationID
	cloned := source
	cloned.BusinessID = businessID
	cloned.LocationID = targetLocationID
	cloned.ClonedFromLocationID = sourceLocationID
	f.profiles[targetLocationID] = cloned
	return nil
}

func (f *fakeBuilderRepo) ListMenuTranslations(_ context.Context, _ int64, _ int64, locale string) ([]domain.MenuTranslation, error) {
	if locale == "" {
		return append([]domain.MenuTranslation{}, f.translations...), nil
	}
	var out []domain.MenuTranslation
	for _, tr := range f.translations {
		if tr.Locale == locale {
			out = append(out, tr)
		}
	}
	return out, nil
}

func (f *fakeBuilderRepo) UpsertMenuTranslation(_ context.Context, tr domain.MenuTranslation) error {
	for i, existing := range f.translations {
		if existing.EntityType == tr.EntityType && existing.EntityID == tr.EntityID && existing.Locale == tr.Locale && existing.Field == tr.Field {
			f.translations[i] = tr
			return nil
		}
	}
	f.translations = append(f.translations, tr)
	return nil
}

func (f *fakeBuilderRepo) ListMenuCurrencies(_ context.Context, _ int64, locationID int64) ([]domain.MenuCurrency, error) {
	var out []domain.MenuCurrency
	for _, cur := range f.currencies {
		if cur.LocationID == locationID {
			out = append(out, cur)
		}
	}
	return out, nil
}

func (f *fakeBuilderRepo) UpsertMenuCurrency(_ context.Context, cur domain.MenuCurrency) error {
	if cur.IsDefault {
		for i := range f.currencies {
			if f.currencies[i].LocationID == cur.LocationID {
				f.currencies[i].IsDefault = false
			}
		}
	}
	for i, existing := range f.currencies {
		if existing.LocationID == cur.LocationID && existing.Code == cur.Code {
			f.currencies[i] = cur
			return nil
		}
	}
	f.currencies = append(f.currencies, cur)
	return nil
}

func (f *fakeBuilderRepo) ListStampCards(_ context.Context, _ int64, _ int64) ([]domain.StampCard, error) {
	return append([]domain.StampCard{}, f.stampCards...), nil
}

func (f *fakeBuilderRepo) AddCustomerStamp(_ context.Context, _, _ int64, _, _ string, _ int32, _ int64) error {
	f.stampsAdded++
	return nil
}

func (f *fakeBuilderRepo) CreateTableAction(_ context.Context, action domain.TableAction) (int64, error) {
	action.ID = int64(len(f.tableActions) + 1)
	f.tableActions = append(f.tableActions, action)
	return action.ID, nil
}

func (f *fakeBuilderRepo) InsertMenuEvents(_ context.Context, _ int64, events []domain.MenuEventInput) error {
	f.events = append(f.events, events...)
	return nil
}

func (f *fakeBuilderRepo) ListActivePublicBanners(_ context.Context, _ int64) ([]domain.Banner, error) {
	return nil, nil
}

func (f *fakeBuilderRepo) ListBranchMenuItems(_ context.Context, _ int64, _ int64) ([]domain.BranchMenuItem, error) {
	return nil, nil
}

func (f *fakeBuilderRepo) GetCustomerStampProgress(_ context.Context, _ int64, _ string) ([]domain.CustomerStampProgress, error) {
	return []domain.CustomerStampProgress{{StampCardID: 7, StampsCount: 3, StampsRequired: 5}}, nil
}

func TestSavePublishAndCloneMenuProfile(t *testing.T) {
	repo := newFakeBuilderRepo()
	var emitted []eventbus.Event
	svc := NewService(context.Background(), repo).WithEventDispatcher(func(_ context.Context, ev eventbus.Event) error {
		emitted = append(emitted, ev)
		return nil
	})

	css := ".rtylr-menu-root { --menu-accent: #d97706; }"
	saved, err := svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      10,
		DraftConfigJSON: `{"items":{"layout":"list"}}`,
		CustomCSS:       &css,
		EnabledLocales:  []string{"en", "ar", "fr"},
	})
	if err != nil {
		t.Fatalf("SaveMenuProfileDraft failed: %v", err)
	}
	if saved.CustomCSS != css || saved.LocationID != 10 {
		t.Fatalf("unexpected saved profile: %+v", saved)
	}

	published, err := svc.PublishMenuProfile(context.Background(), 42, 10)
	if err != nil {
		t.Fatalf("PublishMenuProfile failed: %v", err)
	}
	if !published.IsPublished || published.PublishedConfigJSON != `{"items":{"layout":"list"}}` {
		t.Fatalf("expected published config, got %+v", published)
	}
	if len(emitted) != 1 || emitted[0].Type != "menu.profile.published" {
		t.Fatalf("expected menu.profile.published event, got %+v", emitted)
	}

	cloned, err := svc.CloneMenuProfile(context.Background(), 42, 10, 20)
	if err != nil {
		t.Fatalf("CloneMenuProfile failed: %v", err)
	}
	if cloned.LocationID != 20 || cloned.ClonedFromLocationID != 10 {
		t.Fatalf("unexpected cloned profile: %+v", cloned)
	}
}

func TestTranslateMenuPreservesManualOverrides(t *testing.T) {
	repo := newFakeBuilderRepo()
	repo.translations = []domain.MenuTranslation{
		{
			BusinessID: 42,
			LocationID: 1,
			EntityType: "item",
			EntityID:   "101",
			Locale:     "ar",
			Field:      "name",
			Value:      "برغر خاص",
			Source:     "manual",
		},
	}

	svc := NewService(context.Background(), repo).WithTranslator(func(_ context.Context, locale string, inputs []string) ([]string, error) {
		out := make([]string, len(inputs))
		for i, text := range inputs {
			out[i] = locale + ":" + text
		}
		return out, nil
	})

	translated, skipped, _, err := svc.TranslateMenu(context.Background(), 42, 1, []string{"ar"}, false, []domain.TranslationSourceEntry{
		{EntityType: "item", EntityID: "101", Field: "name", SourceText: "Classic Burger"},
		{EntityType: "item", EntityID: "102", Field: "name", SourceText: "Iced Latte"},
	})
	if err != nil {
		t.Fatalf("TranslateMenu failed: %v", err)
	}
	if translated != 1 || skipped != 1 {
		t.Fatalf("expected 1 translated & 1 skipped manual, got %d / %d", translated, skipped)
	}
}

func TestMenuCurrenciesAndTableActionsAndStamps(t *testing.T) {
	repo := newFakeBuilderRepo()
	repo.stampCards = []domain.StampCard{{ID: 7, BusinessID: 42, Title: "Coffee Club", StampsRequired: 5, QualifyingMinAmount: 10, IsActive: true}}
	var emitted []eventbus.Event
	svc := NewService(context.Background(), repo).WithEventDispatcher(func(_ context.Context, ev eventbus.Event) error {
		emitted = append(emitted, ev)
		return nil
	})

	// Default USD when none stored.
	list, err := svc.ListMenuCurrencies(context.Background(), 42, 1)
	if err != nil || len(list) != 1 || list[0].Code != "USD" {
		t.Fatalf("expected default USD currency, got %+v (%v)", list, err)
	}

	// Add dynamic secondary currency: must preserve default USD alongside LBP.
	list, err = svc.UpsertMenuCurrency(context.Background(), 42, domain.MenuCurrency{
		LocationID: 1, Code: "lbp", Symbol: "L.L.", ExchangeRate: 89500, DecimalPlaces: 0, IsActive: true,
	})
	if err != nil || len(list) != 2 || list[0].Code != "USD" || !list[0].IsDefault || list[1].Code != "LBP" {
		t.Fatalf("expected default USD preserved alongside LBP, got %+v (%v)", list, err)
	}

	// Invalid table action rejected.
	if _, err := svc.CreatePublicTableAction(context.Background(), 42, domain.TableAction{ActionType: "unknown"}); !errors.Is(err, domain.ErrTableActionInvalidInput) {
		t.Fatalf("expected ErrTableActionInvalidInput, got %v", err)
	}
	// Valid table action emits webhook event.
	act, err := svc.CreatePublicTableAction(context.Background(), 42, domain.TableAction{
		LocationID: 1, TableID: 4, TableLabel: "Table 4", ActionType: "call_waiter",
	})
	if err != nil || act.ID == 0 {
		t.Fatalf("expected table action created, got %+v (%v)", act, err)
	}

	// Stamp card award on qualifying order total.
	if err := svc.AwardStampsForOrder(context.Background(), 42, 1, 999, "+15551234", "Abed", 25.00); err != nil {
		t.Fatalf("AwardStampsForOrder failed: %v", err)
	}
	if repo.stampsAdded != 1 {
		t.Fatalf("expected 1 stamp added, got %d", repo.stampsAdded)
	}
	if len(emitted) != 2 {
		t.Fatalf("expected 2 emitted webhook events (table action + stamp), got %d", len(emitted))
	}
}

func TestGetPublicMenuBundleNeverLeaksUnpublishedDraft(t *testing.T) {
	repo := newFakeBuilderRepo()
	svc := NewService(context.Background(), repo)

	css := ".rtylr-menu-root { color: red; }"
	if _, err := svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      0,
		DraftConfigJSON: `{"payment":{"mode":"redirect","redirectUrl":"https://evil.example/pay"}}`,
		CustomCSS:       &css,
	}); err != nil {
		t.Fatalf("SaveMenuProfileDraft failed: %v", err)
	}

	bundle, err := svc.GetPublicMenuBundle(context.Background(), 42, 0, "en")
	if err != nil {
		t.Fatalf("GetPublicMenuBundle failed: %v", err)
	}
	if bundle.Profile.DraftConfigJSON != "" || bundle.Profile.CustomCSS != "" || bundle.Profile.PublishedConfigJSON != "" || bundle.Profile.PublishedCustomCSS != "" {
		t.Fatalf("unpublished draft leaked into public bundle: %+v", bundle.Profile)
	}
	if got := svc.ResolvePaymentRedirectURL(context.Background(), 42, 0); got != "" {
		t.Fatalf("ResolvePaymentRedirectURL must ignore unpublished drafts, got %q", got)
	}

	// Publish, then edit draft again: public bundle must still expose only published values.
	if _, err := svc.PublishMenuProfile(context.Background(), 42, 0); err != nil {
		t.Fatalf("PublishMenuProfile failed: %v", err)
	}
	newDraftCSS := ".rtylr-menu-root { color: blue; }"
	if _, err := svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      0,
		DraftConfigJSON: `{"payment":{"mode":"redirect","redirectUrl":"https://draft.example/pay"}}`,
		CustomCSS:       &newDraftCSS,
	}); err != nil {
		t.Fatalf("second SaveMenuProfileDraft failed: %v", err)
	}
	bundle, err = svc.GetPublicMenuBundle(context.Background(), 42, 0, "en")
	if err != nil {
		t.Fatalf("GetPublicMenuBundle after publish failed: %v", err)
	}
	if bundle.Profile.DraftConfigJSON != "" || bundle.Profile.CustomCSS != "" {
		t.Fatalf("draft fields must be stripped even after publish: %+v", bundle.Profile)
	}
	if bundle.Profile.PublishedCustomCSS != css {
		t.Fatalf("expected published CSS %q, got %q", css, bundle.Profile.PublishedCustomCSS)
	}
	if got := svc.ResolvePaymentRedirectURL(context.Background(), 42, 0); got != "https://evil.example/pay" {
		t.Fatalf("expected published redirect URL, got %q", got)
	}
}

func TestGetPublicCustomerStampsGating(t *testing.T) {
	repo := newFakeBuilderRepo()
	svc := NewService(context.Background(), repo)

	// Too-short phone returns empty without probing repo.
	res, err := svc.GetPublicCustomerStamps(context.Background(), 42, "123")
	if err != nil || len(res) != 0 {
		t.Fatalf("expected empty for short phone, got %+v (%v)", res, err)
	}

	// Published profile with enableLoyaltyStamps disabled hides progress.
	_, _ = svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      0,
		DraftConfigJSON: `{"features":{"enableLoyaltyStamps":false}}`,
	})
	_, _ = svc.PublishMenuProfile(context.Background(), 42, 0)
	res, err = svc.GetPublicCustomerStamps(context.Background(), 42, "+15551234567")
	if err != nil || len(res) != 0 {
		t.Fatalf("expected empty when loyaltyStamps disabled, got %+v (%v)", res, err)
	}

	// Enabled enableLoyaltyStamps allows lookup.
	_, _ = svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      0,
		DraftConfigJSON: `{"features":{"enableLoyaltyStamps":true}}`,
	})
	_, _ = svc.PublishMenuProfile(context.Background(), 42, 0)
	res, err = svc.GetPublicCustomerStamps(context.Background(), 42, "+15551234567")
	if err != nil || len(res) != 1 {
		t.Fatalf("expected stamp progress when enabled, got %+v (%v)", res, err)
	}
}

func TestTenantValidatorRejectsForeignIDs(t *testing.T) {
	repo := newFakeBuilderRepo()
	svc := NewService(context.Background(), repo).WithTenantValidator(TenantValidator{
		ValidateLocation: func(_ context.Context, _, locationID int64) error {
			if locationID == 999 {
				return errors.New("foreign location")
			}
			return nil
		},
		ValidateTable: func(_ context.Context, _, tableID int64) error {
			if tableID == 888 {
				return errors.New("foreign table")
			}
			return nil
		},
		ValidateMenuItem: func(_ context.Context, _, menuItemID int64) error {
			if menuItemID == 777 {
				return errors.New("foreign item")
			}
			return nil
		},
	})

	if _, err := svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      999,
		DraftConfigJSON: `{}`,
	}); !errors.Is(err, domain.ErrMenuProfileInvalidInput) {
		t.Fatalf("expected ErrMenuProfileInvalidInput for foreign location, got %v", err)
	}
	if _, err := svc.CreatePublicTableAction(context.Background(), 42, domain.TableAction{
		TableID: 888, ActionType: "call_waiter",
	}); !errors.Is(err, domain.ErrTableActionInvalidInput) {
		t.Fatalf("expected ErrTableActionInvalidInput for foreign table, got %v", err)
	}
	if _, err := svc.UpsertBranchMenuItem(context.Background(), 42, domain.BranchMenuItem{
		LocationID: 1, MenuItemID: 777, IsVisible: true,
	}); !errors.Is(err, domain.ErrBranchItemInvalidInput) {
		t.Fatalf("expected ErrBranchItemInvalidInput for foreign item, got %v", err)
	}
}

func TestBranchInheritsDefaultBrandCurrencies(t *testing.T) {
	repo := newFakeBuilderRepo()
	svc := NewService(context.Background(), repo)

	// Add EUR on default brand (location_id = 0): must seed USD alongside EUR.
	defaultList, err := svc.UpsertMenuCurrency(context.Background(), 42, domain.MenuCurrency{
		LocationID: 0, Code: "EUR", Symbol: "€", ExchangeRate: 0.92, DecimalPlaces: 2, IsActive: true,
	})
	if err != nil || len(defaultList) != 2 {
		t.Fatalf("expected USD + EUR on default brand, got %+v (%v)", defaultList, err)
	}

	// Branch 4 has no currency rows of its own -> must inherit default brand USD + EUR.
	inherited, err := svc.ListMenuCurrencies(context.Background(), 42, 4)
	if err != nil || len(inherited) != 2 || inherited[0].Code != "USD" || inherited[1].Code != "EUR" {
		t.Fatalf("expected branch 4 to inherit USD + EUR, got %+v (%v)", inherited, err)
	}

	// Public bundle for branch 4 must also surface inherited currencies.
	_, _ = svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{
		LocationID:      4,
		DraftConfigJSON: `{}`,
	})
	_, _ = svc.PublishMenuProfile(context.Background(), 42, 4)
	bundle, err := svc.GetPublicMenuBundle(context.Background(), 42, 4, "en")
	if err != nil || len(bundle.Currencies) != 2 {
		t.Fatalf("expected public bundle for branch 4 to inherit 2 currencies, got %+v (%v)", bundle.Currencies, err)
	}
}

func TestCapturePublicMenuGuestAndFeatureKillSwitches(t *testing.T) {
	repo := newFakeBuilderRepo()
	repo.stampCards = []domain.StampCard{{ID: 7, BusinessID: 42, Title: "Club", StampsRequired: 5, IsActive: true}}
	flags := map[string]bool{}
	var captured []string
	svc := NewService(context.Background(), repo).
		WithFeatureFlagChecker(func(_ context.Context, key string, def bool) bool {
			if v, ok := flags[key]; ok {
				return v
			}
			return def
		}).
		WithGuestCapturer(func(_ context.Context, _, _ int64, name, phone, email, _ string) error {
			captured = append(captured, name+"|"+phone+"|"+email)
			return nil
		})

	// Rejects without consent.
	if err := svc.CapturePublicMenuGuest(context.Background(), 42, 0, "Sara", "+15550001111", "", "", false); !errors.Is(err, domain.ErrGuestCaptureInvalidInput) {
		t.Fatalf("expected ErrGuestCaptureInvalidInput without consent, got %v", err)
	}
	// Rejects without phone or email.
	if err := svc.CapturePublicMenuGuest(context.Background(), 42, 0, "Sara", "", "", "", true); !errors.Is(err, domain.ErrGuestCaptureInvalidInput) {
		t.Fatalf("expected ErrGuestCaptureInvalidInput without contact, got %v", err)
	}
	// Rejects malformed email.
	if err := svc.CapturePublicMenuGuest(context.Background(), 42, 0, "Sara", "", "not-an-email", "", true); !errors.Is(err, domain.ErrGuestCaptureInvalidInput) {
		t.Fatalf("expected ErrGuestCaptureInvalidInput for invalid email, got %v", err)
	}
	// Captures consenting guest cleanly.
	if err := svc.CapturePublicMenuGuest(context.Background(), 42, 0, "Sara", "+15550001111", "sara@example.com", "VIP", true); err != nil {
		t.Fatalf("expected guest capture success, got %v", err)
	}
	if len(captured) != 1 || captured[0] != "Sara|+15550001111|sara@example.com" {
		t.Fatalf("unexpected captured payload: %+v", captured)
	}

	// Super-admin kill-switches disable CRM capture, table actions, AI translate, ordering, loyalty, CSS, currencies.
	flags["app.menu.crmCapture"] = false
	flags["app.menu.tableActions"] = false
	flags["app.menu.aiTranslate"] = false
	flags["app.menu.ordering"] = false
	flags["app.menu.loyalty"] = false
	flags["app.menu.customCss"] = false
	flags["app.menu.currencies"] = false

	if err := svc.CapturePublicMenuGuest(context.Background(), 42, 0, "Sara", "+15550001111", "", "", true); !errors.Is(err, domain.ErrFeatureDisabled) {
		t.Fatalf("expected ErrFeatureDisabled for crmCapture kill-switch, got %v", err)
	}
	if _, err := svc.CreatePublicTableAction(context.Background(), 42, domain.TableAction{ActionType: "call_waiter"}); !errors.Is(err, domain.ErrFeatureDisabled) {
		t.Fatalf("expected ErrFeatureDisabled for tableActions kill-switch, got %v", err)
	}
	if _, _, _, err := svc.TranslateMenu(context.Background(), 42, 0, []string{"ar"}, false, nil); !errors.Is(err, domain.ErrFeatureDisabled) {
		t.Fatalf("expected ErrFeatureDisabled for aiTranslate kill-switch, got %v", err)
	}
	if svc.IsPublishedOrderingEnabled(context.Background(), 42) {
		t.Fatal("expected IsPublishedOrderingEnabled to be false when app.menu.ordering kill-switch is off")
	}

	css := ".rtylr-menu-root { color: green; }"
	_, _ = svc.SaveMenuProfileDraft(context.Background(), 42, domain.SaveMenuProfileDraftInput{LocationID: 0, DraftConfigJSON: `{}`, CustomCSS: &css})
	_, _ = svc.PublishMenuProfile(context.Background(), 42, 0)
	_, _ = svc.UpsertMenuCurrency(context.Background(), 42, domain.MenuCurrency{LocationID: 0, Code: "EUR", Symbol: "€", ExchangeRate: 0.92, IsActive: true})

	bundle, err := svc.GetPublicMenuBundle(context.Background(), 42, 0, "en")
	if err != nil {
		t.Fatalf("GetPublicMenuBundle failed: %v", err)
	}
	if bundle.Profile.PublishedCustomCSS != "" {
		t.Fatalf("expected custom CSS stripped when kill-switch is off, got %q", bundle.Profile.PublishedCustomCSS)
	}
	if len(bundle.StampCards) != 0 {
		t.Fatalf("expected 0 stamp cards when loyalty kill-switch is off, got %d", len(bundle.StampCards))
	}
	if len(bundle.Currencies) != 1 || bundle.Currencies[0].Code != "USD" {
		t.Fatalf("expected only default currency when currencies kill-switch is off, got %+v", bundle.Currencies)
	}
}
