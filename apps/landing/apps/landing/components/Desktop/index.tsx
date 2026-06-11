import React, { useEffect, useState, useCallback } from 'react';
import dynamic from 'next/dynamic';
import { AnimatePresence } from 'motion/react';
import { useDesktopStore, Wallpaper, ColorMode } from '../../stores/useDesktopStore';
import { MenuBar } from '../MenuBar';
import { DesktopIcons } from './DesktopIcons';
import { WallpaperOverlay } from './WallpaperOverlay';
import { BottomDock } from './BottomDock';
import { Spotlight } from './Spotlight';
import { DesktopContextMenu } from './DesktopContextMenu';
import { useKeyboardShortcuts } from '../../hooks/desktop/useKeyboardShortcuts';

const AppWindow = dynamic(() => import('../AppWindow').then((m) => ({ default: m.AppWindow })), {
  ssr: false,
});

const WALLPAPERS: Record<Wallpaper, string> = {
  slate: 'linear-gradient(145deg, #0e0e1e 0%, #101428 100%)',
  midnight: 'linear-gradient(145deg, #0a0a0f 0%, #111116 40%, #0d0d14 100%)',
  aurora: 'linear-gradient(145deg, #0a1628 0%, #0d1a0d 50%, #0a1628 100%)',
  dusk: 'linear-gradient(145deg, #120820 0%, #1e1030 50%, #120820 100%)',
  forest: 'linear-gradient(145deg, #091409 0%, #112011 100%)',
};

// Only the window chrome (frame + titlebar) flips between light/dark.
// Window content always stays dark - apps maintain their own appearance.
const DARK_VARS = {
  '--win-bg': 'rgba(22,22,24,0.97)',
  '--win-border': 'rgba(255,255,255,0.1)',
  '--win-shadow': '0 24px 80px rgba(0,0,0,0.6), 0 4px 12px rgba(0,0,0,0.4)',
  '--titlebar-bg': 'rgba(30,30,32,0.98)',
  '--titlebar-border': 'rgba(255,255,255,0.08)',
  '--titlebar-text': 'rgba(255,255,255,0.7)',
  '--titlebar-icon': 'rgba(255,255,255,0.55)',
};

const LIGHT_VARS = {
  '--win-bg': 'rgba(242,242,244,0.97)',
  '--win-border': 'rgba(0,0,0,0.12)',
  '--win-shadow': '0 24px 80px rgba(0,0,0,0.18), 0 4px 12px rgba(0,0,0,0.1)',
  '--titlebar-bg': 'rgba(225,225,228,0.98)',
  '--titlebar-border': 'rgba(0,0,0,0.08)',
  '--titlebar-text': 'rgba(0,0,0,0.7)',
  '--titlebar-icon': 'rgba(0,0,0,0.55)',
};

function resolveMode(colorMode: ColorMode, systemDark: boolean): 'light' | 'dark' {
  if (colorMode === 'dark') return 'dark';
  if (colorMode === 'light') return 'light';
  return systemDark ? 'dark' : 'light';
}

export const Desktop: React.FC = () => {
  const windows = useDesktopStore((s) => s.windows);
  const wallpaper = useDesktopStore((s) => s.wallpaper);
  const colorMode = useDesktopStore((s) => s.colorMode);
  const animationsEnabled = useDesktopStore((s) => s.animationsEnabled);
  const addWindow = useDesktopStore((s) => s.addWindow);

  const [systemDark, setSystemDark] = useState(false);
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null);

  useKeyboardShortcuts();

  const handleContextMenu = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setContextMenu({ x: e.clientX, y: e.clientY });
  }, []);

  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    setSystemDark(mq.matches);
    const handler = (e: MediaQueryListEvent) => setSystemDark(e.matches);
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, []);

  useEffect(() => {
    if (windows.length === 0) addWindow('landing');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const theme = resolveMode(colorMode, systemDark);
  const vars = theme === 'dark' ? DARK_VARS : LIGHT_VARS;

  return (
    <div
      className="rtylr-desktop-shell"
      onContextMenu={handleContextMenu}
      onClick={() => setContextMenu(null)}
      style={
        {
          position: 'fixed',
          inset: 0,
          background: WALLPAPERS[wallpaper],
          overflow: 'hidden',
          fontFamily: "'Inter', system-ui, sans-serif",
          ...vars,
        } as React.CSSProperties
      }
    >
      <WallpaperOverlay wallpaper={wallpaper} animated={animationsEnabled} />
      <MenuBar />

      {/* Desktop area - below menu bar */}
      <div
        style={{
          position: 'absolute',
          top: 36,
          left: 0,
          right: 0,
          bottom: 0,
          overflow: 'hidden',
        }}
      >
        {/* Free-floating draggable desktop icons */}
        <DesktopIcons />

        {/* Open windows */}
        <AnimatePresence>
          {windows
            .filter((win) => !win.minimized)
            .map((win) => (
              <AppWindow key={win.id} window={win} animationsEnabled={animationsEnabled} />
            ))}
        </AnimatePresence>

        <BottomDock />
      </div>

      <Spotlight />
      {contextMenu && (
        <DesktopContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          onClose={() => setContextMenu(null)}
        />
      )}
    </div>
  );
};
