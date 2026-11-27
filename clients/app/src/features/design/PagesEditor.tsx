import React, { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { uploadAsset } from '@voxire/api';
import { cdnImageUrl } from '@voxire/config';
import { Switch } from '@voxire/ui';
import type {
  AboutPageContent,
  BusinessOffer,
  MenuPageContentConfig,
  PageConfig,
} from '@voxire/menu';
import { toast } from 'sonner';
import { ArrowLeft, ArrowRight, Check, EyeOff, Plus, Star, Trash2, Upload } from 'lucide-react';
import { Badge, Button, Input, Label, Separator } from '../../shared/components/ui';
import {
  useDeleteMenuReview,
  useMenuReviews,
  useUpdateMenuReviewStatus,
} from '../../shared/hooks/queries/useMenuBuilder';

interface PagesEditorProps {
  selectedLocationId: number;
  pages: PageConfig;
  content: MenuPageContentConfig;
  defaultAddress?: string;
  defaultPhone?: string;
  onPatchPages: (patch: Partial<PageConfig>) => void;
  onPatchContent: (patch: Partial<MenuPageContentConfig>) => void;
}

const PagesEditor: React.FC<PagesEditorProps> = ({
  selectedLocationId,
  pages,
  content,
  defaultAddress,
  defaultPhone,
  onPatchPages,
  onPatchContent,
}) => {
  const { t } = useTranslation();
  const about = content.about ?? {};
  const gallery = content.gallery ?? [];
  const offers = content.offers ?? [];
  const includeBannersInOffers = content.includeBannersInOffers !== false;

  const [galleryUrlInput, setGalleryUrlInput] = useState('');
  const [uploadingGallery, setUploadingGallery] = useState(false);
  const galleryFileRef = useRef<HTMLInputElement>(null);

  const [offerTitle, setOfferTitle] = useState('');
  const [offerDesc, setOfferDesc] = useState('');
  const [offerBadge, setOfferBadge] = useState('');
  const [offerValidUntil, setOfferValidUntil] = useState('');
  const [offerImage, setOfferImage] = useState('');
  const [uploadingOfferImage, setUploadingOfferImage] = useState(false);
  const offerFileRef = useRef<HTMLInputElement>(null);

  const [reviewFilter, setReviewFilter] = useState<'' | 'pending' | 'approved' | 'hidden'>('');
  const { data: reviews = [], isLoading: reviewsLoading } = useMenuReviews(
    selectedLocationId,
    reviewFilter || undefined
  );
  const updateReviewStatus = useUpdateMenuReviewStatus();
  const deleteReview = useDeleteMenuReview();

  const patchAbout = (patch: Partial<AboutPageContent>) =>
    onPatchContent({
      about: {
        ...about,
        ...patch,
      },
    });

  const patchSocials = (patch: Partial<NonNullable<AboutPageContent['socials']>>) =>
    patchAbout({
      socials: {
        ...about.socials,
        ...patch,
      },
    });

  const handleGalleryUpload = async (file: File | undefined) => {
    if (!file) return;
    setUploadingGallery(true);
    try {
      const res = await uploadAsset(file, { visibility: 'public' });
      const url = res.url || '';
      if (!url) throw new Error('Missing uploaded URL');
      onPatchContent({ gallery: [...gallery, url] });
      toast.success(t('design.studio.uploadSuccess'));
    } catch (err: unknown) {
      const msg = (err as { message?: string })?.message || t('design.studio.uploadFailed');
      toast.error(msg);
    } finally {
      setUploadingGallery(false);
    }
  };

  const handleAddGalleryUrl = () => {
    const trimmed = galleryUrlInput.trim();
    if (!trimmed) return;
    onPatchContent({ gallery: [...gallery, trimmed] });
    setGalleryUrlInput('');
  };

  const moveGalleryPhoto = (index: number, delta: -1 | 1) => {
    const target = index + delta;
    if (target < 0 || target >= gallery.length) return;
    const next = [...gallery];
    const [moved] = next.splice(index, 1);
    next.splice(target, 0, moved);
    onPatchContent({ gallery: next });
  };

  const removeGalleryPhoto = (index: number) => {
    onPatchContent({ gallery: gallery.filter((_, i) => i !== index) });
  };

  const handleOfferImageUpload = async (file: File | undefined) => {
    if (!file) return;
    setUploadingOfferImage(true);
    try {
      const res = await uploadAsset(file, { visibility: 'public' });
      const url = res.url || '';
      if (!url) throw new Error('Missing uploaded URL');
      setOfferImage(url);
      toast.success(t('design.studio.uploadSuccess'));
    } catch (err: unknown) {
      const msg = (err as { message?: string })?.message || t('design.studio.uploadFailed');
      toast.error(msg);
    } finally {
      setUploadingOfferImage(false);
    }
  };

  const handleAddOffer = () => {
    const title = offerTitle.trim();
    if (!title) return;
    const nextOffer: BusinessOffer = {
      id: `offer-${Date.now()}`,
      title,
      description: offerDesc.trim(),
      ...(offerBadge.trim() ? { badge: offerBadge.trim() } : {}),
      ...(offerValidUntil.trim() ? { validUntil: offerValidUntil.trim() } : {}),
      ...(offerImage.trim() ? { image: offerImage.trim() } : {}),
    };
    onPatchContent({ offers: [...offers, nextOffer] });
    setOfferTitle('');
    setOfferDesc('');
    setOfferBadge('');
    setOfferValidUntil('');
    setOfferImage('');
  };

  const removeOffer = (id: string) => {
    onPatchContent({ offers: offers.filter((o) => o.id !== id) });
  };

  return (
    <div className="space-y-5">
      <div>
        <h2 className="text-sm font-semibold">{t('design.sections.pages')}</h2>
        <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
          {t('design.pagesEditor.subtitle')}
        </p>
      </div>

      {/* 1. ABOUT PAGE */}
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-2">
          <Label
            htmlFor="page-about-toggle"
            className="text-[11px] uppercase tracking-wider cursor-pointer"
          >
            {t('design.pages.about')}
          </Label>
          <Switch
            id="page-about-toggle"
            checked={Boolean(pages.about)}
            onCheckedChange={(checked) => onPatchPages({ about: checked })}
          />
        </div>

        {pages.about && (
          <div className="space-y-2.5">
            <p className="text-[11px]" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.pagesEditor.aboutHint')}
            </p>
            <div className="space-y-1">
              <Label className="text-xs">{t('design.pagesEditor.story')}</Label>
              <textarea
                rows={3}
                value={about.story ?? ''}
                onChange={(e) => patchAbout({ story: e.target.value })}
                placeholder={t('design.pagesEditor.storyPlaceholder')}
                className="w-full rounded-lg p-2.5 text-xs border focus:outline-none focus:ring-2 focus:ring-[var(--dash-border-focus)]"
                style={{
                  background: 'var(--dash-input-bg)',
                  color: 'var(--dash-text)',
                  borderColor: 'var(--dash-border)',
                }}
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">{t('design.pagesEditor.hours')}</Label>
              <Input
                value={about.hours ?? ''}
                onChange={(e) => patchAbout({ hours: e.target.value })}
                placeholder={t('design.pagesEditor.hoursPlaceholder')}
                className="h-8 text-xs"
              />
            </div>
            <div className="space-y-1">
              <Label className="text-xs">{t('design.pagesEditor.address')}</Label>
              <Input
                value={about.address ?? ''}
                onChange={(e) => patchAbout({ address: e.target.value })}
                placeholder={defaultAddress || t('design.pagesEditor.addressPlaceholder')}
                className="h-8 text-xs"
              />
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div className="space-y-1">
                <Label className="text-xs">{t('design.pagesEditor.phone')}</Label>
                <Input
                  value={about.phone ?? ''}
                  onChange={(e) => patchAbout({ phone: e.target.value })}
                  placeholder={defaultPhone || t('design.pagesEditor.phonePlaceholder')}
                  className="h-8 text-xs"
                />
              </div>
              <div className="space-y-1">
                <Label className="text-xs">{t('design.pagesEditor.email')}</Label>
                <Input
                  type="email"
                  value={about.email ?? ''}
                  onChange={(e) => patchAbout({ email: e.target.value })}
                  placeholder="hello@example.com"
                  className="h-8 text-xs"
                />
              </div>
            </div>
            <div className="space-y-1">
              <Label className="text-xs">{t('design.pagesEditor.website')}</Label>
              <Input
                value={about.website ?? ''}
                onChange={(e) => patchAbout({ website: e.target.value })}
                placeholder="https://example.com"
                className="h-8 text-xs"
              />
            </div>
            <div className="grid grid-cols-3 gap-1.5">
              <Input
                value={about.socials?.instagram ?? ''}
                onChange={(e) => patchSocials({ instagram: e.target.value })}
                placeholder="Instagram"
                aria-label="Instagram"
                className="h-8 text-xs"
              />
              <Input
                value={about.socials?.facebook ?? ''}
                onChange={(e) => patchSocials({ facebook: e.target.value })}
                placeholder="Facebook"
                aria-label="Facebook"
                className="h-8 text-xs"
              />
              <Input
                value={about.socials?.tiktok ?? ''}
                onChange={(e) => patchSocials({ tiktok: e.target.value })}
                placeholder="TikTok"
                aria-label="TikTok"
                className="h-8 text-xs"
              />
            </div>
          </div>
        )}
      </div>

      <Separator />

      {/* 2. GALLERY PAGE */}
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-2">
          <Label
            htmlFor="page-gallery-toggle"
            className="text-[11px] uppercase tracking-wider cursor-pointer"
          >
            {t('design.pages.gallery')} ({gallery.length})
          </Label>
          <Switch
            id="page-gallery-toggle"
            checked={Boolean(pages.gallery)}
            onCheckedChange={(checked) => onPatchPages({ gallery: checked })}
          />
        </div>

        {pages.gallery && (
          <div className="space-y-2.5">
            <input
              ref={galleryFileRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(e) => {
                void handleGalleryUpload(e.target.files?.[0]);
                e.target.value = '';
              }}
            />
            <Button
              variant="outline"
              size="sm"
              className="w-full"
              disabled={uploadingGallery}
              onClick={() => galleryFileRef.current?.click()}
            >
              <Upload className="w-3.5 h-3.5 me-1.5" />
              {uploadingGallery
                ? t('design.studio.uploading')
                : t('design.pagesEditor.uploadPhoto')}
            </Button>
            <div className="flex gap-1.5">
              <Input
                value={galleryUrlInput}
                onChange={(e) => setGalleryUrlInput(e.target.value)}
                placeholder={t('design.studio.imageUrlPlaceholder')}
                className="h-8 text-xs flex-1"
              />
              <Button
                variant="outline"
                size="sm"
                className="h-8 px-2.5"
                disabled={!galleryUrlInput.trim()}
                onClick={handleAddGalleryUrl}
              >
                <Plus className="w-3.5 h-3.5" />
              </Button>
            </div>
            {gallery.length > 0 && (
              <div className="grid grid-cols-2 gap-2">
                {gallery.map((src, idx) => (
                  <div
                    key={`${src}-${idx}`}
                    className="rounded-lg border overflow-hidden"
                    style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-card)' }}
                  >
                    <img
                      src={cdnImageUrl(src) || src}
                      alt=""
                      className="w-full h-20 object-cover block"
                    />
                    <div className="flex items-center justify-between p-1 gap-1">
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0"
                          disabled={idx === 0}
                          aria-label={t('design.pagesEditor.movePhotoUp')}
                          onClick={() => moveGalleryPhoto(idx, -1)}
                        >
                          <ArrowLeft className="w-3 h-3" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0"
                          disabled={idx === gallery.length - 1}
                          aria-label={t('design.pagesEditor.movePhotoDown')}
                          onClick={() => moveGalleryPhoto(idx, 1)}
                        >
                          <ArrowRight className="w-3 h-3" />
                        </Button>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0 text-red-500"
                        aria-label={t('design.pagesEditor.removePhoto')}
                        onClick={() => removeGalleryPhoto(idx)}
                      >
                        <Trash2 className="w-3 h-3" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      <Separator />

      {/* 3. OFFERS PAGE */}
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-2">
          <Label
            htmlFor="page-offers-toggle"
            className="text-[11px] uppercase tracking-wider cursor-pointer"
          >
            {t('design.pages.offers')} ({offers.length})
          </Label>
          <Switch
            id="page-offers-toggle"
            checked={Boolean(pages.offers)}
            onCheckedChange={(checked) => onPatchPages({ offers: checked })}
          />
        </div>

        {pages.offers && (
          <div className="space-y-2.5">
            <div className="flex items-center justify-between gap-2 py-1">
              <label
                htmlFor="offers-include-banners"
                className="text-xs cursor-pointer select-none"
                style={{ color: 'var(--dash-text)' }}
              >
                {t('design.pagesEditor.includeBanners')}
              </label>
              <Switch
                id="offers-include-banners"
                checked={includeBannersInOffers}
                onCheckedChange={(checked) => onPatchContent({ includeBannersInOffers: checked })}
              />
            </div>

            {offers.length > 0 && (
              <div className="space-y-1.5">
                {offers.map((offer) => (
                  <div
                    key={offer.id}
                    className="flex items-start justify-between gap-2 rounded-lg border px-2.5 py-2"
                    style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-card)' }}
                  >
                    <div className="min-w-0">
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <span className="text-xs font-semibold truncate">{offer.title}</span>
                        {offer.badge && <Badge variant="secondary">{offer.badge}</Badge>}
                      </div>
                      {offer.description && (
                        <p
                          className="text-[11px] mt-0.5 line-clamp-2"
                          style={{ color: 'var(--dash-text-muted)' }}
                        >
                          {offer.description}
                        </p>
                      )}
                      {offer.validUntil && (
                        <span className="text-[10px]" style={{ color: 'var(--dash-text-muted)' }}>
                          {offer.validUntil}
                        </span>
                      )}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0 shrink-0 text-red-500"
                      aria-label={t('design.pagesEditor.removeOffer')}
                      onClick={() => removeOffer(offer.id)}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                ))}
              </div>
            )}

            <div className="space-y-1.5 pt-1">
              <Input
                value={offerTitle}
                onChange={(e) => setOfferTitle(e.target.value)}
                placeholder={t('design.pagesEditor.offerTitlePlaceholder')}
                className="h-8 text-xs"
              />
              <Input
                value={offerDesc}
                onChange={(e) => setOfferDesc(e.target.value)}
                placeholder={t('design.pagesEditor.offerDescPlaceholder')}
                className="h-8 text-xs"
              />
              <div className="grid grid-cols-2 gap-1.5">
                <Input
                  value={offerBadge}
                  onChange={(e) => setOfferBadge(e.target.value)}
                  placeholder={t('design.pagesEditor.offerBadgePlaceholder')}
                  className="h-8 text-xs"
                />
                <Input
                  type="date"
                  value={offerValidUntil}
                  onChange={(e) => setOfferValidUntil(e.target.value)}
                  aria-label={t('design.pagesEditor.offerValidUntil')}
                  className="h-8 text-xs"
                />
              </div>
              <input
                ref={offerFileRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={(e) => {
                  void handleOfferImageUpload(e.target.files?.[0]);
                  e.target.value = '';
                }}
              />
              <div className="flex gap-1.5">
                <Input
                  value={offerImage}
                  onChange={(e) => setOfferImage(e.target.value)}
                  placeholder={t('design.pagesEditor.offerImagePlaceholder')}
                  className="h-8 text-xs flex-1"
                />
                <Button
                  variant="outline"
                  size="sm"
                  className="h-8 px-2"
                  disabled={uploadingOfferImage}
                  aria-label={t('design.studio.upload')}
                  onClick={() => offerFileRef.current?.click()}
                >
                  <Upload className="w-3.5 h-3.5" />
                </Button>
              </div>
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                disabled={!offerTitle.trim()}
                onClick={handleAddOffer}
              >
                <Plus className="w-3.5 h-3.5 me-1.5" />
                {t('design.pagesEditor.addOffer')}
              </Button>
            </div>
          </div>
        )}
      </div>

      <Separator />

      {/* 4. REVIEWS PAGE + MODERATION QUEUE */}
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-2">
          <Label
            htmlFor="page-reviews-toggle"
            className="text-[11px] uppercase tracking-wider cursor-pointer"
          >
            {t('design.pages.reviews')}
          </Label>
          <Switch
            id="page-reviews-toggle"
            checked={Boolean(pages.reviews)}
            onCheckedChange={(checked) => onPatchPages({ reviews: checked })}
          />
        </div>

        {pages.reviews && (
          <div className="space-y-2.5">
            <p className="text-[11px]" style={{ color: 'var(--dash-text-muted)' }}>
              {t('design.pagesEditor.reviewsHint')}
            </p>
            <div className="flex flex-wrap gap-1">
              {(
                [
                  { id: '', label: t('design.pagesEditor.filterAll') },
                  { id: 'pending', label: t('design.pagesEditor.filterPending') },
                  { id: 'approved', label: t('design.pagesEditor.filterApproved') },
                  { id: 'hidden', label: t('design.pagesEditor.filterHidden') },
                ] as const
              ).map((f) => (
                <button
                  key={f.id || 'all'}
                  type="button"
                  onClick={() => setReviewFilter(f.id)}
                  className="px-2 py-1 rounded-md text-[11px] font-medium border transition-colors"
                  style={{
                    background: reviewFilter === f.id ? 'var(--dash-active)' : 'transparent',
                    borderColor: 'var(--dash-border)',
                    color: reviewFilter === f.id ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                  }}
                >
                  {f.label}
                </button>
              ))}
            </div>

            {reviewsLoading ? (
              <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                {t('design.pagesEditor.loadingReviews')}
              </p>
            ) : reviews.length === 0 ? (
              <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                {t('design.pagesEditor.noReviews')}
              </p>
            ) : (
              <div className="space-y-2">
                {reviews.map((rev) => (
                  <div
                    key={rev.id}
                    className="rounded-lg border p-2.5 space-y-1.5"
                    style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-card)' }}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-xs font-semibold truncate">
                        {rev.authorName || 'Guest'}
                      </span>
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
                    <div className="flex items-center gap-0.5 text-amber-500">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <Star
                          key={i}
                          className="w-3 h-3"
                          fill={i < (rev.rating ?? 0) ? 'currentColor' : 'none'}
                        />
                      ))}
                    </div>
                    <p className="text-xs break-words">{rev.comment}</p>
                    <div className="flex items-center gap-1.5 pt-1">
                      {rev.status !== 'approved' && (
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-6 px-2 text-[11px]"
                          disabled={updateReviewStatus.isPending}
                          onClick={() =>
                            updateReviewStatus.mutate({ id: String(rev.id), status: 'approved' })
                          }
                        >
                          <Check className="w-3 h-3 me-1" />
                          {t('design.pagesEditor.approve')}
                        </Button>
                      )}
                      {rev.status !== 'hidden' && (
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-6 px-2 text-[11px]"
                          disabled={updateReviewStatus.isPending}
                          onClick={() =>
                            updateReviewStatus.mutate({ id: String(rev.id), status: 'hidden' })
                          }
                        >
                          <EyeOff className="w-3 h-3 me-1" />
                          {t('design.pagesEditor.hide')}
                        </Button>
                      )}
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0 ms-auto text-red-500"
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
        )}
      </div>
    </div>
  );
};

export default PagesEditor;
