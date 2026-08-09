import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { ArrowRightLeft, Building2, Check, Loader2, Pencil, Plus, Trash2 } from 'lucide-react';
import { setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';
import { cn } from '../../../../../shared/lib/utils';
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from '../../../../../shared/components/ui';
import { ConfirmDialog } from '../../../../../shared/components/dash';
import {
  useCreateSubBusiness,
  useDeleteBusiness,
  useListMyBusinesses,
  useRenameBusiness,
  useSwitchBusiness,
  type OwnedBusiness,
} from '../../../../../shared/hooks/queries/useBusinesses';
import { useUserStore } from '../../../../../shared/stores/userStore';
import {
  useBillingDetail,
  useMySubscription,
} from '../../../../../shared/hooks/queries/useSubscription';

function errorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object' && 'message' in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === 'string' && message.trim()) return message;
  }
  return fallback;
}

const Businesses: React.FC = () => {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const { data, isLoading } = useListMyBusinesses();
  const { data: subscription } = useMySubscription();
  const { data: billing } = useBillingDetail();
  const createBusiness = useCreateSubBusiness();
  const renameBusiness = useRenameBusiness();
  const deleteBusiness = useDeleteBusiness();
  const switchBusiness = useSwitchBusiness();

  const businesses = data?.businesses ?? [];
  const currentBusinessId = Number(user?.businessId ?? 0);
  const currentBusiness = businesses.find((business) => business.id === currentBusinessId);
  const ownedBusinesses = businesses.filter((business) => business.isOwned !== false);
  const activeOwnedBusinesses = ownedBusinesses.filter((business) => business.isActive !== false);
  const isTrialMode = subscription?.status === 'trial' || billing?.status === 'trial';
  const canDeleteBusiness = activeOwnedBusinesses.length > 1;
  const parentBusinessId = activeOwnedBusinesses.some(
    (business) => business.id === currentBusinessId
  )
    ? currentBusinessId
    : activeOwnedBusinesses[0]?.id || 0;

  const [createOpen, setCreateOpen] = useState(false);
  const [newName, setNewName] = useState('');
  const [renameTarget, setRenameTarget] = useState<OwnedBusiness | null>(null);
  const [renameName, setRenameName] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<OwnedBusiness | null>(null);
  const [switchingId, setSwitchingId] = useState<number | null>(null);

  const handleSwitch = (business: OwnedBusiness, event?: React.MouseEvent): void => {
    event?.stopPropagation();
    if (business.id === currentBusinessId || switchBusiness.isPending) return;

    setSwitchingId(business.id);
    switchBusiness.mutate(business.id, {
      onSuccess: ({ token }) => {
        removeCookie('token');
        setCookie('token', token, {
          days: 30,
          sameSite: 'lax',
          domain: getDefaultCookieDomain(),
        });
        toast.success(t('businesses.switchSuccess', { name: business.name }));
        window.location.reload();
      },
      onError: (error) => {
        setSwitchingId(null);
        toast.error(errorMessage(error, t('businesses.switchError')));
      },
    });
  };

  const closeCreate = (open: boolean): void => {
    setCreateOpen(open);
    if (!open) setNewName('');
  };

  const handleCreate = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    const name = newName.trim();
    if (!name || !parentBusinessId) return;

    createBusiness.mutate(
      { name, parentBusinessId },
      {
        onSuccess: () => {
          toast.success(t('businesses.createSuccess'));
          closeCreate(false);
        },
        onError: (error) => toast.error(errorMessage(error, t('businesses.createError'))),
      }
    );
  };

  const openRename = (business: OwnedBusiness): void => {
    setRenameTarget(business);
    setRenameName(business.name);
  };

  const closeRename = (open: boolean): void => {
    if (!open) {
      setRenameTarget(null);
      setRenameName('');
    }
  };

  const handleRename = (event: React.FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    if (!renameTarget) return;
    const name = renameName.trim();
    if (!name) return;
    if (name === renameTarget.name.trim()) {
      closeRename(false);
      return;
    }

    renameBusiness.mutate(
      { businessId: renameTarget.id, name },
      {
        onSuccess: () => {
          toast.success(t('businesses.renameSuccess'));
          closeRename(false);
        },
        onError: (error) => toast.error(errorMessage(error, t('businesses.renameError'))),
      }
    );
  };

  const handleDelete = async (): Promise<void> => {
    if (!deleteTarget) return;
    if (!canDeleteBusiness) {
      toast.error(t('businesses.onlyBusinessError'));
      return;
    }

    let switched = false;
    try {
      if (deleteTarget.id === currentBusinessId) {
        const replacement = activeOwnedBusinesses.find(
          (business) => business.id !== deleteTarget.id
        );
        if (!replacement) {
          toast.error(t('businesses.activeDeleteError'));
          return;
        }

        const { token } = await switchBusiness.mutateAsync(replacement.id);
        removeCookie('token');
        setCookie('token', token, {
          days: 30,
          sameSite: 'lax',
          domain: getDefaultCookieDomain(),
        });
        switched = true;
      }

      await deleteBusiness.mutateAsync(deleteTarget.id);
      toast.success(t('businesses.deleteSuccess'));
      setDeleteTarget(null);

      // The active business is represented by the JWT, so refresh after the
      // safe switch above to hydrate the new business everywhere.
      if (switched) window.location.reload();
    } catch (error) {
      toast.error(errorMessage(error, t('businesses.deleteError')));
      // Switching already changed the session. Make sure the UI does not
      // continue showing the old business if the delete request failed.
      if (switched) window.location.reload();
      throw error;
    }
  };

  return (
    <>
      <HelmetContainer title={`${t('businesses.manageTitle')} · rtylr`} />
      <div className="w-full max-w-6xl space-y-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3.5">
            <div
              className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl"
              style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
            >
              <Building2 size={22} />
            </div>
            <div>
              <h1 className="text-xl font-bold tracking-tight" style={{ color: 'var(--dash-text)' }}>
                {t('businesses.manageTitle')}
              </h1>
              <p className="mt-0.5 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                {t('businesses.manageDescription')}
              </p>
            </div>
          </div>
          <Button
            className="shrink-0 cursor-pointer gap-2 self-start sm:self-auto"
            onClick={() => setCreateOpen(true)}
            disabled={isTrialMode || !parentBusinessId}
            title={isTrialMode ? t('businesses.trialSingleBusinessTooltip') : undefined}
          >
            <Plus size={16} />
            {t('businesses.addBusiness')}
          </Button>
        </div>

        <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
          <Card className="overflow-hidden">
            <CardHeader
              className="border-b px-6 py-5"
              style={{ borderColor: 'var(--dash-border)' }}
            >
              <CardTitle className="text-base font-semibold">{t('businesses.manageCardTitle')}</CardTitle>
              <CardDescription>{t('businesses.manageCardDescription')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 p-4 sm:p-6">
              {isLoading ? (
                <div className="flex items-center justify-center py-12">
                  <Loader2
                    size={20}
                    className="animate-spin"
                    style={{ color: 'var(--dash-text-dim)' }}
                  />
                </div>
              ) : businesses.length === 0 ? (
                <div
                  className="rounded-xl border border-dashed px-4 py-10 text-center"
                  style={{ borderColor: 'var(--dash-border)' }}
                >
                  <Building2
                    size={24}
                    className="mx-auto mb-2"
                    style={{ color: 'var(--dash-text-dim)' }}
                  />
                  <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                    {t('businesses.noBusinesses')}
                  </p>
                </div>
              ) : (
                businesses.map((business) => {
                  const isCurrent = business.id === currentBusinessId;
                  const isOwned = business.isOwned !== false;
                  const isSwitching = switchingId === business.id;
                  const isDeleteDisabled =
                    !isOwned || !canDeleteBusiness || deleteBusiness.isPending || switchBusiness.isPending;

                  return (
                    <div
                      key={business.id}
                      onClick={(e) => {
                        if (!isCurrent) handleSwitch(business, e);
                      }}
                      className={cn(
                        'group flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-xl border p-4 transition-all duration-150',
                        !isCurrent &&
                          'cursor-pointer hover:border-[var(--dash-border-hover)] hover:bg-[var(--dash-hover)]'
                      )}
                      style={{
                        background: isCurrent ? 'rgba(69,131,196,0.05)' : undefined,
                        borderColor: isCurrent ? 'rgba(69,131,196,0.35)' : 'var(--dash-border)',
                      }}
                    >
                      <div className="flex items-start sm:items-center gap-3.5 min-w-0 flex-1">
                        <div
                          className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl transition-transform group-hover:scale-105"
                          style={{
                            background: isCurrent ? 'rgba(69,131,196,0.14)' : 'var(--dash-hover)',
                            color: isCurrent ? '#4583C4' : 'var(--dash-text-muted)',
                          }}
                        >
                          <Building2 size={20} />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="flex flex-wrap items-center gap-2">
                            <p
                              className="truncate text-base font-semibold"
                              style={{ color: 'var(--dash-text)' }}
                            >
                              {business.name}
                            </p>
                            {isCurrent && (
                              <span
                                className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold"
                                style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
                              >
                                <Check size={12} />
                                {t('businesses.current')}
                              </span>
                            )}
                            {!isOwned && (
                              <span
                                className="rounded-full px-2.5 py-0.5 text-xs font-medium"
                                style={{
                                  background: 'var(--dash-hover)',
                                  color: 'var(--dash-text-dim)',
                                }}
                              >
                                {t('businesses.teamAccess')}
                              </span>
                            )}
                          </div>
                          <p className="mt-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
                            {isOwned
                              ? t('businesses.ownerAccess')
                              : t('businesses.teamAccessDescription')}
                          </p>
                        </div>
                      </div>

                      <div className="flex shrink-0 items-center gap-2 self-end sm:self-auto">
                        {!isCurrent && (
                          <Button
                            variant="outline"
                            size="sm"
                            className="h-8.5 gap-1.5 cursor-pointer px-3 text-xs font-medium border-blue-500/20 hover:border-blue-500/40 hover:bg-blue-500/10 hover:text-[#4583C4]"
                            disabled={switchBusiness.isPending}
                            onClick={(e) => handleSwitch(business, e)}
                          >
                            {isSwitching ? (
                              <Loader2 size={13} className="animate-spin" />
                            ) : (
                              <ArrowRightLeft size={13} />
                            )}
                            <span>{isSwitching ? t('businesses.switching') : t('businesses.switch')}</span>
                          </Button>
                        )}

                        {isOwned && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8.5 cursor-pointer px-2.5 text-xs"
                            onClick={(e) => {
                              e.stopPropagation();
                              openRename(business);
                            }}
                            aria-label={`${t('businesses.rename')} ${business.name}`}
                            style={{ color: 'var(--dash-text-muted)' }}
                          >
                            <Pencil size={13} />
                            <span>{t('businesses.rename')}</span>
                          </Button>
                        )}

                        {isOwned && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8.5 cursor-pointer px-2.5 text-xs hover:bg-red-500/10"
                            onClick={(e) => {
                              e.stopPropagation();
                              setDeleteTarget(business);
                            }}
                            disabled={isDeleteDisabled}
                            aria-label={`${t('businesses.delete')} ${business.name}`}
                            title={
                              !canDeleteBusiness
                                ? t('businesses.onlyBusinessError')
                                : t('businesses.delete')
                            }
                            style={{ color: '#ef4444' }}
                          >
                            <Trash2 size={13} />
                            <span>{t('businesses.delete')}</span>
                          </Button>
                        )}
                      </div>
                    </div>
                  );
                })
              )}
            </CardContent>
          </Card>

          <div className="space-y-5">
            <Card className="overflow-hidden">
              <div
                className="p-5"
                style={{
                  background:
                    'linear-gradient(145deg, rgba(69,131,196,0.15), rgba(69,131,196,0.03))',
                  borderBottom: '1px solid var(--dash-border)',
                }}
              >
                <div className="flex items-center justify-between">
                  <span
                    className="text-[10px] font-bold uppercase tracking-widest"
                    style={{ color: '#4583C4' }}
                  >
                    {t('businesses.current')}
                  </span>
                  <span className="flex h-2 w-2 rounded-full bg-[#4583C4]" />
                </div>
                <p
                  className="mt-2.5 truncate text-lg font-bold"
                  style={{ color: 'var(--dash-text)' }}
                >
                  {currentBusiness?.name ?? user?.businessName ?? t('businesses.myBusinesses')}
                </p>
                <div
                  className="mt-2 inline-flex items-center gap-1.5 text-xs font-medium"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  <Check size={13} style={{ color: '#4583C4' }} />
                  {currentBusiness?.isOwned === false
                    ? t('businesses.teamAccess')
                    : t('businesses.ownerAccess')}
                </div>
              </div>
              <CardContent className="p-5">
                <p className="text-xs leading-relaxed" style={{ color: 'var(--dash-text-dim)' }}>
                  {t('businesses.manageNote')}
                </p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <Dialog open={createOpen} onOpenChange={closeCreate}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('businesses.createSubBusiness')}</DialogTitle>
            <DialogDescription>{t('businesses.createDescription')}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreate}>
            <div className="space-y-1.5 px-6 pb-2">
              <Label htmlFor="new-business-name">{t('businesses.businessName')}</Label>
              <Input
                id="new-business-name"
                autoFocus
                value={newName}
                onChange={(event) => setNewName(event.target.value)}
                disabled={createBusiness.isPending}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="ghost" onClick={() => closeCreate(false)}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={!newName.trim() || createBusiness.isPending}>
                {createBusiness.isPending && <Loader2 size={14} className="animate-spin" />}
                {createBusiness.isPending
                  ? t('businesses.creating')
                  : t('businesses.createSubBusiness')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={renameTarget !== null} onOpenChange={closeRename}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('businesses.renameTitle')}</DialogTitle>
            <DialogDescription>{t('businesses.renameDescription')}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleRename}>
            <div className="space-y-1.5 px-6 pb-2">
              <Label htmlFor="rename-business-name">{t('businesses.businessName')}</Label>
              <Input
                id="rename-business-name"
                autoFocus
                value={renameName}
                onChange={(event) => setRenameName(event.target.value)}
                disabled={renameBusiness.isPending}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="ghost" onClick={() => closeRename(false)}>
                {t('common.cancel')}
              </Button>
              <Button type="submit" disabled={!renameName.trim() || renameBusiness.isPending}>
                {renameBusiness.isPending && <Loader2 size={14} className="animate-spin" />}
                {renameBusiness.isPending ? t('businesses.renaming') : t('common.save')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open && !deleteBusiness.isPending && !switchBusiness.isPending) {
            setDeleteTarget(null);
          }
        }}
        title={t('businesses.deleteTitle')}
        description={
          deleteTarget ? (
            <>
              {t('businesses.deleteDescription')}{' '}
              <strong style={{ color: 'var(--dash-text)' }}>{deleteTarget.name}</strong>
            </>
          ) : undefined
        }
        confirmLabel={
          switchBusiness.isPending || deleteBusiness.isPending
            ? t('businesses.deleting')
            : t('businesses.delete')
        }
        cancelLabel={t('common.cancel')}
        destructive
        isLoading={switchBusiness.isPending || deleteBusiness.isPending}
        onConfirm={handleDelete}
      />
    </>
  );
};

export default Businesses;
