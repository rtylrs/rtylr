import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Plus, Tag, Trash2 } from 'lucide-react';
import { Badge, Button, Input, Select } from '../../shared/components/ui';
import {
  useCreateMenuBanner,
  useDeleteMenuBanner,
  useMenuBanners,
} from '../../shared/hooks/queries/useMenu';
import { useStampCards, useUpsertStampCard } from '../../shared/hooks/queries/useMenuBuilder';
import { useMenuBranchProfile } from './useMenuBranchProfile';

const MenuPromosPage: React.FC = () => {
  const { t } = useTranslation();
  const { branches, selectedLocationId, setSelectedLocationId } = useMenuBranchProfile();

  const { data: banners = [] } = useMenuBanners();
  const createBanner = useCreateMenuBanner();
  const deleteBanner = useDeleteMenuBanner();

  const { data: stampCards = [] } = useStampCards(selectedLocationId);
  const upsertStampCard = useUpsertStampCard();

  const [bannerTitle, setBannerTitle] = useState('');
  const [bannerBody, setBannerBody] = useState('');
  const [stampTitle, setStampTitle] = useState('');
  const [stampTarget, setStampTarget] = useState('10');
  const [stampReward, setStampReward] = useState('');

  return (
    <>
      <HelmetContainer title={`${t('common.nav.menuPromos')} · rtylr`} />
      <div className="w-full max-w-5xl space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <Tag size={18} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold">{t('common.nav.menuPromos')}</h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.menuApp.promosSubtitle')}
            </p>
          </div>

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
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {/* Promo Banners */}
          <section
            className="rounded-xl border p-5 space-y-4"
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            <div>
              <h2 className="text-sm font-semibold">{t('design.studio.promoBanners')}</h2>
            </div>

            <div className="space-y-2">
              {banners.map((banner) => (
                <div
                  key={banner.id}
                  className="flex items-center justify-between gap-2 rounded-lg border px-3 py-2"
                  style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
                >
                  <div className="min-w-0">
                    <div className="text-sm font-medium truncate">{banner.title}</div>
                    {banner.body && (
                      <div className="text-xs truncate" style={{ color: 'var(--dash-text-muted)' }}>
                        {banner.body}
                      </div>
                    )}
                  </div>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="h-7 w-7 p-0 text-red-500 shrink-0"
                    disabled={deleteBanner.isPending}
                    onClick={() => deleteBanner.mutate(String(banner.id))}
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
                </div>
              ))}
            </div>

            <div className="space-y-2 pt-2">
              <Input
                value={bannerTitle}
                onChange={(e) => setBannerTitle(e.target.value)}
                placeholder={t('design.studio.promoBannerTitlePlaceholder')}
                className="h-9 text-sm"
              />
              <Input
                value={bannerBody}
                onChange={(e) => setBannerBody(e.target.value)}
                placeholder={t('design.studio.promoBannerBodyPlaceholder')}
                className="h-9 text-sm"
              />
              <Button
                size="sm"
                variant="outline"
                className="w-full"
                disabled={!bannerTitle.trim() || createBanner.isPending}
                onClick={() => {
                  createBanner.mutate({
                    title: bannerTitle.trim(),
                    body: bannerBody.trim(),
                    isActive: true,
                  });
                  setBannerTitle('');
                  setBannerBody('');
                }}
              >
                <Plus className="w-3.5 h-3.5 me-1.5" />
                {t('design.studio.addBanner')}
              </Button>
            </div>
          </section>

          {/* Digital Stamp Cards */}
          <section
            className="rounded-xl border p-5 space-y-4"
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            <div>
              <h2 className="text-sm font-semibold">{t('design.sections.loyalty')}</h2>
            </div>

            <div className="space-y-2">
              {stampCards.map((card) => (
                <div
                  key={card.id}
                  className="flex items-center justify-between gap-2 rounded-lg border px-3 py-2"
                  style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
                >
                  <div className="min-w-0">
                    <div className="text-sm font-medium truncate">{card.title}</div>
                    <div className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                      {card.stampsRequired} → {card.rewardLabel}
                    </div>
                  </div>
                  <Badge variant={card.isActive ? 'success' : 'secondary'}>
                    {card.isActive ? t('design.studio.liveBadge') : t('design.none')}
                  </Badge>
                </div>
              ))}
            </div>

            <div className="space-y-2 pt-2">
              <Input
                value={stampTitle}
                onChange={(e) => setStampTitle(e.target.value)}
                placeholder={t('design.studio.stampTitlePlaceholder')}
                className="h-9 text-sm"
              />
              <div className="grid grid-cols-3 gap-2">
                <Input
                  type="number"
                  min={1}
                  value={stampTarget}
                  onChange={(e) => setStampTarget(e.target.value)}
                  placeholder="10"
                  className="h-9 text-sm"
                />
                <Input
                  value={stampReward}
                  onChange={(e) => setStampReward(e.target.value)}
                  placeholder={t('design.studio.stampRewardPlaceholder')}
                  className="h-9 text-sm col-span-2"
                />
              </div>
              <Button
                size="sm"
                variant="outline"
                className="w-full"
                disabled={!stampTitle.trim() || !stampReward.trim() || upsertStampCard.isPending}
                onClick={() => {
                  upsertStampCard.mutate({
                    locationId: selectedLocationId,
                    title: stampTitle.trim(),
                    stampsRequired: Math.max(1, Number(stampTarget) || 10),
                    rewardLabel: stampReward.trim(),
                  });
                  setStampTitle('');
                  setStampReward('');
                }}
              >
                <Plus className="w-3.5 h-3.5 me-1.5" />
                {t('design.studio.addStampCard')}
              </Button>
            </div>
          </section>
        </div>
      </div>
    </>
  );
};

export default MenuPromosPage;
