import { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Modal from "react-modal";
import { addCategory } from "../../../redux/reducers/menuSlice";

function CreateCategoryModal({ isOpen, onClose }) {
  const [category, setCategory] = useState("");
  const dispatch = useDispatch();
  const restaurantId = useSelector((state) => state.user.userId);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (category.trim()) {
      dispatch(addCategory({ restaurantId, name: category }));
      setCategory("");
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
        />
        <button type="submit">Save</button>
        <button type="button" onClick={onClose}>
          Cancel
        </button>
      </form>
    </Modal>
  );
}

export default CreateCategoryModal;
