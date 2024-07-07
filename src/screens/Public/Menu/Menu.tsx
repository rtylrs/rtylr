import { useEffect, useState } from "react";
import MenuItem from "./MenuItem";
import useAsync from "../../../hooks/useAsync";
import menuApi from "../../../api/menu";
import { useParams } from "react-router-dom";
import { IconPickerItem } from "react-icons-picker";

function Menu() {
  const [categories, setCategories] = useState([]);
  const [restaurant, setRestaurant] = useState({});
  const [theme, setTheme] = useState({});
  const [status, setStatus] = useState("idle");
  const { id } = useParams();

  const getMenu = useAsync({
    fn: menuApi.getMenu,
    onSuccess: () => {
      const { details } = getMenu.result;
      setCategories(details.menu.items);
      setRestaurant(details.business);
      setTheme(details.theme);
      setStatus("succeeded");
    },
    onError: () => {
      setStatus("failed");
    },
    callOnMount: false,
  });

  useEffect(() => {
    setStatus("loading");
    getMenu.main({ id }); // Replace with actual restaurant ID

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const renderSection = (item) => {
    if (item.type === "category") {
      return (
        <div key={item._id} className="category-section">
          <div className="category-header">
            <div>
              <span>{item?.details?.name}</span>
              {item?.details?.icon && (
                <IconPickerItem value={item?.details?.icon} />
              )}
            </div>
          </div>
          <div className="category-items">
            {item &&
              item?.details?.menuItems.map((menuItem) => (
                <MenuItem
                  key={menuItem._id}
                  item={menuItem}
                  onEdit={() => openEditModal(menuItem)}
                />
              ))}
          </div>
        </div>
      );
    } else if (item.type === "banner") {
      return (
        <div key={item._id} className="banner-section">
          <p>{item?.details?.banner.title}</p>
          <img src={item.details.banner.bannerImage} alt={item.details.title} />
        </div>
      );
    }
    return null;
  };

  if (status === "loading") {
    return <div>Loading...</div>;
  }

  if (status === "failed") {
    return <div>Error loading menu.</div>;
  }

  return (
    <div
      className="menu-container"
      style={{ backgroundColor: theme?.menuBackgroundColor }}
    >
      <h1 style={{ color: theme?.primaryColor }}>{restaurant?.name}'s Menu</h1>
      <div className="category-controls"></div>
      <div className="menu-items">
        {categories && categories.map(renderSection)}
      </div>
    </div>
  );
}

export default Menu;
