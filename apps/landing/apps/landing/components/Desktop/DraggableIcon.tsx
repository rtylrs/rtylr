import React, { useRef, useState, useCallback } from 'react';
import { useDesktopStore, WindowType } from '../../stores/useDesktopStore';

export interface DesktopIconDef {
  key: string;
  label: string;
  svgIcon: React.FC<{ size?: number }>;
  windowType?: WindowType;
  windowUrl?: string;
  href?: string;
  defaultPos: { x: number; y: number };
}

interface DraggableIconProps {
  def: DesktopIconDef;
}

export const DraggableIcon: React.FC<DraggableIconProps> = ({ def }) => {
  const addWindow = useDesktopStore((s) => s.addWindow);
  const hasDragged = useRef(false);
  const [active, setActive] = useState(false);
  const SvgIcon = def.svgIcon;

  const handleClick = useCallback(() => {
    if (hasDragged.current) return;
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
      onMouseDown={() => setActive(true)}
      onMouseUp={() => setActive(false)}
      onMouseLeave={() => setActive(false)}
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
        borderRadius: 12,
        background: 'rgba(0,0,0,0.18)',
        backdropFilter: 'blur(6px)',
        WebkitBackdropFilter: 'blur(6px)',
      }}
    >
      <div
        style={{
          transform: active ? 'scale(0.92)' : 'scale(1)',
          transition: 'transform 0.1s ease',
          filter: 'drop-shadow(0 3px 8px rgba(0,0,0,0.4))',
        }}
      >
        <SvgIcon size={52} />
      </div>

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
