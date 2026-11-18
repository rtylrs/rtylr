import React from "react";
import { cn } from "../lib/cn";
import { avatarColor } from "../lib/avatarColor";

type AvatarSize = "xs" | "sm" | "md" | "lg";

interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  src?: string;
  alt?: string;
  initials?: string;
  size?: AvatarSize;
  /**
   * The person's id. Picks their colour from the shared avatar palette, so
   * colleagues with the same initials still look different. Without it the
   * avatar keeps the accent tint.
   */
  seed?: string;
}

const sizeClasses: Record<AvatarSize, string> = {
  // 24px, for the 32-44px rows in dense tables where sm at 28px crowds.
  xs: "h-6 w-6 text-[10px]",
  sm: "h-7 w-7 text-xs",
  md: "h-9 w-9 text-sm",
  lg: "h-12 w-12 text-base",
};

const Avatar = React.forwardRef<HTMLDivElement, AvatarProps>(
  (
    { src, alt, initials, size = "md", seed, className, style, ...props },
    ref,
  ) => {
    const color = seed && !src ? avatarColor(seed) : null;
    return (
      <div
        ref={ref}
        className={cn(
          "relative inline-flex items-center justify-center rounded-full overflow-hidden bg-accent/20 text-accent font-semibold shrink-0",
          sizeClasses[size],
          className,
        )}
        style={
          color
            ? {
                background: color.background,
                color: color.foreground,
                ...style,
              }
            : style
        }
        data-avatar-seeded={color ? "" : undefined}
        {...props}
      >
        {src ? (
          <img
            src={src}
            alt={alt || ""}
            className="h-full w-full object-cover"
          />
        ) : (
          <span>{initials || "?"}</span>
        )}
      </div>
    );
  },
);

Avatar.displayName = "Avatar";

export { Avatar };
export type { AvatarProps, AvatarSize };
