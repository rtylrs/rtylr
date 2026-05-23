import React, { useState, useRef, useEffect } from 'react';
import Image from 'next/image';
import { useDesktopStore } from '../../stores/useDesktopStore';

export const AppMenu: React.FC = () => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const closeAllWindows = useDesktopStore((s) => s.closeAllWindows);
  const setExperience = useDesktopStore((s) => s.setExperience);
  const addWindow = useDesktopStore((s) => s.addWindow);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    if (open) document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  const items = [
    { label: 'About rtylr', action: () => addWindow('about') },
    { separator: true },
    { label: 'Display Options', shortcut: null, action: () => addWindow('settings') },
    { separator: true },
    { label: 'Switch to Website Mode', shortcut: '⇧M', action: () => setExperience('website') },
    { label: 'Close All Windows', shortcut: '⇧X', action: () => closeAllWindows() },
  ] as const;

  return (
    <div ref={ref} style={{ position: 'relative' }}>
      <button
        onClick={() => setOpen((v) => !v)}
        style={{
          background: open ? 'rgba(255,255,255,0.1)' : 'transparent',
          border: 'none',
          padding: '2px 8px',
          borderRadius: 4,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          transition: 'background 0.1s',
        }}
      >
        <Image
          src="/logo.png"
          alt="rtylr"
          width={52}
          height={18}
          style={{ objectFit: 'contain' }}
        />
      </button>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: 'calc(100% + 4px)',
            left: 0,
            minWidth: 220,
            background: '#151515',
            border: '1px solid rgba(255,255,255,0.12)',
            borderRadius: 8,
            boxShadow: '0 16px 48px rgba(0,0,0,0.6)',
            padding: '4px 0',
            zIndex: 9999,
            backdropFilter: 'blur(20px)',
          }}
        >
          {items.map((item, i) =>
            'separator' in item && item.separator ? (
              <div
                key={i}
                style={{ height: 1, background: 'rgba(255,255,255,0.08)', margin: '4px 0' }}
              />
            ) : (
              <button
                key={i}
                onClick={() => {
                  'action' in item && item.action();
                  setOpen(false);
                }}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  width: '100%',
                  padding: '6px 14px',
                  background: 'transparent',
                  border: 'none',
                  color: 'rgba(255,255,255,0.85)',
                  fontSize: 12,
                  cursor: 'pointer',
                  textAlign: 'left',
                  transition: 'background 0.1s',
                }}
                onMouseEnter={(e) =>
                  ((e.currentTarget as HTMLButtonElement).style.background =
                    'rgba(255,255,255,0.1)')
                }
                onMouseLeave={(e) =>
                  ((e.currentTarget as HTMLButtonElement).style.background = 'transparent')
                }
              >
                <span>{'label' in item ? item.label : ''}</span>
                {'shortcut' in item && item.shortcut && (
                  <span style={{ fontSize: 10, color: 'rgba(255,255,255,0.35)', marginLeft: 16 }}>
                    {item.shortcut}
                  </span>
                )}
              </button>
            )
          )}
        </div>
      )}
    </div>
  );
};
