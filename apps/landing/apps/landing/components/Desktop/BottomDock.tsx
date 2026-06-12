import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { useDesktopStore, WindowType } from '../../stores/useDesktopStore';
import { POSIcon } from './icons/POSIcon';
import { ERPIcon } from './icons/ERPIcon';
import { CRMIcon } from './icons/CRMIcon';
import { HRIcon } from './icons/HRIcon';
import { FinanceIcon } from './icons/FinanceIcon';
import { FlowIcon } from './icons/FlowIcon';
import { PricingIcon } from './icons/PricingIcon';
import { DocsIcon } from './icons/DocsIcon';
import { BlogIcon } from './icons/BlogIcon';
import { ChangelogIcon } from './icons/ChangelogIcon';
import { TerminalIcon } from './icons/TerminalIcon';
import { DoomIcon } from './icons/DoomIcon';
import { StatusIcon } from './icons/StatusIcon';
import { SnakeIcon } from './icons/SnakeIcon';
import { LandingIcon } from './icons/LandingIcon';
import { Keyboard } from 'lucide-react';

const ICON_MAP: Record<WindowType, React.FC<{ size?: number }>> = {
  pos: POSIcon,
  erp: ERPIcon,
  crm: CRMIcon,
  hr: HRIcon,
  finance: FinanceIcon,
  flow: FlowIcon,
  pricing: PricingIcon,
  docs: DocsIcon,
  blog: BlogIcon,
  changelog: ChangelogIcon,
  landing: LandingIcon,
  settings: ({ size = 28 }) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="3" stroke="rgba(255,255,255,0.8)" strokeWidth="1.5" />
      <path
        d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"
        stroke="rgba(255,255,255,0.8)"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  ),
  about: ({ size = 28 }) => (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none">
      <circle cx="12" cy="12" r="10" stroke="rgba(255,255,255,0.8)" strokeWidth="1.5" />
      <path
        d="M12 8h.01M12 11v5"
        stroke="rgba(255,255,255,0.8)"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  ),
  shortcuts: ({ size = 28 }) => (
    <Keyboard size={size} color="rgba(255,255,255,0.8)" strokeWidth={1.5} />
  ),
  terminal: TerminalIcon,
  doom: DoomIcon,
  status: StatusIcon,
  snake: SnakeIcon,
};

export const BottomDock: React.FC = () => {
  const windows = useDesktopStore((s) => s.windows);
  const restoreWindow = useDesktopStore((s) => s.restoreWindow);
  const closeWindow = useDesktopStore((s) => s.closeWindow);

  const minimized = windows.filter((w) => w.minimized);
  if (minimized.length === 0) return null;

  return (
    <div
      style={{
        position: 'absolute',
        bottom: 16,
        left: '50%',
        transform: 'translateX(-50%)',
        zIndex: 8500,
        display: 'flex',
        alignItems: 'flex-end',
        gap: 8,
        padding: '8px 12px',
        background: 'rgba(12,12,16,0.82)',
        backdropFilter: 'blur(24px)',
        border: '1px solid rgba(255,255,255,0.08)',
        borderRadius: 16,
        boxShadow: '0 8px 32px rgba(0,0,0,0.5)',
      }}
    >
      <AnimatePresence initial={false}>
        {minimized.map((win) => {
          const Icon = ICON_MAP[win.type];
          return (
            <motion.div
              key={win.id}
              layout
              initial={{ scale: 0.4, opacity: 0, y: 18 }}
              animate={{ scale: 1, opacity: 1, y: 0 }}
              exit={{ scale: 0.4, opacity: 0, y: 18 }}
              transition={{ type: 'spring', stiffness: 380, damping: 28, mass: 0.7 }}
            >
              <DockItem
                title={win.title}
                icon={Icon}
                onRestore={() => restoreWindow(win.id)}
                onClose={() => closeWindow(win.id)}
              />
            </motion.div>
          );
        })}
      </AnimatePresence>
    </div>
  );
};

interface DockItemProps {
  title: string;
  icon: React.FC<{ size?: number }>;
  onRestore: () => void;
  onClose: () => void;
}

const DockItem: React.FC<DockItemProps> = ({ title, icon: Icon, onRestore, onClose }) => {
  const [hovered, setHovered] = React.useState(false);

  return (
    <div
      style={{
        position: 'relative',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 4,
      }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {/* Tooltip */}
      <AnimatePresence>
        {hovered && (
          <motion.div
            initial={{ opacity: 0, y: 6, scale: 0.92 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 4, scale: 0.95 }}
            transition={{ type: 'spring', stiffness: 480, damping: 30, mass: 0.5 }}
            style={{
              position: 'absolute',
              bottom: 'calc(100% + 8px)',
              left: '50%',
              x: '-50%',
              background: 'rgba(20,20,24,0.95)',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: 6,
              padding: '3px 8px',
              fontSize: 11,
              fontWeight: 500,
              color: 'rgba(255,255,255,0.85)',
              whiteSpace: 'nowrap',
              pointerEvents: 'none',
            }}
          >
            {title}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Close button on hover */}
      {hovered && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
          style={{
            position: 'absolute',
            top: -6,
            right: -6,
            width: 14,
            height: 14,
            borderRadius: '50%',
            background: '#ff5f57',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
            zIndex: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 9,
            color: 'rgba(0,0,0,0.6)',
            fontWeight: 700,
          }}
          aria-label="Close"
        >
          ×
        </button>
      )}

      {/* App icon */}
      <motion.div
        onClick={onRestore}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            onRestore();
          }
        }}
        role="button"
        tabIndex={0}
        aria-label={title}
        animate={{
          y: hovered ? -6 : 0,
          scale: hovered ? 1.14 : 1,
          boxShadow: hovered ? '0 8px 20px rgba(0,0,0,0.5)' : '0 0 0 rgba(0,0,0,0)',
        }}
        whileTap={{ scale: 0.92 }}
        transition={{ type: 'spring', stiffness: 420, damping: 26, mass: 0.6 }}
        style={{
          width: 44,
          height: 44,
          borderRadius: 12,
          background: 'rgba(255,255,255,0.06)',
          border: '1px solid rgba(255,255,255,0.1)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
        }}
      >
        <Icon size={28} />
      </motion.div>

      {/* Minimized dot */}
      <div
        style={{
          width: 4,
          height: 4,
          borderRadius: '50%',
          background: 'rgba(255,255,255,0.4)',
        }}
      />
    </div>
  );
};
