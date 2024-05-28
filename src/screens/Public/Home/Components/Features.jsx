import { useTranslation } from "react-i18next";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faBarsProgress,
  faPenToSquare,
  faStar,
  faCamera,
  faGlobe,
  faFilter,
} from "@fortawesome/free-solid-svg-icons";

function Features() {
  const { t } = useTranslation();

  return (
    <div className="Features">
      <h1 className="vbitez-title features-text">
        {t("Features.Features-title")}
      </h1>

      <div className="features-container">
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faCamera} style={{ color: "#655496" }} />{" "}
          </div>
          <div className="feature-container">
            <h4>
              <b>{t("Features.Feature1-title")}</b>
            </h4>
            <p>{t("Features.Feature1-description")}</p>
          </div>
        </div>

        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faBarsProgress} />{" "}
          </div>
          <div className="feature-container">
            <h4>
              <b>{t("Features.Feature2-title")}</b>
            </h4>
            <p>{t("Features.Feature2-description")}</p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faGlobe} />
          </div>
          <div className="feature-container">
            <h4>
              <b>{t("Features.Feature3-title")}</b>
            </h4>
            <p>{t("Features.Feature3-description")}</p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon
              icon={faPenToSquare}
              style={{ color: "#000000" }}
            />{" "}
          </div>
          <div className="feature-container">
            <h4>
              <b>{t("Features.Feature4-title")}</b>
            </h4>
            <p>{t("Features.Feature4-description")}</p>
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
