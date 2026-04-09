import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  MessageSquare, Heart, BarChart2, Paintbrush, Search,
  Bell, Globe, ShieldCheck, Smartphone, Zap,
} from 'lucide-react';
import { useUserStore } from '../../stores/userStore';
import { Input } from '../../components/ui';
import { cn } from '../../lib/utils';

type Addon = {
  key: string;
  name: string;
  description: string;
  price: string | null; // null = free
  icon: React.ComponentType<{ size?: number }>;
  iconBg: string;
  iconColor: string;
  badge: string | null;
  free: boolean;
};

const ADDONS: Addon[] = [
  {
    key: 'ai_chat',
    name: 'AI Chat',
    description: 'Add an AI-powered chat assistant to help customers with menu questions and recommendations.',
    price: null,
    icon: MessageSquare,
    iconBg: 'rgba(139,92,246,0.12)',
    iconColor: '#8b5cf6',
    badge: 'Free',
    free: true,
  },
  {
    key: 'loyalty',
    name: 'Loyalty Program',
    description: 'Reward repeat customers with points and exclusive discounts to boost retention.',
    price: '$12',
    icon: Heart,
    iconBg: 'rgba(239,68,68,0.12)',
    iconColor: '#ef4444',
    badge: null,
    free: false,
  },
  {
    key: 'advanced_analytics',
    name: 'Advanced Analytics',
    description: 'Deep insights into your sales, customer behavior, and menu performance over time.',
    price: '$19',
    icon: BarChart2,
    iconBg: 'rgba(59,130,246,0.12)',
    iconColor: '#3b82f6',
    badge: 'New',
    free: false,
  },
  {
    key: 'custom_branding',
    name: 'Custom Branding',
    description: 'Fully customize your customer-facing menus and QR codes with your brand identity.',
    price: '$7',
    icon: Paintbrush,
    iconBg: 'rgba(245,158,11,0.12)',
    iconColor: '#f59e0b',
    badge: null,
    free: false,
  },
  {
    key: 'push_notifications',
    name: 'Push Notifications',
    description: 'Send order updates and promotional messages directly to your customers\' devices.',
    price: '$9',
    icon: Bell,
    iconBg: 'rgba(16,185,129,0.12)',
    iconColor: '#10b981',
    badge: 'Coming Soon',
    free: false,
  },
  {
    key: 'multi_language',
    name: 'Multi-Language Menu',
    description: 'Serve your menu in multiple languages to reach a wider audience.',
    price: '$14',
    icon: Globe,
    iconBg: 'rgba(69,131,196,0.12)',
    iconColor: '#4583C4',
    badge: 'Coming Soon',
    free: false,
  },
  {
    key: 'fraud_protection',
    name: 'Fraud Protection',
    description: 'Advanced order verification and fraud detection to protect your revenue.',
    price: '$6',
    icon: ShieldCheck,
    iconBg: 'rgba(139,92,246,0.12)',
    iconColor: '#8b5cf6',
    badge: 'Coming Soon',
    free: false,
  },
  {
    key: 'mobile_app',
    name: 'Branded Mobile App',
    description: 'Your own white-labeled mobile app with your brand, on iOS and Android.',
    price: '$49',
    icon: Smartphone,
    iconBg: 'rgba(245,158,11,0.12)',
    iconColor: '#f59e0b',
    badge: 'Coming Soon',
    free: false,
  },
  {
    key: 'instant_payouts',
    name: 'Instant Payouts',
    description: 'Get your earnings deposited instantly instead of waiting for standard payout cycles.',
    price: '$5',
    icon: Zap,
    iconBg: 'rgba(16,185,129,0.12)',
    iconColor: '#10b981',
    badge: 'Coming Soon',
    free: false,
  },
];

const Addons: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useUserStore();
  const [searchTerm, setSearchTerm] = useState('');
  // AI Chat is free but disabled by default
  const [enabled, setEnabled] = useState<Record<string, boolean>>({});
  const planType = user?.planType;

  const filtered = ADDONS.filter(
    (a) =>
      a.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.description.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const toggleAddon = (addon: Addon) => {
    if (!addon.free && planType === 'free') {
      navigate('/settings/billing');
      return;
    }
    if (addon.badge === 'Coming Soon') return;
    setEnabled((prev) => ({ ...prev, [addon.key]: !prev[addon.key] }));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold" style={{ color: 'var(--hub-text)' }}>Addons</h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Extend your business with powerful add-ons
          </p>
        </div>
        <div className="relative">
          <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2" style={{ color: 'var(--hub-text-dim)' }} />
          <Input
            className="pl-9 w-56"
            placeholder="Search addons..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      {/* Upgrade Banner */}
      {planType === 'free' && (
        <div
          className="rounded-xl p-4 flex items-center justify-between"
          style={{ background: 'rgba(69,131,196,0.08)', border: '1px solid rgba(69,131,196,0.2)' }}
        >
          <div>
            <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              Upgrade to unlock paid add-ons
            </p>
            <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
              Free add-ons like AI Chat are available on all plans
            </p>
          </div>
          <button
            onClick={() => navigate('/settings/billing')}
            className="px-4 py-2 rounded-lg text-sm font-semibold"
            style={{ background: '#4583C4', color: '#fff' }}
          >
            Upgrade
          </button>
        </div>
      )}

      {/* Addons Grid */}
      <div className="grid grid-cols-2 gap-3">
        {filtered.map((addon) => {
          const Icon = addon.icon;
          const isEnabled = !!enabled[addon.key];
          const isComingSoon = addon.badge === 'Coming Soon';

          return (
            <div
              key={addon.key}
              className="rounded-xl p-5 flex flex-col gap-4"
              style={{
                background: 'var(--hub-card)',
                border: '1px solid var(--hub-border)',
                opacity: isComingSoon ? 0.6 : 1,
              }}
            >
              <div className="flex items-start justify-between">
                <div
                  className="h-10 w-10 rounded-xl flex items-center justify-center shrink-0"
                  style={{ background: addon.iconBg }}
                >
                  <Icon size={19} style={{ color: addon.iconColor }} />
                </div>
                <div className="flex items-center gap-2">
                  {addon.badge && (
                    <span
                      className="text-[11px] font-semibold px-2 py-0.5 rounded-full"
                      style={
                        addon.badge === 'Free'
                          ? { background: 'rgba(16,185,129,0.12)', color: '#10b981' }
                          : addon.badge === 'New'
                          ? { background: 'rgba(69,131,196,0.12)', color: '#4583C4' }
                          : { background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }
                      }
                    >
                      {addon.badge}
                    </span>
                  )}
                  {isEnabled && !isComingSoon && (
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
                <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>{addon.name}</p>
                <p className="text-xs mt-1 leading-relaxed" style={{ color: 'var(--hub-text-muted)' }}>
                  {addon.description}
                </p>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-sm font-semibold" style={{ color: addon.free ? '#10b981' : 'var(--hub-text)' }}>
                  {addon.free ? 'Free' : addon.price}
                  {!addon.free && addon.price && (
                    <span className="text-xs font-normal" style={{ color: 'var(--hub-text-dim)' }}>/mo</span>
                  )}
                </span>
                <button
                  onClick={() => toggleAddon(addon)}
                  disabled={isComingSoon}
                  className="px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
                  style={
                    isComingSoon
                      ? { background: 'var(--hub-hover)', color: 'var(--hub-text-dim)', cursor: 'default' }
                      : isEnabled
                      ? { background: 'var(--hub-hover)', color: 'var(--hub-text-muted)', border: '1px solid var(--hub-border)' }
                      : { background: '#4583C4', color: '#fff' }
                  }
                >
                  {isComingSoon ? 'Soon' : isEnabled ? 'Disable' : 'Enable'}
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {filtered.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <p className="text-sm font-medium" style={{ color: 'var(--hub-text-muted)' }}>
            No addons match your search
          </p>
        </div>
      )}
    </div>
  );
};

export default Addons;
