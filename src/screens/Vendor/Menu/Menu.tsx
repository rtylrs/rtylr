import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import MenuItem from "./MenuItem";
import CreateCategoryModal from "./CreateCategoryModal";
import { IconPickerItem } from "react-icons-picker";
import {
  fetchMenu,
  addCategory,
  deleteCategory,
  editMenuItem,
} from "../../../redux/reducers/menuSlice";
import { AppDispatch, RootState } from "../../../redux/store";

const Menu: React.FC<{}> = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const userId = useSelector((state: RootState) => state.user.userId);
  const { items, status } = useSelector((state: RootState) => state.menu);
  const menuId = useSelector((state: RootState) => state.menu.id);

  const [selectedCategory, setSelectedCategory] = useState("");
  const [isEditModalOpen, setEditModalOpen] = useState(false);
  const [isCreateCategoryModalOpen, setCreateCategoryModalOpen] =
    useState(false);
  const [itemToEdit, setItemToEdit] = useState(null);

  useEffect(() => {
    dispatch(fetchMenu(userId));

    const url = new URL(window.location.href);
    if (!url.searchParams.has("menuId")) {
      url.searchParams.append("menuId", menuId);
      window.history.replaceState({}, "", url.toString());
    }
  }, [dispatch, menuId, userId]);

  const handleEditMenuItem = (updatedItem) => {
    dispatch(editMenuItem({ categoryId: selectedCategory, item: updatedItem }));
    setEditModalOpen(false);
  };

  const handleAddCategory = (category) => {
    dispatch(addCategory({ restaurantId: userId, name: category }));
    setCreateCategoryModalOpen(false);
  };

  const handleDeleteCategory = (categoryId: string) => {
    dispatch(deleteCategory(categoryId));
  };

  const renderSection = (item: any) => {
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
            <div>
              <button
                onClick={() => {
                  navigate("/app/business/create-item", {
                    state: { categoryId: item.itemId },
                  });
                }}
              >
                Add a product
              </button>
              <button onClick={() => handleDeleteCategory(item.itemId)}>
                Delete
              </button>
            </div>
          </div>
          <div className="category-items">
            {item?.details?.menuItems?.map((menuItem) => (
              <MenuItem key={menuItem._id} item={menuItem} />
            ))}
          </div>
        </div>
      );
    } else if (item.type === "banner") {
      return (
        <div key={item._id} className="banner-section">
          <p>{item?.details?.banner?.title}</p>
          <img
            src={item?.details?.banner?.bannerImage}
            alt={item?.details?.title}
          />
        </div>
      );
    }
    return null;
  };

  if (status === "loading") {
    return <div>Loading...</div>;
  }

  return (
    <div className="menu-container-vendor">
      <h1>Menu</h1>
      <div className="category-controls">
        <button
          onClick={() => {
            setCreateCategoryModalOpen(true);
          }}
        >
          Create Category
        </button>
        <button
          onClick={() => {
            navigate("/app/business/create-banner");
          }}
        >
          Create Banner
        </button>
      </div>
      <div className="menu-items">{items && items.map(renderSection)}</div>
      <CreateCategoryModal
        isOpen={isCreateCategoryModalOpen}
        onClose={() => setCreateCategoryModalOpen(false)}
        onSave={handleAddCategory}
      />
      {/* <EditMenuItemModal
        isOpen={isEditModalOpen}
        onClose={() => setEditModalOpen(false)}
        onSave={handleEditMenuItem}
        item={itemToEdit}
      /> */}
    </div>
  );
};

export default Menu;
