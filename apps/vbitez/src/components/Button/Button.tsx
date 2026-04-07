import React from "react";
import "./Button.scss";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "danger";
  size?: "sm" | "md" | "lg";
  children: React.ReactNode;
  styles?: React.CSSProperties;
}

export const Button: React.FC<ButtonProps> = ({
  variant = "primary",
  size = "md",
  children,
  className = "",
  styles,
  ...props
}) => {
  const buttonClass = `voxire-button voxire-button--${variant} voxire-button--${size} ${className}`;

  return (
    <button className={buttonClass} style={styles} {...props}>
      {children}
    </button>
  );
};
