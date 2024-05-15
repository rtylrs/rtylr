import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faBars, faTimes } from "@fortawesome/free-solid-svg-icons";
import { Link, useNavigate } from "react-router-dom";

import VbitezButton from "../../components/Buttons/Buttons";
import useColors from "../../hooks/useColors";
import useWindowDimensions from "../../utils/useWindowDimensions";
import { useSelector } from "react-redux";

const Header = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const navigate = useNavigate();
  const { primary } = useColors();
  const { width } = useWindowDimensions();

  const user = useSelector((state) => state.user);

  const LinkWithColor = (props) => {
    // eslint-disable-next-line react/prop-types
    const { color, ...rest } = props;
    return (
      <Link
        style={{
          color: color || primary,
        }}
        {...rest}
      />
    );
  };

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    window.addEventListener("scroll", () => {
      if (window.scrollY > 150) {
        setScrolled(true);
      } else {
        setScrolled(false);
      }
    });
  }, []);

  return (
    <header
      className={
        document.location.pathname === "/"
          ? `header ${isMobileMenuOpen ? "mobile-menu-open" : ""}`
          : `header-pages ${isMobileMenuOpen ? "mobile-menu-open" : ""}
          ${scrolled ? "scrolled" : ""}`
      }
    >
      <div
        className="header-center"
        style={{
          cursor: "pointer",
        }}
        onClick={() => {
          navigate("/");
        }}
      >
        <img src={"/logo.png"} alt="Logo" />
      </div>
      <div
        className="header-left"
        style={{
          transform: "translateX(5%)",
        }}
      >
        <LinkWithColor to="/restaurant/69">Demo</LinkWithColor>
        <LinkWithColor to="/pricing">Product</LinkWithColor>
        <LinkWithColor to="/contact">Testimonials</LinkWithColor>
        <LinkWithColor to="/about">About Us</LinkWithColor>
      </div>

      <div className="header-right">
        <button
          className="mobile-menu-toggle"
          onClick={toggleMobileMenu}
          style={{
            color:
              document.location.pathname === "/"
                ? isMobileMenuOpen
                  ? "#231f1b"
                  : "#f2ebe5"
                : "#231f1b",
          }}
        >
          {isMobileMenuOpen ? (
            <FontAwesomeIcon icon={faTimes} />
          ) : (
            <FontAwesomeIcon icon={faBars} />
          )}
        </button>

        <button
          className="login-button"
          style={{
            display: user.token ? "none" : "flex",
          }}
          onClick={() => {
            navigate("/auth/login");
          }}
        >
          <p>Login</p>
        </button>

        <VbitezButton
          onClick={() => {
            if (user) {
              if (user.role === "admin") {
                navigate("/manage");
                return;
              }
              navigate(
                user.role === "business" ? "/manager/restaurant/" : "/user"
              );
            } else {
              navigate("/auth/register/restaurant");
            }
          }}
          className="browse-products"
          hidden={width < 768}
        >
          {user.token
            ? user.role === "admin"
              ? "Ya Hussein"
              : "My Account"
            : "Register Now"}
        </VbitezButton>
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <LinkWithColor to="/menu">Menu</LinkWithColor>
          <LinkWithColor to="/deals">Deals</LinkWithColor>
          <LinkWithColor to="/blog">blog</LinkWithColor>
          <LinkWithColor to="/about">About</LinkWithColor>
          <LinkWithColor to="/contact">Contact Us</LinkWithColor>
          <button
            onClick={() => {
              navigate("/shop");
            }}
          >
            Explore Products
          </button>
        </div>
      )}
    </header>
  );
};

export default Header;
