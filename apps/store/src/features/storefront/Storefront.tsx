import React from 'react';
import { motion } from 'motion/react';
import { Globe, Palette, Image, Type, Link2, MessageSquare } from 'lucide-react';
import { useStorefrontConfig } from './useStorefront';
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

const Storefront: React.FC = () => {
  const { data: config, isLoading, error } = useStorefrontConfig();

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-3">
        <p className="text-sm font-medium" style={{ color: '#ef4444' }}>
          Failed to load storefront config
        </p>
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          Please check your connection and try again.
        </p>
      </div>
    );
  }

  const sections = [
    {
      title: 'Theme',
      icon: Palette,
      fields: [
        { label: 'Theme Name', value: config?.themeName },
        { label: 'Primary Color', value: config?.primaryColor },
        { label: 'Secondary Color', value: config?.secondaryColor },
      ],
    },
    {
      title: 'Branding',
      icon: Image,
      fields: [
        { label: 'Logo URL', value: config?.logoUrl },
        { label: 'Favicon URL', value: config?.faviconUrl },
        { label: 'Announcement Bar', value: config?.announcementBar },
      ],
    },
    {
      title: 'Domain',
      icon: Globe,
      fields: [{ label: 'Custom Domain', value: config?.customDomain }],
    },
    {
      title: 'SEO',
      icon: Type,
      fields: [
        { label: 'Meta Title', value: config?.metaTitle },
        { label: 'Meta Description', value: config?.metaDescription },
      ],
    },
    {
      title: 'Social Links',
      icon: Link2,
      fields: Object.entries(config?.socialLinks ?? {}).map(([key, val]) => ({
        label: key.charAt(0).toUpperCase() + key.slice(1),
        value: val,
      })),
    },
    {
      title: 'Footer',
      icon: MessageSquare,
      fields: [{ label: 'Footer Text', value: config?.footerText }],
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-4xl"
    >
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Globe size={16} style={{ color: TEAL }} />
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Storefront Builder
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
          {Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="rounded-xl p-5 space-y-3"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <SkeletonRow cols={2} />
              <SkeletonRow cols={2} />
            </div>
          ))}
        </div>
      ) : !config ? (
        <motion.div variants={itemVariants} className="flex flex-col items-center py-16 gap-3">
          <Globe size={32} style={{ color: 'var(--hub-text-dim)' }} />
          <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
            No storefront configuration found
          </p>
          <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
            Set up your online storefront to get started.
          </p>
        </motion.div>
      ) : (
        <div className="space-y-4">
          {sections.map((section) => {
            const Icon = section.icon;
            return (
              <motion.div
                key={section.title}
                variants={itemVariants}
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
                  {section.fields.length === 0 ? (
                    <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                      Not configured
                    </p>
                  ) : (
                    section.fields.map((field) => (
                      <div key={field.label} className="flex items-center gap-4">
                        <label
                          className="text-xs font-medium w-36 shrink-0"
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
                            (e.currentTarget as HTMLElement).style.borderColor =
                              'var(--hub-border)';
                          }}
                        />
                      </div>
                    ))
                  )}
                </div>
              </motion.div>
            );
          })}
        </div>
      )}
    </motion.div>
  );
};

export default Storefront;
