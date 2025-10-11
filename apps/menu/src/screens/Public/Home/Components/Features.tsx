import React from "react";
import { useTranslation } from "react-i18next";
import {
  FaBars,
  FaPen,
  FaStar,
  FaCamera,
  FaGlobe,
  FaFilter,
} from "react-icons/fa";

const Features: React.FC<{}> = () => {
  const { t } = useTranslation();

  return (
    <div className="Features">
      <h1 className="rtylr-title features-text">
        {t("Features.Features-title")}
      </h1>

      <div className="features-container">
        <div className="feature">
          <div className="icon">
            <FaCamera style={{ color: "#655496" }} />{" "}
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
            <FaBars />{" "}
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
            <FaGlobe />
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
            <FaPen
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
            <FaFilter />
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
            <FaStar style={{ color: "#FFD43B" }} />
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

export default React.memo(Features);
