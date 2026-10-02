import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Building2, Check, ChevronsUpDown, Loader2, Search, Settings2 } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import { useBusinessStore } from '../../shared/stores/businessStore';
import { useConfigStore } from '../../shared/stores/configStore';
import { useCaretStore, useChatsStore } from '@voxire/caret';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@voxire/ui';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useListMyBusinesses,
  useSwitchBusiness,
  type OwnedBusiness,
} from '../../shared/hooks/queries/useBusinesses';
import './business-switcher.css';

interface BusinessLogoProps {
  logoUrl?: string;
  size: number;
}

const BusinessLogo: React.FC<BusinessLogoProps> = ({ logoUrl, size }): React.JSX.Element => {
  if (logoUrl) {
    return (
      <img
        src={logoUrl}
        alt=""
        width={size}
        height={size}
        className="portal-business-logo object-cover shrink-0"
        style={{ height: size, width: size }}
      />
    );
  }

  return (
    <span
      className="portal-business-logo flex items-center justify-center shrink-0"
      style={{ height: size, width: size }}
      aria-hidden="true"
    >
      <Building2 size={Math.round(size * 0.6)} />
    </span>
  );
};

const BusinessSwitcher = ({
  onNavigate,
}: {
  onNavigate?: () => void;
}): React.JSX.Element | null => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user } = useUserStore();
  const currentBusinessId = Number(user?.businessId ?? 0);

  const { data, isLoading, isError, refetch } = useListMyBusinesses();
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
        useBusinessStore.getState().reset();
        useConfigStore.getState().reset();
        const caret = useCaretStore.getState();
        for (const tab of caret.tabs.slice(1)) caret.closeTab(tab.id);
        caret.reset();
        useChatsStore.getState().clear();
        window.location.reload();
      },
      onError: () => {
        toast.error(t('businesses.switchError'));
      },
    });
  };

  return (
    <div className="portal-business-picker">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            type="button"
            className="portal-business-trigger"
            disabled={switchMutation.isPending}
          >
            <BusinessLogo logoUrl={currentBusiness?.logoUrl} size={28} />
            <span className="flex-1 truncate text-start">
              {currentBusiness?.name ?? user?.businessName ?? t('businesses.myBusinesses')}
            </span>
            <ChevronsUpDown
              size={13}
              className="shrink-0"
              style={{ color: 'var(--dash-text-dim)' }}
            />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          data-portal-profile-menu
          align="start"
          side="bottom"
          className="portal-business-menu"
        >
          <div className="portal-business-menu-label">{t('businesses.myBusinesses')}</div>
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
                  onKeyDown={(event) => {
                    if (event.key !== 'Escape' && event.key !== 'Tab') event.stopPropagation();
                  }}
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
            {isLoading && <p className="portal-business-menu-label">{t('common.loading')}</p>}
            {isError && (
              <div role="alert" className="portal-business-load-error">
                <p>{t('businesses.loadError')}</p>
                <button type="button" onClick={() => refetch()}>
                  {t('common.tryAgain')}
                </button>
              </div>
            )}
            {filteredBusinesses.map((b) => (
              <DropdownMenuItem
                className="portal-business-option"
                key={b.id}
                onClick={() => handleSwitch(b)}
                disabled={switchMutation.isPending}
              >
                <BusinessLogo logoUrl={b.logoUrl} size={16} />
                <span className="flex-1 truncate">{b.name}</span>
                {b.id === currentBusinessId && <Check size={13} />}
                {switchMutation.isPending && b.id !== currentBusinessId && (
                  <Loader2
                    size={11}
                    className="animate-spin"
                    style={{ color: 'var(--dash-text-dim)' }}
                  />
                )}
              </DropdownMenuItem>
            ))}
            {!isLoading && !isError && filteredBusinesses.length === 0 && (
              <EmptyState
                icon={Search}
                title={t('businesses.noResults')}
                action={{ label: t('common.clearSearch'), onClick: () => setSearch('') }}
                size="compact"
              />
            )}
          </div>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => {
              onNavigate?.();
              navigate({ to: '/settings/businesses' });
            }}
          >
            <Settings2 size={13} />
            {t('businesses.manage')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
};

export default BusinessSwitcher;
