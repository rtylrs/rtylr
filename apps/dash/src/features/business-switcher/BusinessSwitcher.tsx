import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Building2, Check, ChevronRight, Plus, Loader2, Search, Settings2 } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Tooltip,
} from '../../shared/components/ui';
import { useUserStore } from '../../shared/stores/userStore';
import {
  useListMyBusinesses,
  useCreateSubBusiness,
  useSwitchBusiness,
  type OwnedBusiness,
} from '../../shared/hooks/queries/useBusinesses';
import { useBillingDetail, useMySubscription } from '../../shared/hooks/queries/useSubscription';

const BusinessSwitcher: React.FC = (): React.JSX.Element | null => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user } = useUserStore();
  const currentBusinessId = user?.businessId ?? 0;

  const { data } = useListMyBusinesses();
  const businesses: OwnedBusiness[] = data?.businesses ?? [];
  const { data: sub } = useMySubscription();
  const { data: billing } = useBillingDetail();

  const switchMutation = useSwitchBusiness();
  const createMutation = useCreateSubBusiness();

  const [modalOpen, setModalOpen] = useState(false);
  const [newName, setNewName] = useState('');
  const [search, setSearch] = useState('');

  const currentBusiness = businesses.find((b) => b.id === currentBusinessId);
  const isTrialMode = sub?.status === 'trial' || billing?.status === 'trial';

  // Only surface the search field once the list is long enough to warrant it.
  const showSearch = businesses.length > 5;
  const filteredBusinesses = search.trim()
    ? businesses.filter((b) => b.name.toLowerCase().includes(search.trim().toLowerCase()))
    : businesses;

  const handleSwitch = (business: OwnedBusiness): void => {
    if (business.id === currentBusinessId) return;
    switchMutation.mutate(business.id, {
      onSuccess: ({ token }) => {
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

  const handleCreate = (e: React.FormEvent): void => {
    e.preventDefault();
    if (!newName.trim()) return;
    if (!currentBusinessId) {
      toast.error(t('businesses.parentRequired'));
      return;
    }
    createMutation.mutate(
      { name: newName.trim(), parentBusinessId: currentBusinessId },
      {
        onSuccess: () => {
          toast.success(t('businesses.createSuccess'));
          setNewName('');
          setModalOpen(false);
        },
        onError: () => {
          toast.error(t('businesses.createError'));
        },
      }
    );
  };

  const handleModalClose = (open: boolean): void => {
    if (!open) {
      setNewName('');
    }
    setModalOpen(open);
  };

  return (
    <>
      <div
        className="dash-business-switcher px-2 py-1.5"
        style={{ borderBottom: '1px solid var(--dash-border)' }}
      >
        <DropdownMenu>
          <DropdownMenuTrigger>
            <button
              className="flex items-center gap-2 w-full px-3 py-2 rounded-lg cursor-pointer text-sm transition-colors duration-150 hover:bg-[var(--dash-hover)]"
              style={{ color: 'var(--dash-text-muted)' }}
            >
              <span
                className="h-5 w-5 rounded flex items-center justify-center shrink-0"
                style={{ background: 'rgba(69,131,196,0.12)' }}
              >
                <Building2 size={11} style={{ color: '#4583C4' }} />
              </span>
              <span
                className="dash-sidebar-label flex-1 truncate text-left text-xs"
                style={{ color: 'var(--dash-text)' }}
              >
                {currentBusiness?.name ?? user?.businessName ?? t('businesses.myBusinesses')}
              </span>
              <ChevronRight
                size={11}
                className="dash-sidebar-chevron"
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
                <div className="px-3 py-2 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('businesses.noResults')}
                </div>
              )}
            </div>
            <DropdownMenuSeparator />
            <Tooltip
              side="right"
              className="block"
              disabled={!isTrialMode}
              content={t('businesses.trialSingleBusinessTooltip')}
            >
              <DropdownMenuItem
                aria-disabled={isTrialMode}
                className={isTrialMode ? 'cursor-not-allowed opacity-60 hover:bg-transparent' : ''}
                onClick={(e) => {
                  if (isTrialMode) {
                    e.preventDefault();
                    e.stopPropagation();
                    return;
                  }
                  setModalOpen(true);
                }}
              >
                <Plus size={13} />
                {t('businesses.addBusiness')}
              </DropdownMenuItem>
            </Tooltip>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => navigate({ to: '/settings/businesses' })}>
              <Settings2 size={13} />
              {t('businesses.manage')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <Dialog open={modalOpen} onOpenChange={handleModalClose}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('businesses.createSubBusiness')}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate}>
            <div className="px-6 pb-2">
              <input
                autoFocus
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder={t('businesses.businessName')}
                className="w-full text-sm px-3 py-2 rounded-lg outline-none"
                style={{
                  background: 'var(--dash-hover)',
                  color: 'var(--dash-text)',
                  border: '1px solid var(--dash-border)',
                }}
                disabled={createMutation.isPending}
              />
            </div>
            <DialogFooter>
              <button
                type="button"
                onClick={() => handleModalClose(false)}
                className="text-sm px-4 py-2 rounded-lg transition-colors hover:bg-[var(--dash-hover)]"
                style={{ color: 'var(--dash-text-muted)' }}
              >
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                disabled={!newName.trim() || createMutation.isPending}
                className="flex items-center gap-2 text-sm px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
                style={{ background: 'rgba(69,131,196,0.15)', color: '#4583C4' }}
              >
                {createMutation.isPending && <Loader2 size={13} className="animate-spin" />}
                {createMutation.isPending
                  ? t('businesses.creating')
                  : t('businesses.createSubBusiness')}
              </button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default BusinessSwitcher;
