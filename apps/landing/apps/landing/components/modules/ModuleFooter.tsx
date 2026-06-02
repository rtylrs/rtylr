import React from 'react';

interface ModuleFooterProps {
  href: string;
  label: string;
}

export const ModuleFooter: React.FC<ModuleFooterProps> = ({ href, label }) => (
  <div
    style={{
      padding: '10px 16px',
      borderTop: '1px solid rgba(255,255,255,0.07)',
      display: 'flex',
      justifyContent: 'flex-end',
      background: 'rgba(16,16,18,0.9)',
      flexShrink: 0,
    }}
  >
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      style={{
        fontSize: 12,
        fontWeight: 600,
        color: 'rgba(232,230,225,0.85)',
        textDecoration: 'none',
        padding: '5px 12px',
        border: '1px solid rgba(255,255,255,0.15)',
        borderRadius: 6,
        transition: 'background 0.15s, border-color 0.15s',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLAnchorElement).style.background = 'rgba(255,255,255,0.08)';
        (e.currentTarget as HTMLAnchorElement).style.borderColor = 'rgba(255,255,255,0.3)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLAnchorElement).style.background = 'transparent';
        (e.currentTarget as HTMLAnchorElement).style.borderColor = 'rgba(255,255,255,0.15)';
      }}
    >
      {label}
    </a>
  </div>
);
