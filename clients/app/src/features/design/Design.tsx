import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQueryClient } from '@tanstack/react-query';
import { HelmetContainer } from '@voxire/components';
import { Switch } from '@voxire/ui';
import { uploadAsset } from '@voxire/api';
import { FLAG_KEYS, cdnImageUrl, useFeatureFlag } from '@voxire/config';
import { CaretIcon, onCaretToolSuccess, useCaretStore } from '@voxire/caret';
import { toast } from 'sonner';
import {
  Check,
  Code2,
  Copy,
  Eye,
  FileText,
  LayoutGrid,
  Monitor,
  Palette,
  Plus,
  RotateCcw,
  SlidersHorizontal,
  Smartphone,
  Sparkles,
  Store,
  Tablet,
  Trash2,
  Upload,
} from 'lucide-react';
import {
  DEFAULT_FEATURES_CONFIG,
  DEFAULT_LAYOUT_CONFIG,
  MENU_ACCENT_SWATCHES,
  MENU_STARTER_THEMES,
  MenuLayout,
  applyMenuStylePrompt,
  type BusinessInfo,
  type CardRadius,
  type CoverHeight,
  type ImagePosition,
  type ImageSize,
  type ItemLayout,
  type MenuCategoryData,
  type MenuFontSize,
  type MenuLayoutConfig,
  type NavSize,
  type NavStyle,
  type ThemeMode,
} from '@voxire/menu';
import {
  Badge,
  Button,
  Input,
  Label,
  Select,
  Separator,
  Tabs,
  TabsList,
  TabsTrigger,
} from '../../shared/components/ui';
import { useLayoutPresetsStore } from '../../shared/stores/layoutPresetsStore';
import { useUserStore } from '../../shared/stores/userStore';
import { useMenuStore } from '../../shared/stores/menuStore';
import {
  useCloneMenuProfile,
  useCreateMenuPreset,
  useDeleteMenuCurrency,
  useDeleteMenuPreset,
  useMenuAnalyticsSummary,
  useMenuBranches,
  useMenuCatalogItems,
  useMenuCurrencies,
  useMenuPresets,
  useMenuProfile,
  useMenuReviews,
  usePublishMenuProfile,
  useSaveMenuProfileDraft,
  useStampCards,
  useTranslateMenu,
  useUpsertMenuCurrency,
  useUpsertStampCard,
} from '../../shared/hooks/queries/useMenuBuilder';
import {
  useCreateMenuBanner,
  useDeleteMenuBanner,
  useMenuBanners,
} from '../../shared/hooks/queries/useMenu';
import PagesEditor from './PagesEditor';

type RailTab = 'themes' | 'brand' | 'layout' | 'pages' | 'features' | 'ai' | 'css';
type DeviceViewport = 'mobile' | 'tablet' | 'desktop';

const DEFAULT_BUSINESS: Omit<BusinessInfo, 'name' | 'description'> = {
  coverImage: undefined,
  logo: undefined,
  address: '',
  phone: '',
  hours: '',
  averageRating: 0,
  reviewCount: 0,
};

const FONT_OPTIONS = [
  { id: 'system', label: 'System UI' },
  { id: 'inter', label: 'Inter' },
  { id: 'jakarta', label: 'Plus Jakarta' },
  { id: 'poppins', label: 'Poppins' },
  { id: 'playfair', label: 'Playfair' },
  { id: 'merriweather', label: 'Merriweather' },
  { id: 'cairo', label: 'Cairo (عربي)' },
  { id: 'tajawal', label: 'Tajawal (عربي)' },
] as const;

const AI_QUICK_PROMPTS = [
  'Luxury dark gold bistro',
  'Warm beige artisan cafe',
  'Arabic RTL Cairo editorial',
  'Self-service retail list',
] as const;

interface SwitchRowProps {
  id: string;
  label: string;
  hint?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}

const SwitchRow: React.FC<SwitchRowProps> = ({ id, label, hint, checked, onChange }) => (
  <div className="flex items-center justify-between gap-3 py-1.5">
    <div className="min-w-0">
      <label
        htmlFor={id}
        className="block text-[13px] font-medium cursor-pointer select-none"
        style={{ color: 'var(--dash-text)' }}
      >
        {label}
      </label>
      {hint && (
        <span className="block text-[11px]" style={{ color: 'var(--dash-text-muted)' }}>
          {hint}
        </span>
      )}
    </div>
    <Switch id={id} checked={checked} onCheckedChange={onChange} />
  </div>
);

interface SegmentedControlProps<T extends string> {
  label?: string;
  value: T;
  options: { label: string; value: T }[];
  onChange: (value: T) => void;
}

function SegmentedControl<T extends string>({
  label,
  value,
  options,
  onChange,
}: SegmentedControlProps<T>) {
  return (
    <div className="space-y-1.5">
      {label && <Label className="text-[11px] uppercase tracking-wider">{label}</Label>}
      <Tabs value={value} onValueChange={(v) => onChange(v as T)}>
        <TabsList
          className="w-full grid"
          style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }}
        >
          {options.map((opt) => (
            <TabsTrigger key={opt.value} value={opt.value} className="px-2 py-1.5 text-xs">
              {opt.label}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>
    </div>
  );
}

interface RawMenuItem {
  _id?: string;
  name?: string;
  description?: string;
  price?: number;
  currency?: string;
  available?: boolean;
  image?: string;
  badges?: string[];
}

interface RawCategory {
  _id?: string;
  name?: string;
  type?: string;
  menuItems?: RawMenuItem[];
}

const Design: React.FC = () => {
  const {
    presets,
    activePresetId,
    draftConfig,
    activatePreset,
    deletePreset,
    setDraftConfig,
    patchDraftConfig,
    saveActivePreset,
  } = useLayoutPresetsStore();

  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const { user } = useUserStore();
  const { items: menuItems } = useMenuStore();
  const openCaret = useCaretStore((s) => s.open);
  const businessSlug = user?.businessSlug || user?.business_slug || '';

  const platformOrderingEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_ORDERING);
  const platformLoyaltyEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_LOYALTY);
  const platformTableActionsEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_TABLE_ACTIONS);
  const platformCrmCaptureEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CRM_CAPTURE);
  const platformCurrenciesEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CURRENCIES);
  const platformAiTranslateEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_AI_TRANSLATE);
  const platformCustomCssEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CUSTOM_CSS);
  const platformBannersEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_BANNERS);

  const [activeTab, setActiveTab] = useState<RailTab>('themes');
  const [mobilePane, setMobilePane] = useState<'editor' | 'preview'>('editor');
  const [viewport, setViewport] = useState<DeviceViewport>('mobile');
  const [selectedLocationId, setSelectedLocationId] = useState<number>(0);
  const [cloneSourceId, setCloneSourceId] = useState<number>(0);
  const [previewLocale, setPreviewLocale] = useState<string>('en');
  const [customCss, setCustomCss] = useState<string>('');
  const [enabledLocales, setEnabledLocales] = useState<string[]>(['en', 'fr', 'ar']);
  const [newPresetName, setNewPresetName] = useState('');
  const [newCurrencyCode, setNewCurrencyCode] = useState('');
  const [newCurrencySymbol, setNewCurrencySymbol] = useState('');
  const [newCurrencyRate, setNewCurrencyRate] = useState('1');
  const [newStampTitle, setNewStampTitle] = useState('');
  const [newStampTarget, setNewStampTarget] = useState('10');
  const [newStampReward, setNewStampReward] = useState('');
  const [newBannerTitle, setNewBannerTitle] = useState('');
  const [newBannerBody, setNewBannerBody] = useState('');
  const [aiPrompt, setAiPrompt] = useState('');
  const [saved, setSaved] = useState(false);
  const [uploadingField, setUploadingField] = useState<'cover' | 'logo' | null>(null);
  const coverInputRef = useRef<HTMLInputElement>(null);
  const logoInputRef = useRef<HTMLInputElement>(null);

  const { data: branches = [] } = useMenuBranches(businessSlug);
  const { data: catalogItems = [] } = useMenuCatalogItems();
  const { data: remoteProfile } = useMenuProfile(selectedLocationId);
  const { data: remotePresets = [] } = useMenuPresets();
  const { data: currencies = [] } = useMenuCurrencies(selectedLocationId);
  const { data: stampCards = [] } = useStampCards(selectedLocationId);
  const { data: menuBanners = [] } = useMenuBanners();
  const { data: analyticsSummary } = useMenuAnalyticsSummary(selectedLocationId, 30);

  const saveDraftMutation = useSaveMenuProfileDraft();
  const publishMutation = usePublishMenuProfile();
  const cloneMutation = useCloneMenuProfile();
  const createPresetMutation = useCreateMenuPreset();
  const deletePresetMutation = useDeleteMenuPreset();
  const upsertCurrencyMutation = useUpsertMenuCurrency();
  const deleteCurrencyMutation = useDeleteMenuCurrency();
  const upsertStampCardMutation = useUpsertStampCard();
  const createBannerMutation = useCreateMenuBanner();
  const deleteBannerMutation = useDeleteMenuBanner();
  const translateMenuMutation = useTranslateMenu();

  useEffect(() => {
    return onCaretToolSuccess((event) => {
      if (
        event.name === 'pos.update_menu_design' ||
        event.name === 'pos.publish_menu_design' ||
        event.name === 'pos.clone_menu_design'
      ) {
        void queryClient.invalidateQueries({ queryKey: ['menu-builder'] });
      }
    });
  }, [queryClient]);

  useEffect(() => {
    if (!remoteProfile) return;
    try {
      if (remoteProfile.draftConfigJson) {
        const parsed = JSON.parse(remoteProfile.draftConfigJson) as Partial<MenuLayoutConfig>;
        setDraftConfig({
          ...DEFAULT_LAYOUT_CONFIG,
          ...parsed,
          header: { ...DEFAULT_LAYOUT_CONFIG.header, ...parsed.header },
          nav: { ...DEFAULT_LAYOUT_CONFIG.nav, ...parsed.nav },
          items: { ...DEFAULT_LAYOUT_CONFIG.items, ...parsed.items },
          pages: { menu: true, ...parsed.pages },
          typography: { ...DEFAULT_LAYOUT_CONFIG.typography!, ...parsed.typography },
          features: {
            ...DEFAULT_FEATURES_CONFIG,
            ...parsed.features,
            fulfillmentModes: {
              ...DEFAULT_FEATURES_CONFIG.fulfillmentModes,
              ...parsed.features?.fulfillmentModes,
            },
            tableActions: {
              ...DEFAULT_FEATURES_CONFIG.tableActions,
              ...parsed.features?.tableActions,
            },
          },
          payment: { ...DEFAULT_LAYOUT_CONFIG.payment!, ...parsed.payment },
        });
      }
    } catch {
      // Keep current local draft if stored JSON cannot be parsed.
    }
    setCustomCss(remoteProfile.customCss ?? '');
    if (remoteProfile.enabledLocales && remoteProfile.enabledLocales.length > 0) {
      setEnabledLocales(remoteProfile.enabledLocales);
    }
    if (remoteProfile.defaultLocale) {
      setPreviewLocale(remoteProfile.defaultLocale);
    }
  }, [remoteProfile, setDraftConfig]);

  const previewT = useMemo(() => i18n.getFixedT(previewLocale), [i18n, previewLocale]);
  const { data: studioReviews = [] } = useMenuReviews(selectedLocationId);
  const selectedBranch = branches.find((b) => b.id === selectedLocationId);
  const approvedPreviewReviews = useMemo(
    () =>
      studioReviews
        .filter((r) => r.status === 'approved')
        .map((r) => ({
          id: String(r.id ?? ''),
          author: r.authorName || 'Guest',
          rating: r.rating ?? 5,
          comment: r.comment ?? '',
          date: r.createdAt ? r.createdAt.slice(0, 10) : '',
        })),
    [studioReviews]
  );
  const pendingReviewCount = useMemo(
    () => studioReviews.filter((r) => r.status === 'pending').length,
    [studioReviews]
  );
  const u = user as Record<string, unknown> | null;
  const businessInfo: BusinessInfo = {
    name: (u?.businessName as string) || previewT('design.businessPlaceholder.name'),
    description:
      (u?.businessDescription as string) || previewT('design.businessPlaceholder.description'),
    coverImage: (u?.coverImage as string) || DEFAULT_BUSINESS.coverImage,
    logo: (u?.logo as string) || DEFAULT_BUSINESS.logo,
    address: selectedBranch?.address || (u?.address as string) || '',
    phone: (u?.phone as string) || '',
    hours: (u?.hours as string) || '',
    reviews: approvedPreviewReviews,
    averageRating:
      approvedPreviewReviews.length > 0
        ? approvedPreviewReviews.reduce((sum, r) => sum + r.rating, 0) /
          approvedPreviewReviews.length
        : 0,
    reviewCount: approvedPreviewReviews.length,
  };

  const menuCategories: MenuCategoryData[] = useMemo(() => {
    if (catalogItems.length > 0) {
      const grouped = new Map<string, MenuCategoryData['items']>();
      for (const item of catalogItems) {
        const id = String(item.id ?? '');
        const catName = item.category?.trim() || previewT('design.studio.title');
        const list = grouped.get(catName) ?? [];
        list.push({
          id,
          name: item.name ?? '',
          description: item.description ?? '',
          price: item.salePrice ?? 0,
          currency: currencies.find((c) => c.isDefault)?.code ?? 'USD',
          available: item.isActive !== false,
          image: item.imageUrl ?? '',
          badges: [],
        });
        grouped.set(catName, list);
      }
      return Array.from(grouped.entries()).map(([name, items], idx) => ({
        id: `cat-${idx}`,
        name,
        items,
      }));
    }
    if (menuItems && menuItems.length > 0) {
      return (menuItems as RawCategory[])
        .filter((item) => item.type === 'category')
        .map((cat) => ({
          id: cat._id ?? '',
          name: cat.name ?? '',
          items: (cat.menuItems || []).map((mi) => ({
            id: mi._id ?? '',
            name: mi.name ?? '',
            description: mi.description || '',
            price: mi.price ?? 0,
            currency: mi.currency || 'USD',
            available: mi.available !== false,
            image: mi.image || '',
            badges: mi.badges || [],
          })),
        })) as MenuCategoryData[];
    }
    return [];
  }, [catalogItems, currencies, menuItems, previewT]);

  const themeMode: ThemeMode = draftConfig.mode ?? 'light';
  const fontFamily = draftConfig.typography?.bodyFont ?? 'system';
  const fontSize: MenuFontSize = draftConfig.typography?.fontSize ?? 'md';
  const colorPrimary = draftConfig.colors?.primary ?? '#111827';
  const colorAccent = draftConfig.colors?.accent ?? '#4583C4';
  const colorBg = draftConfig.colors?.bg ?? '#f9f9f7';
  const colorSurface = draftConfig.colors?.surface ?? '#ffffff';
  const features = {
    ...DEFAULT_FEATURES_CONFIG,
    ...draftConfig.features,
    fulfillmentModes: {
      ...DEFAULT_FEATURES_CONFIG.fulfillmentModes,
      ...draftConfig.features?.fulfillmentModes,
    },
    tableActions: {
      ...DEFAULT_FEATURES_CONFIG.tableActions,
      ...draftConfig.features?.tableActions,
    },
  };
  const paymentMode = draftConfig.payment?.mode ?? 'pos';
  const paymentRedirectUrl = draftConfig.payment?.redirectUrl ?? '';

  const patchHeader = (patch: Partial<MenuLayoutConfig['header']>) =>
    patchDraftConfig({ header: { ...draftConfig.header, ...patch } });

  const handleMediaUpload = async (field: 'cover' | 'logo', file: File | undefined) => {
    if (!file) return;
    setUploadingField(field);
    try {
      const res = await uploadAsset(file, { visibility: 'public' });
      const uploadedUrl = res.url || '';
      if (!uploadedUrl) throw new Error('Missing uploaded URL');
      patchHeader(field === 'cover' ? { coverImageUrl: uploadedUrl } : { logoUrl: uploadedUrl });
      toast.success(t('design.studio.uploadSuccess'));
    } catch (err: unknown) {
      const msg = (err as { message?: string })?.message || t('design.studio.uploadFailed');
      toast.error(msg);
    } finally {
      setUploadingField(null);
    }
  };

  const patchNav = (patch: Partial<MenuLayoutConfig['nav']>) =>
    patchDraftConfig({ nav: { ...draftConfig.nav, ...patch } });

  const patchItems = (patch: Partial<MenuLayoutConfig['items']>) =>
    patchDraftConfig({ items: { ...draftConfig.items, ...patch } });

  const patchPages = (patch: Partial<MenuLayoutConfig['pages']>) =>
    patchDraftConfig({ pages: { ...draftConfig.pages, ...patch } });

  const patchContent = (patch: Partial<NonNullable<MenuLayoutConfig['content']>>) =>
    patchDraftConfig({ content: { ...draftConfig.content, ...patch } });

  const patchColors = (patch: Partial<NonNullable<MenuLayoutConfig['colors']>>) =>
    patchDraftConfig({
      colors: {
        primary: colorPrimary,
        accent: colorAccent,
        bg: colorBg,
        surface: colorSurface,
        ...draftConfig.colors,
        ...patch,
      },
    });

  const patchFeatures = (patch: Partial<typeof features>) =>
    patchDraftConfig({ features: { ...features, ...patch } });

  const handleAddPreset = () => {
    const name = newPresetName.trim();
    if (!name) return;
    createPresetMutation.mutate({ name, configJson: JSON.stringify(draftConfig) });
    setNewPresetName('');
  };

  const handleSave = async (): Promise<void> => {
    saveActivePreset();
    const { draftConfig: latestDraft } = useLayoutPresetsStore.getState();
    const mergedDraft: MenuLayoutConfig = {
      ...latestDraft,
      customCss,
    };
    const defaultCurrency = currencies.find((c) => c.isDefault)?.code ?? 'USD';
    await saveDraftMutation.mutateAsync({
      locationId: selectedLocationId,
      draftConfigJson: JSON.stringify(mergedDraft),
      customCss,
      defaultLocale: enabledLocales[0] ?? 'en',
      enabledLocales,
      defaultCurrency,
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 1800);
  };

  const handlePublish = async (): Promise<void> => {
    try {
      await handleSave();
      await publishMutation.mutateAsync(selectedLocationId);
    } catch {
      // Mutation hooks surface errors via toast.
    }
  };

  const handleApplyPrompt = (rawText?: string) => {
    const text = (rawText ?? aiPrompt).trim();
    if (!text) return;
    const { patch, locale } = applyMenuStylePrompt(text, { ...draftConfig, features });
    if (Object.keys(patch).length > 0) {
      patchDraftConfig(patch);
    }
    if (locale) setPreviewLocale(locale);
    setAiPrompt('');
  };

  const handleTranslateAll = () => {
    const entries = menuCategories.flatMap((cat) => [
      { entityType: 'category', entityId: cat.id, field: 'name', sourceText: cat.name },
      ...cat.items.flatMap((item) => [
        { entityType: 'item', entityId: item.id, field: 'name', sourceText: item.name },
        ...(item.description
          ? [
              {
                entityType: 'item',
                entityId: item.id,
                field: 'description',
                sourceText: item.description,
              },
            ]
          : []),
      ]),
    ]);
    translateMenuMutation.mutate({
      locationId: selectedLocationId,
      targetLocales: enabledLocales,
      entries,
    });
  };

  const previewConfig: MenuLayoutConfig = {
    ...draftConfig,
    customCss,
    features,
  };

  const previewCurrencies = currencies.map((c) => ({
    code: c.code ?? 'USD',
    symbol: c.symbol ?? '$',
    exchangeRate: c.exchangeRate ?? 1,
    decimalPlaces: c.decimalPlaces ?? 2,
    isDefault: Boolean(c.isDefault),
  }));

  const previewStampCards = platformLoyaltyEnabled
    ? stampCards
        .filter((s) => s.isActive !== false)
        .map((s) => ({
          id: s.id ?? '',
          title: s.title ?? '',
          description: s.description ?? '',
          stampsRequired: s.stampsRequired ?? 10,
          rewardLabel: s.rewardLabel ?? '',
        }))
    : [];

  const previewBanners = platformBannersEnabled
    ? menuBanners
        .filter((b) => b.isActive !== false)
        .map((b) => ({
          id: String(b.id ?? ''),
          title: b.title ?? '',
          body: b.body ?? '',
          imageUrl: b.imageUrl || undefined,
          linkUrl: b.linkUrl || undefined,
        }))
    : [];

  const railItems: { id: RailTab; label: string; icon: React.ReactNode }[] = [
    {
      id: 'themes',
      label: t('design.studio.rail.themes'),
      icon: <LayoutGrid className="w-[18px] h-[18px]" />,
    },
    {
      id: 'brand',
      label: t('design.studio.rail.brand'),
      icon: <Palette className="w-[18px] h-[18px]" />,
    },
    {
      id: 'layout',
      label: t('design.studio.rail.layout'),
      icon: <SlidersHorizontal className="w-[18px] h-[18px]" />,
    },
    {
      id: 'pages',
      label: t('design.studio.rail.pages'),
      icon: (
        <span className="relative inline-flex">
          <FileText className="w-[18px] h-[18px]" />
          {pendingReviewCount > 0 && (
            <span
              className="absolute -top-1 -end-1.5 min-w-[14px] h-[14px] px-0.5 rounded-full text-[9px] font-bold flex items-center justify-center bg-amber-500 text-white"
              aria-label={`${pendingReviewCount} pending reviews`}
            >
              {pendingReviewCount}
            </span>
          )}
        </span>
      ),
    },
    {
      id: 'features',
      label: t('design.studio.rail.features'),
      icon: <Store className="w-[18px] h-[18px]" />,
    },
    {
      id: 'ai',
      label: t('design.studio.rail.ai'),
      icon: <Sparkles className="w-[18px] h-[18px]" />,
    },
    ...(platformCustomCssEnabled
      ? [
          {
            id: 'css' as const,
            label: t('design.studio.rail.css'),
            icon: <Code2 className="w-[18px] h-[18px]" />,
          },
        ]
      : []),
  ];

  const deviceWidth = viewport === 'mobile' ? 392 : viewport === 'tablet' ? 680 : '100%';

  return (
    <>
      <HelmetContainer title={t('design.pageTitle')} />
      <div
        className="relative flex flex-col h-[calc(100vh-56px)] min-h-[600px] overflow-hidden"
        style={{ background: 'var(--dash-bg)', color: 'var(--dash-text)' }}
      >
        {/* Top Studio Toolbar (sticky, wraps cleanly on mobile & narrow desktop) */}
        <header
          className="sticky top-0 z-20 shrink-0 min-h-14 py-2 px-3 sm:px-4 flex flex-wrap items-center justify-between gap-2 border-b"
          style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
        >
          <div className="flex flex-wrap items-center gap-2 min-w-0">
            <h1 className="text-sm font-semibold tracking-tight whitespace-nowrap">
              {t('design.studio.title')}
            </h1>
            {branches.length > 0 && (
              <>
                <span className="hidden sm:inline" style={{ color: 'var(--dash-text-dim)' }}>
                  /
                </span>
                <div className="w-48 sm:w-56">
                  <Select
                    aria-label={t('design.studio.branchLabel')}
                    value={String(selectedLocationId)}
                    onChange={(e) => setSelectedLocationId(Number(e.target.value))}
                    className="h-9 py-1 text-xs leading-normal whitespace-nowrap"
                  >
                    <option value="0">{t('design.studio.defaultBranch')}</option>
                    {branches.map((b) => (
                      <option key={b.id} value={String(b.id)}>
                        {b.name}
                      </option>
                    ))}
                  </Select>
                </div>
                {branches.length > 1 && (
                  <div className="hidden 2xl:flex items-center gap-1.5">
                    <div className="w-36">
                      <Select
                        aria-label={t('design.studio.copyFromBranch')}
                        value={String(cloneSourceId)}
                        onChange={(e) => setCloneSourceId(Number(e.target.value))}
                        className="h-8 text-xs"
                      >
                        <option value="0">{t('design.studio.copyFromBranch')}</option>
                        {branches
                          .filter((b) => b.id !== selectedLocationId)
                          .map((b) => (
                            <option key={b.id} value={String(b.id)}>
                              {b.name}
                            </option>
                          ))}
                      </Select>
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={cloneMutation.isPending || cloneSourceId === selectedLocationId}
                      onClick={() =>
                        cloneMutation.mutate({
                          sourceLocationId: cloneSourceId,
                          targetLocationId: selectedLocationId,
                        })
                      }
                    >
                      <Copy className="w-3.5 h-3.5 me-1.5" />
                      {t('design.studio.copyAction')}
                    </Button>
                  </div>
                )}
              </>
            )}
          </div>

          {/* Center Viewport + Locale Controls */}
          <div className="hidden lg:flex items-center gap-2 shrink-0">
            <Tabs value={viewport} onValueChange={(v) => setViewport(v as DeviceViewport)}>
              <TabsList>
                <TabsTrigger
                  value="mobile"
                  className="px-2 py-1 text-xs gap-1.5"
                  aria-label={t('design.studio.viewports.mobile')}
                >
                  <Smartphone className="w-3.5 h-3.5" />
                  <span className="hidden xl:inline">{t('design.studio.viewports.mobile')}</span>
                </TabsTrigger>
                <TabsTrigger
                  value="tablet"
                  className="px-2 py-1 text-xs gap-1.5"
                  aria-label={t('design.studio.viewports.tablet')}
                >
                  <Tablet className="w-3.5 h-3.5" />
                  <span className="hidden xl:inline">{t('design.studio.viewports.tablet')}</span>
                </TabsTrigger>
                <TabsTrigger
                  value="desktop"
                  className="px-2 py-1 text-xs gap-1.5"
                  aria-label={t('design.studio.viewports.desktop')}
                >
                  <Monitor className="w-3.5 h-3.5" />
                  <span className="hidden xl:inline">{t('design.studio.viewports.desktop')}</span>
                </TabsTrigger>
              </TabsList>
            </Tabs>

            <Tabs value={previewLocale} onValueChange={setPreviewLocale}>
              <TabsList>
                {(enabledLocales.length > 0 ? enabledLocales : ['en', 'fr', 'ar']).map((loc) => (
                  <TabsTrigger key={loc} value={loc} className="px-2 py-1 text-xs uppercase">
                    {loc}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
          </div>

          {/* Right Save, Publish & Mobile Preview Toggle */}
          <div className="flex items-center gap-1.5 sm:gap-2 shrink-0">
            <Button
              variant="outline"
              size="sm"
              className="lg:hidden"
              onClick={() => setMobilePane((p) => (p === 'editor' ? 'preview' : 'editor'))}
            >
              <Eye className="w-3.5 h-3.5 me-1" />
              {mobilePane === 'editor'
                ? t('design.studio.mobilePreview', 'Preview')
                : t('design.studio.mobileEditor', 'Editor')}
            </Button>
            {remoteProfile?.publishedAt && (
              <Badge variant="success" className="hidden 2xl:inline-flex">
                {t('design.studio.liveBadge')}
              </Badge>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={() => void handleSave()}
              disabled={saveDraftMutation.isPending}
            >
              {saved ? (
                <>
                  <Check className="w-3.5 h-3.5 me-1" />
                  {t('design.studio.saved')}
                </>
              ) : (
                t('design.studio.save')
              )}
            </Button>
            <Button
              variant="accent"
              size="sm"
              onClick={() => void handlePublish()}
              disabled={saveDraftMutation.isPending || publishMutation.isPending}
            >
              {t('design.studio.publish')}
            </Button>
          </div>
        </header>

        {/* Main Studio Body: Icon Rail + Contextual Inspector + Neutral Stage */}
        <div className="flex flex-1 min-h-0 overflow-hidden">
          {/* 64px Left Icon Rail */}
          <nav
            aria-label={t('design.studio.railLabel')}
            className={`${
              mobilePane === 'preview' ? 'hidden lg:flex' : 'flex'
            } w-14 sm:w-16 shrink-0 flex-col items-center gap-1 py-3 px-1 border-e overflow-y-auto`}
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            {railItems.map((item) => {
              const active = activeTab === item.id;
              return (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => {
                    setActiveTab(item.id);
                    setMobilePane('editor');
                  }}
                  className="w-full py-2 px-1 rounded-xl flex flex-col items-center gap-1 text-[10px] font-semibold transition-colors"
                  style={{
                    background: active ? 'var(--dash-active)' : 'transparent',
                    color: active ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                  }}
                >
                  {item.icon}
                  <span className="truncate max-w-full">{item.label}</span>
                </button>
              );
            })}
          </nav>

          {/* Responsive Contextual Inspector Panel */}
          <aside
            className={`${
              mobilePane === 'preview' ? 'hidden lg:block' : 'block'
            } flex-1 lg:flex-none lg:w-72 xl:w-80 shrink-0 overflow-y-auto p-4 space-y-5 border-e`}
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            {/* 1. THEMES TAB */}
            {activeTab === 'themes' && (
              <div className="space-y-5">
                <div>
                  <h2 className="text-sm font-semibold">{t('design.studio.themesTitle')}</h2>
                  <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.themesSubtitle')}
                  </p>
                </div>

                <div className="space-y-2">
                  {MENU_STARTER_THEMES.map((theme) => (
                    <button
                      key={theme.id}
                      type="button"
                      onClick={() => {
                        patchDraftConfig(theme.patch);
                        if (theme.locale) setPreviewLocale(theme.locale);
                      }}
                      className="w-full p-3 rounded-xl text-start flex items-center justify-between gap-3 border transition-colors hover:border-[var(--dash-border-focus)]"
                      style={{
                        background: 'var(--dash-card)',
                        borderColor: 'var(--dash-border)',
                      }}
                    >
                      <div className="min-w-0">
                        <div className="text-[13px] font-semibold truncate">{theme.name}</div>
                        <div
                          className="text-[11px] truncate"
                          style={{ color: 'var(--dash-text-muted)' }}
                        >
                          {theme.subtitle}
                        </div>
                      </div>
                      <div className="flex items-center gap-1 shrink-0">
                        {theme.swatches.map((hex, i) => (
                          <span
                            key={i}
                            className="w-4 h-4 rounded-full border border-black/15"
                            style={{ background: hex }}
                          />
                        ))}
                      </div>
                    </button>
                  ))}
                </div>

                <Separator />

                {/* Saved Tenant Presets */}
                <div className="space-y-2.5">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.layouts')}
                  </Label>
                  <div className="flex gap-2">
                    <Input
                      value={newPresetName}
                      onChange={(e) => setNewPresetName(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && handleAddPreset()}
                      placeholder={t('design.newLayoutPlaceholder')}
                      className="h-9 text-xs"
                    />
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={handleAddPreset}
                      aria-label={t('design.studio.addPreset')}
                    >
                      <Plus className="w-4 h-4" />
                    </Button>
                  </div>

                  <div className="space-y-1.5">
                    {presets.map((preset) => {
                      const active = preset.id === activePresetId;
                      return (
                        <div
                          key={preset.id}
                          className="flex items-center justify-between gap-2 px-3 py-2 rounded-lg border"
                          style={{
                            background: active ? 'var(--dash-active)' : 'var(--dash-card)',
                            borderColor: active ? 'var(--dash-border-focus)' : 'var(--dash-border)',
                          }}
                        >
                          <button
                            type="button"
                            onClick={() => activatePreset(preset.id)}
                            className="flex-1 text-start text-xs font-medium truncate"
                          >
                            {preset.name}
                          </button>
                          {presets.length > 1 && (
                            <button
                              type="button"
                              onClick={() => deletePreset(preset.id)}
                              className="p-1 rounded hover:opacity-80"
                              style={{ color: 'var(--dash-text-muted)' }}
                              aria-label={`Delete ${preset.name}`}
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          )}
                        </div>
                      );
                    })}

                    {remotePresets.map((rp) => (
                      <div
                        key={rp.id}
                        className="flex items-center justify-between gap-2 px-3 py-2 rounded-lg border"
                        style={{
                          background: 'var(--dash-card)',
                          borderColor: 'var(--dash-border)',
                        }}
                      >
                        <button
                          type="button"
                          onClick={() => {
                            try {
                              if (rp.configJson) {
                                patchDraftConfig(
                                  JSON.parse(rp.configJson) as Partial<MenuLayoutConfig>
                                );
                              }
                            } catch {
                              // Ignore malformed saved preset JSON.
                            }
                          }}
                          className="flex-1 text-start text-xs font-medium truncate"
                        >
                          {rp.name}
                        </button>
                        {rp.id && (
                          <button
                            type="button"
                            onClick={() => deletePresetMutation.mutate(rp.id!)}
                            className="p-1 rounded hover:opacity-80"
                            style={{ color: 'var(--dash-text-muted)' }}
                            aria-label={`Delete ${rp.name ?? ''}`}
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* 2. BRAND & TYPOGRAPHY TAB */}
            {activeTab === 'brand' && (
              <div className="space-y-5">
                <div>
                  <h2 className="text-sm font-semibold">{t('design.studio.brandTitle')}</h2>
                  <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.brandSubtitle')}
                  </p>
                </div>

                {/* Brand Media: Cover Image & Logo */}
                <div className="space-y-3">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.studio.brandMedia')}
                  </Label>

                  {/* Cover Image */}
                  <div
                    className="rounded-xl border p-2.5 space-y-2"
                    style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-xs font-medium">{t('design.studio.coverImage')}</span>
                      <div className="flex items-center gap-1">
                        {draftConfig.header.coverImageUrl && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => patchHeader({ coverImageUrl: undefined })}
                            aria-label={t('design.studio.resetImage')}
                          >
                            <RotateCcw className="w-3.5 h-3.5" />
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={uploadingField === 'cover'}
                          onClick={() => coverInputRef.current?.click()}
                        >
                          <Upload className="w-3.5 h-3.5 me-1" />
                          {uploadingField === 'cover'
                            ? t('design.studio.uploading')
                            : t('design.studio.upload')}
                        </Button>
                      </div>
                    </div>
                    <input
                      ref={coverInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        void handleMediaUpload('cover', e.target.files?.[0]);
                        e.target.value = '';
                      }}
                    />
                    <div
                      className="h-20 w-full rounded-lg overflow-hidden border flex items-center justify-center"
                      style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-hover)' }}
                    >
                      {cdnImageUrl(draftConfig.header.coverImageUrl || businessInfo.coverImage) ? (
                        <img
                          src={cdnImageUrl(
                            draftConfig.header.coverImageUrl || businessInfo.coverImage
                          )}
                          alt={t('design.studio.coverImage')}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <span
                          className="text-[11px] flex items-center gap-1.5"
                          style={{ color: 'var(--dash-text-muted)' }}
                        >
                          <Upload className="w-3.5 h-3.5" />
                          {t('design.studio.coverImage')}
                        </span>
                      )}
                    </div>
                    <Input
                      value={draftConfig.header.coverImageUrl ?? ''}
                      onChange={(e) => patchHeader({ coverImageUrl: e.target.value })}
                      placeholder={t('design.studio.imageUrlPlaceholder')}
                      className="h-8 text-xs"
                    />
                  </div>

                  {/* Brand Logo */}
                  <div
                    className="rounded-xl border p-2.5 space-y-2"
                    style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2.5 min-w-0">
                        <div
                          className="w-10 h-10 rounded-lg overflow-hidden border shrink-0 flex items-center justify-center"
                          style={{
                            borderColor: 'var(--dash-border)',
                            background: 'var(--dash-hover)',
                          }}
                        >
                          {cdnImageUrl(draftConfig.header.logoUrl || businessInfo.logo) ? (
                            <img
                              src={cdnImageUrl(draftConfig.header.logoUrl || businessInfo.logo)}
                              alt={t('design.studio.logoImage')}
                              className="w-full h-full object-cover"
                            />
                          ) : (
                            <Store
                              className="w-4 h-4"
                              style={{ color: 'var(--dash-text-muted)' }}
                            />
                          )}
                        </div>
                        <span className="text-xs font-medium truncate">
                          {t('design.studio.logoImage')}
                        </span>
                      </div>
                      <div className="flex items-center gap-1 shrink-0">
                        {draftConfig.header.logoUrl && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => patchHeader({ logoUrl: undefined })}
                            aria-label={t('design.studio.resetImage')}
                          >
                            <RotateCcw className="w-3.5 h-3.5" />
                          </Button>
                        )}
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={uploadingField === 'logo'}
                          onClick={() => logoInputRef.current?.click()}
                        >
                          <Upload className="w-3.5 h-3.5 me-1" />
                          {uploadingField === 'logo'
                            ? t('design.studio.uploading')
                            : t('design.studio.upload')}
                        </Button>
                      </div>
                    </div>
                    <input
                      ref={logoInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        void handleMediaUpload('logo', e.target.files?.[0]);
                        e.target.value = '';
                      }}
                    />
                    <Input
                      value={draftConfig.header.logoUrl ?? ''}
                      onChange={(e) => patchHeader({ logoUrl: e.target.value })}
                      placeholder={t('design.studio.imageUrlPlaceholder')}
                      className="h-8 text-xs"
                    />
                  </div>

                  <p className="text-[11px]" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.itemImagesHint')}
                  </p>
                </div>

                <Separator />

                <SegmentedControl<ThemeMode>
                  label={t('design.studio.appearance')}
                  value={themeMode}
                  options={[
                    { label: t('design.studio.modeLight'), value: 'light' },
                    { label: t('design.studio.modeDark'), value: 'dark' },
                  ]}
                  onChange={(mode) =>
                    patchDraftConfig({
                      mode,
                      colors:
                        mode === 'dark'
                          ? {
                              primary: '#ffffff',
                              accent: colorAccent,
                              bg: '#141414',
                              surface: '#1c1c1c',
                              text: '#f5f5f5',
                            }
                          : {
                              primary: '#111111',
                              accent: colorAccent,
                              bg: '#ffffff',
                              surface: '#ffffff',
                              text: '#111111',
                            },
                    })
                  }
                />

                <div className="space-y-2">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.colors.accent')}
                  </Label>
                  <div className="flex flex-wrap gap-2">
                    {MENU_ACCENT_SWATCHES.map((swatch) => (
                      <button
                        key={swatch}
                        type="button"
                        aria-label={`Accent ${swatch}`}
                        onClick={() => patchColors({ accent: swatch })}
                        className="w-7 h-7 rounded-full border-2 transition-transform hover:scale-105"
                        style={{
                          background: swatch,
                          borderColor:
                            colorAccent.toLowerCase() === swatch.toLowerCase()
                              ? 'var(--dash-text)'
                              : 'transparent',
                        }}
                      />
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  {(
                    [
                      { key: 'primary', label: t('design.colors.primary'), val: colorPrimary },
                      { key: 'accent', label: t('design.colors.accent'), val: colorAccent },
                      { key: 'bg', label: t('design.colors.background'), val: colorBg },
                      { key: 'surface', label: t('design.studio.surfaceColor'), val: colorSurface },
                    ] as const
                  ).map((c) => (
                    <label
                      key={c.key}
                      className="flex items-center justify-between gap-2 px-2.5 py-2 rounded-lg border text-xs cursor-pointer"
                      style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
                    >
                      <span className="truncate">{c.label}</span>
                      <input
                        type="color"
                        value={c.val}
                        onChange={(e) => patchColors({ [c.key]: e.target.value })}
                        className="w-6 h-6 rounded border-0 bg-transparent cursor-pointer"
                      />
                    </label>
                  ))}
                </div>

                <Separator />

                <div className="space-y-2.5">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.typography.fontFamily')}
                  </Label>
                  <div className="grid grid-cols-2 gap-1.5">
                    {FONT_OPTIONS.map((f) => {
                      const active = fontFamily === f.id;
                      return (
                        <button
                          key={f.id}
                          type="button"
                          onClick={() =>
                            patchDraftConfig({
                              typography: { headingFont: f.id, bodyFont: f.id, fontSize },
                            })
                          }
                          className="px-2.5 py-2 rounded-lg text-xs font-medium text-start border transition-colors"
                          style={{
                            background: active ? 'var(--dash-active)' : 'var(--dash-card)',
                            borderColor: active ? 'var(--dash-border-focus)' : 'var(--dash-border)',
                          }}
                        >
                          {f.label}
                        </button>
                      );
                    })}
                  </div>

                  <SegmentedControl<MenuFontSize>
                    label={t('design.typography.fontSize')}
                    value={fontSize}
                    options={[
                      { label: 'S', value: 'sm' },
                      { label: 'M', value: 'md' },
                      { label: 'L', value: 'lg' },
                    ]}
                    onChange={(size) =>
                      patchDraftConfig({
                        typography: {
                          headingFont: fontFamily,
                          bodyFont: fontFamily,
                          fontSize: size,
                        },
                      })
                    }
                  />
                </div>
              </div>
            )}

            {/* 3. LAYOUT TAB */}
            {activeTab === 'layout' && (
              <div className="space-y-5">
                <div>
                  <h2 className="text-sm font-semibold">{t('design.studio.layoutTitle')}</h2>
                  <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.layoutSubtitle')}
                  </p>
                </div>

                <SegmentedControl<ItemLayout>
                  label={t('design.items.layout')}
                  value={draftConfig.items.layout}
                  options={[
                    { label: t('design.items.layoutGrid'), value: 'grid' },
                    { label: t('design.items.layoutList'), value: 'list' },
                    { label: t('design.items.layoutMinimal'), value: 'minimal' },
                  ]}
                  onChange={(layout) =>
                    patchItems({
                      layout,
                      imagePosition:
                        layout === 'list' ? 'right' : layout === 'minimal' ? 'none' : 'top',
                    })
                  }
                />

                {draftConfig.items.layout !== 'minimal' && (
                  <>
                    <SegmentedControl<ImagePosition>
                      label={t('design.items.imagePosition')}
                      value={draftConfig.items.imagePosition}
                      options={[
                        { label: t('design.items.imageTop'), value: 'top' },
                        { label: t('design.items.imageRight'), value: 'right' },
                        { label: t('design.items.imageLeft'), value: 'left' },
                        { label: t('design.none'), value: 'none' },
                      ]}
                      onChange={(imagePosition) => patchItems({ imagePosition })}
                    />
                    <SegmentedControl<ImageSize>
                      label={t('design.items.imageSize')}
                      value={draftConfig.items.imageSize}
                      options={[
                        { label: 'S', value: 'sm' },
                        { label: 'M', value: 'md' },
                        { label: 'L', value: 'lg' },
                      ]}
                      onChange={(imageSize) => patchItems({ imageSize })}
                    />
                  </>
                )}

                <SegmentedControl<CardRadius>
                  label={t('design.items.cardRadius')}
                  value={draftConfig.items.cardRadius}
                  options={[
                    { label: t('design.none'), value: 'none' },
                    { label: 'S', value: 'sm' },
                    { label: 'M', value: 'md' },
                    { label: 'L', value: 'lg' },
                  ]}
                  onChange={(cardRadius) => patchItems({ cardRadius })}
                />

                <div className="space-y-1">
                  <SwitchRow
                    id="item-desc"
                    label={t('design.items.showDescription')}
                    checked={draftConfig.items.showDescription}
                    onChange={(showDescription) => patchItems({ showDescription })}
                  />
                  <SwitchRow
                    id="item-badges"
                    label={t('design.items.showBadges')}
                    checked={draftConfig.items.showBadges}
                    onChange={(showBadges) => patchItems({ showBadges })}
                  />
                </div>

                <Separator />

                <div className="space-y-3">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.sections.header')}
                  </Label>
                  <SegmentedControl<CoverHeight>
                    label={t('design.header.coverHeight')}
                    value={draftConfig.header.coverHeight}
                    options={[
                      { label: t('design.none'), value: 'none' },
                      { label: 'S', value: 'sm' },
                      { label: 'M', value: 'md' },
                      { label: 'L', value: 'lg' },
                    ]}
                    onChange={(coverHeight) => patchHeader({ coverHeight })}
                  />
                  <SwitchRow
                    id="hdr-logo"
                    label={t('design.header.showLogo')}
                    checked={draftConfig.header.showLogo}
                    onChange={(showLogo) => patchHeader({ showLogo })}
                  />
                  <SwitchRow
                    id="hdr-desc"
                    label={t('design.header.showDescription')}
                    checked={draftConfig.header.showDescription}
                    onChange={(showDescription) => patchHeader({ showDescription })}
                  />
                  <SwitchRow
                    id="hdr-meta"
                    label={t('design.header.showMeta')}
                    checked={draftConfig.header.showMeta}
                    onChange={(showMeta) => patchHeader({ showMeta })}
                  />
                </div>

                <Separator />

                <div className="space-y-3">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.sections.nav')}
                  </Label>
                  <SegmentedControl<NavStyle>
                    label={t('design.nav.style')}
                    value={draftConfig.nav.style}
                    options={[
                      { label: t('design.nav.styleUnderline'), value: 'underline' },
                      { label: t('design.nav.stylePill'), value: 'pill' },
                      { label: t('design.nav.styleFlat'), value: 'flat' },
                    ]}
                    onChange={(style) => patchNav({ style })}
                  />
                  <SegmentedControl<NavSize>
                    label={t('design.nav.size')}
                    value={draftConfig.nav.size}
                    options={[
                      { label: 'S', value: 'sm' },
                      { label: 'M', value: 'md' },
                      { label: 'L', value: 'lg' },
                    ]}
                    onChange={(size) => patchNav({ size })}
                  />
                </div>

                <Separator />

                <div className="space-y-1">
                  <div className="flex items-center justify-between gap-2">
                    <Label className="text-[11px] uppercase tracking-wider">
                      {t('design.sections.pages')}
                    </Label>
                    <button
                      type="button"
                      onClick={() => setActiveTab('pages')}
                      className="text-[11px] font-semibold hover:underline"
                      style={{ color: colorAccent }}
                    >
                      {t('design.pagesEditor.editContent')} →
                    </button>
                  </div>
                  <SwitchRow
                    id="page-about"
                    label={t('design.pages.about')}
                    checked={Boolean(draftConfig.pages.about)}
                    onChange={(about) => patchPages({ about })}
                  />
                  <SwitchRow
                    id="page-gallery"
                    label={t('design.pages.gallery')}
                    checked={Boolean(draftConfig.pages.gallery)}
                    onChange={(gallery) => patchPages({ gallery })}
                  />
                  <SwitchRow
                    id="page-reviews"
                    label={t('design.pages.reviews')}
                    checked={Boolean(draftConfig.pages.reviews)}
                    onChange={(reviews) => patchPages({ reviews })}
                  />
                  <SwitchRow
                    id="page-offers"
                    label={t('design.pages.offers')}
                    checked={Boolean(draftConfig.pages.offers)}
                    onChange={(offers) => patchPages({ offers })}
                  />
                </div>
              </div>
            )}

            {/* 3b. PAGES CONTENT & REVIEW MODERATION TAB */}
            {activeTab === 'pages' && (
              <PagesEditor
                selectedLocationId={selectedLocationId}
                pages={draftConfig.pages}
                content={draftConfig.content ?? {}}
                defaultAddress={businessInfo.address}
                defaultPhone={businessInfo.phone}
                onPatchPages={patchPages}
                onPatchContent={patchContent}
              />
            )}

            {/* 4. FEATURES, CURRENCIES, LANGUAGES, LOYALTY & PAYMENTS TAB */}
            {activeTab === 'features' && (
              <div className="space-y-5">
                <div>
                  <h2 className="text-sm font-semibold">{t('design.sections.features')}</h2>
                  <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.featuresSubtitle')}
                  </p>
                </div>

                {platformOrderingEnabled && (
                  <div className="space-y-1">
                    <SwitchRow
                      id="feat-ordering"
                      label={t('design.studio.orderingEnabled')}
                      checked={features.enableOrdering}
                      onChange={(enableOrdering) => patchFeatures({ enableOrdering })}
                    />
                    <SwitchRow
                      id="feat-dinein"
                      label={t('design.studio.modeDineIn')}
                      checked={features.fulfillmentModes.dineIn}
                      onChange={(dineIn) =>
                        patchFeatures({
                          fulfillmentModes: { ...features.fulfillmentModes, dineIn },
                        })
                      }
                    />
                    <SwitchRow
                      id="feat-selfservice"
                      label={t('design.studio.modeSelfService')}
                      checked={features.fulfillmentModes.selfService}
                      onChange={(selfService) =>
                        patchFeatures({
                          fulfillmentModes: { ...features.fulfillmentModes, selfService },
                        })
                      }
                    />
                    <SwitchRow
                      id="feat-pickup"
                      label={t('design.studio.modePickup')}
                      checked={features.fulfillmentModes.pickup}
                      onChange={(pickup) =>
                        patchFeatures({
                          fulfillmentModes: { ...features.fulfillmentModes, pickup },
                        })
                      }
                    />
                    <SwitchRow
                      id="feat-delivery"
                      label={t('design.studio.modeDelivery')}
                      checked={features.fulfillmentModes.delivery}
                      onChange={(delivery) =>
                        patchFeatures({
                          fulfillmentModes: { ...features.fulfillmentModes, delivery },
                        })
                      }
                    />
                  </div>
                )}

                <Separator />

                <div className="space-y-1">
                  {platformTableActionsEnabled && (
                    <>
                      <Label className="text-[11px] uppercase tracking-wider">
                        {t('design.studio.tableService')}
                      </Label>
                      <SwitchRow
                        id="act-waiter"
                        label={t('design.studio.callStaff')}
                        checked={features.tableActions.callWaiter}
                        onChange={(callWaiter) =>
                          patchFeatures({ tableActions: { ...features.tableActions, callWaiter } })
                        }
                      />
                      <SwitchRow
                        id="act-bill"
                        label={t('design.studio.requestBill')}
                        checked={features.tableActions.requestBill}
                        onChange={(requestBill) =>
                          patchFeatures({ tableActions: { ...features.tableActions, requestBill } })
                        }
                      />
                      <SwitchRow
                        id="act-split"
                        label={t('design.studio.splitBill')}
                        checked={features.tableActions.splitBill}
                        onChange={(splitBill) =>
                          patchFeatures({ tableActions: { ...features.tableActions, splitBill } })
                        }
                      />
                    </>
                  )}
                  {platformLoyaltyEnabled && (
                    <SwitchRow
                      id="feat-loyalty"
                      label={t('design.sections.loyalty')}
                      checked={features.enableLoyaltyStamps}
                      onChange={(enableLoyaltyStamps) => patchFeatures({ enableLoyaltyStamps })}
                    />
                  )}
                  {platformCrmCaptureEnabled && (
                    <SwitchRow
                      id="feat-crm-capture"
                      label={t('design.studio.crmCapture', 'Guest CRM capture')}
                      hint={t(
                        'design.studio.crmCaptureHint',
                        'Let guests optionally leave their contact info directly into your CRM.'
                      )}
                      checked={features.enableCrmCapture}
                      onChange={(enableCrmCapture) => patchFeatures({ enableCrmCapture })}
                    />
                  )}
                  {platformBannersEnabled && (
                    <>
                      <SwitchRow
                        id="feat-banners"
                        label={t('design.studio.promoBanners')}
                        checked={features.enableBanners}
                        onChange={(enableBanners) => patchFeatures({ enableBanners })}
                      />
                      {features.enableBanners && (
                        <div className="space-y-2 pt-1 pb-2">
                          {menuBanners.map((banner) => (
                            <div
                              key={banner.id}
                              className="flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-lg border text-xs"
                              style={{
                                background: 'var(--dash-card)',
                                borderColor: 'var(--dash-border)',
                              }}
                            >
                              <div className="min-w-0">
                                <div className="font-semibold truncate">{banner.title}</div>
                                {banner.body && (
                                  <div
                                    className="truncate text-[11px]"
                                    style={{ color: 'var(--dash-text-muted)' }}
                                  >
                                    {banner.body}
                                  </div>
                                )}
                              </div>
                              {banner.id && (
                                <button
                                  type="button"
                                  onClick={() => deleteBannerMutation.mutate(String(banner.id))}
                                  style={{ color: 'var(--dash-text-muted)' }}
                                  aria-label={`Remove ${banner.title ?? ''}`}
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                </button>
                              )}
                            </div>
                          ))}
                          <Input
                            value={newBannerTitle}
                            onChange={(e) => setNewBannerTitle(e.target.value)}
                            placeholder={t(
                              'design.studio.promoBannerTitlePlaceholder',
                              'Banner title (e.g. Happy Hour 20% off)'
                            )}
                            className="h-8 text-xs"
                          />
                          <Input
                            value={newBannerBody}
                            onChange={(e) => setNewBannerBody(e.target.value)}
                            placeholder={t(
                              'design.studio.promoBannerBodyPlaceholder',
                              'Short details (optional)'
                            )}
                            className="h-8 text-xs"
                          />
                          <Button
                            variant="outline"
                            size="sm"
                            className="w-full"
                            disabled={!newBannerTitle.trim() || createBannerMutation.isPending}
                            onClick={() => {
                              createBannerMutation.mutate(
                                {
                                  title: newBannerTitle.trim(),
                                  body: newBannerBody.trim(),
                                  isActive: true,
                                },
                                {
                                  onSuccess: () => {
                                    setNewBannerTitle('');
                                    setNewBannerBody('');
                                    toast.success(
                                      t('menu.bannerCreated', 'Promo banner added to menu')
                                    );
                                  },
                                }
                              );
                            }}
                          >
                            <Plus className="w-3.5 h-3.5 me-1.5" />
                            {t('design.studio.addBanner', 'Add banner')}
                          </Button>
                        </div>
                      )}
                    </>
                  )}
                  <SwitchRow
                    id="feat-search"
                    label={t('design.studio.searchAndFilters')}
                    checked={features.enableSearch}
                    onChange={(enableSearch) =>
                      patchFeatures({ enableSearch, enableDietaryFilters: enableSearch })
                    }
                  />
                </div>

                <Separator />

                {/* Payment Redirect */}
                <div className="space-y-2.5">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.sections.payment')}
                  </Label>
                  <SegmentedControl<'pos' | 'redirect'>
                    value={paymentMode === 'redirect' ? 'redirect' : 'pos'}
                    options={[
                      { label: t('design.studio.payAtPos'), value: 'pos' },
                      { label: t('design.studio.payExternal'), value: 'redirect' },
                    ]}
                    onChange={(mode) =>
                      patchDraftConfig({ payment: { ...draftConfig.payment, mode } })
                    }
                  />
                  {paymentMode === 'redirect' && (
                    <Input
                      value={paymentRedirectUrl}
                      onChange={(e) =>
                        patchDraftConfig({
                          payment: {
                            ...draftConfig.payment,
                            mode: 'redirect',
                            redirectUrl: e.target.value,
                          },
                        })
                      }
                      placeholder="https://pay.yourstore.com/checkout"
                      className="h-9 text-xs"
                    />
                  )}
                </div>

                {platformCurrenciesEnabled && (
                  <>
                    <Separator />

                    {/* Dynamic Currencies */}
                    <div className="space-y-2.5">
                      <Label className="text-[11px] uppercase tracking-wider">
                        {t('design.sections.currencies')}
                      </Label>
                      <div className="space-y-1.5">
                        {currencies.map((cur) => (
                          <div
                            key={cur.code}
                            className="flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-lg border text-xs"
                            style={{
                              background: 'var(--dash-card)',
                              borderColor: 'var(--dash-border)',
                            }}
                          >
                            <div className="flex items-center gap-2 min-w-0">
                              <span className="font-semibold">{cur.code}</span>
                              <span style={{ color: 'var(--dash-text-muted)' }}>
                                {cur.symbol} · ×{cur.exchangeRate}
                              </span>
                              {cur.isDefault && (
                                <Badge variant="secondary">{t('design.studio.defaultTag')}</Badge>
                              )}
                            </div>
                            {!cur.isDefault && cur.code && (
                              <button
                                type="button"
                                onClick={() =>
                                  deleteCurrencyMutation.mutate({
                                    locationId: selectedLocationId,
                                    code: cur.code!,
                                  })
                                }
                                style={{ color: 'var(--dash-text-muted)' }}
                                aria-label={`Remove ${cur.code}`}
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                              </button>
                            )}
                          </div>
                        ))}
                      </div>
                      <div className="grid grid-cols-3 gap-1.5">
                        <Input
                          value={newCurrencyCode}
                          onChange={(e) => setNewCurrencyCode(e.target.value.toUpperCase())}
                          placeholder="EUR"
                          className="h-8 text-xs"
                        />
                        <Input
                          value={newCurrencySymbol}
                          onChange={(e) => setNewCurrencySymbol(e.target.value)}
                          placeholder="€"
                          className="h-8 text-xs"
                        />
                        <Input
                          type="number"
                          step="0.01"
                          value={newCurrencyRate}
                          onChange={(e) => setNewCurrencyRate(e.target.value)}
                          placeholder="0.92"
                          className="h-8 text-xs"
                        />
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        className="w-full"
                        disabled={
                          !newCurrencyCode.trim() ||
                          !newCurrencySymbol.trim() ||
                          upsertCurrencyMutation.isPending
                        }
                        onClick={() => {
                          upsertCurrencyMutation.mutate({
                            locationId: selectedLocationId,
                            code: newCurrencyCode.trim().toUpperCase(),
                            symbol: newCurrencySymbol.trim(),
                            exchangeRate: Number(newCurrencyRate) || 1,
                          });
                          setNewCurrencyCode('');
                          setNewCurrencySymbol('');
                          setNewCurrencyRate('1');
                        }}
                      >
                        <Plus className="w-3.5 h-3.5 me-1.5" />
                        {t('design.studio.addCurrency')}
                      </Button>
                    </div>
                  </>
                )}

                <Separator />

                {/* Languages & AI Translation */}
                <div className="space-y-2.5">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.sections.languages')}
                  </Label>
                  <div className="flex flex-wrap gap-1.5">
                    {(['en', 'fr', 'ar'] as const).map((loc) => {
                      const active = enabledLocales.includes(loc);
                      return (
                        <Button
                          key={loc}
                          size="sm"
                          variant={active ? 'default' : 'outline'}
                          onClick={() =>
                            setEnabledLocales((prev) =>
                              prev.includes(loc)
                                ? prev.length > 1
                                  ? prev.filter((l) => l !== loc)
                                  : prev
                                : [...prev, loc]
                            )
                          }
                        >
                          {loc.toUpperCase()}
                        </Button>
                      );
                    })}
                  </div>
                  {platformAiTranslateEnabled && (
                    <Button
                      variant="outline"
                      size="sm"
                      className="w-full whitespace-normal text-center h-auto py-2 leading-snug"
                      disabled={translateMenuMutation.isPending || menuCategories.length === 0}
                      onClick={handleTranslateAll}
                    >
                      <Sparkles className="w-3.5 h-3.5 me-1.5 shrink-0" />
                      <span>{t('design.studio.aiTranslate')}</span>
                    </Button>
                  )}
                </div>

                {platformLoyaltyEnabled && (
                  <>
                    <Separator />

                    {/* Loyalty Stamp Cards */}
                    <div className="space-y-2.5">
                      <Label className="text-[11px] uppercase tracking-wider">
                        {t('design.sections.loyalty')}
                      </Label>
                      {stampCards.map((card) => (
                        <div
                          key={card.id}
                          className="px-2.5 py-2 rounded-lg border text-xs"
                          style={{
                            background: 'var(--dash-card)',
                            borderColor: 'var(--dash-border)',
                          }}
                        >
                          <div className="font-semibold">{card.title}</div>
                          <div style={{ color: 'var(--dash-text-muted)' }}>
                            {card.stampsRequired} → {card.rewardLabel}
                          </div>
                        </div>
                      ))}
                      <Input
                        value={newStampTitle}
                        onChange={(e) => setNewStampTitle(e.target.value)}
                        placeholder={t('design.studio.stampTitlePlaceholder')}
                        className="h-8 text-xs"
                      />
                      <div className="grid grid-cols-3 gap-1.5">
                        <Input
                          type="number"
                          min={1}
                          value={newStampTarget}
                          onChange={(e) => setNewStampTarget(e.target.value)}
                          placeholder="10"
                          className="h-8 text-xs"
                        />
                        <Input
                          value={newStampReward}
                          onChange={(e) => setNewStampReward(e.target.value)}
                          placeholder={t('design.studio.stampRewardPlaceholder')}
                          className="h-8 text-xs col-span-2"
                        />
                      </div>
                      <Button
                        variant="outline"
                        size="sm"
                        className="w-full"
                        disabled={
                          !newStampTitle.trim() ||
                          !newStampReward.trim() ||
                          upsertStampCardMutation.isPending
                        }
                        onClick={() => {
                          upsertStampCardMutation.mutate({
                            locationId: selectedLocationId,
                            title: newStampTitle.trim(),
                            stampsRequired: Math.max(1, Number(newStampTarget) || 10),
                            rewardLabel: newStampReward.trim(),
                          });
                          setNewStampTitle('');
                          setNewStampReward('');
                        }}
                      >
                        <Plus className="w-3.5 h-3.5 me-1.5" />
                        {t('design.studio.addStampCard')}
                      </Button>
                    </div>
                  </>
                )}

                <Separator />

                {/* Compact 30-day Menu Analytics + POS Operations Note */}
                <div className="space-y-2">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.studio.analyticsTitle')}
                  </Label>
                  <div className="grid grid-cols-2 gap-2">
                    {[
                      {
                        label: t('design.studio.metricScans'),
                        val: analyticsSummary?.qrScans ?? 0,
                      },
                      {
                        label: t('design.studio.metricViews'),
                        val: analyticsSummary?.menuViews ?? 0,
                      },
                      {
                        label: t('design.studio.metricCarts'),
                        val: analyticsSummary?.addToCarts ?? 0,
                      },
                      {
                        label: t('design.studio.metricOrders'),
                        val: analyticsSummary?.ordersPlaced ?? 0,
                      },
                    ].map((m) => (
                      <div
                        key={m.label}
                        className="rounded-lg border px-2.5 py-2"
                        style={{
                          background: 'var(--dash-card)',
                          borderColor: 'var(--dash-border)',
                        }}
                      >
                        <div className="text-base font-semibold">{m.val}</div>
                        <div className="text-[11px]" style={{ color: 'var(--dash-text-muted)' }}>
                          {m.label}
                        </div>
                      </div>
                    ))}
                  </div>
                  <p className="text-[11px] pt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t(
                      'design.studio.posOpsHint',
                      'Live table service requests and branch 86 / price overrides are managed in POS.'
                    )}
                  </p>
                </div>
              </div>
            )}

            {/* 6. CARET AI TAB */}
            {activeTab === 'ai' && (
              <div className="space-y-4">
                <div>
                  <h2 className="text-sm font-semibold">{t('design.sections.aiDesigner')}</h2>
                  <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.aiSubtitle')}
                  </p>
                </div>

                <Button variant="accent" className="w-full gap-2" onClick={openCaret}>
                  <CaretIcon size={16} />
                  {t('design.studio.openCaretAi')}
                </Button>

                <Separator />

                <div className="space-y-2">
                  <Label className="text-[11px] uppercase tracking-wider">
                    {t('design.studio.quickStylePrompt')}
                  </Label>
                  <div className="flex gap-2">
                    <Input
                      value={aiPrompt}
                      onChange={(e) => setAiPrompt(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && handleApplyPrompt()}
                      placeholder={t('design.studio.aiPromptPlaceholder')}
                      className="h-9 text-xs"
                    />
                    <Button size="sm" variant="default" onClick={() => handleApplyPrompt()}>
                      {t('design.studio.apply')}
                    </Button>
                  </div>
                  <div className="flex flex-wrap gap-1.5 pt-1">
                    {AI_QUICK_PROMPTS.map((chip) => (
                      <button
                        key={chip}
                        type="button"
                        onClick={() => handleApplyPrompt(chip)}
                        className="px-2.5 py-1 rounded-full text-[11px] font-medium border transition-colors hover:border-[var(--dash-border-focus)]"
                        style={{
                          background: 'var(--dash-card)',
                          borderColor: 'var(--dash-border)',
                        }}
                      >
                        {chip}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* 5. SCOPED CSS TAB */}
            {activeTab === 'css' && platformCustomCssEnabled && (
              <div className="space-y-3">
                <div>
                  <h2 className="text-sm font-semibold">{t('design.sections.advanced')}</h2>
                  <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('design.studio.cssSubtitle')}
                  </p>
                </div>
                <textarea
                  dir="ltr"
                  value={customCss}
                  onChange={(e) => setCustomCss(e.target.value)}
                  rows={12}
                  spellCheck={false}
                  placeholder=".rtylr-menu-root { letter-spacing: 0.01em; }"
                  className="w-full rounded-xl p-3 font-mono text-xs border text-left focus:outline-none focus:ring-2 focus:ring-[var(--dash-border-focus)]"
                  style={{
                    background: 'var(--dash-input-bg)',
                    color: 'var(--dash-text)',
                    borderColor: 'var(--dash-border)',
                  }}
                />
              </div>
            )}
          </aside>

          {/* Calm Neutral Preview Stage */}
          <section
            aria-label={t('design.preview')}
            className={`${
              mobilePane === 'preview' ? 'flex' : 'hidden lg:flex'
            } flex-1 min-w-0 overflow-auto p-3 sm:p-6 items-start justify-center`}
          >
            <div
              className="transition-all duration-200 overflow-hidden max-w-full"
              style={{
                width: deviceWidth,
                maxWidth: viewport === 'desktop' ? 880 : '100%',
                height: viewport === 'mobile' ? 780 : 'calc(100vh - 116px)',
                borderRadius: viewport === 'mobile' ? 28 : 14,
                border: '1px solid var(--dash-border)',
                boxShadow: '0 8px 30px rgba(0, 0, 0, 0.08)',
                transform: 'translateZ(0)',
                background: colorBg,
              }}
            >
              <div className="w-full h-full overflow-y-auto overflow-x-hidden">
                <MenuLayout
                  business={businessInfo}
                  categories={menuCategories}
                  config={previewConfig}
                  banners={previewBanners}
                  activeLocale={previewLocale}
                  availableLocales={enabledLocales}
                  onLocaleChange={setPreviewLocale}
                  currencies={platformCurrenciesEnabled ? previewCurrencies : []}
                  stampCards={previewStampCards}
                  tableLabel="Table 4"
                  enableCart={platformOrderingEnabled && features.enableOrdering}
                  previewMode
                  onTableAction={platformTableActionsEnabled ? () => undefined : undefined}
                  onCaptureCrmGuest={platformCrmCaptureEnabled ? () => undefined : undefined}
                />
              </div>
            </div>
          </section>
        </div>
      </div>
    </>
  );
};

export default Design;
