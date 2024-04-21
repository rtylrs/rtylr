import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import VbitezButton from "../../../../components/Buttons/Buttons";

function Hero() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <div className="Hero">
      <div className="hero-text">
        <h1>{t("Hero.welcome")}</h1>
        <h2>{t("Hero.discover")}</h2>
        <div className="hero-buttons">
          <VbitezButton onClick={() => navigate("/shop")}>
            {t("Hero.Get Started Now")}
          </VbitezButton>
          <VbitezButton onClick={() => navigate("/shop")} variant="secondary">
            {t("Hero.Learn More")}
          </VbitezButton>
        </div>
      </div>
    </div>
  );
}

export default Hero;
