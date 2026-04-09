import React from "react";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "danger";
  size?: "sm" | "md" | "lg";
  children: React.ReactNode;
  styles?: React.CSSProperties;
}

const variantClasses = {
  primary: "bg-admin text-white hover:opacity-90",
  secondary: "bg-black/8 text-primary hover:bg-black/12",
  outline: "border border-black/15 text-primary hover:bg-black/5",
  danger: "bg-destructive text-white hover:opacity-90",
};

const sizeClasses = {
  sm: "px-3 py-1.5 text-xs",
  md: "px-4 py-2 text-sm",
  lg: "px-5 py-2.5 text-base",
};

export const Button: React.FC<ButtonProps> = ({
  variant = "primary",
  size = "md",
  children,
  className = "",
  styles,
  ...props
}) => {
  return (
    <button
      className={`inline-flex items-center justify-center font-medium rounded-lg transition-all disabled:opacity-50 ${variantClasses[variant]} ${sizeClasses[size]} ${className}`}
      style={styles}
      {...props}
    >
      {children}
    </button>
  );
};
