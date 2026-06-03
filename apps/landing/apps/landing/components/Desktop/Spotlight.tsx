import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useDesktopStore, WindowType } from '../../stores/useDesktopStore';

interface SpotlightItem {
  label: string;
  type: string;
  kind: string;
  url?: string;
}

const ITEMS: SpotlightItem[] = [
  { label: 'POS', type: 'pos', kind: 'App' },
  { label: 'ERP', type: 'erp', kind: 'App' },
  { label: 'CRM', type: 'crm', kind: 'App' },
  { label: 'HR', type: 'hr', kind: 'App' },
  { label: 'Finance', type: 'finance', kind: 'App' },
  { label: 'Flow', type: 'flow', kind: 'App' },
  { label: 'Pricing', type: 'pricing', kind: 'Page', url: '/pricing?embed=1' },
  { label: 'Docs', type: 'docs', kind: 'Page', url: 'https://docs.rtylr.com' },
  { label: 'Blog', type: 'docs', kind: 'Page', url: 'https://blog.rtylr.com' },
  { label: 'Status', type: 'status', kind: 'Page', url: 'https://status.rtylr.com' },
  { label: 'Settings', type: 'settings', kind: 'System' },
  { label: 'About rtylr', type: 'about', kind: 'System' },
  { label: 'Terminal', type: 'terminal', kind: 'App' },
  { label: 'Snake', type: 'snake', kind: 'Game' },
  { label: 'Doom', type: 'doom', kind: 'Game' },
];

const KIND_COLOR: Record<string, string> = {
  App: 'rgba(99,179,237,0.8)',
  Page: 'rgba(154,205,150,0.8)',
  System: 'rgba(200,200,200,0.7)',
  Game: 'rgba(239,148,68,0.8)',
};

export const Spotlight: React.FC = () => {
  const spotlightOpen = useDesktopStore((s) => s.spotlightOpen);
  const setSpotlightOpen = useDesktopStore((s) => s.setSpotlightOpen);
  const addWindow = useDesktopStore((s) => s.addWindow);

  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const filtered = query.trim()
    ? ITEMS.filter((item) => item.label.toLowerCase().includes(query.toLowerCase()))
    : ITEMS;

  useEffect(() => {
    if (spotlightOpen) {
      setQuery('');
      setSelected(0);
      setTimeout(() => inputRef.current?.focus(), 30);
    }
  }, [spotlightOpen]);

  useEffect(() => {
    setSelected(0);
  }, [query]);

  const open = useCallback(
    (item: SpotlightItem) => {
      addWindow(item.type as WindowType, item.url);
      setSpotlightOpen(false);
    },
    [addWindow, setSpotlightOpen]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Escape') {
        setSpotlightOpen(false);
        return;
      }
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelected((s) => Math.min(s + 1, filtered.length - 1));
        return;
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelected((s) => Math.max(s - 1, 0));
        return;
      }
      if (e.key === 'Enter') {
        e.preventDefault();
        if (filtered[selected]) open(filtered[selected]);
      }
    },
    [filtered, selected, open, setSpotlightOpen]
  );

  if (!spotlightOpen) return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 9000,
        background: 'rgba(0,0,0,0.55)',
        backdropFilter: 'blur(4px)',
        display: 'flex',
        justifyContent: 'center',
      }}
      onClick={() => setSpotlightOpen(false)}
    >
      <div
        style={{
          position: 'absolute',
          top: '20%',
          width: '100%',
          maxWidth: 560,
          margin: '0 auto',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <div
          style={{
            background: 'rgba(18,18,22,0.97)',
            border: '1px solid rgba(255,255,255,0.15)',
            borderRadius: filtered.length > 0 && query ? '12px 12px 0 0' : 12,
            padding: '12px 16px',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
          }}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none">
            <circle cx="11" cy="11" r="7" stroke="rgba(255,255,255,0.4)" strokeWidth="2" />
            <path
              d="M16.5 16.5L21 21"
              stroke="rgba(255,255,255,0.4)"
              strokeWidth="2"
              strokeLinecap="round"
            />
          </svg>
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Search apps, pages, games…"
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: '#ffffff',
              fontSize: 16,
              fontFamily: "'Inter', system-ui, sans-serif",
            }}
            autoComplete="off"
            spellCheck={false}
          />
          <kbd
            style={{
              fontSize: 10,
              color: 'rgba(255,255,255,0.3)',
              background: 'rgba(255,255,255,0.06)',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: 4,
              padding: '2px 6px',
            }}
          >
            esc
          </kbd>
        </div>

        {filtered.length > 0 && (
          <div
            style={{
              background: 'rgba(14,14,18,0.98)',
              border: '1px solid rgba(255,255,255,0.1)',
              borderTop: 'none',
              borderRadius: '0 0 12px 12px',
              overflow: 'hidden',
              maxHeight: 360,
              overflowY: 'auto',
            }}
          >
            {filtered.map((item, i) => (
              <div
                key={`${item.type}-${item.url ?? ''}-${i}`}
                role="option"
                tabIndex={0}
                aria-selected={i === selected}
                onClick={() => open(item)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    open(item);
                  }
                }}
                onMouseEnter={() => setSelected(i)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '9px 16px',
                  cursor: 'pointer',
                  background: i === selected ? 'rgba(255,255,255,0.07)' : 'transparent',
                  transition: 'background 0.1s',
                }}
              >
                <span style={{ fontSize: 14, color: 'rgba(255,255,255,0.9)' }}>{item.label}</span>
                <span
                  style={{
                    fontSize: 11,
                    color: KIND_COLOR[item.kind] ?? 'rgba(255,255,255,0.4)',
                    background: 'rgba(255,255,255,0.05)',
                    borderRadius: 4,
                    padding: '1px 7px',
                  }}
                >
                  {item.kind}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
