import React, { useMemo, useCallback } from "react";
import { useTranslation } from "react-i18next";
import "./ticker.scss";

const TickerItem: React.FC<{ text: string; emoji: string }> = React.memo(
  ({ text, emoji }) => (
    <div className="ticker-item">
      <p>{text}</p>
      <p>{emoji}</p>
    </div>
  )
);

TickerItem.displayName = "TickerItem";

const Ticker: React.FC = () => {
  const { t } = useTranslation();

  const tickerData = useMemo(
    () => ({
      text: t("Ticker.banner_text"),
      emoji: "🔥",
    }),
    [t]
  );

  const tickerContent = useMemo(() => {
    const items = Array.from({ length: 4 }, (_, index) => (
      <TickerItem key={index} text={tickerData.text} emoji={tickerData.emoji} />
    ));

    return <>{items}</>;
  }, [tickerData]);

  const tickerWrapper = useCallback(
    () => (
      <div className="ticker_wrapper">
        <div className="ticker_content">{tickerContent}</div>
        <div className="ticker_content">{tickerContent}</div>
      </div>
    ),
    [tickerContent]
  );

  return tickerWrapper();
};

export default Ticker;
