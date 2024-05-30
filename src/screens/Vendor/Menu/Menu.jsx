import { useState } from "react";
import MenuItem from "./MenuItem";
import CreateMenuItemModal from "./CreateMenuItemModal";
import EditMenuItemModal from "./EditMenuItemModal";
import CreateCategoryModal from "./CreateCategoryModal";
import CategoryDropdown from "./CategoryDropdown";

function Menu() {
  const [menuItems, setMenuItems] = useState([]);
  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState("");
  const [isCreateModalOpen, setCreateModalOpen] = useState(false);
  const [isEditModalOpen, setEditModalOpen] = useState(false);
  const [isCreateCategoryModalOpen, setCreateCategoryModalOpen] =
    useState(false);
  const [itemToEdit, setItemToEdit] = useState(null);

  const addMenuItem = (item) => {
    setMenuItems([...menuItems, item]);
  };

  const editMenuItem = (updatedItem) => {
    setMenuItems(
      menuItems.map((item) => (item.id === updatedItem.id ? updatedItem : item))
    );
  };

  const addCategory = (category) => {
    setCategories([...categories, category]);
  };

  const deleteCategory = (category) => {
    setCategories(categories.filter((cat) => cat !== category));
    setMenuItems(
      menuItems.map((item) => ({
        ...item,
        categories: item.categories.filter((cat) => cat !== category),
      }))
    );
    if (selectedCategory === category) setSelectedCategory("");
  };

  const openEditModal = (item) => {
    setItemToEdit(item);
    setEditModalOpen(true);
  };

  const filteredMenuItems = selectedCategory
    ? menuItems.filter((item) => item.categories.includes(selectedCategory))
    : menuItems;

  return (
    <div className="menu-container">
      <h1>Menu</h1>
      <CategoryDropdown
        categories={categories}
        selectedCategory={selectedCategory}
        onSelectCategory={setSelectedCategory}
        onCreateCategory={() => setCreateCategoryModalOpen(true)}
        onDeleteCategory={deleteCategory}
      />
      <button
        className="add-item-button"
        onClick={() => setCreateModalOpen(true)}
      >
        Add Menu Item
      </button>
      <div className="menu-items">
        {filteredMenuItems.map((item) => (
          <MenuItem
            key={item.id}
            item={item}
            onEdit={() => openEditModal(item)}
          />
        ))}
      </div>
      <CreateMenuItemModal
        isOpen={isCreateModalOpen}
        onClose={() => setCreateModalOpen(false)}
        categories={categories}
        onSave={addMenuItem}
      />
      <EditMenuItemModal
        isOpen={isEditModalOpen}
        onClose={() => setEditModalOpen(false)}
        item={itemToEdit}
        categories={categories}
        onSave={editMenuItem}
      />
      <CreateCategoryModal
        isOpen={isCreateCategoryModalOpen}
        onClose={() => setCreateCategoryModalOpen(false)}
        onSave={addCategory}
      />
    </div>
  );
}

export default Menu;
