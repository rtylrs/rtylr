import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import MenuItem from "./MenuItem";
import EditMenuItemModal from "./EditMenuItemModal";
import CreateCategoryModal from "./CreateCategoryModal";
import CategoryDropdown from "./CategoryDropdown";
import {
  fetchMenu,
  addCategory,
  deleteCategory,
  editMenuItem,
} from "../../../redux/reducers/menuSlice";

function Menu() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const userId = useSelector((state) => state.user.userId);
  const { items, status } = useSelector((state) => state.menu);
  const menuId = useSelector((state) => state.menu.id);

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

  const handleDeleteCategory = (categoryId) => {
    dispatch(deleteCategory(categoryId));
  };

  const openEditModal = (item) => {
    setItemToEdit(item);
    setEditModalOpen(true);
  };

  const renderSection = (item) => {
    if (item.type === "category") {
      return (
        <div key={item._id} className="category-section">
          <div className="category-header">
            <span>{item?.details?.name}</span>
            <button onClick={() => handleDeleteCategory(item._id)}>
              Delete
            </button>
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
            <button
              onClick={() => {
                navigate("/app/business/create-item", {
                  state: { categoryId: item._id },
                });
              }}
            >
              Add a product
            </button>
          </div>
        </div>
      );
    } else if (item.type === "banner") {
      return (
        <div key={item._id} className="banner-section">
          <img src={item.details.imageUrl} alt={item.details.title} />
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
        <CategoryDropdown
          categories={items.filter((item) => item.type === "category")}
          selectedCategory={selectedCategory}
          onSelectCategory={setSelectedCategory}
          onCreateCategory={() => setCreateCategoryModalOpen(true)}
        />
        <button
          onClick={() => {
            navigate("/app/restaurant/create-banner");
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
}

export default Menu;
