import { useState } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";

function CreateItem({ onSave }) {
  const navigate = useNavigate();
  const categories = useSelector((state) => state.menu.categories);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [price, setPrice] = useState("");
  const [coverImage, setCoverImage] = useState("");
  const [images, setImages] = useState([]);
  const [selectedCategories, setSelectedCategories] = useState([]);
  const [ingredients, setIngredients] = useState([]);
  const [isAvailable, setIsAvailable] = useState(true);
  const [availability, setAvailability] = useState("");
  const [allergens, setAllergens] = useState([]);
  const [tags, setTags] = useState([]);
  const [isVeg, setIsVeg] = useState(false);
  const [isHalal, setIsHalal] = useState(false);

  const handleCategoryChange = (category) => {
    setSelectedCategories((prev) =>
      prev.includes(category)
        ? prev.filter((cat) => cat !== category)
        : [...prev, category]
    );
  };

  const handleInputChange = (e, setState) => {
    setState(e.target.value.split(",").map((item) => item.trim()));
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const newItem = {
      id: Date.now(),
      name,
      description,
      price,
      coverImage,
      images: images.map((url) => ({ url })),
      categories: selectedCategories,
      ingredients,
      isAvailable,
      availability,
      allergens,
      tags,
      isVeg,
      isHalal,
    };
    onSave(newItem);
  };

  return (
    <div className="create-menu-item-page">
      <h2>Create Menu Item</h2>
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
        <input
          type="text"
          placeholder="Cover Image URL"
          value={coverImage}
          onChange={(e) => setCoverImage(e.target.value)}
        />
        <input
          type="text"
          placeholder="Images URLs (comma separated)"
          value={images.join(", ")}
          onChange={(e) => handleInputChange(e, setImages)}
        />
        <div>
          <h3>Categories</h3>
          {categories &&
            categories?.map((cat) => (
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
          placeholder="Ingredients (comma separated)"
          value={ingredients.join(", ")}
          onChange={(e) => handleInputChange(e, setIngredients)}
        />
        <input
          type="text"
          placeholder="Allergens (comma separated)"
          value={allergens.join(", ")}
          onChange={(e) => handleInputChange(e, setAllergens)}
        />
        <input
          type="text"
          placeholder="Tags (comma separated)"
          value={tags.join(", ")}
          onChange={(e) => handleInputChange(e, setTags)}
        />
        <input
          type="text"
          placeholder="Availability"
          value={availability}
          onChange={(e) => setAvailability(e.target.value)}
        />
        <label>
          Available:
          <input
            type="checkbox"
            checked={isAvailable}
            onChange={(e) => setIsAvailable(e.target.checked)}
          />
        </label>
        <label>
          Vegetarian:
          <input
            type="checkbox"
            checked={isVeg}
            onChange={(e) => setIsVeg(e.target.checked)}
          />
        </label>
        <label>
          Halal:
          <input
            type="checkbox"
            checked={isHalal}
            onChange={(e) => setIsHalal(e.target.checked)}
          />
        </label>
        <button type="submit">Save</button>
        <button type="button" onClick={() => navigate(-1)}>
          Cancel
        </button>
      </form>
    </div>
  );
}

export default CreateItem;
