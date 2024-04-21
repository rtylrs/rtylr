import React from "react";
import PropTypes from "prop-types";
import "./buttons.scss";
import useColors from "../../hooks/useColors";

function VbitezButton({
  children,
  onClick,
  styles,
  variant,
  color,
  secondaryColor,
}) {
  const buttonClassName = `vbitez-button-${variant}`;
  const { primary, secondary } = useColors();

  const mergedStyles = {
    ...styles,
    backgroundColor: color || primary,
    borderColor: color || primary,
  };

  const secondaryStyles = {
    ...styles,
    borderColor: color || primary,
    backgroundColor: "#ffffff",
    color: color || primary,
  };

  return (
    <button
      onClick={onClick}
      style={variant === "secondary" ? secondaryStyles : mergedStyles}
      className={buttonClassName}
      onMouseEnter={(e) => {
        (e.target.style.backgroundColor = secondaryColor || secondary),
          variant === "secondary" && (e.target.style.color = "#ffffff");
      }}
      onMouseLeave={
        variant === "secondary"
          ? (e) => {
              (e.target.style.backgroundColor = "#ffffff"),
                (e.target.style.color = color || primary);
            }
          : (e) =>
              (e.target.style.backgroundColor = mergedStyles.backgroundColor)
      }
    >
      {children}
    </button>
  );
}

VbitezButton.propTypes = {
  children: PropTypes.string.isRequired,
  onClick: PropTypes.func,
  styles: PropTypes.object,
  variant: PropTypes.oneOf(["default", "secondary"]).isRequired,
  color: PropTypes.string,
  secondaryColor: PropTypes.string,
};

VbitezButton.defaultProps = {
  variant: "default",
};

export default VbitezButton;
