import React, { useState } from "react";
import RenderSection from "../Components/RenderSection";

const MenuSection: React.FC<{
  menuArray: any;
}> = ({ menuArray }) => {
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const categories = menuArray.filter((item) => item.type === "category");

  const handleCategoryClick = (categoryId: string | null) => {
    setSelectedCategory(categoryId);
  };

  const filteredMenuArray = selectedCategory
    ? menuArray.filter(
        (item) => item.type === "category" && item._id === selectedCategory
      )
    : menuArray;

  return (
    <div className="menu-container">
      <div className="category-navigation">
        <button
          className={`category-link ${
            selectedCategory === null ? "active" : ""
          }`}
          onClick={() => handleCategoryClick(null)}
        >
          Show All
        </button>
        {categories.map((category) => (
          <button
            key={category._id}
            className={`category-link ${
              selectedCategory === category._id ? "active" : ""
            }`}
            onClick={() => handleCategoryClick(category._id)}
          >
            {category.details.name}
          </button>
        ))}
      </div>

      <div className="menu">
        <div className="category-controls"></div>
        <div className="menu-items">
          {filteredMenuArray && filteredMenuArray.map(RenderSection)}
        </div>
      </div>
    </div>
  );
};

export default MenuSection;
