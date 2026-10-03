import React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/cn";
import { useThemeScope } from "./ThemeScope";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 select-none",
  {
    variants: {
      variant: {
        primary: "bg-primary text-white hover:opacity-90 active:opacity-80",
        secondary:
          "bg-secondary text-secondary-foreground hover:opacity-90 active:opacity-80 dark:text-white",
        // Not text-white: white on the accent fill measures 1.9:1 and fails
        // WCAG AA at every size. accent-foreground is near-black (~11:1).
        accent:
          "bg-accent text-accent-foreground hover:opacity-90 active:opacity-80",
        outline:
          "border border-border text-foreground bg-transparent hover:bg-muted/50 active:bg-muted",
        ghost:
          "text-foreground bg-transparent hover:bg-muted/50 active:bg-muted",
        soft: "bg-primary/10 text-primary hover:bg-primary/15",
        danger: "bg-destructive text-white hover:opacity-90",
        "dark-outline":
          "border border-white/20 text-white bg-transparent hover:bg-white/8",
      },
      size: {
        // 30px, for toolbar and table-row actions in dense grids. One requested
        // size rather than each screen inventing its own height.
        xs: "h-[30px] px-3.5 text-xs",
        sm: "h-8 px-3 text-xs",
        md: "h-10 px-4 text-sm",
        lg: "h-12 px-6 text-base",
      },
      fullWidth: {
        true: "w-full",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "md",
    },
  },
);

export interface ButtonProps
  extends
    Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "prefix">,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  loading?: boolean;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = "primary",
      size,
      fullWidth,
      asChild,
      loading,
      prefix,
      suffix,
      children,
      disabled,
      type,
      style,
      ...props
    },
    ref,
  ) => {
    const scope = useThemeScope();
    const Comp = asChild ? Slot : "button";
    const buttonType = type ?? "button";
    return (
      <Comp
        ref={ref}
        className={cn(buttonVariants({ variant, size, fullWidth, className }))}
        style={{
          ...(variant === "primary" && scope?.["--app-action-bg"]
            ? {
                backgroundColor: "var(--app-action-bg)",
                color: "var(--app-on-action)",
              }
            : {}),
          ...style,
        }}
        disabled={disabled || loading}
        {...(!asChild
          ? { type: buttonType }
          : type !== undefined
            ? { type }
            : {})}
        {...props}
      >
        {loading ? (
          <span className="inline-block w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin-slow" />
        ) : (
          prefix && <span className="shrink-0">{prefix}</span>
        )}
        {children}
        {!loading && suffix && <span className="shrink-0">{suffix}</span>}
      </Comp>
    );
  },
);

Button.displayName = "Button";

export { Button, buttonVariants };
