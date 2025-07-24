import React, { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/VbitezButton/VbitezButton";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faArrowRight, faStar } from "@fortawesome/free-solid-svg-icons";
import useColors from "../../../../hooks/useColors";
import { useSelector } from "react-redux";
import { selectIsAuthenticated } from "../../../../redux/reducers/userSlice";
import { QRCode } from "react-qrcode-logo";
import useWindowDimensions from "../../../../utils/useWindowDimensions";

const Hero: React.FC<{}> = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { width } = useWindowDimensions();
  const { primary, secondary } = useColors();
  const isAuthenticated = useSelector(selectIsAuthenticated);

  const memoizedQRCode = useMemo(
    () => (
      <QRCode
        value="https://vbitez.voxire.com/"
        bgColor="#ffffff"
        fgColor={primary}
        size={width < 768 ? 300 : 400}
        qrStyle="fluid"
      />
    ),
    [width, primary]
  );

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
                navigate("/auth/register");
              }
            }}
          >
            {isAuthenticated ? t("Hero.My Account") : t("Hero.Create Account")}
          </VbitezButton>
          <VbitezButton
            onClick={() => navigate("/uses")}
            variant="outline"
            suffix={<FontAwesomeIcon icon={faArrowRight} />}
          >
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
          over 500+ reviews
        </h5>
      </div>
      <div className="hero-image">{memoizedQRCode}</div>
    </div>
  );
};

export default Hero;
