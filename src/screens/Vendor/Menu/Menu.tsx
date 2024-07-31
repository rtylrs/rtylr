import { useEffect, useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import MenuItem from "./MenuItem";
import CreateCategoryModal from "./CreateCategoryModal";
import { fetchMenu, deleteCategory } from "../../../redux/reducers/menuSlice";
import { AppDispatch, RootState } from "../../../redux/store";

const Menu: React.FC<{}> = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const userId = useSelector((state: RootState) => state.user.userId);
  const { items, status } = useSelector((state: RootState) => state.menu);
  const menuId = useSelector((state: RootState) => state.menu.id);

  const [isCreateCategoryModalOpen, setCreateCategoryModalOpen] =
    useState(false);

  useEffect(() => {
    dispatch(fetchMenu({ userId: userId as string, disabled: true }));

    const url = new URL(window.location.href);
    if (!url.searchParams.has("menuId")) {
      url.searchParams.append("menuId", menuId as string);
      window.history.replaceState({}, "", url.toString());
    }
  }, [dispatch, menuId, userId]);

  // const handleEditMenuItem = (updatedItem) => {
  //   dispatch(editMenuItem({ categoryId: selectedCategory, item: updatedItem }));
  //   setEditModalOpen(false);
  // };

  // const handleAddCategory = (category) => {
  //   dispatch(addCategory({ restaurantId: userId, name: category }));
  //   setCreateCategoryModalOpen(false);
  // };

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
                <img src={item?.details?.icon} alt={item?.details?.name} />
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
            {item?.details?.menuItems?.map((menuItem: any) => (
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
            style={{ width: "50%", height: "50%", objectFit: "cover" }}
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
        addOrEdit="add"
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
