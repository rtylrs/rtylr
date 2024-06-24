import React, { CSSProperties, MouseEvent } from "react";

interface VbitezButtonProps {
  children: React.ReactNode;
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void;
  styles?: CSSProperties;
  variant?: "default" | "secondary";
  color?: string;
  secondaryColor?: string;
  hidden?: boolean;
  disabled?: boolean;
}

const VbitezButton: React.FC<VbitezButtonProps> = ({
  children,
  onClick,
  styles,
  variant = "default",
  hidden = false,
  disabled = false,
}) => {
  const buttonClassName = `vbitez-button-${variant}`;

  return (
    <button
      disabled={disabled}
      onClick={onClick}
      style={{
        ...styles,
        display: hidden ? "none" : "",
      }}
      className={buttonClassName}
    >
      {children}
    </button>
  );
};

export default VbitezButton;
