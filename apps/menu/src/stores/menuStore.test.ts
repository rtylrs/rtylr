import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useMenuStore } from './menuStore';

const { getState, setState } = useMenuStore;

// Mock the api module
vi.mock('../shared/lib/api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

beforeEach(() => {
  setState({
    id: null,
    items: [],
    allMenus: [],
    activeMenuId: null,
    status: 'idle',
    error: null,
  });
  vi.clearAllMocks();
});

describe('menuStore', () => {
  it('has correct initial state', () => {
    const s = getState();
    expect(s.id).toBeNull();
    expect(s.items).toEqual([]);
    expect(s.allMenus).toEqual([]);
    expect(s.activeMenuId).toBeNull();
    expect(s.status).toBe('idle');
    expect(s.error).toBeNull();
  });

  it('fetchAllMenus sets allMenus on success', async () => {
    const api = (await import('../shared/lib/api')).default;
    const menus = [{ _id: 'm1', items: [] }];
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ data: menus });

    await getState().fetchAllMenus({ restaurantId: 'r1' });
    expect(getState().allMenus).toEqual(menus);
    expect(getState().status).toBe('succeeded');
  });

  it('fetchAllMenus handles "No menus found" response', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      data: { message: 'No menus found' },
    });

    await getState().fetchAllMenus({ restaurantId: 'r1' });
    expect(getState().allMenus).toEqual([]);
    expect(getState().status).toBe('succeeded');
  });

  it('fetchAllMenus sets error on failure', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.get as ReturnType<typeof vi.fn>).mockRejectedValueOnce({
      response: { data: 'Server error' },
    });

    await getState().fetchAllMenus({ restaurantId: 'r1' });
    expect(getState().status).toBe('failed');
    expect(getState().error).toBe('Server error');
  });

  it('fetchMenu sets items and id on success', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      data: { _id: 'm1', items: [{ _id: 'c1', name: 'Drinks', type: 'category', menuItems: [] }] },
    });

    await getState().fetchMenu({ userId: 'u1', disabled: false });
    expect(getState().id).toBe('m1');
    expect(getState().items).toHaveLength(1);
    expect(getState().status).toBe('succeeded');
  });

  it('fetchMenu sets error on failure', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.get as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Not found'));

    await getState().fetchMenu({ userId: 'u1', disabled: false });
    expect(getState().status).toBe('failed');
    expect(getState().items).toEqual([]);
    expect(getState().id).toBeNull();
  });

  it('addMenuItem appends item to the correct category', async () => {
    const api = (await import('../shared/lib/api')).default;
    const newItem = { _id: 'i1', name: 'Latte' };
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ data: newItem });

    setState({
      items: [{ _id: 'c1', name: 'Drinks', type: 'category' as const, menuItems: [] }],
    });

    await getState().addMenuItem({ categoryId: 'c1', item: { name: 'Latte' } });
    const cat = getState().items[0] as any;
    expect(cat.menuItems).toHaveLength(1);
    expect(cat.menuItems[0].name).toBe('Latte');
  });

  it('deleteMenuItem removes item from category', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.delete as ReturnType<typeof vi.fn>).mockResolvedValueOnce({});

    setState({
      items: [
        {
          _id: 'c1',
          name: 'Drinks',
          type: 'category' as const,
          menuItems: [{ _id: 'i1', name: 'Latte' }],
        },
      ],
    });

    await getState().deleteMenuItem({ categoryId: 'c1', itemId: 'i1' });
    const cat = getState().items[0] as any;
    expect(cat.menuItems).toHaveLength(0);
  });

  it('addCategory appends new category', async () => {
    const api = (await import('../shared/lib/api')).default;
    const newCat = { _id: 'c2', name: 'Food', type: 'category', menuItems: [] };
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ data: newCat });

    await getState().addCategory({ name: 'Food' });
    expect(getState().items).toHaveLength(1);
  });

  it('deleteCategory removes category from items', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.delete as ReturnType<typeof vi.fn>).mockResolvedValueOnce({});

    setState({
      items: [{ _id: 'c1', name: 'Drinks', type: 'category' as const, menuItems: [] }],
    });

    await getState().deleteCategory('c1');
    expect(getState().items).toHaveLength(0);
  });

  it('setActiveMenuId updates activeMenuId and clears items', async () => {
    const api = (await import('../shared/lib/api')).default;
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValueOnce({});

    setState({ items: [{ _id: 'c1', name: 'Old', type: 'category' as const, menuItems: [] }] });
    await getState().setActiveMenuId({ restaurantId: 'r1', menuId: 'm2' });
    expect(getState().activeMenuId).toBe('m2');
    expect(getState().id).toBe('m2');
    expect(getState().items).toEqual([]);
  });

  it('createNewMenu appends to allMenus', async () => {
    const api = (await import('../shared/lib/api')).default;
    const newMenu = { _id: 'm2', items: [] };
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ data: newMenu });

    await getState().createNewMenu({ restaurantId: 'r1', menuName: 'Lunch' });
    expect(getState().allMenus).toHaveLength(1);
  });
});
