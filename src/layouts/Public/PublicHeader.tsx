import React, { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faBars,
  faChevronDown,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";
import { Link, useNavigate } from "react-router-dom";

import VbitezButton from "../../components/Buttons/Buttons";
import useColors from "../../hooks/useColors";
import useWindowDimensions from "../../utils/useWindowDimensions";
import { useSelector } from "react-redux";
import { selectLanguage } from "../../services/i18next";
import { RootState } from "../../redux/store";

interface LinkWithColorProps
  extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  color?: string;
  to: string;
}

const Header: React.FC<{}> = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState<boolean>(false);
  const navigate = useNavigate();
  const { primary } = useColors();
  const { width } = useWindowDimensions();
  const [dropdown, setDropdown] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [language, setLanguage] = useState<string>("en");

  useEffect(() => {
    const storedLanguage = localStorage.getItem("language");
    if (storedLanguage) {
      setLanguage(storedLanguage);
    }
  }, []);

  const toggleDropdown = () => {
    setIsOpen(!isOpen);
  };

  const user = useSelector((state: RootState) => state.user);

  const LinkWithColor: React.FC<LinkWithColorProps> = (props) => {
    const { color, to, ...rest } = props;
    return (
      <Link
        style={{
          color: color || primary,
        }}
        to={to}
        {...rest}
      />
    );
  };

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      if (window.scrollY > 150) {
        setScrolled(true);
      } else {
        setScrolled(false);
      }
    };

    window.addEventListener("scroll", handleScroll);
    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, []);

  const handleClick = () => {
    const token = localStorage.getItem("token");

    if (token) {
      if (user.role === "business" || user.role === "admin") {
        navigate("/app/business/");
      } else {
        navigate("/app/user/");
      }
    } else {
      navigate("/auth/register/business");
    }
  };

  return (
    <header
      className={`header ${isMobileMenuOpen ? "mobile-menu-open" : ""}
          ${scrolled ? "scrolled" : ""}
            ${document.location.pathname !== "/" ? "inner" : ""}
          `}
      style={{
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
        <LinkWithColor to="/pricing">Pricing</LinkWithColor>
        <LinkWithColor to="/uses">Use Cases</LinkWithColor>
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
          onClick={handleClick}
          className="browse-products"
          hidden={width < 768}
        >
          {user.token ? "Dashboard" : "Register Now"}
        </VbitezButton>
        <div className="lang-dropdown-container">
          <div className="selected-language" onClick={toggleDropdown}>
            {language}
            <FontAwesomeIcon
              icon={faChevronDown}
              style={{
                transform: isOpen ? "rotate(180deg)" : "rotate(0deg)",
              }}
            />
          </div>
          {isOpen && (
            <div className="languages-dropdown">
              <span
                onClick={() => {
                  selectLanguage("en");
                  setLanguage("en");
                  setIsOpen(false);
                }}
              >
                English
              </span>
              <span
                onClick={() => {
                  selectLanguage("ar");
                  setLanguage("ar");
                  setIsOpen(false);
                }}
              >
                Arabic
              </span>
              <span
                onClick={() => {
                  selectLanguage("fr");
                  setLanguage("fr");
                  setIsOpen(false);
                }}
              >
                French
              </span>
            </div>
          )}
        </div>
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <LinkWithColor to="/menu">Menu</LinkWithColor>
          <LinkWithColor to="/deals">Deals</LinkWithColor>
          <LinkWithColor to="/blog">Blog</LinkWithColor>
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

          <VbitezButton onClick={handleClick} className="browse-products">
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
