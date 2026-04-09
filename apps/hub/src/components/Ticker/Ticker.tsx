import React, { useMemo, useCallback } from "react";
import { useTranslation } from "react-i18next";

const TickerItem: React.FC<{ text: string; emoji: string }> = React.memo(
  ({ text, emoji }) => (
    <div className="flex items-center gap-3 px-6 shrink-0">
      <p className="text-sm text-white/90">{text}</p>
      <p>{emoji}</p>
    </div>
  )
);

TickerItem.displayName = "TickerItem";

const Ticker: React.FC = () => {
  const { t } = useTranslation();

  const tickerData = useMemo(
    () => ({ text: t("Ticker.banner_text"), emoji: "🔥" }),
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
      <div className="w-full overflow-hidden bg-admin py-2">
        <div className="flex animate-[marquee_20s_linear_infinite]">
          <div className="flex">{tickerContent}</div>
          <div className="flex">{tickerContent}</div>
        </div>
      </div>
    ),
    [tickerContent]
  );

  return tickerWrapper();
};

export default Ticker;
