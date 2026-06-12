import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
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
import { Input } from '../../shared/components/ui';
import { unwrap } from '@voxire/api';
import { safe } from '../../shared/lib/api';
import { ADDONS_QUERY_KEY, type AddonData, useAddons } from '../../shared/hooks/queries/useAddons';

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
  menu: { bg: 'rgba(69,131,196,0.12)', color: '#4583C4' },
  loyalty: { bg: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  ai_chat: { bg: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
  analytics: { bg: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
  custom_branding: { bg: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
  multi_language: { bg: 'rgba(16,185,129,0.12)', color: '#10b981' },
  store: { bg: 'rgba(20,184,166,0.12)', color: '#14b8a6' },
  book: { bg: 'rgba(14,165,233,0.12)', color: '#0ea5e9' },
  support: { bg: 'rgba(69,131,196,0.12)', color: '#4583C4' },
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

const Addons: React.FC = () => {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const { data: addons = [], isLoading } = useAddons();
  const [searchTerm, setSearchTerm] = useState('');
  const [enablingId, setEnablingId] = useState<string | null>(null);
  const [disableTarget, setDisableTarget] = useState<AddonData | null>(null);
  const [confirmText, setConfirmText] = useState('');
  const [showEnableSuccess, setShowEnableSuccess] = useState<string | null>(null);

  const enableMutation = useMutation({
    mutationFn: async (addonId: string) => {
      setEnablingId(addonId);
      await unwrap(safe.post('/addons/enable', { addonId }));
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
      await unwrap(safe.post('/addons/disable', { addonId }));
    },
    onSuccess: () => {
      setDisableTarget(null);
      setConfirmText('');
      qc.invalidateQueries({ queryKey: ADDONS_QUERY_KEY });
    },
    onError: () => toast.error('Failed to disable addon'),
  });

  const filtered = addons.filter(
    (a) =>
      a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const CONFIRM_PHRASE = 'disable this addon';

  const goToBilling = () => {
    void navigate({ to: '/settings/billing' });
  };

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
              const handlePaidAddonClick = () => {
                if (addon.isEnabled && manageRoute) {
                  void navigate({ to: manageRoute });
                  return;
                }
                goToBilling();
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
                      style={{ color: addon.isFree ? '#10b981' : 'var(--dash-text)' }}
                    >
                      {addon.isFree ? 'Free' : `$${addon.price}`}
                      {!addon.isFree && addon.price > 0 && (
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
                        onClick={handlePaidAddonClick}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                        style={{
                          background: addon.isEnabled ? 'var(--dash-hover)' : '#4583C4',
                          color: addon.isEnabled ? 'var(--dash-text-muted)' : '#fff',
                          border: addon.isEnabled ? '1px solid var(--dash-border)' : 'none',
                        }}
                      >
                        {addon.isEnabled ? 'Manage' : 'Purchase'}
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
                        onClick={() => setDisableTarget(addon)}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                        style={{
                          background: 'var(--dash-hover)',
                          color: 'var(--dash-text-muted)',
                          border: '1px solid var(--dash-border)',
                        }}
                      >
                        Disable
                      </button>
                    ) : (
                      <button
                        onClick={() => enableMutation.mutate(addon.id)}
                        className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                        style={{ background: '#4583C4', color: '#fff' }}
                      >
                        Enable
                      </button>
                    )}
                  </div>
                </motion.div>
              );
            })}
          </div>
        )}

        {filtered.length === 0 && !isLoading && (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <p className="text-sm font-medium" style={{ color: 'var(--dash-text-muted)' }}>
              No addons match your search
            </p>
          </div>
        )}

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
                    onClick={() => disableMutation.mutate(disableTarget.id)}
                    disabled={confirmText.toLowerCase() !== CONFIRM_PHRASE}
                    className="px-4 py-2 rounded-lg text-sm font-semibold transition-opacity"
                    style={{
                      background:
                        confirmText.toLowerCase() === CONFIRM_PHRASE
                          ? '#ef4444'
                          : 'var(--dash-hover)',
                      color:
                        confirmText.toLowerCase() === CONFIRM_PHRASE
                          ? '#fff'
                          : 'var(--dash-text-dim)',
                      opacity: confirmText.toLowerCase() === CONFIRM_PHRASE ? 1 : 0.5,
                      cursor:
                        confirmText.toLowerCase() === CONFIRM_PHRASE ? 'pointer' : 'not-allowed',
                    }}
                  >
                    {disableMutation.isPending ? 'Disabling...' : 'Disable'}
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
