import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { uploadAsset } from '@voxire/api';
import UnavailableNotice from '../../../shared/components/dash/UnavailableNotice';
import { useUserStore } from '../../../shared/stores/userStore';
import Dropzone from '../../../shared/components/Dropzone';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

const UNAVAILABLE_ID = 'menu-banner-saving-unavailable';

const CreateBanner = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { userId: restaurantId } = useUserStore();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    title: '',
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
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      let bannerImageUrl = '';
      if (formData.image) {
        const res = await uploadAsset(formData.image);
        bannerImageUrl = res.url || '';
      }

      const bannerObj = {
        _id: String(Date.now()),
        title: formData.title,
        bannerImage: bannerImageUrl,
        menuId: restaurantId ?? '',
        type: 'banner' as const,
        details: {
          banner: {
            title: formData.title,
            bannerImage: bannerImageUrl,
          },
        },
      };

      // Same as CreateItem: addBanner touches only the local store, so the
      // success toast was untrue. Also drops an `as any` the store's own
      // Banner type never needed. rtylr#1150.
      void bannerObj;
      toast.error(t('menu.savingUnavailableToast'));
    } catch {
      toast.error('Failed to upload image. Please try again.');
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
        <UnavailableNotice>
          <span id={UNAVAILABLE_ID}>{t('menu.savingUnavailable')}</span>
        </UnavailableNotice>
        <form
          onSubmit={handleSubmit}
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
              disabled={isSubmitting}
              className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-50"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled
              aria-describedby={UNAVAILABLE_ID}
              className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:opacity-50"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {isSubmitting ? 'Uploading...' : 'Create Banner'}
            </button>
          </div>
        </form>
      </div>
    </>
  );
};

export default CreateBanner;
