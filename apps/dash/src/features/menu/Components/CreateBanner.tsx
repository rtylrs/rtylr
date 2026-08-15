import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useMenuStore } from '../../../shared/stores/menuStore';
import { useUserStore } from '../../../shared/stores/userStore';
import Dropzone from '../../../shared/components/Dropzone';

import { toast } from 'sonner';

const CreateBanner = (): React.JSX.Element => {
  const { addBanner } = useMenuStore();
  const { userId: restaurantId } = useUserStore();

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

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const bannerObj = {
      _id: String(Date.now()),
      title: formData.title,
      bannerImage: formData.image ? URL.createObjectURL(formData.image) : '',
      menuId: restaurantId ?? '',
      type: 'banner' as const,
      details: {
        banner: {
          title: formData.title,
          bannerImage: formData.image ? URL.createObjectURL(formData.image) : '',
        },
      },
    };

    addBanner(bannerObj as any);
    toast.success('Banner created');
    window.history.back();
  };

  return (
    <>
      <HelmetContainer title="Create banner · rtylr" />
      <div className="max-w-2xl mx-auto space-y-6">
        <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          Create Banner
        </h1>
        <form
          onSubmit={handleSubmit}
          className="rounded-xl p-6 border space-y-4"
          style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
        >
          <div className="space-y-1">
            <label className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Banner Title *
            </label>
            <input
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
              className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity text-white"
              style={{ background: '#4583C4' }}
            >
              Create Banner
            </button>
          </div>
        </form>
      </div>
    </>
  );
};

export default CreateBanner;
