import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { Switch } from '@voxire/ui';
import { Check, EyeOff, Star, Trash2 } from 'lucide-react';
import { Badge, Button, Select } from '../../shared/components/ui';
import {
  useDeleteMenuReview,
  useMenuReviews,
  useUpdateMenuReviewStatus,
} from '../../shared/hooks/queries/useMenuBuilder';
import { useMenuBranchProfile } from './useMenuBranchProfile';

const MenuReviewsPage: React.FC = () => {
  const { t } = useTranslation();
  const {
    branches,
    selectedLocationId,
    setSelectedLocationId,
    draftConfig,
    isSaving,
    isPublishing,
    publishLive,
  } = useMenuBranchProfile();

  const [filter, setFilter] = useState<'' | 'pending' | 'approved' | 'hidden'>('');
  const { data: allReviews = [], isLoading } = useMenuReviews(selectedLocationId);
  const updateStatus = useUpdateMenuReviewStatus();
  const deleteReview = useDeleteMenuReview();

  const filteredReviews = useMemo(
    () => (filter ? allReviews.filter((r) => r.status === filter) : allReviews),
    [allReviews, filter]
  );

  const stats = useMemo(() => {
    const approved = allReviews.filter((r) => r.status === 'approved');
    const pending = allReviews.filter((r) => r.status === 'pending');
    const avg =
      approved.length > 0
        ? approved.reduce((sum, r) => sum + (r.rating ?? 0), 0) / approved.length
        : 0;
    return {
      total: allReviews.length,
      pending: pending.length,
      approved: approved.length,
      average: avg,
    };
  }, [allReviews]);

  const handleToggleReviewsPage = (enabled: boolean) => {
    const nextConfig = {
      ...draftConfig,
      pages: { ...draftConfig.pages, reviews: enabled },
    };
    void publishLive(nextConfig);
  };

  return (
    <>
      <HelmetContainer title={`${t('common.nav.menuReviews')} · rtylr`} />
      <div className="w-full max-w-5xl space-y-6">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <Star size={18} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold">{t('common.nav.menuReviews')}</h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.pagesEditor.reviewsHint')}
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3">
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
            <div
              className="flex items-center gap-2 rounded-lg border px-3 py-1.5"
              style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
            >
              <label htmlFor="reviews-live-toggle" className="text-xs font-medium cursor-pointer">
                {t('design.menuApp.enableReviewsPage')}
              </label>
              <Switch
                id="reviews-live-toggle"
                disabled={isSaving || isPublishing}
                checked={Boolean(draftConfig.pages.reviews)}
                onCheckedChange={handleToggleReviewsPage}
              />
            </div>
          </div>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {[
            { label: t('design.menuApp.avgRating'), value: stats.average.toFixed(1) },
            { label: t('design.pagesEditor.filterPending'), value: String(stats.pending) },
            { label: t('design.pagesEditor.filterApproved'), value: String(stats.approved) },
            { label: t('design.pagesEditor.filterAll'), value: String(stats.total) },
          ].map((card) => (
            <div
              key={card.label}
              className="rounded-xl border px-4 py-3"
              style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
            >
              <div className="text-[11px] uppercase" style={{ color: 'var(--dash-text-muted)' }}>
                {card.label}
              </div>
              <div className="mt-1 text-xl font-semibold">{card.value}</div>
            </div>
          ))}
        </div>

        <div className="flex flex-wrap gap-1.5">
          {(
            [
              { id: '', label: t('design.pagesEditor.filterAll') },
              { id: 'pending', label: t('design.pagesEditor.filterPending') },
              { id: 'approved', label: t('design.pagesEditor.filterApproved') },
              { id: 'hidden', label: t('design.pagesEditor.filterHidden') },
            ] as const
          ).map((tab) => (
            <button
              key={tab.id || 'all'}
              type="button"
              onClick={() => setFilter(tab.id)}
              className="px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors"
              style={{
                background: filter === tab.id ? 'var(--dash-active)' : 'var(--dash-surface)',
                borderColor: 'var(--dash-border)',
                color: filter === tab.id ? 'var(--dash-text)' : 'var(--dash-text-muted)',
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {isLoading ? (
          <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
            {t('design.pagesEditor.loadingReviews')}
          </p>
        ) : filteredReviews.length === 0 ? (
          <div
            className="rounded-xl border p-8 text-center text-sm"
            style={{
              background: 'var(--dash-surface)',
              borderColor: 'var(--dash-border)',
              color: 'var(--dash-text-muted)',
            }}
          >
            {t('design.pagesEditor.noReviews')}
          </div>
        ) : (
          <div className="space-y-3">
            {filteredReviews.map((rev) => (
              <div
                key={rev.id}
                className="rounded-xl border p-4 space-y-2"
                style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-semibold">{rev.authorName || 'Guest'}</span>
                    <div className="flex items-center gap-0.5 text-amber-500">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <Star
                          key={i}
                          className="w-3.5 h-3.5"
                          fill={i < (rev.rating ?? 0) ? 'currentColor' : 'none'}
                        />
                      ))}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {rev.createdAt && (
                      <span className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                        {rev.createdAt.slice(0, 10)}
                      </span>
                    )}
                    <Badge
                      variant={
                        rev.status === 'approved'
                          ? 'success'
                          : rev.status === 'hidden'
                            ? 'secondary'
                            : 'warning'
                      }
                    >
                      {rev.status}
                    </Badge>
                  </div>
                </div>
                <p className="text-sm">{rev.comment}</p>
                <div className="flex items-center gap-2 pt-1">
                  {rev.status !== 'approved' && (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={updateStatus.isPending}
                      onClick={() =>
                        updateStatus.mutate({ id: String(rev.id), status: 'approved' })
                      }
                    >
                      <Check className="w-3.5 h-3.5 me-1" />
                      {t('design.pagesEditor.approve')}
                    </Button>
                  )}
                  {rev.status !== 'hidden' && (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={updateStatus.isPending}
                      onClick={() => updateStatus.mutate({ id: String(rev.id), status: 'hidden' })}
                    >
                      <EyeOff className="w-3.5 h-3.5 me-1" />
                      {t('design.pagesEditor.hide')}
                    </Button>
                  )}
                  <Button
                    size="sm"
                    variant="ghost"
                    className="ms-auto text-red-500"
                    disabled={deleteReview.isPending}
                    aria-label={t('design.pagesEditor.deleteReview')}
                    onClick={() => deleteReview.mutate(String(rev.id))}
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default MenuReviewsPage;
