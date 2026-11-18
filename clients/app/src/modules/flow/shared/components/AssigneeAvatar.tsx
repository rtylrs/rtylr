import React from 'react';
import { ACCENT_DEFAULT } from '../lib/colorUtils';
import { withAlpha } from '../lib/colorUtils';
import { avatarColor, avatarColorAt, useAvatarResolver } from '@voxire/ui';

interface AssigneeAvatarProps {
  name: string;
  initials: string;
  size?: number;
  color?: string;
  /** The person's id: picks their colour from the shared avatar palette. */
  seed?: string;
  stacked?: boolean;
}

export const AssigneeAvatar: React.FC<AssigneeAvatarProps> = ({
  name,
  initials,
  size = 22,
  color,
  seed,
  stacked = false,
}) => {
  // A seeded avatar is a solid circle in the person's own colour, the same
  // one they get everywhere else in the portal. An explicit colour keeps the
  // old tinted look for callers that pass one.
  const resolve = useAvatarResolver();
  const chosen = !color && seed && resolve ? resolve.render(seed, size) : null;
  const picked = !color && seed && resolve ? resolve.colorIndex(seed) : undefined;
  const seeded =
    !color && seed ? (picked === undefined ? avatarColor(seed) : avatarColorAt(picked)) : null;
  const tint = color ?? ACCENT_DEFAULT;
  if (chosen) {
    return (
      <span
        title={name}
        data-avatar-chosen=""
        className="rounded-full inline-flex items-center justify-center overflow-hidden shrink-0"
        style={{
          width: size,
          height: size,
          border: stacked ? '2px solid var(--flow-card)' : undefined,
          boxSizing: 'border-box',
        }}
      >
        {chosen}
      </span>
    );
  }
  return (
    <span
      title={name}
      className="rounded-full inline-flex items-center justify-center font-bold shrink-0"
      style={{
        width: size,
        height: size,
        background: seeded ? seeded.background : withAlpha(tint, 13),
        color: seeded ? seeded.foreground : tint,
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
  members: { name: string; initials: string; color?: string; seed?: string }[];
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
          seed={m.seed}
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
