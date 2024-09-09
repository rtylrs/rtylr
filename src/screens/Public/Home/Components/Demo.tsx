import { useTranslation } from "react-i18next";
import VbitezButton from "../../../../components/VbitezButton";
import React from "react";
import { QRCode } from "react-qrcode-logo";
import useColors from "../../../../hooks/useColors";

const Demo: React.FC<{}> = () => {
  const { t } = useTranslation();
  const { primary } = useColors();

  return (
    <div className="demo-section">
      <div className="demo-menu-container">
        <h2>{t("Demo.title")}</h2>
        <p>{t("Demo.description")}</p>
        <div className="qr-code">
          <QRCode
            value="https://vbitez.voxire.com/"
            bgColor="#ffffff"
            fgColor={primary}
            size={300}
          />
        </div>
        <VbitezButton
          variant="default"
          styles={{
            borderColor: primary,
          }}
        >
          {t("Demo.button")}
        </VbitezButton>
      </div>
    </div>
  );
};

export default Demo;
