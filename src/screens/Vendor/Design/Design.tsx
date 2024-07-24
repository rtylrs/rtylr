import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
  setThemeAsync,
  selectUserId,
  selectTheme,
  selectLoading,
  selectError,
} from "../../../redux/reducers/userSlice";
import { AppDispatch } from "../../../redux/store";

const Design: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const userId = useSelector(selectUserId);
  const currentTheme = useSelector(selectTheme);
  const loading = useSelector(selectLoading);
  const error = useSelector(selectError);

  const [themeData, setThemeData] = useState({
    primaryColor: currentTheme?.primaryColor || "#ffffff",
    secondaryColor: currentTheme?.secondaryColor || "#000000",
    loaderColor: currentTheme?.loaderColor || "#000000",
    accentColor: currentTheme?.accentColor || "#000000",
    categoryTitleColor: currentTheme?.categoryTitleColor || "#000000",
    restaurantDescriptionColor:
      currentTheme?.restaurantDescriptionColor || "#000000",
    productTitleColor: currentTheme?.productTitleColor || "#000000",
    productDescriptionColor: currentTheme?.productDescriptionColor || "#000000",
    productPriceColor: currentTheme?.productPriceColor || "#000000",
    menuBackgroundColor: currentTheme?.menuBackgroundColor || "#ffffff",
    font: currentTheme?.font || "Arial",
    fontSize: currentTheme?.fontSize || "16px",
    categoryFontSize: currentTheme?.categoryFontSize || "16px",
    productPriceFontSize: currentTheme?.productPriceFontSize || "16px",
    productDescriptionFontSize:
      currentTheme?.productDescriptionFontSize || "16px",
    backgroundColor: currentTheme?.backgroundColor || "#ffffff",
    backgroundImage: currentTheme?.backgroundImage || "",
    imageSize: currentTheme?.imageSize || "cover",
    themeStyle: currentTheme?.themeStyle || "default",
    displaySearchButton:
      currentTheme?.displaySearchButton !== undefined
        ? currentTheme.displaySearchButton
        : true,
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setThemeData({
      ...themeData,
      [name]: type === "checkbox" ? checked : value,
    });
  };

  const handleSubmit = () => {
    if (userId) {
      dispatch(setThemeAsync({ id: userId, themeData }));
    }
  };

  return (
    <div>
      <h2>Edit Theme</h2>
      {loading && <p>Loading...</p>}
      {error && <p>Error: {error}</p>}
      <div>
        <label>
          Primary Color:
          <input
            type="color"
            name="primaryColor"
            value={themeData.primaryColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Secondary Color:
          <input
            type="color"
            name="secondaryColor"
            value={themeData.secondaryColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Loader Color:
          <input
            type="color"
            name="loaderColor"
            value={themeData.loaderColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Accent Color:
          <input
            type="color"
            name="accentColor"
            value={themeData.accentColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Category Title Color:
          <input
            type="color"
            name="categoryTitleColor"
            value={themeData.categoryTitleColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Restaurant Description Color:
          <input
            type="color"
            name="restaurantDescriptionColor"
            value={themeData.restaurantDescriptionColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Product Title Color:
          <input
            type="color"
            name="productTitleColor"
            value={themeData.productTitleColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Product Description Color:
          <input
            type="color"
            name="productDescriptionColor"
            value={themeData.productDescriptionColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Product Price Color:
          <input
            type="color"
            name="productPriceColor"
            value={themeData.productPriceColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Menu Background Color:
          <input
            type="color"
            name="menuBackgroundColor"
            value={themeData.menuBackgroundColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Font:
          <input
            type="text"
            name="font"
            value={themeData.font}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Font Size:
          <input
            type="text"
            name="fontSize"
            value={themeData.fontSize}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Category Font Size:
          <input
            type="text"
            name="categoryFontSize"
            value={themeData.categoryFontSize}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Product Price Font Size:
          <input
            type="text"
            name="productPriceFontSize"
            value={themeData.productPriceFontSize}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Product Description Font Size:
          <input
            type="text"
            name="productDescriptionFontSize"
            value={themeData.productDescriptionFontSize}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Background Color:
          <input
            type="color"
            name="backgroundColor"
            value={themeData.backgroundColor}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Background Image URL:
          <input
            type="text"
            name="backgroundImage"
            value={themeData.backgroundImage}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Image Size:
          <input
            type="text"
            name="imageSize"
            value={themeData.imageSize}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Theme Style:
          <input
            type="text"
            name="themeStyle"
            value={themeData.themeStyle}
            onChange={handleChange}
          />
        </label>
      </div>
      <div>
        <label>
          Display Search Button:
          <input
            type="checkbox"
            name="displaySearchButton"
            checked={themeData.displaySearchButton}
            onChange={handleChange}
          />
        </label>
      </div>
      <button onClick={handleSubmit}>Save Theme</button>
    </div>
  );
};

export default Design;
