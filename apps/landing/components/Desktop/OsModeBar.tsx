import React from 'react';
import { Monitor } from 'lucide-react';
import { useDesktopStore } from '../../stores/useDesktopStore';

export const OsModeBar: React.FC = () => {
  const setExperience = useDesktopStore((s) => s.setExperience);

  return (
    <div
      style={{
        position: 'fixed',
        bottom: 20,
        right: 20,
        zIndex: 9999,
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '8px 14px',
        background: 'rgba(22,22,24,0.95)',
        border: '1px solid rgba(255,255,255,0.12)',
        borderRadius: 8,
        boxShadow: '0 8px 32px rgba(0,0,0,0.4)',
        backdropFilter: 'blur(12px)',
        cursor: 'pointer',
      }}
      onClick={() => setExperience('os')}
      onMouseEnter={(e) =>
        ((e.currentTarget as HTMLDivElement).style.background = 'rgba(40,40,44,0.97)')
      }
      onMouseLeave={(e) =>
        ((e.currentTarget as HTMLDivElement).style.background = 'rgba(22,22,24,0.95)')
      }
    >
      <Monitor size={13} color="rgba(232,230,225,0.7)" strokeWidth={1.75} />
      <span style={{ fontSize: 12, fontWeight: 500, color: 'rgba(232,230,225,0.7)' }}>
        Switch to OS mode
      </span>
    </div>
  );
};
