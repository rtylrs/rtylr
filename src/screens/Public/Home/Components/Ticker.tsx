import React from "react";
import { useTranslation } from "react-i18next";

const Ticker: React.FC<{}> = () => {
  const { t } = useTranslation();

  return (
    <div className="ticker_section">
      <div className="ticker_wrapper">
        <div className="ticker_content">
          <div className="ticker-item">
            <p>🔥</p>
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
        <div className="ticker_content">
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
        <div className="ticker_content">
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
        <div className="ticker_content">
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
        <div className="ticker_content">
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
        <div className="ticker_content">
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
        <div className="ticker_content">
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
          <div className="ticker-item">
            <p>{t("Ticker.banner_text")}</p>
            <p>🔥</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Ticker;
