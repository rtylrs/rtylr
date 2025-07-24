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
import VbitezButton from "../../../components/VbitezButton/VbitezButton";
import Menu from "../../Public/Menu/Menu";

const Design: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const userId = useSelector(selectUserId);
  const currentTheme = useSelector(selectTheme);
  const loading = useSelector(selectLoading);
  const error = useSelector(selectError);

  const defaultTheme = {
    primaryColor: "#ffffff",
    secondaryColor: "#000000",
    loaderColor: "#000000",
    accentColor: "#000000",
    categoryTitleColor: "#000000",
    restaurantDescriptionColor: "#000000",
    productTitleColor: "#000000",
    productDescriptionColor: "#000000",
    productPriceColor: "#000000",
    menuBackgroundColor: "#ffffff",
    font: "Arial",
    fontSize: "16px",
    categoryFontSize: "16px",
    productPriceFontSize: "16px",
    productDescriptionFontSize: "16px",
    backgroundColor: "#ffffff",
    backgroundImage: "",
    imageSize: "cover",
    themeStyle: "default",
    displaySearchButton: true,
  };

  const [themeData, setThemeData] = useState({
    ...defaultTheme,
    ...currentTheme,
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

  const renderInput = (label: string, name: string, type: string) => {
    return (
      <div key={name}>
        <label>
          {label}:
          <input
            type={type}
            name={name}
            value={themeData[name]}
            checked={type === "checkbox" ? themeData[name] : undefined}
            onChange={handleChange}
          />
        </label>
      </div>
    );
  };

  const themeFields = [
    { label: "Primary Color", name: "primaryColor", type: "color" },
    { label: "Secondary Color", name: "secondaryColor", type: "color" },
    { label: "Loader Color", name: "loaderColor", type: "color" },
    { label: "Accent Color", name: "accentColor", type: "color" },
    {
      label: "Category Title Color",
      name: "categoryTitleColor",
      type: "color",
    },
    {
      label: "Restaurant Description Color",
      name: "restaurantDescriptionColor",
      type: "color",
    },
    { label: "Product Title Color", name: "productTitleColor", type: "color" },
    {
      label: "Product Description Color",
      name: "productDescriptionColor",
      type: "color",
    },
    { label: "Product Price Color", name: "productPriceColor", type: "color" },
    {
      label: "Menu Background Color",
      name: "menuBackgroundColor",
      type: "color",
    },
    { label: "Font", name: "font", type: "text" },
    { label: "Font Size", name: "fontSize", type: "text" },
    { label: "Category Font Size", name: "categoryFontSize", type: "text" },
    {
      label: "Product Price Font Size",
      name: "productPriceFontSize",
      type: "text",
    },
    {
      label: "Product Description Font Size",
      name: "productDescriptionFontSize",
      type: "text",
    },
    { label: "Background Color", name: "backgroundColor", type: "color" },
    { label: "Background Image URL", name: "backgroundImage", type: "text" },
    { label: "Image Size", name: "imageSize", type: "text" },
    { label: "Theme Style", name: "themeStyle", type: "text" },
    {
      label: "Display Search Button",
      name: "displaySearchButton",
      type: "checkbox",
    },
  ];

  return (
    <div>
      <h2>Edit Theme</h2>
      {loading && <p>Loading...</p>}
      {error && (
        <p>
          Error:{" "}
          {typeof error === "string"
            ? error
            : (error as any)?.message || "An error occurred"}
        </p>
      )}
      <div
        style={{
          display: "flex",
          flexDirection: "row",
        }}
      >
        <div>
          {themeFields.map((field) =>
            renderInput(field.label, field.name, field.type)
          )}
        </div>
      </div>

      <VbitezButton onClick={handleSubmit}>Save Theme</VbitezButton>
    </div>
  );
};

export default Design;
