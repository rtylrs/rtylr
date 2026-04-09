import React from 'react';

export interface AuthDividerProps {
  label?: string;
}

const AuthDivider: React.FC<AuthDividerProps> = ({ label = 'or' }) => {
  return (
    <div className="flex items-center gap-3" role="separator">
      <div className="flex-1 h-px bg-white/10" />
      <span className="text-[11px] uppercase tracking-widest text-white/40">
        {label}
      </span>
      <div className="flex-1 h-px bg-white/10" />
    </div>
  );
};

export default AuthDivider;
