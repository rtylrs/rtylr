import { Switch } from '@voxire/components';
import { Menu, Home, RotateCw, X, User, Maximize, Minimize } from "lucide-react";
import { useEffect, useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { useUserStore } from "../../stores/userStore";
import { useFullscreen } from "../../hooks/useFullscreen";
import { usePlatform } from "../../hooks/usePlatform";
import "./ToolBar.scss";


const useFPS = (): number => {
  const [fps, setFps] = useState(0);
  const frameCount = useRef(0);
  const lastFrameTime = useRef(Date.now());
  const animationId = useRef<number>();

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
      if (animationId.current) {
        cancelAnimationFrame(animationId.current);
      }
    };
  }, []);

  return fps;
};

const getPlatform = () => {
  const userAgent = navigator.userAgent;
  if (/Windows/.test(userAgent)) {
    return "Windows";
  } else if (/Mac/.test(userAgent)) {
    return "Mac";
  } else if (/Linux/.test(userAgent)) {
    return "Linux";
  } else {
    return "Unknown";
  }
};

const ToolBarContent: React.FC = () => {
  const { user } = useUserStore();
  const navigate = useNavigate();
  const [mode, setMode] = useState<string>(
    localStorage.getItem("mode") || "prod"
  );
  const [menuOpen, setMenuOpen] = useState<boolean>(false);
  const fps = useFPS();
  const { isFullscreen, toggle: toggleFullscreen } = useFullscreen();
  const { isTauri } = usePlatform();

  const isLocalhost =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";

  const handleMenuToggle = () => {
    setMenuOpen((prev) => !prev);
  };

  useEffect(() => {
    localStorage.setItem("mode", mode);
  }, [mode]);

  const handleModeChange = () => {
    setMode((prevMode) => (prevMode === "dev" ? "prod" : "dev"));
  };

  const [memoryUsage, setMemoryUsage] = useState("N/A");

  useEffect(() => {
    const updateMemoryUsage = () => {
      if ((performance as any).memory) {
        const used = (performance as any).memory.usedJSHeapSize / 1024 / 1024;
        const total = (performance as any).memory.totalJSHeapSize / 1024 / 1024;
        setMemoryUsage(`${used.toFixed(1)}MB / ${total.toFixed(1)}MB`);
      }
    };

    const interval = setInterval(updateMemoryUsage, 5000);
    updateMemoryUsage();

    return () => clearInterval(interval);
  }, []);

  const toolbarNavigate = (path: string) => {
    setMenuOpen(false);
    navigate(path);
  };

  const handleClearStorage = () => {
    if (
      window.confirm(
        "Are you sure you want to clear the local storage? This action cannot be undone."
      )
    ) {
      localStorage.clear();
      window.location.reload();
    }
  };

  return (
    <div className="toolbar-container">
      <div className="toolbar">
        <div
          className="toolbar-item"
          onClick={() => {
            toolbarNavigate("/");
          }}
          title="Home"
        >
          <Home size={16} />
        </div>
        <div
          className="toolbar-item"
          onClick={() => {
            toolbarNavigate("/app/business/");
          }}
        >
          <User size={16} />
        </div>
        <div
          className="toolbar-item"
          onClick={() => {
            window.location.reload();
          }}
        >
          <RotateCw size={16} />
        </div>
        {isTauri && (
          <>
            <div className="vertical-line"></div>
            <div
              className="toolbar-item"
              onClick={toggleFullscreen}
              title={isFullscreen ? "Exit Fullscreen" : "Enter Fullscreen"}
            >
              {isFullscreen ? <Minimize size={16} /> : <Maximize size={16} />}
            </div>
          </>
        )}
        <div className="vertical-line"></div>
        <div className="toolbar-item">
          <Switch checked={mode === "dev"} onChange={handleModeChange} />
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item">
          <span>FPS: {fps}</span>
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item">
          <span>RAM: {memoryUsage}</span>
        </div>

        <div className="toolbar-item" onClick={handleMenuToggle}>
          {menuOpen ? <X size={16} /> : <Menu size={16} />}
        </div>

        {menuOpen && (
          <div className="dropdown-menu">
            <div className="dropdown-item" onClick={() => toolbarNavigate("/")}>
              Home
            </div>
            <div
              className="dropdown-item"
              onClick={() => toolbarNavigate("/app/business/")}
            >
              Statistics
            </div>
            <div
              className="dropdown-item"
              onClick={handleClearStorage}
              style={{
                color: "#ff0000",
                fontWeight: "bold",
              }}
            >
              Clear Storage
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

const ToolBar: React.FC = () => {
  const [showToolbar, setShowToolbar] = useState<boolean>(
    localStorage.getItem("showToolbar") === "true"
  );

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const os = getPlatform();
      const isMac = os === "Mac";

      const isCtrlAltTPressed =
        e.ctrlKey && e.altKey && e.key.toLowerCase() === "t";
      const isCmdOptTPressed =
        e.metaKey && e.ctrlKey && e.key.toLowerCase() === "t";

      if (isCtrlAltTPressed || (isMac && isCmdOptTPressed)) {
        setShowToolbar((prev) => {
          const newShowToolbar = !prev;
          localStorage.setItem("showToolbar", newShowToolbar.toString());
          return newShowToolbar;
        });
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const isProduction =
    window.location.hostname !== "localhost" &&
    window.location.hostname !== "127.0.0.1";

  if (!showToolbar || isProduction) {
    return null;
  }

  return <ToolBarContent />;
};

export default ToolBar;
