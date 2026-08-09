import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useMenuStore } from '../../../shared/stores/menuStore';
import { useUserStore } from '../../../shared/stores/userStore';
import Dropzone from '../../../shared/components/Dropzone';

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
    if (!restaurantId) {
      return;
    }
    const bannerData = new FormData();
    bannerData.append('title', formData.title);
    bannerData.append('restaurantId', restaurantId);
    if (formData.image) {
      bannerData.append('bannerImage', formData.image);
    }

    addBanner(bannerData as any);
  };

  return (
    <>
      <HelmetContainer title="Create banner · rtylr" />
      <div>
        <h1>Create Banner</h1>
        <form onSubmit={handleSubmit}>
          <div>
            <label>
              Title:
              <input
                type="text"
                name="title"
                value={formData.title}
                onChange={handleChange}
                required
              />
            </label>
          </div>
          <Dropzone onFileSelected={handleFileSelected} />
          {formData.image && <p>Selected file: {formData.image.name}</p>}
          <button type="submit">Create Banner</button>
        </form>
      </div>
    </>
  );
};

export default CreateBanner;
