import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { buildPublicMenuUrl } from '@voxire/config';
import { toast } from 'sonner';
import {
  BarChart3,
  Copy,
  ExternalLink,
  FileText,
  Palette,
  Plus,
  QrCode,
  Settings,
  Star,
  Tag,
  UtensilsCrossed,
} from 'lucide-react';
import { Badge, Button, Select } from '../../shared/components/ui';
import {
  useMenuAnalyticsSummary,
  useMenuCatalogItems,
  useMenuReviews,
} from '../../shared/hooks/queries/useMenuBuilder';
import { useMenuBranchProfile } from './useMenuBranchProfile';

const MenuOverviewPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const {
    user,
    businessSlug,
    branches,
    selectedLocationId,
    setSelectedLocationId,
    selectedBranch,
    remoteProfile,
  } = useMenuBranchProfile();

  const { data: catalogItems = [] } = useMenuCatalogItems();
  const { data: reviews = [] } = useMenuReviews(selectedLocationId);
  const { data: analytics } = useMenuAnalyticsSummary(selectedLocationId, 30);

  const pendingReviews = useMemo(
    () => reviews.filter((r) => r.status === 'pending').length,
    [reviews]
  );
  const activeItemsCount = useMemo(
    () => catalogItems.filter((i) => i.isActive !== false).length,
    [catalogItems]
  );

  const publicUrl = useMemo(() => {
    const slug = businessSlug.trim();
    if (slug) {
      const branchSlug = selectedBranch?.slug || (branches.length === 1 ? branches[0].slug : '');
      return buildPublicMenuUrl(slug, branchSlug || undefined);
    }
    const bizId = user?.businessId ?? user?.id;
    return bizId ? `${window.location.origin}/b/${bizId}/menu` : '';
  }, [businessSlug, selectedBranch, branches, user]);

  const handleCopyUrl = () => {
    if (!publicUrl) return;
    void navigator.clipboard?.writeText(publicUrl);
    toast.success(t('design.menuApp.linkCopied'));
  };

  const quickCards = [
    {
      path: '/menus/items',
      icon: UtensilsCrossed,
      title: t('common.nav.menuItems'),
      description: t('design.menuApp.cardItemsDesc'),
    },
    {
      path: '/menus/studio',
      icon: Palette,
      title: t('common.nav.menuStudio'),
      description: t('design.menuApp.cardStudioDesc'),
    },
    {
      path: '/menus/pages',
      icon: FileText,
      title: t('common.nav.menuPages'),
      description: t('design.menuApp.cardPagesDesc'),
    },
    {
      path: '/menus/reviews',
      icon: Star,
      title: t('common.nav.menuReviews'),
      description: t('design.menuApp.cardReviewsDesc'),
      badge: pendingReviews > 0 ? String(pendingReviews) : undefined,
    },
    {
      path: '/menus/promos',
      icon: Tag,
      title: t('common.nav.menuPromos'),
      description: t('design.menuApp.cardPromosDesc'),
    },
    {
      path: '/menus/qr-codes',
      icon: QrCode,
      title: t('common.nav.qrCodes'),
      description: t('design.menuApp.cardQrDesc'),
    },
    {
      path: '/menus/analytics',
      icon: BarChart3,
      title: t('common.nav.menuAnalytics'),
      description: t('design.menuApp.cardAnalyticsDesc'),
    },
    {
      path: '/menus/settings',
      icon: Settings,
      title: t('common.nav.menuSettings'),
      description: t('design.menuApp.cardSettingsDesc'),
    },
  ];

  return (
    <>
      <HelmetContainer title={`${t('appNames.short.menu')} · rtylr`} />
      <div className="w-full max-w-6xl space-y-6">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <UtensilsCrossed size={18} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold">{t('common.nav.menuOverview')}</h1>
              <Badge variant={remoteProfile?.publishedAt ? 'success' : 'secondary'}>
                {remoteProfile?.publishedAt ? t('design.studio.liveBadge') : 'Draft'}
              </Badge>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.menuApp.overviewSubtitle')}
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
              onClick={() => void navigate({ to: '/menus/studio' })}
            >
              <Palette className="w-3.5 h-3.5 me-1.5" />
              {t('common.nav.menuStudio')}
            </Button>
            <Button size="sm" onClick={() => void navigate({ to: '/create-item' })}>
              <Plus className="w-3.5 h-3.5 me-1" />
              {t('design.menuApp.addItem')}
            </Button>
          </div>
        </div>

        {/* Live Storefront Link Banner */}
        {publicUrl && (
          <div
            className="rounded-xl border p-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
            style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
          >
            <div className="min-w-0">
              <div
                className="text-[11px] uppercase font-medium"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                {t('design.menuApp.liveStorefrontUrl')}
              </div>
              <p className="text-sm font-mono font-medium truncate mt-0.5">{publicUrl}</p>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Button size="sm" variant="outline" onClick={handleCopyUrl}>
                <Copy className="w-3.5 h-3.5 me-1.5" />
                {t('design.studio.copyAction')}
              </Button>
              <a
                href={publicUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 h-8 px-3 rounded-lg text-xs font-semibold border transition-colors"
                style={{
                  background: 'var(--dash-card)',
                  borderColor: 'var(--dash-border)',
                  color: 'var(--dash-text)',
                }}
              >
                <ExternalLink className="w-3.5 h-3.5" />
                {t('design.menuApp.openLiveMenu')}
              </a>
            </div>
          </div>
        )}

        {/* KPI Row */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
          {[
            { label: t('design.menuApp.activeItems'), value: activeItemsCount },
            { label: t('design.studio.metricScans'), value: analytics?.qrScans ?? 0 },
            { label: t('design.studio.metricViews'), value: analytics?.menuViews ?? 0 },
            {
              label: t('design.menuApp.pendingReviews'),
              value: pendingReviews,
              onClick: () => void navigate({ to: '/menus/reviews' }),
            },
          ].map((stat) => (
            <div
              key={stat.label}
              onClick={stat.onClick}
              role={stat.onClick ? 'button' : undefined}
              tabIndex={stat.onClick ? 0 : undefined}
              onKeyDown={
                stat.onClick
                  ? (e) => {
                      if (e.key === 'Enter' || e.key === ' ') stat.onClick?.();
                    }
                  : undefined
              }
              className={`rounded-xl border px-4 py-3.5 ${stat.onClick ? 'cursor-pointer hover:bg-[var(--dash-hover)] transition-colors' : ''}`}
              style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
            >
              <div
                className="text-[11px] uppercase font-medium"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                {stat.label}
              </div>
              <div className="mt-1 text-2xl font-semibold">{stat.value}</div>
            </div>
          ))}
        </div>

        {/* Quick Navigation Grid */}
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {quickCards.map((card) => {
            const Icon = card.icon;
            return (
              <button
                key={card.path}
                type="button"
                onClick={() => void navigate({ to: card.path })}
                className="text-start rounded-xl border p-4 space-y-2 transition-colors hover:bg-[var(--dash-hover)]"
                style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
              >
                <div className="flex items-center justify-between gap-2">
                  <Icon size={18} style={{ color: 'var(--dash-text-muted)' }} />
                  {card.badge && <Badge variant="warning">{card.badge}</Badge>}
                </div>
                <div className="text-sm font-semibold">{card.title}</div>
                <p className="text-xs leading-relaxed" style={{ color: 'var(--dash-text-muted)' }}>
                  {card.description}
                </p>
              </button>
            );
          })}
        </div>
      </div>
    </>
  );
};

export default MenuOverviewPage;
