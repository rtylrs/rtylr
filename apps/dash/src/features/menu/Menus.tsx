import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { ExternalLink, Plus, Search, UtensilsCrossed } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useUserStore } from '../../shared/stores/userStore';
import { useGetAllMenus } from '../../shared/hooks/queries/useMenu';
import { Badge, Button } from '../../shared/components/ui';

interface MenuEntry {
  _id: string;
  id?: string;
  menuName?: string;
  name?: string;
  items?: unknown[];
  itemCount?: number;
  isActive?: boolean;
  active?: boolean;
  updatedAt?: string;
  createdAt?: string;
}

function menuName(menu: MenuEntry): string {
  return menu.menuName || menu.name || 'Untitled menu';
}

function menuItemCount(menu: MenuEntry): number {
  return menu.items?.length ?? menu.itemCount ?? 0;
}

function isMenuActive(menu: MenuEntry): boolean {
  return menu.isActive ?? menu.active ?? false;
}

function formatDate(value?: string): string {
  if (!value) return 'Not synced';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Not synced';
  return new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' }).format(date);
}

const Menus: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { userId } = useUserStore();
  const [searchTerm, setSearchTerm] = useState('');

  const { data: allMenus = [], isLoading } = useGetAllMenus(userId);
  const menus = allMenus as MenuEntry[];
  const normalizedSearch = searchTerm.trim().toLowerCase();

  const filteredMenus = useMemo(
    () =>
      menus.filter((menu) =>
        normalizedSearch ? menuName(menu).toLowerCase().includes(normalizedSearch) : true
      ),
    [menus, normalizedSearch]
  );

  const totalItems = menus.reduce((sum, menu) => sum + menuItemCount(menu), 0);
  const activeMenus = menus.filter(isMenuActive).length;

  return (
    <>
      <HelmetContainer title="Menus · rtylr" />
      <div className="w-full min-w-0 space-y-5">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <UtensilsCrossed size={17} style={{ color: 'var(--dash-text-muted)' }} />
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                Menus
              </h1>
            </div>
            <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              Storefront catalog
            </p>
          </div>

          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <div className="relative min-w-0 sm:w-80">
              <Search
                size={14}
                className="absolute left-3 top-1/2 -translate-y-1/2"
                style={{ color: 'var(--dash-text-dim)' }}
              />
              <input
                className="h-9 w-full rounded-lg pl-9 pr-3 text-sm outline-none transition-colors focus:border-[var(--dash-border-focus)]"
                style={{
                  background: 'var(--dash-input-bg)',
                  border: '1px solid var(--dash-border)',
                  color: 'var(--dash-text)',
                }}
                placeholder="Search menus"
                value={searchTerm}
                onChange={(event) => setSearchTerm(event.target.value)}
              />
            </div>
            <Button size="sm" onClick={() => navigate({ to: '/create-item' })}>
              <Plus size={14} />
              Item
            </Button>
          </div>
        </div>

        <div
          className="grid overflow-hidden rounded-lg border sm:grid-cols-3"
          style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-surface)' }}
        >
          {[
            ['Menus', String(menus.length)],
            ['Published', String(activeMenus)],
            ['Items', String(totalItems)],
          ].map(([label, value], index) => (
            <div
              key={label}
              className="px-4 py-3"
              style={{
                borderLeft: index === 0 ? undefined : '1px solid var(--dash-border)',
                borderTop: index === 0 ? undefined : '1px solid var(--dash-border)',
              }}
            >
              <p
                className="text-[11px] font-medium uppercase"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {label}
              </p>
              <p className="mt-1 text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                {value}
              </p>
            </div>
          ))}
        </div>

        <section
          className="w-full overflow-hidden rounded-lg border"
          style={{ borderColor: 'var(--dash-border)', background: 'var(--dash-surface)' }}
        >
          <div
            className="hidden px-4 py-2.5 text-[11px] font-semibold uppercase md:grid md:grid-cols-[minmax(0,1fr)_120px_120px_140px] md:items-center"
            style={{
              color: 'var(--dash-text-dim)',
              borderBottom: '1px solid var(--dash-border)',
              background: 'var(--dash-hover)',
            }}
          >
            <span>Menu</span>
            <span>Items</span>
            <span>Status</span>
            <span className="text-right">Action</span>
          </div>

          {isLoading ? (
            <div className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
              {[0, 1, 2, 3].map((row) => (
                <div
                  key={row}
                  className="grid gap-3 px-4 py-4 md:grid-cols-[minmax(0,1fr)_120px_120px_140px] md:items-center"
                  style={{ borderBottom: row === 3 ? undefined : '1px solid var(--dash-border)' }}
                >
                  <div className="space-y-2">
                    <div
                      className="h-3 w-44 animate-pulse rounded"
                      style={{ background: 'var(--dash-hover)' }}
                    />
                    <div
                      className="h-2.5 w-28 animate-pulse rounded"
                      style={{ background: 'var(--dash-hover)' }}
                    />
                  </div>
                  <div
                    className="h-3 w-14 animate-pulse rounded"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                  <div
                    className="h-6 w-20 animate-pulse rounded-full"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                  <div
                    className="ml-auto h-7 w-20 animate-pulse rounded-md"
                    style={{ background: 'var(--dash-hover)' }}
                  />
                </div>
              ))}
            </div>
          ) : filteredMenus.length === 0 ? (
            <EmptyState
              icon={UtensilsCrossed}
              title={t('menu.emptyNoMenusTitle')}
              description={
                searchTerm ? t('menu.emptyTryAnotherSearch') : t('menu.emptyNoMenuItemsYet')
              }
              action={
                searchTerm
                  ? { label: t('common.clearSearch'), onClick: () => setSearchTerm('') }
                  : {
                      label: t('menu.createItem'),
                      onClick: () => navigate({ to: '/create-item' }),
                    }
              }
            />
          ) : (
            <div>
              {filteredMenus.map((menu, index) => {
                const name = menuName(menu);
                const itemCount = menuItemCount(menu);
                const active = isMenuActive(menu);
                return (
                  <div
                    key={menu._id}
                    className="grid gap-3 px-4 py-4 transition-colors hover:bg-[var(--dash-hover)] md:grid-cols-[minmax(0,1fr)_120px_120px_140px] md:items-center"
                    style={{
                      borderBottom:
                        index === filteredMenus.length - 1
                          ? undefined
                          : '1px solid var(--dash-border)',
                    }}
                  >
                    <div className="min-w-0">
                      <p
                        className="truncate text-sm font-semibold"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {name}
                      </p>
                      <p
                        className="mt-0.5 truncate text-xs"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        Updated {formatDate(menu.updatedAt ?? menu.createdAt)}
                      </p>
                    </div>
                    <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                      {itemCount} {itemCount === 1 ? 'item' : 'items'}
                    </p>
                    <div>
                      <Badge variant={active ? 'success' : 'outline'}>
                        {active ? 'Live' : 'Draft'}
                      </Badge>
                    </div>
                    <div className="flex justify-start md:justify-end">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() =>
                          navigate({
                            to: '/menu',
                            search: { menuId: String(menu.id || menu._id || '') },
                          })
                        }
                      >
                        <ExternalLink size={13} />
                        Open
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </div>
    </>
  );
};

export default Menus;
