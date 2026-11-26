import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Button, Input } from '../../../shared/components/ui';
import { useMenuStore } from '../../../shared/stores/menuStore';
import { useCreateMenuItem, useUpdateMenuItem } from '../../../shared/hooks/queries/useMenu';

interface StoreItemDetailsLike {
  name?: string;
  description?: string;
  price?: string | number;
  sku?: string;
  stock?: number;
  isActive?: boolean;
  isAvailable?: boolean;
  menuItems?: StoreItemLike[];
}

interface StoreItemLike {
  _id?: string;
  itemId?: string;
  type?: string;
  name?: string;
  details?: StoreItemDetailsLike;
}

function generateSku(name: string): string {
  const base = name
    .toUpperCase()
    .replace(/[^A-Z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 12);
  const suffix = Math.random().toString(36).slice(2, 6).toUpperCase();
  return `${base || 'ITEM'}-${suffix}`;
}

const CreateItem: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const { items } = useMenuStore();
  const searchParams = new URLSearchParams(searchStr);
  const itemId = searchParams.get('itemId');

  const createMenuItem = useCreateMenuItem();
  const updateMenuItem = useUpdateMenuItem();

  const [name, setName] = useState('');
  const [sku, setSku] = useState('');
  const [price, setPrice] = useState('');
  const [stock, setStock] = useState('0');
  const [isAvailable, setIsAvailable] = useState(true);

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
      if (details?.sku) setSku(details.sku);
      if (details?.price !== undefined) setPrice(String(Number(details.price)));
      if (details?.stock !== undefined) setStock(String(Number(details.stock)));
      if (details?.isActive !== undefined) setIsAvailable(Boolean(details.isActive));
      else if (details?.isAvailable !== undefined) setIsAvailable(Boolean(details.isAvailable));
    }
  }, [itemId, items]);

  const isSaving = createMenuItem.isPending || updateMenuItem.isPending;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const trimmedName = name.trim();
    if (!trimmedName) {
      toast.error(t('menu.nameRequired', 'Please enter an item name'));
      return;
    }

    const numericPrice = Number(price);
    if (!Number.isFinite(numericPrice) || numericPrice < 0) {
      toast.error(t('menu.priceInvalid', 'Please enter a valid non-negative price'));
      return;
    }

    const numericStock = Number(stock);
    const safeStock = Number.isFinite(numericStock) && numericStock >= 0 ? numericStock : 0;

    try {
      if (itemId) {
        await updateMenuItem.mutateAsync({
          id: itemId,
          name: trimmedName,
          salePrice: numericPrice,
          isActive: isAvailable,
          stock: safeStock,
        });
        toast.success(t('menu.itemUpdated', 'Menu item updated and synced across your apps'));
      } else {
        await createMenuItem.mutateAsync({
          name: trimmedName,
          sku: sku.trim() || generateSku(trimmedName),
          salePrice: numericPrice,
        });
        toast.success(t('menu.itemCreated', 'Menu item created and synced across your apps'));
      }
      void navigate({ to: '/menu', search: { menuId: 'pos-menu' } });
    } catch (err: unknown) {
      const msg =
        (err as { message?: string })?.message ||
        t('menu.itemSaveFailed', 'Failed to save menu item');
      toast.error(msg);
    }
  };

  return (
    <>
      <HelmetContainer title={itemId ? 'Edit item · rtylr' : 'Create item · rtylr'} />
      <div className="max-w-2xl mx-auto space-y-6">
        <div>
          <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
            {itemId
              ? t('menu.editItemTitle', 'Edit Menu Item')
              : t('menu.createItemTitle', 'Create Menu Item')}
          </h2>
          <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
            {t(
              'menu.syncHint',
              'Items saved here automatically stay in sync across Digital Menu, POS, and ERP Inventory.'
            )}
          </p>
        </div>

        <form
          onSubmit={(e) => void handleSubmit(e)}
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
              placeholder="e.g. Iced Matcha Latte"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
                min="0"
                placeholder="0.00"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                required
              />
            </div>

            {!itemId ? (
              <div className="space-y-1">
                <label
                  htmlFor="item-sku"
                  className="text-xs font-semibold"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  SKU (optional)
                </label>
                <Input
                  id="item-sku"
                  type="text"
                  placeholder="Auto-generated if left blank"
                  value={sku}
                  onChange={(e) => setSku(e.target.value)}
                />
              </div>
            ) : (
              <div className="space-y-1">
                <label
                  htmlFor="item-stock"
                  className="text-xs font-semibold"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  Stock
                </label>
                <Input
                  id="item-stock"
                  type="number"
                  step="1"
                  min="0"
                  placeholder="0"
                  value={stock}
                  onChange={(e) => setStock(e.target.value)}
                />
              </div>
            )}
          </div>

          {itemId && (
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
                Available on menu
              </label>
            </div>
          )}

          <div
            className="flex items-center justify-end gap-3 pt-4 border-t"
            style={{ borderColor: 'var(--dash-border)' }}
          >
            <Button variant="ghost" type="button" onClick={() => window.history.back()}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" type="submit" disabled={isSaving}>
              {isSaving ? t('common.saving', 'Saving...') : t('menu.saveItem')}
            </Button>
          </div>
        </form>
      </div>
    </>
  );
};

export default CreateItem;
