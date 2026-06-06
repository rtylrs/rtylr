import React from 'react';
import {
  Utensils,
  Package,
  Users,
  UserCheck,
  DollarSign,
  GitMerge,
  Tag,
  BookOpen,
  Rss,
  History,
  Sparkles,
  LucideProps,
} from 'lucide-react';
import { useDesktopStore, WindowType } from '../../stores/useDesktopStore';

interface DockItem {
  key: WindowType | string;
  label: string;
  icon: React.FC<LucideProps>;
  href?: string;
  windowType?: WindowType;
  accent: string;
}

const LEFT_ITEMS: DockItem[] = [
  { key: 'pos', label: 'POS', icon: Utensils, href: '/pos', accent: 'rgba(232,230,225,0.08)' },
  { key: 'erp', label: 'ERP', icon: Package, href: '/erp', accent: 'rgba(232,230,225,0.08)' },
  { key: 'crm', label: 'CRM', icon: Users, href: '/crm', accent: 'rgba(232,230,225,0.08)' },
  { key: 'hr', label: 'HR', icon: UserCheck, href: '/hr', accent: 'rgba(232,230,225,0.08)' },
  {
    key: 'finance',
    label: 'Finance',
    icon: DollarSign,
    href: '/finance',
    accent: 'rgba(232,230,225,0.08)',
  },
  {
    key: 'flow',
    label: 'Flow',
    icon: GitMerge,
    href: '/flow',
    accent: 'rgba(232,230,225,0.08)',
  },
];

const RIGHT_ITEMS: DockItem[] = [
  {
    key: 'pricing',
    label: 'Pricing',
    icon: Tag,
    windowType: 'pricing',
    accent: 'rgba(232,230,225,0.08)',
  },
  {
    key: 'docs',
    label: 'Docs',
    icon: BookOpen,
    windowType: 'docs',
    accent: 'rgba(232,230,225,0.08)',
  },
  {
    key: 'blog',
    label: 'Blog',
    icon: Rss,
    href: '/blog',
    accent: 'rgba(232,230,225,0.08)',
  },
  {
    key: 'changelog',
    label: 'Changelog',
    icon: History,
    href: 'https://rtylr.com/changelog',
    accent: 'rgba(232,230,225,0.08)',
  },
  {
    key: 'whatsnew',
    label: "What's New",
    icon: Sparkles,
    href: 'https://rtylr.com/changelog',
    accent: 'rgba(232,230,225,0.08)',
  },
];

interface DockIconProps {
  item: DockItem;
  isOpen: boolean;
}

const DockIcon: React.FC<DockIconProps> = ({ item, isOpen }) => {
  const addWindow = useDesktopStore((s) => s.addWindow);
  const Icon = item.icon;

  const handleClick = () => {
    if (item.windowType) addWindow(item.windowType);
    else if (item.href) {
      if (item.href.startsWith('/')) window.location.href = item.href;
      else window.open(item.href, '_blank', 'noopener,noreferrer');
    }
  };

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 5,
        cursor: 'pointer',
        position: 'relative',
      }}
      role="button"
      tabIndex={0}
      aria-label={item.label}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleClick();
        }
      }}
      title={item.label}
    >
      {/* App icon square */}
      <div
        className="dock-icon-bg"
        style={{
          width: 44,
          height: 44,
          borderRadius: 12,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: isOpen ? 'rgba(232,230,225,0.13)' : 'rgba(255,255,255,0.06)',
          border: `1px solid ${isOpen ? 'rgba(232,230,225,0.22)' : 'rgba(255,255,255,0.09)'}`,
          transition: 'transform 0.12s ease, background 0.12s ease, border-color 0.12s ease',
          boxShadow: isOpen ? '0 0 0 1px rgba(232,230,225,0.08) inset' : 'none',
        }}
        onMouseEnter={(e) => {
          const el = e.currentTarget as HTMLDivElement;
          el.style.transform = 'translateY(-3px) scale(1.06)';
          el.style.background = 'rgba(255,255,255,0.1)';
        }}
        onMouseLeave={(e) => {
          const el = e.currentTarget as HTMLDivElement;
          el.style.transform = 'translateY(0) scale(1)';
          el.style.background = isOpen ? 'rgba(232,230,225,0.13)' : 'rgba(255,255,255,0.06)';
        }}
      >
        <Icon size={20} color={isOpen ? '#e8e6e1' : 'rgba(255,255,255,0.65)'} strokeWidth={1.5} />
      </div>

      {/* Label */}
      <span
        style={{
          fontSize: 9.5,
          fontWeight: 500,
          color: isOpen ? 'rgba(232,230,225,0.9)' : 'rgba(255,255,255,0.45)',
          letterSpacing: '0.01em',
          lineHeight: 1,
          whiteSpace: 'nowrap',
        }}
      >
        {item.label}
      </span>

      {/* Open indicator dot */}
      {isOpen && (
        <div
          style={{
            position: 'absolute',
            bottom: -2,
            left: '50%',
            transform: 'translateX(-50%)',
            width: 3,
            height: 3,
            borderRadius: '50%',
            background: 'rgba(232,230,225,0.5)',
            marginTop: -2,
          }}
        />
      )}
    </div>
  );
};

interface DockColumnProps {
  items: DockItem[];
  openTypes: Set<string>;
}

const DockColumn: React.FC<DockColumnProps> = ({ items, openTypes }) => (
  <div
    style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      gap: 4,
      padding: '12px 10px',
      background: 'rgba(14,14,16,0.72)',
      backdropFilter: 'blur(24px)',
      border: '1px solid rgba(255,255,255,0.07)',
      borderRadius: 16,
      boxShadow: '0 8px 32px rgba(0,0,0,0.35)',
    }}
  >
    {items.map((item) => (
      <DockIcon key={item.key} item={item} isOpen={openTypes.has(item.key)} />
    ))}
  </div>
);

export const Dock: React.FC = () => {
  const windows = useDesktopStore((s) => s.windows);
  const openTypes = new Set(windows.filter((w) => !w.minimized).map((w) => w.type));

  return (
    <>
      <div
        style={{
          position: 'absolute',
          left: 16,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 8500,
        }}
      >
        <DockColumn items={LEFT_ITEMS} openTypes={openTypes} />
      </div>
      <div
        style={{
          position: 'absolute',
          right: 16,
          top: '50%',
          transform: 'translateY(-50%)',
          zIndex: 8500,
        }}
      >
        <DockColumn items={RIGHT_ITEMS} openTypes={openTypes} />
      </div>
    </>
  );
};
