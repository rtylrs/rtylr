import { useState, useEffect } from "react";
import Modal from "react-modal";

function EditMenuItemModal({ isOpen, onClose, item, categories, onSave }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [price, setPrice] = useState("");
  const [selectedCategories, setSelectedCategories] = useState([]);
  const [allergens, setAllergens] = useState([]);

  useEffect(() => {
    if (item) {
      setName(item.name);
      setDescription(item.description);
      setPrice(item.price);
      setSelectedCategories(item.categories);
      setAllergens(item.allergens);
    }
  }, [item]);

  const handleCategoryChange = (category) => {
    setSelectedCategories((prev) =>
      prev.includes(category)
        ? prev.filter((cat) => cat !== category)
        : [...prev, category]
    );
  };

  const handleAllergenChange = (e) => {
    const { value } = e.target;
    setAllergens(value.split(",").map((a) => a.trim()));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const updatedItem = {
      ...item,
      name,
      description,
      price,
      categories: selectedCategories,
      allergens,
    };
    onSave(updatedItem);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onRequestClose={onClose}
      className="modal"
      overlayClassName="modal-overlay"
    >
      <h2>Edit Menu Item</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input
          type="text"
          placeholder="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <input
          type="text"
          placeholder="Price"
          value={price}
          onChange={(e) => setPrice(e.target.value)}
        />
        <div>
          <h3>Categories</h3>
          {categories.map((cat) => (
            <label key={cat}>
              <input
                type="checkbox"
                value={cat}
                checked={selectedCategories.includes(cat)}
                onChange={() => handleCategoryChange(cat)}
              />
              {cat}
            </label>
          ))}
        </div>
        <input
          type="text"
          placeholder="Allergens (comma separated)"
          value={allergens.join(", ")}
          onChange={handleAllergenChange}
        />
        <button type="submit">Save</button>
        <button type="button" onClick={onClose}>
          Cancel
        </button>
      </form>
    </Modal>
  );
}

export default EditMenuItemModal;
