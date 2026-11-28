import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Switch } from '@voxire/ui';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import { Check, Plus, Settings, Sparkles, Trash2 } from 'lucide-react';
import { Badge, Button, Input, Label, Select } from '../../shared/components/ui';
import {
  useDeleteMenuCurrency,
  useMenuCatalogItems,
  useTranslateMenu,
  useUpsertMenuCurrency,
} from '../../shared/hooks/queries/useMenuBuilder';
import { useMenuBranchProfile } from './useMenuBranchProfile';

interface SettingSwitchProps {
  id: string;
  label: string;
  hint?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}

const SettingSwitch: React.FC<SettingSwitchProps> = ({ id, label, hint, checked, onChange }) => (
  <div className="flex items-center justify-between gap-3 py-1.5">
    <div className="min-w-0">
      <label htmlFor={id} className="block text-sm font-medium cursor-pointer select-none">
        {label}
      </label>
      {hint && (
        <span className="block text-xs" style={{ color: 'var(--dash-text-muted)' }}>
          {hint}
        </span>
      )}
    </div>
    <Switch id={id} checked={checked} onCheckedChange={onChange} />
  </div>
);

const MenuSettingsPage: React.FC = () => {
  const { t } = useTranslation();
  const platformOrderingEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_ORDERING);
  const platformTableActionsEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_TABLE_ACTIONS);
  const platformCrmCaptureEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CRM_CAPTURE);
  const platformCurrenciesEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_CURRENCIES);
  const platformAiTranslateEnabled = useFeatureFlag(FLAG_KEYS.APP_MENU_AI_TRANSLATE);

  const {
    branches,
    selectedLocationId,
    setSelectedLocationId,
    draftConfig,
    patchConfig,
    enabledLocales,
    setEnabledLocales,
    currencies,
    saved,
    isSaving,
    isPublishing,
    saveDraft,
    publishLive,
  } = useMenuBranchProfile();

  const { data: catalogItems = [] } = useMenuCatalogItems();
  const upsertCurrency = useUpsertMenuCurrency();
  const deleteCurrency = useDeleteMenuCurrency();
  const translateMenu = useTranslateMenu();

  const [newCurCode, setNewCurCode] = useState('');
  const [newCurSymbol, setNewCurSymbol] = useState('');
  const [newCurRate, setNewCurRate] = useState('1');

  const features = draftConfig.features!;
  const paymentMode = draftConfig.payment?.mode ?? 'pos';
  const paymentRedirectUrl = draftConfig.payment?.redirectUrl ?? '';

  const patchFeatures = (patch: Partial<typeof features>) =>
    patchConfig({ features: { ...features, ...patch } });

  const toggleLocale = (code: string) => {
    setEnabledLocales((prev) => {
      if (prev.includes(code)) {
        if (prev.length === 1) return prev;
        return prev.filter((l) => l !== code);
      }
      return [...prev, code];
    });
  };

  const handleAiTranslate = () => {
    const entries: Array<{
      entityType: string;
      entityId: string;
      field: string;
      sourceText: string;
    }> = [];
    for (const item of catalogItems) {
      const id = String(item.id ?? '');
      if (item.name?.trim()) {
        entries.push({ entityType: 'item', entityId: id, field: 'name', sourceText: item.name });
      }
      if (item.description?.trim()) {
        entries.push({
          entityType: 'item',
          entityId: id,
          field: 'description',
          sourceText: item.description,
        });
      }
    }
    translateMenu.mutate({
      locationId: selectedLocationId,
      targetLocales: enabledLocales,
      entries,
    });
  };

  return (
    <>
      <HelmetContainer title={`${t('common.nav.menuSettings')} · rtylr`} />
      <div className="w-full max-w-4xl space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <Settings size={18} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold">{t('common.nav.menuSettings')}</h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.studio.featuresSubtitle')}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            {branches.length > 0 && (
              <Select
                aria-label={t('design.studio.branchLabel')}
                value={String(selectedLocationId)}
                onChange={(e) => setSelectedLocationId(Number(e.target.value) || 0)}
                className="h-8 text-xs w-44"
              >
                <option value="0">{t('design.studio.defaultBranch')}</option>
                {branches.map((branch) => (
                  <option key={branch.id} value={String(branch.id)}>
                    {branch.name}
                  </option>
                ))}
              </Select>
            )}
            <Button
              size="sm"
              variant="outline"
              disabled={isSaving}
              onClick={() => void saveDraft()}
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
              size="sm"
              variant="accent"
              disabled={isSaving || isPublishing}
              onClick={() => void publishLive()}
            >
              {t('design.studio.publish')}
            </Button>
          </div>
        </div>

        <div className="grid gap-6">
          {/* Ordering & Fulfillment */}
          {platformOrderingEnabled && (
            <section
              className="rounded-xl border p-5 space-y-2"
              style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
            >
              <h2 className="text-sm font-semibold mb-2">{t('design.sections.features')}</h2>
              <SettingSwitch
                id="ms-ordering"
                label={t('design.studio.orderingEnabled')}
                checked={Boolean(features.enableOrdering)}
                onChange={(enableOrdering) => patchFeatures({ enableOrdering })}
              />
              <SettingSwitch
                id="ms-dinein"
                label={t('design.studio.modeDineIn')}
                checked={Boolean(features.fulfillmentModes?.dineIn)}
                onChange={(dineIn) =>
                  patchFeatures({
                    fulfillmentModes: { ...features.fulfillmentModes!, dineIn },
                  })
                }
              />
              <SettingSwitch
                id="ms-self"
                label={t('design.studio.modeSelfService')}
                checked={Boolean(features.fulfillmentModes?.selfService)}
                onChange={(selfService) =>
                  patchFeatures({
                    fulfillmentModes: { ...features.fulfillmentModes!, selfService },
                  })
                }
              />
              <SettingSwitch
                id="ms-pickup"
                label={t('design.studio.modePickup')}
                checked={Boolean(features.fulfillmentModes?.pickup)}
                onChange={(pickup) =>
                  patchFeatures({
                    fulfillmentModes: { ...features.fulfillmentModes!, pickup },
                  })
                }
              />
              <SettingSwitch
                id="ms-delivery"
                label={t('design.studio.modeDelivery')}
                checked={Boolean(features.fulfillmentModes?.delivery)}
                onChange={(delivery) =>
                  patchFeatures({
                    fulfillmentModes: { ...features.fulfillmentModes!, delivery },
                  })
                }
              />
            </section>
          )}

          {/* Table Service & Guest Tools */}
          <section
            className="rounded-xl border p-5 space-y-2"
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            <h2 className="text-sm font-semibold mb-2">{t('design.studio.tableService')}</h2>
            {platformTableActionsEnabled && (
              <>
                <SettingSwitch
                  id="ms-call"
                  label={t('design.studio.callStaff')}
                  checked={Boolean(features.tableActions?.callWaiter)}
                  onChange={(callWaiter) =>
                    patchFeatures({ tableActions: { ...features.tableActions!, callWaiter } })
                  }
                />
                <SettingSwitch
                  id="ms-bill"
                  label={t('design.studio.requestBill')}
                  checked={Boolean(features.tableActions?.requestBill)}
                  onChange={(requestBill) =>
                    patchFeatures({ tableActions: { ...features.tableActions!, requestBill } })
                  }
                />
                <SettingSwitch
                  id="ms-split"
                  label={t('design.studio.splitBill')}
                  checked={Boolean(features.tableActions?.splitBill)}
                  onChange={(splitBill) =>
                    patchFeatures({ tableActions: { ...features.tableActions!, splitBill } })
                  }
                />
              </>
            )}
            <SettingSwitch
              id="ms-search"
              label={t('design.studio.searchAndFilters')}
              checked={Boolean(features.enableSearch)}
              onChange={(enableSearch) =>
                patchFeatures({ enableSearch, enableDietaryFilters: enableSearch })
              }
            />
            {platformCrmCaptureEnabled && (
              <SettingSwitch
                id="ms-crm"
                label={t('design.studio.crmCapture')}
                hint={t('design.studio.crmCaptureHint')}
                checked={Boolean(features.enableCrmCapture)}
                onChange={(enableCrmCapture) => patchFeatures({ enableCrmCapture })}
              />
            )}
          </section>

          {/* Currencies */}
          {platformCurrenciesEnabled && (
            <section
              className="rounded-xl border p-5 space-y-3"
              style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
            >
              <h2 className="text-sm font-semibold">{t('design.sections.currencies')}</h2>
              <div className="space-y-2">
                {currencies.map((cur) => (
                  <div
                    key={cur.code}
                    className="flex items-center justify-between rounded-lg border px-3 py-2 text-sm"
                    style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
                  >
                    <div>
                      <span className="font-semibold">{cur.code}</span>{' '}
                      <span style={{ color: 'var(--dash-text-muted)' }}>
                        ({cur.symbol}) · ×{cur.exchangeRate}
                      </span>
                    </div>
                    {cur.isDefault ? (
                      <Badge variant="secondary">{t('design.studio.defaultTag')}</Badge>
                    ) : (
                      <Button
                        size="sm"
                        variant="ghost"
                        className="h-7 w-7 p-0 text-red-500"
                        onClick={() =>
                          deleteCurrency.mutate({
                            locationId: selectedLocationId,
                            code: cur.code ?? '',
                          })
                        }
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
              <div className="grid grid-cols-4 gap-2 pt-1">
                <Input
                  value={newCurCode}
                  onChange={(e) => setNewCurCode(e.target.value.toUpperCase())}
                  placeholder="EUR"
                  className="h-9 text-sm"
                />
                <Input
                  value={newCurSymbol}
                  onChange={(e) => setNewCurSymbol(e.target.value)}
                  placeholder="€"
                  className="h-9 text-sm"
                />
                <Input
                  type="number"
                  step="0.01"
                  value={newCurRate}
                  onChange={(e) => setNewCurRate(e.target.value)}
                  placeholder="0.92"
                  className="h-9 text-sm"
                />
                <Button
                  size="sm"
                  variant="outline"
                  disabled={!newCurCode.trim() || !newCurSymbol.trim() || upsertCurrency.isPending}
                  onClick={() => {
                    upsertCurrency.mutate({
                      locationId: selectedLocationId,
                      code: newCurCode.trim(),
                      symbol: newCurSymbol.trim(),
                      exchangeRate: Math.max(0.0001, Number(newCurRate) || 1),
                    });
                    setNewCurCode('');
                    setNewCurSymbol('');
                    setNewCurRate('1');
                  }}
                >
                  <Plus className="w-3.5 h-3.5 me-1" />
                  {t('design.studio.addCurrency')}
                </Button>
              </div>
            </section>
          )}

          {/* Languages & AI Translation */}
          <section
            className="rounded-xl border p-5 space-y-3"
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            <h2 className="text-sm font-semibold">{t('design.sections.languages')}</h2>
            <div className="flex flex-wrap gap-2">
              {(['en', 'fr', 'ar'] as const).map((loc) => {
                const active = enabledLocales.includes(loc);
                return (
                  <button
                    key={loc}
                    type="button"
                    onClick={() => toggleLocale(loc)}
                    className="px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors"
                    style={{
                      background: active ? 'var(--dash-active)' : 'transparent',
                      borderColor: 'var(--dash-border)',
                      color: active ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                    }}
                  >
                    {loc.toUpperCase()}
                  </button>
                );
              })}
            </div>
            {platformAiTranslateEnabled && (
              <Button
                size="sm"
                variant="outline"
                disabled={translateMenu.isPending || catalogItems.length === 0}
                onClick={handleAiTranslate}
              >
                <Sparkles className="w-3.5 h-3.5 me-1.5" />
                {t('design.studio.aiTranslate')}
              </Button>
            )}
          </section>

          {/* Payment Handoff */}
          <section
            className="rounded-xl border p-5 space-y-3"
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            <h2 className="text-sm font-semibold">{t('design.sections.payment')}</h2>
            <div className="flex gap-2">
              {(
                [
                  { id: 'pos', label: t('design.studio.payAtPos') },
                  { id: 'redirect', label: t('design.studio.payExternal') },
                ] as const
              ).map((opt) => (
                <button
                  key={opt.id}
                  type="button"
                  onClick={() =>
                    patchConfig({
                      payment: { mode: opt.id, redirectUrl: paymentRedirectUrl },
                    })
                  }
                  className="px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors"
                  style={{
                    background: paymentMode === opt.id ? 'var(--dash-active)' : 'transparent',
                    borderColor: 'var(--dash-border)',
                    color: paymentMode === opt.id ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                  }}
                >
                  {opt.label}
                </button>
              ))}
            </div>
            {paymentMode === 'redirect' && (
              <div className="space-y-1">
                <Label className="text-xs">Checkout redirect URL</Label>
                <Input
                  value={paymentRedirectUrl}
                  onChange={(e) =>
                    patchConfig({ payment: { mode: 'redirect', redirectUrl: e.target.value } })
                  }
                  placeholder="https://pay.example.com/checkout"
                  className="h-9 text-sm"
                />
              </div>
            )}
          </section>
        </div>
      </div>
    </>
  );
};

export default MenuSettingsPage;
