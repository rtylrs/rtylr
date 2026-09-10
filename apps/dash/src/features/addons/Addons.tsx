import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import {
  MonitorSpeaker,
  Heart,
  BarChart2,
  Paintbrush,
  Globe,
  MessageSquare,
  Search,
  Loader2,
  AlertTriangle,
  Store,
  CalendarCheck,
  Headphones,
  FileText,
  ShoppingCart,
  Monitor,
  Code,
  EyeOff,
  Download,
  Mail,
} from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { motion, AnimatePresence } from 'motion/react';
import { toast } from 'sonner';
import { EmptyState } from '@voxire/ui';
import { Input } from '../../shared/components/ui';
import {
  ADDONS_QUERY_KEY,
  disableAddon,
  enableAddon,
  type AddonData,
  useAddons,
} from '../../shared/hooks/queries/useAddons';
import {
  centsToDisplay,
  useBillingDetail,
  useMySubscription,
} from '../../shared/hooks/queries/useSubscription';

const ICON_MAP: Record<
  string,
  React.ComponentType<{ size?: number; style?: React.CSSProperties }>
> = {
  'utensils-crossed': MonitorSpeaker,
  heart: Heart,
  'bar-chart-2': BarChart2,
  'bar-chart': BarChart2,
  paintbrush: Paintbrush,
  globe: Globe,
  'message-square': MessageSquare,
  'shopping-bag': Store,
  'shopping-cart': ShoppingCart,
  'calendar-check': CalendarCheck,
  calendar: CalendarCheck,
  headphones: Headphones,
  'file-text': FileText,
  monitor: Monitor,
  code: Code,
  'eye-off': EyeOff,
  download: Download,
  mail: Mail,
};

const COLOR_MAP: Record<string, { bg: string; color: string }> = {
  menu: { bg: 'var(--accent-muted)', color: 'var(--accent-on-muted)' },
  loyalty: { bg: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  ai_chat: { bg: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
  analytics: { bg: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
  custom_branding: { bg: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
  multi_language: { bg: 'rgba(16,185,129,0.12)', color: '#10b981' },
  store: { bg: 'rgba(20,184,166,0.12)', color: '#14b8a6' },
  book: { bg: 'rgba(14,165,233,0.12)', color: '#0ea5e9' },
  support: { bg: 'var(--accent-muted)', color: 'var(--accent-on-muted)' },
  insights: { bg: 'rgba(99,102,241,0.12)', color: '#6366f1' },
  custom_domain: { bg: 'rgba(113,113,122,0.12)', color: '#71717a' },
  email_marketing: { bg: 'rgba(236,72,153,0.12)', color: '#ec4899' },
  sms: { bg: 'rgba(234,179,8,0.12)', color: '#eab308' },
  reservations: { bg: 'rgba(14,165,233,0.12)', color: '#0ea5e9' },
  online_ordering: { bg: 'rgba(249,115,22,0.12)', color: '#f97316' },
  kds: { bg: 'rgba(168,85,247,0.12)', color: '#a855f7' },
  api_access: { bg: 'rgba(34,197,94,0.12)', color: '#22c55e' },
  white_label: { bg: 'rgba(113,113,122,0.12)', color: '#71717a' },
  priority_support: { bg: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  data_export: { bg: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
};

function getAddonManageRoute(addonId: string): '/settings/domain' | '/settings/email' | undefined {
  switch (addonId) {
    case 'custom_domain':
      return '/settings/domain';
    case 'email_marketing':
      return '/settings/email';
    default:
      return undefined;
  }
}

function addonPriceCents(addon: AddonData): number {
  return Math.round(addon.price * 100);
}

function proratedAddonCents(priceCents: number): {
  amountCents: number;
  daysRemaining: number;
  daysInMonth: number;
} {
  const now = new Date();
  const daysInMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate();
  const daysRemaining = daysInMonth - now.getDate() + 1;
  return {
    amountCents: Math.round((daysRemaining * priceCents) / daysInMonth),
    daysRemaining,
    daysInMonth,
  };
}

const Addons: React.FC = () => {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const { data: addons = [], isLoading } = useAddons();
  const { data: billing } = useBillingDetail();
  const { data: sub } = useMySubscription();
  const [searchTerm, setSearchTerm] = useState('');
  const [enablingId, setEnablingId] = useState<string | null>(null);
  const [purchaseTarget, setPurchaseTarget] = useState<AddonData | null>(null);
  const [disableTarget, setDisableTarget] = useState<AddonData | null>(null);
  const [confirmText, setConfirmText] = useState('');
  const [showEnableSuccess, setShowEnableSuccess] = useState<string | null>(null);
  const isTrialMode = sub?.status === 'trial' || billing?.status === 'trial';

  const enableMutation = useMutation({
    mutationFn: async (addonId: string) => {
      if (isTrialMode) {
        throw new Error('addons unavailable during trial');
      }
      setEnablingId(addonId);
      await enableAddon(addonId);
    },
    onSuccess: (_, addonId) => {
      setShowEnableSuccess(addonId);
      setTimeout(() => {
        setShowEnableSuccess(null);
        setEnablingId(null);
        qc.invalidateQueries({ queryKey: ADDONS_QUERY_KEY });
      }, 2000);
    },
    onError: () => {
      setEnablingId(null);
      toast.error('Failed to enable addon');
    },
  });

  const disableMutation = useMutation({
    mutationFn: async (addonId: string) => {
      if (isTrialMode) {
        throw new Error('addons unavailable during trial');
      }
      await disableAddon(addonId);
    },
    onSuccess: () => {
      setDisableTarget(null);
      setConfirmText('');
      qc.invalidateQueries({ queryKey: ADDONS_QUERY_KEY });
    },
    onError: () => toast.error('Failed to disable addon'),
  });

  const purchaseMutation = useMutation({
    mutationFn: async (addonId: string) => {
      if (isTrialMode) {
        throw new Error('addons unavailable during trial');
      }
      await enableAddon(addonId);
    },
    onSuccess: (_, addonId) => {
      const purchased = addons.find((addon) => addon.id === addonId);
      setPurchaseTarget(null);
      qc.invalidateQueries({ queryKey: ADDONS_QUERY_KEY });
      qc.invalidateQueries({ queryKey: ['subscription', 'billing'] });
      qc.invalidateQueries({ queryKey: ['subscription', 'billing', 'history'] });
      toast.success(`${purchased?.name ?? 'Addon'} added to billing`);
    },
    onError: () => toast.error('Failed to purchase addon'),
  });

  const filtered = addons.filter(
    (a) =>
      a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const CONFIRM_PHRASE = 'disable this addon';
  const grantedAddonIds = new Set(
    (billing?.lineItems ?? [])
      .filter((item) => item.type === 'addon' && item.isGranted && item.addonId)
      .map((item) => item.addonId as string)
  );
  const enabledPaidMonthlyCents = addons
    .filter((addon) => addon.isEnabled && !addon.isFree && !grantedAddonIds.has(addon.id))
    .reduce((sum, addon) => sum + addonPriceCents(addon), 0);
  const recurringBaseCents =
    (billing?.billableSeatCount ?? 0) * (billing?.pricePerSeatCents ?? 0) + enabledPaidMonthlyCents;
  const purchasePriceCents = purchaseTarget ? addonPriceCents(purchaseTarget) : 0;
  const purchaseProration = proratedAddonCents(purchasePriceCents);

  return (
    <>
      <HelmetContainer title="Add-ons · rtylr" />
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
              Addons
            </h1>
            <p className="text-sm mt-0.5" style={{ color: 'var(--dash-text-muted)' }}>
              Extend your business with powerful add-ons
            </p>
          </div>
          <div className="relative">
            <Search
              size={13}
              className="absolute left-3 top-1/2 -translate-y-1/2"
              style={{ color: 'var(--dash-text-dim)' }}
            />
            <Input
              className="pl-9 w-56"
              placeholder="Search addons..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {isTrialMode ? (
          <div
            className="rounded-xl px-4 py-3 flex items-center gap-2 text-sm font-semibold"
            style={{
              background: 'var(--status-trial-bg)',
              color: 'var(--status-trial)',
              border: '1px solid var(--dash-border)',
            }}
          >
            <AlertTriangle size={16} className="shrink-0" />
            <span>Not available in trial mode</span>
          </div>
        ) : null}

        {isLoading ? (
          <div className="grid grid-cols-2 gap-3">
            {[1, 2, 3, 4].map((i) => (
              <div
                key={i}
                className="h-44 rounded-xl animate-pulse"
                style={{ background: 'var(--dash-hover)' }}
              />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3">
            {filtered.map((addon) => {
              const colors = COLOR_MAP[addon.id] ?? {
                bg: 'var(--dash-hover)',
                color: 'var(--dash-text-muted)',
              };
              const Icon = ICON_MAP[addon.icon] ?? MessageSquare;
              const isEnabling = enablingId === addon.id;
              const justEnabled = showEnableSuccess === addon.id;
              const manageRoute = getAddonManageRoute(addon.id);
              const isGranted = grantedAddonIds.has(addon.id);
              const handlePaidAddonClick = () => {
                if (isTrialMode) return;
                if (addon.isEnabled && manageRoute) {
                  void navigate({ to: manageRoute });
                  return;
                }
                if (addon.isEnabled) {
                  setDisableTarget(addon);
                  return;
                }
                setPurchaseTarget(addon);
              };

              return (
                <motion.div
                  key={addon.id}
                  layout
                  className="rounded-xl p-5 flex flex-col gap-4"
                  style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
                >
                  <div className="flex items-start justify-between">
                    <div
                      className="h-10 w-10 rounded-xl flex items-center justify-center shrink-0"
                      style={{ background: colors.bg }}
                    >
                      <Icon size={19} style={{ color: colors.color }} />
                    </div>
                    <div className="flex items-center gap-2">
                      {addon.isFree && (
                        <span
                          className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                          style={{ background: 'rgba(16,185,129,0.12)', color: '#10b981' }}
                        >
                          Free
                        </span>
                      )}
                      {isGranted && (
                        <span
                          className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                          style={{
                            background: 'var(--accent-muted)',
                            color: 'var(--accent-on-muted)',
                          }}
                        >
                          Granted
                        </span>
                      )}
                      {addon.isEnabled && (
                        <span
                          className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                          style={{ background: 'rgba(16,185,129,0.12)', color: '#10b981' }}
                        >
                          Enabled
                        </span>
                      )}
                    </div>
                  </div>

                  <div className="flex-1">
                    <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                      {addon.name}
                    </p>
                    <p
                      className="text-xs mt-1 leading-relaxed"
                      style={{ color: 'var(--dash-text-muted)' }}
                    >
                      {addon.description}
                    </p>
                  </div>

                  <div className="flex items-center justify-between">
                    <span
                      className="text-sm font-semibold"
                      style={{
                        color: addon.isFree || isGranted ? '#10b981' : 'var(--dash-text)',
                      }}
                    >
                      {addon.isFree ? 'Free' : isGranted ? 'Granted' : `$${addon.price}`}
                      {!addon.isFree && !isGranted && addon.price > 0 && (
                        <span
                          className="text-xs font-normal"
                          style={{ color: 'var(--dash-text-dim)' }}
                        >
                          /mo
                        </span>
                      )}
                    </span>

                    {!addon.isFree ? (
                      <button
                        type="button"
                        onClick={handlePaidAddonClick}
                        disabled={
                          isTrialMode ||
                          (purchaseMutation.isPending && purchaseTarget?.id === addon.id)
                        }
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all disabled:cursor-not-allowed disabled:opacity-60"
                        style={{
                          background:
                            addon.isEnabled || isTrialMode ? 'var(--dash-hover)' : '#4583C4',
                          color: addon.isEnabled || isTrialMode ? 'var(--dash-text-muted)' : '#fff',
                          border:
                            addon.isEnabled || isTrialMode
                              ? '1px solid var(--dash-border)'
                              : 'none',
                        }}
                      >
                        {isTrialMode
                          ? 'Trial mode'
                          : purchaseMutation.isPending && purchaseTarget?.id === addon.id
                            ? 'Purchasing...'
                            : addon.isEnabled
                              ? 'Manage'
                              : 'Purchase'}
                      </button>
                    ) : isEnabling ? (
                      <div className="flex items-center gap-2 px-3 py-1.5">
                        {justEnabled ? (
                          <motion.span
                            initial={{ scale: 0 }}
                            animate={{ scale: 1 }}
                            className="text-xs font-semibold"
                            style={{ color: '#10b981' }}
                          >
                            Enabled!
                          </motion.span>
                        ) : (
                          <>
                            <Loader2
                              size={14}
                              className="animate-spin"
                              style={{ color: '#4583C4' }}
                            />
                            <span className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                              Setting up...
                            </span>
                          </>
                        )}
                      </div>
                    ) : addon.isEnabled ? (
                      <button
                        type="button"
                        onClick={() => {
                          if (!isTrialMode) setDisableTarget(addon);
                        }}
                        disabled={isTrialMode}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all disabled:cursor-not-allowed disabled:opacity-60"
                        style={{
                          background: 'var(--dash-hover)',
                          color: 'var(--dash-text-muted)',
                          border: '1px solid var(--dash-border)',
                        }}
                      >
                        {isTrialMode ? 'Trial mode' : 'Disable'}
                      </button>
                    ) : (
                      <button
                        type="button"
                        onClick={() => {
                          if (!isTrialMode) enableMutation.mutate(addon.id);
                        }}
                        disabled={isTrialMode}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all disabled:cursor-not-allowed disabled:opacity-60"
                        style={{
                          background: isTrialMode ? 'var(--dash-hover)' : '#4583C4',
                          color: isTrialMode ? 'var(--dash-text-muted)' : '#fff',
                          border: isTrialMode ? '1px solid var(--dash-border)' : 'none',
                        }}
                      >
                        {isTrialMode ? 'Trial mode' : 'Enable'}
                      </button>
                    )}
                  </div>
                </motion.div>
              );
            })}
          </div>
        )}

        {filtered.length === 0 && !isLoading && (
          <EmptyState
            icon={Search}
            title={t('addons.emptyMatchTitle')}
            action={{ label: t('common.clearSearch'), onClick: () => setSearchTerm('') }}
          />
        )}

        {/* Paid addon purchase modal */}
        <AnimatePresence>
          {purchaseTarget && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50 flex items-center justify-center p-4"
              style={{ background: 'rgba(0,0,0,0.5)' }}
              onClick={() => {
                if (!purchaseMutation.isPending) setPurchaseTarget(null);
              }}
            >
              <motion.div
                initial={{ scale: 0.95, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.95, opacity: 0 }}
                onClick={(e) => e.stopPropagation()}
                className="rounded-xl p-6 w-full max-w-md"
                style={{
                  background: 'var(--dash-surface)',
                  border: '1px solid var(--dash-border)',
                }}
              >
                <div className="flex items-start gap-3 mb-5">
                  <div
                    className="h-10 w-10 rounded-xl flex items-center justify-center shrink-0"
                    style={{
                      background: (
                        COLOR_MAP[purchaseTarget.id] ?? {
                          bg: 'var(--dash-hover)',
                        }
                      ).bg,
                    }}
                  >
                    {React.createElement(ICON_MAP[purchaseTarget.icon] ?? ShoppingCart, {
                      size: 19,
                      style: {
                        color: (
                          COLOR_MAP[purchaseTarget.id] ?? {
                            color: 'var(--dash-text-muted)',
                          }
                        ).color,
                      },
                    })}
                  </div>
                  <div className="min-w-0">
                    <h3 className="text-base font-semibold" style={{ color: 'var(--dash-text)' }}>
                      Purchase {purchaseTarget.name}
                    </h3>
                    <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
                      This addon is added to your current billing month and recurring monthly total.
                    </p>
                  </div>
                </div>

                <div
                  className="rounded-xl overflow-hidden mb-5"
                  style={{ border: '1px solid var(--dash-border)' }}
                >
                  {[
                    ['Addon price', `${centsToDisplay(purchasePriceCents)}/mo`],
                    ['This month', `${centsToDisplay(purchaseProration.amountCents)} prorated`],
                    ['Current monthly total', centsToDisplay(recurringBaseCents)],
                    ['New monthly total', centsToDisplay(recurringBaseCents + purchasePriceCents)],
                  ].map(([label, value]) => (
                    <div
                      key={label}
                      className="flex items-center justify-between gap-4 px-4 py-3"
                      style={{ borderBottom: '1px solid var(--dash-border)' }}
                    >
                      <span className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                        {label}
                      </span>
                      <span className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
                        {value}
                      </span>
                    </div>
                  ))}
                </div>

                <p className="text-xs mb-5" style={{ color: 'var(--dash-text-dim)' }}>
                  Proration covers {purchaseProration.daysRemaining} of{' '}
                  {purchaseProration.daysInMonth} days remaining in this billing month.
                </p>

                <div className="flex gap-2 justify-end">
                  <button
                    type="button"
                    onClick={() => setPurchaseTarget(null)}
                    disabled={purchaseMutation.isPending}
                    className="px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-60"
                    style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      if (!isTrialMode) purchaseMutation.mutate(purchaseTarget.id);
                    }}
                    disabled={isTrialMode || purchaseMutation.isPending}
                    className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity disabled:cursor-not-allowed disabled:opacity-60"
                    style={{
                      background: isTrialMode ? 'var(--dash-hover)' : '#4583C4',
                      color: isTrialMode ? 'var(--dash-text-muted)' : '#fff',
                    }}
                  >
                    {isTrialMode
                      ? 'Trial mode'
                      : purchaseMutation.isPending
                        ? 'Purchasing...'
                        : 'Confirm purchase'}
                  </button>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Disable confirmation modal */}
        <AnimatePresence>
          {disableTarget && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-50 flex items-center justify-center"
              style={{ background: 'rgba(0,0,0,0.5)' }}
              onClick={() => {
                setDisableTarget(null);
                setConfirmText('');
              }}
            >
              <motion.div
                initial={{ scale: 0.95, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                exit={{ scale: 0.95, opacity: 0 }}
                onClick={(e) => e.stopPropagation()}
                className="rounded-xl p-6 w-full max-w-md"
                style={{
                  background: 'var(--dash-surface)',
                  border: '1px solid var(--dash-border)',
                }}
              >
                <div className="flex items-center gap-3 mb-4">
                  <div
                    className="h-10 w-10 rounded-xl flex items-center justify-center"
                    style={{ background: 'rgba(239,68,68,0.12)' }}
                  >
                    <AlertTriangle size={20} style={{ color: '#ef4444' }} />
                  </div>
                  <div>
                    <h3 className="text-base font-semibold" style={{ color: 'var(--dash-text)' }}>
                      Disable {disableTarget.name}?
                    </h3>
                    <p className="text-xs" style={{ color: 'var(--dash-text-muted)' }}>
                      Your data will be preserved but the feature will be hidden.
                    </p>
                  </div>
                </div>

                <p className="text-xs mb-2" style={{ color: 'var(--dash-text-muted)' }}>
                  Type <strong style={{ color: 'var(--dash-text)' }}>{CONFIRM_PHRASE}</strong> to
                  confirm:
                </p>
                <input
                  value={confirmText}
                  onChange={(e) => setConfirmText(e.target.value)}
                  placeholder={CONFIRM_PHRASE}
                  autoFocus
                  className="w-full h-10 rounded-lg px-3 text-sm outline-none mb-4"
                  style={{
                    background: 'var(--dash-bg)',
                    border: '1px solid var(--dash-border)',
                    color: 'var(--dash-text)',
                  }}
                />

                <div className="flex gap-2 justify-end">
                  <button
                    type="button"
                    onClick={() => {
                      setDisableTarget(null);
                      setConfirmText('');
                    }}
                    className="px-4 py-2 rounded-lg text-sm font-medium"
                    style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      if (!isTrialMode) disableMutation.mutate(disableTarget.id);
                    }}
                    disabled={isTrialMode || confirmText.toLowerCase() !== CONFIRM_PHRASE}
                    className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity"
                    style={{
                      background:
                        !isTrialMode && confirmText.toLowerCase() === CONFIRM_PHRASE
                          ? '#ef4444'
                          : 'var(--dash-hover)',
                      color:
                        !isTrialMode && confirmText.toLowerCase() === CONFIRM_PHRASE
                          ? '#fff'
                          : 'var(--dash-text-dim)',
                      opacity:
                        !isTrialMode && confirmText.toLowerCase() === CONFIRM_PHRASE ? 1 : 0.5,
                      cursor:
                        !isTrialMode && confirmText.toLowerCase() === CONFIRM_PHRASE
                          ? 'pointer'
                          : 'not-allowed',
                    }}
                  >
                    {isTrialMode
                      ? 'Trial mode'
                      : disableMutation.isPending
                        ? 'Disabling...'
                        : 'Disable'}
                  </button>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </>
  );
};

export default Addons;
