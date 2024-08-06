import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { addBanner } from "../../../../redux/reducers/menuSlice";
import { AppDispatch, RootState } from "../../../../redux/store"; // Adjust the import according to your store setup
import Dropzone from "../../../../components/Dropzone";

const CreateBanner = (): JSX.Element => {
  const dispatch = useDispatch<AppDispatch>();
  const restaurantId = useSelector((state: RootState) => state.user.userId);

  const [formData, setFormData] = useState({
    title: "",
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
      console.error("Restaurant ID is required");
      return;
    }
    const bannerData = new FormData();
    bannerData.append("title", formData.title);
    bannerData.append("restaurantId", restaurantId);
    if (formData.image) {
      bannerData.append("bannerImage", formData.image);
    }

    dispatch(addBanner(bannerData));
  };

  return (
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
  );
};

export default CreateBanner;
