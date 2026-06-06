import React from 'react';
import { Monitor, Sun, Moon } from 'lucide-react';
import { useDesktopStore, ColorMode, Wallpaper } from '../../stores/useDesktopStore';

const COLOR_MODES: {
  key: ColorMode;
  label: string;
  icon: React.FC<{ size?: number; color?: string; strokeWidth?: number }>;
}[] = [
  { key: 'system', label: 'System', icon: Monitor },
  { key: 'light', label: 'Light', icon: Sun },
  { key: 'dark', label: 'Dark', icon: Moon },
];

const WALLPAPERS: { key: Wallpaper; label: string; color: string }[] = [
  { key: 'midnight', label: 'Night City', color: 'linear-gradient(135deg, #040a18 0%, #0e1b38 60%, #ff6b35 100%)' },
  { key: 'aurora', label: 'Street Market', color: 'linear-gradient(135deg, #4aaee8 0%, #7cc8f0 50%, #d4c4a0 100%)' },
  { key: 'dusk', label: 'Coastal Sunset', color: 'linear-gradient(135deg, #1a1040 0%, #e85d20 60%, #1a3860 100%)' },
  { key: 'slate', label: 'Snowy Town', color: 'linear-gradient(135deg, #8898b8 0%, #aab8cc 50%, #e8f2ff 100%)' },
  { key: 'forest', label: 'Forest Trail', color: 'linear-gradient(135deg, #5ab8f0 0%, #448028 50%, #4a8030 100%)' },
];

export default function SettingsWindow(): React.JSX.Element {
  const {
    colorMode,
    setColorMode,
    wallpaper,
    setWallpaper,
    animationsEnabled,
    setAnimationsEnabled,
    setExperience,
  } = useDesktopStore();

  return (
    <div
      style={{
        height: '100%',
        background: '#151515',
        padding: '24px',
        display: 'flex',
        flexDirection: 'column',
        gap: 28,
        overflowY: 'auto',
      }}
    >
      {/* Color mode */}
      <section>
        <div
          style={{
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: '0.1em',
            color: 'rgba(255,255,255,0.35)',
            textTransform: 'uppercase',
            marginBottom: 12,
          }}
        >
          Color Mode
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          {COLOR_MODES.map((m) => {
            const Icon = m.icon;
            return (
              <button
                key={m.key}
                onClick={() => setColorMode(m.key)}
                style={{
                  flex: 1,
                  padding: '10px 0',
                  border: `1px solid ${colorMode === m.key ? 'rgba(232,230,225,0.5)' : 'rgba(255,255,255,0.1)'}`,
                  background:
                    colorMode === m.key ? 'rgba(232,230,225,0.08)' : 'rgba(255,255,255,0.03)',
                  borderRadius: 7,
                  color: colorMode === m.key ? '#e8e6e1' : 'rgba(255,255,255,0.45)',
                  fontSize: 12,
                  fontWeight: 500,
                  cursor: 'pointer',
                  transition: 'all 0.15s',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  gap: 6,
                }}
              >
                <Icon
                  size={14}
                  color={colorMode === m.key ? '#e8e6e1' : 'rgba(255,255,255,0.4)'}
                  strokeWidth={1.75}
                />
                {m.label}
              </button>
            );
          })}
        </div>
      </section>

      {/* Wallpaper */}
      <section>
        <div
          style={{
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: '0.1em',
            color: 'rgba(255,255,255,0.35)',
            textTransform: 'uppercase',
            marginBottom: 12,
          }}
        >
          Wallpaper
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          {WALLPAPERS.map((w) => (
            <button
              key={w.key}
              onClick={() => setWallpaper(w.key)}
              title={w.label}
              style={{
                flex: 1,
                height: 44,
                border: `2px solid ${wallpaper === w.key ? '#e8e6e1' : 'rgba(255,255,255,0.1)'}`,
                borderRadius: 7,
                background: w.color,
                cursor: 'pointer',
                transition: 'border-color 0.15s',
              }}
            />
          ))}
        </div>
        <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)', marginTop: 6 }}>
          {WALLPAPERS.find((w) => w.key === wallpaper)?.label}
        </div>
      </section>

      {/* Animations */}
      <section>
        <div
          style={{
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: '0.1em',
            color: 'rgba(255,255,255,0.35)',
            textTransform: 'uppercase',
            marginBottom: 12,
          }}
        >
          Animations
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          {[
            { v: true, label: 'Enabled' },
            { v: false, label: 'Disabled' },
          ].map(({ v, label }) => (
            <button
              key={String(v)}
              onClick={() => setAnimationsEnabled(v)}
              style={{
                flex: 1,
                padding: '8px 0',
                border: `1px solid ${animationsEnabled === v ? 'rgba(232,230,225,0.5)' : 'rgba(255,255,255,0.1)'}`,
                background:
                  animationsEnabled === v ? 'rgba(232,230,225,0.08)' : 'rgba(255,255,255,0.03)',
                borderRadius: 7,
                color: animationsEnabled === v ? '#e8e6e1' : 'rgba(255,255,255,0.45)',
                fontSize: 12,
                fontWeight: 500,
                cursor: 'pointer',
                transition: 'all 0.15s',
              }}
            >
              {label}
            </button>
          ))}
        </div>
      </section>

      {/* Experience */}
      <section>
        <div
          style={{
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: '0.1em',
            color: 'rgba(255,255,255,0.35)',
            textTransform: 'uppercase',
            marginBottom: 12,
          }}
        >
          Experience
        </div>
        <button
          onClick={() => setExperience('website')}
          style={{
            width: '100%',
            padding: '10px 0',
            border: '1px solid rgba(255,255,255,0.1)',
            background: 'rgba(255,255,255,0.03)',
            borderRadius: 7,
            color: 'rgba(255,255,255,0.6)',
            fontSize: 12,
            fontWeight: 500,
            cursor: 'pointer',
          }}
        >
          Switch to Website Mode
        </button>
      </section>
    </div>
  );
}
