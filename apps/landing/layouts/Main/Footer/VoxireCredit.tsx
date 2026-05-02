import { useEffect, useRef, useState } from 'react';

const PHRASE = 'Designed & developed by';
const SERVICES = ['Web development', 'Digital marketing', 'UI/UX design', 'Mobile apps', 'SEO'];

export function VoxireCredit() {
  const [typed, setTyped] = useState('');
  const [cursorVisible, setCursorVisible] = useState(true);
  const [hovered, setHovered] = useState(false);
  const idxRef = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    timerRef.current = setTimeout(function tick() {
      if (idxRef.current <= PHRASE.length) {
        setTyped(PHRASE.slice(0, idxRef.current));
        idxRef.current++;
        timerRef.current = setTimeout(tick, 48 + Math.random() * 22);
      } else {
        setTimeout(() => setCursorVisible(false), 500);
      }
    }, 900);
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  return (
    <div
      style={{ position: 'relative' }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <div
        aria-hidden="true"
        style={{
          position: 'absolute',
          bottom: 'calc(100% + 10px)',
          left: '50%',
          transform: hovered
            ? 'translateX(-50%) translateY(0)'
            : 'translateX(-50%) translateY(6px)',
          width: 280,
          background: '#14141d',
          border: '0.5px solid rgba(99,198,189,0.3)',
          borderRadius: 12,
          padding: '14px 16px',
          opacity: hovered ? 1 : 0,
          pointerEvents: hovered ? 'auto' : 'none',
          transition: 'opacity 0.2s ease, transform 0.2s ease',
          zIndex: 50,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 10 }}>
          <img
            src="https://voxire.com/assets/images/Voxire-logo-enblm-m.png"
            alt="Voxire"
            style={{ height: 18, width: 'auto' }}
          />
          <span style={{ fontSize: 10.5, color: 'rgba(255,255,255,0.38)' }}>
            Digital agency · Beirut, Lebanon
          </span>
        </div>
        <div style={{ height: '0.5px', background: 'rgba(99,198,189,0.15)', margin: '10px 0' }} />
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 5, marginBottom: 12 }}>
          {SERVICES.map((s) => (
            <span
              key={s}
              style={{
                fontSize: 10,
                padding: '3px 8px',
                borderRadius: 20,
                border: '0.5px solid rgba(99,198,189,0.3)',
                color: '#63C6BD',
                background: 'rgba(99,198,189,0.06)',
                whiteSpace: 'nowrap',
              }}
            >
              {s}
            </span>
          ))}
        </div>
        <a
          href="https://voxire.com/get-a-quote/"
          target="_blank"
          rel="noopener noreferrer"
          style={{
            display: 'block',
            width: '100%',
            padding: '8px',
            background: 'linear-gradient(135deg, #63c6bd 0%, #519ad1 50%, #5a4398 100%)',
            color: '#fff',
            fontSize: 11.5,
            fontWeight: 600,
            textAlign: 'center',
            textDecoration: 'none',
            borderRadius: 8,
          }}
        >
          Get a free quote →
        </a>
        <div
          style={{
            position: 'absolute',
            bottom: -5,
            left: '50%',
            transform: 'translateX(-50%) rotate(45deg)',
            width: 9,
            height: 9,
            background: '#14141d',
            borderRight: '0.5px solid rgba(99,198,189,0.3)',
            borderBottom: '0.5px solid rgba(99,198,189,0.3)',
          }}
        />
      </div>
      <div
        style={{
          background: '#0c0c11',
          borderTop: hovered
            ? '1px solid rgba(99,198,189,0.45)'
            : '1px solid rgba(99,198,189,0.18)',
          transition: 'border-color 0.3s',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 8,
          padding: '9px 24px',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <img
          src="https://voxire.com/assets/images/Voxire-logo-enblm-m.png"
          alt="Voxire"
          style={{
            height: 15,
            width: 'auto',
            opacity: hovered ? 1 : 0.7,
            transition: 'opacity 0.25s',
          }}
        />
        <span style={{ color: 'rgba(99,198,189,0.3)', fontSize: 11 }}>·</span>
        <span
          style={{
            fontSize: 11.5,
            color: hovered ? 'rgba(255,255,255,0.65)' : 'rgba(255,255,255,0.4)',
            transition: 'color 0.25s',
            whiteSpace: 'nowrap',
          }}
        >
          {typed}
          {cursorVisible && (
            <span
              style={{
                display: 'inline-block',
                width: 1.5,
                height: 11,
                background: '#63C6BD',
                marginLeft: 1,
                verticalAlign: 'middle',
                animation: 'vx-blink 0.85s step-end infinite',
              }}
            />
          )}
        </span>
        <a
          href="https://voxire.com"
          target="_blank"
          rel="noopener noreferrer"
          style={{
            color: '#63C6BD',
            textDecoration: 'none',
            fontSize: 11.5,
            fontWeight: 600,
            display: 'inline-flex',
            alignItems: 'center',
            gap: hovered ? 6 : 3,
            transition: 'gap 0.2s',
          }}
        >
          Voxire{' '}
          <span
            style={{
              display: 'inline-block',
              transform: hovered ? 'translateX(3px)' : 'translateX(0)',
              transition: 'transform 0.2s',
            }}
          >
            →
          </span>
        </a>
        <div
          style={{
            position: 'absolute',
            bottom: 0,
            left: '50%',
            transform: 'translateX(-50%)',
            width: hovered ? '70%' : 0,
            height: 1,
            background:
              'linear-gradient(90deg, transparent, #63c6bd 40%, #519ad1 60%, transparent)',
            transition: 'width 0.5s ease',
          }}
        />
        <style>{`@keyframes vx-blink { 0%,100%{opacity:1} 50%{opacity:0} }`}</style>
      </div>
    </div>
  );
}
