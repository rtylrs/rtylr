import React, { useState } from 'react';
import { AppIcon } from '../AppIcon';
import { WindowType } from '../../stores/useDesktopStore';

interface TitleBarProps {
  type: WindowType;
  title: string;
  onClose: () => void;
  onMinimize: () => void;
  onMaximize?: () => void;
}

const CloseIcon = () => (
  <svg width="6" height="6" viewBox="0 0 6 6" fill="none">
    <path d="M1 1l4 4M5 1L1 5" stroke="rgba(0,0,0,0.55)" strokeWidth="1.2" strokeLinecap="round" />
  </svg>
);

const MinimizeIcon = () => (
  <svg width="6" height="6" viewBox="0 0 6 6" fill="none">
    <path d="M1 3h4" stroke="rgba(0,0,0,0.55)" strokeWidth="1.2" strokeLinecap="round" />
  </svg>
);

const MaximizeIcon = () => (
  <svg width="6" height="6" viewBox="0 0 6 6" fill="none">
    <path
      d="M1 5L5 1M3 1h2v2"
      stroke="rgba(0,0,0,0.55)"
      strokeWidth="1.1"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

export const TitleBar: React.FC<TitleBarProps> = ({
  type,
  title,
  onClose,
  onMinimize,
  onMaximize,
}) => {
  const [groupHovered, setGroupHovered] = useState(false);

  return (
    <div
      className="window-titlebar"
      style={{
        display: 'flex',
        alignItems: 'center',
        height: 38,
        padding: '0 12px',
        background: 'var(--titlebar-bg, rgba(30,30,32,0.98))',
        borderBottom: '1px solid var(--titlebar-border, rgba(255,255,255,0.08))',
        borderRadius: '10px 10px 0 0',
        userSelect: 'none',
        cursor: 'grab',
        gap: 8,
        flexShrink: 0,
      }}
    >
      {/* Traffic lights */}
      <div
        style={{ display: 'flex', gap: 6, marginRight: 4 }}
        onMouseEnter={() => setGroupHovered(true)}
        onMouseLeave={() => setGroupHovered(false)}
        onMouseDown={(e) => e.stopPropagation()}
      >
        <button
          className="rtylr-os-window-control rtylr-os-window-control-close"
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            background: '#ff5f57',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
          aria-label="Close"
        >
          {groupHovered && <CloseIcon />}
        </button>
        <button
          className="rtylr-os-window-control rtylr-os-window-control-minimize"
          onClick={(e) => {
            e.stopPropagation();
            onMinimize();
          }}
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            background: '#febc2e',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
          aria-label="Minimize"
        >
          {groupHovered && <MinimizeIcon />}
        </button>
        <button
          className="rtylr-os-window-control rtylr-os-window-control-maximize"
          onClick={(e) => {
            e.stopPropagation();
            onMaximize?.();
          }}
          style={{
            width: 12,
            height: 12,
            borderRadius: '50%',
            background: '#28c840',
            border: 'none',
            cursor: 'pointer',
            padding: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
          aria-label="Maximize"
        >
          {groupHovered && <MaximizeIcon />}
        </button>
      </div>

      <AppIcon
        type={type}
        size={13}
        color="var(--titlebar-icon, rgba(255,255,255,0.55))"
        strokeWidth={1.75}
      />
      <span
        style={{
          fontSize: 12,
          fontWeight: 500,
          color: 'var(--titlebar-text, rgba(255,255,255,0.7))',
          flex: 1,
        }}
      >
        {title}
      </span>
    </div>
  );
};
