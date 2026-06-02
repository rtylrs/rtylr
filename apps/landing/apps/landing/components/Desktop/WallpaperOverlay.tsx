import React from 'react';
import { Wallpaper } from '../../stores/useDesktopStore';

const Midnight: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <radialGradient id="m-glow" cx="72%" cy="30%" r="45%">
        <stop offset="0%" stopColor="#2a2060" stopOpacity="0.35" />
        <stop offset="100%" stopColor="#000" stopOpacity="0" />
      </radialGradient>
      <radialGradient id="m-glow2" cx="20%" cy="80%" r="35%">
        <stop offset="0%" stopColor="#101840" stopOpacity="0.4" />
        <stop offset="100%" stopColor="#000" stopOpacity="0" />
      </radialGradient>
    </defs>
    {/* Ambient glows */}
    <rect width="1440" height="900" fill="url(#m-glow)" />
    <rect width="1440" height="900" fill="url(#m-glow2)" />
    {/* Star field */}
    {[
      [120, 80, 0.9],
      [340, 45, 0.6],
      [560, 120, 0.8],
      [780, 30, 0.5],
      [1020, 90, 0.9],
      [1200, 55, 0.7],
      [1380, 140, 0.6],
      [80, 240, 0.5],
      [260, 300, 0.8],
      [480, 200, 0.6],
      [700, 260, 0.9],
      [920, 180, 0.5],
      [1100, 220, 0.7],
      [1320, 280, 0.6],
      [200, 400, 0.5],
      [440, 370, 0.8],
      [660, 430, 0.6],
      [880, 350, 0.9],
      [1060, 410, 0.5],
      [1260, 360, 0.7],
      [150, 550, 0.6],
      [390, 510, 0.5],
      [610, 580, 0.8],
      [830, 520, 0.6],
      [1010, 560, 0.9],
      [1180, 490, 0.5],
      [1400, 540, 0.7],
      [300, 680, 0.6],
      [520, 720, 0.5],
      [740, 660, 0.8],
      [960, 700, 0.6],
      [1140, 640, 0.9],
      [1360, 720, 0.5],
      [50, 820, 0.7],
      [450, 850, 0.6],
      [850, 800, 0.5],
      [1250, 830, 0.8],
      [1440, 60, 0.6],
      [680, 10, 0.9],
      [1000, 460, 0.4],
    ].map(([x, y, o], i) => (
      <circle key={i} cx={x} cy={y} r={i % 5 === 0 ? 1.5 : 1} fill="white" opacity={o as number} />
    ))}
    {/* Subtle dot grid */}
    {Array.from({ length: 20 }).map((_, row) =>
      Array.from({ length: 30 }).map((_, col) => (
        <circle
          key={`${row}-${col}`}
          cx={col * 50 + 25}
          cy={row * 48 + 24}
          r="0.8"
          fill="white"
          opacity="0.04"
        />
      ))
    )}
  </svg>
);

const Aurora: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="a-band1" x1="0" y1="0" x2="1" y2="0">
        <stop offset="0%" stopColor="#00ff88" stopOpacity="0" />
        <stop offset="30%" stopColor="#00cc66" stopOpacity="0.12" />
        <stop offset="70%" stopColor="#0088ff" stopOpacity="0.1" />
        <stop offset="100%" stopColor="#0044cc" stopOpacity="0" />
      </linearGradient>
      <linearGradient id="a-band2" x1="0" y1="0" x2="1" y2="0">
        <stop offset="0%" stopColor="#00aaff" stopOpacity="0" />
        <stop offset="40%" stopColor="#00ddaa" stopOpacity="0.1" />
        <stop offset="100%" stopColor="#00ff88" stopOpacity="0" />
      </linearGradient>
      <filter id="a-blur">
        <feGaussianBlur stdDeviation="18" />
      </filter>
    </defs>
    {/* Aurora bands */}
    <path
      d="M0 320 Q360 260 720 300 Q1080 340 1440 280 L1440 380 Q1080 440 720 400 Q360 360 0 420Z"
      fill="url(#a-band1)"
      filter="url(#a-blur)"
    />
    <path
      d="M0 200 Q400 150 800 180 Q1100 210 1440 160 L1440 260 Q1100 310 800 280 Q400 250 0 300Z"
      fill="url(#a-band2)"
      filter="url(#a-blur)"
    />
    {/* Stars */}
    {[
      [100, 60, 0.7],
      [320, 40, 0.5],
      [550, 100, 0.8],
      [800, 50, 0.6],
      [1050, 80, 0.7],
      [1280, 40, 0.5],
      [200, 480, 0.4],
      [500, 520, 0.5],
      [900, 500, 0.6],
      [1200, 460, 0.4],
      [700, 700, 0.5],
      [400, 750, 0.4],
      [1100, 720, 0.6],
      [150, 830, 0.3],
      [950, 800, 0.5],
    ].map(([x, y, o], i) => (
      <circle key={i} cx={x} cy={y} r="1" fill="white" opacity={o as number} />
    ))}
  </svg>
);

const Dusk: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <radialGradient id="d-orb1" cx="60%" cy="20%" r="40%">
        <stop offset="0%" stopColor="#8030c0" stopOpacity="0.3" />
        <stop offset="100%" stopColor="#000" stopOpacity="0" />
      </radialGradient>
      <radialGradient id="d-orb2" cx="25%" cy="70%" r="35%">
        <stop offset="0%" stopColor="#c040a0" stopOpacity="0.2" />
        <stop offset="100%" stopColor="#000" stopOpacity="0" />
      </radialGradient>
      <filter id="d-blur">
        <feGaussianBlur stdDeviation="30" />
      </filter>
    </defs>
    <rect width="1440" height="900" fill="url(#d-orb1)" />
    <rect width="1440" height="900" fill="url(#d-orb2)" />
    {/* Geometric lines */}
    {[0, 1, 2, 3, 4].map((i) => (
      <line
        key={i}
        x1={720}
        y1={450}
        x2={i * 360}
        y2={0}
        stroke="rgba(180,100,255,0.06)"
        strokeWidth="1"
      />
    ))}
    {[0, 1, 2, 3, 4].map((i) => (
      <line
        key={i}
        x1={720}
        y1={450}
        x2={i * 360}
        y2={900}
        stroke="rgba(180,100,255,0.06)"
        strokeWidth="1"
      />
    ))}
    {/* Stars */}
    {[
      [80, 70, 0.6],
      [300, 50, 0.8],
      [600, 90, 0.5],
      [900, 40, 0.7],
      [1200, 80, 0.6],
      [1400, 120, 0.5],
      [200, 600, 0.4],
      [700, 650, 0.5],
      [1100, 580, 0.6],
    ].map(([x, y, o], i) => (
      <circle key={i} cx={x} cy={y} r="1.2" fill="white" opacity={o as number} />
    ))}
  </svg>
);

const Slate: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <pattern id="sl-grid" x="0" y="0" width="60" height="60" patternUnits="userSpaceOnUse">
        <path d="M60 0L0 0 0 60" fill="none" stroke="rgba(255,255,255,0.04)" strokeWidth="1" />
      </pattern>
      <radialGradient id="sl-glow" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#1a2a50" stopOpacity="0.5" />
        <stop offset="100%" stopColor="#000" stopOpacity="0" />
      </radialGradient>
    </defs>
    <rect width="1440" height="900" fill="url(#sl-grid)" />
    <rect width="1440" height="900" fill="url(#sl-glow)" />
    {/* Horizontal accent lines */}
    <line x1="0" y1="300" x2="1440" y2="300" stroke="rgba(100,150,255,0.05)" strokeWidth="1" />
    <line x1="0" y1="600" x2="1440" y2="600" stroke="rgba(100,150,255,0.05)" strokeWidth="1" />
    <line x1="360" y1="0" x2="360" y2="900" stroke="rgba(100,150,255,0.04)" strokeWidth="1" />
    <line x1="1080" y1="0" x2="1080" y2="900" stroke="rgba(100,150,255,0.04)" strokeWidth="1" />
  </svg>
);

const Forest: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <radialGradient id="f-glow" cx="50%" cy="40%" r="45%">
        <stop offset="0%" stopColor="#204020" stopOpacity="0.5" />
        <stop offset="100%" stopColor="#000" stopOpacity="0" />
      </radialGradient>
    </defs>
    <rect width="1440" height="900" fill="url(#f-glow)" />
    {/* Organic dot scatter */}
    {[
      [200, 150, 2],
      [450, 200, 1.5],
      [700, 120, 2.5],
      [950, 180, 1.5],
      [1200, 140, 2],
      [100, 400, 1.5],
      [350, 450, 2],
      [600, 380, 1.5],
      [850, 420, 2.5],
      [1100, 390, 1.5],
      [1350, 430, 2],
      [250, 650, 1.5],
      [500, 700, 2],
      [750, 640, 1.5],
      [1000, 680, 2],
      [1250, 650, 1.5],
      [150, 820, 2],
      [600, 800, 1.5],
      [1050, 840, 2],
      [1400, 780, 1.5],
    ].map(([x, y, r], i) => (
      <circle key={i} cx={x} cy={y} r={r} fill="rgba(60,200,80,0.15)" />
    ))}
    {/* Subtle grid */}
    {Array.from({ length: 12 }).map((_, row) =>
      Array.from({ length: 20 }).map((_, col) => (
        <circle
          key={`${row}-${col}`}
          cx={col * 75 + 37}
          cy={row * 80 + 40}
          r="1"
          fill="rgba(60,200,60,0.05)"
        />
      ))
    )}
  </svg>
);

const WALLPAPER_OVERLAYS: Record<Wallpaper, React.FC> = {
  midnight: Midnight,
  aurora: Aurora,
  dusk: Dusk,
  slate: Slate,
  forest: Forest,
};

export const WallpaperOverlay: React.FC<{ wallpaper: Wallpaper }> = ({ wallpaper }) => {
  const Overlay = WALLPAPER_OVERLAYS[wallpaper];
  return <Overlay />;
};
