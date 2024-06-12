/*
  File Created By: Mohamad Homsi
  Last Edited: Thursday 16 May 2024
  Last Edited by: Mohamad Homsi
*/

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faInstagram, faWhatsapp } from "@fortawesome/free-brands-svg-icons";
import { Link } from "react-router-dom";
import { faEnvelope } from "@fortawesome/free-solid-svg-icons";
import { useSelector } from "react-redux";
import PackageJson from "../../../package.json";
import { selectLanguage } from "../../services/i18next";
import { useEffect, useState } from "react";

const Footer = () => {
  const config = useSelector((state) => state.config.contactInfo);
  const [language, setLanguage] = useState("en");

  useEffect(() => {
    const language = localStorage.getItem("language");

    if (language) {
      selectLanguage(language);
      setLanguage(language);
    }
  }, []);

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
            <li>
              <Link to="https://voxire.com/blog">Blog</Link>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Support</h3>
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
              <Link to="/ ">Acceptable Use Policy</Link>
            </li>
            <li>
              <Link to="/ ">End User License Agreement</Link>
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
          <a href="https://api.whatsapp.com/send?phone=96171642565">
            <FontAwesomeIcon icon={faWhatsapp} />
          </a>

          <a href="https://www.instagram.com/voxire.tech">
            <FontAwesomeIcon icon={faInstagram} />
          </a>
          <a href="mailto:TheSanityStore@gmail.com">
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
      </div>

      <div className="footer-version">
        <p
          onMouseEnter={() => {
            document.querySelector(".footer-version p").innerHTML =
              "Version " + PackageJson.version;
          }}
          onMouseLeave={() => {
            document.querySelector(".footer-version p").innerHTML =
              PackageJson.version;
          }}
        >
          {PackageJson.version}
        </p>
      </div>
      {/* 
      <div className="footer-language">
        <select
          onChange={(e) => {
            selectLanguage(e.target.value);
            setLanguage(e.target.value);
          }}
          value={language}
        >
          <option value="en">English</option>
          <option value="fr">French</option>
        </select>
      </div> */}
    </footer>
  );
};

export default Footer;
