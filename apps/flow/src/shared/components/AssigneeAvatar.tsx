import React from 'react';

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
  color = '#64C5B9',
  stacked = false,
}) => {
  return (
    <span
      title={name}
      className="rounded-full inline-flex items-center justify-center font-bold shrink-0"
      style={{
        width: size,
        height: size,
        background: `${color}22`,
        color,
        fontSize: Math.max(9, Math.round(size * 0.4)),
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
            fontSize: 10,
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
