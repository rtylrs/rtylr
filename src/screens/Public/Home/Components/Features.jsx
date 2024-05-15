import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faStar } from "@fortawesome/free-solid-svg-icons";
import useColors from "../../../../hooks/useColors";

import { faUser } from "@fortawesome/free-solid-svg-icons"; // Example icon, import your desired icons similarly
import { faCamera } from "@fortawesome/free-solid-svg-icons";
import { faCircleInfo } from "@fortawesome/free-solid-svg-icons";
import { faGlobe } from "@fortawesome/free-solid-svg-icons";
import { faUserPen } from "@fortawesome/free-solid-svg-icons";
import { faFilter } from "@fortawesome/free-solid-svg-icons";

function Features() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { accent } = useColors();

  return (
    <div className="Features">
      <h1 className="vbitez-title features-text">
        Discover What Sets Us Apart
      </h1>

      <div className="features-container">
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faCamera} style={{ color: "#655496" }} />{" "}
          </div>
          <div className="feature-container">
            <h4>
              <b>Visual Menu Showcase</b>
            </h4>
            <p>
              Showcase your dishes with stunning visuals, enticing customers to
              explore your offerings.
            </p>
          </div>
        </div>

        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faCircleInfo} />{" "}
          </div>
          <div className="feature-container">
            <h4>
              <b>Intuitive Menu Management</b>
            </h4>
            <p>
              Comprehensive information about each menu item, including
              ingredients and potential allergens, to help customers make
              informed choices.
            </p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faGlobe} />
          </div>
          <div className="feature-container">
            <h4>
              <b>Multiple Language Support</b>
            </h4>
            <p>
              Expand your reach with multilingual menu translations, attracting
              customers from around the world.
            </p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faUserPen} style={{ color: "#000000" }} />{" "}
          </div>
          <div className="feature-container">
            <h4>
              <b>Customizable Orders</b>
            </h4>
            <p>
              Tailor orders to meet your customers' preferences and dietary
              requirements, ensuring satisfaction with every meal.
            </p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faFilter} />
          </div>
          <div className="feature-container">
            <h4>
              <b>Comprehensive Menu Insights</b>
            </h4>
            <p>
              Gain detailed insights into your menu, including ingredients and
              allergens, to make informed decisions.
            </p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faStar} style={{ color: "#FFD43B" }} />
          </div>
          <div className="feature-container">
            <h4>
              <b>Loyalty Rewards Integration</b>
            </h4>
            <p>
              Build customer loyalty and drive repeat business with integrated
              loyalty rewards programs tailored to your restaurant.
            </p>
          </div>
        </div>

        {/* Add four more similar divs for other features */}
      </div>
    </div>
  );
}

export default Features;
/*
  File Created By: Abed El-Fattah Amouneh
  Last Edited: Sunday 12 May 2024
  Last Edited by: Abed El-Fattah Amouneh
*/
