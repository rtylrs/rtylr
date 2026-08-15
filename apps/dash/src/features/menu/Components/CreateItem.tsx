import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useRouterState } from '@tanstack/react-router';
import { Button, Input } from '../../../shared/components/ui';
import { useMenuStore } from '../../../shared/stores/menuStore';

const CreateItem: React.FC = () => {
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const { addMenuItem } = useMenuStore();
  const categoryId = new URLSearchParams(searchStr).get('categoryId') || 'uncategorized';

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [salePrice, setSalePrice] = useState('');
  const [coverImage, setCoverImage] = useState<File | null>(null);
  const [images, setImages] = useState<string[]>([]);
  const [ingredients, setIngredients] = useState<string[]>([]);
  const [isAvailable, setIsAvailable] = useState(true);
  const [availability, setAvailability] = useState('');
  const [allergens, setAllergens] = useState<string[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [isVeg, setIsVeg] = useState(false);
  const [isHalal, setIsHalal] = useState(false);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    setState: React.Dispatch<React.SetStateAction<string[]>>
  ) => {
    setState(e.target.value.split(',').map((item) => item.trim()));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const item = {
      name,
      description,
      price: price ? Number(price) * 100 : 0,
      salePrice: salePrice ? Number(salePrice) * 100 : 0,
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
      <div className="max-w-2xl mx-auto space-y-6">
        <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          Create Menu Item
        </h2>
        <form
          onSubmit={handleSubmit}
          className="rounded-xl p-6 border space-y-4"
          style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
        >
          <div className="space-y-1">
            <label
              htmlFor="item-name"
              className="text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Name *
            </label>
            <Input
              id="item-name"
              type="text"
              placeholder="Product name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="space-y-1">
            <label
              htmlFor="item-desc"
              className="text-xs font-semibold"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              Description
            </label>
            <Input
              id="item-desc"
              type="text"
              placeholder="Add a description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <label
                htmlFor="item-price"
                className="text-xs font-semibold"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                Price *
              </label>
              <Input
                id="item-price"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1">
              <label
                htmlFor="item-sale-price"
                className="text-xs font-semibold"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                Sale Price
              </label>
              <Input
                id="item-sale-price"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={salePrice}
                onChange={(e) => setSalePrice(e.target.value)}
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Tags (comma separated)
            </label>
            <Input
              type="text"
              placeholder="Popular, Chef Special"
              value={tags.join(', ')}
              onChange={(e) => handleInputChange(e, setTags)}
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Cover Image
            </label>
            <Input
              type="file"
              onChange={(e) => setCoverImage(e.target.files ? e.target.files[0] : null)}
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Ingredients (comma separated)
            </label>
            <Input
              type="text"
              placeholder="Flour, Sugar, Butter"
              value={ingredients.join(', ')}
              onChange={(e) => handleInputChange(e, setIngredients)}
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold" style={{ color: 'var(--dash-text-muted)' }}>
              Allergens (comma separated)
            </label>
            <Input
              type="text"
              placeholder="Nuts, Dairy, Gluten"
              value={allergens.join(', ')}
              onChange={(e) => handleInputChange(e, setAllergens)}
            />
          </div>

          <div className="flex items-center gap-6 pt-2">
            <label
              className="flex items-center gap-2 text-xs font-medium cursor-pointer select-none"
              style={{ color: 'var(--dash-text)' }}
            >
              <input
                type="checkbox"
                checked={isAvailable}
                onChange={(e) => setIsAvailable(e.target.checked)}
                className="rounded"
              />
              Available
            </label>
            <label
              className="flex items-center gap-2 text-xs font-medium cursor-pointer select-none"
              style={{ color: 'var(--dash-text)' }}
            >
              <input
                type="checkbox"
                checked={isVeg}
                onChange={(e) => setIsVeg(e.target.checked)}
                className="rounded"
              />
              Vegetarian
            </label>
            <label
              className="flex items-center gap-2 text-xs font-medium cursor-pointer select-none"
              style={{ color: 'var(--dash-text)' }}
            >
              <input
                type="checkbox"
                checked={isHalal}
                onChange={(e) => setIsHalal(e.target.checked)}
                className="rounded"
              />
              Halal
            </label>
          </div>

          <div
            className="flex items-center justify-end gap-3 pt-4 border-t"
            style={{ borderColor: 'var(--dash-border)' }}
          >
            <Button variant="ghost" type="button" onClick={() => window.history.back()}>
              Cancel
            </Button>
            <Button variant="default" type="submit">
              Save Item
            </Button>
          </div>
        </form>
      </div>
    </>
  );
};

export default CreateItem;
