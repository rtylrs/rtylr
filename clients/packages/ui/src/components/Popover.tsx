import React from "react";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { cn } from "../lib/cn";
import { useThemeScope } from "./ThemeScope";

const Popover = PopoverPrimitive.Root;
const PopoverTrigger = PopoverPrimitive.Trigger;
const PopoverAnchor = PopoverPrimitive.Anchor;
const PopoverClose = PopoverPrimitive.Close;

/**
 * Popover panel, portalled to the body. Unstyled beyond layering and the
 * open animation, so each app can dress it in its own tokens; collisions
 * keep it inside the viewport.
 */
const PopoverContent = React.forwardRef<
  React.ElementRef<typeof PopoverPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof PopoverPrimitive.Content>
>(
  (
    { className, sideOffset = 6, collisionPadding = 12, style, ...props },
    ref,
  ) => {
    const scope = useThemeScope();
    return (
      <PopoverPrimitive.Portal>
        <PopoverPrimitive.Content
          ref={ref}
          sideOffset={sideOffset}
          collisionPadding={collisionPadding}
          data-theme-scope={scope ? "" : undefined}
          style={{ ...scope, ...style }}
          className={cn(
            "z-50 outline-none data-[state=open]:animate-dropdown-in",
            className,
          )}
          {...props}
        />
      </PopoverPrimitive.Portal>
    );
  },
);
PopoverContent.displayName = "PopoverContent";

export { Popover, PopoverTrigger, PopoverAnchor, PopoverClose, PopoverContent };
