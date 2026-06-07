import React from 'react';
import { Wallpaper } from '../../stores/useDesktopStore';

const SVG_PROPS = {
  style: { position: 'absolute' as const, inset: 0, width: '100%', height: '100%' },
  xmlns: 'http://www.w3.org/2000/svg',
  preserveAspectRatio: 'xMidYMid slice',
  viewBox: '0 0 1440 900',
};

// Deterministic pseudo-random so star fields are stable across renders.
const rng = (seed: number): (() => number) => {
  let s = seed;
  return () => {
    s = (s * 1103515245 + 12345) & 0x7fffffff;
    return s / 0x7fffffff;
  };
};

const StarField: React.FC<{ count: number; seed: number; maxY: number; twinkle?: boolean }> = ({
  count,
  seed,
  maxY,
  twinkle,
}) => {
  const r = rng(seed);
  return (
    <>
      {Array.from({ length: count }).map((_, i) => {
        const cx = r() * 1440;
        const cy = r() * maxY;
        const rad = 0.5 + r() * 1.4;
        const o = 0.4 + r() * 0.6;
        const dur = 2 + r() * 4;
        const delay = r() * 5;
        return (
          <circle
            key={i}
            cx={cx}
            cy={cy}
            r={rad}
            fill="#fff"
            opacity={o}
            style={
              twinkle && i % 3 === 0
                ? { animation: `wp-twinkle ${dur}s ease-in-out ${delay}s infinite` }
                : undefined
            }
          />
        );
      })}
    </>
  );
};

// ─── Midnight — Neon city skyline mirrored on water ──────────────────────────
const Midnight: React.FC = () => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id="mn-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#070a1f" />
        <stop offset="45%" stopColor="#141a44" />
        <stop offset="75%" stopColor="#3a1f5c" />
        <stop offset="100%" stopColor="#6b2f63" />
      </linearGradient>
      <radialGradient id="mn-moon" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#fff7e8" />
        <stop offset="70%" stopColor="#fce9c0" />
        <stop offset="100%" stopColor="#fce9c0" stopOpacity="0" />
      </radialGradient>
      <linearGradient id="mn-water" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#6b2f63" />
        <stop offset="40%" stopColor="#2a1a48" />
        <stop offset="100%" stopColor="#0c0e26" />
      </linearGradient>
      <style>{`
        @keyframes wp-twinkle { 0%,100%{opacity:.9} 50%{opacity:.15} }
        @keyframes mn-shimmer { 0%,100%{opacity:.5;transform:translateX(0)} 50%{opacity:.8;transform:translateX(6px)} }
      `}</style>
    </defs>

    <rect width="1440" height="630" fill="url(#mn-sky)" />
    <StarField count={90} seed={7} maxY={520} twinkle />

    {/* Moon */}
    <circle cx="1140" cy="150" r="150" fill="url(#mn-moon)" opacity="0.5" />
    <circle cx="1140" cy="150" r="64" fill="#fdf3d8" />
    <circle cx="1118" cy="132" r="11" fill="#ecdcb6" opacity="0.6" />
    <circle cx="1158" cy="168" r="8" fill="#ecdcb6" opacity="0.55" />
    <circle cx="1166" cy="128" r="6" fill="#ecdcb6" opacity="0.5" />

    {/* Far skyline — hazy purple */}
    <g fill="#2f2155" opacity="0.85">
      <rect x="60" y="430" width="70" height="200" />
      <rect x="150" y="390" width="54" height="240" />
      <rect x="230" y="450" width="80" height="180" />
      <rect x="340" y="410" width="48" height="220" />
      <rect x="900" y="420" width="64" height="210" />
      <rect x="990" y="380" width="52" height="250" />
      <rect x="1070" y="445" width="76" height="185" />
      <rect x="1170" y="405" width="50" height="225" />
      <rect x="1260" y="440" width="70" height="190" />
    </g>

    {/* Mid skyline */}
    <g fill="#1c1640">
      <rect x="0" y="470" width="90" height="160" />
      <rect x="120" y="430" width="66" height="200" />
      <rect x="210" y="490" width="78" height="140" />
      <rect x="430" y="400" width="60" height="230" />
      <rect x="520" y="350" width="44" height="280" />
      <rect x="600" y="440" width="84" height="190" />
      <rect x="760" y="410" width="58" height="220" />
      <rect x="850" y="470" width="72" height="160" />
      <rect x="1130" y="455" width="62" height="175" />
      <rect x="1230" y="415" width="56" height="215" />
      <rect x="1330" y="465" width="110" height="165" />
    </g>

    {/* Front skyline — darkest, lit windows */}
    <g fill="#0c0a24">
      <rect x="270" y="370" width="92" height="260" />
      <rect x="380" y="320" width="58" height="310" />
      <rect x="660" y="300" width="74" height="330" />
      <rect x="700" y="270" width="20" height="40" />
      <rect x="980" y="350" width="88" height="280" />
      <rect x="1090" y="330" width="60" height="300" />
    </g>
    {/* Lit windows on front buildings */}
    <g fill="#ffd66b">
      {[
        [286, 392], [320, 392], [286, 420], [320, 448], [286, 476], [320, 504], [286, 560],
        [396, 342], [414, 342], [396, 380], [414, 418], [396, 470], [414, 520], [396, 580],
        [676, 322], [702, 322], [676, 360], [702, 398], [676, 440], [702, 490], [676, 540], [702, 590],
        [996, 372], [1030, 372], [996, 410], [1030, 458], [996, 510], [1030, 560],
        [1106, 352], [1124, 352], [1106, 400], [1124, 450], [1106, 520], [1124, 580],
      ].map(([x, y], i) => (
        <rect
          key={i}
          x={x}
          y={y}
          width="12"
          height="16"
          rx="1"
          opacity={0.5 + (i % 4) * 0.13}
        />
      ))}
    </g>

    {/* Water + reflection */}
    <rect y="630" width="1440" height="270" fill="url(#mn-water)" />
    <g fill="#0c0a24" opacity="0.4">
      <rect x="270" y="630" width="92" height="120" />
      <rect x="380" y="630" width="58" height="150" />
      <rect x="660" y="630" width="74" height="170" />
      <rect x="980" y="630" width="88" height="140" />
      <rect x="1090" y="630" width="60" height="150" />
    </g>
    <circle cx="1140" cy="690" r="40" fill="#fdf3d8" opacity="0.25" />
    {[660, 700, 740, 780].map((y, i) => (
      <rect
        key={i}
        x="200"
        y={y}
        width="1040"
        height="3"
        fill="#ffd66b"
        opacity="0.1"
        style={{ animation: `mn-shimmer ${3 + i}s ease-in-out ${i * 0.5}s infinite` }}
      />
    ))}
  </svg>
);

// ─── Aurora — flowing northern lights over a frozen lake ─────────────────────
const Aurora: React.FC = () => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id="au-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#040814" />
        <stop offset="55%" stopColor="#081428" />
        <stop offset="100%" stopColor="#0e2236" />
      </linearGradient>
      <linearGradient id="au-r1" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#48ffb0" stopOpacity="0" />
        <stop offset="40%" stopColor="#3affa0" stopOpacity="0.55" />
        <stop offset="100%" stopColor="#0affc8" stopOpacity="0" />
      </linearGradient>
      <linearGradient id="au-r2" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#7a5cff" stopOpacity="0" />
        <stop offset="45%" stopColor="#9d6bff" stopOpacity="0.5" />
        <stop offset="100%" stopColor="#5cc8ff" stopOpacity="0" />
      </linearGradient>
      <linearGradient id="au-r3" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#22d9a0" stopOpacity="0" />
        <stop offset="50%" stopColor="#3affb0" stopOpacity="0.4" />
        <stop offset="100%" stopColor="#22d9a0" stopOpacity="0" />
      </linearGradient>
      <linearGradient id="au-lake" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#0e2236" />
        <stop offset="100%" stopColor="#04101e" />
      </linearGradient>
      <filter id="au-blur" x="-30%" y="-30%" width="160%" height="160%">
        <feGaussianBlur stdDeviation="22" />
      </filter>
      <style>{`
        @keyframes au-drift1 { 0%,100%{transform:translateX(0) scaleY(1)} 50%{transform:translateX(-40px) scaleY(1.08)} }
        @keyframes au-drift2 { 0%,100%{transform:translateX(0) scaleY(1)} 50%{transform:translateX(50px) scaleY(0.94)} }
        @keyframes au-drift3 { 0%,100%{transform:translateX(0)} 50%{transform:translateX(-28px)} }
      `}</style>
    </defs>

    <rect width="1440" height="660" fill="url(#au-sky)" />
    <StarField count={110} seed={21} maxY={620} twinkle />

    {/* Aurora ribbons */}
    <g filter="url(#au-blur)">
      <path
        d="M-100 120 Q 300 40 700 180 T 1540 120 L1540 470 Q 1100 360 700 480 T -100 420 Z"
        fill="url(#au-r1)"
        style={{ animation: 'au-drift1 14s ease-in-out infinite', transformOrigin: 'center' }}
      />
      <path
        d="M-100 220 Q 400 120 820 260 T 1540 220 L1540 520 Q 1000 420 600 540 T -100 500 Z"
        fill="url(#au-r2)"
        style={{ animation: 'au-drift2 18s ease-in-out infinite', transformOrigin: 'center' }}
      />
      <path
        d="M-100 70 Q 500 10 900 130 T 1540 90 L1540 320 Q 980 240 540 340 T -100 300 Z"
        fill="url(#au-r3)"
        style={{ animation: 'au-drift3 11s ease-in-out infinite', transformOrigin: 'center' }}
      />
    </g>

    {/* Mountain range silhouette */}
    <path
      d="M0 660 L0 540 L120 470 L240 540 L360 430 L520 560 L640 480 L780 580 L920 450 L1080 560 L1220 470 L1340 540 L1440 500 L1440 660 Z"
      fill="#061018"
    />
    {/* Snow caps */}
    <g fill="#16384a" opacity="0.8">
      <path d="M360 430 L400 462 L320 462 Z" />
      <path d="M920 450 L956 478 L884 478 Z" />
      <path d="M120 470 L150 494 L90 494 Z" />
    </g>

    {/* Frozen lake reflection */}
    <rect y="660" width="1440" height="240" fill="url(#au-lake)" />
    <g filter="url(#au-blur)" opacity="0.4" transform="translate(0,1320) scale(1,-1)">
      <path d="M-100 120 Q 300 40 700 180 T 1540 120 L1540 400 Q 1100 320 700 420 T -100 380 Z" fill="url(#au-r1)" />
      <path d="M-100 70 Q 500 10 900 130 T 1540 90 L1540 300 Q 980 220 540 320 T -100 280 Z" fill="url(#au-r3)" />
    </g>
  </svg>
);

// ─── Dusk — layered sunset hills with a low sun ──────────────────────────────
const Dusk: React.FC = () => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id="du-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#2b1a5e" />
        <stop offset="32%" stopColor="#6e2f7a" />
        <stop offset="58%" stopColor="#c5447a" />
        <stop offset="78%" stopColor="#f47c4a" />
        <stop offset="100%" stopColor="#ffc15e" />
      </linearGradient>
      <radialGradient id="du-sunglow" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#fff4c4" />
        <stop offset="35%" stopColor="#ffd06a" stopOpacity="0.7" />
        <stop offset="100%" stopColor="#ff9a4a" stopOpacity="0" />
      </radialGradient>
      <linearGradient id="du-sun" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#fff2c0" />
        <stop offset="100%" stopColor="#ffcf66" />
      </linearGradient>
      <style>{`
        @keyframes du-glow { 0%,100%{opacity:.85;transform:scale(1)} 50%{opacity:1;transform:scale(1.04)} }
        @keyframes du-bird { 0%{transform:translate(0,0)} 50%{transform:translate(20px,-10px)} 100%{transform:translate(40px,0)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#du-sky)" />

    {/* Sun */}
    <circle
      cx="720"
      cy="430"
      r="280"
      fill="url(#du-sunglow)"
      style={{ animation: 'du-glow 7s ease-in-out infinite', transformOrigin: '720px 430px' }}
    />
    <circle cx="720" cy="430" r="115" fill="url(#du-sun)" />

    {/* Birds */}
    {[
      [340, 200],
      [400, 175],
      [452, 205],
      [1010, 170],
      [1068, 150],
      [1120, 185],
    ].map(([x, y], i) => (
      <path
        key={i}
        d={`M${x} ${y} Q ${x + 14} ${y - 12} ${x + 28} ${y} Q ${x + 42} ${y - 12} ${x + 56} ${y}`}
        fill="none"
        stroke="#3a1f3a"
        strokeWidth="3"
        strokeLinecap="round"
        opacity="0.55"
        style={{ animation: `du-bird ${5 + i}s ease-in-out ${i * 0.6}s infinite` }}
      />
    ))}

    {/* Layered hills — back (warm, hazy) to front (dark) */}
    <path d="M0 560 Q 360 500 720 545 T 1440 530 L1440 900 L0 900 Z" fill="#d9633f" opacity="0.55" />
    <path d="M0 620 Q 420 560 840 605 T 1440 600 L1440 900 L0 900 Z" fill="#a83f5e" opacity="0.7" />
    <path d="M0 690 Q 300 630 680 680 Q 1020 720 1440 670 L1440 900 L0 900 Z" fill="#6a2952" />
    <path d="M0 770 Q 380 720 760 765 T 1440 750 L1440 900 L0 900 Z" fill="#3f1838" />
    <path d="M0 840 Q 320 800 720 835 T 1440 825 L1440 900 L0 900 Z" fill="#1f0c22" />

    {/* Lone tree silhouette on a front ridge */}
    <g fill="#1f0c22">
      <rect x="1158" y="690" width="8" height="80" />
      <circle cx="1162" cy="678" r="30" />
      <circle cx="1140" cy="690" r="20" />
      <circle cx="1184" cy="690" r="18" />
    </g>
  </svg>
);

// ─── Slate — minimalist misty mountain ranges ────────────────────────────────
const Slate: React.FC = () => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id="sl-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#c3d2e6" />
        <stop offset="60%" stopColor="#dde7f2" />
        <stop offset="100%" stopColor="#eef3f8" />
      </linearGradient>
      <radialGradient id="sl-sun" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#fff" stopOpacity="0.95" />
        <stop offset="60%" stopColor="#f4f8fc" stopOpacity="0.5" />
        <stop offset="100%" stopColor="#f4f8fc" stopOpacity="0" />
      </radialGradient>
      <filter id="sl-mist" x="-20%" y="-50%" width="140%" height="200%">
        <feGaussianBlur stdDeviation="16" />
      </filter>
      <style>{`
        @keyframes sl-mist { 0%,100%{opacity:.5;transform:translateX(0)} 50%{opacity:.8;transform:translateX(40px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#sl-sky)" />

    {/* Soft sun */}
    <circle cx="1020" cy="240" r="190" fill="url(#sl-sun)" />
    <circle cx="1020" cy="240" r="68" fill="#fbfdff" opacity="0.9" />

    {/* Mountain ranges — atmospheric perspective (light/far → dark/near) */}
    <path
      d="M0 470 L260 300 L470 460 L640 330 L860 500 L1080 360 L1280 480 L1440 400 L1440 900 L0 900 Z"
      fill="#a9bdd4"
      opacity="0.55"
    />
    <path
      d="M0 560 L220 420 L430 560 L660 410 L900 580 L1150 440 L1440 560 L1440 900 L0 900 Z"
      fill="#8298b6"
      opacity="0.7"
    />
    {/* mist band */}
    <rect
      x="-50"
      y="560"
      width="1540"
      height="60"
      fill="#eef3f8"
      filter="url(#sl-mist)"
      style={{ animation: 'sl-mist 16s ease-in-out infinite' }}
    />
    <path
      d="M0 660 L300 520 L560 680 L820 510 L1080 670 L1340 540 L1440 620 L1440 900 L0 900 Z"
      fill="#5a7193"
    />
    <path
      d="M0 760 L240 650 L520 780 L780 640 L1060 780 L1320 660 L1440 720 L1440 900 L0 900 Z"
      fill="#3a4d6e"
    />
    {/* lower mist */}
    <rect
      x="-50"
      y="740"
      width="1540"
      height="50"
      fill="#dde7f2"
      filter="url(#sl-mist)"
      opacity="0.8"
      style={{ animation: 'sl-mist 20s ease-in-out 2s infinite' }}
    />
    <path d="M0 840 L360 770 L760 850 L1120 760 L1440 820 L1440 900 L0 900 Z" fill="#1f2d49" />
  </svg>
);

// ─── Forest — rolling hills at dawn with treeline ────────────────────────────
const Forest: React.FC = () => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id="fo-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#fcd9a0" />
        <stop offset="40%" stopColor="#ffeccb" />
        <stop offset="100%" stopColor="#e6f5e8" />
      </linearGradient>
      <radialGradient id="fo-sun" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#fff6d8" />
        <stop offset="45%" stopColor="#ffe7a0" stopOpacity="0.7" />
        <stop offset="100%" stopColor="#ffe7a0" stopOpacity="0" />
      </radialGradient>
      <style>{`
        @keyframes fo-sun { 0%,100%{opacity:.85} 50%{opacity:1} }
        @keyframes fo-cloud { 0%{transform:translateX(0)} 100%{transform:translateX(60px)} }
        @keyframes fo-bird { 0%{transform:translate(0,0)} 50%{transform:translate(24px,-8px)} 100%{transform:translate(48px,0)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#fo-sky)" />

    {/* Sun + glow */}
    <circle
      cx="500"
      cy="270"
      r="240"
      fill="url(#fo-sun)"
      style={{ animation: 'fo-sun 6s ease-in-out infinite' }}
    />
    <circle cx="500" cy="270" r="92" fill="#fff2cf" />

    {/* Soft clouds */}
    <g fill="#fff" opacity="0.55" style={{ animation: 'fo-cloud 40s linear infinite alternate' }}>
      <ellipse cx="980" cy="180" rx="120" ry="34" />
      <ellipse cx="1060" cy="160" rx="90" ry="40" />
      <ellipse cx="900" cy="195" rx="80" ry="28" />
    </g>
    <g fill="#fff" opacity="0.4" style={{ animation: 'fo-cloud 55s linear 4s infinite alternate' }}>
      <ellipse cx="300" cy="120" rx="100" ry="28" />
      <ellipse cx="370" cy="108" rx="70" ry="32" />
    </g>

    {/* Birds */}
    {[
      [760, 230],
      [810, 212],
      [860, 238],
    ].map(([x, y], i) => (
      <path
        key={i}
        d={`M${x} ${y} Q ${x + 12} ${y - 10} ${x + 24} ${y} Q ${x + 36} ${y - 10} ${x + 48} ${y}`}
        fill="none"
        stroke="#5a7a4a"
        strokeWidth="2.5"
        strokeLinecap="round"
        opacity="0.5"
        style={{ animation: `fo-bird ${6 + i}s ease-in-out ${i * 0.7}s infinite` }}
      />
    ))}

    {/* Rolling hills — far (light) to near (dark), with paper-cut treelines */}
    <path d="M0 540 Q 360 470 720 525 T 1440 510 L1440 900 L0 900 Z" fill="#bfe3a8" opacity="0.8" />
    <path d="M0 615 Q 300 545 680 600 Q 1040 650 1440 590 L1440 900 L0 900 Z" fill="#8fcf7e" />
    {/* treeline on hill 2 */}
    <g fill="#6db35f">
      {Array.from({ length: 24 }).map((_, i) => {
        const x = i * 62 + 20;
        const y = 600 + Math.sin(i * 0.7) * 18;
        return <path key={i} d={`M${x} ${y} l10 -26 l10 26 Z`} />;
      })}
    </g>
    <path d="M0 700 Q 380 630 760 690 T 1440 670 L1440 900 L0 900 Z" fill="#4f9a48" />
    {/* treeline on hill 3 */}
    <g fill="#3d7d3a">
      {Array.from({ length: 20 }).map((_, i) => {
        const x = i * 74 + 30;
        const y = 686 + Math.sin(i * 0.9 + 1) * 16;
        return <path key={i} d={`M${x} ${y} l12 -32 l12 32 Z`} />;
      })}
    </g>
    <path d="M0 790 Q 320 730 720 785 T 1440 760 L1440 900 L0 900 Z" fill="#2c6630" />
    <path d="M0 860 Q 400 820 800 855 T 1440 840 L1440 900 L0 900 Z" fill="#1a4421" />
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
