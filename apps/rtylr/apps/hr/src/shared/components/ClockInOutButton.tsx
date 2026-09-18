import React from 'react';
import { Play, Square, MapPin } from 'lucide-react';
import { ACCENT } from './ui';

interface Props {
  clockedIn: boolean;
  onToggle: () => void;
  location?: string;
  loading?: boolean;
}

const ClockInOutButton: React.FC<Props> = ({ clockedIn, onToggle, location, loading }) => (
  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8 }}>
    <button
      type="button"
      onClick={onToggle}
      disabled={loading}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        padding: '14px 28px',
        borderRadius: 14,
        background: clockedIn ? 'rgba(239,68,68,0.12)' : `${ACCENT}22`,
        border: `1px solid ${clockedIn ? 'rgba(239,68,68,0.4)' : ACCENT}55`,
        color: clockedIn ? '#ef4444' : ACCENT,
        fontSize: 14,
        fontWeight: 700,
        cursor: loading ? 'not-allowed' : 'pointer',
        transition: 'transform 0.15s, background 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1.02)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLButtonElement).style.transform = 'scale(1)';
      }}
    >
      {clockedIn ? <Square size={16} /> : <Play size={16} />}
      {clockedIn ? 'Clock Out' : 'Clock In'}
    </button>
    {location && (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 4,
          fontSize: 10,
          color: 'var(--hr-text-muted)',
        }}
      >
        <MapPin size={10} />
        {location}
      </div>
    )}
  </div>
);

export default ClockInOutButton;
