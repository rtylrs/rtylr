import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { uploadAsset } from '@voxire/api';
import { useNavigate } from '@tanstack/react-router';
import Dropzone from '../../../shared/components/Dropzone';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { useCreateMenuBanner } from '../../../shared/hooks/queries/useMenu';

const CreateBanner = (): React.JSX.Element => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const createBanner = useCreateMenuBanner();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    title: '',
    body: '',
    linkUrl: '',
    image: null as File | null,
  });

  const handleFileSelected = (file: File) => {
    setFormData((prevData) => ({
      ...prevData,
      image: file,
    }));
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const title = formData.title.trim();
    if (!title) {
      toast.error(t('menu.bannerTitleRequired', 'Please enter a banner title'));
      return;
    }

    setIsSubmitting(true);
    try {
      let bannerImageUrl = '';
      if (formData.image) {
        const res = await uploadAsset(formData.image, { visibility: 'public' });
        bannerImageUrl = res.url || '';
      }

      await createBanner.mutateAsync({
        title,
        body: formData.body.trim(),
        imageUrl: bannerImageUrl,
        linkUrl: formData.linkUrl.trim(),
        isActive: true,
      });

      toast.success(t('menu.bannerCreated', 'Promo banner published to your digital menu'));
      void navigate({ to: '/menus/promos' });
    } catch (err: unknown) {
      const msg =
        (err as { message?: string })?.message ||
        t('menu.bannerSaveFailed', 'Failed to create banner. Please try again.');
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <>
      <HelmetContainer title="Create banner · rtylr" />
      <div className="max-w-2xl mx-auto space-y-6">
        <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          Create Banner
        </h1>
        <form
          onSubmit={(e) => void handleSubmit(e)}
          className="rounded-xl p-6 border space-y-4"
          style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
        >
          <div className="space-y-1">
            <label
              htmlFor="banner-title"
              className="text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Banner Title *
            </label>
            <input
              id="banner-title"
              type="text"
              name="title"
              value={formData.title}
              onChange={handleChange}
              placeholder="Summer Special Banner"
              required
              className="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:outline-none"
              style={{
                background: 'var(--dash-hover)',
                borderColor: 'var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            />
          </div>
          <div className="space-y-1">
            <label
              htmlFor="banner-body"
              className="text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Subtitle / Offer Details
            </label>
            <input
              id="banner-body"
              type="text"
              name="body"
              value={formData.body}
              onChange={handleChange}
              placeholder="20% off iced drinks from 2pm to 5pm"
              className="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:outline-none"
              style={{
                background: 'var(--dash-hover)',
                borderColor: 'var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            />
          </div>
          <div className="space-y-1">
            <label
              htmlFor="banner-link"
              className="text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Link URL (optional)
            </label>
            <input
              id="banner-link"
              type="url"
              name="linkUrl"
              value={formData.linkUrl}
              onChange={handleChange}
              placeholder="https://example.com/offer"
              className="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:outline-none"
              style={{
                background: 'var(--dash-hover)',
                borderColor: 'var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Banner Image
            </label>
            <Dropzone onFileSelected={handleFileSelected} />
            {formData.image && (
              <p className="text-xs mt-1 font-medium" style={{ color: 'var(--dash-text-muted)' }}>
                Selected file: {formData.image.name}
              </p>
            )}
          </div>
          <div
            className="flex items-center justify-end gap-3 pt-4 border-t"
            style={{ borderColor: 'var(--dash-border)' }}
          >
            <button
              type="button"
              onClick={() => window.history.back()}
              disabled={isSubmitting || createBanner.isPending}
              className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-50"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting || createBanner.isPending}
              className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-50"
              style={{ background: 'var(--accent)', color: 'var(--dash-on-accent)' }}
            >
              {isSubmitting || createBanner.isPending ? 'Saving...' : 'Create Banner'}
            </button>
          </div>
        </form>
      </div>
    </>
  );
};

export default CreateBanner;
