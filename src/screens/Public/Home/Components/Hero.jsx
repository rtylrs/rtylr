import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faStar } from "@fortawesome/free-solid-svg-icons";
import useColors from "../../../../hooks/useColors";

function Hero() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { accent } = useColors();

  return (
    <div className="Hero">
      <div className="hero-text">
        <h1 className="vbitez-title">{t("Hero.welcome")}</h1>
        <h2>{t("Hero.discover")}</h2>
        <div className="hero-buttons">
          <VbitezButton onClick={() => navigate("/shop")}>
            {t("Hero.Create Account")}
          </VbitezButton>
          <VbitezButton onClick={() => navigate("/shop")} variant="secondary">
            {t("Hero.Learn More")}
          </VbitezButton>
        </div>
        <h5
          style={{
            padding: 0,
            marginTop: 10,
          }}
        >
          <FontAwesomeIcon icon={faStar} color={accent} /> Rated 4.7/5 - from
          over 600 reviews
        </h5>
      </div>
      <div className="hero-image">
        <img src="/images/HeroImage.png" alt="qr code on a paper" />
      </div>
    </div>
  );
}

export default Hero;
