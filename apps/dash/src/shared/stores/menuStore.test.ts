import { describe, it, expect, beforeEach } from 'vitest';
import { useMenuStore } from './menuStore';

const { getState } = useMenuStore;

beforeEach(() => {
  useMenuStore.setState({
    id: null,
    items: [],
    allMenus: [],
    activeMenuId: null,
    status: 'idle',
  });
});

describe('menuStore', () => {
  it('has correct initial state', () => {
    expect(getState().id).toBeNull();
    expect(getState().items).toEqual([]);
    expect(getState().allMenus).toEqual([]);
    expect(getState().activeMenuId).toBeNull();
    expect(getState().status).toBe('idle');
  });

  it('setId updates id', () => {
    getState().setId('menu-1');
    expect(getState().id).toBe('menu-1');
  });

  it('setId with null', () => {
    getState().setId('menu-1');
    getState().setId(null);
    expect(getState().id).toBeNull();
  });

  it('setItems updates items', () => {
    const items = [{ _id: 'c1', name: 'Appetizers', type: 'category' as const, menuItems: [] }];
    getState().setItems(items);
    expect(getState().items).toEqual(items);
  });

  it('setAllMenus updates menus list', () => {
    const menus = [{ _id: 'm1', items: [] }];
    getState().setAllMenus(menus);
    expect(getState().allMenus).toEqual(menus);
  });

  it('setActiveMenuId updates active menu', () => {
    getState().setActiveMenuId('m1');
    expect(getState().activeMenuId).toBe('m1');
  });

  it('addCategory appends a category', () => {
    const category = { _id: 'c1', name: 'Main', type: 'category' as const, menuItems: [] };
    getState().addCategory(category);
    expect(getState().items).toHaveLength(1);
    expect(getState().items[0]).toEqual(category);
  });

  it('addCategory appends multiple categories', () => {
    getState().addCategory({ _id: 'c1', name: 'A', type: 'category' as const, menuItems: [] });
    getState().addCategory({ _id: 'c2', name: 'B', type: 'category' as const, menuItems: [] });
    expect(getState().items).toHaveLength(2);
  });

  it('editCategory updates a category matched by _id', () => {
    getState().addCategory({ _id: 'c1', name: 'A', type: 'category' as const, menuItems: [] });
    getState().editCategory('c1', { name: 'B' });
    expect(getState().items[0]).toMatchObject({ _id: 'c1', name: 'B' });
  });

  it('editCategory matches legacy id and itemId aliases', () => {
    getState().addCategory({
      _id: 'c1',
      id: 'legacy-1',
      name: 'A',
      type: 'category' as const,
      menuItems: [],
    });
    getState().addCategory({
      _id: 'c2',
      itemId: 'legacy-2',
      name: 'C',
      type: 'category' as const,
      menuItems: [],
    });
    getState().editCategory('legacy-1', { name: 'B' });
    getState().editCategory('legacy-2', { name: 'D' });
    expect(getState().items[0]).toMatchObject({ _id: 'c1', name: 'B' });
    expect(getState().items[1]).toMatchObject({ _id: 'c2', name: 'D' });
  });

  it('editCategory leaves non-matching categories untouched', () => {
    getState().addCategory({ _id: 'c1', name: 'A', type: 'category' as const, menuItems: [] });
    getState().editCategory('nope', { name: 'B' });
    expect(getState().items[0]).toMatchObject({ _id: 'c1', name: 'A' });
  });

  it('deleteCategory removes a category', () => {
    getState().addCategory({
      _id: 'c1',
      name: 'A',
      type: 'category' as const,
      menuItems: [],
      id: 'c1',
    });
    getState().deleteCategory('c1');
    expect(getState().items).toHaveLength(0);
  });

  it('addMenuItem inserts into the matching category instead of top level', async () => {
    getState().addCategory({ _id: 'c1', name: 'A', type: 'category' as const, menuItems: [] });
    await getState().addMenuItem('c1', { name: 'Falafel' });
    expect(getState().items).toHaveLength(1);
    const category = getState().items[0];
    expect(category).toMatchObject({ _id: 'c1' });
    if (category.type !== 'category') throw new Error('expected category');
    expect(category.details?.menuItems).toHaveLength(1);
    expect(category.details?.menuItems?.[0]).toMatchObject({ name: 'Falafel' });
    expect(category.menuItems).toHaveLength(1);
  });

  it('addMenuItem falls back to top level for unknown categories', async () => {
    await getState().addMenuItem('uncategorized', { name: 'Soup' });
    expect(getState().items).toHaveLength(1);
    expect(getState().items[0]).toMatchObject({ name: 'Soup', type: 'item' });
  });

  it('deleteMenuItem removes nested items from both storage shapes', () => {
    getState().setItems([
      {
        _id: 'c1',
        name: 'A',
        type: 'category' as const,
        menuItems: [{ _id: 'm1', name: 'Kebab', type: 'item' as const }],
        details: { menuItems: [{ _id: 'm1', name: 'Kebab', type: 'item' as const }] },
      },
    ]);
    getState().deleteMenuItem('m1');
    const category = getState().items[0];
    if (category.type !== 'category') throw new Error('expected category');
    expect(category.menuItems).toHaveLength(0);
    expect(category.details?.menuItems).toHaveLength(0);
  });

  it('deleteMenuItem leaves other nested items untouched', () => {
    getState().setItems([
      {
        _id: 'c1',
        name: 'A',
        type: 'category' as const,
        menuItems: [
          { _id: 'm1', name: 'Kebab', type: 'item' as const },
          { _id: 'm2', name: 'Salad', type: 'item' as const },
        ],
      },
    ]);
    getState().deleteMenuItem('m1');
    const category = getState().items[0];
    if (category.type !== 'category') throw new Error('expected category');
    expect(category.menuItems.map((i) => i._id)).toEqual(['m2']);
  });

  it('addBanner appends a banner', () => {
    const banner = { _id: 'b1', bannerImage: 'img.jpg', menuId: 'm1' };
    getState().addBanner(banner);
    expect(getState().items).toHaveLength(1);
    expect(getState().items[0]).toEqual(banner);
  });

  it('fetchMenu sets status to loading then success', async () => {
    await getState().fetchMenu('user1');
    expect(getState().status).toBe('success');
  });

  it('deleteCategory does not remove banners', () => {
    const banner = { _id: 'b1', bannerImage: 'img.jpg', menuId: 'm1' };
    getState().addBanner(banner);
    getState().deleteCategory('b1');
    expect(getState().items).toHaveLength(1);
  });
});

describe('menuStore.updateMenuItem', () => {
  it('updates a top-level item matched by _id and merges details', () => {
    getState().setItems([
      {
        _id: 'm1',
        name: 'Kebab',
        type: 'item' as const,
        details: { name: 'Kebab', description: 'grilled', price: 500 },
      },
    ]);
    getState().updateMenuItem('m1', { name: 'Shish', details: { name: 'Shish', price: 700 } });
    const item = getState().items[0];
    expect(item.name).toBe('Shish');
    // description was not in the update and must survive the merge.
    expect(item.details).toEqual({ name: 'Shish', description: 'grilled', price: 700 });
  });

  it('updates a top-level item matched by the legacy itemId alias', () => {
    getState().setItems([
      { _id: 'server-1', itemId: 'legacy-1', name: 'Kebab', type: 'item' as const },
    ]);
    getState().updateMenuItem('legacy-1', { name: 'Shish' });
    expect(getState().items[0].name).toBe('Shish');
  });

  it('synthesizes details from the name when the item carries none', () => {
    getState().setItems([{ _id: 'm1', name: 'Kebab', type: 'item' as const }]);
    getState().updateMenuItem('m1', { details: { name: 'Kebab', price: 100 } });
    expect(getState().items[0].details).toEqual({ name: 'Kebab', price: 100 });
  });

  it('updates a nested item under details.menuItems matched by _id', () => {
    getState().setItems([
      {
        _id: 'c1',
        name: 'Grills',
        type: 'category' as const,
        menuItems: [],
        details: {
          menuItems: [
            {
              _id: 'm1',
              name: 'Kebab',
              type: 'item' as const,
              details: { name: 'Kebab', sku: 'K1' },
            },
            { _id: 'm2', name: 'Salad', type: 'item' as const },
          ],
        },
      },
    ]);
    getState().updateMenuItem('m1', { name: 'Shish', details: { name: 'Shish' } });
    const category = getState().items[0];
    if (category.type !== 'category') throw new Error('expected category');
    const nested = category.details?.menuItems ?? [];
    expect(nested[0].name).toBe('Shish');
    // sku was not in the update and must survive the nested merge.
    expect(nested[0].details).toEqual({ name: 'Shish', sku: 'K1' });
    expect(nested[1].name).toBe('Salad');
  });

  it('updates a nested item matched by the legacy itemId alias', () => {
    getState().setItems([
      {
        _id: 'c1',
        name: 'Grills',
        type: 'category' as const,
        menuItems: [],
        details: {
          menuItems: [
            { _id: 'server-1', itemId: 'legacy-1', name: 'Kebab', type: 'item' as const },
          ],
        },
      },
    ]);
    getState().updateMenuItem('legacy-1', { name: 'Shish' });
    const category = getState().items[0];
    if (category.type !== 'category') throw new Error('expected category');
    expect(category.details?.menuItems?.[0].name).toBe('Shish');
  });

  it('leaves a category alone when its own id collides with the target', () => {
    // The top-level pass used to match any entry by _id, cast it to MenuItem
    // and replace details with { name }, wiping details.menuItems. Only a
    // menu item may be updated here.
    getState().setItems([
      {
        _id: 'shared-id',
        name: 'Grills',
        type: 'category' as const,
        menuItems: [{ _id: 'm1', name: 'Kebab', type: 'item' as const }],
        details: { menuItems: [{ _id: 'm1', name: 'Kebab', type: 'item' as const }] },
      },
    ]);
    getState().updateMenuItem('shared-id', { name: 'Clobbered' });
    const category = getState().items[0];
    if (category.type !== 'category') throw new Error('expected category');
    expect(category.name).toBe('Grills');
    expect(category.details?.menuItems).toHaveLength(1);
    expect(category.menuItems).toHaveLength(1);
  });

  it('leaves a banner alone when its id collides with the target', () => {
    getState().setItems([{ _id: 'shared-id', bannerImage: 'a.png', menuId: 'menu-1' }]);
    getState().updateMenuItem('shared-id', { name: 'Clobbered' });
    expect(getState().items[0]).toEqual({
      _id: 'shared-id',
      bannerImage: 'a.png',
      menuId: 'menu-1',
    });
  });

  it('leaves every entry untouched when nothing matches', () => {
    const items = [
      { _id: 'm1', name: 'Kebab', type: 'item' as const },
      { _id: 'c1', name: 'Grills', type: 'category' as const, menuItems: [] },
    ];
    getState().setItems(items);
    getState().updateMenuItem('missing', { name: 'Shish' });
    expect(getState().items).toEqual(items);
  });
});
