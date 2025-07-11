import { useEffect, useState } from "react";
import MenuItem from "./Components/MenuItem";
import useAsync from "../../../hooks/useAsync";
import { menuApi } from "../../../api";
import { useParams } from "react-router-dom";

function MenuWithId() {
  const [categories, setCategories] = useState([]);
  const [restaurant, setRestaurant] = useState({});
  const [status, setStatus] = useState("idle");
  const { id } = useParams();
  const getMenu = useAsync({
    fn: menuApi.getMenuWithMenuId,
    onSuccess: () => {
      setCategories(getMenu.result.details.menu.items);
      setRestaurant(getMenu.result.details.business);
      setStatus("succeeded");
    },
    onError: () => {
      setStatus("failed");
    },
    callOnMount: false,
  });

  useEffect(() => {
    setStatus("loading");
    getMenu.main({ id });

     
  }, []);

  const renderCategorySection = (category) => (
    <div key={category._id} className="category-section">
      <div className="category-header">
        <span>{category?.details?.name}</span>
      </div>
      <div className="category-items">
        {category &&
          category.details.menuItems.map((item) => (
            <MenuItem key={item._id} item={item} />
          ))}
      </div>
    </div>
  );

  if (status === "loading") {
    return <div>Loading...</div>;
  }

  if (status === "failed") {
    return <div>Error loading menu.</div>;
  }

  return (
    <div className="menu-container">
      <h1>{restaurant?.name}'s Menu</h1>
      <div className="category-controls"></div>
      <div className="menu-items">
        {categories &&
          categories
            .filter((item) => item.type === "category")
            .map(renderCategorySection)}
      </div>
    </div>
  );
}

export default MenuWithId;
