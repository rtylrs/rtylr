import { useEffect, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useMenuStore } from '../../../shared/stores/menuStore';

function CreateItem() {
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const { addMenuItem } = useMenuStore();
  const categoryId = new URLSearchParams(searchStr).get('categoryId');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [coverImage, setCoverImage] = useState(null);
  const [images, setImages] = useState([]);
  const [ingredients, setIngredients] = useState([]);
  const [isAvailable, setIsAvailable] = useState(true);
  const [availability, setAvailability] = useState('');
  const [allergens, setAllergens] = useState([]);
  const [tags, setTags] = useState([]);
  const [isVeg, setIsVeg] = useState(false);
  const [isHalal, setIsHalal] = useState(false);

  useEffect(() => {
    if (!categoryId) {
      window.history.back();
    }
  }, [categoryId]);

  const handleInputChange = (e, setState) => {
    setState(e.target.value.split(',').map((item) => item.trim()));
  };

  const handleFileChange = (e, setState) => {
    setState(e.target.files[0]);
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    const item = {
      name,
      description,
      price,
      coverImage,
      images,
      ingredients,
      isAvailable,
      availability,
      allergens,
      tags,
      isVeg,
      isHalal,
    };

    addMenuItem(categoryId, item).then(() => {
      window.history.back();
    });
  };

  return (
    <>
      <HelmetContainer title="Create item · rtylr" />
      <div className="create-menu-item-page">
        <h2>Create Menu Item</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Name *</label>
            <input
              type="text"
              placeholder="Product name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Description</label>
            <input
              type="text"
              placeholder="Add a description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Price *</label>
            <input
              type="number"
              placeholder="0.00"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Lower Sale Price</label>
            <input
              type="number"
              placeholder="0.00"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>Tags</label>
            <input
              type="text"
              placeholder=""
              value={tags.join(', ')}
              onChange={(e) => handleInputChange(e, setTags)}
            />
          </div>
          <div className="form-group">
            <label>Cover Image</label>
            <input type="file" onChange={(e) => handleFileChange(e, setCoverImage)} />
          </div>
          <div className="form-group">
            <label>Images URLs (comma separated)</label>
            <input
              type="text"
              value={images.join(', ')}
              onChange={(e) => handleInputChange(e, setImages)}
            />
          </div>
          <div className="form-group">
            <label>Ingredients (comma separated)</label>
            <input
              type="text"
              value={ingredients.join(', ')}
              onChange={(e) => handleInputChange(e, setIngredients)}
            />
          </div>
          <div className="form-group">
            <label>Allergens (comma separated)</label>
            <input
              type="text"
              value={allergens.join(', ')}
              onChange={(e) => handleInputChange(e, setAllergens)}
            />
          </div>
          <div className="form-group">
            <label>Availability</label>
            <input
              type="text"
              value={availability}
              onChange={(e) => setAvailability(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label>
              Available:
              <input
                type="checkbox"
                checked={isAvailable}
                onChange={(e) => setIsAvailable(e.target.checked)}
              />
            </label>
          </div>
          <div className="form-group">
            <label>
              Vegetarian:
              <input type="checkbox" checked={isVeg} onChange={(e) => setIsVeg(e.target.checked)} />
            </label>
          </div>
          <div className="form-group">
            <label>
              Halal:
              <input
                type="checkbox"
                checked={isHalal}
                onChange={(e) => setIsHalal(e.target.checked)}
              />
            </label>
          </div>
          <button type="submit">Save</button>
          <button type="button" onClick={() => window.history.back()}>
            Cancel
          </button>
        </form>
      </div>
    </>
  );
}

export default CreateItem;
