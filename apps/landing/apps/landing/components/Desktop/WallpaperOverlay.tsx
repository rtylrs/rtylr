import React from 'react';
import { Wallpaper } from '../../stores/useDesktopStore';

// ─── Midnight: Night City Street ─────────────────────────────────────────────
const Midnight: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="mn-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#040a18" />
        <stop offset="100%" stopColor="#0e1b38" />
      </linearGradient>
      <filter id="mn-blur">
        <feGaussianBlur stdDeviation="10" />
      </filter>
      <filter id="mn-glow">
        <feGaussianBlur stdDeviation="3" result="b" />
        <feMerge>
          <feMergeNode in="b" />
          <feMergeNode in="SourceGraphic" />
        </feMerge>
      </filter>
      <pattern id="mn-stripe" x="0" y="0" width="18" height="100" patternUnits="userSpaceOnUse">
        <rect width="18" height="100" fill="#e03030" />
        <rect x="9" width="9" height="100" fill="#f5f5f5" opacity="0.88" />
      </pattern>
      <style>{`
        @keyframes mn-twinkle { 0%,100%{opacity:1} 50%{opacity:0.1} }
        @keyframes mn-car1 { from{transform:translateX(0)} to{transform:translateX(-280px)} }
        @keyframes mn-car2 { from{transform:translateX(0)} to{transform:translateX(240px)} }
        @keyframes mn-neon { 0%,88%,100%{opacity:1} 90%{opacity:0.3} 92%{opacity:1} 94%{opacity:0.5} 96%{opacity:1} }
        @keyframes mn-walk { 0%{transform:translateX(0)} 100%{transform:translateX(45px)} }
        @keyframes mn-walk2 { 0%{transform:translateX(0)} 100%{transform:translateX(-35px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#mn-sky)" />

    {/* Stars */}
    {(
      [
        [55, 28, 1.2], [148, 15, 0.9], [290, 48, 1.4], [418, 22, 0.8], [542, 40, 1.1],
        [675, 18, 1.3], [802, 52, 0.9], [938, 30, 1.2], [1072, 44, 0.8], [1198, 20, 1.3],
        [1348, 38, 1.0], [1415, 62, 0.7], [82, 95, 0.8], [215, 82, 1.1], [362, 108, 0.9],
        [498, 75, 1.2], [632, 100, 0.8], [768, 85, 1.3], [895, 115, 0.9], [1028, 80, 1.1],
        [1155, 102, 0.8], [1302, 88, 1.2], [1382, 118, 0.7], [42, 152, 1.0], [188, 165, 0.8],
        [328, 138, 1.2], [472, 172, 0.9], [612, 155, 1.1], [745, 142, 0.8], [882, 168, 1.3],
        [1015, 148, 0.9], [1148, 162, 0.8], [1288, 145, 1.1], [1432, 162, 0.7],
        [115, 215, 0.7], [252, 232, 1.0], [392, 202, 0.8], [528, 218, 1.1], [672, 235, 0.7],
        [812, 208, 1.2], [948, 225, 0.8], [1088, 212, 1.0], [1228, 238, 0.7], [1365, 222, 0.9],
      ] as [number, number, number][]
    ).map(([cx, cy, r], i) => (
      <circle
        key={`mn-s${i}`}
        cx={cx}
        cy={cy}
        r={r}
        fill="white"
        opacity={0.4 + (i % 4) * 0.12}
        style={
          i % 5 === 0
            ? { animation: `mn-twinkle ${1.5 + (i % 4) * 0.7}s ease-in-out ${(i % 3) * 0.8}s infinite` }
            : undefined
        }
      />
    ))}

    {/* Moon */}
    <ellipse cx="1310" cy="88" rx="75" ry="75" fill="#fffbea" opacity="0.1" filter="url(#mn-blur)" />
    <circle cx="1310" cy="88" r="48" fill="#f4efd8" />
    <circle cx="1296" cy="76" r="8" fill="#e6dec0" opacity="0.65" />
    <circle cx="1320" cy="100" r="5" fill="#e6dec0" opacity="0.55" />
    <circle cx="1330" cy="75" r="3.5" fill="#e6dec0" opacity="0.45" />

    {/* Far background silhouettes */}
    <rect x="0" y="318" width="92" height="302" fill="#0b1226" rx="2" />
    <rect x="22" y="278" width="52" height="342" fill="#0d1430" rx="2" />
    <rect x="72" y="348" width="78" height="272" fill="#0a1020" rx="2" />
    <rect x="138" y="298" width="68" height="322" fill="#0c1228" rx="2" />
    <rect x="192" y="262" width="44" height="358" fill="#0e1530" rx="2" />
    {(
      [
        [12, 332], [12, 362], [12, 392], [12, 422], [12, 452], [28, 292], [28, 322], [28, 352],
        [28, 382], [28, 412], [28, 442], [78, 362], [78, 392], [78, 422], [78, 452],
        [143, 312], [143, 342], [143, 372], [143, 402], [143, 432], [197, 278], [197, 308],
        [197, 338], [197, 368], [197, 398], [197, 428],
      ] as [number, number][]
    ).map(([wx, wy], i) => (
      <rect key={`mn-fw${i}`} x={wx} y={wy} width={7} height={6} fill="#ffd860" opacity={0.15 + (i % 5) * 0.06} rx={1} />
    ))}
    <rect x="1172" y="288" width="88" height="332" fill="#0b1226" rx="2" />
    <rect x="1238" y="258" width="62" height="362" fill="#0d1430" rx="2" />
    <rect x="1288" y="318" width="78" height="302" fill="#0a1020" rx="2" />
    <rect x="1352" y="298" width="88" height="322" fill="#0c1228" rx="2" />
    {(
      [
        [1178, 302], [1178, 332], [1178, 362], [1178, 392], [1178, 422],
        [1243, 272], [1243, 302], [1243, 332], [1243, 362], [1243, 392],
        [1293, 332], [1293, 362], [1293, 392], [1293, 422],
        [1358, 312], [1358, 342], [1358, 372], [1358, 402],
      ] as [number, number][]
    ).map(([wx, wy], i) => (
      <rect key={`mn-rfw${i}`} x={wx} y={wy} width={7} height={6} fill="#ffd860" opacity={0.15 + (i % 5) * 0.06} rx={1} />
    ))}

    {/* Left apartment building */}
    <rect x="0" y="252" width="172" height="368" fill="#131d35" rx="3" />
    {Array.from({ length: 8 }).map((_, row) =>
      Array.from({ length: 4 }).map((_, col) => (
        <rect
          key={`mn-lb1-${row}-${col}`}
          x={14 + col * 38}
          y={268 + row * 36}
          width={22}
          height={18}
          fill={(row + col) % 3 === 0 ? '#fff8e0' : '#ffe070'}
          opacity={0.08 + ((row + col) % 4) * 0.06}
          rx={2}
        />
      ))
    )}
    <rect x="78" y="238" width="22" height="16" fill="#0e1628" rx="2" />
    <ellipse cx="89" cy="238" rx="13" ry="4.5" fill="#0c1422" />
    <line x1="76" y1="254" x2="76" y2="268" stroke="#0e1628" strokeWidth="3" />
    <line x1="103" y1="254" x2="103" y2="268" stroke="#0e1628" strokeWidth="3" />

    {/* Left commercial building */}
    <rect x="165" y="358" width="132" height="262" fill="#152238" rx="3" />
    {Array.from({ length: 5 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect
          key={`mn-lb2-${row}-${col}`}
          x={176 + col * 40}
          y={372 + row * 38}
          width={24}
          height={20}
          fill="#ffc840"
          opacity={0.1 + ((row + col) % 4) * 0.05}
          rx={2}
        />
      ))
    )}
    <rect
      x="176"
      y="364"
      width="112"
      height="9"
      fill="#0066ff"
      opacity={0.55}
      rx={2}
      style={{ animation: 'mn-neon 4s ease-in-out 0.5s infinite' }}
    />

    {/* MAIN SHOP */}
    <rect x="455" y="385" width="268" height="235" fill="#1a2d4a" rx="4" />
    {([[475, 398], [538, 398], [600, 398], [662, 398], [475, 438], [538, 438], [600, 438], [662, 438]] as [number, number][]).map(
      ([wx, wy], i) => (
        <rect key={`mn-sw${i}`} x={wx} y={wy} width={32} height={28} fill="#ffeeaa" opacity={0.3 + (i % 3) * 0.08} rx={3} />
      )
    )}
    <path d="M 445 483 L 733 483 L 713 467 L 465 467 Z" fill="url(#mn-stripe)" />
    {Array.from({ length: 15 }).map((_, i) => (
      <ellipse key={`mn-sc${i}`} cx={452 + i * 19} cy={483} rx={7} ry={8} fill="#c01818" />
    ))}
    <rect x="470" y="490" width="238" height={102} fill="#1e3a5a" rx="4" />
    {([[490, 505, '#ff6b35'], [540, 498, '#4fc3f7'], [590, 506, '#aed581'], [640, 500, '#ff8a65'], [692, 505, '#ce93d8']] as [number, number, string][]).map(
      ([px, py, c], i) => (
        <ellipse key={`mn-prod${i}`} cx={px} cy={py} rx={18} ry={15} fill={c} opacity={0.7} />
      )
    )}
    {([[482, 527], [535, 520], [584, 528], [636, 522], [685, 527]] as [number, number][]).map(([px, py], i) => (
      <rect key={`mn-tag${i}`} x={px} y={py} width={18} height={10} fill="white" opacity={0.5} rx={2} />
    ))}
    <rect x="568" y="516" width={44} height={76} fill="#0d1e32" rx="3" />
    <rect x="572" y="520" width={16} height={32} fill="#1e3555" rx="2" opacity="0.8" />
    <rect x="591" y="520" width={16} height={32} fill="#1e3555" rx="2" opacity="0.8" />
    <circle cx="590" cy="558" r="3" fill="#888" />
    <rect
      x="485"
      y="393"
      width={208}
      height={40}
      fill="#ff6b35"
      rx="6"
      style={{ animation: 'mn-neon 6s ease-in-out 1s infinite' }}
    />
    <text x="589" y="418" textAnchor="middle" fill="white" fontSize="16" fontWeight="700" fontFamily="system-ui, sans-serif">
      THE SHOP
    </text>
    <text x="589" y="479" textAnchor="middle" fill="rgba(255,255,255,0.7)" fontSize="10" fontFamily="system-ui, sans-serif">
      OPEN 24H
    </text>
    <ellipse cx="589" cy="640" rx="160" ry="28" fill="#ffee80" opacity="0.05" filter="url(#mn-blur)" />

    {/* Flower pots */}
    <rect x="456" y="600" width={15} height={18} fill="#8a3a20" rx="2" />
    <ellipse cx="463" cy="600" rx="9" ry="4" fill="#7a3018" />
    <circle cx="463" cy="594" r="8" fill="#2a6030" />
    <circle cx="458" cy="590" r="4" fill="#ff6644" opacity="0.85" />
    <circle cx="468" cy="591" r="3" fill="#ff9966" opacity="0.85" />
    <rect x="718" y="600" width={15} height={18} fill="#8a3a20" rx="2" />
    <ellipse cx="725" cy="600" rx="9" ry="4" fill="#7a3018" />
    <circle cx="725" cy="594" r="8" fill="#2a6030" />
    <circle cx="720" cy="590" r="4" fill="#ffaa33" opacity="0.85" />
    <circle cx="730" cy="591" r="3" fill="#ff6644" opacity="0.85" />

    {/* Right buildings */}
    <rect x="720" y="338" width="148" height="282" fill="#14203a" rx="3" />
    {Array.from({ length: 6 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect key={`mn-rb1-${row}-${col}`} x={730 + col * 42} y={352 + row * 36} width={26} height={22} fill="#ffdc60" opacity={0.08 + ((row * col) % 4) * 0.06} rx={2} />
      ))
    )}
    <rect x="855" y="268" width="128" height="352" fill="#12203a" rx="3" />
    {Array.from({ length: 9 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect key={`mn-rb2-${row}-${col}`} x={864 + col * 38} y={282 + row * 32} width={24} height={20} fill="#ffc830" opacity={0.1 + ((row + col) % 5) * 0.05} rx={2} />
      ))
    )}
    <rect x="982" y="308" width="7" height="278" fill="#1a2840" opacity="0.75" />
    {[338, 368, 398, 428, 458, 488, 518, 548].map((fy, i) => (
      <rect key={`mn-fe${i}`} x={968} y={fy} width={21} height={3} fill="#1a2840" opacity="0.75" />
    ))}
    <rect x="998" y="388" width="108" height="232" fill="#152030" rx="3" />
    {Array.from({ length: 5 }).map((_, row) =>
      Array.from({ length: 2 }).map((_, col) => (
        <rect key={`mn-rb3-${row}-${col}`} x={1008 + col * 48} y={398 + row * 38} width={30} height={24} fill="#ffcc40" opacity={0.08 + ((row * col) % 4) * 0.06} rx={2} />
      ))
    )}
    <rect x="1102" y="358" width="142" height="262" fill="#131e35" rx="3" />
    {Array.from({ length: 6 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect key={`mn-rb4-${row}-${col}`} x={1112 + col * 42} y={372 + row * 35} width={28} height={20} fill="#ffaa20" opacity={0.1 + ((row + col) % 4) * 0.05} rx={2} />
      ))
    )}
    <rect x="1117" y="365" width="122" height="9" fill="#22ffaa" opacity="0.45" rx={2} style={{ animation: 'mn-neon 3.5s ease-in-out 2s infinite' }} />

    {/* Sidewalk & Road */}
    <rect x="0" y="598" width="1440" height="52" fill="#1c2540" />
    {Array.from({ length: 25 }).map((_, i) => (
      <line key={`mn-tile${i}`} x1={i * 58} y1={598} x2={i * 58} y2={650} stroke="#181e38" strokeWidth={1} opacity={0.55} />
    ))}
    <rect x="0" y="650" width="1440" height="112" fill="#10141e" />
    {Array.from({ length: 14 }).map((_, i) => (
      <rect key={`mn-lm${i}`} x={i * 106} y={700} width={64} height={5} fill="#2a3050" rx="2" />
    ))}
    <rect x="0" y="762" width="1440" height="42" fill="#1c2540" />
    {Array.from({ length: 25 }).map((_, i) => (
      <line key={`mn-ltile${i}`} x1={i * 58} y1={762} x2={i * 58} y2={804} stroke="#181e38" strokeWidth={1} opacity={0.55} />
    ))}
    <rect x="0" y="804" width="1440" height="96" fill="#141a2e" />

    {/* Street Lights */}
    {[195, 618, 1038, 1348].map((lx, i) => (
      <g key={`mn-lamp${i}`}>
        <ellipse cx={lx + 32} cy={672} rx={55} ry={18} fill="#ffaa30" opacity={0.1} filter="url(#mn-blur)" />
        <rect x={lx} y={488} width={5} height={162} fill="#2a3550" rx={2} />
        <path d={`M ${lx + 2} 488 Q ${lx + 30} 482 ${lx + 38} 468`} fill="none" stroke="#2a3550" strokeWidth={4} />
        <rect x={lx + 25} y={462} width={24} height={11} fill="#1e2b45" rx={2} />
        <ellipse cx={lx + 37} cy={463} rx={9} ry={5} fill="#ffe060" filter="url(#mn-glow)" />
      </g>
    ))}

    {/* Cars */}
    <g style={{ animation: 'mn-car1 20s linear infinite' }} transform="translate(820, 0)">
      <ellipse cx="792" cy="694" rx="28" ry="8" fill="#ffffaa" opacity={0.12} filter="url(#mn-blur)" />
      <rect x="716" y="662" width="86" height="27" fill="#c02828" rx="5" />
      <rect x="732" y="646" width="52" height="19" fill="#aa1e1e" rx="7" />
      <rect x="735" y="649" width="21" height="13" fill="#88c8ff" opacity={0.5} rx="2" />
      <rect x="759" y="649" width="21" height="13" fill="#88c8ff" opacity={0.5} rx="2" />
      <rect x="800" y="669" width="6" height="10" fill="#ffffaa" rx="1" />
      <rect x="714" y="669" width="5" height="10" fill="#ff3322" rx="1" opacity={0.8} />
      <circle cx="733" cy="690" r="9" fill="#1a1a1a" />
      <circle cx="733" cy="690" r="4.5" fill="#333" />
      <circle cx="786" cy="690" r="9" fill="#1a1a1a" />
      <circle cx="786" cy="690" r="4.5" fill="#333" />
    </g>
    <g style={{ animation: 'mn-car2 24s linear 6s infinite' }} transform="translate(175, 0)">
      <ellipse cx="212" cy="732" rx="22" ry="8" fill="#ff4422" opacity={0.1} filter="url(#mn-blur)" />
      <rect x="218" y="708" width="86" height="27" fill="#2252a8" rx="5" />
      <rect x="234" y="692" width="52" height="19" fill="#1a3e8e" rx="7" />
      <rect x="237" y="695" width="21" height="13" fill="#88c8ff" opacity={0.5} rx="2" />
      <rect x="261" y="695" width="21" height="13" fill="#88c8ff" opacity={0.5} rx="2" />
      <rect x="302" y="715" width="6" height="10" fill="#ffffaa" rx="1" />
      <rect x="216" y="715" width="5" height="10" fill="#ff3322" rx="1" />
      <circle cx="235" cy="736" r="9" fill="#1a1a1a" />
      <circle cx="235" cy="736" r="4.5" fill="#333" />
      <circle cx="288" cy="736" r="9" fill="#1a1a1a" />
      <circle cx="288" cy="736" r="4.5" fill="#333" />
    </g>
    <g transform="translate(1080, 0)">
      <rect x="0" y="662" width="80" height="26" fill="#e8c010" rx="5" />
      <rect x="14" y="647" width="50" height="18" fill="#d4ac0c" rx="6" />
      <rect x="16" y="649" width="20" height="12" fill="#88c8ff" opacity={0.5} rx="2" />
      <rect x="38" y="649" width="20" height="12" fill="#88c8ff" opacity={0.5} rx="2" />
      <rect x="78" y="668" width="5" height="9" fill="#ffffaa" rx="1" />
      <circle cx="16" cy="689" r="8" fill="#1a1a1a" />
      <circle cx="16" cy="689" r="4" fill="#333" />
      <circle cx="64" cy="689" r="8" fill="#1a1a1a" />
      <circle cx="64" cy="689" r="4" fill="#333" />
    </g>

    {/* People */}
    <g style={{ animation: 'mn-walk 9s linear infinite' }} transform="translate(328, 0)">
      <circle cx="0" cy="588" r="7" fill="#e8b888" />
      <rect x="-5" y="595" width="10" height="13" fill="#4455aa" rx="2" />
      <rect x="-4" y="608" width="4" height="10" fill="#2a3555" rx="1" />
      <rect x="1" y="608" width="4" height="10" fill="#2a3555" rx="1" />
    </g>
    <g transform="translate(428, 0)">
      <circle cx="0" cy="584" r="7" fill="#c89870" />
      <rect x="-5" y="591" width="11" height="14" fill="#cc4433" rx="2" />
      <rect x="-5" y="605" width="4" height="12" fill="#1a2235" rx="1" />
      <rect x="2" y="605" width="4" height="12" fill="#1a2235" rx="1" />
    </g>
    <g transform="translate(778, 0)">
      <circle cx="0" cy="588" r="7" fill="#d4aa80" />
      <rect x="-5" y="595" width="10" height="14" fill="#338855" rx="2" />
      <rect x="-4" y="609" width="4" height="10" fill="#1a2235" rx="1" />
      <rect x="1" y="609" width="4" height="10" fill="#1a2235" rx="1" />
      <rect x="9" y="599" width="9" height="12" fill="#ffaa22" rx="2" />
      <path d="M 10 599 Q 13 595 17 599" fill="none" stroke="#aa7010" strokeWidth="1.5" />
    </g>
    <g style={{ animation: 'mn-walk2 11s linear 2s infinite' }} transform="translate(1082, 0)">
      <circle cx="0" cy="786" r="7" fill="#e0c0a0" />
      <rect x="-5" y="793" width="10" height="14" fill="#554488" rx="2" />
      <rect x="-4" y="807" width="4" height="10" fill="#1a1a2a" rx="1" />
      <rect x="1" y="807" width="4" height="10" fill="#1a1a2a" rx="1" />
    </g>
    <g transform="translate(1258, 0)">
      <rect x="-14" y="793" width="48" height="5" fill="#2a3550" rx="2" />
      <rect x="-11" y="798" width="6" height="12" fill="#243040" />
      <rect x="20" y="798" width="6" height="12" fill="#243040" />
      <circle cx="8" cy="783" r="6" fill="#d4b090" />
      <rect x="3" y="789" width="10" height="10" fill="#3366cc" rx="2" />
      <rect x="3" y="796" width="15" height="4" fill="#334466" rx="1" />
    </g>

    {/* Sidewalk tree */}
    <rect x="350" y="543" width="8" height="58" fill="#1e2d40" />
    <circle cx="354" cy="528" r="26" fill="#1a3022" />
    <circle cx="338" cy="538" r="18" fill="#1e3528" />
    <circle cx="370" cy="540" r="16" fill="#1e3525" />
    <rect x="1178" y="543" width="8" height="58" fill="#1e2d40" />
    <circle cx="1182" cy="526" r="24" fill="#1a3022" />
    <circle cx="1165" cy="537" r="16" fill="#1e3528" />

    {/* Manhole + crosswalk */}
    <circle cx="718" cy="706" r="13" fill="#161d32" />
    <circle cx="718" cy="706" r="11" fill="#121828" stroke="#1a2240" strokeWidth="2" />
    {Array.from({ length: 6 }).map((_, i) => (
      <rect key={`mn-cw${i}`} x={1148 + i * 16} y={650} width={10} height={52} fill="#222a45" rx="1" />
    ))}
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
        <stop offset="0%" stopColor="#4aaee8" />
        <stop offset="65%" stopColor="#7cc8f0" />
        <stop offset="100%" stopColor="#a8dbf5" />
      </linearGradient>
      <filter id="au-blur">
        <feGaussianBlur stdDeviation="8" />
      </filter>
      <style>{`
        @keyframes au-cloud1 { 0%{transform:translateX(0)} 100%{transform:translateX(-40px)} }
        @keyframes au-cloud2 { 0%{transform:translateX(0)} 100%{transform:translateX(32px)} }
        @keyframes au-flag { 0%,100%{transform:skewX(0deg)} 50%{transform:skewX(-6deg)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#au-sky)" />

    {/* Sun */}
    <circle cx="1338" cy="92" r="85" fill="#fff070" opacity="0.2" filter="url(#au-blur)" />
    <circle cx="1338" cy="92" r="52" fill="#fff380" />
    {Array.from({ length: 12 }).map((_, i) => {
      const a = (i * 30 * Math.PI) / 180;
      return (
        <line key={`au-sr${i}`} x1={1338 + Math.cos(a) * 60} y1={92 + Math.sin(a) * 60} x2={1338 + Math.cos(a) * 82} y2={92 + Math.sin(a) * 82} stroke="#ffe050" strokeWidth={i % 2 === 0 ? 3 : 2} opacity={0.7} />
      );
    })}

    {/* Clouds */}
    <g style={{ animation: 'au-cloud1 14s ease-in-out infinite alternate' }}>
      <ellipse cx="282" cy="108" rx="78" ry="33" fill="white" opacity={0.9} />
      <ellipse cx="340" cy="92" rx="60" ry="40" fill="white" opacity={0.85} />
      <ellipse cx="222" cy="114" rx="54" ry="28" fill="white" opacity={0.8} />
    </g>
    <g style={{ animation: 'au-cloud2 18s ease-in-out 3s infinite alternate' }}>
      <ellipse cx="798" cy="72" rx="68" ry="28" fill="white" opacity={0.8} />
      <ellipse cx="852" cy="58" rx="52" ry="35" fill="white" opacity={0.75} />
      <ellipse cx="752" cy="78" rx="48" ry="24" fill="white" opacity={0.7} />
    </g>
    <g style={{ animation: 'au-cloud1 20s ease-in-out 7s infinite alternate' }}>
      <ellipse cx="1078" cy="128" rx="62" ry="26" fill="white" opacity={0.72} />
      <ellipse cx="1128" cy="115" rx="48" ry="30" fill="white" opacity={0.68} />
    </g>

    {/* Background buildings */}
    <rect x="0" y="308" width="118" height="322" fill="#e8d5b0" rx="3" />
    <rect x="18" y="268" width="58" height="362" fill="#f0dcbc" rx="3" />
    {([[28, 282], [28, 318], [28, 354], [28, 390], [88, 322], [88, 358], [88, 394]] as [number, number][]).map(([x, y], i) => (
      <rect key={`au-bw${i}`} x={x} y={y} width={18} height={20} fill="#d4a870" opacity={0.4} rx={2} />
    ))}
    <rect x="1228" y="282" width="108" height="348" fill="#dcc8a0" rx="3" />
    <rect x="1292" y="258" width="78" height="372" fill="#e8d4b0" rx="3" />
    {([[1238, 298], [1238, 338], [1238, 378], [1298, 272], [1298, 312], [1298, 352], [1298, 392]] as [number, number][]).map(([x, y], i) => (
      <rect key={`au-bw2${i}`} x={x} y={y} width={18} height={20} fill="#c4a460" opacity={0.5} rx={2} />
    ))}

    {/* Market ground */}
    <rect x="0" y="578" width="1440" height="322" fill="#d4c4a0" />
    {Array.from({ length: 25 }).map((_, i) => (
      <line key={`au-gv${i}`} x1={i * 58} y1={578} x2={i * 58} y2={900} stroke="#c4b490" strokeWidth={1} opacity={0.45} />
    ))}
    {Array.from({ length: 6 }).map((_, i) => (
      <line key={`au-gh${i}`} x1={0} y1={578 + i * 55} x2={1440} y2={578 + i * 55} stroke="#c4b490" strokeWidth={1} opacity={0.45} />
    ))}

    {/* Bunting */}
    <path d="M 60 450 Q 350 420 720 448 Q 1090 420 1385 452" fill="none" stroke="#ddd" strokeWidth={1.5} opacity={0.7} />
    {Array.from({ length: 22 }).map((_, i) => {
      const x = 62 + i * 60;
      const y = 450 + Math.sin(i * 0.55) * 10;
      const colors = ['#e03030', '#ff9920', '#22aa44', '#3366cc', '#cc44aa', '#ff6622', '#ffcc00'];
      return (
        <polygon
          key={`au-flag${i}`}
          points={`${x},${y} ${x + 15},${y} ${x + 7.5},${y + 17}`}
          fill={colors[i % colors.length]}
          opacity={0.88}
          style={{ animation: `au-flag ${1.5 + (i % 3) * 0.5}s ease-in-out ${i * 0.18}s infinite alternate`, transformOrigin: `${x + 7.5}px ${y}px` }}
        />
      );
    })}

    {/* Stall 1 - Red */}
    <rect x="75" y="478" width="218" height="122" fill="#fdf8ee" rx="2" />
    <rect x="62" y="468" width="244" height="28" fill="#e03030" rx="3" />
    {Array.from({ length: 5 }).map((_, i) => (
      <rect key={`au-s1s${i}`} x={62 + i * 49} y={468} width={24} height={28} fill="#c01010" opacity={0.45} rx="1" />
    ))}
    {Array.from({ length: 13 }).map((_, i) => (
      <ellipse key={`au-s1f${i}`} cx={68 + i * 18} cy={496} rx={7} ry={8} fill="#c01010" />
    ))}
    <rect x="80" y="524" width="208" height="12" fill="#d4a060" rx="2" />
    <rect x="80" y="536" width="5" height={62} fill="#b88040" />
    <rect x="283" y="536" width="5" height={62} fill="#b88040" />
    {([[95, 516, '#ff6b22'], [114, 512, '#ffcc00'], [133, 516, '#55bb22'], [152, 512, '#ff6b22'], [171, 516, '#ffcc00'], [190, 512, '#55bb22'], [209, 516, '#ff6b22'], [228, 512, '#ffcc00'], [247, 516, '#55bb22'], [266, 512, '#ff6b22']] as [number, number, string][]).map(([px, py, c], i) => (
      <circle key={`au-s1p${i}`} cx={px} cy={py} r={9} fill={c} opacity={0.9} />
    ))}
    <rect x="128" y="476" width="112" height="16" fill="#c01010" rx="3" />
    <text x="184" y="488" textAnchor="middle" fill="white" fontSize="11" fontWeight="600" fontFamily="system-ui, sans-serif">FRESH MARKET</text>

    {/* Stall 2 - Orange */}
    <rect x="355" y="472" width="198" height="128" fill="#fffcf0" rx="2" />
    <rect x="342" y="462" width="224" height="28" fill="#ff9920" rx="3" />
    {Array.from({ length: 5 }).map((_, i) => (
      <rect key={`au-s2s${i}`} x={342 + i * 45} y={462} width={22} height={28} fill="#ee7e0e" opacity={0.45} rx="1" />
    ))}
    {Array.from({ length: 12 }).map((_, i) => (
      <ellipse key={`au-s2f${i}`} cx={348 + i * 18} cy={490} rx={7} ry={8} fill="#ee7e0e" />
    ))}
    <rect x="360" y="522" width="188" height="12" fill="#d4a060" rx="2" />
    <rect x="360" y="534" width="5" height={64} fill="#b88040" />
    <rect x="543" y="534" width="5" height={64} fill="#b88040" />
    {Array.from({ length: 9 }).map((_, i) => (
      <ellipse key={`au-s2h${i}`} cx={375 + i * 21} cy={512} rx={7} ry={9} fill={i % 2 === 0 ? '#ff6633' : '#ffaa22'} opacity={0.88} />
    ))}

    {/* Stall 3 - Green hero */}
    <rect x="615" y="458" width="238" height="142" fill="#f0f8f0" rx="2" />
    <rect x="600" y="446" width="268" height="32" fill="#22aa44" rx="3" />
    {Array.from({ length: 6 }).map((_, i) => (
      <rect key={`au-s3s${i}`} x={600 + i * 45} y={446} width={22} height={32} fill="#188833" opacity={0.45} rx="1" />
    ))}
    {Array.from({ length: 14 }).map((_, i) => (
      <ellipse key={`au-s3f${i}`} cx={606 + i * 18} cy={478} rx={7} ry={8} fill="#188833" />
    ))}
    <rect x="620" y="524" width="228" height="14" fill="#d4a060" rx="2" />
    <rect x="620" y="538" width="5" height={62} fill="#b88040" />
    <rect x="843" y="538" width="5" height={62} fill="#b88040" />
    {([[636, 516, '#55cc22'], [658, 510, '#ff6622'], [680, 516, '#ffcc00'], [702, 510, '#44aaff'], [724, 516, '#55cc22'], [746, 510, '#ff6622'], [768, 516, '#ffcc00'], [790, 510, '#44aaff'], [812, 516, '#55cc22'], [834, 510, '#ff6622']] as [number, number, string][]).map(([px, py, c], i) => (
      <circle key={`au-s3p${i}`} cx={px} cy={py} r={10} fill={c} opacity={0.9} />
    ))}
    <rect x="648" y="455" width="172" height="20" fill="#188833" rx="4" />
    <text x="734" y="469" textAnchor="middle" fill="white" fontSize="12" fontWeight="700" fontFamily="system-ui, sans-serif">ORGANIC GOODS</text>

    {/* Stall 4 - Blue */}
    <rect x="924" y="472" width="208" height="128" fill="#f0f0ff" rx="2" />
    <rect x="910" y="462" width="236" height="28" fill="#3366cc" rx="3" />
    {Array.from({ length: 5 }).map((_, i) => (
      <rect key={`au-s4s${i}`} x={910 + i * 47} y={462} width={23} height={28} fill="#2255bb" opacity={0.45} rx="1" />
    ))}
    {Array.from({ length: 12 }).map((_, i) => (
      <ellipse key={`au-s4f${i}`} cx={916 + i * 18} cy={490} rx={7} ry={8} fill="#2255bb" />
    ))}
    <rect x="929" y="522" width="198" height="12" fill="#d4a060" rx="2" />
    <rect x="929" y="534" width="5" height={64} fill="#b88040" />
    <rect x="1122" y="534" width="5" height={64} fill="#b88040" />
    {Array.from({ length: 8 }).map((_, i) => (
      <rect key={`au-s4p${i}`} x={942 + i * 22} y={508} width={16} height={18} fill={i % 3 === 0 ? '#ff8844' : i % 3 === 1 ? '#44bbff' : '#ffcc33'} opacity={0.82} rx={2} />
    ))}

    {/* Stall 5 - Pink */}
    <rect x="1194" y="476" width="200" height="124" fill="#fff0f8" rx="2" />
    <rect x="1180" y="466" width="228" height="28" fill="#cc44aa" rx="3" />
    {Array.from({ length: 5 }).map((_, i) => (
      <rect key={`au-s5s${i}`} x={1180 + i * 46} y={466} width={23} height={28} fill="#bb3399" opacity={0.45} rx="1" />
    ))}
    {Array.from({ length: 12 }).map((_, i) => (
      <ellipse key={`au-s5f${i}`} cx={1186 + i * 18} cy={494} rx={7} ry={8} fill="#bb3399" />
    ))}
    <rect x="1199" y="522" width="188" height="12" fill="#d4a060" rx="2" />
    <rect x="1199" y="534" width="5" height={64} fill="#b88040" />
    <rect x="1382" y="534" width="5" height={64} fill="#b88040" />

    {/* Trees */}
    <rect x="202" y="445" width="10" height="133" fill="#8a6030" />
    <circle cx="207" cy="422" r="40" fill="#2a8828" />
    <circle cx="190" cy="438" r="28" fill="#338833" />
    <circle cx="225" cy="436" r="26" fill="#2a7a25" />
    <circle cx="207" cy="402" r="24" fill="#2d9a2d" />
    <rect x="1148" y="442" width="10" height="136" fill="#8a6030" />
    <circle cx="1153" cy="418" r="38" fill="#258825" />
    <circle cx="1135" cy="434" r="26" fill="#2d9230" />
    <circle cx="1170" cy="432" r="24" fill="#228020" />

    {/* Lamp post */}
    <rect x="739" y="438" width="6" height="162" fill="#888070" />
    <ellipse cx="742" cy="438" rx="14" ry="5.5" fill="#aaaa80" />
    <circle cx="742" cy="435" r="7" fill="#ffeeaa" opacity={0.7} />

    {/* Vendor 1 */}
    <g transform="translate(188, 0)">
      <circle cx="0" cy="562" r="8" fill="#d4a870" />
      <rect x="-6" y="570" width="12" height="14" fill="#e03030" rx="2" />
      <rect x="-4" y="572" width="8" height="11" fill="white" opacity={0.55} rx="1" />
      <rect x="-5" y="584" width="4" height="12" fill="#333" rx="1" />
      <rect x="2" y="584" width="4" height="12" fill="#333" rx="1" />
    </g>
    {/* Customer 1 with basket */}
    <g transform="translate(318, 0)">
      <circle cx="0" cy="565" r="7" fill="#e8c8a0" />
      <rect x="-5" y="572" width="10" height="14" fill="#4488ff" rx="2" />
      <rect x="-4" y="586" width="4" height="11" fill="#225" rx="1" />
      <rect x="1" y="586" width="4" height="11" fill="#225" rx="1" />
      <rect x="8" y="575" width="12" height="10" fill="#c4903a" rx="2" />
      <path d="M 9 575 Q 14 570 20 575" fill="none" stroke="#a07028" strokeWidth="1.5" />
    </g>
    {/* Vendor green */}
    <g transform="translate(728, 0)">
      <circle cx="0" cy="558" r="8" fill="#c8986a" />
      <rect x="-6" y="566" width="12" height="15" fill="#22aa44" rx="2" />
      <rect x="-4" y="568" width="8" height="12" fill="white" opacity={0.5} rx="1" />
      <rect x="-5" y="581" width="4" height="12" fill="#333" rx="1" />
      <rect x="2" y="581" width="4" height="12" fill="#333" rx="1" />
    </g>
    {/* Customer 2 + child */}
    <g transform="translate(850, 0)">
      <circle cx="0" cy="560" r="7" fill="#e0b888" />
      <rect x="-5" y="567" width="10" height="13" fill="#aa44cc" rx="2" />
      <rect x="-4" y="580" width="4" height="11" fill="#224" rx="1" />
      <rect x="1" y="580" width="4" height="11" fill="#224" rx="1" />
    </g>
    <g transform="translate(866, 0)">
      <circle cx="0" cy="568" r="5" fill="#e8c8a0" />
      <rect x="-4" y="573" width="8" height="10" fill="#ff8844" rx="2" />
      <rect x="-3" y="583" width="3" height="8" fill="#445" rx="1" />
      <rect x="1" y="583" width="3" height="8" fill="#445" rx="1" />
    </g>
    {/* Customer 3 */}
    <g transform="translate(1058, 0)">
      <circle cx="0" cy="562" r="7" fill="#d4a070" />
      <rect x="-5" y="569" width="10" height="14" fill="#336699" rx="2" />
      <rect x="-4" y="583" width="4" height="11" fill="#223" rx="1" />
      <rect x="1" y="583" width="4" height="11" fill="#223" rx="1" />
    </g>
    {/* Customer 4 */}
    <g transform="translate(1378, 0)">
      <circle cx="0" cy="560" r="7" fill="#c89870" />
      <rect x="-5" y="567" width="10" height="14" fill="#ff6622" rx="2" />
      <rect x="-4" y="581" width="4" height="11" fill="#333" rx="1" />
      <rect x="1" y="581" width="4" height="11" fill="#333" rx="1" />
    </g>

    {/* Ground details */}
    <circle cx="308" cy="598" r="5.5" fill="#ff5520" opacity={0.82} />
    <line x1="308" y1="592" x2="310" y2="588" stroke="#228822" strokeWidth="1.5" />
    <rect x="548" y="584" width="17" height="19" fill="#eecc88" rx="2" />
    <path d="M 552 584 Q 556 579 562 584" fill="none" stroke="#cc9944" strokeWidth="1.5" />
  </svg>
);

// ─── Dusk: Coastal Sunset ─────────────────────────────────────────────────────
const Dusk: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="du-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#1a1040" />
        <stop offset="30%" stopColor="#6b2d8b" />
        <stop offset="58%" stopColor="#e85d20" />
        <stop offset="78%" stopColor="#f5a030" />
        <stop offset="100%" stopColor="#ffc860" />
      </linearGradient>
      <linearGradient id="du-water" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#e07828" />
        <stop offset="45%" stopColor="#b05820" />
        <stop offset="100%" stopColor="#1a3860" />
      </linearGradient>
      <filter id="du-blur">
        <feGaussianBlur stdDeviation="10" />
      </filter>
      <style>{`
        @keyframes du-bird { 0%{transform:translate(0,0)} 50%{transform:translate(30px,-8px)} 100%{transform:translate(60px,0)} }
        @keyframes du-sail { 0%,100%{transform:rotate(-2deg)} 50%{transform:rotate(2deg)} }
        @keyframes du-ripple { 0%,100%{opacity:0.15} 50%{opacity:0.35} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#du-sky)" />

    {/* Sun */}
    <ellipse cx="720" cy="490" rx="110" ry="110" fill="#ffee80" opacity="0.15" filter="url(#du-blur)" />
    <circle cx="720" cy="490" r="62" fill="#fff380" />
    <path d="M 660 530 Q 720 520 780 530 Q 760 560 720 570 Q 680 560 660 530 Z" fill="#ffdd60" opacity="0.3" />

    {/* Stars top */}
    {([[80, 45], [180, 28], [310, 52], [450, 30], [580, 58], [200, 85], [380, 70], [520, 92]] as [number, number][]).map(([cx, cy], i) => (
      <circle key={`du-star${i}`} cx={cx} cy={cy} r={1.2} fill="white" opacity={0.6} />
    ))}

    {/* Sunset clouds */}
    <ellipse cx="220" cy="200" rx="95" ry="32" fill="#ff9060" opacity="0.5" />
    <ellipse cx="285" cy="185" rx="75" ry="42" fill="#ff7a48" opacity="0.45" />
    <ellipse cx="155" cy="210" rx="65" ry="26" fill="#ff9060" opacity="0.4" />
    <ellipse cx="1050" cy="165" rx="85" ry="28" fill="#ff9060" opacity="0.48" />
    <ellipse cx="1115" cy="150" rx="68" ry="38" fill="#e87845" opacity="0.42" />

    {/* Water */}
    <rect x="0" y="528" width="1440" height="372" fill="url(#du-water)" />
    {[548, 568, 590, 615, 645].map((wy, i) => (
      <path key={`du-wv${i}`}
        d={`M 0 ${wy} Q ${180 + i * 20} ${wy - 6} ${360} ${wy} Q ${540} ${wy + 6} 720 ${wy} Q ${900} ${wy - 6} 1080 ${wy} Q 1260 ${wy + 6} 1440 ${wy}`}
        fill="none" stroke={`rgba(255,${140 - i * 18},30,${0.25 - i * 0.03})`} strokeWidth={1.5}
        style={{ animation: `du-ripple ${2 + i * 0.4}s ease-in-out ${i * 0.3}s infinite` }} />
    ))}
    {Array.from({ length: 8 }).map((_, i) => (
      <ellipse key={`du-ref${i}`} cx={720} cy={540 + i * 22} rx={20 + i * 12} ry={3} fill="#ffcc40" opacity={0.15 - i * 0.015} />
    ))}

    {/* Left buildings */}
    <rect x="0" y="350" width="145" height="182" fill="#e8ddd0" rx="4" />
    <rect x="0" y="332" width="105" height="200" fill="#ede4d8" rx="4" />
    <rect x="18" y="318" width="78" height="214" fill="#f0e6da" rx="4" />
    {([[25, 330], [55, 330], [25, 362], [55, 362], [25, 394], [55, 394], [25, 426], [55, 426], [110, 362], [110, 394], [110, 426], [110, 458], [110, 490]] as [number, number][]).map(([wx, wy], i) => (
      <rect key={`du-lw${i}`} x={wx} y={wy} width={22} height={18} fill="#ff9840" opacity={0.45} rx={2} />
    ))}
    <rect x="0" y="440" width="145" height="8" fill="#d4c8ba" rx="2" />
    <rect x="40" y="462" width="28" height="38" fill="#c05828" rx="14 14 0 0" opacity={0.5} />

    {/* Center buildings */}
    <rect x="320" y="372" width="165" height="160" fill="#ede4d8" rx="4" />
    <rect x="352" y="350" width="102" height="182" fill="#f0e6da" rx="4" />
    {([[328, 385], [368, 385], [410, 385], [328, 420], [368, 420], [410, 420], [328, 455], [368, 455], [410, 455]] as [number, number][]).map(([wx, wy], i) => (
      <rect key={`du-cw${i}`} x={wx} y={wy} width={26} height={22} fill="#ff9840" opacity={0.4} rx={2} />
    ))}

    {/* Right buildings */}
    <rect x="980" y="360" width="155" height="172" fill="#e8ddd0" rx="4" />
    <rect x="1008" y="338" width="118" height="194" fill="#ede4d8" rx="4" />
    <rect x="1120" y="374" width="145" height="158" fill="#e0d4c4" rx="4" />
    <rect x="1258" y="352" width="182" height="180" fill="#ede4d8" rx="4" />
    {([[988, 372], [1028, 372], [1068, 372], [988, 408], [1028, 408], [1068, 408], [988, 444], [1028, 444], [1068, 444], [1128, 386], [1168, 386], [1128, 422], [1168, 422], [1266, 365], [1310, 365], [1354, 365], [1266, 402], [1310, 402], [1354, 402]] as [number, number][]).map(([wx, wy], i) => (
      <rect key={`du-rw${i}`} x={wx} y={wy} width={24} height={20} fill="#ff9840" opacity={0.42} rx={2} />
    ))}

    {/* Dock */}
    <rect x="580" y="520" width="285" height="20" fill="#a06840" rx="2" />
    {Array.from({ length: 14 }).map((_, i) => (
      <rect key={`du-plank${i}`} x={582 + i * 20} y={520} width={17} height={20} fill="#9a6038" opacity={0.6} rx="1" />
    ))}
    {[595, 645, 698, 752, 805, 855].map((px, i) => (
      <g key={`du-post${i}`}>
        <rect x={px - 4} y={522} width={8} height={55} fill="#8a5030" rx="2" />
        <ellipse cx={px} cy={575} rx={6} ry={3} fill="#7a4428" />
        {i < 5 && <path d={`M ${px + 4} 535 Q ${px + 30} 530 ${px + 57} 535`} fill="none" stroke="#c09060" strokeWidth={2} opacity={0.6} />}
      </g>
    ))}
    <g transform="translate(628, 0)">
      <circle cx="0" cy="512" r="7" fill="#e8c090" />
      <rect x="-5" y="519" width="11" height="14" fill="#2255aa" rx="2" />
      <rect x="-4" y="533" width="4" height="10" fill="#334" rx="1" />
      <rect x="2" y="533" width="4" height="10" fill="#334" rx="1" />
    </g>
    <g transform="translate(808, 0)">
      <circle cx="0" cy="511" r="7" fill="#d4a878" />
      <rect x="-5" y="518" width="11" height="14" fill="#cc4422" rx="2" />
      <rect x="-4" y="532" width="4" height="10" fill="#334" rx="1" />
      <rect x="2" y="532" width="4" height="10" fill="#334" rx="1" />
    </g>

    {/* Sailboats */}
    <g style={{ animation: 'du-sail 4s ease-in-out infinite', transformOrigin: '440px 480px' }}>
      <path d="M 390 490 L 440 490 L 455 520 L 375 520 Z" fill="#e8d0a8" />
      <rect x="438" y="420" width="4" height="72" fill="#8a6030" />
      <path d="M 442 425 L 490 480 L 442 480 Z" fill="white" opacity={0.9} />
      <path d="M 438 430 L 395 478 L 438 478 Z" fill="#ff8844" opacity={0.8} />
    </g>
    <g>
      <path d="M 840 500 Q 870 498 900 500 L 888 518 L 852 518 Z" fill="#d8c898" />
      <rect x="868" y="458" width="3" height="44" fill="#7a5028" />
      <path d="M 871 462 L 900 496 L 871 496 Z" fill="white" opacity={0.85} />
      <path d="M 868 466 L 842 494 L 868 494 Z" fill="#4488cc" opacity={0.75} />
    </g>

    {/* Palm trees */}
    <g transform="translate(152, 0)">
      <path d="M 0 528 Q -5 480 8 440 Q 15 420 5 395" fill="none" stroke="#6a4820" strokeWidth={8} strokeLinecap="round" />
      {[[-35, 395, 20], [-10, 380, 10], [20, 388, -15], [32, 402, -30], [0, 370, 5]].map(([dx, dy, rot], i) => (
        <ellipse key={i} cx={5 + dx} cy={dy} rx={28} ry={8} fill="#2d6820" transform={`rotate(${rot} ${5 + dx} ${dy})`} opacity={0.88} />
      ))}
    </g>
    <g transform="translate(1328, 0)">
      <path d="M 0 528 Q 5 480 -6 440 Q -12 420 -2 395" fill="none" stroke="#6a4820" strokeWidth={7} strokeLinecap="round" />
      {[[28, 395, -20], [8, 380, -10], [-18, 388, 15], [-30, 402, 30], [-2, 372, -5]].map(([dx, dy, rot], i) => (
        <ellipse key={i} cx={-2 + dx} cy={dy} rx={26} ry={7} fill="#2d6820" transform={`rotate(${rot} ${-2 + dx} ${dy})`} opacity={0.88} />
      ))}
    </g>

    {/* Seagulls */}
    {([[220, 280], [250, 265], [275, 278], [900, 295], [930, 280], [960, 292]] as [number, number][]).map(([bx, by], i) => (
      <g key={`du-bird${i}`} style={{ animation: `du-bird ${3 + i * 0.4}s ease-in-out ${i * 0.8}s infinite` }}>
        <path d={`M ${bx - 10} ${by} Q ${bx - 5} ${by - 5} ${bx} ${by}`} fill="none" stroke="white" strokeWidth={1.8} />
        <path d={`M ${bx} ${by} Q ${bx + 5} ${by - 5} ${bx + 10} ${by}`} fill="none" stroke="white" strokeWidth={1.8} />
      </g>
    ))}

    {/* Beach umbrellas */}
    <line x1="100" y1="488" x2="115" y2="538" stroke="#8a5030" strokeWidth={3} />
    <path d="M 60 490 Q 100 470 140 490 Z" fill="#ff4433" opacity={0.85} />
    <path d="M 78 490 Q 100 478 122 490 Z" fill="white" opacity={0.7} />
    <line x1="1300" y1="492" x2="1315" y2="540" stroke="#8a5030" strokeWidth={3} />
    <path d="M 1260 494 Q 1300 474 1340 494 Z" fill="#3366cc" opacity={0.82} />
    <path d="M 1278 494 Q 1300 482 1322 494 Z" fill="white" opacity={0.65} />
  </svg>
);

// ─── Slate: Snowy Town ────────────────────────────────────────────────────────
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
        <stop offset="100%" stopColor="#c8d4e4" />
      </linearGradient>
      <filter id="sl-blur">
        <feGaussianBlur stdDeviation="8" />
      </filter>
      <style>{`
        @keyframes sl-snow1 { 0%{transform:translate(0,0)} 100%{transform:translate(-20px,900px)} }
        @keyframes sl-snow2 { 0%{transform:translate(0,0)} 100%{transform:translate(25px,900px)} }
        @keyframes sl-smoke { 0%{transform:translateY(0) scaleX(1);opacity:0.4} 100%{transform:translateY(-40px) scaleX(1.6);opacity:0} }
        @keyframes sl-walk3 { 0%{transform:translateX(0)} 100%{transform:translateX(40px)} }
        @keyframes sl-xmas { 0%,100%{opacity:0.8} 50%{opacity:0.4} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#sl-sky)" />

    {/* Snowflakes */}
    {Array.from({ length: 40 }).map((_, i) => {
      const x = (i * 97 + 30) % 1440;
      const y = (i * 43 + 10) % 400;
      const s = 0.8 + (i % 4) * 0.4;
      return (
        <g key={`sl-flake${i}`} transform={`translate(${x},${y})`}
          style={{ animation: `${i % 2 === 0 ? 'sl-snow1' : 'sl-snow2'} ${6 + (i % 6)}s linear ${(i % 8) * 0.5}s infinite` }}>
          <circle cx="0" cy="0" r={s} fill="white" opacity={0.7} />
          <line x1={-s * 3} y1="0" x2={s * 3} y2="0" stroke="white" strokeWidth={s * 0.6} opacity={0.5} />
          <line x1="0" y1={-s * 3} x2="0" y2={s * 3} stroke="white" strokeWidth={s * 0.6} opacity={0.5} />
        </g>
      );
    })}

    {/* Snow ground */}
    <rect x="0" y="620" width="1440" height="280" fill="#dce8f4" />
    <path d="M 0 620 Q 240 608 480 622 Q 720 636 960 618 Q 1200 604 1440 620 L 1440 900 L 0 900 Z" fill="#e8f2ff" />

    {/* Far left building */}
    <rect x="0" y="348" width="138" height="274" fill="#6a7898" rx="3" />
    <rect x="-4" y="340" width="146" height="14" fill="#e8f2ff" rx="3" />
    <path d="M 0 348 L 69 325 L 138 348 Z" fill="#7a88a8" />
    <rect x="15" y="333" width="108" height="20" fill="#e8f2ff" rx="2" opacity={0.9} />
    {Array.from({ length: 4 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect key={`sl-lw-${row}-${col}`} x={12 + col * 40} y={362 + row * 44} width={26} height={22} fill="#ffe898" opacity={0.38 + (row + col) % 3 * 0.08} rx={2} />
      ))
    )}
    <rect x="52" y="318" width="16" height="28" fill="#5a6888" rx="2" />
    <rect x="48" y="314" width="24" height="8" fill="#4e5c7a" rx="1" />
    {[0, 1, 2].map((i) => (
      <ellipse key={`sl-sm1-${i}`} cx={60} cy={304 - i * 12} rx={5 + i * 2} ry={4 + i} fill="#c8d4e4" opacity={0.4}
        style={{ animation: `sl-smoke ${2.5 + i * 0.5}s ease-out ${i * 0.6}s infinite` }} />
    ))}

    {/* Main left building */}
    <rect x="125" y="290" width="188" height="332" fill="#788aaa" rx="3" />
    <rect x="125" y="282" width="188" height="14" fill="#e8f2ff" rx="3" />
    <path d="M 125 290 L 219 258 L 313 290 Z" fill="#8898b8" />
    <rect x="130" y="265" width="178" height="28" fill="#e8f2ff" rx="2" opacity={0.9} />
    {Array.from({ length: 6 }).map((_, row) =>
      Array.from({ length: 4 }).map((_, col) => (
        <rect key={`sl-mw-${row}-${col}`} x={135 + col * 44} y={302 + row * 40} width={28} height={24} fill="#ffe898" opacity={0.35 + (row * col) % 4 * 0.07} rx={2} />
      ))
    )}
    <rect x="194" y="522" width="32" height="52" fill="#3a4a60" rx="3" />
    <circle cx="210" cy="538" r="12" fill="none" stroke="#448833" strokeWidth={4} opacity={0.8} />
    <circle cx="210" cy="534" r="3" fill="#cc2222" opacity={0.9} />
    <rect x="168" y="248" width="18" height="30" fill="#6a7898" rx="2" />
    <rect x="163" y="244" width="28" height="9" fill="#5e6e8a" rx="1" />
    {[0, 1, 2].map((i) => (
      <ellipse key={`sl-sm2-${i}`} cx={177} cy={234 - i * 12} rx={5 + i * 2} ry={4 + i} fill="#c8d4e4" opacity={0.38}
        style={{ animation: `sl-smoke ${3 + i * 0.5}s ease-out ${0.8 + i * 0.4}s infinite` }} />
    ))}

    {/* Center bakery */}
    <rect x="458" y="360" width="265" height="262" fill="#7a8eae" rx="3" />
    <rect x="454" y="352" width="273" height="14" fill="#e8f2ff" rx="3" />
    <path d="M 458 360 L 590 330 L 723 360 Z" fill="#8898b8" />
    <rect x="460" y="338" width="263" height="22" fill="#e8f2ff" rx="2" opacity={0.9} />
    {([[468, 374], [540, 374], [612, 374], [682, 374], [468, 424], [540, 424], [612, 424], [682, 424]] as [number, number][]).map(([wx, wy], i) => (
      <rect key={`sl-cw${i}`} x={wx} y={wy} width={38} height={32} fill="#ffe898" opacity={0.45} rx={3} />
    ))}
    <rect x="492" y="368" width="155" height="22" fill="#3a4e6a" rx="3" />
    <rect x="492" y="368" width="155" height="7" fill="#e8f2ff" rx="3 3 0 0" opacity={0.9} />
    <text x="569" y="385" textAnchor="middle" fill="white" fontSize="11" fontWeight="600" fontFamily="system-ui, sans-serif">BAKERY &amp; CAFÉ</text>
    <rect x="466" y="456" width="252" height="78" fill="#3a4e6a" rx="4" />
    {[[485, 490, '#f4c2a0'], [520, 485, '#d4a870'], [558, 490, '#e8b888'], [595, 484, '#ff9988'], [635, 490, '#d4c888'], [672, 484, '#e8a878']].map(([px, py, c], i) => (
      <ellipse key={`sl-cake${i}`} cx={px as number} cy={py as number} rx={16} ry={12} fill={c as string} opacity={0.85} />
    ))}
    <rect x="464" y="452" width="256" height="8" fill="#e8f2ff" rx="2" />
    <rect x="560" y="484" width="40" height="76" fill="#2a3a50" rx="3" />
    <circle cx="580" cy="524" r="3" fill="#aaa" />
    <rect x="548" y="322" width="20" height="32" fill="#6a7898" rx="2" />
    <rect x="543" y="318" width="30" height="9" fill="#5e6e8a" rx="1" />
    {[0, 1, 2, 3].map((i) => (
      <ellipse key={`sl-sm3-${i}`} cx={558} cy={308 - i * 11} rx={6 + i * 2} ry={4 + i} fill="#c8d4e4" opacity={0.4}
        style={{ animation: `sl-smoke ${2 + i * 0.4}s ease-out ${i * 0.3}s infinite` }} />
    ))}
    {/* Christmas lights */}
    {Array.from({ length: 18 }).map((_, i) => {
      const colors = ['#ff2222', '#22ff22', '#ffff22', '#2222ff', '#ff8822'];
      return <circle key={`sl-xmas${i}`} cx={462 + i * 14} cy={460} r={3.5} fill={colors[i % colors.length]} opacity={0.8}
        style={{ animation: `sl-xmas ${0.8 + (i % 4) * 0.3}s ease-in-out ${i * 0.1}s infinite` }} />;
    })}
    <path d="M 462 460 Q 640 456 723 460" fill="none" stroke="#888" strokeWidth={1} opacity={0.5} />

    {/* Right buildings */}
    <rect x="755" y="332" width="165" height="290" fill="#788aaa" rx="3" />
    <rect x="751" y="324" width="173" height="14" fill="#e8f2ff" rx="3" />
    <path d="M 755 332 L 837 305 L 920 332 Z" fill="#8898b8" />
    <rect x="757" y="311" width="162" height="22" fill="#e8f2ff" rx="2" opacity={0.88} />
    {Array.from({ length: 5 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect key={`sl-rw1-${row}-${col}`} x={765 + col * 50} y={346 + row * 40} width={32} height={26} fill="#ffe898" opacity={0.35 + (row + col) % 4 * 0.07} rx={2} />
      ))
    )}
    <rect x="1042" y="358" width="148" height="264" fill="#6a7898" rx="3" />
    <rect x="1038" y="350" width="156" height="12" fill="#e8f2ff" rx="3" />
    <path d="M 1042 358 L 1116 332 L 1190 358 Z" fill="#7a88a8" />
    <rect x="1044" y="338" width="148" height="20" fill="#e8f2ff" rx="2" opacity={0.88} />
    {Array.from({ length: 5 }).map((_, row) =>
      Array.from({ length: 3 }).map((_, col) => (
        <rect key={`sl-rw2-${row}-${col}`} x={1050 + col * 42} y={370 + row * 38} width={28} height={24} fill="#ffe898" opacity={0.3 + (row * col) % 4 * 0.07} rx={2} />
      ))
    )}
    <rect x="1228" y="370" width="212" height="252" fill="#788aaa" rx="3" />
    <rect x="1224" y="362" width="220" height="12" fill="#e8f2ff" rx="3" />
    <path d="M 1228 370 L 1334 340 L 1440 370 Z" fill="#8898b8" />
    <rect x="1230" y="347" width="210" height="23" fill="#e8f2ff" rx="2" opacity={0.88} />
    {Array.from({ length: 5 }).map((_, row) =>
      Array.from({ length: 4 }).map((_, col) => (
        <rect key={`sl-rw3-${row}-${col}`} x={1237 + col * 48} y={382 + row * 38} width={30} height={24} fill="#ffe898" opacity={0.33 + (row + col) % 4 * 0.06} rx={2} />
      ))
    )}

    {/* Street + lamps */}
    <rect x="0" y="618" width="1440" height="8" fill="#e8f2ff" rx="2" />
    {[240, 590, 960, 1280].map((lx, i) => (
      <g key={`sl-lamp${i}`}>
        <ellipse cx={lx + 30} cy={660} rx={45} ry={15} fill="#ffe898" opacity={0.2} filter="url(#sl-blur)" />
        <rect x={lx} y={488} width={5} height={132} fill="#4a5878" rx={2} />
        <path d={`M ${lx + 2} 488 Q ${lx + 28} 482 ${lx + 36} 470`} fill="none" stroke="#4a5878" strokeWidth={4} />
        <rect x={lx + 23} y={464} width={22} height={10} fill="#3a4868" rx={2} />
        <ellipse cx={lx + 34} cy={465} rx={8} ry={4.5} fill="#ffe060" opacity={0.85} />
        <rect x={lx - 2} y={616} width={9} height={28} fill="#e8f2ff" rx="2" />
      </g>
    ))}

    {/* Bare trees */}
    <g transform="translate(398, 0)">
      <rect x="-3" y="540" width="6" height="80" fill="#5a5a6a" />
      {[[-20, 555], [18, 548], [-12, 572], [14, 568]].map(([dx, dy], i) => (
        <line key={i} x1="0" y1={dy} x2={dx} y2={dy + 18} stroke="#5a5a6a" strokeWidth={2.5} opacity={0.8} />
      ))}
      {[[-20, 573, 14], [18, 566, 12], [-12, 590, 10], [14, 586, 10]].map(([bx, by, bw], i) => (
        <ellipse key={i} cx={bx as number} cy={by as number} rx={bw as number / 2} ry={3} fill="#e8f2ff" opacity={0.85} />
      ))}
    </g>
    <g transform="translate(1095, 0)">
      <rect x="-3" y="545" width="6" height="75" fill="#5a5a6a" />
      {[[-18, 558], [16, 552], [-10, 575], [12, 570]].map(([dx, dy], i) => (
        <line key={i} x1="0" y1={dy} x2={dx} y2={dy + 16} stroke="#5a5a6a" strokeWidth={2.5} opacity={0.8} />
      ))}
      {[[-18, 574, 12], [16, 568, 12], [-10, 591, 9], [12, 586, 10]].map(([bx, by, bw], i) => (
        <ellipse key={i} cx={bx as number} cy={by as number} rx={bw as number / 2} ry={3} fill="#e8f2ff" opacity={0.85} />
      ))}
    </g>

    {/* Snowman */}
    <g transform="translate(890, 0)">
      <ellipse cx="0" cy="666" rx="38" ry="10" fill="#d8eaf8" />
      <circle cx="0" cy="638" r="28" fill="white" opacity={0.95} />
      <circle cx="0" cy="596" r="20" fill="white" opacity={0.95} />
      <circle cx="0" cy="570" r="13" fill="white" opacity={0.95} />
      <circle cx="-5" cy="566" r="2.5" fill="#2a3040" />
      <circle cx="5" cy="566" r="2.5" fill="#2a3040" />
      <path d="M 0 570 L 8 572 L 0 574 Z" fill="#ff6622" />
      <circle cx="0" cy="590" r={2.5} fill="#2a3040" />
      <circle cx="0" cy="598" r={2.5} fill="#2a3040" />
      <circle cx="0" cy="606" r={2.5} fill="#2a3040" />
      <path d="M -18 584 Q 0 578 18 584 Q 18 590 0 586 Q -18 590 -18 584 Z" fill="#cc2222" opacity={0.9} />
      <rect x="-12" y="550" width="24" height="18" fill="#1a2030" rx="2" />
      <rect x="-16" y="566" width="32" height="5" fill="#1a2030" rx="1" />
      <line x1="-28" y1="592" x2="-50" y2="575" stroke="#5a5a6a" strokeWidth={3} />
      <line x1="-50" y1="575" x2="-58" y2="565" stroke="#5a5a6a" strokeWidth={2} />
      <line x1="28" y1="592" x2="50" y2="575" stroke="#5a5a6a" strokeWidth={3} />
      <line x1="50" y1="575" x2="58" y2="565" stroke="#5a5a6a" strokeWidth={2} />
    </g>

    {/* People in coats */}
    <g transform="translate(318, 0)">
      <circle cx="0" cy="608" r="8" fill="#e8c8a0" />
      <rect x="-7" y="616" width="14" height="18" fill="#2244aa" rx="3" />
      <rect x="-8" y="623" width="16" height="5" fill="#cc2222" rx="2" />
      <rect x="-5" y="634" width="5" height="14" fill="#1a3380" rx="1" />
      <rect x="1" y="634" width="5" height="14" fill="#1a3380" rx="1" />
    </g>
    <g style={{ animation: 'sl-walk3 9s linear infinite' }} transform="translate(748, 0)">
      <circle cx="0" cy="606" r="8" fill="#e0b890" />
      <rect x="-7" y="614" width="14" height="18" fill="#556633" rx="3" />
      <rect x="-8" y="622" width="16" height="5" fill="#aa8833" rx="2" />
      <rect x="-5" y="632" width="5" height="13" fill="#3a4422" rx="1" />
      <rect x="1" y="632" width="5" height="13" fill="#3a4422" rx="1" />
    </g>
    <g transform="translate(1182, 0)">
      <circle cx="0" cy="608" r="8" fill="#d4a878" />
      <rect x="-7" y="616" width="14" height="18" fill="#443322" rx="3" />
      <rect x="-8" y="622" width="16" height="5" fill="#884422" rx="2" />
      <rect x="-5" y="634" width="5" height="13" fill="#332211" rx="1" />
      <rect x="1" y="634" width="5" height="13" fill="#332211" rx="1" />
    </g>

    {/* Footprints */}
    {Array.from({ length: 8 }).map((_, i) => (
      <ellipse key={`sl-foot${i}`} cx={740 + i * 28} cy={650 + (i % 2) * 6} rx={5} ry={3} fill="#c8dcf0" opacity={0.7} />
    ))}
  </svg>
);

// ─── Forest: Sunny Forest Trail ────────────────────────────────────────────────
const Forest: React.FC = () => (
  <svg
    style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid slice"
    viewBox="0 0 1440 900"
  >
    <defs>
      <linearGradient id="fo-sky" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#5ab8f0" />
        <stop offset="55%" stopColor="#8dcef5" />
        <stop offset="100%" stopColor="#b8e4f8" />
      </linearGradient>
      <linearGradient id="fo-path" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#c4a870" />
        <stop offset="100%" stopColor="#a8905a" />
      </linearGradient>
      <filter id="fo-blur">
        <feGaussianBlur stdDeviation="8" />
      </filter>
      <style>{`
        @keyframes fo-sway { 0%,100%{transform:rotate(-1.5deg)} 50%{transform:rotate(1.5deg)} }
        @keyframes fo-sway2 { 0%,100%{transform:rotate(2deg)} 50%{transform:rotate(-2deg)} }
        @keyframes fo-bird { 0%{transform:translate(0,0)} 50%{transform:translate(35px,-12px)} 100%{transform:translate(70px,0)} }
        @keyframes fo-butterfly { 0%{transform:translate(0,0)} 25%{transform:translate(8px,-5px)} 50%{transform:translate(15px,0)} 75%{transform:translate(8px,5px)} 100%{transform:translate(0,0)} }
        @keyframes fo-light { 0%,100%{opacity:0.18} 50%{opacity:0.32} }
        @keyframes fo-deer { 0%{transform:translateX(0)} 100%{transform:translateX(-8px)} }
      `}</style>
    </defs>

    <rect width="1440" height="900" fill="url(#fo-sky)" />

    {/* Sun glow */}
    <circle cx="720" cy="200" r="200" fill="#f0e888" opacity="0.15" filter="url(#fo-blur)" />

    {/* Background canopy */}
    {([[0, 280, 130], [80, 260, 110], [180, 240, 140], [295, 255, 125], [400, 235, 150], [520, 248, 120],
      [625, 228, 155], [748, 240, 145], [870, 232, 138], [988, 245, 125], [1100, 235, 148],
      [1225, 248, 130], [1335, 255, 120]] as [number, number, number][]).map(([tx, ty, tw], i) => (
        <ellipse key={`fo-bg${i}`} cx={tx + tw / 2} cy={ty} rx={tw / 2} ry={tw / 3.5}
          fill={i % 2 === 0 ? '#2d7028' : '#347530'} opacity={0.7} />
      ))}

    {/* Sunbeams */}
    {[580, 660, 740, 820, 900].map((bx, i) => (
      <path key={`fo-beam${i}`} d={`M ${bx - 30} 180 L ${bx - 80} 620 L ${bx + 80} 620 L ${bx + 30} 180 Z`}
        fill="#f8f060" opacity={0.06}
        style={{ animation: `fo-light ${3 + i * 0.6}s ease-in-out ${i * 0.4}s infinite` }} />
    ))}

    {/* Ground */}
    <rect x="0" y="580" width="1440" height="320" fill="#4a8030" />
    <rect x="0" y="580" width="1440" height="60" fill="#558838" />
    <rect x="0" y="640" width="1440" height="260" fill="#448028" />

    {/* Trail */}
    <path d="M 560 580 Q 590 640 620 720 Q 650 800 670 900 L 780 900 Q 800 800 790 720 Q 780 640 820 580 Z" fill="url(#fo-path)" opacity={0.85} />
    {Array.from({ length: 6 }).map((_, i) => (
      <ellipse key={`fo-pebble${i}`} cx={640 + (i % 3) * 18 - 12} cy={650 + i * 42} rx={5} ry={3} fill="#c4a060" opacity={0.6} transform={`rotate(${i * 22} ${640 + (i % 3) * 18 - 12} ${650 + i * 42})`} />
    ))}

    {/* Main trees - Big oaks */}
    <g style={{ animation: 'fo-sway 7s ease-in-out infinite', transformOrigin: '155px 580px' }}>
      <path d="M 150 575 Q 125 578 115 585" fill="none" stroke="#5a3a18" strokeWidth={5} strokeLinecap="round" />
      <path d="M 162 578 Q 182 580 192 588" fill="none" stroke="#5a3a18" strokeWidth={4} strokeLinecap="round" />
      <rect x="150" y="445" width="12" height="135" fill="#6a4820" rx="3" />
      <circle cx="156" cy="400" r="68" fill="#3a8030" />
      <circle cx="156" cy="375" r="52" fill="#448838" />
      <circle cx="128" cy="415" r="45" fill="#3d8530" />
      <circle cx="184" cy="412" r="42" fill="#388030" />
      <circle cx="145" cy="358" r="36" fill="#4a9038" />
      <circle cx="170" cy="375" r="22" fill="#58a840" opacity={0.5} />
    </g>
    <g style={{ animation: 'fo-sway2 9s ease-in-out 1s infinite', transformOrigin: '285px 580px' }}>
      <rect x="280" y="468" width="10" height="112" fill="#5a4018" rx="2" />
      <circle cx="285" cy="430" r="58" fill="#357830" />
      <circle cx="285" cy="408" r="45" fill="#3d8835" />
      <circle cx="262" cy="445" r="38" fill="#308028" />
      <circle cx="308" cy="442" r="36" fill="#357830" />
      <circle cx="275" cy="392" r="30" fill="#3e9035" />
      <circle cx="295" cy="395" r="18" fill="#50a840" opacity={0.5} />
    </g>
    <g style={{ animation: 'fo-sway 8s ease-in-out 2s infinite', transformOrigin: '1165px 580px' }}>
      <path d="M 1155 580 Q 1132 582 1122 590" fill="none" stroke="#5a3a18" strokeWidth={4} strokeLinecap="round" />
      <path d="M 1172 582 Q 1192 585 1200 590" fill="none" stroke="#5a3a18" strokeWidth={4} strokeLinecap="round" />
      <rect x="1160" y="440" width="12" height="140" fill="#6a4820" rx="3" />
      <circle cx="1166" cy="395" r="72" fill="#3a8030" />
      <circle cx="1166" cy="368" r="56" fill="#448838" />
      <circle cx="1138" cy="412" r="48" fill="#3d8530" />
      <circle cx="1194" cy="410" r="44" fill="#388030" />
      <circle cx="1155" cy="350" r="38" fill="#4a9038" />
      <circle cx="1178" cy="370" r="24" fill="#58a840" opacity={0.48} />
    </g>
    <g style={{ animation: 'fo-sway2 6.5s ease-in-out 3s infinite', transformOrigin: '1292px 580px' }}>
      <rect x="1288" y="472" width="10" height="108" fill="#5a4018" rx="2" />
      <circle cx="1293" cy="432" r="60" fill="#368030" />
      <circle cx="1293" cy="410" r="47" fill="#3e8838" />
      <circle cx="1270" cy="448" r="40" fill="#318030" />
      <circle cx="1316" cy="445" r="38" fill="#368030" />
      <circle cx="1283" cy="395" r="32" fill="#409038" />
      <circle cx="1303" cy="400" r="20" fill="#50a840" opacity={0.48} />
    </g>

    {/* Mid-ground trees */}
    {([[390, 488, 48, '#357a28'], [480, 495, 42, '#2d7028'], [960, 490, 46, '#357a28'], [1050, 485, 52, '#2d7028']] as [number, number, number, string][]).map(([tx, ty, tr, tc], i) => (
      <g key={`fo-mt${i}`} style={{ animation: i % 2 === 0 ? 'fo-sway 7s ease-in-out 1s infinite' : 'fo-sway2 8s ease-in-out 2s infinite', transformOrigin: `${tx}px 580px` }}>
        <rect x={tx - 5} y={ty + 45} width={10} height={90} fill="#5a4018" rx="2" />
        <circle cx={tx} cy={ty + 20} r={tr} fill={tc} />
        <circle cx={tx} cy={ty} r={tr * 0.7} fill={tc} opacity={0.9} />
        <circle cx={tx - tr * 0.5} cy={ty + 25} r={tr * 0.55} fill={tc} opacity={0.85} />
        <circle cx={tx + tr * 0.5} cy={ty + 22} r={tr * 0.5} fill={tc} opacity={0.85} />
      </g>
    ))}

    {/* Flowers */}
    {([[120, 598, '#ffdd22'], [142, 602, '#ff6688'], [168, 596, '#ff6688'], [192, 600, '#ffdd22'],
      [1252, 596, '#ffdd22'], [1276, 600, '#ff6688'], [1298, 596, '#ffaa33'], [1322, 602, '#ff6688'],
      [440, 590, '#ff6688'], [462, 594, '#ffdd22'], [484, 590, '#ff6688'],
      [972, 590, '#ffdd22'], [996, 594, '#ff6688'], [1020, 590, '#ffaa33']] as [number, number, string][]).map(([fx, fy, fc], i) => (
        <g key={`fo-flower${i}`}>
          <line x1={fx} y1={fy} x2={fx} y2={fy + 12} stroke="#3a7828" strokeWidth={1.5} />
          <circle cx={fx} cy={fy} r={5} fill={fc} opacity={0.9} />
          <circle cx={fx} cy={fy} r={2.5} fill="white" opacity={0.7} />
        </g>
      ))}

    {/* Mushrooms */}
    {([[200, 608, '#cc3322', 14], [218, 612, '#cc3322', 10], [1230, 606, '#dd4422', 12]] as [number, number, string, number][]).map(([mx, my, mc, mr], i) => (
      <g key={`fo-mush${i}`}>
        <rect x={(mx as number) - 3} y={my as number} width={6} height={14} fill="#e8d8b0" rx="1" />
        <ellipse cx={mx as number} cy={my as number} rx={mr as number} ry={(mr as number) * 0.55} fill={mc as string} />
        {Array.from({ length: 3 }).map((_, j) => (
          <ellipse key={j} cx={(mx as number) - (mr as number) * 0.5 + j * (mr as number) * 0.5} cy={(my as number) + 2} rx={2} ry={1.2} fill="white" opacity={0.6} />
        ))}
      </g>
    ))}

    {/* Ferns */}
    {([[340, 590], [360, 594], [380, 590], [1068, 590], [1088, 592], [1108, 590]] as [number, number][]).map(([fx, fy], i) => (
      <g key={`fo-fern${i}`}>
        {[[-14, -10], [-7, -16], [0, -18], [7, -16], [14, -10]].map(([dx, dy], j) => (
          <ellipse key={j} cx={fx + dx} cy={fy + dy} rx={6} ry={3.5} fill="#2a7020"
            transform={`rotate(${dx * 2} ${fx + dx} ${fy + dy})`} opacity={0.85} />
        ))}
      </g>
    ))}

    {/* Deer */}
    <g style={{ animation: 'fo-deer 3s ease-in-out infinite alternate' }} transform="translate(680, 0)">
      <ellipse cx="0" cy="560" rx="30" ry="18" fill="#8a5a30" />
      <ellipse cx="22" cy="546" rx="10" ry="15" fill="#8a5a30" />
      <ellipse cx="28" cy="534" rx="13" ry="10" fill="#8a5a30" />
      <ellipse cx="35" cy="527" rx="5" ry="8" fill="#8a5a30" />
      <ellipse cx="35" cy="527" rx="3" ry="5" fill="#c87860" />
      <circle cx="32" cy="532" r="2.5" fill="#1a1a1a" />
      <circle cx="33" cy="531" r="1" fill="white" opacity={0.8} />
      <path d="M 30 524 L 35 510 L 42 502 M 35 510 L 28 504" fill="none" stroke="#6a4020" strokeWidth={2.5} strokeLinecap="round" />
      {[[-18, 578], [-8, 578], [8, 578], [18, 578]].map(([lx, ly], i) => (
        <line key={i} x1={lx} y1={574} x2={lx - (i < 2 ? 2 : -2)} y2={ly} stroke="#6a4020" strokeWidth={4} strokeLinecap="round" />
      ))}
      <ellipse cx="-28" cy="556" rx="6" ry="4" fill="white" opacity={0.9} />
    </g>

    {/* Birds */}
    {([[380, 240], [415, 228], [445, 238], [900, 215], [935, 202], [962, 218]] as [number, number][]).map(([bx, by], i) => (
      <g key={`fo-bird${i}`} style={{ animation: `fo-bird ${4 + i * 0.5}s ease-in-out ${i * 0.7}s infinite` }}>
        <path d={`M ${bx - 10} ${by} Q ${bx - 4} ${by - 6} ${bx} ${by}`} fill="none" stroke="#2a2a3a" strokeWidth={1.8} />
        <path d={`M ${bx} ${by} Q ${bx + 4} ${by - 6} ${bx + 10} ${by}`} fill="none" stroke="#2a2a3a" strokeWidth={1.8} />
      </g>
    ))}

    {/* Butterflies */}
    {[[540, 520], [480, 530], [1000, 515]].map(([btx, bty], i) => (
      <g key={`fo-bf${i}`} style={{ animation: `fo-butterfly ${2.5 + i}s ease-in-out ${i * 0.6}s infinite`, transformOrigin: `${btx}px ${bty}px` }}>
        <ellipse cx={btx - 8} cy={bty} rx={9} ry={6} fill={i === 0 ? '#ff8844' : i === 1 ? '#ee44cc' : '#44aaff'} opacity={0.85} transform={`rotate(20 ${btx - 8} ${bty})`} />
        <ellipse cx={btx + 8} cy={bty} rx={9} ry={6} fill={i === 0 ? '#ff8844' : i === 1 ? '#ee44cc' : '#44aaff'} opacity={0.85} transform={`rotate(-20 ${btx + 8} ${bty})`} />
        <ellipse cx={btx - 5} cy={bty + 4} rx={5} ry={3.5} fill={i === 0 ? '#ffaa66' : i === 1 ? '#ff66dd' : '#66bbff'} opacity={0.75} transform={`rotate(30 ${btx - 5} ${bty + 4})`} />
        <ellipse cx={btx + 5} cy={bty + 4} rx={5} ry={3.5} fill={i === 0 ? '#ffaa66' : i === 1 ? '#ff66dd' : '#66bbff'} opacity={0.75} transform={`rotate(-30 ${btx + 5} ${bty + 4})`} />
        <line x1={btx - 3} y1={bty - 5} x2={btx - 10} y2={bty - 14} stroke="#2a2a3a" strokeWidth={1} />
        <line x1={btx + 3} y1={bty - 5} x2={btx + 10} y2={bty - 14} stroke="#2a2a3a" strokeWidth={1} />
      </g>
    ))}

    {/* Hikers on trail */}
    <g transform="translate(630, 0)">
      <circle cx="0" cy="574" r="7" fill="#e0b888" />
      <rect x="-8" y="560" width="16" height="7" fill="#aa6622" rx="2" />
      <rect x="-5" y="556" width="10" height="7" fill="#993311" rx="2" />
      <rect x="-6" y="581" width="12" height="14" fill="#338844" rx="2" />
      <rect x="-5" y="595" width="4" height="12" fill="#1a2a1a" rx="1" />
      <rect x="2" y="595" width="4" height="12" fill="#1a2a1a" rx="1" />
      <rect x="-12" y="582" width="8" height="12" fill="#ee8833" rx="2" />
      <line x1="10" y1="572" x2="16" y2="610" stroke="#8a6030" strokeWidth={2.5} strokeLinecap="round" />
    </g>
    <g transform="translate(752, 0)">
      <circle cx="0" cy="568" r="6" fill="#d4a070" />
      <rect x="-5" y="574" width="10" height="12" fill="#4455aa" rx="2" />
      <rect x="-4" y="586" width="4" height="10" fill="#223" rx="1" />
      <rect x="1" y="586" width="4" height="10" fill="#223" rx="1" />
      <rect x="-10" y="576" width="7" height="10" fill="#cc4422" rx="2" />
      <line x1="8" y1="566" x2="14" y2="598" stroke="#8a6030" strokeWidth={2} strokeLinecap="round" />
    </g>
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
