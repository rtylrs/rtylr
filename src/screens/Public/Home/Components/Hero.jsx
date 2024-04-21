import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

function Hero() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <div className="Hero">
      <div className="hero-text">
        <h1>{t("Hero.welcome")}</h1>
        <h2>{t("Hero.discover")}</h2>
        <button className="hero-button" onClick={() => navigate("/shop")}>
          {t("Hero.shop")}
        </button>
      </div>
    </div>
  );
}

export default Hero;
