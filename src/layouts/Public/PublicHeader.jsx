/*
  File Created By: Mohamad Homsi
  Last Edited: Thursday 16 May 2024
  Last Edited by: Mohamad Homsi
*/

import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faBars,
  faChevronDown,
  faChevronUp,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";
import { Link, useNavigate } from "react-router-dom";

import VbitezButton from "../../components/Buttons/Buttons";
import useColors from "../../hooks/useColors";
import useWindowDimensions from "../../utils/useWindowDimensions";
import { useSelector } from "react-redux";

const Header = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const navigate = useNavigate();
  const { primary, secondary } = useColors();
  const { width } = useWindowDimensions();
  const [dropdown, setDropdown] = useState(false);

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
      className={`header ${isMobileMenuOpen ? "mobile-menu-open" : ""}
          ${scrolled ? "scrolled" : ""}`}
      style={{
        backgroundColor:
          document.location.pathname === "/" ? "#fff" : "#f5f5f5",
        position: "fixed",
      }}
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
        <p
          style={{
            color: primary,
          }}
          onClick={() => {
            setDropdown(!dropdown);
          }}
        >
          Company
          <FontAwesomeIcon
            icon={faChevronDown}
            style={{
              marginLeft: "5px",
              fontSize: "14px",
              transform: dropdown ? "rotate(180deg)" : "rotate(0deg)",
            }}
          />
        </p>

        <div
          className="header-dropdown"
          style={{
            display: dropdown ? "block" : "none",
          }}
        >
          <h6
            onClick={() => {
              setDropdown(false);
              navigate("/about");
            }}
          >
            About Us
          </h6>
          <h6
            onClick={() => {
              setDropdown(false);
              navigate("/contact");
            }}
          >
            Contact Us
          </h6>

          <h6
            onClick={() => {
              setDropdown(false);
              navigate("/faqs");
            }}
          >
            FAQs
          </h6>
          <h6
            onClick={() => {
              setDropdown(false);
              window.open("https://voxire.com/careers", "_blank");
            }}
          >
            Careers (Jobs)
          </h6>
        </div>
      </div>

      <div className="header-right">
        <button
          className="mobile-menu-toggle"
          onClick={toggleMobileMenu}
          style={{
            color: primary,
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
            display: width < 768 ? "none" : user.token ? "none" : "flex",
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
              navigate(
                user.role === "business"
                  ? "/app/restaurant/"
                  : "/auth/register/restaurant"
              );
            } else {
              navigate("/auth/register/restaurant");
            }
          }}
          className="browse-products"
          hidden={width < 768}
        >
          {user.token ? "My Dashboard" : "Register Now"}
        </VbitezButton>
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <LinkWithColor to="/menu">Menu</LinkWithColor>
          <LinkWithColor to="/deals">Deals</LinkWithColor>
          <LinkWithColor to="/blog">blog</LinkWithColor>
          <LinkWithColor to="/about">About</LinkWithColor>
          <LinkWithColor to="/contact">Contact Us</LinkWithColor>
          <VbitezButton
            styles={{
              display: user.token ? "none" : "flex",
              marginBottom: "0px",
            }}
            onClick={() => {
              navigate("/auth/login");
            }}
          >
            <p>Login</p>
          </VbitezButton>

          <VbitezButton
            onClick={() => {
              if (user) {
                if (user.role === "admin") {
                  navigate("/manage");
                  return;
                }
                navigate(
                  user.role === "business" ? "/app/restaurant/" : "/user"
                );
              } else {
                navigate("/auth/register/restaurant");
              }
            }}
            className="browse-products"
          >
            {user.token
              ? user.role === "admin"
                ? "Ya Hussein"
                : "My Account"
              : "Register Now"}
          </VbitezButton>
        </div>
      )}
    </header>
  );
};

export default Header;
