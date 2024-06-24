import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faStar } from "@fortawesome/free-solid-svg-icons";
import useColors from "../../../../hooks/useColors";
import { useSelector } from "react-redux";
import { selectIsAuthenticated } from "../../../../redux/reducers/userSlice";
import React from "react";

const Hero: React.FC<{}> = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { secondary } = useColors();
  const isAuthenticated = useSelector(selectIsAuthenticated);

  return (
    <div className="Hero">
      <div className="hero-text">
        <h1 className="vbitez-title">{t("Hero.welcome")}</h1>
        <h2>{t("Hero.discover")}</h2>
        <div className="hero-buttons">
          <VbitezButton
            onClick={() => {
              if (isAuthenticated) {
                navigate("/app/business");
              } else {
                navigate("/auth/register/business");
              }
            }}
          >
            {isAuthenticated ? t("Hero.My Account") : t("Hero.Create Account")}
          </VbitezButton>
          <VbitezButton onClick={() => navigate("/uses")} variant="secondary">
            {t("Hero.Learn More")}
          </VbitezButton>
        </div>
        <h5
          style={{
            padding: 0,
            marginTop: 10,
          }}
        >
          <FontAwesomeIcon icon={faStar} color={secondary} /> Rated 4.7/5 - from
          over 150 reviews
        </h5>
      </div>
      <div className="hero-image">
        <img src="/svg/Hero.svg" alt="qr code on a paper" />
      </div>
    </div>
  );
};

export default Hero;
