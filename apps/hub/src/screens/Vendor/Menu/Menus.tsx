import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, UtensilsCrossed, Search, MoreHorizontal, CheckCircle } from 'lucide-react';
import { useUserStore } from '../../../stores/userStore';
import { useGetAllMenus, useCreateMenu, useSetActiveMenu } from '../../../hooks/queries/useMenu';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Badge,
  Button,
  Input,
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
  Label,
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '../../../components/ui';

const Menus: React.FC = () => {
  const navigate = useNavigate();
  const { userId } = useUserStore();
  const [searchTerm, setSearchTerm] = useState('');
  const [newMenuName, setNewMenuName] = useState('');

  const { data: allMenus = [], isLoading } = useGetAllMenus(userId);
  const createMenu = useCreateMenu();
  const setActiveMenu = useSetActiveMenu();

  const filteredMenus = (allMenus as any[]).filter((menu: any) =>
    (menu.menuName || menu.name || '').toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleCreate = () => {
    if (!newMenuName.trim() || !userId) return;
    createMenu.mutate({ restaurantId: userId, menuName: newMenuName }, {
      onSuccess: () => setNewMenuName(''),
    });
  };

  const handleSetActive = (menuId: string) => {
    if (!userId) return;
    setActiveMenu.mutate({ restaurantId: userId, menuId });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-primary">Menus</h1>
          <p className="text-sm text-gray-400 mt-0.5">Manage your restaurant menus</p>
        </div>
        <Dialog>
          <DialogTrigger>
            <Button size="sm">
              <Plus size={14} />
              New Menu
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Menu</DialogTitle>
              <DialogDescription>Give your menu a name to get started.</DialogDescription>
            </DialogHeader>
            <div className="p-6 pt-2 pb-4 space-y-3">
              <Label htmlFor="menuName">Menu Name</Label>
              <Input
                id="menuName"
                placeholder="e.g. Lunch Menu, Drinks, Specials..."
                value={newMenuName}
                onChange={(e) => setNewMenuName(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') handleCreate(); }}
              />
            </div>
            <DialogFooter>
              <DialogClose>
                <Button variant="outline">Cancel</Button>
              </DialogClose>
              <Button onClick={handleCreate} disabled={!newMenuName.trim() || createMenu.isPending}>
                {createMenu.isPending ? 'Creating...' : 'Create Menu'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Search */}
      <div className="relative max-w-sm">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <Input
          className="pl-9"
          placeholder="Search menus..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="grid grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <Card key={i}>
              <CardContent className="pt-5">
                <div className="h-4 animate-pulse rounded mb-2 w-3/4" style={{ background: 'var(--hub-hover)' }} />
                <div className="h-3 animate-pulse rounded w-1/2" style={{ background: 'var(--hub-hover)' }} />
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Menus Grid */}
      {!isLoading && filteredMenus.length > 0 && (
        <div className="grid grid-cols-3 gap-4">
          {filteredMenus.map((menu: any) => {
            const name = menu.menuName || menu.name || 'Untitled';
            const itemCount = menu.items?.length ?? menu.itemCount ?? 0;
            const isActive = menu.isActive || menu.active;
            return (
              <Card key={menu._id} className="group hover:shadow-md transition-shadow cursor-pointer">
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <CardTitle className="truncate">{name}</CardTitle>
                      <CardDescription className="mt-1">{itemCount} items</CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                      {isActive && <Badge variant="success">Active</Badge>}
                      <DropdownMenu>
                        <DropdownMenuTrigger>
                          <button className="h-7 w-7 rounded-md flex items-center justify-center text-gray-400 hover:text-primary hover:bg-[var(--hub-hover)] transition-colors">
                            <MoreHorizontal size={15} />
                          </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => navigate(`/app/business/menu/${menu._id}`)}>
                            View Menu
                          </DropdownMenuItem>
                          {!isActive && (
                            <DropdownMenuItem onClick={() => handleSetActive(menu._id)}>
                              <CheckCircle size={14} />
                              Set as Active
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full"
                    onClick={() => navigate(`/app/business/menu/${menu._id}`)}
                  >
                    View Menu
                  </Button>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && filteredMenus.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <div className="h-16 w-16 rounded-2xl flex items-center justify-center mb-4" style={{ background: 'var(--hub-hover)' }}>
            <UtensilsCrossed size={28} className="text-gray-300" />
          </div>
          <h3 className="text-sm font-semibold text-primary">No menus yet</h3>
          <p className="text-sm text-gray-400 mt-1 mb-4">Create your first menu to get started</p>
          <Dialog>
            <DialogTrigger>
              <Button size="sm">
                <Plus size={14} />
                Create your first menu
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create New Menu</DialogTitle>
                <DialogDescription>Give your menu a name to get started.</DialogDescription>
              </DialogHeader>
              <div className="p-6 pt-2 pb-4 space-y-3">
                <Label htmlFor="menuName2">Menu Name</Label>
                <Input
                  id="menuName2"
                  placeholder="e.g. Lunch Menu, Drinks..."
                  value={newMenuName}
                  onChange={(e) => setNewMenuName(e.target.value)}
                />
              </div>
              <DialogFooter>
                <DialogClose>
                  <Button variant="outline">Cancel</Button>
                </DialogClose>
                <Button onClick={handleCreate} disabled={!newMenuName.trim()}>Create Menu</Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      )}
    </div>
  );
};

export default Menus;
