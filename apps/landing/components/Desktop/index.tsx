import React, { useEffect } from 'react';
import dynamic from 'next/dynamic';
import { useDesktopStore, Wallpaper } from '../../stores/useDesktopStore';
import { MenuBar } from '../MenuBar';
import { DesktopIcons } from './DesktopIcons';
import { WallpaperOverlay } from './WallpaperOverlay';
import { BottomDock } from './BottomDock';
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

export const Desktop: React.FC = () => {
  const windows = useDesktopStore((s) => s.windows);
  const wallpaper = useDesktopStore((s) => s.wallpaper);
  const animationsEnabled = useDesktopStore((s) => s.animationsEnabled);
  const addWindow = useDesktopStore((s) => s.addWindow);

  useKeyboardShortcuts();

  useEffect(() => {
    if (windows.length === 0) addWindow('landing');
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: WALLPAPERS[wallpaper],
        overflow: 'hidden',
        fontFamily: "'Inter', system-ui, sans-serif",
      }}
    >
      <WallpaperOverlay wallpaper={wallpaper} />
      <MenuBar />

      {/* Desktop area — below menu bar */}
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
        {windows.map((win) => (
          <AppWindow key={win.id} window={win} animationsEnabled={animationsEnabled} />
        ))}

        <BottomDock />
      </div>
    </div>
  );
};
