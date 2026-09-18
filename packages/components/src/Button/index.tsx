import React, { CSSProperties, MouseEvent, ButtonHTMLAttributes } from 'react';
import classNames from 'classnames';
import LoadingSpinner from '../LoadingSpinner';
import './Button.scss';

interface ButtonProps extends Omit<
  ButtonHTMLAttributes<HTMLButtonElement>,
  'onClick' | 'prefix' | 'suffix'
> {
  children: React.ReactNode;
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void;
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'soft' | 'dark-outline';
  size?: 'sm' | 'md' | 'lg';
  color?: string;
  secondaryColor?: string;
  hidden?: boolean;
  loading?: boolean;
  fullWidth?: boolean;
  width?: string | number;
  height?: string | number;
  styles?: CSSProperties;
  prefix?: React.ReactNode;
  suffix?: React.ReactNode;
}

/**
 * @param children - The content to display inside the button
 * @param onClick - Click event handler (works exactly like onClick prop of a button)
 * @param variant - Button style variant: primary, secondary, outline, ghost, soft, or dark-outline (default: primary)
 * @param size - Button size: sm, md, or lg (default: md)
 * @param color - Custom primary color override
 * @param secondaryColor - Custom secondary color override
 * @param hidden - Whether to hide the button (default: false)
 * @param loading - Show loading spinner and disable interaction (default: false)
 * @param fullWidth - Make button full width (default: false)
 * @param width - Custom width override (string or number, e.g., "200px", "50%", 300)
 * @param height - Custom height override (string or number, e.g., "40px", "100%", 50)
 * @param styles - Additional inline styles
 * @param disabled - Disable the button (default: false)
 * @param className - Additional CSS classes
 * @param prefix - Icon/image to display before the text
 * @param suffix - Icon/image to display after the text
 */

export const Button: React.FC<ButtonProps> = ({
  children,
  onClick,
  variant = 'primary',
  size = 'md',
  hidden = false,
  disabled = false,
  loading = false,
  fullWidth = false,
  width,
  height,
  color,
  secondaryColor,
  styles,
  className,
  prefix,
  suffix,
  ...props
}) => {
  const buttonClassName = classNames(
    'voxire-button',
    `voxire-button-${variant}`,
    `voxire-button-${size}`,
    {
      'voxire-button-disabled': disabled,
      'voxire-button-loading': loading,
      'voxire-button-full-width': fullWidth,
    },
    className
  );

  const buttonStyle: CSSProperties = {
    ...(color && variant === 'primary' ? { backgroundColor: color, borderColor: color } : {}),
    ...(secondaryColor && variant === 'secondary'
      ? { backgroundColor: secondaryColor, borderColor: secondaryColor }
      : {}),
    ...(color && variant === 'outline' ? { borderColor: color, color } : {}),
    ...(color && variant === 'ghost' ? { color } : {}),
    ...(width ? { width: typeof width === 'number' ? `${width}px` : width } : {}),
    ...(height
      ? {
          height: typeof height === 'number' ? `${height}px` : height,
          minHeight: typeof height === 'number' ? `${height}px` : height,
          maxHeight: typeof height === 'number' ? `${height}px` : height,
        }
      : {}),
    ...styles,
    display: hidden ? 'none' : 'inline-flex',
  };

  const handleClick = (event: MouseEvent<HTMLButtonElement>) => {
    if (!disabled && !loading && onClick) {
      onClick(event);
    }
  };

  const renderContent = () => {
    if (loading) {
      return (
        <div className="voxire-button-content">
          <LoadingSpinner size="small" color="inherit" />
          <span className="voxire-button-text">{children}</span>
        </div>
      );
    }

    return (
      <div className="voxire-button-content">
        {prefix && <span className="voxire-button-prefix">{prefix}</span>}
        <span className="voxire-button-text">{children}</span>
        {suffix && <span className="voxire-button-suffix">{suffix}</span>}
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
