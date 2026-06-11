import React, { useCallback } from 'react';
import { Rnd } from 'react-rnd';
import { motion } from 'motion/react';
import { TitleBar } from './TitleBar';
import { WindowContent } from './WindowContent';
import { useDesktopStore, WindowState } from '../../stores/useDesktopStore';

interface AppWindowProps {
  window: WindowState;
  animationsEnabled: boolean;
}

export const AppWindow: React.FC<AppWindowProps> = React.memo(
  ({ window: win, animationsEnabled }) => {
    const closeWindow = useDesktopStore((s) => s.closeWindow);
    const bringToFront = useDesktopStore((s) => s.bringToFront);
    const minimizeWindow = useDesktopStore((s) => s.minimizeWindow);
    const updateWindowPosition = useDesktopStore((s) => s.updateWindowPosition);
    const updateWindowSize = useDesktopStore((s) => s.updateWindowSize);
    const [dragging, setDragging] = React.useState(false);

    const handleClose = useCallback(() => closeWindow(win.id), [win.id, closeWindow]);
    const handleMinimize = useCallback(() => minimizeWindow(win.id), [win.id, minimizeWindow]);
    const handleFocus = useCallback(() => bringToFront(win.id), [win.id, bringToFront]);

    return (
      <motion.div
        initial={animationsEnabled ? { scale: 0.85, opacity: 0 } : false}
        animate={{ scale: 1, opacity: 1 }}
        exit={animationsEnabled ? { scale: 0.9, opacity: 0 } : undefined}
        transition={
          animationsEnabled ? { type: 'spring', stiffness: 400, damping: 28 } : { duration: 0 }
        }
        // z-index lives here - this full-desktop overlay establishes the stacking context
        // so desktop icons (z=50) are always below all open windows (z=601+)
        style={{ position: 'absolute', inset: 0, pointerEvents: 'none', zIndex: win.zIndex }}
      >
        <Rnd
          size={{ width: win.width, height: win.height }}
          position={{ x: win.x, y: win.y }}
          style={{ contain: 'layout style', pointerEvents: 'auto' }}
          dragHandleClassName="window-titlebar"
          bounds="parent"
          onDragStart={() => {
            handleFocus();
            setDragging(true);
          }}
          onDragStop={(_e, d) => {
            setDragging(false);
            updateWindowPosition(win.id, d.x, d.y);
          }}
          onResizeStart={() => setDragging(true)}
          onResizeStop={(_e, _dir, ref, _delta, pos) => {
            setDragging(false);
            updateWindowSize(win.id, parseInt(ref.style.width), parseInt(ref.style.height));
            updateWindowPosition(win.id, pos.x, pos.y);
          }}
          minWidth={320}
          minHeight={240}
          enableResizing={{
            top: true,
            right: true,
            bottom: true,
            left: true,
            topRight: true,
            bottomRight: true,
            bottomLeft: true,
            topLeft: true,
          }}
        >
          <div
            onClick={handleFocus}
            style={{
              width: '100%',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              background: 'var(--win-bg, rgba(22,22,24,0.97))',
              border: '1px solid var(--win-border, rgba(255,255,255,0.1))',
              borderRadius: 10,
              overflow: 'hidden',
              boxShadow:
                'var(--win-shadow, 0 24px 80px rgba(0,0,0,0.6), 0 4px 12px rgba(0,0,0,0.4))',
            }}
          >
            <TitleBar
              type={win.type}
              title={win.title}
              onClose={handleClose}
              onMinimize={handleMinimize}
            />
            <div style={{ flex: 1, position: 'relative', overflow: 'hidden' }}>
              <WindowContent type={win.type} url={win.url} />
              {/* Blocks iframe from stealing mouse events during drag/resize */}
              {dragging && <div style={{ position: 'absolute', inset: 0, zIndex: 10 }} />}
            </div>
          </div>
        </Rnd>
      </motion.div>
    );
  }
);

AppWindow.displayName = 'AppWindow';
