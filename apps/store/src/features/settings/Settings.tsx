import React from 'react';
import { motion } from 'motion/react';
import { Settings as SettingsIcon, DollarSign, Bell, Globe } from 'lucide-react';
import { useStoreSettings } from './useSettings';
import { SkeletonRow } from '../../shared/components/Skeleton';

const TEAL = '#14b8a6';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Settings: React.FC = () => {
  const { data: settings, isLoading, error } = useStoreSettings();

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>
          Failed to load settings
        </p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Please check your connection and try again.
        </p>
      </div>
    );
  }

  const sections = [
    {
      title: 'General',
      icon: SettingsIcon,
      fields: [
        { label: 'Store Name', value: settings?.storeName, type: 'text' as const },
        { label: 'Order Prefix', value: settings?.orderPrefix, type: 'text' as const },
        { label: 'Timezone', value: settings?.timezone, type: 'text' as const },
      ],
    },
    {
      title: 'Currency & Tax',
      icon: DollarSign,
      fields: [
        { label: 'Currency', value: settings?.currency, type: 'text' as const },
        { label: 'Tax Rate (%)', value: settings?.taxRate?.toString(), type: 'text' as const },
        {
          label: 'Tax Inclusive',
          value: settings?.taxInclusive ? 'Yes' : 'No',
          type: 'text' as const,
        },
      ],
    },
    {
      title: 'Inventory',
      icon: Globe,
      fields: [
        {
          label: 'Low Stock Threshold',
          value: settings?.lowStockThreshold?.toString(),
          type: 'text' as const,
        },
      ],
    },
    {
      title: 'Notifications',
      icon: Bell,
      fields: [
        {
          label: 'Email Notifications',
          value: settings?.emailNotifications ? 'Enabled' : 'Disabled',
          type: 'text' as const,
        },
        {
          label: 'Order Confirmation',
          value: settings?.orderConfirmationEmail ? 'Enabled' : 'Disabled',
          type: 'text' as const,
        },
        {
          label: 'Shipping Updates',
          value: settings?.shippingNotificationEmail ? 'Enabled' : 'Disabled',
          type: 'text' as const,
        },
      ],
    },
  ];

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-4xl"
    >
      <motion.div variants={itemVariants as any} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <SettingsIcon size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Store Settings
          </h2>
        </div>
        <button
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-semibold"
          style={{ background: TEAL, color: '#fff' }}
        >
          Save Changes
        </button>
      </motion.div>

      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className="rounded-xl p-5 space-y-3"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <SkeletonRow cols={2} />
              <SkeletonRow cols={2} />
              <SkeletonRow cols={2} />
            </div>
          ))}
        </div>
      ) : !settings ? (
        <motion.div
          variants={itemVariants as any}
          className="flex flex-col items-center py-16 gap-3"
        >
          <SettingsIcon size={32} style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
            No settings found
          </p>
          <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
            Configure your store settings to get started.
          </p>
        </motion.div>
      ) : (
        <div className="space-y-4">
          {sections.map((section) => {
            const Icon = section.icon;
            return (
              <motion.div
                key={section.title}
                variants={itemVariants as any}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center gap-2 mb-4">
                  <Icon size={14} style={{ color: TEAL }} />
                  <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {section.title}
                  </h3>
                </div>
                <div className="space-y-3">
                  {section.fields.map((field) => (
                    <div key={field.label} className="flex items-center gap-4">
                      <label
                        className="text-xs font-medium w-40 shrink-0"
                        style={{ color: 'var(--hub-text-muted)' }}
                      >
                        {field.label}
                      </label>
                      <input
                        type="text"
                        defaultValue={field.value ?? ''}
                        className="flex-1 px-3 py-2 rounded-lg text-sm outline-none"
                        style={{
                          background: 'var(--hub-input-bg)',
                          border: '1px solid var(--hub-border)',
                          color: 'var(--hub-text)',
                        }}
                        onFocus={(e) => {
                          (e.currentTarget as HTMLElement).style.borderColor =
                            'rgba(20,184,166,0.5)';
                        }}
                        onBlur={(e) => {
                          (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
                        }}
                      />
                    </div>
                  ))}
                </div>
              </motion.div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default Settings;
