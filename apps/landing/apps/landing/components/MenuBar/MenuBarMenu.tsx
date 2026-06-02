import React, { useState, useRef, useEffect } from 'react';

export interface MenuItem {
  label: string;
  shortcut?: string;
  onClick?: () => void;
  separator?: boolean;
}

interface MenuBarMenuProps {
  label: string;
  items: MenuItem[];
}

export const MenuBarMenu: React.FC<MenuBarMenuProps> = ({ label, items }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    if (open) document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  return (
    <div ref={ref} style={{ position: 'relative' }}>
      <button
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        style={{
          background: open ? 'rgba(255,255,255,0.1)' : 'transparent',
          border: 'none',
          color: 'rgba(255,255,255,0.8)',
          fontSize: 12,
          fontWeight: 500,
          padding: '3px 8px',
          borderRadius: 4,
          cursor: 'pointer',
          transition: 'background 0.1s',
        }}
        onMouseEnter={(e) => {
          if (!open)
            (e.currentTarget as HTMLButtonElement).style.background = 'rgba(255,255,255,0.07)';
        }}
        onMouseLeave={(e) => {
          if (!open) (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
        }}
      >
        {label}
      </button>

      {open && (
        <div
          style={{
            position: 'absolute',
            top: 'calc(100% + 4px)',
            left: 0,
            minWidth: 200,
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
            item.separator ? (
              <div
                key={i}
                style={{ height: 1, background: 'rgba(255,255,255,0.08)', margin: '4px 0' }}
              />
            ) : (
              <button
                key={i}
                onClick={() => {
                  item.onClick?.();
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
                <span>{item.label}</span>
                {item.shortcut && (
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
