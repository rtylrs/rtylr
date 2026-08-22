import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Building2, Check, ChevronsUpDown, Loader2, Search, Settings2 } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';
import { EmptyState } from '@voxire/ui';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../../shared/components/ui';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useListMyBusinesses,
  useSwitchBusiness,
  type OwnedBusiness,
} from '../../shared/hooks/queries/useBusinesses';

interface BusinessLogoProps {
  logoUrl?: string;
  name: string;
  size: number;
}

const BusinessLogo: React.FC<BusinessLogoProps> = ({ logoUrl, name, size }): React.JSX.Element => {
  if (logoUrl) {
    return (
      <img
        src={logoUrl}
        alt=""
        width={size}
        height={size}
        className="rounded object-cover shrink-0"
        style={{ height: size, width: size }}
      />
    );
  }

  return (
    <span
      className="rounded flex items-center justify-center shrink-0"
      style={{ height: size, width: size, background: 'rgba(69,131,196,0.12)' }}
      aria-label={name}
    >
      <Building2 size={Math.round(size * 0.55)} style={{ color: '#4583C4' }} />
    </span>
  );
};

const BusinessSwitcher: React.FC = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user } = useUserStore();
  const currentBusinessId = Number(user?.businessId ?? 0);

  const { data } = useListMyBusinesses();
  const businesses: OwnedBusiness[] = data?.businesses ?? [];

  const switchMutation = useSwitchBusiness();

  const [search, setSearch] = useState('');

  const currentBusiness = businesses.find((b) => b.id === currentBusinessId);

  // Only surface the search field once the list is long enough to warrant it.
  const showSearch = businesses.length > 5;
  const filteredBusinesses = search.trim()
    ? businesses.filter((b) => b.name.toLowerCase().includes(search.trim().toLowerCase()))
    : businesses;

  const handleSwitch = (business: OwnedBusiness): void => {
    if (business.id === currentBusinessId) return;
    switchMutation.mutate(business.id, {
      onSuccess: ({ token }) => {
        if (!token) {
          toast.error(t('businesses.switchError'));
          return;
        }
        removeCookie('token', { domain: getDefaultCookieDomain() });
        removeCookie('token');
        setCookie('token', token, {
          days: 30,
          sameSite: 'lax',
          domain: getDefaultCookieDomain(),
        });
        window.location.reload();
      },
      onError: () => {
        toast.error(t('businesses.switchError'));
      },
    });
  };

  return (
    <div
      className="dash-business-switcher px-2 pt-2 pb-2"
      style={{ borderBottom: '1px solid var(--dash-border)' }}
    >
      <DropdownMenu>
        <DropdownMenuTrigger>
          <button
            type="button"
            className="flex items-center gap-2.5 w-full px-2 py-2 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-[var(--dash-hover)]"
            style={{ color: 'var(--dash-text-muted)' }}
          >
            <BusinessLogo
              logoUrl={currentBusiness?.logoUrl}
              name={currentBusiness?.name ?? user?.businessName ?? ''}
              size={32}
            />
            <span
              className="dash-sidebar-label flex-1 truncate text-left text-sm font-semibold"
              style={{ color: 'var(--dash-text)' }}
            >
              {currentBusiness?.name ?? user?.businessName ?? t('businesses.myBusinesses')}
            </span>
            <ChevronsUpDown
              size={13}
              className="dash-sidebar-chevron shrink-0"
              style={{ color: 'var(--dash-text-dim)' }}
            />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" side="bottom" className="w-52">
          <div
            className="px-2 py-1.5 text-[10px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            {t('businesses.myBusinesses')}
          </div>
          {showSearch && (
            <div className="px-2 pb-1.5">
              <div className="relative">
                <Search
                  size={12}
                  className="absolute left-2.5 top-1/2 -translate-y-1/2"
                  style={{ color: 'var(--dash-text-dim)' }}
                />
                <input
                  autoFocus
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder={t('businesses.searchPlaceholder')}
                  aria-label={t('businesses.searchPlaceholder')}
                  className="w-full rounded-lg py-1.5 pl-7 pr-2 text-xs outline-none"
                  style={{
                    fontSize: '16px',
                    background: 'var(--dash-hover)',
                    color: 'var(--dash-text)',
                    border: '1px solid var(--dash-border)',
                  }}
                />
              </div>
            </div>
          )}
          <div className="max-h-60 overflow-y-auto">
            {filteredBusinesses.map((b) => (
              <DropdownMenuItem key={b.id} onClick={() => handleSwitch(b)}>
                <BusinessLogo logoUrl={b.logoUrl} name={b.name} size={16} />
                <span className="flex-1 truncate">{b.name}</span>
                {b.id === currentBusinessId && <Check size={13} style={{ color: '#4583C4' }} />}
                {switchMutation.isPending && b.id !== currentBusinessId && (
                  <Loader2
                    size={11}
                    className="animate-spin"
                    style={{ color: 'var(--dash-text-dim)' }}
                  />
                )}
              </DropdownMenuItem>
            ))}
            {filteredBusinesses.length === 0 && (
              <EmptyState
                icon={Search}
                title={t('businesses.noResults')}
                action={{ label: t('common.clearSearch'), onClick: () => setSearch('') }}
                size="compact"
              />
            )}
          </div>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => navigate({ to: '/settings/businesses' })}>
            <Settings2 size={13} />
            {t('businesses.manage')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export default BusinessSwitcher;
