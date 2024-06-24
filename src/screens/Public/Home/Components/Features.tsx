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

const Features: React.FC<{}> = () => {
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
              <b>{t("Features.Feature5-title")}</b>
            </h4>
            <p>{t("Features.Feature5-description")}</p>
          </div>
        </div>
        <div className="feature">
          <div className="icon">
            <FontAwesomeIcon icon={faStar} style={{ color: "#FFD43B" }} />
          </div>
          <div className="feature-container">
            <h4>
              <b>{t("Features.Feature6-title")}</b>
            </h4>
            <p>{t("Features.Feature6-description")}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Features;
