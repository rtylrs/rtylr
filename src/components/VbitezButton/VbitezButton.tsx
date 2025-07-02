import React, { CSSProperties, MouseEvent, ButtonHTMLAttributes } from "react";
import classNames from "classnames";
import { CircularProgress } from "@mui/material";
import "./VbitezButton.scss";

interface VbitezButtonProps
  extends Omit<
    ButtonHTMLAttributes<HTMLButtonElement>,
    "onClick" | "prefix" | "suffix"
  > {
  children: React.ReactNode;
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void;
  variant?: "primary" | "secondary" | "outline" | "ghost";
  size?: "sm" | "md" | "lg";
  color?: string;
  secondaryColor?: string;
  hidden?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
  styles?: CSSProperties;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
}

/**
 * Modern, accessible button component for the Vbitez UI.
 *
 * @param children - The content to display inside the button
 * @param onClick - Click event handler
 * @param variant - Button style variant: primary, secondary, outline, or ghost
 * @param size - Button size: sm, md, or lg
 * @param color - Custom primary color
 * @param secondaryColor - Custom secondary color
 * @param hidden - Whether to hide the button
 * @param loading - Show loading spinner
 * @param fullWidth - Make button full width
 * @param styles - Additional inline styles
 * @param disabled - Disable the button
 * @param className - Additional CSS classes
 * @param prefix - Icon/image to display before the text
 * @param suffix - Icon/image to display after the text
 */
const VbitezButton: React.FC<VbitezButtonProps> = ({
  children,
  onClick,
  variant = "primary",
  size = "md",
  hidden = false,
  disabled = false,
  loading = false,
  fullWidth = false,
  color,
  secondaryColor,
  styles,
  className,
  prefix,
  suffix,
  ...props
}) => {
  const buttonClassName = classNames(
    "vbitez-button",
    `vbitez-button-${variant}`,
    `vbitez-button-${size}`,
    {
      "vbitez-button-disabled": disabled,
      "vbitez-button-loading": loading,
      "vbitez-button-full-width": fullWidth,
    },
    className
  );

  const buttonStyle: CSSProperties = {
    ...(color && variant === "primary"
      ? { backgroundColor: color, borderColor: color }
      : {}),
    ...(secondaryColor && variant === "secondary"
      ? { backgroundColor: secondaryColor, borderColor: secondaryColor }
      : {}),
    ...(color && variant === "outline" ? { borderColor: color, color } : {}),
    ...(color && variant === "ghost" ? { color } : {}),
    ...styles,
    display: hidden ? "none" : "inline-flex",
  };

  const handleClick = (event: MouseEvent<HTMLButtonElement>) => {
    if (!disabled && !loading && onClick) {
      onClick(event);
    }
  };

  const renderContent = () => {
    if (loading) {
      return (
        <div className="vbitez-button-content">
          <CircularProgress size={20} color="inherit" />
          <span className="vbitez-button-text">{children}</span>
        </div>
      );
    }

    return (
      <div className="vbitez-button-content">
        {prefix && <span className="vbitez-button-prefix">{prefix}</span>}
        <span className="vbitez-button-text">{children}</span>
        {suffix && <span className="vbitez-button-suffix">{suffix}</span>}
      </div>
    );
  };

  return (
    <button
      className={buttonClassName}
      onClick={handleClick}
      disabled={disabled || loading}
      style={buttonStyle}
      aria-disabled={disabled || loading}
      {...props}
    >
      {renderContent()}
    </button>
  );
};

export default VbitezButton;
