import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { addBanner } from "../../../redux/reducers/menuSlice";
import { AppDispatch, RootState } from "../../../redux/store"; // Adjust the import according to your store setup

const CreateBanner = (): JSX.Element => {
  const dispatch = useDispatch<AppDispatch>();
  const restaurantId = useSelector((state: RootState) => state.user.userId);

  const [formData, setFormData] = useState({
    title: "",
    image: null as File | null,
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, files } = e.target;
    if (name === "image" && files) {
      setFormData({
        ...formData,
        image: files[0],
      });
    } else {
      setFormData({
        ...formData,
        [name]: value,
      });
    }
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
        <div>
          <label>
            Image:
            <input
              type="file"
              name="image"
              accept="image/*"
              onChange={handleChange}
            />
          </label>
        </div>
        <button type="submit">Create Banner</button>
      </form>
    </div>
  );
};

export default CreateBanner;
