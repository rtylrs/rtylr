import {
  List as Menu,
  House as Home,
  ArrowClockwise as RotateCw,
  X,
  User,
  Warning,
} from 'phosphor-react';
import { useEffect, useState, useRef } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@voxire/ui';
import { useUserStore } from '../../stores/userStore';
import { useFullscreen } from '../../hooks/useFullscreen';

const getPlatform = () => {
  const ua = navigator.userAgent;
  if (/Windows/.test(ua)) return 'Windows';
  if (/Mac/.test(ua)) return 'Mac';
  if (/Linux/.test(ua)) return 'Linux';
  return 'Unknown';
};

const useFPS = (): number => {
  const [fps, setFps] = useState(0);
  const frameCount = useRef(0);
  const lastFrameTime = useRef(Date.now());
  const animationId = useRef<number | undefined>(undefined);

  useEffect(() => {
    const calculateFPS = () => {
      frameCount.current++;
      const now = Date.now();
      const deltaTime = now - lastFrameTime.current;
      if (deltaTime >= 2000) {
        setFps(frameCount.current);
        frameCount.current = 0;
        lastFrameTime.current = now;
      }
      animationId.current = requestAnimationFrame(calculateFPS);
    };
    animationId.current = requestAnimationFrame(calculateFPS);
    return () => {
      if (animationId.current) cancelAnimationFrame(animationId.current);
    };
  }, []);

  return fps;
};

const ToolBarContent: React.FC = () => {
  const { user } = useUserStore();
  const navigate = useNavigate();
  const [mode, setMode] = useState<string>(localStorage.getItem('mode') || 'prod');
  const [menuOpen, setMenuOpen] = useState<boolean>(false);
  const fps = useFPS();
  const { isFullscreen, toggle: toggleFullscreen } = useFullscreen();
  const [memoryUsage, setMemoryUsage] = useState('N/A');

  const isLocalhost =
    window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

  useEffect(() => {
    localStorage.setItem('mode', mode);
  }, [mode]);

  useEffect(() => {
    const updateMemoryUsage = () => {
      const perf = performance as typeof performance & {
        memory?: { usedJSHeapSize: number; totalJSHeapSize: number };
      };
      if (perf.memory) {
        const used = perf.memory.usedJSHeapSize / 1024 / 1024;
        const total = perf.memory.totalJSHeapSize / 1024 / 1024;
        setMemoryUsage(`${used.toFixed(1)}MB / ${total.toFixed(1)}MB`);
      }
    };
    const interval = setInterval(updateMemoryUsage, 5000);
    updateMemoryUsage();
    return () => clearInterval(interval);
  }, []);

  const toolbarNavigate = (path: string) => {
    setMenuOpen(false);
    navigate({ to: path });
  };

  const [confirmClearStorage, setConfirmClearStorage] = useState(false);
  const handleClearStorage = () => setConfirmClearStorage(true);
  const doClearStorage = () => {
    localStorage.clear();
    window.location.reload();
  };

  const itemCls =
    'w-8 h-8 flex items-center justify-center rounded-lg hover:bg-white/10 text-white/70 hover:text-white cursor-pointer transition-colors';

  return (
    <>
      <Dialog open={confirmClearStorage} onOpenChange={setConfirmClearStorage}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-[14px] font-semibold">
              <Warning size={14} weight="fill" style={{ color: '#ef4444' }} />
              Clear local storage
            </DialogTitle>
            <DialogDescription className="text-[12px]">
              This wipes every cached preference, token, and saved view in this browser and reloads.
              You will need to sign in again. This cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 mt-4">
            <button
              type="button"
              onClick={() => setConfirmClearStorage(false)}
              className="px-3 py-1.5 rounded-lg text-[12px] font-medium text-white/70"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={doClearStorage}
              className="px-3 py-1.5 rounded-lg text-[12px] font-semibold"
              style={{ background: '#ef4444', color: '#fff' }}
            >
              Clear and reload
            </button>
          </div>
        </DialogContent>
      </Dialog>
      <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50">
        <div className="flex items-center gap-1 bg-gray-900/95 backdrop-blur-sm border border-white/10 rounded-xl px-2 py-1.5 shadow-xl">
          <div className={itemCls} onClick={() => toolbarNavigate('/')} title="Home">
            <Home size={14} />
          </div>
          <div
            className={itemCls}
            onClick={() => toolbarNavigate('/app/business/')}
            title="Dashboard"
          >
            <User size={14} />
          </div>
          <div className={itemCls} onClick={() => window.location.reload()} title="Reload">
            <RotateCw size={14} />
          </div>

          <div className="w-px h-4 bg-white/15 mx-0.5" />

          <label className="flex items-center gap-1.5 cursor-pointer">
            <span className="text-[10px] text-white/50">DEV</span>
            <div className="relative">
              <input
                type="checkbox"
                className="sr-only peer"
                checked={mode === 'dev'}
                onChange={() => setMode((m) => (m === 'dev' ? 'prod' : 'dev'))}
              />
              <div className="w-7 h-3.5 bg-white/15 rounded-full peer peer-checked:bg-accent after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:w-2.5 after:h-2.5 after:rounded-full after:bg-white after:transition-all peer-checked:after:translate-x-3.5" />
            </div>
          </label>

          <div className="w-px h-4 bg-white/15 mx-0.5" />

          <span className="text-[10px] text-white/50 px-1">FPS: {fps}</span>
          <span className="text-[10px] text-white/50 px-1">RAM: {memoryUsage}</span>

          <div className={itemCls} onClick={() => setMenuOpen((v) => !v)}>
            {menuOpen ? <X size={14} /> : <Menu size={14} />}
          </div>

          {menuOpen && (
            <div className="absolute bottom-full left-0 mb-2 w-40 bg-gray-900 border border-white/10 rounded-xl overflow-hidden shadow-xl">
              <div className="p-1">
                <div
                  className="px-3 py-2 text-sm text-white/80 hover:bg-white/8 rounded-lg cursor-pointer"
                  onClick={() => toolbarNavigate('/')}
                >
                  Home
                </div>
                <div
                  className="px-3 py-2 text-sm text-white/80 hover:bg-white/8 rounded-lg cursor-pointer"
                  onClick={() => toolbarNavigate('/app/business/')}
                >
                  Statistics
                </div>
                <div
                  className="px-3 py-2 text-sm text-red-400 font-semibold hover:bg-white/8 rounded-lg cursor-pointer"
                  onClick={handleClearStorage}
                >
                  Clear Storage
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
};

const ToolBar: React.FC = () => {
  const [showToolbar, setShowToolbar] = useState<boolean>(
    localStorage.getItem('showToolbar') === 'true'
  );

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const isMac = getPlatform() === 'Mac';
      const isCtrlAltTPressed = e.ctrlKey && e.altKey && e.key.toLowerCase() === 't';
      const isCmdOptTPressed = e.metaKey && e.ctrlKey && e.key.toLowerCase() === 't';
      if (isCtrlAltTPressed || (isMac && isCmdOptTPressed)) {
        setShowToolbar((prev) => {
          const next = !prev;
          localStorage.setItem('showToolbar', next.toString());
          return next;
        });
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const isProduction =
    window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1';

  if (!showToolbar || isProduction) return null;

  return <ToolBarContent />;
};

export default ToolBar;
