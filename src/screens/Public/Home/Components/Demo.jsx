import { useTranslation } from "react-i18next";
import VbitezButton from "../../../../components/Buttons/Buttons";

function Demo() {
  const { t } = useTranslation();
  return (
    <div className="demo-section">
      <div className="demo-menu-container">
        <h2>{t("Demo.title")}</h2>
        <p>{t("Demo.description")}</p>
        <div className="qr-code">
          <img src="/images/VoxireWebsiteQR.png" alt="QR Code" />
        </div>
        <VbitezButton variant="default">{t("Demo.button")}</VbitezButton>
      </div>
    </div>
  );
}

export default Demo;
