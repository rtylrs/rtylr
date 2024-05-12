import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faStar } from "@fortawesome/free-solid-svg-icons";
import useColors from "../../../../hooks/useColors";

import { faUser } from "@fortawesome/free-solid-svg-icons"; // Example icon, import your desired icons similarly
import { faCamera } from "@fortawesome/free-solid-svg-icons";
function Features() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { accent } = useColors();

  return (
    <div className="Features">
      <div className="features-text">
        <h1>Elevate Your Experience with Modern Features</h1>
      </div>

      <div className="features-container">
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faCamera} />
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
            <FontAwesomeIcon icon={faUser} />
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
            <FontAwesomeIcon icon={faCamera} />
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
            <FontAwesomeIcon icon={faCamera} />
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
            <FontAwesomeIcon icon={faCamera} />
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
            <FontAwesomeIcon icon={faCamera} />
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
