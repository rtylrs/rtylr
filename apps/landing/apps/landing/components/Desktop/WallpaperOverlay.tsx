import React from 'react';
import { Wallpaper } from '../../stores/useDesktopStore';

const SVG_PROPS = {
  style: {
    position: 'absolute' as const,
    inset: 0,
    width: '100%',
    height: '100%',
    pointerEvents: 'none' as const,
  },
  xmlns: 'http://www.w3.org/2000/svg',
  preserveAspectRatio: 'xMidYMid slice',
  viewBox: '0 0 1440 900',
};

const KEYFRAMES = `
@keyframes wp-twinkle { 0%,100%{opacity:.9} 50%{opacity:.12} }
@keyframes mn-beacon { 0%,45%,100%{opacity:1} 55%,95%{opacity:.1} }
@keyframes au-ray { 0%,100%{opacity:.35;transform:scaleY(1)} 50%{opacity:.85;transform:scaleY(1.12)} }
@keyframes du-glow { 0%,100%{opacity:.85;transform:scale(1)} 50%{opacity:1;transform:scale(1.04)} }
@keyframes du-bird { 0%{transform:translate(0,0)} 50%{transform:translate(20px,-10px)} 100%{transform:translate(40px,0)} }
@keyframes sl-mist { 0%,100%{opacity:.5;transform:translateX(0)} 50%{opacity:.8;transform:translateX(40px)} }
@keyframes fo-sun { 0%,100%{opacity:.85} 50%{opacity:1} }
@keyframes fo-cloud { 0%{transform:translateX(0)} 100%{transform:translateX(70px)} }
@keyframes fo-bird { 0%{transform:translate(0,0)} 50%{transform:translate(24px,-8px)} 100%{transform:translate(48px,0)} }
@keyframes fo-mist { 0%,100%{opacity:.55;transform:translateX(0)} 50%{opacity:.85;transform:translateX(36px)} }
`;

// Static CSS approximations of each scene, used for cheap previews (e.g. the
// Settings thumbnails) instead of mounting the full SVG five times over.
export const WALLPAPER_PREVIEWS: Record<Wallpaper, string> = {
  midnight:
    'linear-gradient(180deg, #070b22 0%, #141d45 36%, #34315e 58%, #7a4a5e 72%, #1a1c42 76%, #0c0d26 100%)',
  aurora:
    'linear-gradient(180deg, rgba(61,255,154,0) 18%, rgba(61,255,154,0.4) 42%, rgba(176,107,255,0.25) 58%, rgba(31,217,192,0) 72%), linear-gradient(180deg, #03060f 0%, #06121f 55%, #0a2030 76%, #04101a 78%, #030b14 100%)',
  dusk: 'linear-gradient(180deg, #2b1a5e 0%, #6e2f7a 30%, #c5447a 52%, #f47c4a 68%, #6a2952 78%, #1f0c22 100%)',
  slate:
    'linear-gradient(180deg, #c3d2e6 0%, #dde7f2 40%, #8298b6 62%, #5a7193 75%, #3a4d6e 88%, #1f2d49 100%)',
  forest:
    'linear-gradient(180deg, #ffd29a 0%, #ffe7c4 32%, #e2f1e6 50%, #cfe6b4 62%, #86c873 76%, #3f8c45 88%, #1f5a2b 100%)',
};

type SceneProps = { id: (name: string) => string };

// Deterministic pseudo-random so generated detail is stable across renders.
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
        const rad = 0.5 + r() * 1.5;
        const o = 0.35 + r() * 0.6;
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

// ─── Midnight — city skyline glowing under a hazy twilight ───────────────────
// Skyline built as continuous stepped silhouettes (one path per depth layer)
// so it reads as a connected cityscape, not scattered boxes. Light pollution
// glow + base haze tie it together; soft window bokeh gives it life.
const buildSkyline = (
  seed: number,
  baseY: number,
  hMin: number,
  hMax: number,
  wMin: number,
  wMax: number
): { path: string; segs: { x: number; w: number; top: number }[] } => {
  const r = rng(seed);
  const segs: { x: number; w: number; top: number }[] = [];
  let x = -40;
  while (x < 1480) {
    const w = wMin + r() * (wMax - wMin);
    const h = hMin + r() * (hMax - hMin);
    segs.push({ x, w, top: baseY - h });
    x += w;
  }
  let d = `M${segs[0].x} 900 L${segs[0].x} ${segs[0].top}`;
  segs.forEach((s) => {
    d += ` L${s.x} ${s.top} L${s.x + s.w} ${s.top}`;
  });
  d += ` L${x} 900 Z`;
  return { path: d, segs };
};

const Windows: React.FC<{
  segs: { x: number; w: number; top: number }[];
  baseY: number;
  seed: number;
}> = ({ segs, baseY, seed }) => {
  const r = rng(seed);
  const rects: React.ReactNode[] = [];
  segs.forEach((s, si) => {
    if (s.w < 26) return;
    for (let wy = s.top + 14; wy < baseY - 10; wy += 16) {
      for (let wx = s.x + 8; wx < s.x + s.w - 12; wx += 14) {
        if (r() > 0.55) continue;
        rects.push(
          <rect
            key={`${si}-${wx}-${wy}`}
            x={wx}
            y={wy}
            width={6}
            height={8}
            rx={1}
            opacity={0.25 + r() * 0.6}
          />
        );
      }
    }
  });
  return <g fill="#ffd27a">{rects}</g>;
};

const Midnight: React.FC<SceneProps> = ({ id }) => {
  const far = buildSkyline(3, 560, 80, 200, 70, 130);
  const mid = buildSkyline(11, 610, 120, 300, 80, 150);
  const near = buildSkyline(29, 660, 160, 380, 90, 170);
  return (
    <svg {...SVG_PROPS}>
      <defs>
        <linearGradient id={id('mn-sky')} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#070b22" />
          <stop offset="40%" stopColor="#141d45" />
          <stop offset="70%" stopColor="#34315e" />
          <stop offset="100%" stopColor="#7a4a5e" />
        </linearGradient>
        <radialGradient id={id('mn-moon')} cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#fff7e8" />
          <stop offset="65%" stopColor="#fce9c0" stopOpacity="0.55" />
          <stop offset="100%" stopColor="#fce9c0" stopOpacity="0" />
        </radialGradient>
        <radialGradient id={id('mn-pollution')} cx="50%" cy="100%" r="75%">
          <stop offset="0%" stopColor="#ffb27a" stopOpacity="0.5" />
          <stop offset="45%" stopColor="#b9628a" stopOpacity="0.22" />
          <stop offset="100%" stopColor="#b9628a" stopOpacity="0" />
        </radialGradient>
        <filter id={id('mn-haze')} x="-20%" y="-60%" width="140%" height="220%">
          <feGaussianBlur stdDeviation="26" />
        </filter>
      </defs>

      <rect width="1440" height="900" fill={`url(#${id('mn-sky')})`} />
      <StarField count={80} seed={7} maxY={460} twinkle />

      {/* Moon */}
      <circle cx="1130" cy="160" r="170" fill={`url(#${id('mn-moon')})`} />
      <circle cx="1130" cy="160" r="62" fill="#fdf3d8" />
      <circle cx="1110" cy="144" r="10" fill="#ecdcb6" opacity="0.55" />
      <circle cx="1148" cy="176" r="7" fill="#ecdcb6" opacity="0.5" />

      {/* Light-pollution glow behind the skyline */}
      <rect x="0" y="360" width="1440" height="540" fill={`url(#${id('mn-pollution')})`} />

      {/* Far layer */}
      <path d={far.path} fill="#2c2f58" opacity="0.7" />
      {/* Mid layer */}
      <path d={mid.path} fill="#1a1c42" />
      <Windows segs={mid.segs} baseY={610} seed={101} />
      {/* Near layer */}
      <path d={near.path} fill="#0c0d26" />
      <Windows segs={near.segs} baseY={660} seed={202} />

      {/* Beacon lights on the two tallest near towers */}
      {near.segs
        .filter((s) => s.top < 380)
        .slice(0, 3)
        .map((s, i) => (
          <circle
            key={i}
            cx={s.x + s.w / 2}
            cy={s.top - 2}
            r={3}
            fill="#ff4d4d"
            style={{ animation: `mn-beacon ${2 + i * 0.6}s ease-in-out ${i * 0.4}s infinite` }}
          />
        ))}

      {/* Ground haze — glowing fog that bleeds the city light upward */}
      <rect
        x="-40"
        y="640"
        width="1520"
        height="120"
        fill="#9a6a7a"
        opacity="0.5"
        filter={`url(#${id('mn-haze')})`}
      />
      <rect
        x="-40"
        y="700"
        width="1520"
        height="160"
        fill="#d98a6a"
        opacity="0.35"
        filter={`url(#${id('mn-haze')})`}
      />
    </svg>
  );
};

// ─── Aurora — full shimmering curtains of light over a low ridge ─────────────
const Aurora: React.FC<SceneProps> = ({ id }) => {
  const rays = (seed: number, yBase: number, amp: number, count: number) => {
    const r = rng(seed);
    return Array.from({ length: count }).map((_, i) => {
      const x = (i / count) * 1520 - 40;
      const w = 10 + r() * 26;
      const top = yBase + Math.sin(i * 0.4) * amp - 30 - r() * 60;
      const h = 240 + r() * 220;
      const dur = 4 + r() * 5;
      const delay = r() * 5;
      return { x, w, top, h, dur, delay };
    });
  };
  const greenRays = rays(5, 230, 70, 46);
  const violetRays = rays(17, 330, 90, 34);
  return (
    <svg {...SVG_PROPS}>
      <defs>
        <linearGradient id={id('au-sky')} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#03060f" />
          <stop offset="55%" stopColor="#06121f" />
          <stop offset="100%" stopColor="#0a2030" />
        </linearGradient>
        <linearGradient id={id('au-green')} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#4dffb0" stopOpacity="0" />
          <stop offset="35%" stopColor="#3dff9a" stopOpacity="0.75" />
          <stop offset="70%" stopColor="#1fd9c0" stopOpacity="0.4" />
          <stop offset="100%" stopColor="#1fd9c0" stopOpacity="0" />
        </linearGradient>
        <linearGradient id={id('au-violet')} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#9d6bff" stopOpacity="0" />
          <stop offset="45%" stopColor="#b06bff" stopOpacity="0.55" />
          <stop offset="100%" stopColor="#ff6bd0" stopOpacity="0" />
        </linearGradient>
        <radialGradient id={id('au-bloom')} cx="50%" cy="38%" r="60%">
          <stop offset="0%" stopColor="#2dffb0" stopOpacity="0.28" />
          <stop offset="100%" stopColor="#2dffb0" stopOpacity="0" />
        </radialGradient>
        <filter id={id('au-soft')} x="-40%" y="-40%" width="180%" height="180%">
          <feGaussianBlur stdDeviation="14" />
        </filter>
        <linearGradient id={id('au-lake')} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#0a2030" />
          <stop offset="100%" stopColor="#030b14" />
        </linearGradient>
      </defs>

      <rect width="1440" height="700" fill={`url(#${id('au-sky')})`} />
      <StarField count={140} seed={21} maxY={520} twinkle />

      {/* Broad green bloom filling the sky */}
      <rect width="1440" height="700" fill={`url(#${id('au-bloom')})`} />

      {/* Violet curtain (behind) */}
      <g filter={`url(#${id('au-soft')})`}>
        {violetRays.map((r, i) => (
          <rect
            key={`v${i}`}
            x={r.x}
            y={r.top}
            width={r.w}
            height={r.h}
            rx={r.w / 2}
            fill={`url(#${id('au-violet')})`}
            style={{
              animation: `au-ray ${r.dur}s ease-in-out ${r.delay}s infinite`,
              transformBox: 'fill-box',
              transformOrigin: 'center top',
            }}
          />
        ))}
      </g>
      {/* Green curtain (front, brighter) */}
      <g filter={`url(#${id('au-soft')})`}>
        {greenRays.map((r, i) => (
          <rect
            key={`g${i}`}
            x={r.x}
            y={r.top}
            width={r.w}
            height={r.h}
            rx={r.w / 2}
            fill={`url(#${id('au-green')})`}
            style={{
              animation: `au-ray ${r.dur}s ease-in-out ${r.delay}s infinite`,
              transformBox: 'fill-box',
              transformOrigin: 'center top',
            }}
          />
        ))}
      </g>

      {/* Low ridge so the aurora dominates */}
      <path
        d="M0 700 L0 600 L160 560 L320 610 L470 540 L640 620 L820 560 L1000 630 L1180 560 L1320 610 L1440 580 L1440 700 Z"
        fill="#04101a"
      />
      <g fill="#123040" opacity="0.7">
        <path d="M470 540 L505 568 L435 568 Z" />
        <path d="M1180 560 L1212 586 L1148 586 Z" />
      </g>

      {/* Frozen-lake reflection */}
      <rect y="700" width="1440" height="200" fill={`url(#${id('au-lake')})`} />
      <g filter={`url(#${id('au-soft')})`} opacity="0.4" transform="translate(0,1400) scale(1,-1)">
        {greenRays.slice(0, 24).map((r, i) => (
          <rect
            key={`gr${i}`}
            x={r.x}
            y={520}
            width={r.w}
            height={170}
            rx={r.w / 2}
            fill={`url(#${id('au-green')})`}
          />
        ))}
      </g>
    </svg>
  );
};

// ─── Dusk — layered sunset hills with a low sun ──────────────────────────────
const Dusk: React.FC<SceneProps> = ({ id }) => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id={id('du-sky')} x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#2b1a5e" />
        <stop offset="32%" stopColor="#6e2f7a" />
        <stop offset="58%" stopColor="#c5447a" />
        <stop offset="78%" stopColor="#f47c4a" />
        <stop offset="100%" stopColor="#ffc15e" />
      </linearGradient>
      <radialGradient id={id('du-sunglow')} cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#fff4c4" />
        <stop offset="35%" stopColor="#ffd06a" stopOpacity="0.7" />
        <stop offset="100%" stopColor="#ff9a4a" stopOpacity="0" />
      </radialGradient>
      <linearGradient id={id('du-sun')} x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#fff2c0" />
        <stop offset="100%" stopColor="#ffcf66" />
      </linearGradient>
    </defs>

    <rect width="1440" height="900" fill={`url(#${id('du-sky')})`} />

    {/* Sun */}
    <circle
      cx="720"
      cy="430"
      r="280"
      fill={`url(#${id('du-sunglow')})`}
      style={{
        animation: 'du-glow 7s ease-in-out infinite',
        transformBox: 'fill-box',
        transformOrigin: 'center',
      }}
    />
    <circle cx="720" cy="430" r="115" fill={`url(#${id('du-sun')})`} />

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
const Slate: React.FC<SceneProps> = ({ id }) => (
  <svg {...SVG_PROPS}>
    <defs>
      <linearGradient id={id('sl-sky')} x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#c3d2e6" />
        <stop offset="60%" stopColor="#dde7f2" />
        <stop offset="100%" stopColor="#eef3f8" />
      </linearGradient>
      <radialGradient id={id('sl-sun')} cx="50%" cy="50%" r="50%">
        <stop offset="0%" stopColor="#fff" stopOpacity="0.95" />
        <stop offset="60%" stopColor="#f4f8fc" stopOpacity="0.5" />
        <stop offset="100%" stopColor="#f4f8fc" stopOpacity="0" />
      </radialGradient>
      <filter id={id('sl-mist')} x="-20%" y="-50%" width="140%" height="200%">
        <feGaussianBlur stdDeviation="16" />
      </filter>
    </defs>

    <rect width="1440" height="900" fill={`url(#${id('sl-sky')})`} />

    {/* Soft sun */}
    <circle cx="1020" cy="240" r="190" fill={`url(#${id('sl-sun')})`} />
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
    <rect
      x="-50"
      y="560"
      width="1540"
      height="60"
      fill="#eef3f8"
      filter={`url(#${id('sl-mist')})`}
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
    <rect
      x="-50"
      y="740"
      width="1540"
      height="50"
      fill="#dde7f2"
      filter={`url(#${id('sl-mist')})`}
      opacity="0.8"
      style={{ animation: 'sl-mist 20s ease-in-out 2s infinite' }}
    />
    <path d="M0 840 L360 770 L760 850 L1120 760 L1440 820 L1440 900 L0 900 Z" fill="#1f2d49" />
  </svg>
);

// ─── Forest — misty layered hills at dawn with soft treelines ────────────────
// Hills flow as smooth curves; treelines are rounded canopy bumps (arcs), not
// spikes; mist bands between layers add depth like the Misty Peaks scene.
const Forest: React.FC<SceneProps> = ({ id }) => {
  // A ridge with soft rounded tree-canopy bumps along the top.
  const treeRidge = (seed: number, baseY: number, amp: number): string => {
    const r = rng(seed);
    let d = `M0 900 L0 ${baseY}`;
    let x = 0;
    while (x < 1440) {
      const w = Math.min(26 + r() * 30, 1440 - x);
      const h = 14 + r() * amp;
      const my = baseY - Math.sin(x * 0.004) * 28;
      d += ` Q ${x + w / 2} ${my - h} ${x + w} ${my}`;
      x += w;
    }
    d += ` L1440 900 Z`;
    return d;
  };
  return (
    <svg {...SVG_PROPS}>
      <defs>
        <linearGradient id={id('fo-sky')} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#ffd29a" />
          <stop offset="38%" stopColor="#ffe7c4" />
          <stop offset="100%" stopColor="#e2f1e6" />
        </linearGradient>
        <radialGradient id={id('fo-sun')} cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#fff7dc" />
          <stop offset="42%" stopColor="#ffe39a" stopOpacity="0.7" />
          <stop offset="100%" stopColor="#ffe39a" stopOpacity="0" />
        </radialGradient>
        <filter id={id('fo-mist')} x="-20%" y="-60%" width="140%" height="220%">
          <feGaussianBlur stdDeviation="18" />
        </filter>
      </defs>

      <rect width="1440" height="900" fill={`url(#${id('fo-sky')})`} />

      {/* Sun + glow */}
      <circle
        cx="520"
        cy="250"
        r="250"
        fill={`url(#${id('fo-sun')})`}
        style={{ animation: 'fo-sun 6s ease-in-out infinite' }}
      />
      <circle cx="520" cy="250" r="86" fill="#fff4cf" />

      {/* Soft clouds */}
      <g fill="#fff" opacity="0.5" style={{ animation: 'fo-cloud 45s linear infinite alternate' }}>
        <ellipse cx="980" cy="170" rx="120" ry="32" />
        <ellipse cx="1064" cy="152" rx="86" ry="38" />
        <ellipse cx="906" cy="184" rx="74" ry="26" />
      </g>

      {/* Birds */}
      {[
        [780, 220],
        [832, 202],
        [884, 228],
      ].map(([x, y], i) => (
        <path
          key={i}
          d={`M${x} ${y} Q ${x + 12} ${y - 10} ${x + 24} ${y} Q ${x + 36} ${y - 10} ${x + 48} ${y}`}
          fill="none"
          stroke="#7a8a5a"
          strokeWidth="2.5"
          strokeLinecap="round"
          opacity="0.5"
          style={{ animation: `fo-bird ${6 + i}s ease-in-out ${i * 0.7}s infinite` }}
        />
      ))}

      {/* Far hills — smooth, hazy */}
      <path d="M0 560 Q 380 500 760 548 T 1440 528 L1440 900 L0 900 Z" fill="#cfe6b4" opacity="0.85" />
      {/* mist band */}
      <rect
        x="-40"
        y="548"
        width="1520"
        height="56"
        fill="#eef6e8"
        filter={`url(#${id('fo-mist')})`}
        style={{ animation: 'fo-mist 18s ease-in-out infinite' }}
      />
      {/* Mid hill with soft treeline */}
      <path d={treeRidge(13, 640, 34)} fill="#86c873" />
      {/* mist band */}
      <rect
        x="-40"
        y="660"
        width="1520"
        height="48"
        fill="#dbeede"
        filter={`url(#${id('fo-mist')})`}
        opacity="0.8"
        style={{ animation: 'fo-mist 22s ease-in-out 2s infinite' }}
      />
      {/* Near hill with denser treeline */}
      <path d={treeRidge(41, 740, 46)} fill="#3f8c45" />
      {/* Foreground ridge */}
      <path d="M0 840 Q 420 795 840 838 T 1440 822 L1440 900 L0 900 Z" fill="#1f5a2b" />

      {/* A couple of soft foreground trees for character */}
      <g fill="#173f20">
        <rect x="196" y="788" width="10" height="60" rx="2" />
        <circle cx="201" cy="776" r="34" />
        <circle cx="176" cy="792" r="22" />
        <circle cx="226" cy="790" r="20" />
        <rect x="1204" y="800" width="9" height="52" rx="2" />
        <circle cx="1208" cy="790" r="28" />
        <circle cx="1188" cy="802" r="18" />
        <circle cx="1228" cy="802" r="17" />
      </g>
    </svg>
  );
};

const WALLPAPER_OVERLAYS: Record<Wallpaper, React.FC<SceneProps>> = {
  midnight: Midnight,
  aurora: Aurora,
  dusk: Dusk,
  slate: Slate,
  forest: Forest,
};

export const WallpaperOverlay: React.FC<{ wallpaper: Wallpaper }> = ({ wallpaper }) => {
  const uid = React.useId().replace(/[^a-zA-Z0-9_-]/g, '');
  const id = React.useCallback((name: string) => `${name}-${uid}`, [uid]);
  const Overlay = WALLPAPER_OVERLAYS[wallpaper];
  return (
    <>
      <style>{KEYFRAMES}</style>
      <Overlay id={id} />
    </>
  );
};
