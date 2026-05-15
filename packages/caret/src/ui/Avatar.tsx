import React from 'react';
import { caretTokens } from './tokens';

export interface AvatarProps {
  role: 'user' | 'assistant' | 'tool' | 'system';
  initials?: string;
}

const ROLE_COLORS: Record<AvatarProps['role'], { bg: string; fg: string }> = {
  user: { bg: caretTokens.accent, fg: '#ffffff' },
  assistant: { bg: '#0ea5e9', fg: '#ffffff' },
  tool: { bg: caretTokens.warn, fg: '#1a1a1a' },
  system: { bg: caretTokens.textDim, fg: '#ffffff' },
};

export const Avatar: React.FC<AvatarProps> = ({ role, initials }) => {
  const c = ROLE_COLORS[role];
  return (
    <span
      aria-hidden="true"
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: 24,
        height: 24,
        borderRadius: 12,
        background: c.bg,
        color: c.fg,
        fontSize: 11,
        fontWeight: 600,
        flexShrink: 0,
      }}
    >
      {initials ?? role[0].toUpperCase()}
    </span>
  );
};

export default Avatar;
