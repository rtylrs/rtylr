import React from 'react';
import { Wallpaper } from '../../stores/useDesktopStore';

// ─── Isometric helpers ────────────────────────────────────────────────────────
// T=30, origin=(720, 480)
// p(x, y, z) = [720 + (x-y)*30, 480 + (x+y)*15 - z*30]
const px = (x: number, y: number, z: number): [number, number] => [
  720 + (x - y) * 30,
  480 + (x + y) * 15 - z * 30,
];

// Top face polygon (z = gz+h): corners (gx,gy)→(gx+W,gy)→(gx+W,gy+D)→(gx,gy+D)
function topFace(gx: number, gy: number, gz: number, W: number, D: number, h: number): string {
  const [ax, ay] = px(gx, gy, gz + h);
  const [bx, by] = px(gx + W, gy, gz + h);
  const [cx, cy] = px(gx + W, gy + D, gz + h);
  const [dx, dy] = px(gx, gy + D, gz + h);
  return `${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`;
}

// East face polygon (x = gx+W): corners at z=gz and z=gz+h, y from gy to gy+D
function eastFace(gx: number, gy: number, gz: number, W: number, D: number, h: number): string {
  const [ax, ay] = px(gx + W, gy, gz + h);
  const [bx, by] = px(gx + W, gy + D, gz + h);
  const [cx, cy] = px(gx + W, gy + D, gz);
  const [dx, dy] = px(gx + W, gy, gz);
  return `${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`;
}

// South face polygon (y = gy+D): corners at z=gz and z=gz+h, x from gx to gx+W
function southFace(gx: number, gy: number, gz: number, W: number, D: number, h: number): string {
  const [ax, ay] = px(gx, gy + D, gz + h);
  const [bx, by] = px(gx + W, gy + D, gz + h);
  const [cx, cy] = px(gx + W, gy + D, gz);
  const [dx, dy] = px(gx, gy + D, gz);
  return `${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`;
}

interface BuildingProps {
  gx: number;
  gy: number;
  gz?: number;
  W: number;
  D: number;
  h: number;
  topColor: string;
  eastColor: string;
  southColor: string;
  outline?: string;
  // optional window grid
  winRows?: number;
  winCols?: number;
  winColor?: string;
  winOpacity?: number;
}

function Building({
  gx, gy, gz = 0, W, D, h,
  topColor, eastColor, southColor,
  outline = '#0d111e',
  winRows = 0, winCols = 0, winColor = '#ffe080', winOpacity = 0.6,
}: BuildingProps) {
  const stroke = { stroke: outline, strokeWidth: 1, strokeLinejoin: 'round' as const };

  // Window parallelograms on east face
  const eastWins: React.ReactNode[] = [];
  if (winRows > 0 && winCols > 0) {
    for (let row = 0; row < winRows; row++) {
      for (let col = 0; col < winCols; col++) {
        const yStep = D / (winCols + 1);
        const zStep = h / (winRows + 1);
        const wy = gy + D - yStep * (col + 1);
        const wz = gz + zStep * (row + 1);
        const wW = 0;
        const wD = yStep * 0.45;
        const wH = zStep * 0.5;
        const [ax, ay] = px(gx + W, wy, wz + wH);
        const [bx, by] = px(gx + W, wy + wD, wz + wH);
        const [cx, cy] = px(gx + W, wy + wD, wz);
        const [dx, dy] = px(gx + W, wy, wz);
        eastWins.push(
          <polygon key={`ew-${row}-${col}`} points={`${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`}
            fill={winColor} opacity={winOpacity} stroke={outline} strokeWidth={0.5} strokeLinejoin="round" />
        );
      }
    }
  }

  // Window parallelograms on south face
  const southWins: React.ReactNode[] = [];
  if (winRows > 0 && winCols > 0) {
    for (let row = 0; row < winRows; row++) {
      for (let col = 0; col < winCols; col++) {
        const xStep = W / (winCols + 1);
        const zStep = h / (winRows + 1);
        const wx = gx + xStep * (col + 1);
        const wz = gz + zStep * (row + 1);
        const wW = xStep * 0.45;
        const wH = zStep * 0.5;
        const [ax, ay] = px(wx, gy + D, wz + wH);
        const [bx, by] = px(wx + wW, gy + D, wz + wH);
        const [cx, cy] = px(wx + wW, gy + D, wz);
        const [dx, dy] = px(wx, gy + D, wz);
        southWins.push(
          <polygon key={`sw-${row}-${col}`} points={`${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`}
            fill={winColor} opacity={winOpacity * 0.8} stroke={outline} strokeWidth={0.5} strokeLinejoin="round" />
        );
      }
    }
  }

  return (
    <g>
      <polygon points={southFace(gx, gy, gz, W, D, h)} fill={southColor} {...stroke} />
      <polygon points={eastFace(gx, gy, gz, W, D, h)} fill={eastColor} {...stroke} />
      <polygon points={topFace(gx, gy, gz, W, D, h)} fill={topColor} {...stroke} />
      {eastWins}
      {southWins}
    </g>
  );
}

// Ground tile parallelogram
function GroundTile({ gx, gy, W, D, fill }: { gx: number; gy: number; W: number; D: number; fill: string }) {
  const [ax, ay] = px(gx, gy, 0);
  const [bx, by] = px(gx + W, gy, 0);
  const [cx, cy] = px(gx + W, gy + D, 0);
  const [dx, dy] = px(gx, gy + D, 0);
  return <polygon points={`${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`} fill={fill} stroke="#0d111e" strokeWidth={0.5} strokeLinejoin="round" />;
}

// Tree: trunk + 3 layered canopy tiers
function IsoTree({ gx, gy, gz = 0, colors = ['#3aaa50', '#2a8838', '#1e6828'] }: { gx: number; gy: number; gz?: number; colors?: string[] }) {
  return (
    <g>
      <Building gx={gx + 0.15} gy={gy + 0.15} gz={gz} W={0.7} D={0.7} h={1.2} topColor="#5a3010" eastColor="#4a2208" southColor="#3e1c06" />
      {/* tier 1 - wide base */}
      <polygon points={topFace(gx - 0.4, gy - 0.4, gz, 1.8, 1.8, 1.2 + 0.8)} fill={colors[2]} stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
      {/* tier 2 */}
      <polygon points={topFace(gx - 0.2, gy - 0.2, gz, 1.4, 1.4, 1.2 + 1.5)} fill={colors[1]} stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
      {/* tier 3 top */}
      <polygon points={topFace(gx, gy, gz, 1.0, 1.0, 1.2 + 2.1)} fill={colors[0]} stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
    </g>
  );
}

// Simple isometric person silhouette
function IsoPerson({ gx, gy, gz = 0, bodyColor, skinColor, hatColor }: { gx: number; gy: number; gz?: number; bodyColor: string; skinColor: string; hatColor?: string }) {
  const [hx, hy] = px(gx, gy, gz + 1.8);
  const [fx, fy] = px(gx, gy, gz);
  // body
  const bpts = topFace(gx - 0.15, gy - 0.15, gz, 0.3, 0.3, 1.4);
  const bE = eastFace(gx - 0.15, gy - 0.15, gz, 0.3, 0.3, 1.4);
  const bS = southFace(gx - 0.15, gy - 0.15, gz, 0.3, 0.3, 1.4);
  return (
    <g>
      <polygon points={bS} fill={bodyColor} stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
      <polygon points={bE} fill={bodyColor} stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
      <polygon points={bpts} fill={bodyColor} stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
      <circle cx={hx} cy={hy - 12} r={9} fill={skinColor} stroke="#0d111e" strokeWidth={0.7} />
      {hatColor && <ellipse cx={hx} cy={hy - 20} rx={10} ry={5} fill={hatColor} stroke="#0d111e" strokeWidth={0.7} />}
    </g>
  );
}

// ─── Midnight: Night City ─────────────────────────────────────────────────────
const Midnight: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="mn-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#030818" />
        <stop offset="100%" stopColor="#0a1632" />
      </linearGradient>
      <filter id="mn-glow">
        <feGaussianBlur stdDeviation="4" result="b" />
        <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
      </filter>
      <filter id="mn-blur">
        <feGaussianBlur stdDeviation="12" />
      </filter>
      <style>{`
        @keyframes mn-twinkle { 0%,100%{opacity:0.9} 50%{opacity:0.1} }
        @keyframes mn-neon { 0%,85%,100%{opacity:1} 87%{opacity:0.2} 89%{opacity:1} 91%{opacity:0.5} 93%{opacity:1} }
        @keyframes mn-car1 { from{transform:translateX(0)} to{transform:translateX(-420px)} }
        @keyframes mn-car2 { from{transform:translateX(0)} to{transform:translateX(380px)} }
        @keyframes mn-bob  { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-3px)} }
      `}</style>
    </defs>

    {/* Sky */}
    <rect width="1440" height="900" fill="url(#mn-sky)" />

    {/* Moon */}
    <circle cx="1260" cy="95" r="70" fill="#f8f0d8" opacity="0.12" filter="url(#mn-blur)" />
    <circle cx="1260" cy="95" r="44" fill="#f0e8c8" stroke="#0d111e" strokeWidth="1" />
    <circle cx="1247" cy="83" r="7" fill="#e0d4a8" opacity="0.65" />
    <circle cx="1272" cy="106" r="4.5" fill="#e0d4a8" opacity="0.5" />
    <circle cx="1278" cy="82" r="3" fill="#e0d4a8" opacity="0.4" />

    {/* Stars */}
    {(
      [
        [58,22,1.3],[148,12,0.9],[295,48,1.5],[420,18,0.8],[550,42,1.2],
        [682,16,1.4],[810,55,0.9],[945,28,1.2],[1075,46,0.8],[1198,19,1.1],
        [1350,40,1.0],[88,92,0.8],[218,80,1.1],[365,108,0.9],[500,74,1.3],
        [635,100,0.8],[775,88,1.4],[902,118,0.9],[1032,82,1.1],[1162,105,0.8],
        [1305,90,1.2],[42,148,1.0],[188,162,0.8],[332,138,1.2],[475,172,0.9],
        [618,155,1.1],[748,140,0.8],[885,168,1.3],[1018,148,0.9],[1152,162,0.8],
        [1290,145,1.1],[115,215,0.7],[255,232,1.0],[395,202,0.8],[530,220,1.1],
        [675,235,0.7],[815,208,1.2],[952,225,0.8],[1092,212,1.0],[1232,238,0.7],
      ] as [number,number,number][]
    ).map(([cx,cy,r],i) => (
      <circle key={`mn-s${i}`} cx={cx} cy={cy} r={r} fill="white"
        opacity={0.35 + (i % 4) * 0.12}
        style={i % 5 === 0 ? { animation: `mn-twinkle ${1.4 + (i % 4) * 0.8}s ease-in-out ${(i % 3) * 0.7}s infinite` } : undefined} />
    ))}

    {/* Ground tiles - road + sidewalk */}
    {Array.from({ length: 16 }).map((_, i) =>
      Array.from({ length: 8 }).map((_, j) => {
        const roadY = j >= 4;
        return (
          <GroundTile key={`mn-gt-${i}-${j}`} gx={i * 1 - 4} gy={j * 1 + 4} W={1} D={1}
            fill={roadY ? (j === 4 || j === 7 ? '#1e2438' : '#141828') : '#1c2438'} />
        );
      })
    )}
    {/* Road markings */}
    {Array.from({ length: 8 }).map((_, i) => {
      const [ax, ay] = px(-4 + i * 2, 6.5, 0);
      const [bx, by] = px(-4 + i * 2 + 0.8, 6.5, 0);
      const [cx, cy] = px(-4 + i * 2 + 0.8, 7.5, 0);
      const [dx, dy] = px(-4 + i * 2, 7.5, 0);
      return <polygon key={`mn-rm-${i}`} points={`${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`} fill="#2a3050" stroke="none" />;
    })}

    {/* Buildings - back to front (painter's algorithm) */}

    {/* Tall tower (6,1 h=8) */}
    <Building gx={6} gy={1} gz={0} W={2} D={2} h={8}
      topColor="#1a2848" eastColor="#0e1830" southColor="#0a1228"
      winRows={6} winCols={2} winColor="#ffc840" winOpacity={0.55} />
    {/* Neon sign on tower east */}
    {(() => {
      const [sx, sy] = px(8, 2, 5.5);
      return (
        <g style={{ animation: 'mn-neon 5s ease-in-out 0.5s infinite' }}>
          <rect x={sx - 28} y={sy - 10} width={56} height={18} fill="#3366ff" rx="3" filter="url(#mn-glow)" />
          <text x={sx} y={sy + 4} textAnchor="middle" fill="white" fontSize="10" fontWeight="700" fontFamily="system-ui,sans-serif">TOWER</text>
        </g>
      );
    })()}

    {/* Apartment (3,1 h=6) */}
    <Building gx={3} gy={1} gz={0} W={2} D={2} h={6}
      topColor="#1e2e50" eastColor="#121e38" southColor="#0e1830"
      winRows={4} winCols={2} winColor="#ffe060" winOpacity={0.5} />

    {/* Left building (-1,2 size 2x3 h=5) */}
    <Building gx={-1} gy={2} gz={0} W={2} D={3} h={5}
      topColor="#182240" eastColor="#0c1428" southColor="#081020"
      winRows={4} winCols={2} winColor="#ffaa30" winOpacity={0.45} />

    {/* Rtylr shop (1,4 size 2x2 h=3) */}
    <Building gx={1} gy={4} gz={0} W={2} D={2} h={3}
      topColor="#1e3258" eastColor="#122040" southColor="#0c1830"
      winRows={2} winCols={2} winColor="#ffd060" winOpacity={0.6} />
    {/* Awning on south face */}
    {(() => {
      const [ax, ay] = px(1, 6, 2.2);
      const [bx, by] = px(3, 6, 2.2);
      const [cx, cy] = px(3, 6, 1.8);
      const [dx, dy] = px(1, 6, 1.8);
      return (
        <polygon points={`${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`}
          fill="#e05520" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
      );
    })()}
    {/* Neon RTYLR sign on south face */}
    {(() => {
      const [sx, sy] = px(2, 6, 2.8);
      return (
        <g style={{ animation: 'mn-neon 7s ease-in-out 1s infinite' }}>
          <rect x={sx - 32} y={sy - 12} width={64} height={22} fill="#ff6030" rx="4" filter="url(#mn-glow)" />
          <text x={sx} y={sy + 4} textAnchor="middle" fill="white" fontSize="13" fontWeight="800" fontFamily="system-ui,sans-serif" letterSpacing="1">Rtylr</text>
        </g>
      );
    })()}
    {/* OPEN 24H sign */}
    {(() => {
      const [sx, sy] = px(2, 6, 1.2);
      return (
        <g>
          <rect x={sx - 24} y={sy - 8} width={48} height={14} fill="#00aa66" rx="3" filter="url(#mn-glow)"
            style={{ animation: 'mn-neon 3s ease-in-out 2s infinite' }} />
          <text x={sx} y={sy + 3} textAnchor="middle" fill="white" fontSize="8" fontWeight="600" fontFamily="system-ui,sans-serif">OPEN 24H</text>
        </g>
      );
    })()}

    {/* Small foreground building (5,5 h=3) */}
    <Building gx={5} gy={5} gz={0} W={1} D={1} h={3}
      topColor="#1a2848" eastColor="#10183a" southColor="#0c1230"
      winRows={2} winCols={1} winColor="#ffcc40" winOpacity={0.5} />

    {/* Street lights */}
    {[[-1.5, 3.5], [2.5, 3.5], [5.5, 3.5]].map(([lx, ly], i) => {
      const [bx, by] = px(lx, ly, 0);
      const [tx, ty] = px(lx, ly, 2.5);
      const [hx, hy] = px(lx + 0.4, ly, 2.5);
      const [ex, ey] = px(lx + 0.4, ly, 2.3);
      return (
        <g key={`mn-lamp${i}`}>
          <line x1={bx} y1={by} x2={tx} y2={ty} stroke="#2a3858" strokeWidth={3} />
          <line x1={tx} y1={ty} x2={hx} y2={hy} stroke="#2a3858" strokeWidth={2.5} />
          <circle cx={hx} cy={hy} r={6} fill="#ffe050" filter="url(#mn-glow)" stroke="#0d111e" strokeWidth={0.5} />
          <ellipse cx={hx} cy={hy + 18} rx={30} ry={8} fill="#ffee80" opacity="0.08" filter="url(#mn-blur)" />
        </g>
      );
    })}

    {/* Neon signs floating */}
    {(() => {
      const [sx, sy] = px(-0.5, 2, 3.5);
      return (
        <g style={{ animation: 'mn-neon 4s ease-in-out 1.5s infinite' }}>
          <rect x={sx - 22} y={sy - 8} width={44} height={14} fill="#cc22aa" rx="3" filter="url(#mn-glow)" />
          <text x={sx} y={sy + 3} textAnchor="middle" fill="white" fontSize="8" fontWeight="600" fontFamily="system-ui,sans-serif">NOODLES</text>
        </g>
      );
    })()}
    {(() => {
      const [sx, sy] = px(3.5, 1, 4.5);
      return (
        <g style={{ animation: 'mn-neon 6s ease-in-out 3s infinite' }}>
          <rect x={sx - 20} y={sy - 7} width={40} height={13} fill="#22aaff" rx="3" filter="url(#mn-glow)" />
          <text x={sx} y={sy + 3} textAnchor="middle" fill="white" fontSize="7.5" fontWeight="600" fontFamily="system-ui,sans-serif">24/7 BAR</text>
        </g>
      );
    })()}

    {/* Cars on road */}
    <g style={{ animation: 'mn-car1 18s linear infinite' }}>
      {(() => {
        const [ax, ay] = px(2, 5.2, 0.3);
        const [bx, by] = px(3, 5.2, 0.3);
        const [cx, cy] = px(3, 5.7, 0.3);
        const [dx, dy] = px(2, 5.7, 0.3);
        const [ax2, ay2] = px(2.2, 5.2, 0.55);
        const [bx2, by2] = px(2.8, 5.2, 0.55);
        const [cx2, cy2] = px(2.8, 5.7, 0.55);
        const [dx2, dy2] = px(2.2, 5.7, 0.55);
        const [fx, fy] = px(3, 5.3, 0.15);
        const [gxx, gy2] = px(3, 5.6, 0.15);
        return (
          <g>
            {/* body */}
            <polygon points={`${ax2},${ay2} ${bx2},${by2} ${cx2},${cy2} ${dx2},${dy2}`} fill="#c82020" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={southFace(2, 5.2, 0, 1, 0.5, 0.3)} fill="#a01818" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={eastFace(2, 5.2, 0, 1, 0.5, 0.3)} fill="#b01c1c" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={`${ax},${ay} ${bx},${by} ${cx},${cy} ${dx},${dy}`} fill="#c82020" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            {/* headlights */}
            <circle cx={fx} cy={fy} r={4} fill="#ffffaa" opacity={0.9} filter="url(#mn-glow)" />
            <circle cx={gxx} cy={gy2} r={4} fill="#ffffaa" opacity={0.9} filter="url(#mn-glow)" />
          </g>
        );
      })()}
    </g>
    <g style={{ animation: 'mn-car2 22s linear 5s infinite' }}>
      {(() => {
        const [sx, sy] = px(-2, 6.2, 0.3);
        const bS = southFace(-2, 6.2, 0, 1, 0.5, 0.3);
        const bE = eastFace(-2, 6.2, 0, 1, 0.5, 0.3);
        const bT = topFace(-2, 6.2, 0, 1, 0.5, 0.3);
        const bCT = topFace(-1.8, 6.2, 0.3, 0.6, 0.5, 0.25);
        const [lx, ly] = px(-2, 6.3, 0.15);
        const [lx2, ly2] = px(-2, 6.6, 0.15);
        return (
          <g>
            <polygon points={bS} fill="#1a3a88" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={bE} fill="#1e4099" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={bT} fill="#2244aa" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={bCT} fill="#2860cc" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <circle cx={lx} cy={ly} r={4} fill="#ff2200" opacity={0.9} filter="url(#mn-glow)" />
            <circle cx={lx2} cy={ly2} r={4} fill="#ff2200" opacity={0.9} filter="url(#mn-glow)" />
          </g>
        );
      })()}
    </g>

    {/* People silhouettes on sidewalk */}
    <g style={{ animation: 'mn-bob 2.2s ease-in-out infinite' }}>
      <IsoPerson gx={0.5} gy={3.5} gz={0} bodyColor="#2255aa" skinColor="#d4a878" hatColor="#1a2030" />
    </g>
    <IsoPerson gx={2.5} gy={3.8} gz={0} bodyColor="#cc3322" skinColor="#c89060" />
    <g style={{ animation: 'mn-bob 2.8s ease-in-out 0.5s infinite' }}>
      <IsoPerson gx={4.5} gy={4.2} gz={0} bodyColor="#338844" skinColor="#e8c090" hatColor="#4a2808" />
    </g>
    <IsoPerson gx={-0.8} gy={3.2} gz={0} bodyColor="#554488" skinColor="#ddb898" />

    {/* Flower pots at shop */}
    <Building gx={0.8} gy={5.9} gz={0} W={0.25} D={0.25} h={0.3} topColor="#7a3010" eastColor="#5a2008" southColor="#4e1c06" />
    <Building gx={2.9} gy={5.9} gz={0} W={0.25} D={0.25} h={0.3} topColor="#7a3010" eastColor="#5a2008" southColor="#4e1c06" />
    {(() => {
      const [cx, cy] = px(0.925, 5.925, 0.3);
      const [cx2, cy2] = px(3.025, 5.925, 0.3);
      return (
        <>
          <circle cx={cx} cy={cy - 4} r={10} fill="#2a6030" stroke="#0d111e" strokeWidth={0.8} />
          <circle cx={cx - 4} cy={cy - 9} r={4} fill="#ff5530" stroke="#0d111e" strokeWidth={0.6} />
          <circle cx={cx + 3} cy={cy - 8} r={3} fill="#ff9960" stroke="#0d111e" strokeWidth={0.6} />
          <circle cx={cx2} cy={cy2 - 4} r={10} fill="#2a6030" stroke="#0d111e" strokeWidth={0.8} />
          <circle cx={cx2 - 3} cy={cy2 - 9} r={4} fill="#ffaa30" stroke="#0d111e" strokeWidth={0.6} />
          <circle cx={cx2 + 4} cy={cy2 - 7} r={3} fill="#ff5530" stroke="#0d111e" strokeWidth={0.6} />
        </>
      );
    })()}
  </svg>
);

// ─── Aurora: Sunny Street Market ─────────────────────────────────────────────
const Aurora: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="au-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#48b0e8" />
        <stop offset="60%" stopColor="#7acaf0" />
        <stop offset="100%" stopColor="#a8dcf5" />
      </linearGradient>
      <filter id="au-blur"><feGaussianBlur stdDeviation="10" /></filter>
      <filter id="au-glow">
        <feGaussianBlur stdDeviation="3" result="b" />
        <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
      </filter>
      <style>{`
        @keyframes au-cloud1 { 0%{transform:translateX(0)} 100%{transform:translateX(-50px)} }
        @keyframes au-cloud2 { 0%{transform:translateX(0)} 100%{transform:translateX(40px)} }
        @keyframes au-flag { 0%,100%{transform:skewX(0deg)} 50%{transform:skewX(-8deg)} }
        @keyframes au-bob { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-4px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#au-sky)" />

    {/* Sun */}
    <circle cx="1330" cy="88" r="90" fill="#fff060" opacity="0.18" filter="url(#au-blur)" />
    <circle cx="1330" cy="88" r="52" fill="#fff380" stroke="#0d111e" strokeWidth="1" />
    {Array.from({ length: 12 }).map((_, i) => {
      const a = (i * 30 * Math.PI) / 180;
      return <line key={`au-sr${i}`} x1={1330 + Math.cos(a) * 60} y1={88 + Math.sin(a) * 60}
        x2={1330 + Math.cos(a) * 82} y2={88 + Math.sin(a) * 82} stroke="#ffe050" strokeWidth={i % 2 === 0 ? 3 : 2} opacity={0.7} />;
    })}

    {/* Clouds */}
    <g style={{ animation: 'au-cloud1 16s ease-in-out infinite alternate' }}>
      <ellipse cx="285" cy="112" rx="82" ry="34" fill="white" opacity={0.92} />
      <ellipse cx="345" cy="95" rx="62" ry="42" fill="white" opacity={0.88} />
      <ellipse cx="225" cy="118" rx="55" ry="28" fill="white" opacity={0.82} />
    </g>
    <g style={{ animation: 'au-cloud2 20s ease-in-out 4s infinite alternate' }}>
      <ellipse cx="805" cy="75" rx="70" ry="29" fill="white" opacity={0.8} />
      <ellipse cx="858" cy="60" rx="54" ry="36" fill="white" opacity={0.75} />
      <ellipse cx="755" cy="80" rx="50" ry="25" fill="white" opacity={0.7} />
    </g>
    <g style={{ animation: 'au-cloud1 22s ease-in-out 8s infinite alternate' }}>
      <ellipse cx="1082" cy="130" rx="65" ry="27" fill="white" opacity={0.72} />
      <ellipse cx="1132" cy="115" rx="50" ry="32" fill="white" opacity={0.68} />
    </g>

    {/* Ground tiles - cobblestone */}
    {Array.from({ length: 14 }).map((_, i) =>
      Array.from({ length: 8 }).map((_, j) => (
        <GroundTile key={`au-gt-${i}-${j}`} gx={i * 1 - 4} gy={j * 1 + 1} W={1} D={1}
          fill={(i + j) % 2 === 0 ? '#c8b888' : '#b8a878'} />
      ))
    )}

    {/* Background buildings */}
    <Building gx={-3} gy={0} gz={0} W={2} D={2} h={4}
      topColor="#e8d8a8" eastColor="#d4c090" southColor="#c4aa78"
      winRows={3} winCols={2} winColor="#88bbff" winOpacity={0.5} />
    <Building gx={8} gy={0} gz={0} W={2} D={2} h={5}
      topColor="#dcc8a0" eastColor="#c8b488" southColor="#b8a070"
      winRows={4} winCols={2} winColor="#88bbff" winOpacity={0.5} />
    <Building gx={-3} gy={-2} gz={0} W={1} D={1} h={3}
      topColor="#e4d4a8" eastColor="#cec090" southColor="#baa878" />

    {/* Bunting rope */}
    <path d="M 80 440 Q 360 415 720 442 Q 1080 415 1370 445" fill="none" stroke="#ccc" strokeWidth={1.5} opacity={0.6} />
    {Array.from({ length: 22 }).map((_, i) => {
      const x = 82 + i * 58;
      const y = 442 + Math.sin(i * 0.6) * 12;
      const colors = ['#e02020', '#ff9910', '#22aa44', '#2255cc', '#cc44aa', '#ff6622', '#ffcc00'];
      return (
        <polygon key={`au-flag${i}`}
          points={`${x},${y} ${x + 16},${y} ${x + 8},${y + 18}`}
          fill={colors[i % colors.length]} opacity={0.9}
          style={{ animation: `au-flag ${1.5 + (i % 3) * 0.5}s ease-in-out ${i * 0.15}s infinite alternate`, transformOrigin: `${x + 8}px ${y}px` }} />
      );
    })}

    {/* Market stall 1 - RED: Fresh Produce */}
    {(() => {
      const sx = -3, sy = 3;
      const [cx, cy] = px(sx + 1, sy + 1, 2.5);
      return (
        <g>
          <Building gx={sx} gy={sy} gz={0} W={2} D={2} h={2}
            topColor="#f0e8c0" eastColor="#dcd2a8" southColor="#c8bc90" />
          {/* Canopy top */}
          <polygon points={topFace(sx - 0.2, sy - 0.2, 0, 2.4, 2.4, 2.1)} fill="#e02828" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
          {/* Fringe east */}
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + 2.2, sy + i * 0.4, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#c01818" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          {/* Fringe south */}
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + i * 0.4, sy + 2.2, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#c01818" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          {/* Sign */}
          <rect x={cx - 40} y={cy - 16} width={80} height={20} fill="#a01010" rx="3" stroke="#0d111e" strokeWidth={0.8} />
          <text x={cx} y={cy - 2} textAnchor="middle" fill="white" fontSize="10" fontWeight="700" fontFamily="system-ui,sans-serif">FRESH MARKET</text>
          {/* Products on table */}
          {[[-0.7, 0.3], [-0.2, 0.8], [0.3, 0.3], [0.8, 0.8]].map(([ox, oy], pi) => {
            const [px2, py2] = px(sx + 1 + ox, sy + 1 + oy, 1.2);
            const cols = ['#ff6020', '#ff9920', '#44cc22', '#ffdd00'];
            return <circle key={pi} cx={px2} cy={py2} r={8} fill={cols[pi]} stroke="#0d111e" strokeWidth={0.7} />;
          })}
        </g>
      );
    })()}

    {/* Market stall 2 - ORANGE: Spices */}
    {(() => {
      const sx = 1, sy = 3;
      const [cx, cy] = px(sx + 1, sy + 1, 2.5);
      return (
        <g>
          <Building gx={sx} gy={sy} gz={0} W={2} D={2} h={2}
            topColor="#f0e8c0" eastColor="#dcd2a8" southColor="#c8bc90" />
          <polygon points={topFace(sx - 0.2, sy - 0.2, 0, 2.4, 2.4, 2.1)} fill="#ff8810" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + 2.2, sy + i * 0.4, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#dd6608" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + i * 0.4, sy + 2.2, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#dd6608" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          <rect x={cx - 36} y={cy - 16} width={72} height={20} fill="#bb5500" rx="3" stroke="#0d111e" strokeWidth={0.8} />
          <text x={cx} y={cy - 2} textAnchor="middle" fill="white" fontSize="10" fontWeight="700" fontFamily="system-ui,sans-serif">SPICE BAZAAR</text>
          {[[-0.6, 0.5], [0.1, 0.3], [0.7, 0.6]].map(([ox, oy], pi) => {
            const [px2, py2] = px(sx + 1 + ox, sy + 1 + oy, 1.2);
            const cols = ['#ffaa30', '#cc4422', '#ffdd00'];
            return <ellipse key={pi} cx={px2} cy={py2} rx={10} ry={7} fill={cols[pi]} stroke="#0d111e" strokeWidth={0.7} />;
          })}
        </g>
      );
    })()}

    {/* Market stall 3 - GREEN: Organic */}
    {(() => {
      const sx = 5, sy = 3;
      const [cx, cy] = px(sx + 1, sy + 1, 2.5);
      return (
        <g>
          <Building gx={sx} gy={sy} gz={0} W={2} D={2} h={2}
            topColor="#f0e8c0" eastColor="#dcd2a8" southColor="#c8bc90" />
          <polygon points={topFace(sx - 0.2, sy - 0.2, 0, 2.4, 2.4, 2.1)} fill="#22aa44" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + 2.2, sy + i * 0.4, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#189833" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + i * 0.4, sy + 2.2, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#189833" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          <rect x={cx - 44} y={cy - 16} width={88} height={20} fill="#148828" rx="3" stroke="#0d111e" strokeWidth={0.8} />
          <text x={cx} y={cy - 2} textAnchor="middle" fill="white" fontSize="10" fontWeight="700" fontFamily="system-ui,sans-serif">ORGANIC GOODS</text>
          {[[-0.7, 0.4], [0, 0.7], [0.6, 0.3], [0.8, 0.9]].map(([ox, oy], pi) => {
            const [px2, py2] = px(sx + 1 + ox, sy + 1 + oy, 1.2);
            const cols = ['#55cc22', '#ff6622', '#ffdd00', '#44aaff'];
            return <circle key={pi} cx={px2} cy={py2} r={9} fill={cols[pi]} stroke="#0d111e" strokeWidth={0.7} />;
          })}
        </g>
      );
    })()}

    {/* Market stall 4 - BLUE: Crafts */}
    {(() => {
      const sx = 2.5, sy = 5.5;
      const [cx, cy] = px(sx + 1, sy + 1, 2.5);
      return (
        <g>
          <Building gx={sx} gy={sy} gz={0} W={2} D={2} h={2}
            topColor="#f0e8c0" eastColor="#dcd2a8" southColor="#c8bc90" />
          <polygon points={topFace(sx - 0.2, sy - 0.2, 0, 2.4, 2.4, 2.1)} fill="#2255cc" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + 2.2, sy + i * 0.4, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#1844aa" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          {Array.from({ length: 5 }).map((_, i) => {
            const [fx, fy] = px(sx + i * 0.4, sy + 2.2, 2.1);
            return <circle key={i} cx={fx} cy={fy} r={5} fill="#1844aa" stroke="#0d111e" strokeWidth={0.5} />;
          })}
          <rect x={cx - 36} y={cy - 16} width={72} height={20} fill="#1438aa" rx="3" stroke="#0d111e" strokeWidth={0.8} />
          <text x={cx} y={cy - 2} textAnchor="middle" fill="white" fontSize="10" fontWeight="700" fontFamily="system-ui,sans-serif">ARTISAN CRAFTS</text>
          {[[-0.5, 0.5], [0.3, 0.3], [0.7, 0.7]].map(([ox, oy], pi) => {
            const [px2, py2] = px(sx + 1 + ox, sy + 1 + oy, 1.2);
            const cols = ['#ff8844', '#44bbff', '#ffcc33'];
            return <rect key={pi} x={px2 - 9} y={py2 - 7} width={18} height={14} fill={cols[pi]} stroke="#0d111e" strokeWidth={0.7} rx={2} />;
          })}
        </g>
      );
    })()}

    {/* Trees */}
    <IsoTree gx={-2.5} gy={2} gz={0} colors={['#44bb44', '#339933', '#226622']} />
    <IsoTree gx={9} gy={2.5} gz={0} colors={['#3aaa50', '#2a8838', '#1e6828']} />
    <IsoTree gx={-1} gy={7} gz={0} colors={['#55cc44', '#44aa33', '#2d7a22']} />

    {/* Lamp post in square center */}
    {(() => {
      const [bx, by] = px(4.5, 4.5, 0);
      const [tx, ty] = px(4.5, 4.5, 3);
      const [hx, hy] = px(4.9, 4.5, 3);
      return (
        <g>
          <line x1={bx} y1={by} x2={tx} y2={ty} stroke="#8a7050" strokeWidth={4} />
          <line x1={tx} y1={ty} x2={hx} y2={hy} stroke="#8a7050" strokeWidth={3} />
          <circle cx={hx} cy={hy} r={7} fill="#ffe898" filter="url(#au-glow)" stroke="#0d111e" strokeWidth={0.8} />
        </g>
      );
    })()}

    {/* Vendors and customers */}
    <g style={{ animation: 'au-bob 2.5s ease-in-out infinite' }}>
      <IsoPerson gx={-2} gy={3.5} gz={0} bodyColor="#e02020" skinColor="#d4a878" />
    </g>
    <IsoPerson gx={0} gy={4.2} gz={0} bodyColor="#4488ff" skinColor="#e8c8a0" hatColor="#3a2808" />
    <g style={{ animation: 'au-bob 3s ease-in-out 0.5s infinite' }}>
      <IsoPerson gx={4} gy={3.5} gz={0} bodyColor="#22aa44" skinColor="#c8986a" />
    </g>
    <IsoPerson gx={6} gy={4.3} gz={0} bodyColor="#aa44cc" skinColor="#e0b888" />
    <IsoPerson gx={3.5} gy={5.8} gz={0} bodyColor="#ff6622" skinColor="#d4a070" hatColor="#1a2030" />
    {/* Child customer */}
    {(() => {
      const [hx, hy] = px(1.5, 5.0, 1.1);
      const bS = southFace(1.38, 4.88, 0, 0.24, 0.24, 1);
      const bE = eastFace(1.38, 4.88, 0, 0.24, 0.24, 1);
      const bT = topFace(1.38, 4.88, 0, 0.24, 0.24, 1);
      return (
        <g>
          <polygon points={bS} fill="#ff8844" stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <polygon points={bE} fill="#ff8844" stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <polygon points={bT} fill="#ff8844" stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <circle cx={hx} cy={hy - 10} r={7} fill="#e8c8a0" stroke="#0d111e" strokeWidth={0.7} />
        </g>
      );
    })()}

    {/* Ground decoration: flowers */}
    {[[-3.5, 2.5], [-3.5, 8.5], [9.5, 2.5], [9.5, 8.5]].map(([fx, fy], i) => {
      const [cx2, cy2] = px(fx, fy, 0);
      return (
        <g key={`au-fl${i}`}>
          <circle cx={cx2} cy={cy2} r={7} fill={i % 2 === 0 ? '#ffdd22' : '#ff6688'} stroke="#0d111e" strokeWidth={0.7} />
          <circle cx={cx2} cy={cy2} r={3} fill="white" opacity={0.7} />
        </g>
      );
    })}

    {/* Basket */}
    {(() => {
      const [bx, by] = px(0.2, 4.5, 0.4);
      return (
        <g>
          <ellipse cx={bx} cy={by} rx={14} ry={9} fill="#c4903a" stroke="#0d111e" strokeWidth={0.8} />
          <path d={`M ${bx - 14} ${by} Q ${bx} ${by - 18} ${bx + 14} ${by}`} fill="none" stroke="#a07028" strokeWidth={1.5} />
        </g>
      );
    })()}
  </svg>
);

// ─── Dusk: Coastal Harbor ────────────────────────────────────────────────────
const Dusk: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="du-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#1a1048" />
        <stop offset="28%" stopColor="#6a2a90" />
        <stop offset="55%" stopColor="#e85520" />
        <stop offset="75%" stopColor="#f59828" />
        <stop offset="100%" stopColor="#ffc850" />
      </linearGradient>
      <linearGradient id="du-water" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#d07020" />
        <stop offset="40%" stopColor="#a04e18" />
        <stop offset="100%" stopColor="#1a3860" />
      </linearGradient>
      <filter id="du-blur"><feGaussianBlur stdDeviation="12" /></filter>
      <filter id="du-glow">
        <feGaussianBlur stdDeviation="4" result="b" />
        <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
      </filter>
      <style>{`
        @keyframes du-ripple { 0%,100%{opacity:0.15} 50%{opacity:0.35} }
        @keyframes du-sail { 0%,100%{transform:rotate(-2.5deg)} 50%{transform:rotate(2.5deg)} }
        @keyframes du-bird { 0%{transform:translate(0,0)} 50%{transform:translate(32px,-10px)} 100%{transform:translate(64px,0)} }
        @keyframes du-bob { 0%,100%{transform:translateY(0)} 50%{transform:translateY(-3px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#du-sky)" />

    {/* Sun on horizon */}
    <circle cx="720" cy="488" r="120" fill="#ffe880" opacity="0.14" filter="url(#du-blur)" />
    <circle cx="720" cy="488" r="64" fill="#fff380" stroke="#0d111e" strokeWidth="1" />
    <path d="M 658 528 Q 720 515 782 528 Q 762 560 720 572 Q 678 560 658 528 Z" fill="#ffdd60" opacity="0.35" />

    {/* Stars in upper sky */}
    {([[80,45],[182,28],[315,52],[455,30],[582,58],[200,88],[382,72],[528,95]] as [number,number][]).map(([cx,cy],i) => (
      <circle key={`du-star${i}`} cx={cx} cy={cy} r={1.2} fill="white" opacity={0.6} />
    ))}

    {/* Sunset clouds */}
    <ellipse cx="222" cy="205" rx="98" ry="33" fill="#ff8855" opacity="0.5" />
    <ellipse cx="288" cy="188" rx="78" ry="44" fill="#ff6a40" opacity="0.45" />
    <ellipse cx="158" cy="215" rx="67" ry="27" fill="#ff8855" opacity="0.4" />
    <ellipse cx="1055" cy="168" rx="88" ry="30" fill="#ff8855" opacity="0.48" />
    <ellipse cx="1120" cy="152" rx="70" ry="40" fill="#e07040" opacity="0.42" />

    {/* Water */}
    <rect x="0" y="525" width="1440" height="375" fill="url(#du-water)" />
    {[548,570,595,623,655].map((wy,i) => (
      <path key={`du-wv${i}`}
        d={`M 0 ${wy} Q ${180+i*22} ${wy-7} ${360} ${wy} Q ${540} ${wy+7} 720 ${wy} Q ${900} ${wy-7} 1080 ${wy} Q ${1260} ${wy+7} 1440 ${wy}`}
        fill="none" stroke={`rgba(255,${135-i*18},25,${0.25-i*0.03})`} strokeWidth={1.5}
        style={{ animation: `du-ripple ${2+i*0.45}s ease-in-out ${i*0.3}s infinite` }} />
    ))}
    {/* Sun reflection */}
    {Array.from({ length: 8 }).map((_,i) => (
      <ellipse key={`du-ref${i}`} cx={720} cy={542+i*24} rx={18+i*14} ry={4} fill="#ffcc40" opacity={0.18-i*0.018} />
    ))}

    {/* Isometric ground/dock area */}
    {Array.from({ length: 14 }).map((_,i) =>
      Array.from({ length: 4 }).map((_,j) => (
        <GroundTile key={`du-gt-${i}-${j}`} gx={i-5} gy={j+2} W={1} D={1}
          fill={(i+j)%3===0 ? '#c8b88a' : (i+j)%3===1 ? '#b8a878' : '#ccc0a0'} />
      ))
    )}

    {/* Mediterranean buildings */}
    {/* Building 1 left */}
    <Building gx={-4} gy={0} gz={0} W={2} D={2} h={4}
      topColor="#f8f0e0" eastColor="#e8dbc8" southColor="#d8c8b0"
      winRows={3} winCols={2} winColor="#ff9838" winOpacity={0.55} />
    <Building gx={-4} gy={-2} gz={0} W={1.5} D={1.5} h={3}
      topColor="#f0e8d8" eastColor="#e0d4bc" southColor="#d0c0a4"
      winRows={2} winCols={1} winColor="#ff8828" winOpacity={0.5} />

    {/* Building 2 mid-left */}
    <Building gx={0} gy={0} gz={0} W={2} D={1.5} h={3.5}
      topColor="#f5ede0" eastColor="#e5d8c8" southColor="#d5c8b0"
      winRows={2} winCols={2} winColor="#ff9940" winOpacity={0.5} />
    {/* Arch window */}
    {(() => {
      const [ax, ay] = px(0, 1.5, 1.5);
      return (
        <g>
          <rect x={ax - 12} y={ay - 18} width={24} height={30} fill="#e08830" opacity={0.55} rx="12 12 0 0" stroke="#0d111e" strokeWidth={0.8} />
        </g>
      );
    })()}

    {/* Building 3 right */}
    <Building gx={6} gy={0} gz={0} W={2} D={2} h={4}
      topColor="#f8f0e0" eastColor="#e8dbc8" southColor="#d8c8b0"
      winRows={3} winCols={2} winColor="#ff9838" winOpacity={0.55} />
    {/* Dome / blue roof */}
    {(() => {
      const [cx, cy] = px(7, 0, 4);
      return (
        <ellipse cx={cx} cy={cy - 15} rx={35} ry={18} fill="#5588cc" stroke="#0d111e" strokeWidth={1} />
      );
    })()}

    {/* Building 4 far right */}
    <Building gx={8} gy={-1} gz={0} W={1.5} D={1.5} h={5}
      topColor="#f0e8d8" eastColor="#e0d4c0" southColor="#d0c0a8"
      winRows={4} winCols={1} winColor="#ff9940" winOpacity={0.5} />
    {/* Bell tower top */}
    <Building gx={8.2} gy={-0.8} gz={5} W={1.1} D={1.1} h={1}
      topColor="#f8f0e0" eastColor="#e8dac8" southColor="#d8c8b0" />

    {/* Dock planks */}
    {Array.from({ length: 8 }).map((_,i) => (
      <GroundTile key={`du-dock-${i}`} gx={1+i*0.6} gy={4} W={0.55} D={1} fill={i%2===0?'#9a6028':'#8a5020'} />
    ))}
    {/* Dock posts */}
    {[1.2, 2.4, 3.6, 4.8].map((dx,i) => {
      const [bx, by] = px(dx, 5, 0);
      const [tx, ty] = px(dx, 5, 0.8);
      return (
        <g key={`du-post${i}`}>
          <line x1={bx} y1={by} x2={tx} y2={ty} stroke="#7a4820" strokeWidth={5} />
          <ellipse cx={tx} cy={ty} rx={6} ry={3} fill="#6a3818" stroke="#0d111e" strokeWidth={0.7} />
        </g>
      );
    })}
    {/* Rope between posts */}
    {[1.2,2.4,3.6].map((dx,i) => {
      const [ax,ay] = px(dx, 5, 0.7);
      const [bx,by] = px(dx+1.2, 5, 0.7);
      return <line key={i} x1={ax} y1={ay} x2={bx} y2={by} stroke="#c09060" strokeWidth={1.5} opacity={0.7} />;
    })}

    {/* Sailboat 1 */}
    <g style={{ animation: 'du-sail 5s ease-in-out infinite', transformOrigin: `${px(2,8,0)[0]}px ${px(2,8,0)[1]}px` }}>
      {(() => {
        const hull = southFace(1, 7.5, 0, 1.5, 1, 0.35);
        const hullE = eastFace(1, 7.5, 0, 1.5, 1, 0.35);
        const hullT = topFace(1, 7.5, 0, 1.5, 1, 0.35);
        const [mx, my] = px(2.2, 8, 0.35);
        const [mtx, mty] = px(2.2, 8, 3);
        const [s1ax, s1ay] = px(2.2, 8, 3);
        const [s1bx, s1by] = px(2.2, 9, 0.35);
        const [s1cx, s1cy] = px(2.2, 8.5, 0.35);
        const [s2ax, s2ay] = px(2.2, 8, 3);
        const [s2bx, s2by] = px(1, 8, 0.35);
        const [s2cx, s2cy] = px(1.5, 8, 0.35);
        return (
          <g>
            <polygon points={hull} fill="#e8d5a8" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={hullE} fill="#d8c598" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={hullT} fill="#e8d5a8" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <line x1={mx} y1={my} x2={mtx} y2={mty} stroke="#7a5028" strokeWidth={3} />
            <polygon points={`${s1ax},${s1ay} ${s1bx},${s1by} ${s1cx},${s1cy}`} fill="white" opacity={0.92} stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
            <polygon points={`${s2ax},${s2ay} ${s2bx},${s2by} ${s2cx},${s2cy}`} fill="#ff6622" opacity={0.8} stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
          </g>
        );
      })()}
    </g>

    {/* Sailboat 2 */}
    <g>
      {(() => {
        const hull = southFace(5, 8, 0, 1.2, 0.8, 0.28);
        const hullE = eastFace(5, 8, 0, 1.2, 0.8, 0.28);
        const hullT = topFace(5, 8, 0, 1.2, 0.8, 0.28);
        const [mx, my] = px(5.8, 8.3, 0.28);
        const [mtx, mty] = px(5.8, 8.3, 2.5);
        const [s1ax, s1ay] = px(5.8, 8.3, 2.5);
        const [s1bx, s1by] = px(5.8, 8.9, 0.28);
        const [s1cx, s1cy] = px(5.8, 8.6, 0.28);
        const [s2ax, s2ay] = px(5.8, 8.3, 2.5);
        const [s2bx, s2by] = px(5, 8.3, 0.28);
        const [s2cx, s2cy] = px(5.3, 8.3, 0.28);
        return (
          <g>
            <polygon points={hull} fill="#d8c898" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={hullE} fill="#c8b888" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <polygon points={hullT} fill="#d8c898" stroke="#0d111e" strokeWidth={1} strokeLinejoin="round" />
            <line x1={mx} y1={my} x2={mtx} y2={mty} stroke="#7a5028" strokeWidth={2.5} />
            <polygon points={`${s1ax},${s1ay} ${s1bx},${s1by} ${s1cx},${s1cy}`} fill="white" opacity={0.88} stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
            <polygon points={`${s2ax},${s2ay} ${s2bx},${s2by} ${s2cx},${s2cy}`} fill="#4488cc" opacity={0.78} stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
          </g>
        );
      })()}
    </g>

    {/* Palm trees */}
    {(() => {
      const [tx, ty] = px(-3, 3, 0);
      const [t2x, t2y] = px(9.5, 1, 0);
      const renderPalm = (bx: number, by: number, flip: boolean) => (
        <g>
          <path d={`M ${bx} ${by} Q ${bx + (flip?-8:8)} ${by - 48} ${bx + (flip?-5:5)} ${by - 95}`}
            fill="none" stroke="#6a4820" strokeWidth={8} strokeLinecap="round" />
          {[[-30, -90, 20], [-8, -105, 8], [22, -96, -18], [35, -80, -32], [0, -118, 5]].map(([dx, dy2, rot], i) => (
            <ellipse key={i} cx={bx + (flip ? -dx : dx) + (flip?-5:5)} cy={by + dy2}
              rx={26} ry={7} fill="#2d6820"
              transform={`rotate(${flip ? -rot : rot} ${bx + (flip?-dx:dx) + (flip?-5:5)} ${by + dy2})`}
              stroke="#0d111e" strokeWidth={0.8} opacity={0.9} />
          ))}
        </g>
      );
      return (
        <>
          {renderPalm(tx, ty, false)}
          {renderPalm(t2x, t2y, true)}
        </>
      );
    })()}

    {/* People on dock */}
    <g style={{ animation: 'du-bob 2s ease-in-out infinite' }}>
      <IsoPerson gx={1.5} gy={3.5} gz={0} bodyColor="#2255aa" skinColor="#e8c090" />
    </g>
    <IsoPerson gx={4} gy={3.2} gz={0} bodyColor="#cc4422" skinColor="#d4a878" hatColor="#1a2038" />

    {/* Beach umbrellas */}
    {[[-3.8, 1.5], [9.5, 2]].map(([ux, uy], i) => {
      const [cx, cy] = px(ux, uy, 0);
      return (
        <g key={i}>
          <line x1={cx} y1={cy} x2={cx + 5} y2={cy + 50} stroke="#8a5030" strokeWidth={3} />
          <path d={`M ${cx - 32} ${cy + 2} Q ${cx} ${cy - 22} ${cx + 32} ${cy + 2} Z`}
            fill={i === 0 ? '#e83020' : '#2244aa'} opacity={0.88} stroke="#0d111e" strokeWidth={0.8} />
          <path d={`M ${cx - 15} ${cy + 2} Q ${cx} ${cy - 10} ${cx + 15} ${cy + 2} Z`}
            fill="white" opacity={0.65} />
        </g>
      );
    })}

    {/* Seagulls */}
    {([[225,282],[258,265],[290,278],[908,298],[940,282],[968,295]] as [number,number][]).map(([bx,by],i) => (
      <g key={`du-bird${i}`} style={{ animation: `du-bird ${3+i*0.4}s ease-in-out ${i*0.8}s infinite` }}>
        <path d={`M ${bx-10} ${by} Q ${bx-5} ${by-5} ${bx} ${by}`} fill="none" stroke="white" strokeWidth={1.8} />
        <path d={`M ${bx} ${by} Q ${bx+5} ${by-5} ${bx+10} ${by}`} fill="none" stroke="white" strokeWidth={1.8} />
      </g>
    ))}
  </svg>
);

// ─── Slate: Snowy Town ───────────────────────────────────────────────────────
const Slate: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="sl-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#8898b8" />
        <stop offset="55%" stopColor="#aab8cc" />
        <stop offset="100%" stopColor="#c8d4e0" />
      </linearGradient>
      <filter id="sl-blur"><feGaussianBlur stdDeviation="8" /></filter>
      <filter id="sl-glow">
        <feGaussianBlur stdDeviation="3" result="b" />
        <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
      </filter>
      <style>{`
        @keyframes sl-snow1 { 0%{transform:translate(0,0)} 100%{transform:translate(-22px,900px)} }
        @keyframes sl-snow2 { 0%{transform:translate(0,0)} 100%{transform:translate(28px,900px)} }
        @keyframes sl-smoke { 0%{transform:translateY(0) scaleX(1);opacity:0.45} 100%{transform:translateY(-50px) scaleX(1.8);opacity:0} }
        @keyframes sl-xmas { 0%,100%{opacity:0.9} 50%{opacity:0.3} }
        @keyframes sl-walk { 0%{transform:translateX(0)} 100%{transform:translateX(50px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#sl-sky)" />

    {/* Snowflakes */}
    {Array.from({ length: 44 }).map((_, i) => {
      const x = (i * 101 + 35) % 1440;
      const y = (i * 47 + 15) % 420;
      const s = 0.9 + (i % 4) * 0.4;
      return (
        <g key={`sl-flake${i}`} transform={`translate(${x},${y})`}
          style={{ animation: `${i%2===0?'sl-snow1':'sl-snow2'} ${6+(i%6)}s linear ${(i%8)*0.5}s infinite` }}>
          <circle cx="0" cy="0" r={s} fill="white" opacity={0.75} />
          <line x1={-s*3} y1="0" x2={s*3} y2="0" stroke="white" strokeWidth={s*0.6} opacity={0.5} />
          <line x1="0" y1={-s*3} x2="0" y2={s*3} stroke="white" strokeWidth={s*0.6} opacity={0.5} />
        </g>
      );
    })}

    {/* Ground tiles - snowy */}
    {Array.from({ length: 14 }).map((_,i) =>
      Array.from({ length: 6 }).map((_,j) => (
        <GroundTile key={`sl-gt-${i}-${j}`} gx={i-4} gy={j+3} W={1} D={1}
          fill={(i+j)%2===0 ? '#dce8f4' : '#e8f2fc'} />
      ))
    )}

    {/* Street */}
    {Array.from({ length: 14 }).map((_,i) => (
      <GroundTile key={`sl-rd-${i}`} gx={i-4} gy={4} W={1} D={1}
        fill={(i)%3===0?'#2a3250':'#222a42'} />
    ))}

    {/* Building 1 - far left small */}
    <Building gx={-4} gy={0} gz={0} W={1.5} D={1.5} h={3}
      topColor="#e8f2ff" eastColor="#6a7898" southColor="#5a6888"
      winRows={2} winCols={1} winColor="#ffe898" winOpacity={0.6} />
    {/* Snow on roof */}
    <polygon points={topFace(-4, 0, 0, 1.5, 1.5, 3.05)} fill="#e8f2ff" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
    {/* Chimney */}
    <Building gx={-3.6} gy={0.2} gz={3} W={0.4} D={0.4} h={0.8} topColor="#5a6880" eastColor="#485870" southColor="#3c4c60" />
    {[0,1,2].map(i => (
      <ellipse key={`sl-sm1-${i}`} cx={px(-3.4,0.4,3.8+i*0.3)[0]} cy={px(-3.4,0.4,3.8+i*0.3)[1]-8} rx={5+i*2} ry={3+i}
        fill="#c8d4e4" opacity={0.4} style={{ animation: `sl-smoke ${2.5+i*0.5}s ease-out ${i*0.6}s infinite` }} />
    ))}

    {/* Building 2 - main left */}
    <Building gx={-2.5} gy={0} gz={0} W={2} D={2} h={5}
      topColor="#e8f2ff" eastColor="#788aaa" southColor="#6878a0"
      winRows={4} winCols={2} winColor="#ffe898" winOpacity={0.5} />
    <polygon points={topFace(-2.5, 0, 0, 2, 2, 5.05)} fill="#e8f2ff" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
    {/* Pitched snow roof shape */}
    {(() => {
      const [rx, ry] = px(-1.5, 0, 5.4);
      const [lx, ly] = px(-2.5, 0, 5.05);
      const [ex, ey] = px(-0.5, 0, 5.05);
      const [sx, sy] = px(-2.5, 2, 5.05);
      const [sex, sey] = px(-0.5, 2, 5.05);
      return <polygon points={`${lx},${ly} ${rx},${ry} ${ex},${ey} ${sex},${sey} ${rx},${ry} ${sx},${sy}`}
        fill="#dde8f5" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />;
    })()}
    <Building gx={-1.8} gy={0.2} gz={5} W={0.5} D={0.5} h={0.9} topColor="#5a6880" eastColor="#485870" southColor="#3c4c60" />
    {[0,1,2,3].map(i => (
      <ellipse key={`sl-sm2-${i}`} cx={px(-1.55,0.45,5.9+i*0.28)[0]} cy={px(-1.55,0.45,5.9+i*0.28)[1]-8} rx={5+i*2} ry={3+i}
        fill="#c8d4e4" opacity={0.4} style={{ animation: `sl-smoke ${3+i*0.5}s ease-out ${0.8+i*0.4}s infinite` }} />
    ))}

    {/* Building 3 - bakery center */}
    <Building gx={1} gy={1} gz={0} W={3} D={2} h={4}
      topColor="#e8f2ff" eastColor="#7a8eae" southColor="#6a7ea0"
      winRows={3} winCols={3} winColor="#ffe898" winOpacity={0.55} />
    <polygon points={topFace(1, 1, 0, 3, 2, 4.05)} fill="#e8f2ff" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
    {/* Shop sign */}
    {(() => {
      const [sx, sy] = px(2.5, 3, 2.5);
      return (
        <g>
          <rect x={sx-48} y={sy-14} width={96} height={22} fill="#3a4e6a" rx="4" stroke="#0d111e" strokeWidth={0.8} />
          <rect x={sx-48} y={sy-14} width={96} height={7} fill="#e8f2ff" rx="4 4 0 0" opacity={0.9} />
          <text x={sx} y={sy+4} textAnchor="middle" fill="white" fontSize="11" fontWeight="600" fontFamily="system-ui,sans-serif">BAKERY &amp; CAFÉ</text>
        </g>
      );
    })()}
    {/* Window display - baked goods */}
    {(() => {
      const [wx, wy] = px(1.5, 3, 1.5);
      const [wx2, wy2] = px(3, 3, 1.5);
      return (
        <g>
          <rect x={wx-15} y={wy-10} width={30} height={22} fill="#3a4e6a" rx="2" stroke="#0d111e" strokeWidth={0.8} />
          {[0,1,2].map(i => <circle key={i} cx={wx-8+i*8} cy={wy-2} r={5} fill={['#f4c2a0','#d4a870','#e8b888'][i]} stroke="#0d111e" strokeWidth={0.5} />)}
          <rect x={wx2-15} y={wy2-10} width={30} height={22} fill="#3a4e6a" rx="2" stroke="#0d111e" strokeWidth={0.8} />
          {[0,1].map(i => <ellipse key={i} cx={wx2-6+i*12} cy={wy2-2} rx={7} ry={5} fill={['#ff9988','#d4c888'][i]} stroke="#0d111e" strokeWidth={0.5} />)}
        </g>
      );
    })()}
    <Building gx={2} gy={1.2} gz={4} W={0.5} D={0.5} h={0.9} topColor="#5a6880" eastColor="#485870" southColor="#3c4c60" />
    {[0,1,2,3].map(i => (
      <ellipse key={`sl-sm3-${i}`} cx={px(2.25,1.45,4.9+i*0.28)[0]} cy={px(2.25,1.45,4.9+i*0.28)[1]-8} rx={6+i*2} ry={4+i}
        fill="#c8d4e4" opacity={0.4} style={{ animation: `sl-smoke ${2+i*0.4}s ease-out ${i*0.3}s infinite` }} />
    ))}
    {/* Christmas lights on bakery */}
    {Array.from({ length: 14 }).map((_,i) => {
      const [lx, ly] = px(1 + i * (3/14), 3, 4.06);
      const colors = ['#ff2222','#22ff22','#ffff22','#2222ff','#ff8822'];
      return <circle key={`sl-xmas${i}`} cx={lx} cy={ly} r={4} fill={colors[i%colors.length]} opacity={0.85}
        style={{ animation: `sl-xmas ${0.8+(i%4)*0.3}s ease-in-out ${i*0.1}s infinite` }}
        filter="url(#sl-glow)" />;
    })}

    {/* Building 4 - right residential */}
    <Building gx={5} gy={0} gz={0} W={2} D={2} h={4}
      topColor="#e8f2ff" eastColor="#788aaa" southColor="#6878a0"
      winRows={3} winCols={2} winColor="#ffe898" winOpacity={0.5} />
    <polygon points={topFace(5, 0, 0, 2, 2, 4.05)} fill="#e8f2ff" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
    <Building gx={5.8} gy={0.2} gz={4} W={0.4} D={0.4} h={0.7} topColor="#5a6880" eastColor="#485870" southColor="#3c4c60" />
    {[0,1,2].map(i => (
      <ellipse key={`sl-sm4-${i}`} cx={px(6,0.4,4.7+i*0.28)[0]} cy={px(6,0.4,4.7+i*0.28)[1]-8} rx={4+i*2} ry={3+i}
        fill="#c8d4e4" opacity={0.38} style={{ animation: `sl-smoke ${3.5+i*0.5}s ease-out ${1+i*0.5}s infinite` }} />
    ))}

    {/* Building 5 - far right church */}
    <Building gx={8} gy={-1} gz={0} W={1.5} D={1.5} h={4}
      topColor="#e8f2ff" eastColor="#6a7898" southColor="#5a6888"
      winRows={3} winCols={1} winColor="#ffe898" winOpacity={0.55} />
    <polygon points={topFace(8, -1, 0, 1.5, 1.5, 4.05)} fill="#e8f2ff" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
    {/* Steeple */}
    <Building gx={8.4} gy={-0.6} gz={4} W={0.7} D={0.7} h={2} topColor="#dce8f4" eastColor="#5a6888" southColor="#4a5878" />
    {(() => {
      const [tx, ty] = px(8.75, -0.25, 6.1);
      return <line x1={tx} y1={ty} x2={tx} y2={ty - 22} stroke="#cc2222" strokeWidth={2} />;
    })()}
    <Building gx={8.5} gy={-0.5} gz={6} W={0.5} D={0.5} h={0.7} topColor="#cc2222" eastColor="#aa1818" southColor="#881010" />

    {/* Street lamps */}
    {[[-2, 2.8], [2.5, 2.8], [6.5, 2.8]].map(([lx, ly], i) => {
      const [bx, by] = px(lx, ly, 0);
      const [tx, ty] = px(lx, ly, 2.5);
      const [hx, hy] = px(lx + 0.4, ly, 2.5);
      return (
        <g key={`sl-lamp${i}`}>
          <line x1={bx} y1={by} x2={tx} y2={ty} stroke="#4a5878" strokeWidth={3} />
          <line x1={tx} y1={ty} x2={hx} y2={hy} stroke="#4a5878" strokeWidth={2.5} />
          <circle cx={hx} cy={hy} r={7} fill="#ffe060" filter="url(#sl-glow)" stroke="#0d111e" strokeWidth={0.6} />
          <ellipse cx={hx} cy={hy+20} rx={35} ry={10} fill="#ffe898" opacity={0.15} filter="url(#sl-blur)" />
          {/* Snow cap on lamp */}
          <ellipse cx={hx} cy={hy-7} rx={9} ry={4} fill="#e8f2ff" stroke="#0d111e" strokeWidth={0.5} />
        </g>
      );
    })}

    {/* Bare trees */}
    {[[-3.5, 2.2], [4.8, 2.2]].map(([tx, ty], ti) => {
      const [bx, by] = px(tx, ty, 0);
      return (
        <g key={`sl-tree${ti}`}>
          <line x1={bx} y1={by} x2={bx} y2={by-78} stroke="#5a5a6a" strokeWidth={5} />
          {[[-22,55],[ 20,48],[-14,72],[16,68]].map(([dx,dz],i) => {
            const [lx2, ly2] = px(tx, ty, dz/30);
            return (
              <g key={i}>
                <line x1={lx2} y1={ly2} x2={lx2+dx} y2={ly2+18} stroke="#5a5a6a" strokeWidth={2.5} opacity={0.8} />
                <ellipse cx={lx2+dx} cy={ly2+18} rx={Math.abs(dx)*0.55} ry={4} fill="#dce8f4" opacity={0.8} />
              </g>
            );
          })}
          {/* Snow on branches */}
        </g>
      );
    })}

    {/* Snowman */}
    {(() => {
      const [cx, cy] = px(3.5, 6.5, 0);
      return (
        <g>
          <ellipse cx={cx} cy={cy + 5} rx={40} ry={11} fill="#d8eaf8" stroke="#0d111e" strokeWidth={0.8} />
          <circle cx={cx} cy={cy - 24} r={28} fill="white" stroke="#0d111e" strokeWidth={0.8} />
          <circle cx={cx} cy={cy - 62} r={20} fill="white" stroke="#0d111e" strokeWidth={0.8} />
          <circle cx={cx} cy={cy - 88} r={13} fill="white" stroke="#0d111e" strokeWidth={0.8} />
          {/* Face */}
          <circle cx={cx-5} cy={cy-91} r={2.5} fill="#2a3040" />
          <circle cx={cx+5} cy={cy-91} r={2.5} fill="#2a3040" />
          <path d={`M ${cx-2} ${cy-86} L ${cx+8} ${cy-84} L ${cx} ${cy-82} Z`} fill="#ff6622" />
          {/* Buttons */}
          {[0,1,2].map(i => <circle key={i} cx={cx} cy={cy-68+i*8} r={2.5} fill="#2a3040" />)}
          {/* Scarf */}
          <path d={`M ${cx-18} ${cy-66} Q ${cx} ${cy-60} ${cx+18} ${cy-66} Q ${cx+18} ${cy-58} ${cx} ${cy-54} Q ${cx-18} ${cy-58} ${cx-18} ${cy-66} Z`}
            fill="#cc2222" stroke="#0d111e" strokeWidth={0.8} opacity={0.9} />
          {/* Hat */}
          <rect x={cx-13} y={cy-104} width={26} height={19} fill="#1a2030" rx="2" stroke="#0d111e" strokeWidth={0.8} />
          <rect x={cx-18} y={cy-87} width={36} height={5} fill="#1a2030" rx="2" stroke="#0d111e" strokeWidth={0.8} />
          {/* Arms */}
          <line x1={cx-28} y1={cy-64} x2={cx-52} y2={cy-48} stroke="#5a5a6a" strokeWidth={3} />
          <line x1={cx-52} y1={cy-48} x2={cx-60} y2={cy-38} stroke="#5a5a6a" strokeWidth={2} />
          <line x1={cx+28} y1={cy-64} x2={cx+52} y2={cy-48} stroke="#5a5a6a" strokeWidth={3} />
          <line x1={cx+52} y1={cy-48} x2={cx+60} y2={cy-38} stroke="#5a5a6a" strokeWidth={2} />
        </g>
      );
    })()}

    {/* People in winter coats */}
    <IsoPerson gx={-1} gy={3.5} gz={0} bodyColor="#2244aa" skinColor="#e8c8a0" hatColor="#1a2030" />
    <g style={{ animation: 'sl-walk 9s linear infinite' }}>
      <IsoPerson gx={2} gy={3.5} gz={0} bodyColor="#556633" skinColor="#e0b890" hatColor="#aa8833" />
    </g>
    <IsoPerson gx={6.5} gy={3.8} gz={0} bodyColor="#443322" skinColor="#d4a878" hatColor="#884422" />

    {/* Footprints in snow */}
    {Array.from({ length: 8 }).map((_,i) => {
      const [fx, fy] = px(2 + i * 0.5, 4.5, 0);
      return <ellipse key={i} cx={fx} cy={fy + (i%2)*5} rx={5} ry={3} fill="#c8dcf0" opacity={0.7} />;
    })}
  </svg>
);

// ─── Forest: Sunny Trail ─────────────────────────────────────────────────────
const Forest: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="fo-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#58baf2" />
        <stop offset="55%" stopColor="#8dd0f5" />
        <stop offset="100%" stopColor="#b8e4f8" />
      </linearGradient>
      <filter id="fo-blur"><feGaussianBlur stdDeviation="10" /></filter>
      <filter id="fo-glow">
        <feGaussianBlur stdDeviation="3" result="b" />
        <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
      </filter>
      <style>{`
        @keyframes fo-sway { 0%,100%{transform:rotate(-1.5deg)} 50%{transform:rotate(1.5deg)} }
        @keyframes fo-sway2 { 0%,100%{transform:rotate(2deg)} 50%{transform:rotate(-2deg)} }
        @keyframes fo-bird { 0%{transform:translate(0,0)} 50%{transform:translate(35px,-12px)} 100%{transform:translate(70px,0)} }
        @keyframes fo-butterfly { 0%{transform:translate(0,0)} 25%{transform:translate(8px,-5px)} 50%{transform:translate(16px,0)} 75%{transform:translate(8px,5px)} 100%{transform:translate(0,0)} }
        @keyframes fo-light { 0%,100%{opacity:0.06} 50%{opacity:0.14} }
        @keyframes fo-deer { 0%,100%{transform:translateX(0)} 50%{transform:translateX(-6px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#fo-sky)" />

    {/* Sun glow */}
    <circle cx="720" cy="200" r="210" fill="#f0e888" opacity="0.12" filter="url(#fo-blur)" />

    {/* Background canopy silhouette */}
    {([[0,285,140],[88,262,118],[195,242,148],[308,258,132],[415,238,158],[532,250,128],
      [638,230,162],[758,242,152],[882,234,142],[1002,248,130],[1115,238,155],
      [1238,250,135],[1352,258,125]] as [number,number,number][]).map(([tx,ty,tw],i) => (
        <ellipse key={`fo-bg${i}`} cx={tx+tw/2} cy={ty} rx={tw/2} ry={tw/3.5}
          fill={i%2===0?'#2d7028':'#347530'} opacity={0.68} />
      ))}

    {/* Sunbeams */}
    {[585,665,745,825,905].map((bx,i) => (
      <path key={`fo-beam${i}`} d={`M ${bx-32} 178 L ${bx-85} 625 L ${bx+85} 625 L ${bx+32} 178 Z`}
        fill="#f8f060" opacity={0.07}
        style={{ animation: `fo-light ${3+i*0.6}s ease-in-out ${i*0.4}s infinite` }} />
    ))}

    {/* Ground tiles - grass */}
    {Array.from({ length: 14 }).map((_,i) =>
      Array.from({ length: 8 }).map((_,j) => {
        const isDirt = (i >= 5 && i <= 8) && j >= 2;
        return (
          <GroundTile key={`fo-gt-${i}-${j}`} gx={i-4} gy={j+1} W={1} D={1}
            fill={isDirt ? (j%2===0?'#c4a870':'#b8985a') : ((i+j)%3===0?'#4a8030':(i+j)%3===1?'#559038':'#488028')} />
        );
      })
    )}
    {/* More grass tiles */}
    {Array.from({ length: 14 }).map((_,i) =>
      Array.from({ length: 8 }).map((_,j) => {
        const isDirt = (i >= 5 && i <= 8) && j >= 2;
        if (!isDirt) return null;
        return (
          <GroundTile key={`fo-dirt-${i}-${j}`} gx={i-4} gy={j+1} W={1} D={1}
            fill={j%2===0?'#c4a870':'#b8985a'} />
        );
      })
    )}

    {/* Main trees (6 large, stacked octagonal layers via IsoTree) */}
    <g style={{ animation: 'fo-sway 7s ease-in-out infinite', transformOrigin: `${px(-3,2,0)[0]}px ${px(-3,2,0)[1]}px` }}>
      <IsoTree gx={-3} gy={2} gz={0} colors={['#4ab840','#38a030','#268820']} />
    </g>
    <g style={{ animation: 'fo-sway2 9s ease-in-out 1s infinite', transformOrigin: `${px(-2,4,0)[0]}px ${px(-2,4,0)[1]}px` }}>
      <IsoTree gx={-2} gy={4} gz={0} colors={['#42b038','#309028','#207818']} />
    </g>
    <g style={{ animation: 'fo-sway 8s ease-in-out 2s infinite', transformOrigin: `${px(-3,6,0)[0]}px ${px(-3,6,0)[1]}px` }}>
      <IsoTree gx={-3} gy={6} gz={0} colors={['#50c048','#3ea835','#2a9025']} />
    </g>
    <g style={{ animation: 'fo-sway2 6.5s ease-in-out 3s infinite', transformOrigin: `${px(7,2,0)[0]}px ${px(7,2,0)[1]}px` }}>
      <IsoTree gx={7} gy={2} gz={0} colors={['#48b840','#369030','#247820']} />
    </g>
    <g style={{ animation: 'fo-sway 7.5s ease-in-out 0.5s infinite', transformOrigin: `${px(6,4,0)[0]}px ${px(6,4,0)[1]}px` }}>
      <IsoTree gx={6} gy={4} gz={0} colors={['#44b03c','#328830','#207018']} />
    </g>
    <g style={{ animation: 'fo-sway2 8.5s ease-in-out 1.5s infinite', transformOrigin: `${px(7,6,0)[0]}px ${px(7,6,0)[1]}px` }}>
      <IsoTree gx={7} gy={6} gz={0} colors={['#4cc044','#38a832','#269020']} />
    </g>

    {/* Smaller mid-ground trees */}
    {[[-1,2],[0,5],[5,2],[4,5]].map(([tx,ty],i) => (
      <g key={`fo-sm-tree${i}`}>
        <Building gx={tx+0.25} gy={ty+0.25} gz={0} W={0.5} D={0.5} h={0.9}
          topColor="#4a3010" eastColor="#3a2208" southColor="#2e1806" />
        <polygon points={topFace(tx-0.3, ty-0.3, 0, 1.6, 1.6, 1.6)} fill="#267820" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
        <polygon points={topFace(tx-0.15, ty-0.15, 0, 1.3, 1.3, 2.2)} fill="#308828" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
        <polygon points={topFace(tx, ty, 0, 1.0, 1.0, 2.7)} fill="#3a9830" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
      </g>
    ))}

    {/* Dirt path tiles */}
    {Array.from({ length: 8 }).map((_,j) => (
      <g key={`fo-path-${j}`}>
        <GroundTile gx={1} gy={j+1} W={1} D={1} fill="#c4a870" />
        <GroundTile gx={2} gy={j+1} W={1} D={1} fill="#b8985a" />
      </g>
    ))}
    {/* Path pebbles */}
    {Array.from({ length: 8 }).map((_,i) => {
      const [px2, py2] = px(1.3 + (i%3)*0.25, 1.5+i, 0);
      return <ellipse key={i} cx={px2} cy={py2} rx={5} ry={3} fill="#c09858" stroke="#0d111e" strokeWidth={0.4} opacity={0.7}
        transform={`rotate(${i*25} ${px2} ${py2})`} />;
    })}

    {/* Flower patches */}
    {[[-3.6,1.5,'#ffdd22'],[-3.2,1.8,'#ff6688'],[-2.8,1.5,'#ff9933'],
      [7.5,1.5,'#ffdd22'],[7.8,1.8,'#ff6688'],[8.2,1.4,'#44aaff'],
      [-3.5,7.5,'#ff6688'],[-3,7.8,'#ffdd22'],
      [7.4,7.5,'#ffaa33'],[7.9,7.8,'#ff6688']].map(([fx,fy,fc],i) => {
      const [cx2, cy2] = px(fx as number, fy as number, 0);
      return (
        <g key={`fo-fl${i}`}>
          <line x1={cx2} y1={cy2} x2={cx2} y2={cy2+12} stroke="#3a7828" strokeWidth={1.5} />
          <circle cx={cx2} cy={cy2} r={5.5} fill={fc as string} stroke="#0d111e" strokeWidth={0.6} opacity={0.92} />
          <circle cx={cx2} cy={cy2} r={2.5} fill="white" opacity={0.7} />
        </g>
      );
    })}

    {/* Mushrooms */}
    {[[-2.8,7.5,'#cc3322',14],[-2.5,7.8,'#cc3322',10],[7.2,7.5,'#dd4422',12]].map(([mx,my,mc,mr],i) => {
      const [cx2, cy2] = px(mx as number, my as number, 0);
      return (
        <g key={`fo-mush${i}`}>
          <rect x={cx2-3} y={cy2-10} width={6} height={12} fill="#e8d8b0" rx="1" stroke="#0d111e" strokeWidth={0.5} />
          <ellipse cx={cx2} cy={cy2-10} rx={mr as number} ry={(mr as number)*0.55} fill={mc as string} stroke="#0d111e" strokeWidth={0.6} />
          {[0,1,2].map(j => <ellipse key={j} cx={cx2-(mr as number)*0.4+j*(mr as number)*0.4} cy={cy2-8} rx={2} ry={1.2} fill="white" opacity={0.6} />)}
        </g>
      );
    })}

    {/* Ferns */}
    {[[-2.2,2.5],[-2.5,6.5],[6.8,2.5],[6.5,6.5]].map(([fx,fy],i) => {
      const [cx2, cy2] = px(fx, fy, 0);
      return (
        <g key={`fo-fern${i}`}>
          {[[-14,-10],[-7,-16],[0,-18],[7,-16],[14,-10]].map(([dx,dy],j) => (
            <ellipse key={j} cx={cx2+dx} cy={cy2+dy} rx={6} ry={3.5} fill="#2a7020"
              transform={`rotate(${dx*2} ${cx2+dx} ${cy2+dy})`} opacity={0.85} stroke="#0d111e" strokeWidth={0.4} />
          ))}
        </g>
      );
    })}

    {/* Deer */}
    <g style={{ animation: 'fo-deer 3s ease-in-out infinite alternate' }}>
      {(() => {
        const [bx, by] = px(3, 4, 0);
        return (
          <g>
            {/* body */}
            <ellipse cx={bx} cy={by-18} rx={28} ry={16} fill="#8a5a30" stroke="#0d111e" strokeWidth={0.8} />
            {/* neck */}
            <ellipse cx={bx+18} cy={by-28} rx={9} ry={13} fill="#8a5a30" stroke="#0d111e" strokeWidth={0.8} />
            {/* head */}
            <ellipse cx={bx+28} cy={by-40} rx={12} ry={9} fill="#8a5a30" stroke="#0d111e" strokeWidth={0.8} />
            <ellipse cx={bx+34} cy={by-42} rx={4.5} ry={7} fill="#8a5a30" stroke="#0d111e" strokeWidth={0.8} />
            <ellipse cx={bx+34} cy={by-42} rx={2.5} ry={4.5} fill="#c87860" />
            <circle cx={bx+31} cy={by-44} r={2.2} fill="#1a1a1a" />
            <circle cx={bx+32} cy={by-45} r={0.9} fill="white" opacity={0.8} />
            {/* antlers */}
            <path d={`M ${bx+28} ${by-49} L ${bx+32} ${by-65} L ${bx+40} ${by-58} M ${bx+32} ${by-65} L ${bx+26} ${by-60}`}
              fill="none" stroke="#6a4020" strokeWidth={2.5} strokeLinecap="round" />
            {/* legs */}
            {[[-18,0],[-8,0],[8,0],[18,0]].map(([lx2],idx) => (
              <line key={idx} x1={bx+lx2} y1={by-6} x2={bx+lx2+(idx<2?-2:2)} y2={by+10}
                stroke="#6a4020" strokeWidth={3.5} strokeLinecap="round" />
            ))}
            {/* tail */}
            <ellipse cx={bx-26} cy={by-22} rx={6} ry={4} fill="white" stroke="#0d111e" strokeWidth={0.6} opacity={0.9} />
          </g>
        );
      })()}
    </g>

    {/* Birds */}
    {([[385,242],[418,228],[450,242],[905,218],[938,204],[966,220]] as [number,number][]).map(([bx,by],i) => (
      <g key={`fo-bird${i}`} style={{ animation: `fo-bird ${4+i*0.5}s ease-in-out ${i*0.7}s infinite` }}>
        <path d={`M ${bx-10} ${by} Q ${bx-4} ${by-6} ${bx} ${by}`} fill="none" stroke="#2a2a3a" strokeWidth={1.8} />
        <path d={`M ${bx} ${by} Q ${bx+4} ${by-6} ${bx+10} ${by}`} fill="none" stroke="#2a2a3a" strokeWidth={1.8} />
      </g>
    ))}

    {/* Butterflies */}
    {[[540,522],[482,535],[1008,518]].map(([btx,bty],i) => (
      <g key={`fo-bf${i}`}
        style={{ animation: `fo-butterfly ${2.5+i}s ease-in-out ${i*0.6}s infinite`, transformOrigin: `${btx}px ${bty}px` }}>
        <ellipse cx={btx-8} cy={bty} rx={9} ry={6} fill={i===0?'#ff8844':i===1?'#ee44cc':'#44aaff'} opacity={0.88}
          transform={`rotate(20 ${btx-8} ${bty})`} />
        <ellipse cx={btx+8} cy={bty} rx={9} ry={6} fill={i===0?'#ff8844':i===1?'#ee44cc':'#44aaff'} opacity={0.88}
          transform={`rotate(-20 ${btx+8} ${bty})`} />
        <ellipse cx={btx-5} cy={bty+4} rx={5} ry={3.5} fill={i===0?'#ffaa66':i===1?'#ff66dd':'#66bbff'} opacity={0.75}
          transform={`rotate(30 ${btx-5} ${bty+4})`} />
        <ellipse cx={btx+5} cy={bty+4} rx={5} ry={3.5} fill={i===0?'#ffaa66':i===1?'#ff66dd':'#66bbff'} opacity={0.75}
          transform={`rotate(-30 ${btx+5} ${bty+4})`} />
        <line x1={btx-3} y1={bty-5} x2={btx-10} y2={bty-14} stroke="#2a2a3a" strokeWidth={1} />
        <line x1={btx+3} y1={bty-5} x2={btx+10} y2={bty-14} stroke="#2a2a3a" strokeWidth={1} />
      </g>
    ))}

    {/* Hikers on path */}
    <IsoPerson gx={1.5} gy={3} gz={0} bodyColor="#338844" skinColor="#e0b888" hatColor="#aa6622" />
    {/* Hiker 2 with backpack */}
    {(() => {
      const bS = southFace(1.7, 5.85, 0, 0.3, 0.3, 1.4);
      const bE = eastFace(1.7, 5.85, 0, 0.3, 0.3, 1.4);
      const bT = topFace(1.7, 5.85, 0, 0.3, 0.3, 1.4);
      const [hx, hy] = px(1.85, 5.85, 1.4);
      const [bpx, bpy] = px(1.65, 6, 1.1);
      const [wx, wy] = px(2.1, 5.85, 0);
      const [wx2, wy2] = px(2.1, 5.85, 1.4);
      return (
        <g>
          <polygon points={bS} fill="#4455aa" stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <polygon points={bE} fill="#4455aa" stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <polygon points={bT} fill="#4455aa" stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <circle cx={hx} cy={hy-12} r={8} fill="#d4a070" stroke="#0d111e" strokeWidth={0.7} />
          {/* Backpack */}
          <rect x={bpx-8} y={bpy-18} width={14} height={20} fill="#cc4422" rx="2" stroke="#0d111e" strokeWidth={0.7} />
          {/* Walking stick */}
          <line x1={wx} y1={wy} x2={wx2} y2={wy2-30} stroke="#8a6030" strokeWidth={2.5} strokeLinecap="round" />
        </g>
      );
    })()}

    {/* Log on ground */}
    {(() => {
      const logS = southFace(-1.5, 7, 0, 1, 0.25, 0.25);
      const logE = eastFace(-1.5, 7, 0, 1, 0.25, 0.25);
      const logT = topFace(-1.5, 7, 0, 1, 0.25, 0.25);
      return (
        <g>
          <polygon points={logS} fill="#7a5030" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
          <polygon points={logE} fill="#6a4020" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
          <polygon points={logT} fill="#8a6040" stroke="#0d111e" strokeWidth={0.8} strokeLinejoin="round" />
        </g>
      );
    })()}

    {/* Rock cluster */}
    {[[-3.8,4.5,'#9a9888'],[-3.4,4.8,'#b0ae9c'],[-3.7,5.1,'#8a8878']].map(([rx,ry,rc],i) => {
      const rockS = southFace(rx as number, ry as number, 0, 0.3, 0.25, 0.2+i*0.05);
      const rockE = eastFace(rx as number, ry as number, 0, 0.3, 0.25, 0.2+i*0.05);
      const rockT = topFace(rx as number, ry as number, 0, 0.3, 0.25, 0.2+i*0.05);
      return (
        <g key={i}>
          <polygon points={rockS} fill={rc as string} stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <polygon points={rockE} fill={rc as string} stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
          <polygon points={rockT} fill={rc as string} stroke="#0d111e" strokeWidth={0.7} strokeLinejoin="round" />
        </g>
      );
    })}
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
