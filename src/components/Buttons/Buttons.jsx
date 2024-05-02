import PropTypes from "prop-types";
import useColors from "../../hooks/useColors";

function VbitezButton({
  children,
  onClick,
  styles,
  variant,
  color,
  secondaryColor,
  hidden,
}) {
  const buttonClassName = `vbitez-button-${variant}`;
  const { primary, secondary } = useColors();

  const mergedStyles = {
    ...styles,
    backgroundColor: color || primary,
    borderColor: color || primary,
    display: hidden ? "none" : "",
  };

  const secondaryStyles = {
    ...styles,
    borderColor: color || primary,
    backgroundColor: "#ffffff",
    color: color || primary,
    display: hidden ? "none" : "",
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
  hidden: PropTypes.bool,
};

VbitezButton.defaultProps = {
  variant: "default",
  hidden: false,
};

export default VbitezButton;
