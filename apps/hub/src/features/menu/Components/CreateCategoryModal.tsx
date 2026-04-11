import { useState, useEffect } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import { useMenuStore } from "../../../shared/stores/menuStore";
import { useUserStore } from "../../../shared/stores/userStore";

interface CreateCategoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  addOrEdit: "add" | "edit";
  categoryData?: {
    _id: string;
    name: string;
    description: string;
    hasAlcohol: boolean;
    availability: string;
    icon: string;
    parentCategory?: string;
  };
}

const CreateCategoryModal: React.FC<CreateCategoryModalProps> = ({
  isOpen,
  onClose,
  addOrEdit,
  categoryData,
}) => {
  const [category, setCategory] = useState(categoryData?.name || "");
  const [description, setDescription] = useState(
    categoryData?.description || ""
  );
  const [hasAlcohol, setHasAlcohol] = useState(
    categoryData?.hasAlcohol || false
  );
  const [availability, setAvailability] = useState(
    categoryData?.availability || "always available"
  );
  const [iconURL, setIconURL] = useState(categoryData?.icon || "");
  const [parentCategory, setParentCategory] = useState(
    categoryData?.parentCategory || ""
  );

  const { userId: restaurantId } = useUserStore();
  const { items, addCategory, editCategory } = useMenuStore();
  const parentCategories = items.filter(
    (item: any) => item.type === "category" && item.details.isParentCategory
  );

  useEffect(() => {
    if (isOpen && categoryData) {
      setCategory(categoryData.name || "");
      setDescription(categoryData.description || "");
      setHasAlcohol(categoryData.hasAlcohol || false);
      setAvailability(categoryData.availability || "always available");
      setIconURL(categoryData.icon || "");
      setParentCategory(categoryData.parentCategory || "");
    }
  }, [isOpen, categoryData]);

  const handleSubmit = (e: any) => {
    e.preventDefault();
    if (category.trim() && parentCategory) {
      const categoryPayload = {
        restaurantId,
        name: category,
        description,
        hasAlcohol,
        availability,
        icon: iconURL,
        parentCategory,
      };

      if (addOrEdit === "add") {
        addCategory(categoryPayload);
      } else if (addOrEdit === "edit" && categoryData) {
        editCategory(categoryData?._id, categoryPayload);
      }
      setCategory("");
      setDescription("");
      setHasAlcohol(false);
      setAvailability("always available");
      setIconURL("");
      setParentCategory("");
      onClose();
    } else {
      alert("Category name and parent category cannot be empty.");
    }
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="modal-overlay" />
        <Dialog.Content className="category-modal">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Dialog.Title asChild>
              <h2>{addOrEdit === "add" ? "Create" : "Edit"} Category</h2>
            </Dialog.Title>
            <Dialog.Close asChild>
              <button style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4 }}>
                <X size={18} />
              </button>
            </Dialog.Close>
          </div>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label>Category Name*</label>
              <input
                type="text"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label>Description</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="e.g. Appetizers, Salads, Drinks, etc."
              />
            </div>
            <div className="form-group">
              <label>
                <input
                  type="checkbox"
                  checked={hasAlcohol}
                  onChange={(e) => setHasAlcohol(e.target.checked)}
                />
                Show alcohol warning
              </label>
            </div>
            <div className="form-group">
              <label>Category Availability:</label>
              <select
                value={availability}
                onChange={(e) => setAvailability(e.target.value)}
              >
                <option value="hidden">Hidden</option>
                <option value="available at certain time">
                  Available at certain time
                </option>
                <option value="always available">Always available</option>
              </select>
            </div>
            <div className="form-group">
              <label>Parent Category*</label>
              <select
                value={parentCategory}
                onChange={(e) => setParentCategory(e.target.value)}
                required
              >
                <option value="">Select Parent Category</option>
                {parentCategories.map((parent: any) => (
                  <option key={parent.itemId} value={parent.itemId}>
                    {parent.details.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Icon URL</label>
              <input
                type="text"
                placeholder="Paste custom icon URL"
                value={iconURL}
                onChange={(e) => setIconURL(e.target.value)}
              />
            </div>
            <div className="form-actions">
              <button type="submit">Save</button>
              <button type="button" onClick={onClose}>
                Cancel
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
};

export default CreateCategoryModal;
