import { useEffect, useState } from 'react';
import {
  DEFAULT_FEATURES_CONFIG,
  DEFAULT_LAYOUT_CONFIG,
  type MenuLayoutConfig,
} from '@voxire/menu';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useMenuBranches,
  useMenuCurrencies,
  useMenuProfile,
  usePublishMenuProfile,
  useSaveMenuProfileDraft,
} from '../../shared/hooks/queries/useMenuBuilder';

export function normalizeMenuConfig(raw?: Partial<MenuLayoutConfig>): MenuLayoutConfig {
  return {
    ...DEFAULT_LAYOUT_CONFIG,
    ...raw,
    pages: { menu: true, ...DEFAULT_LAYOUT_CONFIG.pages, ...raw?.pages },
    header: { ...DEFAULT_LAYOUT_CONFIG.header, ...raw?.header },
    nav: { ...DEFAULT_LAYOUT_CONFIG.nav, ...raw?.nav },
    items: { ...DEFAULT_LAYOUT_CONFIG.items, ...raw?.items },
    typography: { ...DEFAULT_LAYOUT_CONFIG.typography!, ...raw?.typography },
    features: {
      ...DEFAULT_FEATURES_CONFIG,
      ...raw?.features,
      fulfillmentModes: {
        ...DEFAULT_FEATURES_CONFIG.fulfillmentModes,
        ...raw?.features?.fulfillmentModes,
      },
      tableActions: {
        ...DEFAULT_FEATURES_CONFIG.tableActions,
        ...raw?.features?.tableActions,
      },
    },
    payment: {
      mode: raw?.payment?.mode ?? 'pos',
      redirectUrl: raw?.payment?.redirectUrl ?? '',
    },
    content: {
      ...raw?.content,
    },
  };
}

export function useMenuBranchProfile() {
  const { user } = useUserStore();
  const businessSlug = (user as { businessSlug?: string } | null)?.businessSlug ?? '';
  const { data: branches = [] } = useMenuBranches(businessSlug);
  const [selectedLocationId, setSelectedLocationId] = useState<number>(0);

  const { data: remoteProfile, isLoading } = useMenuProfile(selectedLocationId);
  const { data: currencies = [] } = useMenuCurrencies(selectedLocationId);
  const saveDraftMutation = useSaveMenuProfileDraft();
  const publishMutation = usePublishMenuProfile();

  const [draftConfig, setDraftConfig] = useState<MenuLayoutConfig>(() => normalizeMenuConfig());
  const [enabledLocales, setEnabledLocales] = useState<string[]>(['en', 'fr', 'ar']);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (!remoteProfile) return;
    if (remoteProfile.draftConfigJson) {
      try {
        const parsed = JSON.parse(remoteProfile.draftConfigJson) as Partial<MenuLayoutConfig>;
        setDraftConfig(normalizeMenuConfig(parsed));
      } catch {
        setDraftConfig(normalizeMenuConfig());
      }
    } else {
      setDraftConfig(normalizeMenuConfig());
    }
    if (remoteProfile.enabledLocales && remoteProfile.enabledLocales.length > 0) {
      setEnabledLocales(remoteProfile.enabledLocales);
    }
  }, [remoteProfile]);

  const patchConfig = (patch: Partial<MenuLayoutConfig>) => {
    setDraftConfig((prev) => normalizeMenuConfig({ ...prev, ...patch }));
  };

  const saveDraft = async (overrideConfig?: MenuLayoutConfig) => {
    const cfg = overrideConfig ?? draftConfig;
    const defaultCurrency = currencies.find((c) => c.isDefault)?.code ?? 'USD';
    await saveDraftMutation.mutateAsync({
      locationId: selectedLocationId,
      draftConfigJson: JSON.stringify(cfg),
      customCss: remoteProfile?.customCss ?? '',
      defaultLocale: enabledLocales[0] ?? 'en',
      enabledLocales,
      defaultCurrency,
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 1800);
  };

  const publishLive = async (overrideConfig?: MenuLayoutConfig) => {
    await saveDraft(overrideConfig);
    await publishMutation.mutateAsync(selectedLocationId);
  };

  const selectedBranch = branches.find((b) => b.id === selectedLocationId);

  return {
    user,
    businessSlug,
    branches,
    selectedLocationId,
    setSelectedLocationId,
    selectedBranch,
    remoteProfile,
    isLoading,
    draftConfig,
    setDraftConfig,
    patchConfig,
    enabledLocales,
    setEnabledLocales,
    currencies,
    saved,
    isSaving: saveDraftMutation.isPending,
    isPublishing: publishMutation.isPending,
    saveDraft,
    publishLive,
  };
}
