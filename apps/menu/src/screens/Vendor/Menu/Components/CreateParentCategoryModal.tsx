import { useState, useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "react-modal";
import {
  addCategory,
  editCategory,
} from "../../../../redux/reducers/menuSlice";
import { AppDispatch, RootState } from "../../../../redux/store";

interface CreateParentCategoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  addOrEdit: "add" | "edit";
  categoryData?: {
    _id: string;
    name: string;
    description: string;
    icon: string;
  };
}

const CreateParentCategoryModal: React.FC<CreateParentCategoryModalProps> = ({
  isOpen,
  onClose,
  addOrEdit,
  categoryData,
}) => {
  const [category, setCategory] = useState(categoryData?.name || "");
  const [description, setDescription] = useState(
    categoryData?.description || ""
  );
  const [iconURL, setIconURL] = useState(categoryData?.icon || "");
  const dispatch = useDispatch<AppDispatch>();
  const restaurantId = useSelector((state: RootState) => state.user.userId);

  useEffect(() => {
    // Modal open state changed
  }, [isOpen]);

  const handleSubmit = (e: any) => {
    e.preventDefault();
    if (category.trim()) {
      const categoryPayload = {
        restaurantId,
        name: category,
        description,
        isParentCategory: true,
        icon: iconURL,
      };

      if (addOrEdit === "add") {
        dispatch(addCategory(categoryPayload));
      } else if (addOrEdit === "edit" && categoryData) {
        dispatch(
          editCategory({
            categoryId: categoryData?._id,
            category: categoryPayload,
          })
        );
      }

      setCategory("");
      setDescription("");
      setIconURL("");
      onClose();
    } else {
      alert("Category name cannot be empty.");
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onRequestClose={onClose}
      className="category-modal"
      overlayClassName="modal-overlay"
    >
      <h2>{addOrEdit === "add" ? "Create" : "Edit"} Parent Category</h2>
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
            placeholder="e.g. Main Courses, Beverages, etc."
          />
        </div>
        <input
          type="text"
          placeholder="Paste custom icon URL"
          value={iconURL}
          onChange={(e) => setIconURL(e.target.value)}
        />
        <div className="form-actions">
          <button type="submit">Save</button>
          <button type="button" onClick={onClose}>
            Cancel
          </button>
        </div>
      </form>
    </Modal>
  );
};

export default CreateParentCategoryModal;
