import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "react-modal";
import { addCategory } from "../../../redux/reducers/menuSlice";
import { IconPicker } from "react-fa-icon-picker";

function CreateCategoryModal({ isOpen, onClose }) {
  const [category, setCategory] = useState("");
  const dispatch = useDispatch();
  const restaurantId = useSelector((state) => state.user.userId);
  const [icon, setIcon] = useState("");
  const [iconURL, setIconURL] = useState("");
  const [iconInputType, setIconInputType] = useState("picker"); // State to manage input type

  const handleSubmit = (e) => {
    e.preventDefault();
    const iconValue = iconInputType === "url" ? iconURL : icon;
    if (category.trim()) {
      dispatch(addCategory({ restaurantId, name: category, icon: iconValue }));
      setCategory("");
      setIcon("");
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
      className="modal"
      overlayClassName="modal-overlay"
    >
      <h2>Create Category</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Category Name"
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          required
        />
        <div>
          <label>
            <input
              type="radio"
              name="iconInputType"
              value="picker"
              checked={iconInputType === "picker"}
              onChange={() => setIconInputType("picker")}
            />
            Use Icon Picker
          </label>
          <label>
            <input
              type="radio"
              name="iconInputType"
              value="url"
              checked={iconInputType === "url"}
              onChange={() => setIconInputType("url")}
            />
            Use Custom URL
          </label>
        </div>
        {iconInputType === "picker" ? (
          <IconPicker value={icon} onChange={setIcon} />
        ) : (
          <input
            type="text"
            placeholder="Paste custom icon URL"
            value={iconURL}
            onChange={(e) => setIconURL(e.target.value)}
          />
        )}
        <button type="submit">Save</button>
        <button type="button" onClick={onClose}>
          Cancel
        </button>
      </form>
    </Modal>
  );
}

export default CreateCategoryModal;
