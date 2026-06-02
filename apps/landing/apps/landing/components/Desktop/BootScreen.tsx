import React from 'react';
import { motion } from 'motion/react';

export const BootScreen: React.FC = () => (
  <motion.div
    initial={{ opacity: 1 }}
    exit={{ opacity: 0 }}
    transition={{ duration: 0.5, ease: 'easeInOut' }}
    style={{
      position: 'fixed',
      inset: 0,
      background: '#0a0a0f',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 99999,
      fontFamily: "'Inter', system-ui, sans-serif",
    }}
  >
    <span
      style={{
        fontSize: 36,
        fontWeight: 700,
        letterSpacing: '-0.04em',
        color: '#e8e6e1',
        marginBottom: 48,
      }}
    >
      rtylr
    </span>

    <Spinner />
  </motion.div>
);

const Spinner: React.FC = () => {
  const COUNT = 12;
  return (
    <svg width={32} height={32} viewBox="0 0 32 32">
      {Array.from({ length: COUNT }).map((_, i) => {
        const angle = (i / COUNT) * 360;
        const opacity = (i + 1) / COUNT;
        const rad = (angle * Math.PI) / 180;
        const x1 = 16 + Math.sin(rad) * 8;
        const y1 = 16 - Math.cos(rad) * 8;
        const x2 = 16 + Math.sin(rad) * 12;
        const y2 = 16 - Math.cos(rad) * 12;
        return (
          <line
            key={i}
            x1={x1}
            y1={y1}
            x2={x2}
            y2={y2}
            stroke="#e8e6e1"
            strokeWidth={2}
            strokeLinecap="round"
            style={{
              opacity,
              animation: `rtylr-spin 1s linear infinite`,
              animationDelay: `${-(i / COUNT)}s`,
            }}
          />
        );
      })}
      <style>{`
        @keyframes rtylr-spin {
          0%   { opacity: 0.12 }
          100% { opacity: 1 }
        }
      `}</style>
    </svg>
  );
};
