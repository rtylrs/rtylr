import React from "react";
import PropTypes from "prop-types";
import "./buttons.scss";
import useColors from "../../hooks/useColors";

function VbitezButton({
  text,
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
  };

  return (
    <button
      onClick={onClick}
      style={mergedStyles}
      className={buttonClassName}
      onMouseEnter={(e) =>
        (e.target.style.backgroundColor = secondaryColor || secondary)
      }
      onMouseLeave={(e) =>
        (e.target.style.backgroundColor = mergedStyles.backgroundColor)
      }
    >
      {text}
    </button>
  );
}

VbitezButton.propTypes = {
  text: PropTypes.string.isRequired,
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
