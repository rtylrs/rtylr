import React from 'react';
import { ACCENT_DEFAULT } from '../lib/colorUtils';
import { withAlpha } from '../lib/colorUtils';

interface AssigneeAvatarProps {
  name: string;
  initials: string;
  size?: number;
  color?: string;
  stacked?: boolean;
}

export const AssigneeAvatar: React.FC<AssigneeAvatarProps> = ({
  name,
  initials,
  size = 22,
  color = ACCENT_DEFAULT,
  stacked = false,
}) => {
  return (
    <span
      title={name}
      className="rounded-full inline-flex items-center justify-center font-bold shrink-0"
      style={{
        width: size,
        height: size,
        background: withAlpha(color, 13),
        color,
        // Initials scale with the circle, so the scale cannot supply this
        // one: at the sizes in use (18 to 28) a 12px floor overflows a two
        // character monogram, and reaching 12px proportionally needs a 30px
        // avatar. Growing avatars changes row height, so that is a density
        // call and not this file's to make. The floor is 11 rather than 9 in
        // the meantime, and typeScale.test.ts names this as its one
        // exception so it cannot be copied silently.
        fontSize: Math.max(11, Math.round(size * 0.4)),
        border: stacked ? '2px solid var(--flow-card)' : undefined,
        boxSizing: 'border-box',
      }}
    >
      {initials}
    </span>
  );
};

interface AvatarStackProps {
  members: { name: string; initials: string; color?: string }[];
  max?: number;
  size?: number;
}

export const AvatarStack: React.FC<AvatarStackProps> = ({ members, max = 4, size = 22 }) => {
  const shown = members.slice(0, max);
  const remaining = members.length - shown.length;
  return (
    <div className="flex items-center -space-x-1.5">
      {shown.map((m, i) => (
        <AssigneeAvatar
          key={i}
          name={m.name}
          initials={m.initials}
          size={size}
          color={m.color}
          stacked
        />
      ))}
      {remaining > 0 && (
        <span
          className="rounded-full inline-flex items-center justify-center font-semibold"
          style={{
            width: size,
            height: size,
            background: 'var(--flow-hover)',
            color: 'var(--flow-text-muted)',
            fontSize: 'var(--text-flow-meta)',
            border: '2px solid var(--flow-card)',
          }}
        >
          +{remaining}
        </span>
      )}
    </div>
  );
};

export default AssigneeAvatar;
