import React from 'react';

const DOOM_URL = 'https://archive.org/embed/doom-1993';

export default function DoomWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        background: '#000',
        color: '#fff',
        fontFamily: 'monospace',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 24,
        userSelect: 'none',
      }}
    >
      <pre
        style={{
          color: '#b91c1c',
          fontSize: 10,
          lineHeight: 1.2,
          letterSpacing: 1,
          textAlign: 'center',
          margin: 0,
        }}
      >
        {[
          ' ####   #####   #####  ##   ##',
          ' ## ##  ##  ##  ## ##  ### ###',
          ' ##  ## #####   ##  ## ## # ##',
          ' ## ##  ##  ##  ## ##  ##   ##',
          ' ####   ##  ##  #####  ##   ##',
        ].join('\n')}
      </pre>

      <div style={{ color: 'rgba(255,255,255,0.35)', fontSize: 11, letterSpacing: 2 }}>
        ID SOFTWARE 1993
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 10 }}>
        <a
          href={DOOM_URL}
          target="_blank"
          rel="noreferrer"
          style={{
            display: 'inline-block',
            padding: '10px 28px',
            background: '#b91c1c',
            color: '#fff',
            fontFamily: 'monospace',
            fontSize: 13,
            fontWeight: 700,
            letterSpacing: 2,
            textDecoration: 'none',
            borderRadius: 4,
          }}
        >
          PLAY DOOM
        </a>
        <span style={{ color: 'rgba(255,255,255,0.2)', fontSize: 10 }}>
          Opens in a new tab (Internet Archive)
        </span>
      </div>

      <div
        style={{
          borderTop: '1px solid rgba(255,255,255,0.07)',
          paddingTop: 20,
          color: 'rgba(255,255,255,0.25)',
          fontSize: 10,
          letterSpacing: 1,
          textAlign: 'center',
          lineHeight: 2,
        }}
      >
        WASD / ARROW KEYS TO MOVE &nbsp; CTRL TO SHOOT &nbsp; SPACE TO USE
      </div>
    </div>
  );
}
