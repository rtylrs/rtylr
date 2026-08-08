import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Building2, Check, Loader2, Pencil, Plus, Trash2 } from 'lucide-react';
import { setCookie, removeCookie, getDefaultCookieDomain } from '@voxire/auth';
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
      <div className="max-w-5xl space-y-7">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex items-start gap-3">
            <div
              className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl"
              style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
            >
              <Building2 size={21} />
            </div>
            <div>
              <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
                {t('businesses.manageTitle')}
              </h1>
              <p className="mt-1 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
                {t('businesses.manageDescription')}
              </p>
            </div>
          </div>
          <Button
            className="shrink-0 cursor-pointer self-start sm:self-auto"
            onClick={() => setCreateOpen(true)}
            disabled={isTrialMode || !parentBusinessId}
            title={isTrialMode ? t('businesses.trialSingleBusinessTooltip') : undefined}
          >
            <Plus size={15} />
            {t('businesses.addBusiness')}
          </Button>
        </div>

        <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_260px]">
          <Card className="overflow-hidden">
            <CardHeader
              className="border-b px-5 py-5"
              style={{ borderColor: 'var(--dash-border)' }}
            >
              <CardTitle>{t('businesses.manageCardTitle')}</CardTitle>
              <CardDescription>{t('businesses.manageCardDescription')}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2 p-3 sm:p-4">
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
                  const isDeleteDisabled =
                    !isOwned || !canDeleteBusiness || deleteBusiness.isPending;

                  return (
                    <div
                      key={business.id}
                      className="group flex items-center gap-3 rounded-xl border px-3 py-3.5 transition-colors hover:bg-[var(--dash-hover)] sm:px-4"
                      style={{
                        background: isCurrent ? 'rgba(69,131,196,0.05)' : undefined,
                        borderColor: isCurrent ? 'rgba(69,131,196,0.35)' : 'var(--dash-border)',
                      }}
                    >
                      <div
                        className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl"
                        style={{
                          background: isCurrent ? 'rgba(69,131,196,0.14)' : 'var(--dash-hover)',
                          color: isCurrent ? '#4583C4' : 'var(--dash-text-muted)',
                        }}
                      >
                        <Building2 size={18} />
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <p
                            className="truncate text-sm font-semibold"
                            style={{ color: 'var(--dash-text)' }}
                          >
                            {business.name}
                          </p>
                          {isCurrent && (
                            <span
                              className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium"
                              style={{ background: 'rgba(69,131,196,0.12)', color: '#4583C4' }}
                            >
                              <Check size={10} />
                              {t('businesses.current')}
                            </span>
                          )}
                          {!isOwned && (
                            <span
                              className="rounded-full px-2 py-0.5 text-[10px] font-medium"
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
                      {isOwned && (
                        <div className="ml-auto flex shrink-0 items-center gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 cursor-pointer px-2 text-xs"
                            onClick={() => openRename(business)}
                            aria-label={`${t('businesses.rename')} ${business.name}`}
                            style={{ color: 'var(--dash-text-muted)' }}
                          >
                            <Pencil size={13} />
                            <span className="hidden md:inline">{t('businesses.rename')}</span>
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 cursor-pointer px-2 text-xs"
                            onClick={() => setDeleteTarget(business)}
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
                            <span className="hidden md:inline">{t('businesses.delete')}</span>
                          </Button>
                        </div>
                      )}
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
                    'linear-gradient(145deg, rgba(69,131,196,0.16), rgba(69,131,196,0.04))',
                  borderBottom: '1px solid var(--dash-border)',
                }}
              >
                <p
                  className="text-[10px] font-semibold uppercase tracking-widest"
                  style={{ color: '#4583C4' }}
                >
                  {t('businesses.current')}
                </p>
                <p
                  className="mt-2 truncate text-base font-semibold"
                  style={{ color: 'var(--dash-text)' }}
                >
                  {currentBusiness?.name ?? user?.businessName ?? t('businesses.myBusinesses')}
                </p>
                <div
                  className="mt-3 inline-flex items-center gap-1.5 text-xs"
                  style={{ color: 'var(--dash-text-muted)' }}
                >
                  <Check size={13} style={{ color: '#4583C4' }} />
                  {currentBusiness?.isOwned === false
                    ? t('businesses.teamAccess')
                    : t('businesses.ownerAccess')}
                </div>
              </div>
              <CardContent className="p-5">
                <p className="text-xs leading-5" style={{ color: 'var(--dash-text-dim)' }}>
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
