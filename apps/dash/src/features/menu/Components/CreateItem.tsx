import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Button, Input } from '../../../shared/components/ui';
import { useMenuStore } from '../../../shared/stores/menuStore';
import UnavailableNotice from '../../../shared/components/dash/UnavailableNotice';

const UNAVAILABLE_ID = 'menu-item-saving-unavailable';

interface StoreItemDetailsLike {
  name?: string;
  description?: string;
  price?: string | number;
  salePrice?: string | number;
  isAvailable?: boolean;
  ingredients?: string[];
  allergens?: string[];
  tags?: string[];
  isVeg?: boolean;
  isHalal?: boolean;
  menuItems?: StoreItemLike[];
}

interface StoreItemLike {
  _id?: string;
  itemId?: string;
  type?: string;
  name?: string;
  details?: StoreItemDetailsLike;
}

const CreateItem: React.FC = () => {
  const { t } = useTranslation();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  // Only `items` is read now: the two write actions were the false-success
  // path. They stay in the store for whoever wires the real mutations.
  const { items } = useMenuStore();
  const searchParams = new URLSearchParams(searchStr);
  const itemId = searchParams.get('itemId');

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [price, setPrice] = useState('');
  const [salePrice, setSalePrice] = useState('');
  const [coverImage, setCoverImage] = useState<File | null>(null);
  const [images] = useState<string[]>([]);
  const [ingredients, setIngredients] = useState<string[]>([]);
  const [isAvailable, setIsAvailable] = useState(true);
  const [availability] = useState('');
  const [allergens, setAllergens] = useState<string[]>([]);
  const [tags, setTags] = useState<string[]>([]);
  const [isVeg, setIsVeg] = useState(false);
  const [isHalal, setIsHalal] = useState(false);

  React.useEffect(() => {
    if (!itemId) return;
    let found: StoreItemLike | undefined = items.find(
      (i) => (i as StoreItemLike)._id === itemId || (i as StoreItemLike).itemId === itemId
    ) as StoreItemLike | undefined;
    if (!found) {
      for (const cat of items) {
        const candidate = cat as StoreItemLike;
        if (candidate.type === 'category' && candidate.details?.menuItems) {
          found = candidate.details.menuItems.find((i) => i._id === itemId || i.itemId === itemId);
          if (found) break;
        }
      }
    }
    if (found) {
      const details = found.details;
      setName(details?.name || found.name || '');
      setDescription(details?.description || '');
      if (details?.price !== undefined) {
        setPrice(String(Number(details.price)));
      }
      if (details?.salePrice) {
        setSalePrice(String(Number(details.salePrice)));
      }
      if (details?.isAvailable !== undefined) setIsAvailable(Boolean(details.isAvailable));
      if (details?.ingredients) setIngredients(details.ingredients);
      if (details?.allergens) setAllergens(details.allergens);
      if (details?.tags) setTags(details.tags);
      if (details?.isVeg !== undefined) setIsVeg(Boolean(details.isVeg));
      if (details?.isHalal !== undefined) setIsHalal(Boolean(details.isHalal));
    }
  }, [itemId, items]);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    setState: React.Dispatch<React.SetStateAction<string[]>>
  ) => {
    setState(e.target.value.split(',').map((item) => item.trim()));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    const trimmedName = name.trim();
    if (!trimmedName) {
      toast.error('Please enter an item name');
      return;
    }

    const numericPrice = Number(price);
    if (isNaN(numericPrice) || numericPrice < 0) {
      toast.error('Please enter a valid non-negative price');
      return;
    }

    const item = {
      name: trimmedName,
      description,
      // Major units, matching pos_menu_items.sale_price DECIMAL(10,2) and
      // what useMenu's toStoreItems reads back. Scaling by 100 here meant the
      // form seeded itself by dividing by 100 to compensate, so the field
      // showed 0.125 for a $12.50 item.
      price: numericPrice,
      salePrice: salePrice ? Number(salePrice) : 0,
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

    // Saving is off. These two calls change only the local store, and
    // useGetMenu refetches over the top, so the success toast was telling
    // someone their menu item had been created when nothing had left the
    // browser. Reporting the state is the honest behaviour until the
    // mutations reach pos/menu_item/v1. rtylr#1150.
    void item;
    toast.error(t('menu.savingUnavailableToast'));
  };

  return (
    <>
      <HelmetContainer title={itemId ? 'Edit item · rtylr' : 'Create item · rtylr'} />
      <div className="max-w-2xl mx-auto space-y-6">
        <UnavailableNotice>
          <span id={UNAVAILABLE_ID}>{t('menu.savingUnavailable')}</span>
        </UnavailableNotice>
        <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
          {itemId ? 'Edit Menu Item' : 'Create Menu Item'}
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
              {t('common.cancel')}
            </Button>
            {/* Disabled as well as guarded in the handler: a disabled submit
                button does not stop Enter submitting the form from a field. */}
            <Button variant="default" type="submit" disabled aria-describedby={UNAVAILABLE_ID}>
              {t('menu.saveItem')}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
};

export default CreateItem;
