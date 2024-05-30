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
  const { items: categories, status } = useSelector((state) => state.menu);
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
    dispatch(addCategory(category));
    setCreateCategoryModalOpen(false);
  };

  const handleDeleteCategory = (categoryId) => {
    dispatch(deleteCategory(categoryId));
  };

  const openEditModal = (item) => {
    setItemToEdit(item);
    setEditModalOpen(true);
  };

  const renderCategorySection = (category) => (
    <div key={category._id} className="category-section">
      <div className="category-header">
        <span>{category.name}</span>
        <button onClick={() => handleDeleteCategory(category._id)}>
          Delete
        </button>
      </div>
      <div className="category-items">
        {category &&
          category.menuItems.map((item) => (
            <MenuItem
              key={item._id}
              item={item}
              onEdit={() => openEditModal(item)}
            />
          ))}
        <button
          onClick={() => {
            navigate("/app/restaurant/create-item", {
              state: { categoryId: category._id },
            });
          }}
        >
          Add a product
        </button>
      </div>
    </div>
  );

  if (status === "loading") {
    return <div>Loading...</div>;
  }

  return (
    <div className="menu-container">
      <h1>Menu</h1>
      <div className="category-controls">
        {/* <CategoryDropdown
          categories={categories}
          selectedCategory={selectedCategory}
          onSelectCategory={setSelectedCategory}
          onCreateCategory={() => setCreateCategoryModalOpen(true)}
        /> */}
      </div>
      <div className="menu-items">
        {categories &&
          categories
            .filter((item) => item.type === "category")
            .map(renderCategorySection)}
      </div>
      {/* <EditMenuItemModal
        isOpen={isEditModalOpen}
        onClose={() => setEditModalOpen(false)}
        item={itemToEdit}
        categories={categories}
        onSave={handleEditMenuItem}
      /> */}
      <CreateCategoryModal
        isOpen={isCreateCategoryModalOpen}
        onClose={() => setCreateCategoryModalOpen(false)}
        onSave={handleAddCategory}
      />
    </div>
  );
}

export default Menu;
