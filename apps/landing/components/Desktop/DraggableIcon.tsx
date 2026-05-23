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
  const iconPositions = useDesktopStore((s) => s.iconPositions);
  const setIconPosition = useDesktopStore((s) => s.setIconPosition);

  const pos = iconPositions[def.key] ?? def.defaultPos;

  const dragging = useRef(false);
  const hasDragged = useRef(false);
  const startMouse = useRef({ x: 0, y: 0 });
  const startPos = useRef({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [currentPos, setCurrentPos] = useState(pos);
  const SvgIcon = def.svgIcon;

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      dragging.current = true;
      hasDragged.current = false;
      startMouse.current = { x: e.clientX, y: e.clientY };
      startPos.current = currentPos;

      const onMove = (ev: MouseEvent) => {
        if (!dragging.current) return;
        const dx = ev.clientX - startMouse.current.x;
        const dy = ev.clientY - startMouse.current.y;
        if (Math.abs(dx) > 4 || Math.abs(dy) > 4) {
          hasDragged.current = true;
          setIsDragging(true);
        }
        if (hasDragged.current) {
          setCurrentPos({ x: startPos.current.x + dx, y: startPos.current.y + dy });
        }
      };

      const onUp = () => {
        dragging.current = false;
        setIsDragging(false);
        if (hasDragged.current) {
          setIconPosition(def.key, currentPos);
        }
        window.removeEventListener('mousemove', onMove);
        window.removeEventListener('mouseup', onUp);
      };

      window.addEventListener('mousemove', onMove);
      window.addEventListener('mouseup', onUp);
    },
    [currentPos, def.key, setIconPosition]
  );

  const handleClick = useCallback(() => {
    if (hasDragged.current) return;
    if (def.windowType) addWindow(def.windowType, def.windowUrl);
    else if (def.href) window.open(def.href, '_blank', 'noopener,noreferrer');
  }, [def, addWindow]);

  return (
    <div
      onMouseDown={handleMouseDown}
      onClick={handleClick}
      style={{
        position: 'absolute',
        left: currentPos.x,
        top: currentPos.y,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: 5,
        cursor: isDragging ? 'grabbing' : 'pointer',
        userSelect: 'none',
        width: 64,
        zIndex: isDragging ? 599 : 50,
      }}
    >
      <div
        style={{
          transform: isDragging ? 'scale(1.1) translateY(-4px)' : 'scale(1)',
          transition: isDragging ? 'none' : 'transform 0.12s ease',
          filter: isDragging
            ? 'drop-shadow(0 12px 20px rgba(0,0,0,0.6))'
            : 'drop-shadow(0 3px 8px rgba(0,0,0,0.4))',
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
          maxWidth: 64,
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
