import { useState, useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "react-modal";
import { addCategory, editCategory } from "../../../redux/reducers/menuSlice";
import { AppDispatch, RootState } from "../../../redux/store";

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
  const [iconURL, setIconURL] = useState("");
  const dispatch = useDispatch<AppDispatch>();
  const restaurantId = useSelector((state: RootState) => state.user.userId);

  useEffect(() => {
    console.log("Modal opened:", isOpen);
  }, [isOpen]);

  const handleSubmit = (e: any) => {
    e.preventDefault();
    if (category.trim()) {
      const categoryPayload = {
        restaurantId,
        name: category,
        description,
        hasAlcohol,
        availability,
        icon: "" || iconURL,
      };
      console.log("Category Payload:", categoryPayload);

      if (addOrEdit === "add") {
        dispatch(addCategory(categoryPayload));
        console.log("Dispatching addCategory");
      } else if (addOrEdit === "edit" && categoryData) {
        dispatch(
          editCategory({
            categoryId: categoryData?._id,
            category: categoryPayload,
          })
        );
        console.log("Dispatching editCategory");
      }
      setCategory("");
      setDescription("");
      setHasAlcohol(false);
      setAvailability("always available");
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
      <h2>{addOrEdit === "add" ? "Create" : "Edit"} Category</h2>
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

export default CreateCategoryModal;
