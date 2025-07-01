import React, { CSSProperties, MouseEvent, ButtonHTMLAttributes } from "react";
import classNames from "classnames";
import { CircularProgress } from "@mui/material";

interface VbitezButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void;
  variant?: "default" | "secondary";
  color?: string;
  secondaryColor?: string;
  hidden?: boolean;
  disabled?: boolean;
  loading?: boolean;
  styles?: CSSProperties;
}

/**
 * The main button component for the Vbitez UI.
 *
 * @typedef {import('react').ButtonHTMLAttributes<HTMLButtonElement>} ButtonProps
 * @typedef {Object} VbitezButtonProps
 * @property {React.ReactNode} children The content of the button.
 * @property {ButtonProps['onClick']} onClick The button's onClick event handler.
 * @property {ButtonProps['disabled']} disabled Whether the button is disabled.
 * @property {ButtonProps['style']} style The button's style.
 * @property {ButtonProps['className']} className The button's className.
 * @property {string} [variant='default'] The button's variant. Can be either 'default' or 'secondary'.
 * @property {string} [color] The button's color.
 * @property {string} [secondaryColor] The button's secondary color.
 * @property {boolean} [hidden=false] Whether the button should be hidden.
 * @property {boolean} [loading=false] Whether the button is in a loading state.
 */
const VbitezButton: React.FC<VbitezButtonProps> = ({
  children,
  onClick,
  variant = "default",
  hidden = false,
  disabled = false,
  loading = false,
  color,
  secondaryColor,
  styles,
  ...props
}) => {
  const buttonClassName = classNames("vbitez-button", {
    [`vbitez-button-${variant}`]: variant,
    "vbitez-button-disabled": disabled,
  });

  const buttonStyle: CSSProperties = {
    ...(variant === "default" && color ? { backgroundColor: color } : {}),
    ...(variant === "secondary" && secondaryColor
      ? { backgroundColor: secondaryColor }
      : {}),
    ...styles,
    display: hidden ? "none" : undefined,
    cursor: disabled ? "not-allowed" : "pointer",
  };

  return (
    <button
      className={buttonClassName}
      onClick={!disabled ? onClick : undefined}
      disabled={disabled}
      style={buttonStyle}
      {...props}
    >
      {loading ? (
        <CircularProgress size={20} color="inherit" />
      ) : (
        <span>{children}</span>
      )}
    </button>
  );
};

export default VbitezButton;
