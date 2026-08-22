import React from 'react';
import { ScheduleShift } from '../services/hrExtras';

interface Props {
  shift: ScheduleShift;
  onClick?: () => void;
}

const ShiftCell: React.FC<Props> = ({ shift, onClick }) => {
  const color = shift.color ?? '#8b5cf6';
  return (
    <button
      type="button"
      onClick={onClick}
      style={{
        width: '100%',
        padding: '6px 8px',
        borderRadius: 8,
        background: `${color}22`,
        border: `1px solid ${color}55`,
        color: 'var(--hr-text)',
        fontSize: 11,
        textAlign: 'left',
        cursor: onClick ? 'pointer' : 'default',
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
        borderLeft: `3px solid ${color}`,
      }}
    >
      <div style={{ fontWeight: 600, color: 'var(--hr-text)' }}>{shift.staffName}</div>
      <div style={{ color: 'var(--hr-text-muted)', fontSize: 10 }}>
        {shift.start} - {shift.end} • {shift.role}
      </div>
    </button>
  );
};

export default ShiftCell;
