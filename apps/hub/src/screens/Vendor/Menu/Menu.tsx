import { Spinner } from '@voxire/components';
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import MenuItem from "./Components/MenuItem";
import CreateCategoryModal from "./Components/CreateCategoryModal";
import { useMenuStore } from "../../../stores/menuStore";
import { useUserStore } from "../../../stores/userStore";
import CreateParentCategoryModal from "./Components/CreateParentCategoryModal";

const Menu: React.FC<{}> = () => {
  const navigate = useNavigate();
  const { userId } = useUserStore();
  const {
    items,
    status,
    id: menuId,
    fetchMenu,
    deleteCategory,
  } = useMenuStore();

  const [isCreateCategoryModalOpen, setCreateCategoryModalOpen] =
    useState(false);
  const [isCreateParentCategoryModalOpen, setCreateParentCategoryModalOpen] =
    useState(false);

  useEffect(() => {
    fetchMenu(userId as string, true);

    const url = new URL(window.location.href);
    if (!url.searchParams.has("menuId")) {
      url.searchParams.append("menuId", menuId as string);
      window.history.replaceState({}, "", url.toString());
    }
  }, [fetchMenu, menuId, userId]);

  const handleDeleteCategory = (categoryId: string) => {
    deleteCategory(categoryId);
  };

  const renderItems = (items: any) => {
    return items.map((item: any) => <MenuItem key={item._id} item={item} />);
  };

  const renderCategory = (category: any) => {
    return (
      <div key={category._id} className={styles.categorySection}>
        <div className={styles.categoryHeader}>
          <div>
            <span>{category?.details?.name}</span>
            {category?.details?.icon && (
              <img
                src={category?.details?.icon}
                alt={category?.details?.name}
              />
            )}
          </div>
          <div>
            <button
              onClick={() => {
                navigate("/app/business/create-item", {
                  state: { categoryId: category.itemId },
                });
              }}
            >
              Add a product
            </button>
            <button onClick={() => handleDeleteCategory(category.itemId)}>
              Delete
            </button>
          </div>
        </div>
        <div className={styles.categoryItems}>
          {renderItems(category.details.menuItems)}
        </div>
      </div>
    );
  };

  const renderParentCategory = (parentCategory: any, childCategories: any) => {
    return (
      <div key={parentCategory._id} className={styles.parentCategorySection}>
        <div className={styles.parentCategoryHeader}>
          <div>
            <span>{parentCategory?.details?.name}</span>
            {parentCategory?.details?.icon && (
              <img
                src={parentCategory?.details?.icon}
                alt={parentCategory?.details?.name}
              />
            )}
          </div>
          <div>
            {!parentCategory && (
              <button
                onClick={() => {
                  navigate("/app/business/create-item", {
                    state: { categoryId: parentCategory.itemId },
                  });
                }}
              >
                Add a product
              </button>
            )}
            <button onClick={() => handleDeleteCategory(parentCategory.itemId)}>
              Delete
            </button>
          </div>
        </div>
        <div className={styles.childCategories}>
          {childCategories.map(renderCategory)}
        </div>
      </div>
    );
  };

  const groupedCategories = items.reduce((acc: any, item: any) => {
    if (item.type === "category") {
      if (item.details.isParentCategory) {
        acc[item.itemId] = { parent: item, children: [] };
      } else if (item.details.parentCategory) {
        if (!acc[item.details.parentCategory]) {
          acc[item.details.parentCategory] = { parent: null, children: [] };
        }
        acc[item.details.parentCategory].children.push(item);
      } else {
        acc[item.itemId] = { parent: item, children: [] };
      }
    } else if (item.type === "banner") {
      acc[item.itemId] = item;
    }
    return acc;
  }, {});

  const renderSections = (items: any) => {
    return items.map((item: any) => {
      if (item.type === "banner") {
        return (
          <div key={item._id} className={styles.bannerSectionVendor}>
            <div>
              <p>{item?.details?.banner?.title}</p>

              <div>
                <button onClick={() => setCreateCategoryModalOpen(true)}>
                  Delete
                </button>
              </div>
            </div>
            <img
              style={{ width: "50%", height: "50%", objectFit: "cover" }}
              src={item?.details?.banner?.bannerImage}
              alt={item?.details?.title}
            />
          </div>
        );
      } else if (item.type === "category") {
        if (item.details.isParentCategory) {
          const childCategories = groupedCategories[item.itemId].children;
          return renderParentCategory(item, childCategories);
        } else if (item.details.parentCategory) {
          return null;
        } else {
          return renderCategory(item);
        }
      }
      return null;
    });
  };

  if (status === "loading") {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "80vh",
        }}
      >
        <CircularProgress size={60} color="primary" />
      </div>
    );
  }

  return (
    <div className={styles.menuContainerVendor}>
      <div className={styles.menuHeader}>
        <h1>Menu</h1>
        <div className={styles.menuButtons}>
          <button
            onClick={() => setCreateParentCategoryModalOpen(true)}
            className={styles.menuButton}
          >
            Create Group
          </button>
          <button
            onClick={() => {
              setCreateCategoryModalOpen(true);
            }}
            className={styles.menuButton}
          >
            Create Category
          </button>
          <button
            onClick={() => {
              navigate("/app/business/create-banner");
            }}
            className={styles.menuButton}
          >
            Create Banner
          </button>
        </div>
      </div>
      <div className={styles.menuItems}>{renderSections(items)}</div>
      <CreateCategoryModal
        isOpen={isCreateCategoryModalOpen}
        onClose={() => setCreateCategoryModalOpen(false)}
        addOrEdit="add"
      />
      <CreateParentCategoryModal
        isOpen={isCreateParentCategoryModalOpen}
        onClose={() => setCreateParentCategoryModalOpen(false)}
        addOrEdit="add"
      />
    </div>
  );
};

export default Menu;
