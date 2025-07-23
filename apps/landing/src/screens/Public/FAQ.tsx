import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { useSelector } from "react-redux";
import { RootState } from "../../redux/store";

const Faq: React.FC = () => {
  const { t } = useTranslation();
  const [activeIndex, setActiveIndex] = useState(null);
  const faqs = useSelector((state: RootState) => state.config.faqs);

  const toggleAnswer = (index) => {
    setActiveIndex(activeIndex === index ? null : index);
  };

  return (
    <div className="faq-section">
      <h2 className="vbitez-title">{t("QandA.title")}</h2>
      <div className="faq-container">
        {faqs &&
          faqs.slice(0, 8).map((item, index) => (
            <div key={index} className="faq-item">
              <div className="faq-question" onClick={() => toggleAnswer(index)}>
                <span>Q. {item.question}</span>
                <span className="faq-icon">
                  {activeIndex === index ? "−" : "+"}
                </span>
              </div>
              {activeIndex === index && (
                <div className="faq-answer">
                  <p>{item.answer}</p>
                </div>
              )}
            </div>
          ))}
      </div>
    </div>
  );
};
export default Faq;
