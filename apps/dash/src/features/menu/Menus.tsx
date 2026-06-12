import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { UtensilsCrossed, Search, MoreHorizontal } from 'lucide-react';
import { useUserStore } from '../../shared/stores/userStore';
import { useGetAllMenus } from '../../shared/hooks/queries/useMenu';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
  Badge,
  Button,
  Input,
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '../../shared/components/ui';

interface MenuEntry {
  _id: string;
  menuName?: string;
  name?: string;
  items?: unknown[];
  itemCount?: number;
  isActive?: boolean;
  active?: boolean;
}

const Menus: React.FC = () => {
  const navigate = useNavigate();
  const { userId } = useUserStore();
  const [searchTerm, setSearchTerm] = useState('');

  const { data: allMenus = [], isLoading } = useGetAllMenus(userId);

  const filteredMenus = (allMenus as MenuEntry[]).filter((menu) =>
    (menu.menuName || menu.name || '').toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <>
      <HelmetContainer title="Menus · rtylr" />
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-primary">Menus</h1>
            <p className="text-sm text-gray-400 mt-0.5">Manage your restaurant menus</p>
          </div>
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
                  <div
                    className="h-4 animate-pulse rounded mb-2 w-3/4"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                  <div
                    className="h-3 animate-pulse rounded w-1/2"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Menus Grid */}
        {!isLoading && filteredMenus.length > 0 && (
          <div className="grid grid-cols-3 gap-4">
            {filteredMenus.map((menu) => {
              const name = menu.menuName || menu.name || 'Untitled';
              const itemCount = menu.items?.length ?? menu.itemCount ?? 0;
              const isActive = menu.isActive || menu.active;
              return (
                <Card
                  key={menu._id}
                  className="group hover:shadow-md transition-shadow cursor-pointer"
                >
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
                            <button className="h-7 w-7 rounded-md flex items-center justify-center text-gray-400 hover:text-primary hover:bg-[var(--dash-hover)] transition-colors">
                              <MoreHorizontal size={15} />
                            </button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => navigate({ to: `/menu/${menu._id}` })}>
                              View Menu
                            </DropdownMenuItem>
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
                      onClick={() => navigate({ to: `/menu/${menu._id}` })}
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
            <div
              className="h-16 w-16 rounded-2xl flex items-center justify-center mb-4"
              style={{ background: 'var(--dash-hover)' }}
            >
              <UtensilsCrossed size={28} className="text-gray-300" />
            </div>
            <h3 className="text-sm font-semibold text-primary">No menus yet</h3>
            <p className="text-sm text-gray-400 mt-1">No POS menu items have been created yet.</p>
          </div>
        )}
      </div>
    </>
  );
};

export default Menus;
