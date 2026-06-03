import React from 'react';

const DOOM_URL = 'https://archive.org/embed/doom-1993';

export default function DoomWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        background: '#000',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <iframe
        src={DOOM_URL}
        title="DOOM (1993)"
        style={{ flex: 1, width: '100%', border: 'none', display: 'block', background: '#000' }}
        allow="autoplay; fullscreen; gamepad; keyboard-map"
        allowFullScreen
      />
      <div
        style={{
          flexShrink: 0,
          background: '#0a0a0a',
          borderTop: '1px solid rgba(255,255,255,0.08)',
          color: 'rgba(255,255,255,0.4)',
          fontFamily: 'monospace',
          fontSize: 10,
          letterSpacing: 1,
          textAlign: 'center',
          padding: '8px 12px',
          userSelect: 'none',
        }}
      >
        CLICK THE SCREEN TO START &nbsp;·&nbsp; ARROWS / WASD MOVE &nbsp;·&nbsp; CTRL SHOOT
        &nbsp;·&nbsp; SPACE USE
      </div>
    </div>
  );
}
