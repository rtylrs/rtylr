import React, { useState } from 'react';
import { Star } from 'lucide-react';

interface Props {
  value: number;
  onChange?: (v: number) => void;
  size?: number;
  readOnly?: boolean;
}

const RatingStars: React.FC<Props> = ({ value, onChange, size = 18, readOnly = false }) => {
  const [hover, setHover] = useState(0);
  const display = hover || value;
  return (
    <div style={{ display: 'inline-flex', gap: 2 }}>
      {[1, 2, 3, 4, 5].map((i) => {
        const filled = i <= display;
        return (
          <button
            key={i}
            type="button"
            disabled={readOnly}
            onClick={() => !readOnly && onChange?.(i)}
            onMouseEnter={() => !readOnly && setHover(i)}
            onMouseLeave={() => !readOnly && setHover(0)}
            style={{
              background: 'transparent',
              border: 'none',
              padding: 2,
              cursor: readOnly ? 'default' : 'pointer',
              color: filled ? '#f59e0b' : 'var(--hr-text-dim)',
            }}
          >
            <Star size={size} fill={filled ? '#f59e0b' : 'transparent'} />
          </button>
        );
      })}
    </div>
  );
};

export default RatingStars;
