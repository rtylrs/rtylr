import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { HelmetContainer } from '@voxire/components';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@voxire/ui';
import {
  type LucideIcon,
  Search,
  ChevronDown,
  Check,
  Plug2,
  Package,
  Globe,
  Mail,
  Heart,
  BarChart2,
  CalendarCheck,
  Monitor,
  Code,
  Download,
  ShoppingCart,
  Paintbrush,
  Headphones,
  MessageSquare,
  FileText,
  EyeOff,
  MonitorSpeaker,
  Store as StoreIcon,
} from 'lucide-react';
import { toast } from 'sonner';
import { PageHeader } from '../../shared/components/dash/PageHeader';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import {
  Button,
  Input,
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../shared/components/ui';
import {
  ADDONS_QUERY_KEY,
  useAddons,
  enableAddon,
  disableAddon,
} from '../../shared/hooks/queries/useAddons';
import {
  useIntegrations,
  useConnectIntegration,
  useDisconnectIntegration,
} from '../../shared/hooks/queries/useIntegrations';
import {
  useBillingDetail,
  useMySubscription,
  centsToDisplay,
} from '../../shared/hooks/queries/useSubscription';
import { useUserStore } from '../../shared/stores/userStore';
import {
  buildExtensionCatalog,
  getFeatureSettings,
  EXTENSION_CATEGORIES,
  type ExtensionCategory,
  type ExtensionEntry,
} from './catalog';
import './extensions.css';

const FEATURE_ICONS: Record<string, LucideIcon> = {
  globe: Globe,
  mail: Mail,
  heart: Heart,
  'bar-chart-2': BarChart2,
  'bar-chart': BarChart2,
  'calendar-check': CalendarCheck,
  calendar: CalendarCheck,
  monitor: Monitor,
  code: Code,
  download: Download,
  'shopping-cart': ShoppingCart,
  'shopping-bag': StoreIcon,
  paintbrush: Paintbrush,
  headphones: Headphones,
  'message-square': MessageSquare,
  'file-text': FileText,
  'eye-off': EyeOff,
  'utensils-crossed': MonitorSpeaker,
};
type ExtensionAction = { entry: ExtensionEntry; operation: 'activate' | 'remove' };

export default function Extensions() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const user = useUserStore((state) => state.user);
  const canManage = !!(user?.effectiveAccess?.isOwner || user?.effectiveAccess?.isAdmin);
  const addons = useAddons();
  const integrations = useIntegrations();
  const subscription = useMySubscription();
  const billing = useBillingDetail(canManage);
  const connect = useConnectIntegration();
  const disconnect = useDisconnectIntegration();
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<ExtensionCategory>('all');
  const [addedOnly, setAddedOnly] = useState(false);
  const [action, setAction] = useState<ExtensionAction | null>(null);
  const [confirmation, setConfirmation] = useState('');
  const [actionError, setActionError] = useState(false);
  const trial = subscription.data?.status === 'trial' || billing.data?.status === 'trial';
  const subscriptionReady = !!subscription.data && !subscription.isError;
  const mutateFeature = useMutation({
    mutationFn: ({ id, remove }: { id: string; remove: boolean }) =>
      remove ? disableAddon(id) : enableAddon(id),
    onSuccess: async (_, { remove }) => {
      toast.success(t(remove ? 'extensions.removedSuccess' : 'extensions.addedSuccess'));
      setAction(null);
      setConfirmation('');
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ADDONS_QUERY_KEY }),
        queryClient.invalidateQueries({ queryKey: ['subscription', 'billing'] }),
      ]);
    },
    onError: () => setActionError(true),
  });
  const busy = mutateFeature.isPending || connect.isPending || disconnect.isPending;
  const loading = addons.isLoading || integrations.isLoading;
  const entries = buildExtensionCatalog(
    addons.isError ? [] : (addons.data ?? []),
    integrations.isError ? [] : (integrations.data ?? [])
  ).filter((entry) => !(integrations.isError && entry.kind === 'connection'));
  const categories = EXTENSION_CATEGORIES.filter(
    (value) => value === 'all' || entries.some((entry) => entry.category === value)
  );
  const filtered = entries.filter(
    (entry) =>
      (!addedOnly || entry.added) &&
      (category === 'all' || entry.category === category) &&
      `${entry.name} ${entry.description} ${t(`extensions.categories.${entry.category}`)}`
        .toLocaleLowerCase()
        .includes(search.trim().toLocaleLowerCase())
  );
  const granted = new Set(
    billing.data?.lineItems
      .filter((item) => item.type === 'addon' && item.isGranted)
      .map((item) => item.addonId)
  );
  const reset = () => {
    setSearch('');
    setCategory('all');
    setAddedOnly(false);
  };
  const openAction = (entry: ExtensionEntry, operation: ExtensionAction['operation']) => {
    setActionError(false);
    setConfirmation('');
    setAction({ entry, operation });
  };
  const closeAction = () => {
    if (!busy) {
      setAction(null);
      setActionError(false);
      setConfirmation('');
    }
  };
  const runAction = async () => {
    if (!action || !canManage || busy) return;
    setActionError(false);
    if (action.entry.kind === 'feature') {
      if (trial || !subscriptionReady) return;
      mutateFeature.mutate({ id: action.entry.addon.id, remove: action.operation === 'remove' });
    } else {
      try {
        await disconnect.mutateAsync(action.entry.integration.id);
        setAction(null);
      } catch {
        setActionError(true);
      }
    }
  };

  return (
    <div className="portal-management-page portal-extensions">
      <HelmetContainer title={`${t('common.nav.extensions')} · rtylr`} />
      <PageHeader title={t('common.nav.extensions')} description={t('extensions.description')} />
      <div className="portal-extensions-toolbar">
        <label className="portal-extensions-search">
          <Search size={16} aria-hidden />
          <input
            type="search"
            aria-label={t('extensions.search')}
            placeholder={t('extensions.search')}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </label>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" aria-label={t('extensions.category')}>
              {t(`extensions.categories.${category}`)}
              <ChevronDown size={14} aria-hidden />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="portal-extensions-category-menu" align="end">
            {categories.map((value) => (
              <DropdownMenuItem key={value} onSelect={() => setCategory(value)}>
                {t(`extensions.categories.${value}`)}
                {category === value && <Check size={14} aria-hidden />}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
        <Button
          variant={addedOnly ? 'secondary' : 'ghost'}
          aria-pressed={addedOnly}
          onClick={() => setAddedOnly(!addedOnly)}
        >
          <Check size={15} aria-hidden />
          {t('extensions.added')}
        </Button>
      </div>
      {!canManage && <p className="portal-extensions-notice">{t('extensions.readOnly')}</p>}
      {trial && <p className="portal-extensions-notice">{t('extensions.trial')}</p>}
      {canManage && subscription.isError && (
        <div className="portal-extensions-notice" role="alert">
          {t('extensions.subscriptionError')}
          <Button size="sm" variant="ghost" onClick={() => void subscription.refetch()}>
            {t('extensions.retry')}
          </Button>
        </div>
      )}
      {(addons.isError || integrations.isError) && (
        <div className="portal-extensions-notice" role="alert">
          {t(
            addons.isError && integrations.isError
              ? 'extensions.loadError'
              : addons.isError
                ? 'extensions.featuresError'
                : 'extensions.connectionsError'
          )}
          <Button
            size="sm"
            variant="ghost"
            onClick={() => {
              if (addons.isError) void addons.refetch();
              if (integrations.isError) void integrations.refetch();
            }}
          >
            {t('extensions.retry')}
          </Button>
        </div>
      )}
      {loading ? (
        <p role="status" className="portal-extensions-notice">
          {t('extensions.loading')}
        </p>
      ) : (
        <>
          <p className="portal-extensions-results" role="status">
            {t('extensions.results', { count: filtered.length })}
          </p>
          <div className="portal-extensions-grid">
            {filtered.map((entry) => {
              const Icon =
                entry.kind === 'connection' ? Plug2 : (FEATURE_ICONS[entry.addon.icon] ?? Package);
              const feature = entry.kind === 'feature' ? entry.addon : undefined;
              const route = feature && entry.added ? getFeatureSettings(feature.id) : undefined;
              const unavailable =
                entry.kind === 'connection' && entry.integration.status === 'coming_soon';
              const failedConnection =
                entry.kind === 'connection' && entry.integration.status === 'error';
              return (
                <article key={entry.key} className="portal-extensions-item">
                  <div className="portal-extensions-item-top">
                    <span className="portal-extensions-icon">
                      <Icon size={23} strokeWidth={1.5} aria-hidden />
                    </span>
                    <span className="portal-extensions-category">
                      {t(`extensions.categories.${entry.category}`)}
                    </span>
                    {entry.added && (
                      <span className="portal-extensions-status">
                        <Check size={12} aria-hidden />
                        {t('extensions.added')}
                      </span>
                    )}
                  </div>
                  <h2>{entry.name}</h2>
                  <p>
                    {entry.kind === 'connection' &&
                    entry.integration.provider.toLowerCase() === 'stripe'
                      ? t('extensions.stripeDescription')
                      : entry.description}
                  </p>
                  {failedConnection && (
                    <p className="portal-extensions-connection-error">
                      {t('extensions.connectionError')}
                    </p>
                  )}
                  <div className="portal-extensions-item-footer">
                    <span className="portal-extensions-price">
                      {feature
                        ? feature.isFree
                          ? t('extensions.free')
                          : granted.has(feature.id)
                            ? t('extensions.included')
                            : t('extensions.monthlyPrice', {
                                price: centsToDisplay(Math.round(feature.price * 100)),
                              })
                        : t('extensions.connection')}
                    </span>
                    <div className="portal-extensions-item-actions">
                      {route && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => void navigate({ to: route })}
                        >
                          {t('extensions.manage')}
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="secondary"
                        disabled={
                          !canManage ||
                          busy ||
                          unavailable ||
                          (!!feature && (trial || !subscriptionReady))
                        }
                        onClick={() => {
                          if (entry.added || failedConnection) openAction(entry, 'remove');
                          else if (entry.kind === 'feature') openAction(entry, 'activate');
                          else connect.mutate(entry.integration.provider);
                        }}
                      >
                        {unavailable
                          ? t('extensions.comingSoon')
                          : busy &&
                              ((mutateFeature.variables?.id === feature?.id &&
                                mutateFeature.isPending) ||
                                (connect.variables ===
                                  (entry.kind === 'connection'
                                    ? entry.integration.provider
                                    : undefined) &&
                                  connect.isPending))
                            ? t('extensions.working')
                            : entry.added || failedConnection
                              ? t('extensions.remove')
                              : feature
                                ? t('extensions.activate')
                                : t('extensions.connect')}
                      </Button>
                    </div>
                  </div>
                </article>
              );
            })}
          </div>
          {filtered.length === 0 && !(addons.isError && integrations.isError) && (
            <EmptyState
              icon={Search}
              title={t(addedOnly ? 'extensions.noAdded' : 'extensions.noResults')}
              description={t('extensions.emptyDescription')}
              action={{ label: t('extensions.reset'), onClick: reset }}
            />
          )}
        </>
      )}
      <Dialog
        open={!!action}
        onOpenChange={(open) => {
          if (!open) closeAction();
        }}
      >
        <DialogContent
          onEscapeKeyDown={(event) => {
            if (busy) event.preventDefault();
          }}
          onPointerDownOutside={(event) => {
            if (busy) event.preventDefault();
          }}
        >
          <DialogHeader>
            <DialogTitle>
              {t(
                action?.operation === 'remove'
                  ? 'extensions.removeTitle'
                  : 'extensions.activateTitle',
                { name: action?.entry.name }
              )}
            </DialogTitle>
            <DialogDescription>
              {t(
                action?.operation === 'remove'
                  ? action.entry.kind === 'feature'
                    ? 'extensions.removeFeatureDescription'
                    : 'extensions.disconnectDescription'
                  : action?.entry.kind === 'feature' && !action.entry.addon.isFree
                    ? 'extensions.entitlementDescription'
                    : 'extensions.activateDescription'
              )}
            </DialogDescription>
          </DialogHeader>
          <div className="portal-extensions-dialog-body">
            {action?.operation === 'activate' &&
              action.entry.kind === 'feature' &&
              !action.entry.addon.isFree && (
                <p>
                  {t('extensions.listedPrice', {
                    price: centsToDisplay(Math.round(action.entry.addon.price * 100)),
                  })}
                </p>
              )}
            {action?.operation === 'remove' && (
              <label className="portal-extensions-confirmation">
                {t('extensions.confirmName', { name: action.entry.name })}
                <Input
                  autoFocus
                  value={confirmation}
                  onChange={(event) => setConfirmation(event.target.value)}
                />
              </label>
            )}
            {actionError && (
              <p role="alert">
                {t(
                  action?.operation === 'activate'
                    ? 'extensions.activationError'
                    : 'extensions.removeError'
                )}
              </p>
            )}
          </div>
          <DialogFooter>
            <Button variant="ghost" disabled={busy} onClick={closeAction}>
              {t('extensions.cancel')}
            </Button>
            <Button
              disabled={
                busy ||
                (action?.operation === 'remove' && confirmation.trim() !== action.entry.name)
              }
              onClick={() => void runAction()}
            >
              {t(
                busy
                  ? 'extensions.working'
                  : action?.operation === 'remove'
                    ? 'extensions.remove'
                    : 'extensions.activate'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
