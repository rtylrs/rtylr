import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faInstagram, faWhatsapp } from "@fortawesome/free-brands-svg-icons";
import { Link } from "react-router-dom";
import { faEnvelope } from "@fortawesome/free-solid-svg-icons";
import { useSelector } from "react-redux";
import PackageJson from "../../../package.json";
import { RootState } from "../../redux/store";
import React from "react";

const Footer: React.FC<{}> = () => {
  const config = useSelector((state: RootState) => state.config.contactInfo);

  return (
    <footer className="footer">
      <div className="footer-top">
        <img src={"/logo.png"} alt="Logo" className="footer-logo" />
      </div>
      <div className="footer-sections">
        <div className="footer-section">
          <h3>Company</h3>
          <ul>
            <li>
              <Link to="/about">About Us</Link>
            </li>
            <li>
              <Link to="/contact">Contact Us</Link>
            </li>
            <li>
              <Link to="https://voxire.com/careers">Careers (Jobs)</Link>
            </li>
            {/* <li>
              <Link to="https://voxire.com/blog">Blog</Link>
            </li> */}
          </ul>
        </div>
        <div className="footer-section">
          <h3>Quick Links</h3>
          <ul>
            <li>
              <Link to="/faqs">FAQs</Link>
            </li>
            <li>
              <Link to="/pricing">Pricing</Link>
            </li>
            <li>
              <Link to="/feedback">Feedback</Link>
            </li>
            <li>
              <Link to="/feedback">Feature Request</Link>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Legal</h3>
          <ul>
            <li>
              <Link to="/">Privacy Policy</Link>
            </li>
            <li>
              <Link to="/">Terms and Conditions</Link>
            </li>
            <li>
              <Link to="/ ">Disclaimer</Link>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Contact</h3>
          <ul>
            <li>
              Email:{" "}
              <span
                style={{
                  fontWeight: "bold",
                  cursor: "pointer",
                }}
                onClick={() => {
                  window.open("mailto:vbitez@voxire.com");
                }}
              >
                {config && config.contactEmail}
              </span>
            </li>
            <li>
              Phone:
              {config && config.contactNumber}
            </li>
          </ul>
        </div>
      </div>
      <div className="footer-social">
        <h3>Follow or Contact Us</h3>
        <div className="social-icons">
          <a href="https://www.instagram.com/voxire.tech">
            <FontAwesomeIcon icon={faInstagram} />
          </a>
          <a href="mailto:info@vbitez.voxire.com">
            <FontAwesomeIcon icon={faEnvelope} />
          </a>
        </div>
      </div>
      <div className="footer-bottom">
        <p>
          &copy;{new Date().getFullYear()},{" "}
          <a
            style={{
              color: "#434297",
              textDecoration: "underline",
              cursor: "pointer",
            }}
            href="https://voxire.com"
          >
            Voxire
          </a>{" "}
          All rights reserved.
        </p>
        <span>
          Unless otherwise noted, the use of third-party logos does not imply
          endorsement, <br /> sponsorship, or affiliation with Vbitez.
        </span>
      </div>

      <div className="footer-version">
        <p
          onMouseEnter={() => {
            const footerVersionElement =
              document.querySelector(".footer-version p");
            if (footerVersionElement) {
              footerVersionElement.innerHTML = "Version " + PackageJson.version;
            }
          }}
          onMouseLeave={() => {
            const footerVersionElement =
              document.querySelector(".footer-version p");
            if (footerVersionElement) {
              footerVersionElement.innerHTML = PackageJson.version;
            }
          }}
        >
          {PackageJson.version}
        </p>
      </div>
    </footer>
  );
};

export default Footer;
