import React, { useCallback } from 'react';
import { motion } from 'motion/react';
import { useDesktopStore, WindowType } from '../../stores/useDesktopStore';

export interface DesktopIconDef {
  key: string;
  label: string;
  svgIcon: React.FC<{ size?: number }>;
  windowType?: WindowType;
  windowUrl?: string;
  href?: string;
}

interface DesktopIconProps {
  def: DesktopIconDef;
}

export const DesktopIcon: React.FC<DesktopIconProps> = ({ def }) => {
  const addWindow = useDesktopStore((s) => s.addWindow);
  const SvgIcon = def.svgIcon;

  const handleClick = useCallback(() => {
    if (def.windowType) addWindow(def.windowType, def.windowUrl);
    else if (def.href) {
      if (def.href.startsWith('/')) window.location.href = def.href;
      else window.open(def.href, '_blank', 'noopener,noreferrer');
    }
  }, [def, addWindow]);

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={handleClick}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleClick();
        }
      }}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 5,
        cursor: 'pointer',
        userSelect: 'none',
        width: 72,
        flexShrink: 0,
        pointerEvents: 'auto',
        padding: '6px 4px',
      }}
    >
      <motion.div
        whileHover={{ scale: 1.12, y: -3 }}
        whileTap={{ scale: 0.88 }}
        transition={{ type: 'spring', stiffness: 420, damping: 26, mass: 0.6 }}
        style={{
          filter:
            'drop-shadow(0 4px 10px rgba(0,0,0,0.45)) drop-shadow(0 1px 2px rgba(0,0,0,0.3))',
          willChange: 'transform',
        }}
      >
        <SvgIcon size={52} />
      </motion.div>

      <span
        style={{
          fontSize: 11,
          fontWeight: 500,
          color: 'rgba(255,255,255,0.9)',
          textAlign: 'center',
          lineHeight: 1.3,
          textShadow: '0 1px 5px rgba(0,0,0,0.9)',
          maxWidth: 68,
          overflow: 'hidden',
          whiteSpace: 'nowrap',
          textOverflow: 'ellipsis',
        }}
      >
        {def.label}
      </span>
    </div>
  );
};
