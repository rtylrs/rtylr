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
